package handlers

import (
	"github.com/gofiber/fiber/v2"

	"fratelli-feccia/internal/services"
	"fratelli-feccia/pkg/utils"
)

type MapHandler struct {
	Service services.Map
}

func NewMapHandler(service services.Map) *MapHandler {
	return &MapHandler{Service: service}
}

// @Summary Live map: active trips, POI, garages, stats
// @Tags Map
// @Security BearerAuth
// @Produce json
// @Success 200 {object} dto.MapTripsResponse
// @Router /api/v1/map/trips [get]
func (h *MapHandler) Trips(c *fiber.Ctx) error {
	result, err := h.Service.Trips(utils.RequestContext(c))
	if err != nil {
		return utils.HandleDatabaseError(c, err)
	}
	return utils.SuccessResponse(c, 200, result)
}
