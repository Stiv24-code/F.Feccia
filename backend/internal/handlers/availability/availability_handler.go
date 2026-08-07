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

// @Summary Motrice availability for a date range
// @Tags Availability
// @Security BearerAuth
// @Produce json
// @Param data_da query string true "Range start (YYYY-MM-DD)"
// @Param data_a query string true "Range end (YYYY-MM-DD)"
// @Success 200 {array} dto.MotriceAvailabilityResponse
// @Router /api/v1/availability/motrici [get]
func (h *AvailabilityHandler) Motrici(c *fiber.Ctx) error {
	result, err := h.Service.MotriceAvailability(utils.RequestContext(c), c.Query("data_da"), c.Query("data_a"))
	if err != nil {
		return utils.HandleDatabaseError(c, err)
	}
	return utils.SuccessResponse(c, 200, result)
}

// @Summary Semirimorchio availability for a date range
// @Tags Availability
// @Security BearerAuth
// @Produce json
// @Param data_da query string true "Range start (YYYY-MM-DD)"
// @Param data_a query string true "Range end (YYYY-MM-DD)"
// @Success 200 {array} dto.SemirimorchioAvailabilityResponse
// @Router /api/v1/availability/semirimorchi [get]
func (h *AvailabilityHandler) Semirimorchi(c *fiber.Ctx) error {
	result, err := h.Service.SemirimorchioAvailability(utils.RequestContext(c), c.Query("data_da"), c.Query("data_a"))
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
