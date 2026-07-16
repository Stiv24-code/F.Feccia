package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"fratelli-feccia/internal/services"
	"fratelli-feccia/pkg/utils"
)

type DashboardHandler struct {
	Service services.Dashboard
}

func NewDashboardHandler(service services.Dashboard) *DashboardHandler {
	return &DashboardHandler{Service: service}
}

// @Summary Global dashboard KPIs
// @Tags Dashboard
// @Security BearerAuth
// @Produce json
// @Success 200 {object} dto.DashboardStatsResponse
// @Router /api/v1/dashboard/stats [get]
func (h *DashboardHandler) Stats(c *fiber.Ctx) error {
	result, err := h.Service.Stats(utils.RequestContext(c))
	if err != nil {
		return utils.HandleDatabaseError(c, err)
	}
	return utils.SuccessResponse(c, 200, result)
}

// @Summary Per-customer commercial dashboard
// @Tags Dashboard
// @Security BearerAuth
// @Produce json
// @Param customer_id path string true "Customer ID (UUID)"
// @Success 200 {object} dto.CustomerDashboardResponse
// @Failure 404 {object} map[string]string
// @Router /api/v1/dashboard/customer/{customer_id} [get]
func (h *DashboardHandler) CustomerDashboard(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("customer_id"))
	if err != nil {
		return utils.ErrorResponse(c, 400, "Invalid ID")
	}
	result, err := h.Service.CustomerDashboard(utils.RequestContext(c), id)
	if err != nil {
		return utils.HandleDatabaseError(c, err)
	}
	return utils.SuccessResponse(c, 200, result)
}

// @Summary Recent orders (last 10)
// @Tags Dashboard
// @Security BearerAuth
// @Produce json
// @Success 200 {array} dto.OrderResponse
// @Router /api/v1/dashboard/recent-orders [get]
func (h *DashboardHandler) RecentOrders(c *fiber.Ctx) error {
	result, err := h.Service.RecentOrders(utils.RequestContext(c))
	if err != nil {
		return utils.HandleDatabaseError(c, err)
	}
	return utils.SuccessResponse(c, 200, result)
}
