package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"fratelli-feccia/internal/dto"
	"fratelli-feccia/internal/services"
	"fratelli-feccia/pkg/utils"
)

type WashStationHandler struct {
	Service services.WashStation
}

func NewWashStationHandler(service services.WashStation) *WashStationHandler {
	return &WashStationHandler{Service: service}
}

// ListWashStations godoc
// @Summary List wash stations
// @Tags WashStations
// @Security BearerAuth
// @Produce json
// @Param include_inactive query bool false "Include logically deleted (active=false) wash stations"
// @Param page query int false "Page number (default 1)"
// @Param limit query int false "Items per page (default 20, max 100)"
// @Success 200 {object} dto.WashStationListResponse
// @Router /api/v1/wash-stations [get]
func (h *WashStationHandler) ListWashStations(c *fiber.Ctx) error {
	ctx := utils.RequestContext(c)
	items, total, err := h.Service.List(ctx, c.QueryBool("include_inactive", false), utils.ParsePageParams(c))
	if err != nil {
		return utils.HandleDatabaseError(c, err)
	}
	return utils.SuccessResponse(c, 200, dto.WashStationListResponse{Data: items, Total: total})
}

// CreateWashStation godoc
// @Summary Create wash station
// @Tags WashStations
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param washStation body dto.WashStationRequest true "Wash station data"
// @Success 201 {object} dto.WashStationResponse
// @Failure 400 {object} map[string]string
// @Router /api/v1/wash-stations [post]
func (h *WashStationHandler) CreateWashStation(c *fiber.Ctx) error {
	ctx := utils.RequestContext(c)
	var req dto.WashStationRequest
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

// UpdateWashStation godoc
// @Summary Update wash station (full replace)
// @Tags WashStations
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Wash station ID (UUID)"
// @Param washStation body dto.WashStationRequest true "Wash station data"
// @Success 200 {object} dto.WashStationResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/wash-stations/{id} [put]
func (h *WashStationHandler) UpdateWashStation(c *fiber.Ctx) error {
	ctx := utils.RequestContext(c)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.ErrorResponse(c, 400, "Invalid ID")
	}
	var req dto.WashStationRequest
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

// DeleteWashStation godoc
// @Summary Delete wash station (logical, sets active=false)
// @Tags WashStations
// @Security BearerAuth
// @Param id path string true "Wash station ID (UUID)"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]string
// @Router /api/v1/wash-stations/{id} [delete]
func (h *WashStationHandler) DeleteWashStation(c *fiber.Ctx) error {
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
