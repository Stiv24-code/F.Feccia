package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"fratelli-feccia/internal/dto"
	"fratelli-feccia/internal/services"
	"fratelli-feccia/pkg/utils"
)

type PriceListHandler struct {
	Service services.PriceList
}

func NewPriceListHandler(service services.PriceList) *PriceListHandler {
	return &PriceListHandler{Service: service}
}

// @Summary List pricelists
// @Tags PriceLists
// @Security BearerAuth
// @Produce json
// @Param cliente_id query string false "Filter by cliente_id"
// @Success 200 {array} dto.PriceListResponse
// @Router /api/v1/pricelists [get]
func (h *PriceListHandler) ListPriceLists(c *fiber.Ctx) error {
	items, err := h.Service.List(utils.RequestContext(c), c.Query("cliente_id"))
	if err != nil {
		return utils.HandleDatabaseError(c, err)
	}
	return utils.SuccessResponse(c, 200, items)
}

// @Summary Create pricelist
// @Tags PriceLists
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param pricelist body dto.PriceListRequest true "Pricelist data"
// @Success 201 {object} dto.PriceListResponse
// @Router /api/v1/pricelists [post]
func (h *PriceListHandler) CreatePriceList(c *fiber.Ctx) error {
	var req dto.PriceListRequest
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

// @Summary Lookup the best matching tariff for an order context
// @Tags PriceLists
// @Security BearerAuth
// @Produce json
// @Param cliente_id query string true "Cliente ID"
// @Param carico_id query string false "Destinazione carico ID"
// @Param scarico_id query string false "Destinazione scarico ID"
// @Param prodotto_id query string false "Prodotto ID"
// @Param peso query number false "Peso (Kg)"
// @Success 200 {object} dto.TariffLookupResult
// @Router /api/v1/pricelists/lookup-tariff [get]
func (h *PriceListHandler) LookupTariff(c *fiber.Ctx) error {
	result, err := h.Service.LookupTariff(
		utils.RequestContext(c), c.Query("cliente_id"), c.Query("carico_id"),
		c.Query("scarico_id"), c.Query("prodotto_id"), c.QueryFloat("peso", 0),
	)
	if err != nil {
		return utils.HandleDatabaseError(c, err)
	}
	return utils.SuccessResponse(c, 200, result)
}

// @Summary Get pricelist by ID
// @Tags PriceLists
// @Security BearerAuth
// @Produce json
// @Param id path string true "Pricelist ID (UUID)"
// @Success 200 {object} dto.PriceListResponse
// @Failure 404 {object} map[string]string
// @Router /api/v1/pricelists/{id} [get]
func (h *PriceListHandler) GetPriceListByID(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.ErrorResponse(c, 400, "Invalid ID")
	}
	item, err := h.Service.GetByID(utils.RequestContext(c), id)
	if err != nil {
		return utils.HandleDatabaseError(c, err)
	}
	if item == nil {
		return utils.ErrorResponse(c, 404, "Listino non trovato")
	}
	return utils.SuccessResponse(c, 200, item)
}

// @Summary Update pricelist (duplicates if in_uso, else in-place)
// @Tags PriceLists
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Pricelist ID (UUID)"
// @Param pricelist body dto.PriceListRequest true "Pricelist data"
// @Success 200 {object} dto.PriceListUpdateResult
// @Failure 404 {object} map[string]string
// @Router /api/v1/pricelists/{id} [put]
func (h *PriceListHandler) UpdatePriceList(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.ErrorResponse(c, 400, "Invalid ID")
	}
	var req dto.PriceListRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, 400, "Invalid request body")
	}
	result, err := h.Service.Update(utils.RequestContext(c), id, req)
	if err != nil {
		return utils.HandleDatabaseError(c, err)
	}
	return utils.SuccessResponse(c, 200, result)
}

// @Summary Delete pricelist (logical, sets active=false)
// @Tags PriceLists
// @Security BearerAuth
// @Param id path string true "Pricelist ID (UUID)"
// @Success 204 "No Content"
// @Router /api/v1/pricelists/{id} [delete]
func (h *PriceListHandler) DeletePriceList(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.ErrorResponse(c, 400, "Invalid ID")
	}
	if err := h.Service.Delete(utils.RequestContext(c), id); err != nil {
		return utils.HandleDatabaseError(c, err)
	}
	return c.SendStatus(204)
}

// @Summary Add a tariff rule to a pricelist
// @Tags PriceLists
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Pricelist ID (UUID)"
// @Param item body dto.PriceListItemRequestDTO true "Item data"
// @Success 200 {object} dto.PriceListItemAddResult
// @Failure 404 {object} map[string]string
// @Router /api/v1/pricelists/{id}/items [post]
func (h *PriceListHandler) AddItem(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.ErrorResponse(c, 400, "Invalid ID")
	}
	var item dto.PriceListItemRequestDTO
	if err := c.BodyParser(&item); err != nil {
		return utils.ErrorResponse(c, 400, "Invalid request body")
	}
	result, err := h.Service.AddItem(utils.RequestContext(c), id, item)
	if err != nil {
		return utils.HandleDatabaseError(c, err)
	}
	return utils.SuccessResponse(c, 200, result)
}

// @Summary Update a tariff rule
// @Tags PriceLists
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Pricelist ID (UUID)"
// @Param item_id path string true "Item ID (UUID)"
// @Param item body dto.PriceListItemRequestDTO true "Item data"
// @Success 200 {object} dto.PriceListItemUpdateResult
// @Failure 404 {object} map[string]string
// @Router /api/v1/pricelists/{id}/items/{item_id} [put]
func (h *PriceListHandler) UpdateItem(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.ErrorResponse(c, 400, "Invalid ID")
	}
	itemID, err := uuid.Parse(c.Params("item_id"))
	if err != nil {
		return utils.ErrorResponse(c, 400, "Invalid item_id")
	}
	var item dto.PriceListItemRequestDTO
	if err := c.BodyParser(&item); err != nil {
		return utils.ErrorResponse(c, 400, "Invalid request body")
	}
	result, err := h.Service.UpdateItem(utils.RequestContext(c), id, itemID, item)
	if err != nil {
		return utils.HandleDatabaseError(c, err)
	}
	return utils.SuccessResponse(c, 200, result)
}

// @Summary Delete a tariff rule
// @Tags PriceLists
// @Security BearerAuth
// @Produce json
// @Param id path string true "Pricelist ID (UUID)"
// @Param item_id path string true "Item ID (UUID)"
// @Success 200 {object} dto.PriceListItemDeleteResult
// @Failure 404 {object} map[string]string
// @Router /api/v1/pricelists/{id}/items/{item_id} [delete]
func (h *PriceListHandler) DeleteItem(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.ErrorResponse(c, 400, "Invalid ID")
	}
	itemID, err := uuid.Parse(c.Params("item_id"))
	if err != nil {
		return utils.ErrorResponse(c, 400, "Invalid item_id")
	}
	result, err := h.Service.DeleteItem(utils.RequestContext(c), id, itemID)
	if err != nil {
		return utils.HandleDatabaseError(c, err)
	}
	return utils.SuccessResponse(c, 200, result)
}
