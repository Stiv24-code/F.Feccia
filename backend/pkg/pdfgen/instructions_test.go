package pdfgen

import (
	"bytes"
	"testing"

	"github.com/google/uuid"

	"fratelli-feccia/internal/models"
)

func TestBuildInstructionsPDF_ProducesValidPDF(t *testing.T) {
	trip := models.Trip{
		ID: uuid.New(), TargaMotrice: "AB123CD", TargaRimorchio: "TR001AA",
		AutistaNome: "Mario Rossi", GarageNome: "Garage Lodi",
		DataPartenza: "2026-01-10", DataArrivo: "2026-01-11", KmTotali: 350.5,
	}
	orders := []models.Order{
		{
			ClienteID: "c1", ClienteNome: "Cliente Uno", RifOrdineCliente: "RIF-1",
			DestinazioneCaricoNome: "Milano (MI)", DestinazioneScaricoNome: "Lodi (LO)",
			DataRitiro: "2026-01-10", DataConsegna: "2026-01-10", Note: "Fragile",
			Items: []models.OrderItem{{ProdottoDescrizione: "Pallet", Quantita: 2, Peso: 400}},
		},
	}
	segments := []models.TripSegment{
		{Ordine: 1, Tipo: "base_carico", OrigineNome: "Garage Lodi", DestinazioneNome: "Milano (MI)", Km: 40, TempoStimatoMin: 50},
	}
	driver := &models.Driver{Nome: "Mario", Cognome: "Rossi", Telefono: "3331234567"}
	customersByID := map[string]models.Customer{"c1": {Telefono: "0212345678"}}

	data, err := BuildInstructionsPDF(trip, orders, segments, driver, customersByID)
	if err != nil {
		t.Fatalf("BuildInstructionsPDF returned error: %v", err)
	}
	if !bytes.HasPrefix(data, []byte("%PDF-")) {
		t.Fatal("expected valid PDF bytes")
	}
}

func TestBuildInstructionsPDF_HandlesEmptyOrdersAndSegments(t *testing.T) {
	trip := models.Trip{ID: uuid.New()}
	data, err := BuildInstructionsPDF(trip, nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("BuildInstructionsPDF returned error: %v", err)
	}
	if !bytes.HasPrefix(data, []byte("%PDF-")) {
		t.Fatal("expected valid PDF bytes even with no orders/segments/driver")
	}
}

func TestMakeInstructionsFilename(t *testing.T) {
	trip := models.Trip{ID: uuid.MustParse("12345678-1234-1234-1234-123456789012"), DataPartenza: "2026-04-01"}
	got := MakeInstructionsFilename(trip)
	want := "ISTRUZIONI_20260401_12345678.pdf"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}
