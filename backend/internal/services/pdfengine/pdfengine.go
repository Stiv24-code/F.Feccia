// Package pdfengine ports OrderMesh's pure-Go PDF engine: page rendering and
// text layer via poppler-utils (pdftoppm, pdftotext -bbox-layout), with the
// Claude vision fallback for zones with no text (scanned PDFs). The engine
// is stateless — it turns a PDF plus a template into an inbound-order draft,
// persistence stays with the inbound-orders service.
package pdfengine

import (
	"bytes"
	"context"
	"encoding/base64"
	"image"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"fratelli-feccia/internal/dto"
	"fratelli-feccia/internal/models"
	"fratelli-feccia/pkg/utils"
)

type PdfEngineService struct {
	anthropicAPIKey string
}

func NewPdfEngineService(anthropicAPIKey string) *PdfEngineService {
	return &PdfEngineService{anthropicAPIKey: anthropicAPIKey}
}

// Ready reports whether the poppler binaries are available in PATH. The
// engine degrades without them (endpoints answer 422), same posture as
// OrderMesh's /api/config pdfReady flag.
func (s *PdfEngineService) Ready() bool {
	_, err1 := exec.LookPath("pdftoppm")
	_, err2 := exec.LookPath("pdftotext")
	return err1 == nil && err2 == nil
}

// VisionReady reports whether the Claude vision fallback is configured.
func (s *PdfEngineService) VisionReady() bool { return s.anthropicAPIKey != "" }

func (s *PdfEngineService) popplerErr() error {
	return utils.NewAPIError(422, "poppler-utils non disponibile: installare pdftoppm/pdftotext (winget install poppler, apk add poppler-utils)")
}

// Render produces, for each page, the PNG (base64) plus the detected text
// blocks in normalized coordinates — the input for the template editor.
func (s *PdfEngineService) Render(ctx context.Context, filename string, pdf []byte) (*dto.PdfRenderResponse, error) {
	if !s.Ready() {
		return nil, s.popplerErr()
	}
	dir, err := os.MkdirTemp("", "feccia-pdf-*")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(dir)
	pdfPath := filepath.Join(dir, "in.pdf")
	if err := os.WriteFile(pdfPath, pdf, 0o600); err != nil {
		return nil, err
	}

	pngPaths, err := renderAllPages(dir, pdfPath)
	if err != nil {
		return nil, utils.NewAPIError(422, err.Error())
	}
	textPages, err := pdfTextLayout(dir, pdfPath)
	if err != nil {
		return nil, utils.NewAPIError(422, err.Error())
	}

	res := &dto.PdfRenderResponse{Filename: filename, PageCount: len(pngPaths), Pages: []dto.PdfRenderPageDTO{}}
	for i, pngPath := range pngPaths {
		raw, err := os.ReadFile(pngPath)
		if err != nil {
			return nil, err
		}
		cfgImg, err := png.DecodeConfig(bytes.NewReader(raw))
		if err != nil {
			return nil, err
		}
		page := dto.PdfRenderPageDTO{
			PageNum:  i,
			ImageB64: base64.StdEncoding.EncodeToString(raw),
			Width:    cfgImg.Width,
			Height:   cfgImg.Height,
			Blocks:   []dto.PdfRenderBlockDTO{},
		}
		if i < len(textPages) {
			tp := textPages[i]
			for _, b := range tp.Blocks {
				text := blockText(b)
				if text == "" || tp.Width == 0 || tp.Height == 0 {
					continue
				}
				if len(text) > 120 {
					text = text[:120]
				}
				page.Blocks = append(page.Blocks, dto.PdfRenderBlockDTO{
					Text: text,
					BoundsNorm: map[string]float64{
						"x":      b.XMin / tp.Width,
						"y":      b.YMin / tp.Height,
						"width":  (b.XMax - b.XMin) / tp.Width,
						"height": (b.YMax - b.YMin) / tp.Height,
					},
				})
			}
		}
		res.Pages = append(res.Pages, page)
	}
	return res, nil
}

// Extract reads the template zones from the PDF: text layer first (poppler
// words intersecting the zone), Claude vision as fallback. Returns the
// extracted value per template field ID.
func (s *PdfEngineService) Extract(ctx context.Context, pdf []byte, tpl dto.PdfTemplateResponse) (map[string]dto.PdfExtractedValueDTO, error) {
	if !s.Ready() {
		return nil, s.popplerErr()
	}
	dir, err := os.MkdirTemp("", "feccia-ext-*")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(dir)
	pdfPath := filepath.Join(dir, "in.pdf")
	if err := os.WriteFile(pdfPath, pdf, 0o600); err != nil {
		return nil, err
	}

	textPages, err := pdfTextLayout(dir, pdfPath)
	if err != nil {
		return nil, utils.NewAPIError(422, err.Error())
	}

	values := map[string]dto.PdfExtractedValueDTO{}
	pageImgs := map[int]image.Image{} // lazy render cache for vision fallback

	for _, f := range tpl.Fields {
		if f.Page < 0 || f.Page >= len(textPages) {
			values[f.ID] = dto.PdfExtractedValueDTO{Method: "page-out-of-range"}
			continue
		}
		// 1) layer testo
		text := zoneText(textPages[f.Page], f)
		if text != "" {
			values[f.ID] = dto.PdfExtractedValueDTO{Value: text, Confidence: 1.0, Method: "poppler-text"}
			continue
		}
		// 2) fallback visione (solo con API key)
		if s.anthropicAPIKey == "" {
			values[f.ID] = dto.PdfExtractedValueDTO{Method: "empty"}
			continue
		}
		img, ok := pageImgs[f.Page]
		if !ok {
			img, err = renderOnePage(dir, pdfPath, f.Page)
			if err != nil {
				values[f.ID] = dto.PdfExtractedValueDTO{Method: "render-error"}
				continue
			}
			pageImgs[f.Page] = img
		}
		crop, rotated := cropZone(img, f)
		if crop == nil {
			values[f.ID] = dto.PdfExtractedValueDTO{Method: "skipped-too-small"}
			continue
		}
		v, err := claudeVision(ctx, s.anthropicAPIKey, crop, rotated)
		if err != nil || v == "" {
			values[f.ID] = dto.PdfExtractedValueDTO{Method: "empty"}
			continue
		}
		values[f.ID] = dto.PdfExtractedValueDTO{Value: v, Confidence: 0.9, Method: "claude-vision"}
	}
	return values, nil
}

// BuildDraft maps the extracted values onto an inbound-order draft using the
// template field targets. Values for the same target are joined with a space
// (e.g. a template may map two zones onto "notes"). Returns the draft plus
// the merged value per target, for the UI's review step.
func (s *PdfEngineService) BuildDraft(tpl dto.PdfTemplateResponse, values map[string]dto.PdfExtractedValueDTO, sender, filename string) (dto.InboundOrderDraftDTO, map[string]string) {
	byTarget := map[string]string{}
	for _, f := range tpl.Fields {
		v, ok := values[f.ID]
		if !ok || strings.TrimSpace(v.Value) == "" {
			continue
		}
		val := strings.TrimSpace(v.Value)
		if prev := byTarget[f.Target]; prev != "" {
			val = prev + " " + val
		}
		byTarget[f.Target] = val
	}

	client := byTarget["client"]
	if client == "" {
		client = tpl.Client
	}
	if client == "" {
		client = tpl.Name
	}
	senderEmail := byTarget["sender_email"]
	if senderEmail == "" {
		senderEmail = strings.ToLower(strings.TrimSpace(sender))
	}
	if senderEmail == "" && len(tpl.Senders) > 0 && !strings.HasPrefix(tpl.Senders[0], "@") {
		senderEmail = tpl.Senders[0]
	}

	kg := 0
	if n, err := strconv.Atoi(cleanNumber(byTarget["kg"])); err == nil {
		kg = n
	}

	templateID := tpl.ID
	draft := dto.InboundOrderDraftDTO{
		Client:        client,
		SenderEmail:   senderEmail,
		Ref:           byTarget["ref"],
		Product:       byTarget["product"],
		Kg:            kg,
		LoadDate:      byTarget["load_date"],
		LoadPlace:     byTarget["load_place"],
		DeliveryDate:  byTarget["delivery_date"],
		DeliveryPlace: byTarget["delivery_place"],
		Rate:          byTarget["rate"],
		Notes:         byTarget["notes"],
		Status:        string(models.InboundOrderStatusPending),
		Source:        models.InboundOrderSourcePDF,
		TemplateID:    &templateID,
		ReceivedAt:    time.Now(),
	}
	if draft.Ref == "" {
		draft.Ref = strings.TrimSuffix(filename, ".pdf")
	}
	return draft, byTarget
}

// cleanNumber strips thousand separators and everything after a decimal comma,
// keeping only the leading integer part ("28.000 kg" -> "28000").
func cleanNumber(s string) string {
	s = strings.TrimSpace(s)
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '.' || r == ' ' || r == '\'':
			continue // thousand separators
		case r == ',':
			return b.String() // drop decimals
		default:
			if b.Len() > 0 {
				return b.String() // stop at first non-numeric after digits
			}
		}
	}
	return b.String()
}
