package dashboard

import (
	"context"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"fratelli-feccia/internal/models"
)

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}
	if err := db.AutoMigrate(&models.Order{}, &models.OrderItem{}, &models.Customer{}, &models.Vehicle{}, &models.Driver{}, &models.Invoice{}, &models.InvoiceLine{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	return db
}

func seedOrder(t *testing.T, db *gorm.DB, clienteID, stato, dataRitiro, destScarico, tipologia, categoria string, tariffa float64) models.Order {
	t.Helper()
	return seedOrderWithFattura(t, db, clienteID, stato, dataRitiro, destScarico, tipologia, categoria, tariffa, "")
}

// seedOrderWithFattura is like seedOrder but also stamps fattura_id — "is
// this order billed" is tracked via fattura_id, not a dedicated stato value.
func seedOrderWithFattura(t *testing.T, db *gorm.DB, clienteID, stato, dataRitiro, destScarico, tipologia, categoria string, tariffa float64, fatturaID string) models.Order {
	t.Helper()
	o := models.Order{
		ID: uuid.New(), ClienteID: clienteID, Stato: stato, DataRitiro: dataRitiro,
		DestinazioneScaricoNome: destScarico, Tipologia: tipologia, CategoriaTrasporto: categoria,
		Tariffa: tariffa, FatturaID: fatturaID, ServiziAccessori: []byte("[]"), CostiAccessori: []byte("[]"),
	}
	if err := db.Create(&o).Error; err != nil {
		t.Fatalf("failed to seed order: %v", err)
	}
	return o
}

func TestDashboardService_Stats_CountsAndRevenue(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	svc := NewDashboardService(db)

	seedOrder(t, db, "c1", "PIANIFICABILE", "2026-01-10", "", "nazionale", "", 100)
	seedOrder(t, db, "c1", "VIAGGIO", "2026-01-15", "", "nazionale", "", 200)
	seedOrder(t, db, "c1", "CHIUSO", "2026-02-01", "", "nazionale", "", 300)
	seedOrderWithFattura(t, db, "c1", "CHIUSO", "2026-02-10", "", "nazionale", "", 400, "inv-1")

	db.Create(&models.Customer{ID: uuid.New(), RagioneSociale: "C1", Active: true})
	db.Create(&models.Vehicle{ID: uuid.New(), Targa: "AB123CD", Active: true})
	db.Create(&models.Driver{ID: uuid.New(), Nome: "M", Cognome: "R", Active: true})

	inv := models.Invoice{ID: uuid.New(), ClienteID: "c1", Stato: "DEFINITIVA", Totale: 750, CostiAccessori: []byte("[]")}
	db.Create(&inv)
	db.Create(&models.Invoice{ID: uuid.New(), ClienteID: "c1", Stato: "PROFORMA", Totale: 999, CostiAccessori: []byte("[]")})

	stats, err := svc.Stats(ctx)
	if err != nil {
		t.Fatalf("Stats returned error: %v", err)
	}
	if stats.TotalOrders != 4 || stats.Pianificabili != 1 || stats.InViaggio != 1 || stats.Chiusi != 2 || stats.Fatturati != 1 {
		t.Fatalf("unexpected order counts: %+v", stats)
	}
	if stats.TotalCustomers != 1 || stats.TotalVehicles != 1 || stats.TotalDrivers != 1 {
		t.Fatalf("unexpected master-data counts: %+v", stats)
	}
	if stats.TotalRevenue != 750 {
		t.Fatalf("expected total_revenue 750 (only DEFINITIVA), got %v", stats.TotalRevenue)
	}
	if len(stats.MonthlyTrend) != 2 {
		t.Fatalf("expected 2 monthly groups (2026-01, 2026-02), got %+v", stats.MonthlyTrend)
	}
}

func TestDashboardService_CustomerDashboard_KPIsAndBreakdowns(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	svc := NewDashboardService(db)

	customer := models.Customer{ID: uuid.New(), RagioneSociale: "Acme", Citta: "Milano", PartitaIva: "123", Active: true}
	db.Create(&customer)
	clienteID := customer.ID.String()

	seedOrderWithFattura(t, db, clienteID, "CHIUSO", "2026-01-10", "Roma", "nazionale", "frigo", 500, "inv-1")
	seedOrder(t, db, clienteID, "CHIUSO", "2026-02-10", "Roma", "nazionale", "", 300)
	seedOrder(t, db, clienteID, "PIANIFICABILE", "2026-02-20", "Napoli", "internazionale", "", 200)
	seedOrderWithFattura(t, db, "other-client", "CHIUSO", "2026-01-10", "Torino", "nazionale", "", 999, "inv-2")

	result, err := svc.CustomerDashboard(ctx, customer.ID)
	if err != nil {
		t.Fatalf("CustomerDashboard returned error: %v", err)
	}
	if result.Customer.RagioneSociale != "Acme" {
		t.Fatalf("expected customer summary to be populated, got %+v", result.Customer)
	}
	if result.KPI.OrdiniTotali != 3 || result.KPI.OrdiniFatturati != 1 || result.KPI.OrdiniChiusi != 2 || result.KPI.OrdiniPianificabili != 1 {
		t.Fatalf("unexpected KPI: %+v", result.KPI)
	}
	if result.KPI.FatturatoNetto != 500 {
		t.Fatalf("expected fatturato_netto 500 (only orders with fattura_id set), got %v", result.KPI.FatturatoNetto)
	}
	if len(result.TopDestinazioni) != 2 {
		t.Fatalf("expected 2 distinct destinations (Roma, Napoli), got %+v", result.TopDestinazioni)
	}
	if result.TopDestinazioni[0].Nome != "Roma" || result.TopDestinazioni[0].Ordini != 2 {
		t.Fatalf("expected Roma to be the top destination with 2 orders, got %+v", result.TopDestinazioni[0])
	}
	if len(result.PerCategoria) != 1 || result.PerCategoria[0].Categoria != "frigo" {
		t.Fatalf("expected 1 categoria (frigo, empty ones excluded), got %+v", result.PerCategoria)
	}
}

func TestDashboardService_CustomerDashboard_NotFound(t *testing.T) {
	svc := NewDashboardService(newTestDB(t))
	_, err := svc.CustomerDashboard(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("expected an error for a missing customer")
	}
}

func TestDashboardService_RecentOrders_LimitsToTen(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	svc := NewDashboardService(db)

	for i := 0; i < 15; i++ {
		seedOrder(t, db, "c1", "PIANIFICABILE", "2026-01-01", "", "nazionale", "", 0)
	}

	recent, err := svc.RecentOrders(ctx)
	if err != nil {
		t.Fatalf("RecentOrders returned error: %v", err)
	}
	if len(recent) != 10 {
		t.Fatalf("expected 10 recent orders, got %d", len(recent))
	}
}
