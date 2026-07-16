package invoices

import (
	"bytes"
	"context"
	"testing"

	"github.com/google/uuid"

	"fratelli-feccia/internal/models"
)

func TestGetPDF_GeneratesOnTheFlyWhenNotArchived(t *testing.T) {
	db := newTestDB(t)
	svc := NewInvoiceService(db, nil)

	customer := models.Customer{ID: uuid.New(), RagioneSociale: "Cliente Fattura Srl", Active: true}
	if err := db.Create(&customer).Error; err != nil {
		t.Fatal(err)
	}
	inv := models.Invoice{
		ID: uuid.New(), Numero: "O/F-26/0001", ClienteID: customer.ID.String(), ClienteNome: customer.RagioneSociale,
		Stato: "PROFORMA", CostiAccessori: []byte("[]"),
		Righe: []models.InvoiceLine{{ID: uuid.New(), Descrizione: "Trasporto", Totale: 100, IvaCodice: "N8"}},
	}
	if err := db.Create(&inv).Error; err != nil {
		t.Fatal(err)
	}

	data, filename, err := svc.GetPDF(context.Background(), inv.ID)
	if err != nil {
		t.Fatalf("GetPDF returned error: %v", err)
	}
	if !bytes.HasPrefix(data, []byte("%PDF-")) {
		t.Fatal("expected valid PDF bytes")
	}
	if filename == "" {
		t.Fatal("expected a non-empty filename")
	}
}

func TestGetPDF_NotFoundReturns404(t *testing.T) {
	db := newTestDB(t)
	svc := NewInvoiceService(db, nil)

	_, _, err := svc.GetPDF(context.Background(), uuid.New())
	assertAPIError(t, err, 404)
}

func TestGetPDFPresignedURL_404WhenNotArchived(t *testing.T) {
	db := newTestDB(t)
	svc := NewInvoiceService(db, nil)

	inv := models.Invoice{ID: uuid.New(), ClienteID: "c1", Stato: "PROFORMA", CostiAccessori: []byte("[]")}
	if err := db.Create(&inv).Error; err != nil {
		t.Fatal(err)
	}

	_, err := svc.GetPDFPresignedURL(context.Background(), inv.ID)
	assertAPIError(t, err, 404)
}

func TestGetPDFPresignedURL_404WhenS3Disabled(t *testing.T) {
	db := newTestDB(t)
	svc := NewInvoiceService(db, nil)

	key := "invoices/2026/x.pdf"
	inv := models.Invoice{ID: uuid.New(), ClienteID: "c1", Stato: "DEFINITIVA", CostiAccessori: []byte("[]"), PdfS3Key: &key}
	if err := db.Create(&inv).Error; err != nil {
		t.Fatal(err)
	}

	_, err := svc.GetPDFPresignedURL(context.Background(), inv.ID)
	assertAPIError(t, err, 404)
}

func TestFinalize_PdfArchivedFalseWhenS3Disabled(t *testing.T) {
	db := newTestDB(t)
	svc := NewInvoiceService(db, nil)

	inv := models.Invoice{ID: uuid.New(), ClienteID: "c1", Stato: "PROFORMA", CostiAccessori: []byte("[]")}
	if err := db.Create(&inv).Error; err != nil {
		t.Fatal(err)
	}

	result, err := svc.Finalize(context.Background(), inv.ID)
	if err != nil {
		t.Fatalf("Finalize returned error: %v", err)
	}
	if !result.OK || result.PdfArchived || result.PdfS3Key != nil {
		t.Fatalf("expected pdf_archived=false with nil S3 client, got %+v", result)
	}
}
