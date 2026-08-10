package pdfengine

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/go-pdf/fpdf"
	"github.com/google/uuid"

	"fratelli-feccia/internal/dto"
)

func TestCleanNumber(t *testing.T) {
	cases := []struct{ in, expected string }{
		{"28.000 kg", "28000"},
		{"28,5", "28"},
		{"1'250", "1250"},
		{" 30 000 ", "30000"},
		{"kg 28000", "28000"},
		{"", ""},
		{"n/d", ""},
	}
	for _, tc := range cases {
		if got := cleanNumber(tc.in); got != tc.expected {
			t.Errorf("cleanNumber(%q): expected %q, got %q", tc.in, tc.expected, got)
		}
	}
}

func TestZoneText_WordCenterInclusion(t *testing.T) {
	// One page 100x100pt with two words: "inside" centered at (25,25) and
	// "outside" centered at (75,75).
	page := bboxPage{
		Width:  100,
		Height: 100,
		Blocks: []bboxBlock{{
			XMin: 20, YMin: 20, XMax: 80, YMax: 80,
			Lines: []bboxLine{{Words: []bboxWord{
				{XMin: 20, YMin: 20, XMax: 30, YMax: 30, Text: "inside"},
				{XMin: 70, YMin: 70, XMax: 80, YMax: 80, Text: "outside"},
			}}},
		}},
	}
	// Zone covering the top-left quadrant only.
	f := dto.PdfTemplateFieldDTO{X: 0, Y: 0, W: 0.5, H: 0.5}
	if got := zoneText(page, f); got != "inside" {
		t.Fatalf("expected %q, got %q", "inside", got)
	}
	// Full-page zone captures both, in reading order.
	f = dto.PdfTemplateFieldDTO{X: 0, Y: 0, W: 1, H: 1}
	if got := zoneText(page, f); got != "inside outside" {
		t.Fatalf("expected both words, got %q", got)
	}
}

func TestBuildDraft_MapsTargetsAndFallbacks(t *testing.T) {
	svc := NewPdfEngineService("")
	tpl := dto.PdfTemplateResponse{
		ID:      uuid.New(),
		Name:    "ACME layout",
		Client:  "ACME S.r.l.",
		Senders: []string{"ordini@acme.it"},
		Fields: []dto.PdfTemplateFieldDTO{
			{ID: "f1", Target: "ref"},
			{ID: "f2", Target: "kg"},
			{ID: "f3", Target: "notes"},
			{ID: "f4", Target: "notes"}, // two zones onto the same target
		},
	}
	values := map[string]dto.PdfExtractedValueDTO{
		"f1": {Value: " ORD-42 ", Method: "poppler-text"},
		"f2": {Value: "28.000 kg", Method: "poppler-text"},
		"f3": {Value: "prima", Method: "poppler-text"},
		"f4": {Value: "seconda", Method: "poppler-text"},
	}

	draft, byTarget := svc.BuildDraft(tpl, values, "", "ordine.pdf")

	if draft.Ref != "ORD-42" {
		t.Fatalf("expected ref ORD-42, got %q", draft.Ref)
	}
	if draft.Kg != 28000 {
		t.Fatalf("expected kg 28000, got %d", draft.Kg)
	}
	if draft.Notes != "prima seconda" {
		t.Fatalf("expected joined notes, got %q", draft.Notes)
	}
	// Client falls back to the template client, sender to the first exact
	// template sender.
	if draft.Client != "ACME S.r.l." {
		t.Fatalf("expected template client fallback, got %q", draft.Client)
	}
	if draft.SenderEmail != "ordini@acme.it" {
		t.Fatalf("expected sender fallback from template, got %q", draft.SenderEmail)
	}
	if draft.Status != "pending" || draft.Source != "pdf" {
		t.Fatalf("expected pending/pdf, got %q/%q", draft.Status, draft.Source)
	}
	if draft.TemplateID == nil || *draft.TemplateID != tpl.ID {
		t.Fatalf("expected template id on draft, got %v", draft.TemplateID)
	}
	if byTarget["notes"] != "prima seconda" {
		t.Fatalf("expected byTarget to carry merged values, got %v", byTarget)
	}
}

func TestBuildDraft_RefFallsBackToFilename(t *testing.T) {
	svc := NewPdfEngineService("")
	tpl := dto.PdfTemplateResponse{ID: uuid.New(), Name: "X"}
	draft, _ := svc.BuildDraft(tpl, nil, "Chi@Manda.it", "ordine-luglio.pdf")
	if draft.Ref != "ordine-luglio" {
		t.Fatalf("expected filename fallback, got %q", draft.Ref)
	}
	if draft.SenderEmail != "chi@manda.it" {
		t.Fatalf("expected lowercased sender, got %q", draft.SenderEmail)
	}
	if draft.Client != "X" {
		t.Fatalf("expected template name as last client fallback, got %q", draft.Client)
	}
}

// makeTestPDF builds a one-page A4 PDF with known text at a known position
// (in points), so zones can be expressed in normalized coordinates.
func makeTestPDF(t *testing.T) []byte {
	t.Helper()
	pdf := fpdf.New("P", "pt", "A4", "") // 595.28 x 841.89 pt
	pdf.AddPage()
	pdf.SetFont("Helvetica", "", 14)
	pdf.Text(100, 100, "ORD-42")
	pdf.Text(100, 200, "28.000")
	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		t.Fatalf("failed to build test PDF: %v", err)
	}
	return buf.Bytes()
}

// End-to-end through real poppler binaries; skipped when they are not in
// PATH (they are always present in the Docker image).
func TestRenderAndExtract_WithPoppler(t *testing.T) {
	svc := NewPdfEngineService("")
	if !svc.Ready() {
		t.Skip("poppler-utils (pdftoppm/pdftotext) not in PATH")
	}
	ctx := context.Background()
	pdfBytes := makeTestPDF(t)

	res, err := svc.Render(ctx, "test.pdf", pdfBytes)
	if err != nil {
		t.Fatalf("Render returned error: %v", err)
	}
	if res.PageCount != 1 || len(res.Pages) != 1 {
		t.Fatalf("expected 1 page, got count=%d pages=%d", res.PageCount, len(res.Pages))
	}
	if res.Pages[0].ImageB64 == "" || res.Pages[0].Width == 0 {
		t.Fatalf("expected rendered PNG with size, got width=%d", res.Pages[0].Width)
	}
	foundBlock := false
	for _, b := range res.Pages[0].Blocks {
		if strings.Contains(b.Text, "ORD-42") {
			foundBlock = true
		}
	}
	if !foundBlock {
		t.Fatalf("expected a text block containing ORD-42, got %+v", res.Pages[0].Blocks)
	}

	// A4 in points; the zone around Text(100,100) — baseline at y=100, so
	// the glyph box sits roughly in y 85..105.
	const pw, ph = 595.28, 841.89
	tpl := dto.PdfTemplateResponse{
		ID:     uuid.New(),
		Name:   "test",
		Client: "ACME",
		Fields: []dto.PdfTemplateFieldDTO{
			{ID: "zref", Target: "ref", Page: 0, X: 80 / pw, Y: 75 / ph, W: 120 / pw, H: 40 / ph},
			{ID: "zkg", Target: "kg", Page: 0, X: 80 / pw, Y: 175 / ph, W: 120 / pw, H: 40 / ph},
			{ID: "zempty", Target: "notes", Page: 0, X: 0.7, Y: 0.7, W: 0.1, H: 0.05},
		},
	}
	values, err := svc.Extract(ctx, pdfBytes, tpl)
	if err != nil {
		t.Fatalf("Extract returned error: %v", err)
	}
	if v := values["zref"]; v.Value != "ORD-42" || v.Method != "poppler-text" {
		t.Fatalf("expected ORD-42 via poppler-text, got %+v", v)
	}
	// Empty zone with no API key -> method "empty", not an error.
	if v := values["zempty"]; v.Method != "empty" || v.Value != "" {
		t.Fatalf("expected empty method for blank zone, got %+v", v)
	}

	draft, _ := svc.BuildDraft(tpl, values, "ordini@acme.it", "test.pdf")
	if draft.Ref != "ORD-42" || draft.Kg != 28000 {
		t.Fatalf("expected draft ORD-42/28000, got %q/%d", draft.Ref, draft.Kg)
	}
}
