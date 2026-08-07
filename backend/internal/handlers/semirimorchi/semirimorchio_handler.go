package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"fratelli-feccia/internal/dto"
	"fratelli-feccia/internal/services"
	"fratelli-feccia/pkg/utils"
)

type SemirimorchioHandler struct {
	Service services.Semirimorchio
}

func NewSemirimorchioHandler(service services.Semirimorchio) *SemirimorchioHandler {
	return &SemirimorchioHandler{Service: service}
}

// @Summary List semirimorchi
// @Tags Semirimorchi
// @Security BearerAuth
// @Produce json
// @Param search query string false "Filter by targa"
// @Success 200 {array} dto.SemirimorchioResponse
// @Router /api/v1/semirimorchi [get]
func (h *SemirimorchioHandler) ListSemirimorchi(c *fiber.Ctx) error {
	items, err := h.Service.List(utils.RequestContext(c), c.Query("search"))
	if err != nil {
		return utils.HandleDatabaseError(c, err)
	}
	return utils.SuccessResponse(c, 200, items)
}

// @Summary Create semirimorchio
// @Tags Semirimorchi
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param semirimorchio body dto.SemirimorchioRequest true "Semirimorchio data"
// @Success 201 {object} dto.SemirimorchioResponse
// @Failure 400 {object} map[string]string
// @Router /api/v1/semirimorchi [post]
func (h *SemirimorchioHandler) CreateSemirimorchio(c *fiber.Ctx) error {
	var req dto.SemirimorchioRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, 400, "Invalid request body")
	}
	if validationErrors := utils.NewValidator().Validate(&req); len(validationErrors) > 0 {
		return utils.ValidationErrorResponse(c, validationErrors)
	}
	item, err := h.Service.Create(utils.RequestContext(c), req)
	if err != nil {
		return utils.HandleDatabaseError(c, err)
	}
	return utils.SuccessResponse(c, 201, item)
}

// @Summary Update semirimorchio (full replace)
// @Tags Semirimorchi
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Semirimorchio ID (UUID)"
// @Param semirimorchio body dto.SemirimorchioRequest true "Semirimorchio data"
// @Success 200 {object} dto.SemirimorchioResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/semirimorchi/{id} [put]
func (h *SemirimorchioHandler) UpdateSemirimorchio(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.ErrorResponse(c, 400, "Invalid ID")
	}
	var req dto.SemirimorchioRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, 400, "Invalid request body")
	}
	if validationErrors := utils.NewValidator().Validate(&req); len(validationErrors) > 0 {
		return utils.ValidationErrorResponse(c, validationErrors)
	}
	item, err := h.Service.Update(utils.RequestContext(c), id, req)
	if err != nil {
		return utils.HandleDatabaseError(c, err)
	}
	return utils.SuccessResponse(c, 200, item)
}

// @Summary Delete semirimorchio (logical, sets active=false)
// @Tags Semirimorchi
// @Security BearerAuth
// @Param id path string true "Semirimorchio ID (UUID)"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]string
// @Router /api/v1/semirimorchi/{id} [delete]
func (h *SemirimorchioHandler) DeleteSemirimorchio(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.ErrorResponse(c, 400, "Invalid ID")
	}
	if err := h.Service.Delete(utils.RequestContext(c), id); err != nil {
		return utils.HandleDatabaseError(c, err)
	}
	return c.SendStatus(204)
}
