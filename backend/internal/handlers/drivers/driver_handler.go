package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"fratelli-feccia/internal/dto"
	"fratelli-feccia/internal/services"
	"fratelli-feccia/pkg/utils"
)

type DriverHandler struct {
	Service services.Driver
}

func NewDriverHandler(service services.Driver) *DriverHandler {
	return &DriverHandler{Service: service}
}

// ListDrivers godoc
// @Summary List drivers
// @Tags Drivers
// @Security BearerAuth
// @Produce json
// @Param search query string false "Filter by nome/cognome"
// @Param page query int false "Page number (default 1)"
// @Param limit query int false "Items per page (default 20, max 100)"
// @Success 200 {object} dto.DriverListResponse
// @Router /api/v1/drivers [get]
func (h *DriverHandler) ListDrivers(c *fiber.Ctx) error {
	ctx := utils.RequestContext(c)
	items, total, err := h.Service.List(ctx, c.Query("search"), utils.ParsePageParams(c))
	if err != nil {
		return utils.HandleDatabaseError(c, err)
	}
	return utils.SuccessResponse(c, 200, dto.DriverListResponse{Data: items, Total: total})
}

// CreateDriver godoc
// @Summary Create driver
// @Tags Drivers
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param driver body dto.DriverRequest true "Driver data"
// @Success 201 {object} dto.DriverResponse
// @Failure 400 {object} map[string]string
// @Router /api/v1/drivers [post]
func (h *DriverHandler) CreateDriver(c *fiber.Ctx) error {
	ctx := utils.RequestContext(c)
	var req dto.DriverRequest
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

// UpdateDriver godoc
// @Summary Update driver (full replace)
// @Tags Drivers
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Driver ID (UUID)"
// @Param driver body dto.DriverRequest true "Driver data"
// @Success 200 {object} dto.DriverResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/drivers/{id} [put]
func (h *DriverHandler) UpdateDriver(c *fiber.Ctx) error {
	ctx := utils.RequestContext(c)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.ErrorResponse(c, 400, "Invalid ID")
	}
	var req dto.DriverRequest
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

// DeleteDriver godoc
// @Summary Delete driver (logical, sets active=false)
// @Tags Drivers
// @Security BearerAuth
// @Param id path string true "Driver ID (UUID)"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]string
// @Router /api/v1/drivers/{id} [delete]
func (h *DriverHandler) DeleteDriver(c *fiber.Ctx) error {
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
