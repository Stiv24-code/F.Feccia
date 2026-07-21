package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"fratelli-feccia/internal/dto"
	"fratelli-feccia/internal/services"
	"fratelli-feccia/pkg/utils"
)

type GarageHandler struct {
	Service services.Garage
}

func NewGarageHandler(service services.Garage) *GarageHandler {
	return &GarageHandler{Service: service}
}

// ListGarages godoc
// @Summary List garages
// @Tags Garages
// @Security BearerAuth
// @Produce json
// @Param include_inactive query bool false "Include logically deleted (active=false) garages"
// @Success 200 {array} dto.GarageResponse
// @Router /api/v1/garages [get]
func (h *GarageHandler) ListGarages(c *fiber.Ctx) error {
	ctx := utils.RequestContext(c)
	items, err := h.Service.List(ctx, c.QueryBool("include_inactive", false))
	if err != nil {
		return utils.HandleDatabaseError(c, err)
	}
	return utils.SuccessResponse(c, 200, items)
}

// CreateGarage godoc
// @Summary Create garage
// @Tags Garages
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param garage body dto.GarageRequest true "Garage data"
// @Success 201 {object} dto.GarageResponse
// @Failure 400 {object} map[string]string
// @Router /api/v1/garages [post]
func (h *GarageHandler) CreateGarage(c *fiber.Ctx) error {
	ctx := utils.RequestContext(c)
	var req dto.GarageRequest
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

// UpdateGarage godoc
// @Summary Update garage (full replace)
// @Tags Garages
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Garage ID (UUID)"
// @Param garage body dto.GarageRequest true "Garage data"
// @Success 200 {object} dto.GarageResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/garages/{id} [put]
func (h *GarageHandler) UpdateGarage(c *fiber.Ctx) error {
	ctx := utils.RequestContext(c)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.ErrorResponse(c, 400, "Invalid ID")
	}
	var req dto.GarageRequest
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

// DeleteGarage godoc
// @Summary Delete garage (logical, sets active=false)
// @Tags Garages
// @Security BearerAuth
// @Param id path string true "Garage ID (UUID)"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]string
// @Router /api/v1/garages/{id} [delete]
func (h *GarageHandler) DeleteGarage(c *fiber.Ctx) error {
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
