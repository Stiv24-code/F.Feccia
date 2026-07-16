package pdfgen

import (
	"bytes"
	"testing"
	"time"

	"fratelli-feccia/internal/models"
)

func sampleInvoice() models.Invoice {
	return models.Invoice{
		Numero: "O/F-26/0001", DataFattura: "2026-03-01", DataScadenza: "2026-04-01",
		CondizioniPagamento: "30gg df fm", Stato: "PROFORMA", Note: "Pagamento gradito entro scadenza",
		Righe: []models.InvoiceLine{
			{Descrizione: "Trasporto Milano-Lodi", Prodotto: "P001", Peso: 500, Quantita: 1, Tariffa: 250, Totale: 250, IvaCodice: "N8"},
		},
		TotaleImponibile: 250, TotaleIva: 0, Totale: 250,
		CreatedAt: time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
	}
}

func TestBuildInvoicePDF_ProducesValidPDF(t *testing.T) {
	customer := models.Customer{
		RagioneSociale: "Cliente Fatturato Srl", Indirizzo: "Via Roma 1",
		Citta: "Milano", Cap: "20100", Provincia: "MI", Nazione: "ITALIA",
		PartitaIva: "IT12345678901", Pec: "cliente@pec.it",
	}
	data, err := BuildInvoicePDF(sampleInvoice(), customer, nil)
	if err != nil {
		t.Fatalf("BuildInvoicePDF returned error: %v", err)
	}
	if !bytes.HasPrefix(data, []byte("%PDF-")) {
		t.Fatal("expected valid PDF bytes")
	}
}

func TestBuildInvoicePDF_DefinitivaShowsFatturaLabel(t *testing.T) {
	invoice := sampleInvoice()
	invoice.Stato = "DEFINITIVA"
	data, err := BuildInvoicePDF(invoice, models.Customer{RagioneSociale: "X"}, nil)
	if err != nil {
		t.Fatalf("BuildInvoicePDF returned error: %v", err)
	}
	if !bytes.HasPrefix(data, []byte("%PDF-")) {
		t.Fatal("expected valid PDF bytes")
	}
}

func TestBuildInvoicePDF_HandlesEmptyLines(t *testing.T) {
	invoice := sampleInvoice()
	invoice.Righe = nil
	data, err := BuildInvoicePDF(invoice, models.Customer{}, nil)
	if err != nil {
		t.Fatalf("BuildInvoicePDF returned error: %v", err)
	}
	if !bytes.HasPrefix(data, []byte("%PDF-")) {
		t.Fatal("expected valid PDF bytes even with no line items")
	}
}

func TestMakeInvoiceFilename(t *testing.T) {
	invoice := models.Invoice{Numero: "O/F-26/0007", DataFattura: "2026-03-15"}
	customer := models.Customer{RagioneSociale: "Cliente Test S.r.l."}
	got := MakeInvoiceFilename(invoice, customer)
	want := "FATT_2026_O-F-26-0007_cliente_test_s_r_l.pdf"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestMakeInvoiceFilename_FallsBackToDraftAndCliente(t *testing.T) {
	invoice := models.Invoice{CreatedAt: time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC)}
	customer := models.Customer{}
	got := MakeInvoiceFilename(invoice, customer)
	want := "FATT_2027_DRAFT_cliente.pdf"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}
