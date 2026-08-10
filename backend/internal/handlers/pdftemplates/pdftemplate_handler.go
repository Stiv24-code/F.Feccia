package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"fratelli-feccia/internal/dto"
	"fratelli-feccia/internal/services"
	"fratelli-feccia/pkg/utils"
)

type PdfTemplateHandler struct {
	Service services.PdfTemplate
}

func NewPdfTemplateHandler(service services.PdfTemplate) *PdfTemplateHandler {
	return &PdfTemplateHandler{Service: service}
}

// ListPdfTemplates godoc
// @Summary List PDF templates
// @Tags PdfTemplates
// @Security BearerAuth
// @Produce json
// @Success 200 {array} dto.PdfTemplateResponse
// @Router /api/v1/pdf-templates [get]
func (h *PdfTemplateHandler) ListPdfTemplates(c *fiber.Ctx) error {
	ctx := utils.RequestContext(c)
	items, err := h.Service.List(ctx)
	if err != nil {
		return utils.HandleDatabaseError(c, err)
	}
	return utils.SuccessResponse(c, 200, items)
}

// MatchPdfTemplate godoc
// @Summary Best template for a mail sender (exact address beats domain); match is null when nothing matches
// @Tags PdfTemplates
// @Security BearerAuth
// @Produce json
// @Param sender query string true "Sender address, e.g. ordini@cliente.it"
// @Success 200 {object} dto.PdfTemplateMatchResponse
// @Router /api/v1/pdf-templates/match [get]
func (h *PdfTemplateHandler) MatchPdfTemplate(c *fiber.Ctx) error {
	ctx := utils.RequestContext(c)
	match, err := h.Service.Match(ctx, c.Query("sender"))
	if err != nil {
		return utils.HandleDatabaseError(c, err)
	}
	return utils.SuccessResponse(c, 200, dto.PdfTemplateMatchResponse{Match: match})
}

// CreatePdfTemplate godoc
// @Summary Create PDF template
// @Tags PdfTemplates
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param template body dto.PdfTemplateRequest true "Template data"
// @Success 201 {object} dto.PdfTemplateResponse
// @Failure 400 {object} map[string]string
// @Router /api/v1/pdf-templates [post]
func (h *PdfTemplateHandler) CreatePdfTemplate(c *fiber.Ctx) error {
	ctx := utils.RequestContext(c)
	var req dto.PdfTemplateRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, 400, "Invalid request body")
	}
	if validationErrors := utils.NewValidator().Validate(&req); len(validationErrors) > 0 {
		return utils.ValidationErrorResponse(c, validationErrors)
	}
	item, err := h.Service.Create(ctx, req)
	if err != nil {
		return utils.HandleDatabaseError(c, err)
	}
	return utils.SuccessResponse(c, 201, item)
}

// UpdatePdfTemplate godoc
// @Summary Update PDF template (full replace)
// @Tags PdfTemplates
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Template ID (UUID)"
// @Param template body dto.PdfTemplateRequest true "Template data"
// @Success 200 {object} dto.PdfTemplateResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/pdf-templates/{id} [put]
func (h *PdfTemplateHandler) UpdatePdfTemplate(c *fiber.Ctx) error {
	ctx := utils.RequestContext(c)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.ErrorResponse(c, 400, "Invalid ID")
	}
	var req dto.PdfTemplateRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, 400, "Invalid request body")
	}
	if validationErrors := utils.NewValidator().Validate(&req); len(validationErrors) > 0 {
		return utils.ValidationErrorResponse(c, validationErrors)
	}
	item, err := h.Service.Update(ctx, id, req)
	if err != nil {
		return utils.HandleDatabaseError(c, err)
	}
	return utils.SuccessResponse(c, 200, item)
}

// DeletePdfTemplate godoc
// @Summary Delete PDF template
// @Tags PdfTemplates
// @Security BearerAuth
// @Param id path string true "Template ID (UUID)"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/pdf-templates/{id} [delete]
func (h *PdfTemplateHandler) DeletePdfTemplate(c *fiber.Ctx) error {
	ctx := utils.RequestContext(c)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.ErrorResponse(c, 400, "Invalid ID")
	}
	if err := h.Service.Delete(ctx, id); err != nil {
		return utils.HandleDatabaseError(c, err)
	}
	return c.SendStatus(204)
}
