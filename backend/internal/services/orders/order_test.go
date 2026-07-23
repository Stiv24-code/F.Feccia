package orders

import (
	"context"
	"errors"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"fratelli-feccia/internal/dto"
	"fratelli-feccia/internal/models"
	"fratelli-feccia/pkg/database"
	"fratelli-feccia/pkg/utils"
)

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}
	if err := db.AutoMigrate(&database.Counter{}, &models.Order{}, &models.OrderItem{}, &models.Customer{}, &models.Destination{}, &models.Product{}, &models.Garage{}, &models.Driver{}, &models.Carrier{}, &models.WashStation{}, &models.Vehicle{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	return db
}

func seedCustomer(t *testing.T, db *gorm.DB, ragioneSociale string) uuid.UUID {
	t.Helper()
	c := models.Customer{ID: uuid.New(), RagioneSociale: ragioneSociale, Active: true}
	if err := db.Create(&c).Error; err != nil {
		t.Fatalf("failed to seed customer: %v", err)
	}
	return c.ID
}

func seedDestination(t *testing.T, db *gorm.DB, nome string) uuid.UUID {
	t.Helper()
	d := models.Destination{ID: uuid.New(), Nome: nome, Active: true, Lat: geoPtr(0), Lng: geoPtr(0)}
	if err := db.Create(&d).Error; err != nil {
		t.Fatalf("failed to seed destination: %v", err)
	}
	return d.ID
}

func seedProduct(t *testing.T, db *gorm.DB, codice string) uuid.UUID {
	t.Helper()
	p := models.Product{ID: uuid.New(), Codice: codice, Descrizione: codice, Active: true}
	if err := db.Create(&p).Error; err != nil {
		t.Fatalf("failed to seed product: %v", err)
	}
	return p.ID
}

// baseRequest seeds a generic customer/destinations/product and returns an
// OrderRequest referencing their real ids — the write-side DTO only carries
// ids now, the server resolves names via Preload.
func baseRequest(t *testing.T, db *gorm.DB) dto.OrderRequest {
	t.Helper()
	clienteID := seedCustomer(t, db, "Acme S.r.l.")
	caricoID := seedDestination(t, db, "Milano")
	scaricoID := seedDestination(t, db, "Roma")
	prodottoID := seedProduct(t, db, "P-"+uuid.New().String())
	return dto.OrderRequest{
		ClienteID:             clienteID.String(),
		DestinazioneCaricoID:  caricoID.String(),
		DestinazioneScaricoID: scaricoID.String(),
		DataRitiro:            "2026-01-10",
		DataConsegna:          "2026-01-12",
		Tariffa:               500,
		Items: []dto.OrderItemRequestDTO{
			{ProdottoID: prodottoID.String(), Quantita: 1, Peso: 10000},
		},
	}
}

func TestOrderService_Create_AssignsProgressivoAndDefaults(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	svc := NewOrderService(db)

	resp, err := svc.Create(ctx, baseRequest(t, db))
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if resp.Stato != "PIANIFICABILE" {
		t.Fatalf("expected initial stato PIANIFICABILE, got %q", resp.Stato)
	}
	if resp.TipoTariffa != "forfait" || resp.Tipologia != "nazionale" {
		t.Fatalf("expected defaults forfait/nazionale, got %q/%q", resp.TipoTariffa, resp.Tipologia)
	}
	if resp.Progressivo == "" {
		t.Fatalf("expected a non-empty progressivo")
	}
	if len(resp.Items) != 1 || resp.Items[0].Prodotto == nil || resp.Items[0].Prodotto.Codice == "" {
		t.Fatalf("expected 1 item to be persisted with a resolved product, got %+v", resp.Items)
	}
}

func TestOrderService_Create_ProgressivoIncrements(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	svc := NewOrderService(db)

	first, err := svc.Create(ctx, baseRequest(t, db))
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	second, err := svc.Create(ctx, baseRequest(t, db))
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if first.Progressivo == second.Progressivo {
		t.Fatalf("expected distinct progressivo values, both were %q", first.Progressivo)
	}
}

func TestOrderService_Update_ReplacesItemsAndDoesNotTouchState(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	svc := NewOrderService(db)

	created, err := svc.Create(ctx, baseRequest(t, db))
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	betaID := seedCustomer(t, db, "Beta S.p.A.")
	p002 := seedProduct(t, db, "P002")
	p003 := seedProduct(t, db, "P003")

	updateReq := baseRequest(t, db)
	updateReq.ClienteID = betaID.String()
	updateReq.Items = []dto.OrderItemRequestDTO{
		{ProdottoID: p002.String(), Quantita: 2, Peso: 500},
		{ProdottoID: p003.String(), Quantita: 3, Peso: 700},
	}
	updated, err := svc.Update(ctx, created.ID, updateReq)
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
	if updated.Cliente == nil || updated.Cliente.RagioneSociale != "Beta S.p.A." {
		t.Fatalf("expected updated cliente, got %+v", updated.Cliente)
	}
	if len(updated.Items) != 2 {
		t.Fatalf("expected item replacement to leave exactly 2 items, got %+v", updated.Items)
	}
	if updated.Stato != "PIANIFICABILE" {
		t.Fatalf("expected Update to leave stato untouched, got %q", updated.Stato)
	}
}

func TestOrderService_AssignStartCloseDelete_StateMachine(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	svc := NewOrderService(db)

	order, err := svc.Create(ctx, baseRequest(t, db))
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	// Start/Close before assign must fail (still PIANIFICABILE).
	_, err = svc.Start(ctx, order.ID)
	assertAPIError(t, err, 400)
	_, err = svc.Close(ctx, order.ID)
	assertAPIError(t, err, 400)

	assigned, err := svc.Assign(ctx, order.ID, dto.OrderAssignRequest{TargaMotrice: "AB123CD", AutistaID: uuid.New().String()})
	if err != nil {
		t.Fatalf("Assign returned error: %v", err)
	}
	if assigned.Stato != "PIANIFICATO" {
		t.Fatalf("expected stato PIANIFICATO after assign, got %q", assigned.Stato)
	}

	// Assign again must fail (no longer PIANIFICABILE).
	_, err = svc.Assign(ctx, order.ID, dto.OrderAssignRequest{})
	assertAPIError(t, err, 400)

	// Close before start must fail (still PIANIFICATO, not VIAGGIO).
	_, err = svc.Close(ctx, order.ID)
	assertAPIError(t, err, 400)

	started, err := svc.Start(ctx, order.ID)
	if err != nil {
		t.Fatalf("Start returned error: %v", err)
	}
	if started.Stato != "VIAGGIO" {
		t.Fatalf("expected stato VIAGGIO after start, got %q", started.Stato)
	}

	// Start again must fail (no longer PIANIFICATO).
	_, err = svc.Start(ctx, order.ID)
	assertAPIError(t, err, 400)

	// Delete while VIAGGIO must fail.
	err = svc.Delete(ctx, order.ID)
	assertAPIError(t, err, 400)

	closed, err := svc.Close(ctx, order.ID)
	if err != nil {
		t.Fatalf("Close returned error: %v", err)
	}
	if closed.Stato != "CHIUSO" {
		t.Fatalf("expected stato CHIUSO after close, got %q", closed.Stato)
	}

	// Close again must fail (no longer VIAGGIO).
	_, err = svc.Close(ctx, order.ID)
	assertAPIError(t, err, 400)
}

func TestOrderService_Start_RejectsOrdersOnATrip(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	svc := NewOrderService(db)

	order, err := svc.Create(ctx, baseRequest(t, db))
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if _, err := svc.Assign(ctx, order.ID, dto.OrderAssignRequest{}); err != nil {
		t.Fatalf("Assign returned error: %v", err)
	}
	// Simulate the order having been grouped into a Trip meanwhile.
	if err := svc.db.Model(&models.Order{}).Where("id = ?", order.ID).Update("viaggio_id", uuid.New()).Error; err != nil {
		t.Fatalf("failed to set viaggio_id: %v", err)
	}

	_, err = svc.Start(ctx, order.ID)
	assertAPIError(t, err, 400)
}

func TestOrderService_Discard_ValidTransitionsOnly(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	svc := NewOrderService(db)

	// From PIANIFICABILE.
	a, err := svc.Create(ctx, baseRequest(t, db))
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	discardedA, err := svc.Discard(ctx, a.ID)
	if err != nil {
		t.Fatalf("Discard returned error: %v", err)
	}
	if discardedA.Stato != "SCARTATO" {
		t.Fatalf("expected stato SCARTATO, got %q", discardedA.Stato)
	}

	// From PIANIFICATO.
	b, err := svc.Create(ctx, baseRequest(t, db))
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if _, err := svc.Assign(ctx, b.ID, dto.OrderAssignRequest{}); err != nil {
		t.Fatalf("Assign returned error: %v", err)
	}
	discardedB, err := svc.Discard(ctx, b.ID)
	if err != nil {
		t.Fatalf("Discard returned error: %v", err)
	}
	if discardedB.Stato != "SCARTATO" {
		t.Fatalf("expected stato SCARTATO, got %q", discardedB.Stato)
	}

	// Discard again must fail (no longer PIANIFICABILE/PIANIFICATO).
	_, err = svc.Discard(ctx, a.ID)
	assertAPIError(t, err, 400)

	// From VIAGGIO must fail.
	c, err := svc.Create(ctx, baseRequest(t, db))
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if _, err := svc.Assign(ctx, c.ID, dto.OrderAssignRequest{}); err != nil {
		t.Fatalf("Assign returned error: %v", err)
	}
	if _, err := svc.Start(ctx, c.ID); err != nil {
		t.Fatalf("Start returned error: %v", err)
	}
	_, err = svc.Discard(ctx, c.ID)
	assertAPIError(t, err, 400)
}

func TestOrderService_Delete_OnlyWhenPianificabile(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	svc := NewOrderService(db)

	order, err := svc.Create(ctx, baseRequest(t, db))
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if err := svc.Delete(ctx, order.ID); err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}

	resp, err := svc.GetByID(ctx, order.ID)
	if err != nil {
		t.Fatalf("GetByID returned error: %v", err)
	}
	if resp != nil {
		t.Fatalf("expected order to be hard-deleted, got %+v", resp)
	}
}

func TestOrderService_List_FiltersByStatoAndSearch(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	svc := NewOrderService(db)

	a := baseRequest(t, db)
	a.ClienteID = seedCustomer(t, db, "Acme").String()
	if _, err := svc.Create(ctx, a); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	b := baseRequest(t, db)
	b.ClienteID = seedCustomer(t, db, "Beta").String()
	created, err := svc.Create(ctx, b)
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if _, err := svc.Assign(ctx, created.ID, dto.OrderAssignRequest{}); err != nil {
		t.Fatalf("Assign returned error: %v", err)
	}

	byStato, err := svc.List(ctx, ListFilters{Stato: "PIANIFICATO"})
	if err != nil || len(byStato) != 1 || byStato[0].Cliente == nil || byStato[0].Cliente.RagioneSociale != "Beta" {
		t.Fatalf("expected 1 PIANIFICATO order (Beta), got %+v (err=%v)", byStato, err)
	}

	bySearch, err := svc.List(ctx, ListFilters{Search: "acme"})
	if err != nil || len(bySearch) != 1 {
		t.Fatalf("expected search to match Acme, got %+v (err=%v)", bySearch, err)
	}
}

func TestOrderService_ReturnSuggestions_ScoresAndFilters(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	svc := NewOrderService(db)

	romaID := seedDestination(t, db, "Roma")

	source := baseRequest(t, db)
	source.ClienteID = seedCustomer(t, db, "Cliente A").String()
	source.DestinazioneScaricoID = romaID.String()
	source.DataConsegna = "2026-01-12"
	source.Tariffa = 1000
	source.Tipologia = "nazionale"
	orderA, err := svc.Create(ctx, source)
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	goodMatch := baseRequest(t, db)
	goodMatch.ClienteID = seedCustomer(t, db, "Cliente B").String()
	goodMatch.DestinazioneCaricoID = romaID.String()
	goodMatch.DataRitiro = "2026-01-12"
	goodMatch.Tariffa = 800
	goodMatch.Tipologia = "nazionale"
	if _, err := svc.Create(ctx, goodMatch); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	tooLate := baseRequest(t, db)
	tooLate.DestinazioneCaricoID = romaID.String()
	tooLate.DataRitiro = "2026-02-01"
	if _, err := svc.Create(ctx, tooLate); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	result, err := svc.ReturnSuggestions(ctx, orderA.ID, 2, 20)
	if err != nil {
		t.Fatalf("ReturnSuggestions returned error: %v", err)
	}
	if result.Count != 1 {
		t.Fatalf("expected exactly 1 candidate within the date window, got %+v", result.Candidates)
	}
	best := result.Candidates[0]
	if best.Score < 60 {
		t.Fatalf("expected a high score for same-day + different-client + good tariffa match, got %d (%v)", best.Score, best.Reasons)
	}
}

func TestOrderService_ReturnSuggestions_InvalidParams(t *testing.T) {
	ctx := context.Background()
	svc := NewOrderService(newTestDB(t))

	_, err := svc.ReturnSuggestions(ctx, uuid.New(), 15, 20)
	assertAPIError(t, err, 400)

	_, err = svc.ReturnSuggestions(ctx, uuid.New(), 2, 0)
	assertAPIError(t, err, 400)
}

func assertAPIError(t *testing.T, err error, code int) {
	t.Helper()
	var apiErr utils.APIError
	if err == nil {
		t.Fatalf("expected an APIError %d, got nil", code)
	}
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected an APIError %d, got %v (%T)", code, err, err)
	}
	if apiErr.Code != code {
		t.Fatalf("expected APIError code %d, got %d (%s)", code, apiErr.Code, apiErr.Message)
	}
}
