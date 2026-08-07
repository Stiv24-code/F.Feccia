package pdfgen

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/go-pdf/fpdf"
	"github.com/google/uuid"

	"fratelli-feccia/internal/models"
)

// Sender mirrors cmr_pdf.py's SENDER_DEFAULT (FECCIA F.lli).
type Sender struct {
	RagioneSociale string
	Indirizzo      string
	CapCitta       string
	Nazione        string
	PartitaIva     string
	Telefono       string
}

var DefaultSender = Sender{
	RagioneSociale: "FECCIA F.lli S.r.l.",
	Indirizzo:      "Via Lodi 12",
	CapCitta:       "26900 Lodi (LO)",
	Nazione:        "ITALIA",
	PartitaIva:     "IT00000000000",
	Telefono:       "+39 0371 000000",
}

// BuildCMRPDF mirrors build_cmr_pdf: a "courtesy" rendering of the 24-box
// CMR (Convention de Marchandises par Route, UNECE 1956) international
// road transport waybill — not the official 4-copy carbon form.
func BuildCMRPDF(order models.Order, consignee models.Customer, sender *Sender, motrice *models.Motrice, semirimorchio *models.Semirimorchio) ([]byte, error) {
	snd := DefaultSender
	if sender != nil {
		snd = *sender
	}

	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(10, 10, 10)
	pdf.SetAutoPageBreak(true, 12)
	pdf.AliasNbPages("")

	pdf.SetHeaderFunc(func() {
		pdf.SetFont("Helvetica", "B", 12)
		pdf.CellFormat(0, 6, "LETTERA DI VETTURA INTERNAZIONALE - CMR", "", 1, "C", false, 0, "")
		pdf.SetFont("Helvetica", "", 8)
		pdf.CellFormat(0, 4, safe("Convenzione relativa al contratto di trasporto internazionale "+
			"di merci su strada (Ginevra 19/05/1956)"), "", 1, "C", false, 0, "")
		pdf.Ln(2)
	})
	pdf.SetFooterFunc(func() {
		pdf.SetY(-12)
		pdf.SetFont("Helvetica", "I", 7)
		pdf.SetTextColor(120, 120, 120)
		pdf.CellFormat(0, 4, safe(fmt.Sprintf(
			"CMR generata da TMS LoginBusiness - Pagina %d/{nb} - "+
				"Documento di cortesia, non sostituisce il modulo cartaceo a 4 copie", pdf.PageNo())),
			"", 0, "C", false, 0, "")
	})

	pdf.AddPage()

	deliveryPlace := strings.TrimSpace(strings.Join(filterEmpty([]string{
		consignee.Citta,
		provinciaParens(consignee.Provincia),
		consignee.Nazione,
	}), " "))
	if deliveryPlace == "" && order.DestinazioneScarico != nil {
		deliveryPlace = order.DestinazioneScarico.Nome
	}

	pickupPlace := snd.CapCitta
	if order.DestinazioneCarico != nil && order.DestinazioneCarico.Nome != "" {
		pickupPlace = order.DestinazioneCarico.Nome
	}
	pickupDate := fmtDate(order.DataRitiro)

	drawCMRFirstBlock(pdf, snd, consignee, deliveryPlace, pickupPlace, pickupDate)

	var totalWeight float64
	for _, it := range order.Items {
		totalWeight += it.Peso
	}
	drawCMRGoodsTable(pdf, order.Items, totalWeight)

	drawCMRBottomBlock(pdf, order, motrice, semirimorchio)

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func filterEmpty(vals []string) []string {
	out := make([]string, 0, len(vals))
	for _, v := range vals {
		if v != "" {
			out = append(out, v)
		}
	}
	return out
}

func provinciaParens(p string) string {
	if p == "" {
		return ""
	}
	return "(" + p + ")"
}

func cmrBox(pdf *fpdf.Fpdf, x, y, w, h float64, n int, label, content string, contentSize float64) {
	pdf.SetXY(x, y)
	pdf.SetDrawColor(0, 0, 0)
	pdf.Rect(x, y, w, h, "")

	pdf.SetXY(x+1, y+1)
	pdf.SetFont("Helvetica", "B", 7)
	pdf.CellFormat(5, 3, fmt.Sprintf("%d", n), "", 0, "L", false, 0, "")

	pdf.SetXY(x+6, y+1)
	pdf.SetFont("Helvetica", "", 7)
	labelText := label
	if len(labelText) > 60 {
		labelText = labelText[:60]
	}
	pdf.CellFormat(w-7, 3, safe(labelText), "", 0, "L", false, 0, "")

	if content != "" {
		pdf.SetXY(x+2, y+5)
		pdf.SetFont("Helvetica", "", contentSize)
		pdf.MultiCell(w-4, contentSize*0.4, safe(content), "", "L", false)
	}
}

func drawCMRFirstBlock(pdf *fpdf.Fpdf, sender Sender, consignee models.Customer, deliveryPlace, pickupPlace, pickupDate string) {
	senderText := fmt.Sprintf("%s\n%s\n%s\n%s\nP.IVA: %s",
		sender.RagioneSociale, sender.Indirizzo, sender.CapCitta, sender.Nazione, sender.PartitaIva)

	nazione := consignee.Nazione
	consigneeText := fmt.Sprintf("%s\n%s\n%s %s %s\n%s",
		orDash(consignee.RagioneSociale), consignee.Indirizzo,
		consignee.Cap, consignee.Citta, provinciaParens(consignee.Provincia), nazione)
	consigneeText = strings.TrimSpace(consigneeText)
	if consignee.PartitaIva != "" {
		consigneeText += "\nP.IVA: " + consignee.PartitaIva
	}

	cmrBox(pdf, 10, 22, 95, 30, 1, "Mittente (nome, indirizzo, paese)", senderText, 8)
	cmrBox(pdf, 10, 52, 95, 30, 2, "Destinatario (nome, indirizzo, paese)", consigneeText, 8)
	cmrBox(pdf, 105, 22, 95, 20, 3, "Luogo riconsegna della merce (luogo / paese)", deliveryPlace, 9)
	cmrBox(pdf, 105, 42, 95, 20, 4, "Luogo e data presa in carico", pickupPlace+"\n"+pickupDate, 9)
	cmrBox(pdf, 105, 62, 95, 20, 5, "Documenti allegati", "", 9)
}

func orDash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}

func drawCMRGoodsTable(pdf *fpdf.Fpdf, items []models.OrderItem, totalWeight float64) {
	y := 82.0
	type col struct {
		w     float64
		label string
	}
	headers := []col{
		{12, "6 Marche e numeri"}, {15, "7 N. colli"}, {20, "8 Imballaggio"},
		{60, "9 Natura della merce"}, {20, "10 N. statistico"},
		{25, "11 Peso lordo (kg)"}, {28, "12 Volume (mc)"},
	}
	pdf.SetXY(10, y)
	pdf.SetFont("Helvetica", "B", 7)
	pdf.SetFillColor(235, 238, 242)
	for _, h := range headers {
		pdf.CellFormat(h.w, 5, safe(h.label), "1", 0, "L", true, 0, "")
	}
	pdf.Ln(-1)

	pdf.SetFont("Helvetica", "", 8)
	rows := items
	emptyRow := len(rows) == 0
	if emptyRow {
		rows = []models.OrderItem{{Quantita: 0, Peso: totalWeight}}
	}
	if len(rows) > 6 {
		rows = rows[:6]
	}
	for _, it := range rows {
		desc := "merce"
		switch {
		case emptyRow:
			desc = "Merce varia"
		case it.Prodotto.Descrizione != "":
			desc = it.Prodotto.Descrizione
		case it.Prodotto.Codice != "":
			desc = it.Prodotto.Codice
		}
		if len(desc) > 55 {
			desc = desc[:55]
		}
		type cell struct {
			w     float64
			txt   string
			align string
		}
		cells := []cell{
			{12, "-", "L"}, {15, fmtG(it.Quantita), "R"}, {20, "Pallet", "L"},
			{60, safe(desc), "L"}, {20, "-", "L"}, {25, fmtG(it.Peso), "R"}, {28, "-", "L"},
		}
		for _, c := range cells {
			pdf.CellFormat(c.w, 5, c.txt, "1", 0, c.align, false, 0, "")
		}
		pdf.Ln(-1)
	}

	pdf.SetFont("Helvetica", "B", 8)
	pdf.CellFormat(107, 5, "Totale", "1", 0, "R", false, 0, "")
	pdf.CellFormat(20, 5, "", "1", 0, "L", false, 0, "")
	pdf.CellFormat(25, 5, fmtG(totalWeight), "1", 0, "R", false, 0, "")
	pdf.CellFormat(28, 5, "", "1", 0, "L", false, 0, "")
	pdf.Ln(8)
}

func drawCMRBottomBlock(pdf *fpdf.Fpdf, order models.Order, motrice *models.Motrice, semirimorchio *models.Semirimorchio) {
	y := pdf.GetY() + 2
	cmrBox(pdf, 10, y, 95, 22, 13, "Istruzioni del mittente", order.Note, 8)
	cmrBox(pdf, 105, y, 95, 22, 14, "Prescrizioni di affrancazione", "", 9)

	y2 := y + 22
	cmrBox(pdf, 10, y2, 95, 12, 15, "Rimborso", "", 9)
	vehicleInfo := ""
	if motrice != nil {
		vehicleInfo = fmt.Sprintf("Targa motrice: %s\nMarca: %s %s", orDash(motrice.Targa), motrice.Marca, motrice.Modello)
	}
	if semirimorchio != nil {
		vehicleInfo = strings.TrimSpace(vehicleInfo + fmt.Sprintf("\nTarga rimorchio: %s", orDash(semirimorchio.Targa)))
	}
	if vehicleInfo == "" && order.Vettore != nil {
		vehicleInfo = order.Vettore.RagioneSociale
	}
	if vehicleInfo == "" {
		vehicleInfo = DefaultSender.RagioneSociale
	}
	cmrBox(pdf, 105, y2, 95, 12, 16, "Vettore (nome, indirizzo, paese)", vehicleInfo, 8)

	y3 := y2 + 12
	cmrBox(pdf, 10, y3, 95, 15, 17, "Vettori successivi", "", 9)
	cmrBox(pdf, 105, y3, 95, 15, 18, "Riserve e osservazioni dei vettori", "", 9)

	y4 := y3 + 15
	cmrBox(pdf, 10, y4, 190, 10, 19, "Convenzioni speciali", "", 9)
	cmrBox(pdf, 10, y4+10, 190, 10, 20, "Da pagare", "", 8)

	y5 := y4 + 20
	clienteNome := ""
	if order.Cliente.ID != uuid.Nil {
		clienteNome = order.Cliente.RagioneSociale
	}
	cmrBox(pdf, 10, y5, 190, 10, 21, "Stilato a / il",
		fmt.Sprintf("   %s - %s", orDash(clienteNome), fmtDate(order.DataRitiro)), 9)

	y6 := y5 + 10
	sigW := 63.3
	cmrBox(pdf, 10, y6, sigW, 22, 22, "Firma e timbro del mittente", "", 9)
	cmrBox(pdf, 10+sigW, y6, sigW, 22, 23, "Firma e timbro del vettore", "", 9)
	cmrBox(pdf, 10+2*sigW, y6, sigW, 22, 24, "Firma e timbro del destinatario", "", 9)
}

// MakeCMRFilename mirrors make_cmr_filename: `CMR_{data}_{progressivo}.pdf`.
func MakeCMRFilename(order models.Order) string {
	d := order.DataRitiro
	if d == "" {
		d = order.CreatedAt.Format("2006-01-02")
	}
	if len(d) > 10 {
		d = d[:10]
	}
	d = strings.ReplaceAll(d, "-", "")
	if d == "" {
		d = "DRAFT"
	}
	rif := order.Progressivo
	if rif == "" {
		rif = order.ID.String()
	}
	if len(rif) > 12 {
		rif = rif[:12]
	}
	rif = strings.ReplaceAll(rif, "/", "-")
	rif = strings.ReplaceAll(rif, " ", "_")
	if rif == "" {
		rif = "DRAFT"
	}
	return fmt.Sprintf("CMR_%s_%s.pdf", d, rif)
}
