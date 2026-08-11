package mailscraper

import (
	"testing"

	"fratelli-feccia/internal/dto"
)

func TestParseOrderBodyRoundTrip(t *testing.T) {
	src := dto.InboundOrderResponse{
		Client: "Xchance Italia Srl", SenderEmail: "office@xchance.it",
		Ref: "SBG26-0022", Product: "Latte crudo", Kg: 25750,
		LoadDate: "25/07 · 14:30", LoadPlace: "SCA Laitnaa 24, La Capelle (FR)",
		DeliveryDate: "26/07", DeliveryPlace: "Valform S.r.l., Martiniana Po (CN)",
		Rate: "€ 1.900", Notes: "Prodotto ≤ 4 °C", Portal: true,
	}
	got, ok := parseOrderBody(OrderMailBody(src), "fallback@feccia.it")
	if !ok {
		t.Fatal("parseOrderBody returned ok=false")
	}
	if got.Client != src.Client || got.SenderEmail != src.SenderEmail || got.Ref != src.Ref ||
		got.Product != src.Product || got.Kg != src.Kg ||
		got.LoadDate != src.LoadDate || got.LoadPlace != src.LoadPlace ||
		got.DeliveryDate != src.DeliveryDate || got.DeliveryPlace != src.DeliveryPlace ||
		got.Rate != src.Rate || got.Notes != src.Notes || !got.Portal {
		t.Errorf("round trip mismatch:\n got: %+v\nwant: %+v", got, src)
	}
	if got.Source != "mail" {
		t.Errorf("expected source mail, got %q", got.Source)
	}
}

func TestParseOrderBodyFallbackSender(t *testing.T) {
	body := "CLIENTE: Mazzoleni S.p.A.\nRIF: Viaggio 31/07\nPRODOTTO: Mangime sfuso\nKG: 28.000"
	got, ok := parseOrderBody(body, "marco.cattaneo@mazzoleni.com")
	if !ok {
		t.Fatal("parseOrderBody returned ok=false")
	}
	if got.SenderEmail != "marco.cattaneo@mazzoleni.com" {
		t.Errorf("sender fallback failed: %q", got.SenderEmail)
	}
	if got.Kg != 28000 {
		t.Errorf("kg with thousands separator: got %d, want 28000", got.Kg)
	}
}

func TestParseOrderBodyEnglishKeys(t *testing.T) {
	body := "CLIENT: ACME Ltd\nREF: PO-99\nPRODUCT: Molasses\nLOAD: 10/08 | Ravenna\nDELIVERY: 11/08 | Verona\nPORTAL: yes"
	got, ok := parseOrderBody(body, "x@acme.com")
	if !ok {
		t.Fatal("parseOrderBody returned ok=false")
	}
	if got.LoadDate != "10/08" || got.LoadPlace != "Ravenna" ||
		got.DeliveryDate != "11/08" || got.DeliveryPlace != "Verona" {
		t.Errorf("date|place split failed: %+v", got)
	}
	if !got.Portal {
		t.Errorf("expected portal true for 'yes'")
	}
}

func TestParseOrderBodyRejectsNonOrder(t *testing.T) {
	if _, ok := parseOrderBody("Ciao, ci vediamo domani alle 10:30", "x@y.it"); ok {
		t.Error("random mail parsed as order")
	}
}

func TestExtractTextBody_FallbackAfterHeaders(t *testing.T) {
	raw := []byte("Subject: [ORDINE] test\r\nFrom: a@b.it\r\n\r\nCLIENTE: ACME\r\nRIF: X")
	body := extractTextBody(raw)
	if body != "CLIENTE: ACME\r\nRIF: X" {
		t.Fatalf("unexpected body: %q", body)
	}
}
