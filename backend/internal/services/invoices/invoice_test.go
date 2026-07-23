package invoices

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

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}
	if err := db.AutoMigrate(&database.Counter{}, &models.Invoice{}, &models.InvoiceLine{}, &models.Order{}, &models.OrderItem{}, &models.Customer{}, &models.Destination{}, &models.Product{}, &models.Garage{}, &models.Driver{}, &models.Carrier{}, &models.WashStation{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	return db
}

func TestInvoiceService_Create_AssignsProgressivoAndDefaults(t *testing.T) {
	ctx := context.Background()
	svc := NewInvoiceService(newTestDB(t), nil)

	inv, err := svc.Create(ctx, dto.InvoiceRequest{ClienteID: uuid.New().String(), Totale: 500})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if inv.Stato != "PROFORMA" {
		t.Fatalf("expected initial stato PROFORMA, got %q", inv.Stato)
	}
	if inv.Numero == "" {
		t.Fatalf("expected a non-empty numero")
	}
}

func TestInvoiceService_Finalize_CascadesToOrders(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	svc := NewInvoiceService(db, nil)

	order := models.Order{
		ID: uuid.New(), ClienteID: uuid.New(), Stato: "CHIUSO",
		ServiziAccessori: []byte("[]"), CostiAccessori: []byte("[]"),
	}
	if err := db.Create(&order).Error; err != nil {
		t.Fatalf("failed to seed order: %v", err)
	}

	inv, err := svc.Create(ctx, dto.InvoiceRequest{
		ClienteID: uuid.New().String(),
		Righe:     []dto.InvoiceLineDTO{{OrdineID: order.ID.String(), Descrizione: "Trasporto", Totale: 500}},
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	result, err := svc.Finalize(ctx, inv.ID)
	if err != nil {
		t.Fatalf("Finalize returned error: %v", err)
	}
	if !result.OK || result.PdfArchived {
		t.Fatalf("expected ok=true, pdf_archived=false (S3 not configured), got %+v", result)
	}

	refreshed, err := svc.GetByID(ctx, inv.ID)
	if err != nil || refreshed.Stato != "DEFINITIVA" {
		t.Fatalf("expected invoice stato DEFINITIVA, got %+v (err=%v)", refreshed, err)
	}

	var updatedOrder models.Order
	db.First(&updatedOrder, "id = ?", order.ID)
	if updatedOrder.Stato != "CHIUSO" || updatedOrder.FatturaID == nil || *updatedOrder.FatturaID != inv.ID {
		t.Fatalf("expected order to stay CHIUSO with fattura_id stamped, got %+v", updatedOrder)
	}
}

func TestInvoiceService_Finalize_OnlyFromProforma(t *testing.T) {
	ctx := context.Background()
	svc := NewInvoiceService(newTestDB(t), nil)

	inv, err := svc.Create(ctx, dto.InvoiceRequest{ClienteID: uuid.New().String()})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if _, err := svc.Finalize(ctx, inv.ID); err != nil {
		t.Fatalf("first Finalize returned error: %v", err)
	}

	_, err = svc.Finalize(ctx, inv.ID)
	assertAPIError(t, err, 400)
}

func TestInvoiceService_Finalize_DoesNotCascadeNonChiusoOrders(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	svc := NewInvoiceService(db, nil)

	order := models.Order{
		ID: uuid.New(), ClienteID: uuid.New(), Stato: "VIAGGIO",
		ServiziAccessori: []byte("[]"), CostiAccessori: []byte("[]"),
	}
	db.Create(&order)

	inv, err := svc.Create(ctx, dto.InvoiceRequest{
		ClienteID: uuid.New().String(),
		Righe:     []dto.InvoiceLineDTO{{OrdineID: order.ID.String(), Totale: 500}},
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if _, err := svc.Finalize(ctx, inv.ID); err != nil {
		t.Fatalf("Finalize returned error: %v", err)
	}

	var untouched models.Order
	db.First(&untouched, "id = ?", order.ID)
	if untouched.Stato != "VIAGGIO" {
		t.Fatalf("expected non-CHIUSO order to be left untouched, got stato=%q", untouched.Stato)
	}
}

func TestInvoiceService_Delete_OnlyFromProforma(t *testing.T) {
	ctx := context.Background()
	svc := NewInvoiceService(newTestDB(t), nil)

	inv, err := svc.Create(ctx, dto.InvoiceRequest{ClienteID: uuid.New().String()})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if _, err := svc.Finalize(ctx, inv.ID); err != nil {
		t.Fatalf("Finalize returned error: %v", err)
	}

	err = svc.Delete(ctx, inv.ID)
	assertAPIError(t, err, 400)

	inv2, err := svc.Create(ctx, dto.InvoiceRequest{ClienteID: uuid.New().String()})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if err := svc.Delete(ctx, inv2.ID); err != nil {
		t.Fatalf("expected PROFORMA invoice to be deletable, got %v", err)
	}
}

func TestInvoiceService_List_FiltersByStatoAndCliente(t *testing.T) {
	ctx := context.Background()
	svc := NewInvoiceService(newTestDB(t), nil)

	clienteA := uuid.New().String()
	clienteB := uuid.New().String()
	a, err := svc.Create(ctx, dto.InvoiceRequest{ClienteID: clienteA})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if _, err := svc.Create(ctx, dto.InvoiceRequest{ClienteID: clienteB}); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if _, err := svc.Finalize(ctx, a.ID); err != nil {
		t.Fatalf("Finalize returned error: %v", err)
	}

	byCliente, err := svc.List(ctx, "", clienteB)
	if err != nil || len(byCliente) != 1 {
		t.Fatalf("expected 1 invoice for cliente-B, got %+v (err=%v)", byCliente, err)
	}

	byStato, err := svc.List(ctx, "DEFINITIVA", "")
	if err != nil || len(byStato) != 1 || byStato[0].ClienteID != clienteA {
		t.Fatalf("expected 1 DEFINITIVA invoice (cliente-A), got %+v (err=%v)", byStato, err)
	}
}
