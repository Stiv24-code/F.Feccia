package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"fratelli-feccia/internal/dto"
	"fratelli-feccia/internal/services"
	"fratelli-feccia/pkg/utils"
)

type InboundOrderHandler struct {
	Service services.InboundOrder
	Scraper services.MailScraper
	Engine  services.PdfEngine
}

func NewInboundOrderHandler(service services.InboundOrder, scraper services.MailScraper, engine services.PdfEngine) *InboundOrderHandler {
	return &InboundOrderHandler{Service: service, Scraper: scraper, Engine: engine}
}

// ListInboundOrders godoc
// @Summary List inbound orders (acceptance dashboard)
// @Tags InboundOrders
// @Security BearerAuth
// @Produce json
// @Success 200 {array} dto.InboundOrderResponse
// @Router /api/v1/inbound-orders [get]
func (h *InboundOrderHandler) ListInboundOrders(c *fiber.Ctx) error {
	ctx := utils.RequestContext(c)
	items, err := h.Service.List(ctx)
	if err != nil {
		return utils.HandleDatabaseError(c, err)
	}
	return utils.SuccessResponse(c, 200, items)
}

// CreateInboundOrder godoc
// @Summary Confirm an inbound-order draft (e.g. from /pdf/import)
// @Description 409 when an order with the same (ref, client) already exists — mailbox re-reads and double submissions never duplicate.
// @Tags InboundOrders
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param order body dto.InboundOrderRequest true "Order data"
// @Success 201 {object} dto.InboundOrderResponse
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Router /api/v1/inbound-orders [post]
func (h *InboundOrderHandler) CreateInboundOrder(c *fiber.Ctx) error {
	ctx := utils.RequestContext(c)
	var req dto.InboundOrderRequest
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

// AcceptInboundOrder godoc
// @Summary Accept an inbound order (+ confirmation mail when SMTP is configured)
// @Description The mail is sent BEFORE the status change: on send failure (502) the order stays pending for a retry. Without SMTP the order is accepted anyway and the response says no mail was sent.
// @Tags InboundOrders
// @Security BearerAuth
// @Produce json
// @Param id path string true "Order ID (UUID)"
// @Success 200 {object} dto.InboundOrderActionResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 502 {object} map[string]string
// @Router /api/v1/inbound-orders/{id}/accept [post]
func (h *InboundOrderHandler) AcceptInboundOrder(c *fiber.Ctx) error {
	ctx := utils.RequestContext(c)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.ErrorResponse(c, 400, "Invalid ID")
	}
	res, err := h.Service.Accept(ctx, id)
	if err != nil {
		return utils.HandleDatabaseError(c, err)
	}
	return utils.SuccessResponse(c, 200, res)
}

// ModifyInboundOrder godoc
// @Summary Mark an inbound order as under revision (the UI opens a mailto: to the sender)
// @Tags InboundOrders
// @Security BearerAuth
// @Produce json
// @Param id path string true "Order ID (UUID)"
// @Success 200 {object} dto.InboundOrderResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/inbound-orders/{id}/modify [post]
func (h *InboundOrderHandler) ModifyInboundOrder(c *fiber.Ctx) error {
	ctx := utils.RequestContext(c)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.ErrorResponse(c, 400, "Invalid ID")
	}
	item, err := h.Service.Modify(ctx, id)
	if err != nil {
		return utils.HandleDatabaseError(c, err)
	}
	return utils.SuccessResponse(c, 200, item)
}

// ScrapeInboundOrders godoc
// @Summary Read the orders mailbox now and store the new orders
// @Description Runs one scrape of the configured mailbox (Microsoft Graph or IMAP). 502 when the mailbox is unreachable or not authenticated.
// @Tags InboundOrders
// @Security BearerAuth
// @Produce json
// @Success 200 {object} dto.InboundScrapeResponse
// @Failure 502 {object} map[string]string
// @Router /api/v1/inbound-orders/scrape [post]
func (h *InboundOrderHandler) ScrapeInboundOrders(c *fiber.Ctx) error {
	ctx := utils.RequestContext(c)
	added, scanned, err := h.Scraper.Scrape(ctx)
	if err != nil {
		return utils.ErrorResponse(c, 502, err.Error())
	}
	return utils.SuccessResponse(c, 200, dto.InboundScrapeResponse{Added: added, Scanned: scanned})
}

// GetInboundConfig godoc
// @Summary Runtime status of the inbound pipeline (SMTP, mailbox, poppler, vision)
// @Tags InboundOrders
// @Security BearerAuth
// @Produce json
// @Success 200 {object} dto.InboundConfigResponse
// @Router /api/v1/inbound-config [get]
func (h *InboundOrderHandler) GetInboundConfig(c *fiber.Ctx) error {
	status := h.Scraper.Status()
	status.PdfReady = h.Engine.Ready()
	status.VisionReady = h.Engine.VisionReady()
	return utils.SuccessResponse(c, 200, status)
}

// ResetInboundOrder godoc
// @Summary Reset an inbound order to pending
// @Tags InboundOrders
// @Security BearerAuth
// @Produce json
// @Param id path string true "Order ID (UUID)"
// @Success 200 {object} dto.InboundOrderResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/inbound-orders/{id}/reset [post]
func (h *InboundOrderHandler) ResetInboundOrder(c *fiber.Ctx) error {
	ctx := utils.RequestContext(c)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.ErrorResponse(c, 400, "Invalid ID")
	}
	item, err := h.Service.Reset(ctx, id)
	if err != nil {
		return utils.HandleDatabaseError(c, err)
	}
	return utils.SuccessResponse(c, 200, item)
}
