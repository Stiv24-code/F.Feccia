package pricelists

import (
	"context"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"fratelli-feccia/internal/dto"
	"fratelli-feccia/internal/models"
)

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}
	if err := db.AutoMigrate(&models.PriceList{}, &models.PriceListItem{}, &models.Customer{}, &models.Destination{}, &models.Product{}); err != nil {
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

func geoPtr(v float64) *float64 { return &v }

func TestPriceListService_CreateAndAddItem(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	svc := NewPriceListService(db)
	clienteID := seedCustomer(t, db, "Cliente Uno")

	pl, err := svc.Create(ctx, dto.PriceListRequest{ClienteID: clienteID.String(), DataInizio: "2026-01-01", DataFine: "2026-12-31"})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	added, err := svc.AddItem(ctx, pl.ID, dto.PriceListItemRequestDTO{Tariffa: 500, TipoTariffa: "forfait"})
	if err != nil {
		t.Fatalf("AddItem returned error: %v", err)
	}
	if added.ItemsCount != 1 {
		t.Fatalf("expected items_count 1, got %d", added.ItemsCount)
	}

	refreshed, err := svc.GetByID(ctx, pl.ID)
	if err != nil || len(refreshed.Items) != 1 {
		t.Fatalf("expected 1 item on refresh, got %v (err=%v)", refreshed, err)
	}
}

func TestPriceListService_Update_DuplicatesWhenInUso(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	svc := NewPriceListService(db)
	clienteID := seedCustomer(t, db, "Cliente Uno")
	rinnovatoID := seedCustomer(t, db, "Rinnovato")

	pl, err := svc.Create(ctx, dto.PriceListRequest{ClienteID: clienteID.String()})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	db.Model(&models.PriceList{}).Where("id = ?", pl.ID).Update("in_uso", true)

	result, err := svc.Update(ctx, pl.ID, dto.PriceListRequest{ClienteID: rinnovatoID.String()})
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
	if !result.Duplicated || result.NewID == nil {
		t.Fatalf("expected a duplicated update for an in_uso pricelist, got %+v", result)
	}

	var original models.PriceList
	db.First(&original, "id = ?", pl.ID)
	if original.Active {
		t.Fatalf("expected original pricelist to be deactivated after duplicate-update")
	}

	var duplicate models.PriceList
	db.First(&duplicate, "id = ?", *result.NewID)
	if !duplicate.Active || duplicate.InUso {
		t.Fatalf("expected duplicate to be active and not in_uso, got %+v", duplicate)
	}
}

func TestPriceListService_Update_InPlaceWhenNotInUso(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	svc := NewPriceListService(db)
	clienteID := seedCustomer(t, db, "Cliente Uno")
	aggiornatoID := seedCustomer(t, db, "Aggiornato")

	pl, err := svc.Create(ctx, dto.PriceListRequest{ClienteID: clienteID.String()})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	result, err := svc.Update(ctx, pl.ID, dto.PriceListRequest{ClienteID: aggiornatoID.String()})
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
	if result.Duplicated {
		t.Fatalf("expected an in-place update, got duplicated=%v", result)
	}

	refreshed, err := svc.GetByID(ctx, pl.ID)
	if err != nil || refreshed.Cliente == nil || refreshed.Cliente.RagioneSociale != "Aggiornato" {
		t.Fatalf("expected in-place update to apply, got %+v (err=%v)", refreshed, err)
	}
}

func TestPriceListService_DeleteItem(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	svc := NewPriceListService(db)
	clienteID := seedCustomer(t, db, "Cliente Uno")

	pl, err := svc.Create(ctx, dto.PriceListRequest{ClienteID: clienteID.String()})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	added, err := svc.AddItem(ctx, pl.ID, dto.PriceListItemRequestDTO{Tariffa: 100})
	if err != nil {
		t.Fatalf("AddItem returned error: %v", err)
	}

	result, err := svc.DeleteItem(ctx, pl.ID, added.ItemID)
	if err != nil {
		t.Fatalf("DeleteItem returned error: %v", err)
	}
	if result.ItemsCount != 0 {
		t.Fatalf("expected 0 items after delete, got %d", result.ItemsCount)
	}

	_, err = svc.UpdateItem(ctx, pl.ID, added.ItemID, dto.PriceListItemRequestDTO{Tariffa: 200})
	if err == nil {
		t.Fatal("expected error updating a deleted item")
	}
}

func TestPriceListService_LookupTariff_PicksBestScoringRule(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	svc := NewPriceListService(db)
	clienteID := seedCustomer(t, db, "Cliente Uno")
	caricoID := seedDestination(t, db, "Carico Uno")
	scaricoID := seedDestination(t, db, "Scarico Uno")

	today := time.Now().UTC()
	pl, err := svc.Create(ctx, dto.PriceListRequest{
		ClienteID:  clienteID.String(),
		DataInizio: today.AddDate(0, 0, -1).Format("2006-01-02"),
		DataFine:   today.AddDate(0, 0, 1).Format("2006-01-02"),
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	// Generic rule: no tratta/prodotto/peso constraints -> score 0.
	if _, err := svc.AddItem(ctx, pl.ID, dto.PriceListItemRequestDTO{Tariffa: 300, TipoTariffa: "forfait"}); err != nil {
		t.Fatalf("AddItem returned error: %v", err)
	}
	// Specific rule: exact tratta match -> score 10, should win.
	if _, err := svc.AddItem(ctx, pl.ID, dto.PriceListItemRequestDTO{
		Tariffa: 500, TipoTariffa: "forfait",
		DestinazioneCaricoID: caricoID.String(), DestinazioneScaricoID: scaricoID.String(),
	}); err != nil {
		t.Fatalf("AddItem returned error: %v", err)
	}

	result, err := svc.LookupTariff(ctx, clienteID.String(), caricoID.String(), scaricoID.String(), "", 0)
	if err != nil {
		t.Fatalf("LookupTariff returned error: %v", err)
	}
	if !result.Found || result.Tariffa != 500 {
		t.Fatalf("expected the specific 500 rule to win (score 10 > 0), got %+v", result)
	}
}

func TestPriceListService_LookupTariff_EuroKgWithMinimoAndFuel(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	svc := NewPriceListService(db)
	clienteID := seedCustomer(t, db, "Cliente Uno")

	today := time.Now().UTC().Format("2006-01-02")
	pl, err := svc.Create(ctx, dto.PriceListRequest{ClienteID: clienteID.String(), DataInizio: today, DataFine: today})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if _, err := svc.AddItem(ctx, pl.ID, dto.PriceListItemRequestDTO{
		Tariffa: 2, TipoTariffa: "euro_kg", MinimoTassabile: 100, PercAdeguamentoCarburante: 10,
	}); err != nil {
		t.Fatalf("AddItem returned error: %v", err)
	}

	// peso (50) below minimo (100) -> billed at minimo: 2 * 100 * 1.10 = 220.
	result, err := svc.LookupTariff(ctx, clienteID.String(), "", "", "", 50)
	if err != nil {
		t.Fatalf("LookupTariff returned error: %v", err)
	}
	if !result.Found || result.Tariffa != 220 {
		t.Fatalf("expected tariffa 220 (minimo-billed + 10%% fuel), got %+v", result)
	}
}

func TestPriceListService_LookupTariff_NotFound(t *testing.T) {
	ctx := context.Background()
	svc := NewPriceListService(newTestDB(t))

	result, err := svc.LookupTariff(ctx, uuid.New().String(), "", "", "", 0)
	if err != nil {
		t.Fatalf("LookupTariff returned error: %v", err)
	}
	if result.Found || result.TipoTariffa != "forfait" {
		t.Fatalf("expected not-found default response, got %+v", result)
	}
}

func TestScoreRuleMatch(t *testing.T) {
	a := uuid.New()
	b := uuid.New()
	p1 := uuid.New()
	item := models.PriceListItem{
		ID: uuid.New(), DestinazioneCaricoID: &a, DestinazioneScaricoID: &b,
		ProdottoID: &p1, RangePesoMin: 10, RangePesoMax: 100,
	}

	other := uuid.New()
	if score := scoreRuleMatch(item, &a, &b, &p1, 50); score != 18 {
		t.Fatalf("expected score 10+5+3=18 for full match, got %d", score)
	}
	if score := scoreRuleMatch(item, &other, &other, &p1, 50); score != -1 {
		t.Fatalf("expected -1 for tratta mismatch, got %d", score)
	}
	if score := scoreRuleMatch(item, &a, &b, &other, 50); score != -1 {
		t.Fatalf("expected -1 for prodotto mismatch, got %d", score)
	}
	if score := scoreRuleMatch(item, &a, &b, &p1, 5); score != -1 {
		t.Fatalf("expected -1 for peso below range, got %d", score)
	}
}
