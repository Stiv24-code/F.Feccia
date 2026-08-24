package handlers

import (
	"context"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"fratelli-feccia/internal/dto"
	"fratelli-feccia/internal/models"
	"fratelli-feccia/internal/services"
	"fratelli-feccia/pkg/utils"
)

type InboundOrderHandler struct {
	Service      services.InboundOrder
	Scraper      services.MailScraper
	Engine       services.PdfEngine
	Customers    services.Customer
	Destinations services.Destination
}

func NewInboundOrderHandler(service services.InboundOrder, scraper services.MailScraper, engine services.PdfEngine, customers services.Customer, destinations services.Destination) *InboundOrderHandler {
	return &InboundOrderHandler{Service: service, Scraper: scraper, Engine: engine, Customers: customers, Destinations: destinations}
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

// ConvertInboundOrder godoc
// @Summary Convert an inbound order into a real TMS order
// @Description Creates a models.Order from the draft and links the two (inbound_order.order_id), which makes the call idempotent: a second convert answers 409 with the existing order id. Separate from /accept, which only sends the confirmation mail — an order can be created before or after mailing the customer. cliente_id is required in the body unless the draft already carries a trusted one (portal submissions do); the free-text client field is never resolved to an anagrafica by name, so a spoofed sender cannot get an order billed to someone else. tariffa defaults to the customer's proposed rate for portal drafts and to 0 otherwise (a mail draft's rate is free text and is never parsed) — the response flags when the applied rate is the customer's own.
// @Tags InboundOrders
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Inbound order ID (UUID)"
// @Param payload body dto.InboundOrderConvertRequest true "Conversion overrides (all fields optional)"
// @Success 201 {object} dto.InboundOrderConvertResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Router /api/v1/inbound-orders/{id}/convert [post]
func (h *InboundOrderHandler) ConvertInboundOrder(c *fiber.Ctx) error {
	ctx := utils.RequestContext(c)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.ErrorResponse(c, 400, "Invalid ID")
	}
	// Body opzionale: convertire un draft da portale senza override e' il
	// caso normale, e un POST senza corpo non deve valere 400.
	var req dto.InboundOrderConvertRequest
	if len(c.Body()) > 0 {
		if err := c.BodyParser(&req); err != nil {
			return utils.ErrorResponse(c, 400, "Invalid request body")
		}
	}
	if validationErrors := utils.NewValidator().Validate(&req); len(validationErrors) > 0 {
		return utils.ValidationErrorResponse(c, validationErrors)
	}
	res, err := h.Service.Convert(ctx, id, req)
	if err != nil {
		return utils.HandleDatabaseError(c, err)
	}
	return utils.SuccessResponse(c, 201, res)
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

// CreateMyInboundOrder godoc
// @Summary Submit a new transport request as the logged-in client (self-service portal)
// @Description Creates a pending InboundOrder draft (source=portal) — visible to staff on /inbound-orders and to the client itself on GET /me/inbound-orders — instead of creating a live Order directly. An operator must accept it (same review step already applied to mail/PDF-sourced orders) before it becomes a plannable Order.
// @Tags Auth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param order body dto.ClientInboundOrderRequest true "Same shape as the internal order form (destinazione ids, tariffa, date/orari, note), plus product/kg as plain text (InboundOrder has no product master-data FK)"
// @Success 201 {object} dto.InboundOrderResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /api/v1/me/inbound-orders [post]
func (h *InboundOrderHandler) CreateMyInboundOrder(c *fiber.Ctx) error {
	ctx := utils.RequestContext(c)
	customerID, err := utils.RequestCustomerID(c)
	if err != nil {
		return utils.ErrorResponse(c, 401, "Account cliente non valido")
	}

	var req dto.ClientInboundOrderRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, 400, "Invalid request body")
	}

	customer, err := h.Customers.GetByID(ctx, customerID)
	if err != nil {
		return utils.HandleDatabaseError(c, err)
	}

	caricoID, caricoName := h.resolveDestination(ctx, req.DestinazioneCaricoID)
	scaricoID, scaricoName := h.resolveDestination(ctx, req.DestinazioneScaricoID)

	// InboundOrderRequest requires ref-or-product non-empty (Create's own
	// business rule) — rif_ordine_cliente is optional on the client form, so
	// when the client leaves Product blank too, a fallback description
	// guarantees that rule is met regardless of what was filled in.
	product := strings.TrimSpace(req.Product)
	if product == "" {
		product = "Richiesta di trasporto"
		if caricoName != "" || scaricoName != "" {
			product = fmt.Sprintf("Trasporto %s -> %s", firstNonEmpty(caricoName, "?"), firstNonEmpty(scaricoName, "?"))
		}
	}

	notesParts := make([]string, 0, 3)
	if req.Note != "" {
		notesParts = append(notesParts, req.Note)
	}
	if req.OraRitiroDa != "" || req.OraRitiroA != "" {
		notesParts = append(notesParts, fmt.Sprintf("Orario ritiro: %s-%s", req.OraRitiroDa, req.OraRitiroA))
	}
	if req.OraConsegnaDa != "" || req.OraConsegnaA != "" {
		notesParts = append(notesParts, fmt.Sprintf("Orario consegna: %s-%s", req.OraConsegnaDa, req.OraConsegnaA))
	}

	inboundReq := dto.InboundOrderRequest{
		Client:        customer.RagioneSociale,
		SenderEmail:   customer.Email,
		Ref:           req.RifOrdineCliente,
		Product:       product,
		Kg:            req.Kg,
		LoadDate:      req.DataRitiro,
		LoadPlace:     caricoName,
		DeliveryDate:  req.DataConsegna,
		DeliveryPlace: scaricoName,
		Rate:          fmt.Sprintf("€ %.2f", req.Tariffa),
		Notes:         strings.Join(notesParts, " | "),
		Source:        models.InboundOrderSourcePortal,
		ClienteID:     &customerID,

		// Payload strutturato: gli stessi dati salvati anche come testo
		// libero sopra, ma nella forma in cui sono arrivati, cosi'
		// InboundOrderService.Convert ricostruisce l'ordine senza dover
		// risalire da un nome a una FK. CommittenteID e' deliberatamente
		// escluso: il form del portale non lo espone, e accettarlo dal body
		// lascerebbe un cliente puntare l'anagrafica di chiunque altro come
		// parte ordinante — se serve, lo imposta l'operatore convertendo.
		DestinazioneCaricoID:  caricoID,
		DestinazioneScaricoID: scaricoID,
		OraRitiroDa:           req.OraRitiroDa,
		OraRitiroA:            req.OraRitiroA,
		OraConsegnaDa:         req.OraConsegnaDa,
		OraConsegnaA:          req.OraConsegnaA,
		// "Tariffa desiderata" del form: una proposta, non un prezzo. Resta
		// separata da Order.Tariffa fino a che un operatore non la conferma.
		TariffaProposta: req.Tariffa,
	}

	item, err := h.Service.Create(ctx, inboundReq)
	if err != nil {
		return utils.HandleDatabaseError(c, err)
	}
	return utils.SuccessResponse(c, 201, item)
}

// resolveDestination validates a destination id sent by the client portal,
// returning the id only when the row really exists (so a draft never stores
// a dangling reference for Convert to trip over) plus its name, which feeds
// the free-text LoadPlace/DeliveryPlace the acceptance dashboard renders.
func (h *InboundOrderHandler) resolveDestination(ctx context.Context, raw string) (*uuid.UUID, string) {
	id, err := uuid.Parse(raw)
	if err != nil {
		return nil, ""
	}
	d, err := h.Destinations.GetByID(ctx, id)
	if err != nil || d == nil {
		return nil, ""
	}
	return &id, d.Nome
}

func firstNonEmpty(v, fallback string) string {
	if v == "" {
		return fallback
	}
	return v
}

// ListMyInboundOrders godoc
// @Summary List the logged-in client's own pending/under-revision requests (self-service portal)
// @Description Only source=portal drafts still pending or under revision — once staff accepts one it becomes a normal Order and drops off this list (see GET /me/orders instead).
// @Tags Auth
// @Security BearerAuth
// @Produce json
// @Success 200 {array} dto.InboundOrderResponse
// @Failure 401 {object} map[string]string
// @Router /api/v1/me/inbound-orders [get]
func (h *InboundOrderHandler) ListMyInboundOrders(c *fiber.Ctx) error {
	customerID, err := utils.RequestCustomerID(c)
	if err != nil {
		return utils.ErrorResponse(c, 401, "Account cliente non valido")
	}
	items, err := h.Service.ListForClient(utils.RequestContext(c), customerID)
	if err != nil {
		return utils.HandleDatabaseError(c, err)
	}
	return utils.SuccessResponse(c, 200, items)
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
