package pdfgen

import (
	"bytes"
	"testing"

	"github.com/google/uuid"

	"fratelli-feccia/internal/models"
)

func TestBuildCMRPDF_ProducesValidPDF(t *testing.T) {
	order := models.Order{
		ID: uuid.New(), Progressivo: "26/0001", Cliente: models.Customer{RagioneSociale: "Cliente Test"},
		DestinazioneCarico: &models.Destination{Nome: "Milano (MI)"}, DestinazioneScarico: &models.Destination{Nome: "Lodi (LO)"},
		DataRitiro: "2026-01-10", Note: "Consegna urgente",
		Items: []models.OrderItem{{Prodotto: models.Product{Descrizione: "Pallet EPAL"}, Quantita: 2, Peso: 500}},
	}
	consignee := models.Customer{
		RagioneSociale: "Destinatario Srl", Indirizzo: "Via Roma 1",
		Citta: "Milano", Cap: "20100", Provincia: "MI", Nazione: "ITALIA", PartitaIva: "IT12345678901",
	}
	motrice := &models.Motrice{Targa: "AB123CD", Marca: "Iveco", Modello: "Daily"}

	data, err := BuildCMRPDF(order, consignee, nil, motrice, nil)
	if err != nil {
		t.Fatalf("BuildCMRPDF returned error: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("expected non-empty PDF bytes")
	}
	if !bytes.HasPrefix(data, []byte("%PDF-")) {
		t.Fatalf("expected output to start with %%PDF- header, got: %q", data[:min(20, len(data))])
	}
	if !bytes.Contains(data, []byte("%%EOF")) {
		t.Fatal("expected output to contain an EOF trailer marker")
	}
}

func TestBuildCMRPDF_HandlesNoItemsAndNoVehicle(t *testing.T) {
	order := models.Order{
		ID: uuid.New(), Cliente: models.Customer{RagioneSociale: "Cliente Senza Items"},
		DestinazioneCarico: &models.Destination{Nome: "Cuneo (CN)"}, DestinazioneScarico: &models.Destination{Nome: "Milano (MI)"},
		DataRitiro: "2026-02-01",
	}
	consignee := models.Customer{RagioneSociale: "Dest Srl"}

	data, err := BuildCMRPDF(order, consignee, nil, nil, nil)
	if err != nil {
		t.Fatalf("BuildCMRPDF returned error: %v", err)
	}
	if !bytes.HasPrefix(data, []byte("%PDF-")) {
		t.Fatal("expected a valid PDF even with no items/vehicle")
	}
}

func TestMakeCMRFilename(t *testing.T) {
	order := models.Order{Progressivo: "26/0007", DataRitiro: "2026-03-15"}
	got := MakeCMRFilename(order)
	want := "CMR_20260315_26-0007.pdf"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}
