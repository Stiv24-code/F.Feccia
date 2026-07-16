package pdfgen

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/go-pdf/fpdf"

	"fratelli-feccia/internal/models"
)

// tripSegmentTypeLabels mirrors instructions_pdf.py's inline tipo_label dict.
var tripSegmentTypeLabels = map[string]string{
	"base_carico":               "Garage->Carico",
	"carico_scarico":            "Carico->Scarico",
	"scarico_carico_successivo": "Scarico->Prossimo carico",
	"scarico_base":              "Scarico->Garage",
}

// BuildInstructionsPDF mirrors build_instructions_pdf: the operational
// service sheet handed to the driver at loading time.
func BuildInstructionsPDF(
	trip models.Trip,
	orders []models.Order,
	segments []models.TripSegment,
	driver *models.Driver,
	customersByID map[string]models.Customer,
) ([]byte, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 18, 15)
	pdf.SetAutoPageBreak(true, 20)
	pdf.AliasNbPages("")

	pdf.SetHeaderFunc(func() {
		pdf.SetFont("Helvetica", "B", 14)
		pdf.CellFormat(0, 7, "ISTRUZIONI OPERATIVE - VIAGGIO", "", 1, "L", false, 0, "")
		pdf.SetFont("Helvetica", "", 9)
		pdf.CellFormat(0, 4, safe("ID viaggio: "+orDash(trip.ID.String())), "", 1, "L", false, 0, "")
		pdf.SetDrawColor(180, 180, 180)
		pdf.Line(15, pdf.GetY()+1, 195, pdf.GetY()+1)
		pdf.Ln(4)
	})
	pdf.SetFooterFunc(func() {
		pdf.SetY(-15)
		pdf.SetFont("Helvetica", "I", 8)
		pdf.SetTextColor(120, 120, 120)
		pdf.CellFormat(0, 5, safe(fmt.Sprintf(
			"Pagina %d/{nb} - In caso di problemi contattare la centrale operativa", pdf.PageNo())),
			"", 0, "C", false, 0, "")
	})

	pdf.AddPage()

	drawTripHeader(pdf, trip, driver)
	drawOrdersBlock(pdf, orders, customersByID)
	drawSegmentsBlock(pdf, segments)
	drawSignatureBlock(pdf)

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func drawTripHeader(pdf *fpdf.Fpdf, trip models.Trip, driver *models.Driver) {
	pdf.SetFont("Helvetica", "B", 10)
	pdf.CellFormat(0, 5, "AUTISTA E MEZZO", "", 1, "L", false, 0, "")
	pdf.SetFont("Helvetica", "", 9)

	autistaNome := trip.AutistaNome
	if autistaNome == "" && driver != nil {
		autistaNome = strings.TrimSpace(driver.Cognome + " " + driver.Nome)
	}
	autistaTelefono := "-"
	if driver != nil && driver.Telefono != "" {
		autistaTelefono = driver.Telefono
	}
	pdf.CellFormat(0, 4, safe(fmt.Sprintf("Autista: %s   Telefono: %s", orDash(autistaNome), autistaTelefono)), "", 1, "L", false, 0, "")

	pdf.CellFormat(0, 4, safe(fmt.Sprintf(
		"Targa motrice: %s   Targa rimorchio: %s",
		orDash(trip.TargaMotrice), orDash(trip.TargaRimorchio))), "", 1, "L", false, 0, "")

	pdf.CellFormat(0, 4, safe(fmt.Sprintf(
		"Garage: %s   Partenza: %s   Arrivo prev.: %s",
		orDash(trip.GarageNome), fmtDate(trip.DataPartenza), fmtDate(trip.DataArrivo))), "", 1, "L", false, 0, "")

	if trip.VettoreNome != "" {
		pdf.CellFormat(0, 4, safe("Vettore: "+trip.VettoreNome), "", 1, "L", false, 0, "")
	}
	pdf.CellFormat(0, 4, safe("Km totali stimati: "+fmtG(trip.KmTotali)), "", 1, "L", false, 0, "")
	pdf.Ln(3)
}

func drawOrdersBlock(pdf *fpdf.Fpdf, orders []models.Order, customersByID map[string]models.Customer) {
	pdf.SetFont("Helvetica", "B", 10)
	pdf.CellFormat(0, 5, "ORDINI ASSEGNATI", "", 1, "L", false, 0, "")
	pdf.Ln(1)

	for idx, o := range orders {
		cliente := customersByID[o.ClienteID]

		pdf.SetFillColor(240, 243, 247)
		pdf.SetFont("Helvetica", "B", 9)
		pdf.CellFormat(0, 6, safe(fmt.Sprintf("%d. %s - rif. %s", idx+1, orDash(o.ClienteNome), orDash(o.RifOrdineCliente))), "", 1, "L", true, 0, "")

		pdf.SetFont("Helvetica", "", 9)
		pdf.CellFormat(0, 4, safe(fmt.Sprintf(
			"Carico: %s   Data: %s   Orario: %s",
			orDash(o.DestinazioneCaricoNome), fmtDate(o.DataRitiro), fmtTimeWindow(o.OraRitiroDa, o.OraRitiroA))), "", 1, "L", false, 0, "")
		pdf.CellFormat(0, 4, safe(fmt.Sprintf(
			"Scarico: %s   Data: %s   Orario: %s",
			orDash(o.DestinazioneScaricoNome), fmtDate(o.DataConsegna), fmtTimeWindow(o.OraConsegnaDa, o.OraConsegnaA))), "", 1, "L", false, 0, "")

		if len(o.Items) > 0 {
			var totPeso, totQta float64
			for _, it := range o.Items {
				totPeso += it.Peso
				totQta += it.Quantita
			}
			shown := o.Items
			extra := 0
			if len(shown) > 5 {
				extra = len(shown) - 5
				shown = shown[:5]
			}
			parts := make([]string, len(shown))
			for i, it := range shown {
				desc := it.ProdottoDescrizione
				if desc == "" {
					desc = it.ProdottoCodice
				}
				if desc == "" {
					desc = "merce"
				}
				parts[i] = fmt.Sprintf("%s (%s x %s kg)", desc, fmtG(it.Quantita), fmtG(it.Peso))
			}
			righe := strings.Join(parts, ", ")
			if extra > 0 {
				righe += fmt.Sprintf(" + altri %d", extra)
			}
			pdf.CellFormat(0, 4, safe("Merce: "+righe), "", 1, "L", false, 0, "")
			pdf.CellFormat(0, 4, safe(fmt.Sprintf("Totale: %s colli, %s kg", fmtG(totQta), fmtG(totPeso))), "", 1, "L", false, 0, "")
		}

		if cliente.Telefono != "" {
			pdf.CellFormat(0, 4, safe(fmt.Sprintf("Riferimento cliente: -   Telefono: %s", cliente.Telefono)), "", 1, "L", false, 0, "")
		}

		if o.Note != "" {
			pdf.SetFont("Helvetica", "I", 9)
			pdf.MultiCell(0, 4, safe("Note: "+o.Note), "", "L", false)
			pdf.SetFont("Helvetica", "", 9)
		}

		pdf.Ln(2)
	}
}

func drawSegmentsBlock(pdf *fpdf.Fpdf, segments []models.TripSegment) {
	if len(segments) == 0 {
		return
	}
	pdf.SetFont("Helvetica", "B", 10)
	pdf.CellFormat(0, 5, "PERCORSO STIMATO", "", 1, "L", false, 0, "")

	headers := []string{"#", "Tipo", "Da", "A", "Km", "Tempo"}
	widths := []float64{10, 35, 55, 55, 15, 20}

	pdf.SetFillColor(235, 238, 242)
	pdf.SetFont("Helvetica", "B", 8)
	for i, h := range headers {
		align := "L"
		if h == "Km" || h == "Tempo" {
			align = "R"
		}
		pdf.CellFormat(widths[i], 6, h, "1", 0, align, true, 0, "")
	}
	pdf.Ln(-1)

	pdf.SetFont("Helvetica", "", 8)
	for _, s := range segments {
		tipoLabel, ok := tripSegmentTypeLabels[s.Tipo]
		if !ok {
			tipoLabel = orDash(s.Tipo)
		}
		origine := s.OrigineNome
		if len(origine) > 35 {
			origine = origine[:35]
		}
		destinazione := s.DestinazioneNome
		if len(destinazione) > 35 {
			destinazione = destinazione[:35]
		}
		tempo := s.TempoStimatoMin

		type cell struct {
			w     float64
			txt   string
			align string
		}
		cells := []cell{
			{widths[0], fmt.Sprintf("%d", s.Ordine), "C"},
			{widths[1], safe(tipoLabel), "L"},
			{widths[2], safe(orDash(origine)), "L"},
			{widths[3], safe(orDash(destinazione)), "L"},
			{widths[4], fmtG(s.Km), "R"},
			{widths[5], fmt.Sprintf("%dh%02d", tempo/60, tempo%60), "R"},
		}
		for _, c := range cells {
			pdf.CellFormat(c.w, 5, c.txt, "1", 0, c.align, false, 0, "")
		}
		pdf.Ln(-1)
	}
	pdf.Ln(2)
}

func drawSignatureBlock(pdf *fpdf.Fpdf) {
	pdf.Ln(8)
	pdf.SetFont("Helvetica", "", 9)
	pdf.CellFormat(95, 5, "Firma autista alla partenza:", "", 0, "L", false, 0, "")
	pdf.CellFormat(95, 5, "Firma di consegna:", "", 1, "L", false, 0, "")
	pdf.SetDrawColor(180, 180, 180)
	y := pdf.GetY() + 8
	pdf.Line(20, y, 100, y)
	pdf.Line(115, y, 195, y)
}

// MakeInstructionsFilename mirrors make_instructions_filename.
func MakeInstructionsFilename(trip models.Trip) string {
	d := trip.DataPartenza
	if d == "" {
		d = trip.CreatedAt.Format("2006-01-02")
	}
	if len(d) > 10 {
		d = d[:10]
	}
	d = strings.ReplaceAll(d, "-", "")
	if d == "" {
		d = "DRAFT"
	}
	short := trip.ID.String()
	if len(short) > 8 {
		short = short[:8]
	}
	if short == "" {
		short = "DRAFT"
	}
	return fmt.Sprintf("ISTRUZIONI_%s_%s.pdf", d, short)
}
