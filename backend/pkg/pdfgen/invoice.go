package pdfgen

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/go-pdf/fpdf"

	"fratelli-feccia/internal/models"
)

// Issuer mirrors invoice_pdf.py's ISSUER_DEFAULT (FECCIA F.lli).
type Issuer struct {
	RagioneSociale string
	Indirizzo      string
	CapCitta       string
	PartitaIva     string
	CodiceFiscale  string
	Iban           string
	Telefono       string
	Email          string
	Pec            string
}

var DefaultIssuer = Issuer{
	RagioneSociale: "FECCIA F.lli S.r.l.",
	Indirizzo:      "Via Lodi 12",
	CapCitta:       "26900 Lodi (LO)",
	PartitaIva:     "IT00000000000",
	CodiceFiscale:  "00000000000",
	Iban:           "IT00 X000 0000 0000 0000 0000 000",
	Telefono:       "+39 0371 000000",
	Email:          "amministrazione@feccia.it",
	Pec:            "feccia@pec.it",
}

// BuildInvoicePDF mirrors build_invoice_pdf. Not XML SDI e-invoicing format —
// electronic invoicing to SDI is handled separately by the fiscal provider
// (see finalize_invoice's own comment in Python); this is a courtesy summary.
func BuildInvoicePDF(invoice models.Invoice, customer models.Customer, issuer *Issuer) ([]byte, error) {
	iss := DefaultIssuer
	if issuer != nil {
		iss = *issuer
	}

	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 18, 15)
	pdf.SetAutoPageBreak(true, 20)
	pdf.AliasNbPages("")

	pdf.SetHeaderFunc(func() { drawInvoiceHeader(pdf, iss, invoice) })
	pdf.SetFooterFunc(func() {
		pdf.SetY(-15)
		pdf.SetFont("Helvetica", "I", 8)
		pdf.SetTextColor(120, 120, 120)
		pdf.CellFormat(0, 5, safe(fmt.Sprintf("Pagina %d/{nb} - IBAN %s - %s", pdf.PageNo(), iss.Iban, iss.Email)),
			"", 0, "C", false, 0, "")
	})

	pdf.AddPage()

	drawCustomerBlock(pdf, customer)
	drawLinesTable(pdf, invoice.Righe)
	drawTotals(pdf, invoice)
	drawPaymentBlock(pdf, invoice, iss)
	drawFEFooter(pdf)

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func drawInvoiceHeader(pdf *fpdf.Fpdf, issuer Issuer, invoice models.Invoice) {
	pdf.SetFont("Helvetica", "B", 14)
	pdf.CellFormat(0, 7, safe(issuer.RagioneSociale), "", 1, "L", false, 0, "")
	pdf.SetFont("Helvetica", "", 9)
	pdf.CellFormat(0, 4, safe(issuer.Indirizzo), "", 1, "L", false, 0, "")
	pdf.CellFormat(0, 4, safe(issuer.CapCitta), "", 1, "L", false, 0, "")
	pdf.CellFormat(0, 4, safe(fmt.Sprintf("P.IVA %s - C.F. %s", issuer.PartitaIva, issuer.CodiceFiscale)), "", 1, "L", false, 0, "")

	tipo := strings.ToUpper(invoice.Stato)
	numero := invoice.Numero
	if numero == "" {
		numero = "-"
	}
	data := fmtDate(invoice.DataFattura)
	pdf.SetXY(140, 18)
	pdf.SetFont("Helvetica", "B", 12)
	label := "PROFORMA"
	if tipo == "DEFINITIVA" {
		label = "FATTURA"
	}
	pdf.CellFormat(55, 6, safe(fmt.Sprintf("%s %s", label, numero)), "", 1, "R", false, 0, "")
	pdf.SetX(140)
	pdf.SetFont("Helvetica", "", 9)
	pdf.CellFormat(55, 4, "Data: "+data, "", 1, "R", false, 0, "")

	pdf.Ln(2)
	pdf.SetDrawColor(180, 180, 180)
	pdf.Line(15, pdf.GetY()+1, 195, pdf.GetY()+1)
	pdf.Ln(4)
}

func drawCustomerBlock(pdf *fpdf.Fpdf, customer models.Customer) {
	pdf.SetFont("Helvetica", "B", 9)
	pdf.CellFormat(0, 5, "DESTINATARIO", "", 1, "L", false, 0, "")
	pdf.SetFont("Helvetica", "B", 11)
	pdf.CellFormat(0, 6, safe(orDash(customer.RagioneSociale)), "", 1, "L", false, 0, "")
	pdf.SetFont("Helvetica", "", 9)

	indirizzo := strings.TrimSpace(customer.Indirizzo)
	cittaRiga := strings.TrimSpace(strings.Join(filterEmpty([]string{
		customer.Cap, customer.Citta, provinciaParens(customer.Provincia),
	}), " "))

	if indirizzo != "" {
		pdf.CellFormat(0, 4, safe(indirizzo), "", 1, "L", false, 0, "")
	}
	if cittaRiga != "" {
		pdf.CellFormat(0, 4, safe(cittaRiga), "", 1, "L", false, 0, "")
	}
	if customer.Nazione != "" {
		pdf.CellFormat(0, 4, safe(customer.Nazione), "", 1, "L", false, 0, "")
	}
	if customer.PartitaIva != "" {
		pdf.CellFormat(0, 4, safe("P.IVA: "+customer.PartitaIva), "", 1, "L", false, 0, "")
	}
	if customer.CodiceFiscale != "" {
		pdf.CellFormat(0, 4, safe("C.F.: "+customer.CodiceFiscale), "", 1, "L", false, 0, "")
	}
	if customer.Pec != "" {
		pdf.CellFormat(0, 4, safe("PEC: "+customer.Pec), "", 1, "L", false, 0, "")
	}
	pdf.Ln(3)
}

func drawLinesTable(pdf *fpdf.Fpdf, righe []models.InvoiceLine) {
	headers := []string{"Descrizione", "Prodotto", "Peso", "Qta", "Tariffa", "IVA", "Totale"}
	widths := []float64{55, 30, 18, 12, 23, 12, 30}

	pdf.SetFillColor(235, 238, 242)
	pdf.SetFont("Helvetica", "B", 8)
	for i, h := range headers {
		align := "L"
		if h == "Peso" || h == "Qta" || h == "Tariffa" || h == "Totale" {
			align = "R"
		}
		pdf.CellFormat(widths[i], 7, h, "1", 0, align, true, 0, "")
	}
	pdf.Ln(-1)

	pdf.SetFont("Helvetica", "", 8)
	for _, r := range righe {
		descrizione := r.Descrizione
		if len(descrizione) > 60 {
			descrizione = descrizione[:60]
		}
		prodotto := r.Prodotto
		if len(prodotto) > 30 {
			prodotto = prodotto[:30]
		}
		pesoTxt := ""
		if r.Peso != 0 {
			pesoTxt = fmtG(r.Peso) + " kg"
		}
		qtaTxt := ""
		if r.Quantita != 0 {
			qtaTxt = fmtG(r.Quantita)
		}
		ivaCodice := r.IvaCodice
		if ivaCodice == "" {
			ivaCodice = "N8"
		}

		type cell struct {
			w     float64
			txt   string
			align string
		}
		cells := []cell{
			{widths[0], safe(descrizione), "L"}, {widths[1], safe(prodotto), "L"},
			{widths[2], pesoTxt, "R"}, {widths[3], qtaTxt, "R"},
			{widths[4], fmtEuro(r.Tariffa), "R"}, {widths[5], safe(ivaCodice), "C"},
			{widths[6], fmtEuro(r.Totale), "R"},
		}
		for _, c := range cells {
			pdf.CellFormat(c.w, 6, c.txt, "1", 0, c.align, false, 0, "")
		}
		pdf.Ln(-1)
	}
}

func drawTotals(pdf *fpdf.Fpdf, invoice models.Invoice) {
	pdf.Ln(2)

	pdf.SetX(115)
	pdf.SetFont("Helvetica", "", 10)
	pdf.CellFormat(50, 6, "Totale imponibile", "", 0, "R", false, 0, "")
	pdf.CellFormat(30, 6, fmtEuro(invoice.TotaleImponibile), "", 1, "R", false, 0, "")

	pdf.SetX(115)
	pdf.CellFormat(50, 6, "IVA", "", 0, "R", false, 0, "")
	pdf.CellFormat(30, 6, fmtEuro(invoice.TotaleIva), "", 1, "R", false, 0, "")

	pdf.SetX(115)
	pdf.SetFont("Helvetica", "B", 11)
	pdf.SetFillColor(11, 18, 32)
	pdf.SetTextColor(255, 255, 255)
	pdf.CellFormat(50, 8, "TOTALE DOCUMENTO", "", 0, "R", true, 0, "")
	pdf.CellFormat(30, 8, fmtEuro(invoice.Totale), "", 1, "R", true, 0, "")
	pdf.SetTextColor(0, 0, 0)
}

func drawPaymentBlock(pdf *fpdf.Fpdf, invoice models.Invoice, issuer Issuer) {
	pdf.Ln(6)
	pdf.SetFont("Helvetica", "B", 9)
	pdf.CellFormat(0, 5, "MODALITA DI PAGAMENTO", "", 1, "L", false, 0, "")
	pdf.SetFont("Helvetica", "", 9)
	pdf.CellFormat(0, 4, safe("Bonifico bancario su IBAN: "+issuer.Iban), "", 1, "L", false, 0, "")
	if invoice.CondizioniPagamento != "" {
		pdf.CellFormat(0, 4, safe("Condizioni: "+invoice.CondizioniPagamento), "", 1, "L", false, 0, "")
	}
	if invoice.DataScadenza != "" {
		pdf.CellFormat(0, 4, safe("Scadenza: "+fmtDate(invoice.DataScadenza)), "", 1, "L", false, 0, "")
	}
	if invoice.Note != "" {
		pdf.Ln(3)
		pdf.SetFont("Helvetica", "B", 9)
		pdf.CellFormat(0, 5, "NOTE", "", 1, "L", false, 0, "")
		pdf.SetFont("Helvetica", "", 9)
		pdf.MultiCell(0, 4, safe(invoice.Note), "", "L", false)
	}
}

func drawFEFooter(pdf *fpdf.Fpdf) {
	pdf.Ln(8)
	pdf.SetFont("Helvetica", "I", 8)
	pdf.SetTextColor(120, 120, 120)
	pdf.MultiCell(0, 4, safe(
		"Documento generato dal sistema TMS LoginBusiness. La fatturazione elettronica "+
			"verso lo SDI in formato FatturaPA XML viene gestita separatamente dal provider "+
			"fiscale convenzionato; questo PDF e' solo un riepilogo cortese per il cliente."),
		"", "L", false)
	pdf.SetTextColor(0, 0, 0)
}

// MakeInvoiceFilename mirrors make_filename: `FATT_{anno}_{numero}_{slug}.pdf`.
func MakeInvoiceFilename(invoice models.Invoice, customer models.Customer) string {
	anno := invoice.DataFattura
	if len(anno) >= 4 {
		anno = anno[:4]
	} else {
		anno = ""
	}
	if anno == "" {
		anno = fmt.Sprintf("%d", invoice.CreatedAt.Year())
	}
	numero := invoice.Numero
	if numero == "" {
		numero = "DRAFT"
	}
	numero = strings.ReplaceAll(numero, "/", "-")

	slugSrc := strings.ToLower(customer.RagioneSociale)
	if slugSrc == "" {
		slugSrc = "cliente"
	}
	var slugBuf strings.Builder
	for _, r := range slugSrc {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			slugBuf.WriteRune(r)
		} else {
			slugBuf.WriteByte('_')
		}
	}
	slug := slugBuf.String()
	if len(slug) > 40 {
		slug = slug[:40]
	}
	slug = strings.Trim(slug, "_")

	return fmt.Sprintf("FATT_%s_%s_%s.pdf", anno, numero, slug)
}
