package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"fratelli-feccia/internal/dto"
	"fratelli-feccia/internal/services"
	"fratelli-feccia/pkg/utils"
)

type CarrierHandler struct {
	Service services.Carrier
}

func NewCarrierHandler(service services.Carrier) *CarrierHandler {
	return &CarrierHandler{Service: service}
}

// ListCarriers godoc
// @Summary List carriers
// @Tags Carriers
// @Security BearerAuth
// @Produce json
// @Param search query string false "Filter by ragione sociale"
// @Success 200 {array} dto.CarrierResponse
// @Router /api/v1/carriers [get]
func (h *CarrierHandler) ListCarriers(c *fiber.Ctx) error {
	ctx := utils.RequestContext(c)
	items, err := h.Service.List(ctx, c.Query("search"))
	if err != nil {
		return utils.HandleDatabaseError(c, err)
	}
	return utils.SuccessResponse(c, 200, items)
}

// CreateCarrier godoc
// @Summary Create carrier
// @Tags Carriers
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param carrier body dto.CarrierRequest true "Carrier data"
// @Success 201 {object} dto.CarrierResponse
// @Failure 400 {object} map[string]string
// @Router /api/v1/carriers [post]
func (h *CarrierHandler) CreateCarrier(c *fiber.Ctx) error {
	ctx := utils.RequestContext(c)
	var req dto.CarrierRequest
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

// UpdateCarrier godoc
// @Summary Update carrier (full replace)
// @Tags Carriers
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Carrier ID (UUID)"
// @Param carrier body dto.CarrierRequest true "Carrier data"
// @Success 200 {object} dto.CarrierResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/carriers/{id} [put]
func (h *CarrierHandler) UpdateCarrier(c *fiber.Ctx) error {
	ctx := utils.RequestContext(c)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.ErrorResponse(c, 400, "Invalid ID")
	}
	var req dto.CarrierRequest
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

// DeleteCarrier godoc
// @Summary Delete carrier (logical, sets active=false)
// @Tags Carriers
// @Security BearerAuth
// @Param id path string true "Carrier ID (UUID)"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]string
// @Router /api/v1/carriers/{id} [delete]
func (h *CarrierHandler) DeleteCarrier(c *fiber.Ctx) error {
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
