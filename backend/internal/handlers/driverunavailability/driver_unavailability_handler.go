package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"fratelli-feccia/internal/dto"
	"fratelli-feccia/internal/services"
	"fratelli-feccia/pkg/utils"
)

type DriverUnavailabilityHandler struct {
	Service services.DriverUnavailability
}

func NewDriverUnavailabilityHandler(service services.DriverUnavailability) *DriverUnavailabilityHandler {
	return &DriverUnavailabilityHandler{Service: service}
}

// @Summary List driver unavailability periods
// @Tags DriverUnavailability
// @Security BearerAuth
// @Produce json
// @Param autista_id query string false "Filter by driver ID (UUID)"
// @Param data_da query string false "Range start (overlap filter, requires data_a too)"
// @Param data_a query string false "Range end (overlap filter, requires data_da too)"
// @Success 200 {array} dto.DriverUnavailabilityResponse
// @Failure 400 {object} map[string]string
// @Router /api/v1/driver-unavailability [get]
func (h *DriverUnavailabilityHandler) List(c *fiber.Ctx) error {
	var autistaID uuid.UUID
	if raw := c.Query("autista_id"); raw != "" {
		parsed, err := uuid.Parse(raw)
		if err != nil {
			return utils.ErrorResponse(c, 400, "Invalid autista_id")
		}
		autistaID = parsed
	}

	items, err := h.Service.List(utils.RequestContext(c), autistaID, c.Query("data_da"), c.Query("data_a"))
	if err != nil {
		return utils.HandleDatabaseError(c, err)
	}
	return utils.SuccessResponse(c, 200, items)
}

// @Summary Create driver unavailability period
// @Tags DriverUnavailability
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param item body dto.DriverUnavailabilityRequest true "Driver unavailability data"
// @Success 201 {object} dto.DriverUnavailabilityResponse
// @Failure 400 {object} map[string]string
// @Router /api/v1/driver-unavailability [post]
func (h *DriverUnavailabilityHandler) Create(c *fiber.Ctx) error {
	var req dto.DriverUnavailabilityRequest
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

// @Summary Delete driver unavailability period (hard delete)
// @Tags DriverUnavailability
// @Security BearerAuth
// @Param id path string true "Record ID (UUID)"
// @Success 200 {object} map[string]bool
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/driver-unavailability/{id} [delete]
func (h *DriverUnavailabilityHandler) Delete(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.ErrorResponse(c, 400, "Invalid ID")
	}
	if err := h.Service.Delete(utils.RequestContext(c), id); err != nil {
		return utils.HandleDatabaseError(c, err)
	}
	return utils.SuccessResponse(c, 200, fiber.Map{"ok": true})
}
