package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"fratelli-feccia/internal/dto"
	"fratelli-feccia/internal/services"
	"fratelli-feccia/pkg/utils"
)

type InvoiceHandler struct {
	Service services.Invoice
}

func NewInvoiceHandler(service services.Invoice) *InvoiceHandler {
	return &InvoiceHandler{Service: service}
}

// @Summary List invoices
// @Tags Invoices
// @Security BearerAuth
// @Produce json
// @Param stato query string false "Filter by stato"
// @Param cliente_id query string false "Filter by cliente_id"
// @Success 200 {array} dto.InvoiceResponse
// @Router /api/v1/invoices [get]
func (h *InvoiceHandler) ListInvoices(c *fiber.Ctx) error {
	items, err := h.Service.List(utils.RequestContext(c), c.Query("stato"), c.Query("cliente_id"))
	if err != nil {
		return utils.HandleDatabaseError(c, err)
	}
	return utils.SuccessResponse(c, 200, items)
}

// @Summary Create invoice (PROFORMA)
// @Tags Invoices
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param invoice body dto.InvoiceRequest true "Invoice data"
// @Success 201 {object} dto.InvoiceResponse
// @Router /api/v1/invoices [post]
func (h *InvoiceHandler) CreateInvoice(c *fiber.Ctx) error {
	var req dto.InvoiceRequest
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

// @Summary Get invoice by ID
// @Tags Invoices
// @Security BearerAuth
// @Produce json
// @Param id path string true "Invoice ID (UUID)"
// @Success 200 {object} dto.InvoiceResponse
// @Failure 404 {object} map[string]string
// @Router /api/v1/invoices/{id} [get]
func (h *InvoiceHandler) GetInvoiceByID(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.ErrorResponse(c, 400, "Invalid ID")
	}
	item, err := h.Service.GetByID(utils.RequestContext(c), id)
	if err != nil {
		return utils.HandleDatabaseError(c, err)
	}
	if item == nil {
		return utils.ErrorResponse(c, 404, "Fattura non trovata")
	}
	return utils.SuccessResponse(c, 200, item)
}

// @Summary Finalize invoice (PROFORMA -> DEFINITIVA)
// @Tags Invoices
// @Security BearerAuth
// @Produce json
// @Param id path string true "Invoice ID (UUID)"
// @Success 200 {object} dto.InvoiceFinalizeResult
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/invoices/{id}/finalize [patch]
func (h *InvoiceHandler) FinalizeInvoice(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.ErrorResponse(c, 400, "Invalid ID")
	}
	result, err := h.Service.Finalize(utils.RequestContext(c), id)
	if err != nil {
		return utils.HandleDatabaseError(c, err)
	}
	return utils.SuccessResponse(c, 200, result)
}

// @Summary Delete invoice (only PROFORMA, hard delete)
// @Tags Invoices
// @Security BearerAuth
// @Param id path string true "Invoice ID (UUID)"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/invoices/{id} [delete]
func (h *InvoiceHandler) DeleteInvoice(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.ErrorResponse(c, 400, "Invalid ID")
	}
	if err := h.Service.Delete(utils.RequestContext(c), id); err != nil {
		return utils.HandleDatabaseError(c, err)
	}
	return c.SendStatus(204)
}

// @Summary Invoice PDF (S3 archived copy if available, else generated on the fly)
// @Tags Invoices
// @Security BearerAuth
// @Produce application/pdf
// @Param id path string true "Invoice ID (UUID)"
// @Success 200 {file} binary
// @Failure 404 {object} map[string]string
// @Router /api/v1/invoices/{id}/pdf [get]
func (h *InvoiceHandler) GetInvoicePDF(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.ErrorResponse(c, 400, "Invalid ID")
	}
	data, filename, err := h.Service.GetPDF(utils.RequestContext(c), id)
	if err != nil {
		return utils.HandleDatabaseError(c, err)
	}
	c.Set(fiber.HeaderContentType, "application/pdf")
	c.Set(fiber.HeaderContentDisposition, `attachment; filename="`+filename+`"`)
	return c.Send(data)
}

// @Summary Presigned S3 URL for the invoice PDF (DEFINITIVA + archived only)
// @Tags Invoices
// @Security BearerAuth
// @Produce json
// @Param id path string true "Invoice ID (UUID)"
// @Success 200 {object} dto.InvoicePDFURLResult
// @Failure 404 {object} map[string]string
// @Router /api/v1/invoices/{id}/pdf-url [get]
func (h *InvoiceHandler) GetInvoicePDFURL(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.ErrorResponse(c, 400, "Invalid ID")
	}
	result, err := h.Service.GetPDFPresignedURL(utils.RequestContext(c), id)
	if err != nil {
		return utils.HandleDatabaseError(c, err)
	}
	return utils.SuccessResponse(c, 200, result)
}
