package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"fratelli-feccia/internal/dto"
	"fratelli-feccia/internal/services"
	"fratelli-feccia/pkg/utils"
)

type TripHandler struct {
	Service services.Trip
}

func NewTripHandler(service services.Trip) *TripHandler {
	return &TripHandler{Service: service}
}

// @Summary List trips
// @Tags Trips
// @Security BearerAuth
// @Produce json
// @Param stato query string false "Filter by stato"
// @Param limit query int false "Max results (default 200)"
// @Success 200 {array} dto.TripResponse
// @Router /api/v1/trips [get]
func (h *TripHandler) ListTrips(c *fiber.Ctx) error {
	items, err := h.Service.List(utils.RequestContext(c), c.Query("stato"), c.QueryInt("limit", 0))
	if err != nil {
		return utils.HandleDatabaseError(c, err)
	}
	return utils.SuccessResponse(c, 200, items)
}

// @Summary Create trip
// @Description Syncs any PIANIFICABILE orders in ordini_ids to PIANIFICATO and computes route segments via OSRM
// @Tags Trips
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param trip body dto.TripRequest true "Trip data"
// @Success 201 {object} dto.TripResponse
// @Router /api/v1/trips [post]
func (h *TripHandler) CreateTrip(c *fiber.Ctx) error {
	var req dto.TripRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, 400, "Invalid request body")
	}
	item, err := h.Service.Create(utils.RequestContext(c), req)
	if err != nil {
		return utils.HandleDatabaseError(c, err)
	}
	return utils.SuccessResponse(c, 201, item)
}

// @Summary Get trip by ID (includes joined orders)
// @Tags Trips
// @Security BearerAuth
// @Produce json
// @Param id path string true "Trip ID (UUID)"
// @Success 200 {object} dto.TripDetailResponse
// @Failure 404 {object} map[string]string
// @Router /api/v1/trips/{id} [get]
func (h *TripHandler) GetTripByID(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.ErrorResponse(c, 400, "Invalid ID")
	}
	item, err := h.Service.GetByID(utils.RequestContext(c), id)
	if err != nil {
		return utils.HandleDatabaseError(c, err)
	}
	if item == nil {
		return utils.ErrorResponse(c, 404, "Viaggio non trovato")
	}
	return utils.SuccessResponse(c, 200, item)
}

// @Summary Recompute a trip's route segments
// @Tags Trips
// @Security BearerAuth
// @Produce json
// @Param id path string true "Trip ID (UUID)"
// @Success 200 {object} dto.RecomputeSegmentsResult
// @Failure 404 {object} map[string]string
// @Router /api/v1/trips/{id}/recompute-segments [post]
func (h *TripHandler) RecomputeSegments(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.ErrorResponse(c, 400, "Invalid ID")
	}
	result, err := h.Service.RecomputeSegments(utils.RequestContext(c), id)
	if err != nil {
		return utils.HandleDatabaseError(c, err)
	}
	return utils.SuccessResponse(c, 200, result)
}

// @Summary Start a trip (PIANIFICATO -> IN_CORSO, starts its PIANIFICATO orders to VIAGGIO)
// @Tags Trips
// @Security BearerAuth
// @Produce json
// @Param id path string true "Trip ID (UUID)"
// @Success 200 {object} dto.OKResult
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/trips/{id}/start [patch]
func (h *TripHandler) StartTrip(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.ErrorResponse(c, 400, "Invalid ID")
	}
	result, err := h.Service.Start(utils.RequestContext(c), id)
	if err != nil {
		return utils.HandleDatabaseError(c, err)
	}
	return utils.SuccessResponse(c, 200, result)
}

// @Summary Complete a trip (closes its VIAGGIO orders)
// @Tags Trips
// @Security BearerAuth
// @Produce json
// @Param id path string true "Trip ID (UUID)"
// @Success 200 {object} dto.OKResult
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/trips/{id}/complete [patch]
func (h *TripHandler) CompleteTrip(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.ErrorResponse(c, 400, "Invalid ID")
	}
	result, err := h.Service.Complete(utils.RequestContext(c), id)
	if err != nil {
		return utils.HandleDatabaseError(c, err)
	}
	return utils.SuccessResponse(c, 200, result)
}

// @Summary Add an order to a trip
// @Tags Trips
// @Security BearerAuth
// @Produce json
// @Param id path string true "Trip ID (UUID)"
// @Param order_id query string true "Order ID (UUID)"
// @Success 200 {object} dto.OKResult
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/trips/{id}/add-order [patch]
func (h *TripHandler) AddOrderToTrip(c *fiber.Ctx) error {
	tripID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.ErrorResponse(c, 400, "Invalid ID")
	}
	orderID, err := uuid.Parse(c.Query("order_id"))
	if err != nil {
		return utils.ErrorResponse(c, 400, "Invalid order_id")
	}
	result, err := h.Service.AddOrder(utils.RequestContext(c), tripID, orderID)
	if err != nil {
		return utils.HandleDatabaseError(c, err)
	}
	return utils.SuccessResponse(c, 200, result)
}

// @Summary Operational instructions PDF for a trip
// @Tags Trips
// @Security BearerAuth
// @Produce application/pdf
// @Param id path string true "Trip ID"
// @Success 200 {file} binary
// @Failure 404 {object} map[string]string
// @Router /api/v1/trips/{id}/instructions/pdf [get]
func (h *TripHandler) GetInstructionsPDF(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.ErrorResponse(c, 400, "Invalid ID")
	}

	data, filename, err := h.Service.GetInstructionsPDF(utils.RequestContext(c), id)
	if err != nil {
		return utils.HandleDatabaseError(c, err)
	}
	c.Set(fiber.HeaderContentType, "application/pdf")
	c.Set(fiber.HeaderContentDisposition, `attachment; filename="`+filename+`"`)
	return c.Send(data)
}
