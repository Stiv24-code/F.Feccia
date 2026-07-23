package trips

import (
	"bytes"
	"context"
	"testing"

	"github.com/google/uuid"

	"fratelli-feccia/internal/models"
	"fratelli-feccia/pkg/utils"
)

func TestGetInstructionsPDF_ResolvesOrdersCustomersAndDriver(t *testing.T) {
	db := newTestDB(t)
	svc := NewTripService(db)

	driver := models.Driver{ID: uuid.New(), Nome: "Mario", Cognome: "Rossi", Telefono: "3331234567", Active: true}
	if err := db.Create(&driver).Error; err != nil {
		t.Fatal(err)
	}
	customer := models.Customer{ID: uuid.New(), RagioneSociale: "Cliente Uno", Telefono: "0212345678", Active: true}
	if err := db.Create(&customer).Error; err != nil {
		t.Fatal(err)
	}

	trip := models.Trip{
		ID: uuid.New(), TargaMotrice: "AB123CD", AutistaID: &driver.ID,
		DataPartenza: "2026-01-10", KmTotali: 200,
		Segments: []models.TripSegment{{Ordine: 1, Tipo: "base_carico", Km: 40, TempoStimatoMin: 30}},
	}
	if err := db.Create(&trip).Error; err != nil {
		t.Fatal(err)
	}
	caricoID := seedDestination(t, db, "Milano (MI)", 45.0, 9.0)
	scaricoID := seedDestination(t, db, "Lodi (LO)", 45.5, 9.5)
	order := models.Order{
		ID: uuid.New(), ClienteID: customer.ID,
		DestinazioneCaricoID: &caricoID, DestinazioneScaricoID: &scaricoID,
		DataRitiro: "2026-01-10", Stato: "VIAGGIO", ViaggioID: &trip.ID,
		ServiziAccessori: []byte("[]"), CostiAccessori: []byte("[]"),
	}
	if err := db.Create(&order).Error; err != nil {
		t.Fatal(err)
	}

	data, filename, err := svc.GetInstructionsPDF(context.Background(), trip.ID)
	if err != nil {
		t.Fatalf("GetInstructionsPDF returned error: %v", err)
	}
	if !bytes.HasPrefix(data, []byte("%PDF-")) {
		t.Fatal("expected valid PDF bytes")
	}
	if filename != "ISTRUZIONI_20260110_"+trip.ID.String()[:8]+".pdf" {
		t.Fatalf("unexpected filename: %q", filename)
	}
}

func TestGetInstructionsPDF_NotFoundReturns404(t *testing.T) {
	db := newTestDB(t)
	svc := NewTripService(db)

	_, _, err := svc.GetInstructionsPDF(context.Background(), uuid.New())
	apiErr, ok := err.(utils.APIError)
	if !ok || apiErr.StatusCode() != 404 {
		t.Fatalf("expected a 404 APIError, got %v (%T)", err, err)
	}
}

func TestGetInstructionsPDF_HandlesTripWithNoOrdersOrDriver(t *testing.T) {
	db := newTestDB(t)
	svc := NewTripService(db)

	trip := models.Trip{ID: uuid.New(), DataPartenza: "2026-01-10"}
	if err := db.Create(&trip).Error; err != nil {
		t.Fatal(err)
	}

	data, _, err := svc.GetInstructionsPDF(context.Background(), trip.ID)
	if err != nil {
		t.Fatalf("GetInstructionsPDF returned error: %v", err)
	}
	if !bytes.HasPrefix(data, []byte("%PDF-")) {
		t.Fatal("expected valid PDF bytes even with no linked orders/driver")
	}
}
