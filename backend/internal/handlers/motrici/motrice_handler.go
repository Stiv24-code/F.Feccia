package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"fratelli-feccia/internal/dto"
	"fratelli-feccia/internal/services"
	"fratelli-feccia/pkg/utils"
)

type MotriceHandler struct {
	Service services.Motrice
}

func NewMotriceHandler(service services.Motrice) *MotriceHandler {
	return &MotriceHandler{Service: service}
}

// @Summary List motrici
// @Tags Motrici
// @Security BearerAuth
// @Produce json
// @Param search query string false "Filter by targa"
// @Success 200 {array} dto.MotriceResponse
// @Router /api/v1/motrici [get]
func (h *MotriceHandler) ListMotrici(c *fiber.Ctx) error {
	items, err := h.Service.List(utils.RequestContext(c), c.Query("search"))
	if err != nil {
		return utils.HandleDatabaseError(c, err)
	}
	return utils.SuccessResponse(c, 200, items)
}

// @Summary Create motrice
// @Tags Motrici
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param motrice body dto.MotriceRequest true "Motrice data"
// @Success 201 {object} dto.MotriceResponse
// @Failure 400 {object} map[string]string
// @Router /api/v1/motrici [post]
func (h *MotriceHandler) CreateMotrice(c *fiber.Ctx) error {
	var req dto.MotriceRequest
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

// @Summary Update motrice (full replace)
// @Tags Motrici
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Motrice ID (UUID)"
// @Param motrice body dto.MotriceRequest true "Motrice data"
// @Success 200 {object} dto.MotriceResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/motrici/{id} [put]
func (h *MotriceHandler) UpdateMotrice(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.ErrorResponse(c, 400, "Invalid ID")
	}
	var req dto.MotriceRequest
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

// @Summary Delete motrice (logical, sets active=false)
// @Tags Motrici
// @Security BearerAuth
// @Param id path string true "Motrice ID (UUID)"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]string
// @Router /api/v1/motrici/{id} [delete]
func (h *MotriceHandler) DeleteMotrice(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.ErrorResponse(c, 400, "Invalid ID")
	}
	if err := h.Service.Delete(utils.RequestContext(c), id); err != nil {
		return utils.HandleDatabaseError(c, err)
	}
	return c.SendStatus(204)
}
