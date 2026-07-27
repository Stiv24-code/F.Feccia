package handlers

import (
	"github.com/gofiber/fiber/v2"

	"fratelli-feccia/internal/services"
	"fratelli-feccia/pkg/utils"
)

type GeocodeHandler struct {
	Service services.Geocode
}

func NewGeocodeHandler(service services.Geocode) *GeocodeHandler {
	return &GeocodeHandler{Service: service}
}

// @Summary Forward-geocode an address (Destination/Garage/WashStation forms)
// @Tags Geocode
// @Security BearerAuth
// @Produce json
// @Param q query string true "Free-text address/place to search"
// @Success 200 {array} dto.GeocodeResultDTO
// @Router /api/v1/geocode/search [get]
func (h *GeocodeHandler) Search(c *fiber.Ctx) error {
	query := c.Query("q")
	if query == "" {
		return utils.ErrorResponse(c, 400, "Parametro 'q' obbligatorio")
	}
	results, err := h.Service.Search(utils.RequestContext(c), query)
	if err != nil {
		return utils.HandleDatabaseError(c, err)
	}
	return utils.SuccessResponse(c, 200, results)
}
