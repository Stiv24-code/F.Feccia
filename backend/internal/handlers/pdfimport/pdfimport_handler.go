// Package handlers (pdfimport) exposes the OrderMesh PDF import flow:
// render pages for the template editor, dry-run an unsaved template, and
// extract an inbound-order draft from a PDF with a saved template. None of
// these endpoints persist anything — the draft is confirmed by the operator
// via the inbound-orders API.
package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"fratelli-feccia/internal/dto"
	"fratelli-feccia/internal/services"
	"fratelli-feccia/pkg/utils"
)

// maxPDFBytes mirrors OrderMesh's 25 MB multipart cap (the Fiber BodyLimit
// is raised accordingly in internal/app/bootstrap.go).
const maxPDFBytes = 25 << 20

type PdfImportHandler struct {
	Templates services.PdfTemplate
	Engine    services.PdfEngine
}

func NewPdfImportHandler(templates services.PdfTemplate, engine services.PdfEngine) *PdfImportHandler {
	return &PdfImportHandler{Templates: templates, Engine: engine}
}

// readPDFUpload extracts the "file" part from the multipart upload.
func readPDFUpload(c *fiber.Ctx) (filename string, pdf []byte, err error) {
	hdr, err := c.FormFile("file")
	if err != nil {
		return "", nil, utils.NewAPIError(400, "nessun file PDF nel campo 'file'")
	}
	if !strings.HasSuffix(strings.ToLower(hdr.Filename), ".pdf") {
		return "", nil, utils.NewAPIError(400, "sono accettati solo file .pdf")
	}
	if hdr.Size > maxPDFBytes {
		return "", nil, utils.NewAPIError(413, "file troppo grande (max 25 MB)")
	}
	f, err := hdr.Open()
	if err != nil {
		return "", nil, err
	}
	defer f.Close()
	pdf, err = io.ReadAll(f)
	if err != nil {
		return "", nil, err
	}
	if len(pdf) == 0 {
		return "", nil, utils.NewAPIError(400, "il file caricato e' vuoto")
	}
	return hdr.Filename, pdf, nil
}

// RenderPdf godoc
// @Summary Render PDF pages + text blocks for the template editor
// @Description Multipart upload; returns per-page PNG (base64) and detected text blocks in normalized coordinates.
// @Tags PdfImport
// @Security BearerAuth
// @Accept mpfd
// @Produce json
// @Param file formData file true "PDF file (max 25 MB)"
// @Success 200 {object} dto.PdfRenderResponse
// @Failure 400 {object} map[string]string
// @Failure 422 {object} map[string]string
// @Router /api/v1/pdf/render [post]
func (h *PdfImportHandler) RenderPdf(c *fiber.Ctx) error {
	ctx := utils.RequestContext(c)
	filename, pdf, err := readPDFUpload(c)
	if err != nil {
		return utils.HandleDatabaseError(c, err)
	}
	res, err := h.Engine.Render(ctx, filename, pdf)
	if err != nil {
		return utils.HandleDatabaseError(c, err)
	}
	return utils.SuccessResponse(c, 200, res)
}

// TestPdfTemplate godoc
// @Summary Dry-run a (possibly unsaved) template against a PDF
// @Description Multipart {file, template: JSON, sender?}. Nothing is persisted — returns the draft the template would produce.
// @Tags PdfImport
// @Security BearerAuth
// @Accept mpfd
// @Produce json
// @Param file formData file true "PDF file (max 25 MB)"
// @Param template formData string true "Template JSON (dto.PdfTemplateRequest shape)"
// @Param sender formData string false "Sender address for the draft"
// @Success 200 {object} dto.PdfExtractionResponse
// @Failure 400 {object} map[string]string
// @Failure 422 {object} map[string]string
// @Router /api/v1/pdf/test [post]
func (h *PdfImportHandler) TestPdfTemplate(c *fiber.Ctx) error {
	ctx := utils.RequestContext(c)
	filename, pdf, err := readPDFUpload(c)
	if err != nil {
		return utils.HandleDatabaseError(c, err)
	}
	var tpl dto.PdfTemplateResponse
	if err := json.Unmarshal([]byte(c.FormValue("template")), &tpl); err != nil {
		return utils.ErrorResponse(c, 400, "campo 'template' non valido: "+err.Error())
	}
	if len(tpl.Fields) == 0 {
		return utils.ErrorResponse(c, 422, "il template non ha campi mappati")
	}
	// An unsaved template from the editor may carry fields without IDs yet —
	// the extraction map is keyed by field ID, so assign stable ones.
	for i := range tpl.Fields {
		if tpl.Fields[i].ID == "" {
			tpl.Fields[i].ID = fmt.Sprintf("fld-%d", i)
		}
	}
	values, err := h.Engine.Extract(ctx, pdf, tpl)
	if err != nil {
		return utils.HandleDatabaseError(c, err)
	}
	draft, byTarget := h.Engine.BuildDraft(tpl, values, c.FormValue("sender"), filename)
	return utils.SuccessResponse(c, 200, dto.PdfExtractionResponse{
		Order:      draft,
		Values:     byTarget,
		Extraction: values,
	})
}

// ImportPdf godoc
// @Summary Extract an inbound-order draft from a PDF using a saved template
// @Description Multipart {file, template_id?, sender?}. Template resolution: explicit template_id wins, otherwise the best sender match. Nothing is persisted — the returned draft is confirmed via POST /inbound-orders.
// @Tags PdfImport
// @Security BearerAuth
// @Accept mpfd
// @Produce json
// @Param file formData file true "PDF file (max 25 MB)"
// @Param template_id formData string false "Template ID (UUID); wins over sender matching"
// @Param sender formData string false "Sender address used to preselect the template"
// @Success 200 {object} dto.PdfExtractionResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 422 {object} map[string]string
// @Router /api/v1/pdf/import [post]
func (h *PdfImportHandler) ImportPdf(c *fiber.Ctx) error {
	ctx := utils.RequestContext(c)
	filename, pdf, err := readPDFUpload(c)
	if err != nil {
		return utils.HandleDatabaseError(c, err)
	}
	sender := c.FormValue("sender")

	var tpl *dto.PdfTemplateResponse
	if rawID := c.FormValue("template_id"); rawID != "" {
		id, err := uuid.Parse(rawID)
		if err != nil {
			return utils.ErrorResponse(c, 400, "template_id non valido")
		}
		tpl, err = h.Templates.GetByID(ctx, id)
		if err != nil {
			return utils.HandleDatabaseError(c, err)
		}
	} else {
		tpl, err = h.Templates.Match(ctx, sender)
		if err != nil {
			return utils.HandleDatabaseError(c, err)
		}
		if tpl == nil {
			return utils.ErrorResponse(c, 422, "nessun template associato al mittente "+sender+": selezionane uno manualmente")
		}
	}
	if len(tpl.Fields) == 0 {
		return utils.ErrorResponse(c, 422, "il template «"+tpl.Name+"» non ha campi mappati")
	}

	values, err := h.Engine.Extract(ctx, pdf, *tpl)
	if err != nil {
		return utils.HandleDatabaseError(c, err)
	}
	draft, byTarget := h.Engine.BuildDraft(*tpl, values, sender, filename)
	return utils.SuccessResponse(c, 200, dto.PdfExtractionResponse{
		Order:      draft,
		Values:     byTarget,
		Extraction: values,
		Template:   &dto.PdfTemplateRef{ID: tpl.ID, Name: tpl.Name},
	})
}
