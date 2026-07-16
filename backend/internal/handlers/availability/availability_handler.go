package handlers

import (
	"github.com/gofiber/fiber/v2"

	"fratelli-feccia/internal/services"
	"fratelli-feccia/pkg/utils"
)

type AvailabilityHandler struct {
	Service services.Availability
}

func NewAvailabilityHandler(service services.Availability) *AvailabilityHandler {
	return &AvailabilityHandler{Service: service}
}

// @Summary Vehicle availability for a date range
// @Tags Availability
// @Security BearerAuth
// @Produce json
// @Param data_da query string true "Range start (YYYY-MM-DD)"
// @Param data_a query string true "Range end (YYYY-MM-DD)"
// @Success 200 {array} dto.VehicleAvailabilityResponse
// @Router /api/v1/availability/vehicles [get]
func (h *AvailabilityHandler) Vehicles(c *fiber.Ctx) error {
	result, err := h.Service.VehicleAvailability(utils.RequestContext(c), c.Query("data_da"), c.Query("data_a"))
	if err != nil {
		return utils.HandleDatabaseError(c, err)
	}
	return utils.SuccessResponse(c, 200, result)
}

// @Summary Driver availability for a date range
// @Tags Availability
// @Security BearerAuth
// @Produce json
// @Param data_da query string true "Range start (YYYY-MM-DD)"
// @Param data_a query string true "Range end (YYYY-MM-DD)"
// @Success 200 {array} dto.DriverAvailabilityResponse
// @Router /api/v1/availability/drivers [get]
func (h *AvailabilityHandler) Drivers(c *fiber.Ctx) error {
	result, err := h.Service.DriverAvailability(utils.RequestContext(c), c.Query("data_da"), c.Query("data_a"))
	if err != nil {
		return utils.HandleDatabaseError(c, err)
	}
	return utils.SuccessResponse(c, 200, result)
}
