package orders

import (
	"bytes"
	"context"
	"testing"

	"github.com/google/uuid"

	"fratelli-feccia/internal/models"
	"fratelli-feccia/pkg/utils"
)

func TestGetCMRPDF_ResolvesCustomerAndVehicle(t *testing.T) {
	db := newTestDB(t)
	svc := NewOrderService(db)

	customer := models.Customer{ID: uuid.New(), RagioneSociale: "Cliente Reale Srl", Active: true}
	if err := db.Create(&customer).Error; err != nil {
		t.Fatalf("failed to seed customer: %v", err)
	}
	vehicle := models.Vehicle{ID: uuid.New(), Targa: "AB123CD", Marca: "Iveco", Active: true}
	if err := db.Create(&vehicle).Error; err != nil {
		t.Fatalf("failed to seed vehicle: %v", err)
	}

	order := models.Order{
		ID: uuid.New(), ClienteID: customer.ID.String(), ClienteNome: customer.RagioneSociale,
		DestinazioneCaricoNome: "Milano (MI)", DestinazioneScaricoNome: "Lodi (LO)",
		DataRitiro: "2026-01-10", TargaMotrice: "AB123CD", Progressivo: "26/0001",
		ServiziAccessori: []byte("[]"), CostiAccessori: []byte("[]"),
	}
	if err := db.Create(&order).Error; err != nil {
		t.Fatalf("failed to seed order: %v", err)
	}

	data, filename, err := svc.GetCMRPDF(context.Background(), order.ID)
	if err != nil {
		t.Fatalf("GetCMRPDF returned error: %v", err)
	}
	if !bytes.HasPrefix(data, []byte("%PDF-")) {
		t.Fatal("expected valid PDF bytes")
	}
	if filename != "CMR_20260110_26-0001.pdf" {
		t.Fatalf("unexpected filename: %q", filename)
	}
}

func TestGetCMRPDF_FallsBackWhenCustomerMissing(t *testing.T) {
	db := newTestDB(t)
	svc := NewOrderService(db)

	order := models.Order{
		ID: uuid.New(), ClienteID: "nonexistent-customer-id", ClienteNome: "Cliente Fantasma",
		DestinazioneCaricoNome: "Milano (MI)", DestinazioneScaricoNome: "Lodi (LO)",
		DataRitiro: "2026-01-10", ServiziAccessori: []byte("[]"), CostiAccessori: []byte("[]"),
	}
	if err := db.Create(&order).Error; err != nil {
		t.Fatalf("failed to seed order: %v", err)
	}

	data, _, err := svc.GetCMRPDF(context.Background(), order.ID)
	if err != nil {
		t.Fatalf("expected fallback to cliente_nome, not an error: %v", err)
	}
	if !bytes.HasPrefix(data, []byte("%PDF-")) {
		t.Fatal("expected valid PDF bytes even with an unresolvable customer")
	}
}

func TestGetCMRPDF_NotFoundReturns404(t *testing.T) {
	db := newTestDB(t)
	svc := NewOrderService(db)

	_, _, err := svc.GetCMRPDF(context.Background(), uuid.New())
	apiErr, ok := err.(utils.APIError)
	if !ok || apiErr.StatusCode() != 404 {
		t.Fatalf("expected a 404 APIError, got %v (%T)", err, err)
	}
}
