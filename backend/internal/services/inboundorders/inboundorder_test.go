package inboundorders

import (
	"context"
	"errors"
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
	svc := NewInboundOrderService(newTestDB(t), nil)

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
	svc := NewInboundOrderService(newTestDB(t), nil)

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
	svc := NewInboundOrderService(newTestDB(t), nil)

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
	svc := NewInboundOrderService(newTestDB(t), nil)

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
	svc := NewInboundOrderService(newTestDB(t), mailer)

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
	svc := NewInboundOrderService(newTestDB(t), mailer)

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
	svc := NewInboundOrderService(newTestDB(t), nil)

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
