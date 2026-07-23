package export

import (
	"bytes"
	"context"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"

	"fratelli-feccia/internal/models"
)

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}
	if err := db.AutoMigrate(&models.Order{}, &models.OrderItem{}, &models.Customer{}, &models.Destination{}, &models.Product{}, &models.Garage{}, &models.Driver{}, &models.Carrier{}, &models.WashStation{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	return db
}

func createOrder(t *testing.T, db *gorm.DB, progressivo, dataRitiro, stato string) {
	t.Helper()
	cliente := models.Customer{ID: uuid.New(), RagioneSociale: "Cliente Uno", Active: true}
	if err := db.Create(&cliente).Error; err != nil {
		t.Fatalf("failed to seed customer: %v", err)
	}
	o := models.Order{
		ID: uuid.New(), ClienteID: cliente.ID, Progressivo: progressivo,
		DataRitiro: dataRitiro, Stato: models.OrderStato(stato), Tariffa: 100,
		ServiziAccessori: []byte("[]"), CostiAccessori: []byte("[]"),
	}
	if err := db.Create(&o).Error; err != nil {
		t.Fatalf("failed to seed order: %v", err)
	}
}

func readSheetRows(t *testing.T, data []byte) [][]string {
	t.Helper()
	xf, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("failed to open generated xlsx: %v", err)
	}
	defer xf.Close()
	rows, err := xf.GetRows("Ordini")
	if err != nil {
		t.Fatalf("failed to read Ordini sheet: %v", err)
	}
	return rows
}

func TestOrdersExcel_HeaderAndRows(t *testing.T) {
	db := newTestDB(t)
	svc := NewExportService(db)

	createOrder(t, db, "26/0001", "2026-01-05", "CHIUSO")
	createOrder(t, db, "26/0002", "2026-01-10", "VIAGGIO")

	data, err := svc.OrdersExcel(context.Background(), OrdersFilter{})
	if err != nil {
		t.Fatalf("OrdersExcel returned error: %v", err)
	}

	rows := readSheetRows(t, data)
	if len(rows) != 3 {
		t.Fatalf("expected header + 2 rows, got %d rows: %+v", len(rows), rows)
	}
	wantHeader := []string{"Progressivo", "Cliente", "Carico", "Scarico", "Data Ritiro", "Data Consegna", "Tariffa", "Tipologia", "Stato"}
	for i, h := range wantHeader {
		if rows[0][i] != h {
			t.Fatalf("header mismatch at col %d: got %q want %q", i, rows[0][i], h)
		}
	}
	if rows[1][0] != "26/0001" || rows[2][0] != "26/0002" {
		t.Fatalf("expected rows sorted ascending by data_ritiro, got %+v", rows)
	}
}

func TestOrdersExcel_FiltersByStatoAndDateRange(t *testing.T) {
	db := newTestDB(t)
	svc := NewExportService(db)

	createOrder(t, db, "26/0001", "2026-01-05", "CHIUSO")
	createOrder(t, db, "26/0002", "2026-01-10", "VIAGGIO")
	createOrder(t, db, "26/0003", "2026-02-01", "CHIUSO")

	data, err := svc.OrdersExcel(context.Background(), OrdersFilter{Stato: "CHIUSO", DataDa: "2026-01-01", DataA: "2026-01-31"})
	if err != nil {
		t.Fatalf("OrdersExcel returned error: %v", err)
	}

	rows := readSheetRows(t, data)
	if len(rows) != 2 {
		t.Fatalf("expected header + 1 filtered row, got %d rows: %+v", len(rows), rows)
	}
	if rows[1][0] != "26/0001" {
		t.Fatalf("expected only 26/0001 to match stato+range filter, got %+v", rows[1])
	}
}

func TestOrdersExcel_EmptyResultStillProducesValidHeader(t *testing.T) {
	db := newTestDB(t)
	svc := NewExportService(db)

	data, err := svc.OrdersExcel(context.Background(), OrdersFilter{})
	if err != nil {
		t.Fatalf("OrdersExcel returned error: %v", err)
	}
	rows := readSheetRows(t, data)
	if len(rows) != 1 {
		t.Fatalf("expected only the header row for no orders, got %d rows", len(rows))
	}
}
