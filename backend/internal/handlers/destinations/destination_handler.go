package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"fratelli-feccia/internal/dto"
	"fratelli-feccia/internal/services"
	"fratelli-feccia/pkg/utils"
)

type DestinationHandler struct {
	Service services.Destination
}

func NewDestinationHandler(service services.Destination) *DestinationHandler {
	return &DestinationHandler{Service: service}
}

// ListDestinations godoc
// @Summary List destinations
// @Tags Destinations
// @Security BearerAuth
// @Produce json
// @Param search query string false "Filter by nome (substring, case-insensitive)"
// @Param include_inactive query bool false "Include logically deleted (active=false) destinations"
// @Success 200 {array} dto.DestinationResponse
// @Router /api/v1/destinations [get]
func (h *DestinationHandler) ListDestinations(c *fiber.Ctx) error {
	ctx := utils.RequestContext(c)
	items, err := h.Service.List(ctx, c.Query("search"), c.QueryBool("include_inactive", false))
	if err != nil {
		return utils.HandleDatabaseError(c, err)
	}
	return utils.SuccessResponse(c, 200, items)
}

// GetDestinationByID godoc
// @Summary Get destination by ID
// @Tags Destinations
// @Security BearerAuth
// @Produce json
// @Param id path string true "Destination ID (UUID)"
// @Success 200 {object} dto.DestinationResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/destinations/{id} [get]
func (h *DestinationHandler) GetDestinationByID(c *fiber.Ctx) error {
	ctx := utils.RequestContext(c)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.ErrorResponse(c, 400, "Invalid ID")
	}
	item, err := h.Service.GetByID(ctx, id)
	if err != nil {
		return utils.HandleDatabaseError(c, err)
	}
	if item == nil {
		return utils.ErrorResponse(c, 404, "Destinazione non trovata")
	}
	return utils.SuccessResponse(c, 200, item)
}

// CreateDestination godoc
// @Summary Create destination
// @Tags Destinations
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param destination body dto.DestinationRequest true "Destination data"
// @Success 201 {object} dto.DestinationResponse
// @Failure 400 {object} map[string]string
// @Router /api/v1/destinations [post]
func (h *DestinationHandler) CreateDestination(c *fiber.Ctx) error {
	ctx := utils.RequestContext(c)
	var req dto.DestinationRequest
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

// UpdateDestination godoc
// @Summary Update destination (full replace)
// @Tags Destinations
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Destination ID (UUID)"
// @Param destination body dto.DestinationRequest true "Destination data"
// @Success 200 {object} dto.DestinationResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/destinations/{id} [put]
func (h *DestinationHandler) UpdateDestination(c *fiber.Ctx) error {
	ctx := utils.RequestContext(c)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.ErrorResponse(c, 400, "Invalid ID")
	}
	var req dto.DestinationRequest
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

// DeleteDestination godoc
// @Summary Delete destination (logical, sets active=false)
// @Tags Destinations
// @Security BearerAuth
// @Param id path string true "Destination ID (UUID)"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]string
// @Router /api/v1/destinations/{id} [delete]
func (h *DestinationHandler) DeleteDestination(c *fiber.Ctx) error {
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
