package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"fratelli-feccia/internal/dto"
	"fratelli-feccia/internal/services"
	"fratelli-feccia/pkg/utils"
)

type CustomerHandler struct {
	Service services.Customer
}

func NewCustomerHandler(service services.Customer) *CustomerHandler {
	return &CustomerHandler{Service: service}
}

// ListCustomers godoc
// @Summary List customers
// @Description Returns active customers, optionally filtered by ragione sociale
// @Tags Customers
// @Security BearerAuth
// @Produce json
// @Param search query string false "Filter by ragione sociale (substring, case-insensitive)"
// @Success 200 {array} dto.CustomerResponse
// @Router /api/v1/customers [get]
func (h *CustomerHandler) ListCustomers(c *fiber.Ctx) error {
	ctx := utils.RequestContext(c)
	search := c.Query("search")

	customers, err := h.Service.List(ctx, search)
	if err != nil {
		return utils.HandleDatabaseError(c, err)
	}

	return utils.SuccessResponse(c, 200, customers)
}

// GetCustomerByID godoc
// @Summary Get customer by ID
// @Tags Customers
// @Security BearerAuth
// @Produce json
// @Param id path string true "Customer ID (UUID)"
// @Success 200 {object} dto.CustomerResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/customers/{id} [get]
func (h *CustomerHandler) GetCustomerByID(c *fiber.Ctx) error {
	ctx := utils.RequestContext(c)

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.ErrorResponse(c, 400, "Invalid ID")
	}

	customer, err := h.Service.GetByID(ctx, id)
	if err != nil {
		return utils.HandleDatabaseError(c, err)
	}
	if customer == nil {
		return utils.ErrorResponse(c, 404, "Cliente non trovato")
	}

	return utils.SuccessResponse(c, 200, customer)
}

// CreateCustomer godoc
// @Summary Create customer
// @Tags Customers
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param customer body dto.CustomerRequest true "Customer data"
// @Success 201 {object} dto.CustomerResponse
// @Failure 400 {object} map[string]string
// @Router /api/v1/customers [post]
func (h *CustomerHandler) CreateCustomer(c *fiber.Ctx) error {
	ctx := utils.RequestContext(c)

	var req dto.CustomerRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, 400, "Invalid request body")
	}
	if validationErrors := utils.NewValidator().Validate(&req); len(validationErrors) > 0 {
		return utils.ValidationErrorResponse(c, validationErrors)
	}

	customer, err := h.Service.Create(ctx, req)
	if err != nil {
		return utils.HandleDatabaseError(c, err)
	}

	return utils.SuccessResponse(c, 201, customer)
}

// UpdateCustomer godoc
// @Summary Update customer (full replace)
// @Tags Customers
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Customer ID (UUID)"
// @Param customer body dto.CustomerRequest true "Customer data"
// @Success 200 {object} dto.CustomerResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/customers/{id} [put]
func (h *CustomerHandler) UpdateCustomer(c *fiber.Ctx) error {
	ctx := utils.RequestContext(c)

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.ErrorResponse(c, 400, "Invalid ID")
	}

	var req dto.CustomerRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, 400, "Invalid request body")
	}
	if validationErrors := utils.NewValidator().Validate(&req); len(validationErrors) > 0 {
		return utils.ValidationErrorResponse(c, validationErrors)
	}

	customer, err := h.Service.Update(ctx, id, req)
	if err != nil {
		return utils.HandleDatabaseError(c, err)
	}

	return utils.SuccessResponse(c, 200, customer)
}

// DeleteCustomer godoc
// @Summary Delete customer (logical, sets active=false)
// @Tags Customers
// @Security BearerAuth
// @Param id path string true "Customer ID (UUID)"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]string
// @Router /api/v1/customers/{id} [delete]
func (h *CustomerHandler) DeleteCustomer(c *fiber.Ctx) error {
	ctx := utils.RequestContext(c)

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.ErrorResponse(c, 400, "Invalid ID")
	}

	if err := h.Service.Delete(ctx, id); err != nil {
		return utils.HandleDatabaseError(c, err)
	}

	return c.SendStatus(204)
}
