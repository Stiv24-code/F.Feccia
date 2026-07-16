package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"fratelli-feccia/internal/dto"
	"fratelli-feccia/internal/services"
	"fratelli-feccia/internal/services/orders"
	"fratelli-feccia/pkg/utils"
)

type OrderHandler struct {
	Service services.Order
}

func NewOrderHandler(service services.Order) *OrderHandler {
	return &OrderHandler{Service: service}
}

// @Summary List orders
// @Tags Orders
// @Security BearerAuth
// @Produce json
// @Param stato query string false "Filter by stato"
// @Param cliente_id query string false "Filter by cliente_id"
// @Param data_da query string false "data_ritiro >= data_da"
// @Param data_a query string false "data_ritiro <= data_a"
// @Param search query string false "Filter by cliente_nome/progressivo/rif_ordine_cliente/destinazioni"
// @Param tipologia query string false "Filter by tipologia"
// @Param limit query int false "Max results (default 500)"
// @Success 200 {array} dto.OrderResponse
// @Router /api/v1/orders [get]
func (h *OrderHandler) ListOrders(c *fiber.Ctx) error {
	items, err := h.Service.List(utils.RequestContext(c), orders.ListFilters{
		Stato:     c.Query("stato"),
		ClienteID: c.Query("cliente_id"),
		DataDa:    c.Query("data_da"),
		DataA:     c.Query("data_a"),
		Search:    c.Query("search"),
		Tipologia: c.Query("tipologia"),
		Limit:     c.QueryInt("limit", 0),
	})
	if err != nil {
		return utils.HandleDatabaseError(c, err)
	}
	return utils.SuccessResponse(c, 200, items)
}

// @Summary Create order
// @Tags Orders
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param order body dto.OrderRequest true "Order data"
// @Success 201 {object} dto.OrderResponse
// @Failure 400 {object} map[string]string
// @Router /api/v1/orders [post]
func (h *OrderHandler) CreateOrder(c *fiber.Ctx) error {
	var req dto.OrderRequest
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

// @Summary Get order by ID
// @Tags Orders
// @Security BearerAuth
// @Produce json
// @Param id path string true "Order ID (UUID)"
// @Success 200 {object} dto.OrderResponse
// @Failure 404 {object} map[string]string
// @Router /api/v1/orders/{id} [get]
func (h *OrderHandler) GetOrderByID(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.ErrorResponse(c, 400, "Invalid ID")
	}
	item, err := h.Service.GetByID(utils.RequestContext(c), id)
	if err != nil {
		return utils.HandleDatabaseError(c, err)
	}
	if item == nil {
		return utils.ErrorResponse(c, 404, "Ordine non trovato")
	}
	return utils.SuccessResponse(c, 200, item)
}

// @Summary Update order (full replace of the create-able fields)
// @Tags Orders
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Order ID (UUID)"
// @Param order body dto.OrderRequest true "Order data"
// @Success 200 {object} dto.OrderResponse
// @Failure 404 {object} map[string]string
// @Router /api/v1/orders/{id} [put]
func (h *OrderHandler) UpdateOrder(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.ErrorResponse(c, 400, "Invalid ID")
	}
	var req dto.OrderRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, 400, "Invalid request body")
	}
	if validationErrors := utils.NewValidator().Validate(&req); len(validationErrors) > 0 {
		return utils.ValidationErrorResponse(c, validationErrors)
	}
	item, err := h.Service.Update(utils.RequestContext(c), id, req)
	if err != nil {
		return utils.HandleDatabaseError(c, err)
	}
	return utils.SuccessResponse(c, 200, item)
}

// @Summary Assign order to a vehicle/driver/carrier (PIANIFICABILE -> VIAGGIO)
// @Tags Orders
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Order ID (UUID)"
// @Param assign body dto.OrderAssignRequest true "Assignment data"
// @Success 200 {object} dto.OrderResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/orders/{id}/assign [patch]
func (h *OrderHandler) AssignOrder(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.ErrorResponse(c, 400, "Invalid ID")
	}
	var req dto.OrderAssignRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, 400, "Invalid request body")
	}
	item, err := h.Service.Assign(utils.RequestContext(c), id, req)
	if err != nil {
		return utils.HandleDatabaseError(c, err)
	}
	return utils.SuccessResponse(c, 200, item)
}

// @Summary Close order (VIAGGIO -> CHIUSO)
// @Tags Orders
// @Security BearerAuth
// @Produce json
// @Param id path string true "Order ID (UUID)"
// @Success 200 {object} dto.OrderResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/orders/{id}/close [patch]
func (h *OrderHandler) CloseOrder(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.ErrorResponse(c, 400, "Invalid ID")
	}
	item, err := h.Service.Close(utils.RequestContext(c), id)
	if err != nil {
		return utils.HandleDatabaseError(c, err)
	}
	return utils.SuccessResponse(c, 200, item)
}

// @Summary Delete order (only PIANIFICABILE, hard delete)
// @Tags Orders
// @Security BearerAuth
// @Param id path string true "Order ID (UUID)"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/orders/{id} [delete]
func (h *OrderHandler) DeleteOrder(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.ErrorResponse(c, 400, "Invalid ID")
	}
	if err := h.Service.Delete(utils.RequestContext(c), id); err != nil {
		return utils.HandleDatabaseError(c, err)
	}
	return c.SendStatus(204)
}

// @Summary Return-trip order suggestions
// @Tags Orders
// @Security BearerAuth
// @Produce json
// @Param id path string true "Order ID (UUID)"
// @Param max_days_gap query int false "Days after data_consegna to search (0-14, default 2)"
// @Param limit query int false "Max candidates (1-100, default 20)"
// @Success 200 {object} dto.OrderReturnSuggestionsResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/orders/{id}/return-suggestions [get]
func (h *OrderHandler) ReturnSuggestions(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.ErrorResponse(c, 400, "Invalid ID")
	}
	maxDaysGap := c.QueryInt("max_days_gap", 2)
	limit := c.QueryInt("limit", 20)

	result, err := h.Service.ReturnSuggestions(utils.RequestContext(c), id, maxDaysGap, limit)
	if err != nil {
		return utils.HandleDatabaseError(c, err)
	}
	return utils.SuccessResponse(c, 200, result)
}

// @Summary CMR waybill PDF for an order
// @Tags Orders
// @Security BearerAuth
// @Produce application/pdf
// @Param id path string true "Order ID"
// @Success 200 {file} binary
// @Failure 404 {object} map[string]string
// @Router /api/v1/orders/{id}/cmr/pdf [get]
func (h *OrderHandler) GetCMRPDF(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.ErrorResponse(c, 400, "Invalid ID")
	}

	data, filename, err := h.Service.GetCMRPDF(utils.RequestContext(c), id)
	if err != nil {
		return utils.HandleDatabaseError(c, err)
	}
	c.Set(fiber.HeaderContentType, "application/pdf")
	c.Set(fiber.HeaderContentDisposition, `attachment; filename="`+filename+`"`)
	return c.Send(data)
}
