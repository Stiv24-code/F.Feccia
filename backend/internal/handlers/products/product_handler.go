package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"fratelli-feccia/internal/dto"
	"fratelli-feccia/internal/services"
	"fratelli-feccia/pkg/utils"
)

type ProductHandler struct {
	Service services.Product
}

func NewProductHandler(service services.Product) *ProductHandler {
	return &ProductHandler{Service: service}
}

// ListProducts godoc
// @Summary List products
// @Tags Products
// @Security BearerAuth
// @Produce json
// @Param search query string false "Filter by codice/descrizione"
// @Success 200 {array} dto.ProductResponse
// @Router /api/v1/products [get]
func (h *ProductHandler) ListProducts(c *fiber.Ctx) error {
	ctx := utils.RequestContext(c)
	items, err := h.Service.List(ctx, c.Query("search"))
	if err != nil {
		return utils.HandleDatabaseError(c, err)
	}
	return utils.SuccessResponse(c, 200, items)
}

// CreateProduct godoc
// @Summary Create product
// @Tags Products
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param product body dto.ProductRequest true "Product data"
// @Success 201 {object} dto.ProductResponse
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Router /api/v1/products [post]
func (h *ProductHandler) CreateProduct(c *fiber.Ctx) error {
	ctx := utils.RequestContext(c)
	var req dto.ProductRequest
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

// UpdateProduct godoc
// @Summary Update product (full replace)
// @Tags Products
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Product ID (UUID)"
// @Param product body dto.ProductRequest true "Product data"
// @Success 200 {object} dto.ProductResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/products/{id} [put]
func (h *ProductHandler) UpdateProduct(c *fiber.Ctx) error {
	ctx := utils.RequestContext(c)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.ErrorResponse(c, 400, "Invalid ID")
	}
	var req dto.ProductRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, 400, "Invalid request body")
	}
	if validationErrors := utils.NewValidator().Validate(&req); len(validationErrors) > 0 {
		return utils.ValidationErrorResponse(c, validationErrors)
	}
	item, err := h.Service.Update(ctx, id, req)
	if err != nil {
		return utils.HandleDatabaseError(c, err)
	}
	return utils.SuccessResponse(c, 200, item)
}

// DeleteProduct godoc
// @Summary Delete product (logical, sets active=false)
// @Tags Products
// @Security BearerAuth
// @Param id path string true "Product ID (UUID)"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]string
// @Router /api/v1/products/{id} [delete]
func (h *ProductHandler) DeleteProduct(c *fiber.Ctx) error {
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
