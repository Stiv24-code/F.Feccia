package inboundorders

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"fratelli-feccia/internal/dto"
	"fratelli-feccia/internal/models"
	"fratelli-feccia/pkg/utils"
)

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}
	if err := db.AutoMigrate(&models.InboundOrder{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	return db
}

func baseRequest() dto.InboundOrderRequest {
	return dto.InboundOrderRequest{
		Client:        "ACME S.r.l.",
		SenderEmail:   "Ordini@ACME.it",
		Ref:           "ORD-42",
		Product:       "Melassa",
		Kg:            28000,
		LoadDate:      "2026-08-10",
		LoadPlace:     "Ravenna",
		DeliveryDate:  "2026-08-11",
		DeliveryPlace: "Verona",
	}
}

func assertAPIError(t *testing.T, err error, code int) {
	t.Helper()
	var apiErr utils.APIError
	if !errors.As(err, &apiErr) || apiErr.Code != code {
		t.Fatalf("expected APIError %d, got %v", code, err)
	}
}

func TestInboundOrderService_CreateDefaultsAndValidation(t *testing.T) {
	ctx := context.Background()
	svc := NewInboundOrderService(newTestDB(t), nil, nil)

	created, err := svc.Create(ctx, baseRequest())
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if created.Status != "pending" || created.Source != "pdf" {
		t.Fatalf("expected pending/pdf defaults, got %q/%q", created.Status, created.Source)
	}
	if created.SenderEmail != "ordini@acme.it" {
		t.Fatalf("expected lowercased sender, got %q", created.SenderEmail)
	}
	if created.ReceivedAt.IsZero() {
		t.Fatalf("expected received_at to default to now")
	}

	// Neither ref nor product -> 400.
	req := baseRequest()
	req.Client = "Altro"
	req.Ref = "  "
	req.Product = ""
	_, err = svc.Create(ctx, req)
	assertAPIError(t, err, 400)
}

func TestInboundOrderService_CreateDedupCaseInsensitive(t *testing.T) {
	ctx := context.Background()
	svc := NewInboundOrderService(newTestDB(t), nil, nil)

	if _, err := svc.Create(ctx, baseRequest()); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	// Same (ref, client) modulo case and surrounding spaces -> 409.
	dup := baseRequest()
	dup.Ref = "  ord-42 "
	dup.Client = "acme S.R.L."
	_, err := svc.Create(ctx, dup)
	assertAPIError(t, err, 409)

	// Different ref on the same client is fine.
	other := baseRequest()
	other.Ref = "ORD-43"
	if _, err := svc.Create(ctx, other); err != nil {
		t.Fatalf("Create returned error for distinct ref: %v", err)
	}

	list, err := svc.List(ctx)
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 orders, got %d", len(list))
	}
}

func TestInboundOrderService_AddIfNew(t *testing.T) {
	ctx := context.Background()
	svc := NewInboundOrderService(newTestDB(t), nil, nil)

	inserted, err := svc.AddIfNew(ctx, baseRequest())
	if err != nil || !inserted {
		t.Fatalf("expected first AddIfNew to insert, got inserted=%v err=%v", inserted, err)
	}
	// Duplicate is a normal outcome for the scraper, not an error.
	inserted, err = svc.AddIfNew(ctx, baseRequest())
	if err != nil {
		t.Fatalf("expected duplicate to be swallowed, got %v", err)
	}
	if inserted {
		t.Fatalf("expected inserted=false on duplicate")
	}
	// Real validation errors still surface.
	bad := baseRequest()
	bad.Ref, bad.Product = "", ""
	if _, err := svc.AddIfNew(ctx, bad); err == nil {
		t.Fatalf("expected validation error to propagate")
	}
}

func TestInboundOrderService_AcceptWithoutMailer(t *testing.T) {
	ctx := context.Background()
	svc := NewInboundOrderService(newTestDB(t), nil, nil)

	created, err := svc.Create(ctx, baseRequest())
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	res, err := svc.Accept(ctx, created.ID)
	if err != nil {
		t.Fatalf("Accept returned error: %v", err)
	}
	if res.Order.Status != "accepted" {
		t.Fatalf("expected accepted, got %q", res.Order.Status)
	}
	if res.Mail != "nessuna email inviata: SMTP non configurato" {
		t.Fatalf("unexpected mail info: %q", res.Mail)
	}
}

// fakeMailer implements AcceptanceMailer for tests.
type fakeMailer struct {
	recipient string
	err       error
	sent      []dto.InboundOrderResponse
}

func (m *fakeMailer) SendAcceptance(_ context.Context, o dto.InboundOrderResponse) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	m.sent = append(m.sent, o)
	return m.recipient, nil
}

func TestInboundOrderService_AcceptWithMailer(t *testing.T) {
	ctx := context.Background()
	mailer := &fakeMailer{recipient: "test@feccia.it"}
	svc := NewInboundOrderService(newTestDB(t), mailer, nil)

	created, err := svc.Create(ctx, baseRequest())
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	res, err := svc.Accept(ctx, created.ID)
	if err != nil {
		t.Fatalf("Accept returned error: %v", err)
	}
	if res.Mail != "email di conferma inviata a test@feccia.it" {
		t.Fatalf("unexpected mail info: %q", res.Mail)
	}
	if len(mailer.sent) != 1 || mailer.sent[0].Ref != "ORD-42" {
		t.Fatalf("expected the order to be handed to the mailer, got %+v", mailer.sent)
	}
}

func TestInboundOrderService_AcceptMailFailureLeavesPending(t *testing.T) {
	ctx := context.Background()
	mailer := &fakeMailer{err: errors.New("smtp down")}
	svc := NewInboundOrderService(newTestDB(t), mailer, nil)

	created, err := svc.Create(ctx, baseRequest())
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	_, err = svc.Accept(ctx, created.ID)
	assertAPIError(t, err, 502)

	// Send failed BEFORE the status change: the order must still be pending
	// so the operator can retry.
	list, err := svc.List(ctx)
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(list) != 1 || list[0].Status != "pending" {
		t.Fatalf("expected order still pending after mail failure, got %+v", list)
	}
}

func TestInboundOrderService_ModifyResetAndNotFound(t *testing.T) {
	ctx := context.Background()
	svc := NewInboundOrderService(newTestDB(t), nil, nil)

	created, err := svc.Create(ctx, baseRequest())
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	modified, err := svc.Modify(ctx, created.ID)
	if err != nil {
		t.Fatalf("Modify returned error: %v", err)
	}
	if modified.Status != "modify" {
		t.Fatalf("expected modify, got %q", modified.Status)
	}

	reset, err := svc.Reset(ctx, created.ID)
	if err != nil {
		t.Fatalf("Reset returned error: %v", err)
	}
	if reset.Status != "pending" {
		t.Fatalf("expected pending, got %q", reset.Status)
	}

	if _, err := svc.Modify(ctx, uuid.New()); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected ErrRecordNotFound on missing order, got %v", err)
	}
	if _, err := svc.Accept(ctx, uuid.New()); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected ErrRecordNotFound on missing order, got %v", err)
	}
}

// fakeOrderCreator stands in for the orders service: Convert's contract with
// it is just "hand me an OrderRequest, get back an id", and what the tests
// care about is the request that was built — which fields came from the
// draft, which from the operator, and which were refused.
type fakeOrderCreator struct {
	calls int
	got   dto.OrderRequest
	id    uuid.UUID
	err   error
}

func (f *fakeOrderCreator) Create(_ context.Context, req dto.OrderRequest) (*dto.OrderResponse, error) {
	f.calls++
	f.got = req
	if f.err != nil {
		return nil, f.err
	}
	if f.id == uuid.Nil {
		f.id = uuid.New()
	}
	return &dto.OrderResponse{ID: f.id, Progressivo: "26/0001"}, nil
}

// portalRequest is a draft as CreateMyInboundOrder builds one: free text for
// the dashboard plus the structured payload taken from the authenticated
// submission.
func portalRequest(clienteID, caricoID, scaricoID uuid.UUID) dto.InboundOrderRequest {
	req := baseRequest()
	req.Ref = "PORTAL-1"
	req.Source = models.InboundOrderSourcePortal
	req.ClienteID = &clienteID
	req.DestinazioneCaricoID = &caricoID
	req.DestinazioneScaricoID = &scaricoID
	req.OraRitiroDa = "08:00"
	req.OraRitiroA = "12:00"
	req.TariffaProposta = 850
	return req
}

// A mail draft's Client is whatever the sender typed, so there is nothing
// trustworthy to bill: conversion must refuse rather than name-match an
// anagrafica, which is what would let a spoofed sender have an order created
// against a real customer.
func TestInboundOrderService_ConvertRefusesUntrustedClient(t *testing.T) {
	ctx := context.Background()
	creator := &fakeOrderCreator{}
	svc := NewInboundOrderService(newTestDB(t), nil, creator)

	req := baseRequest()
	req.Source = models.InboundOrderSourceMail
	created, err := svc.Create(ctx, req)
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	_, err = svc.Convert(ctx, created.ID, dto.InboundOrderConvertRequest{})
	assertAPIError(t, err, 400)
	if creator.calls != 0 {
		t.Fatalf("no order must be created without a trusted cliente_id, got %d calls", creator.calls)
	}

	// The same draft converts once staff names the customer explicitly.
	clienteID := uuid.New()
	res, err := svc.Convert(ctx, created.ID, dto.InboundOrderConvertRequest{ClienteID: clienteID.String()})
	if err != nil {
		t.Fatalf("Convert with explicit cliente_id returned error: %v", err)
	}
	if creator.got.ClienteID != clienteID.String() {
		t.Fatalf("expected cliente_id %s, got %q", clienteID, creator.got.ClienteID)
	}
	// A mail draft's Rate is attacker-controlled free text and is never
	// parsed into a price.
	if creator.got.Tariffa != 0 {
		t.Fatalf("expected tariffa 0 for a mail draft, got %v", creator.got.Tariffa)
	}
	if res.TariffaFromClient {
		t.Fatal("a mail draft has no client-proposed rate to flag")
	}
}

func TestInboundOrderService_ConvertPortalDraftKeepsUUIDs(t *testing.T) {
	ctx := context.Background()
	creator := &fakeOrderCreator{}
	svc := NewInboundOrderService(newTestDB(t), nil, creator)

	clienteID, caricoID, scaricoID := uuid.New(), uuid.New(), uuid.New()
	created, err := svc.Create(ctx, portalRequest(clienteID, caricoID, scaricoID))
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	res, err := svc.Convert(ctx, created.ID, dto.InboundOrderConvertRequest{})
	if err != nil {
		t.Fatalf("Convert returned error: %v", err)
	}

	if creator.got.ClienteID != clienteID.String() {
		t.Fatalf("cliente_id: expected %s, got %q", clienteID, creator.got.ClienteID)
	}
	if creator.got.DestinazioneCaricoID != caricoID.String() {
		t.Fatalf("destinazione_carico_id: expected %s, got %q", caricoID, creator.got.DestinazioneCaricoID)
	}
	if creator.got.DestinazioneScaricoID != scaricoID.String() {
		t.Fatalf("destinazione_scarico_id: expected %s, got %q", scaricoID, creator.got.DestinazioneScaricoID)
	}
	if creator.got.OraRitiroDa != "08:00" || creator.got.OraRitiroA != "12:00" {
		t.Fatalf("pickup window lost: %q-%q", creator.got.OraRitiroDa, creator.got.OraRitiroA)
	}
	if creator.got.RifOrdineCliente != "PORTAL-1" {
		t.Fatalf("expected ref carried as rif_ordine_cliente, got %q", creator.got.RifOrdineCliente)
	}
	// Product and Kg have no typed home on an Order, so they must survive in
	// the note rather than be dropped.
	if !strings.Contains(creator.got.Note, "Melassa") || !strings.Contains(creator.got.Note, "28000") {
		t.Fatalf("product/kg missing from note: %q", creator.got.Note)
	}
	// The customer's proposed rate applies but is flagged as theirs.
	if creator.got.Tariffa != 850 {
		t.Fatalf("expected proposed tariffa 850, got %v", creator.got.Tariffa)
	}
	if !res.TariffaFromClient {
		t.Fatal("expected tariffa_from_client=true when the proposal is applied unchanged")
	}
	if res.InboundOrder.OrderID == nil || *res.InboundOrder.OrderID != creator.id {
		t.Fatalf("draft not linked to the created order: %v", res.InboundOrder.OrderID)
	}
}

func TestInboundOrderService_ConvertOperatorTariffaWins(t *testing.T) {
	ctx := context.Background()
	creator := &fakeOrderCreator{}
	svc := NewInboundOrderService(newTestDB(t), nil, creator)

	created, err := svc.Create(ctx, portalRequest(uuid.New(), uuid.New(), uuid.New()))
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	// An explicit zero must win over the customer's proposal, which is why
	// the field is a pointer.
	zero := 0.0
	res, err := svc.Convert(ctx, created.ID, dto.InboundOrderConvertRequest{Tariffa: &zero})
	if err != nil {
		t.Fatalf("Convert returned error: %v", err)
	}
	if creator.got.Tariffa != 0 {
		t.Fatalf("expected operator tariffa 0 to win, got %v", creator.got.Tariffa)
	}
	if res.TariffaFromClient {
		t.Fatal("expected tariffa_from_client=false when the operator set the rate")
	}
}

func TestInboundOrderService_ConvertIsIdempotent(t *testing.T) {
	ctx := context.Background()
	creator := &fakeOrderCreator{}
	svc := NewInboundOrderService(newTestDB(t), nil, creator)

	created, err := svc.Create(ctx, portalRequest(uuid.New(), uuid.New(), uuid.New()))
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if _, err := svc.Convert(ctx, created.ID, dto.InboundOrderConvertRequest{}); err != nil {
		t.Fatalf("first Convert returned error: %v", err)
	}

	_, err = svc.Convert(ctx, created.ID, dto.InboundOrderConvertRequest{})
	assertAPIError(t, err, 409)
	if creator.calls != 1 {
		t.Fatalf("expected exactly one order created, got %d", creator.calls)
	}
}

// The portal list is keyed on the conversion link, not on the status: a
// request accepted by staff but not yet converted used to disappear from the
// customer's portal while no order existed to replace it.
func TestInboundOrderService_ListForClientKeepsAcceptedUntilConverted(t *testing.T) {
	ctx := context.Background()
	creator := &fakeOrderCreator{}
	svc := NewInboundOrderService(newTestDB(t), nil, creator)

	clienteID := uuid.New()
	created, err := svc.Create(ctx, portalRequest(clienteID, uuid.New(), uuid.New()))
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	if _, err := svc.Accept(ctx, created.ID); err != nil {
		t.Fatalf("Accept returned error: %v", err)
	}
	items, err := svc.ListForClient(ctx, clienteID)
	if err != nil {
		t.Fatalf("ListForClient returned error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("an accepted but unconverted request must stay visible, got %d items", len(items))
	}

	if _, err := svc.Convert(ctx, created.ID, dto.InboundOrderConvertRequest{}); err != nil {
		t.Fatalf("Convert returned error: %v", err)
	}
	items, err = svc.ListForClient(ctx, clienteID)
	if err != nil {
		t.Fatalf("ListForClient returned error: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("a converted request must drop out in favour of the order, got %d items", len(items))
	}
}

func TestInboundOrderService_ConvertWithoutOrderCreator(t *testing.T) {
	ctx := context.Background()
	svc := NewInboundOrderService(newTestDB(t), nil, nil)

	created, err := svc.Create(ctx, portalRequest(uuid.New(), uuid.New(), uuid.New()))
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	_, err = svc.Convert(ctx, created.ID, dto.InboundOrderConvertRequest{})
	assertAPIError(t, err, 500)
}
