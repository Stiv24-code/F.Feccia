package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"fratelli-feccia/internal/dto"
	"fratelli-feccia/internal/services"
	"fratelli-feccia/pkg/utils"
)

// AnagraficheHandler groups countries/banks/accounting-entries, mirroring
// backend/routers/anagrafiche_extra.py's own rationale.
type AnagraficheHandler struct {
	Service services.Anagrafiche
}

func NewAnagraficheHandler(service services.Anagrafiche) *AnagraficheHandler {
	return &AnagraficheHandler{Service: service}
}

// ── Countries ────────────────────────────────────────────────────────────

// @Summary List countries
// @Tags Anagrafiche
// @Security BearerAuth
// @Produce json
// @Param search query string false "Filter by nome/iso2"
// @Param page query int false "Page number (default 1)"
// @Param limit query int false "Items per page (default 20, max 100)"
// @Success 200 {object} dto.CountryListResponse
// @Router /api/v1/countries [get]
func (h *AnagraficheHandler) ListCountries(c *fiber.Ctx) error {
	items, total, err := h.Service.ListCountries(utils.RequestContext(c), c.Query("search"), utils.ParsePageParams(c))
	if err != nil {
		return utils.HandleDatabaseError(c, err)
	}
	return utils.SuccessResponse(c, 200, dto.CountryListResponse{Data: items, Total: total})
}

// @Summary Create country
// @Tags Anagrafiche
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param country body dto.CountryRequest true "Country data"
// @Success 201 {object} dto.CountryResponse
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Router /api/v1/countries [post]
func (h *AnagraficheHandler) CreateCountry(c *fiber.Ctx) error {
	var req dto.CountryRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, 400, "Invalid request body")
	}
	if validationErrors := utils.NewValidator().Validate(&req); len(validationErrors) > 0 {
		return utils.ValidationErrorResponse(c, validationErrors)
	}
	item, err := h.Service.CreateCountry(utils.RequestContext(c), req)
	if err != nil {
		return utils.HandleDatabaseError(c, err)
	}
	return utils.SuccessResponse(c, 201, item)
}

// @Summary Update country (full replace)
// @Tags Anagrafiche
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Country ID (UUID)"
// @Param country body dto.CountryRequest true "Country data"
// @Success 200 {object} dto.CountryResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/countries/{id} [put]
func (h *AnagraficheHandler) UpdateCountry(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.ErrorResponse(c, 400, "Invalid ID")
	}
	var req dto.CountryRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, 400, "Invalid request body")
	}
	if validationErrors := utils.NewValidator().Validate(&req); len(validationErrors) > 0 {
		return utils.ValidationErrorResponse(c, validationErrors)
	}
	item, err := h.Service.UpdateCountry(utils.RequestContext(c), id, req)
	if err != nil {
		return utils.HandleDatabaseError(c, err)
	}
	return utils.SuccessResponse(c, 200, item)
}

// @Summary Delete country (logical, sets active=false)
// @Tags Anagrafiche
// @Security BearerAuth
// @Param id path string true "Country ID (UUID)"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]string
// @Router /api/v1/countries/{id} [delete]
func (h *AnagraficheHandler) DeleteCountry(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.ErrorResponse(c, 400, "Invalid ID")
	}
	if err := h.Service.DeleteCountry(utils.RequestContext(c), id); err != nil {
		return utils.HandleDatabaseError(c, err)
	}
	return c.SendStatus(204)
}

// ── Banks ────────────────────────────────────────────────────────────────

// @Summary List banks
// @Tags Anagrafiche
// @Security BearerAuth
// @Produce json
// @Param search query string false "Filter by nome/bic_swift"
// @Param page query int false "Page number (default 1)"
// @Param limit query int false "Items per page (default 20, max 100)"
// @Success 200 {object} dto.BankListResponse
// @Router /api/v1/banks [get]
func (h *AnagraficheHandler) ListBanks(c *fiber.Ctx) error {
	items, total, err := h.Service.ListBanks(utils.RequestContext(c), c.Query("search"), utils.ParsePageParams(c))
	if err != nil {
		return utils.HandleDatabaseError(c, err)
	}
	return utils.SuccessResponse(c, 200, dto.BankListResponse{Data: items, Total: total})
}

// @Summary Create bank
// @Tags Anagrafiche
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param bank body dto.BankRequest true "Bank data"
// @Success 201 {object} dto.BankResponse
// @Failure 400 {object} map[string]string
// @Router /api/v1/banks [post]
func (h *AnagraficheHandler) CreateBank(c *fiber.Ctx) error {
	var req dto.BankRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, 400, "Invalid request body")
	}
	if validationErrors := utils.NewValidator().Validate(&req); len(validationErrors) > 0 {
		return utils.ValidationErrorResponse(c, validationErrors)
	}
	item, err := h.Service.CreateBank(utils.RequestContext(c), req)
	if err != nil {
		return utils.HandleDatabaseError(c, err)
	}
	return utils.SuccessResponse(c, 201, item)
}

// @Summary Update bank (full replace)
// @Tags Anagrafiche
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Bank ID (UUID)"
// @Param bank body dto.BankRequest true "Bank data"
// @Success 200 {object} dto.BankResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/banks/{id} [put]
func (h *AnagraficheHandler) UpdateBank(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.ErrorResponse(c, 400, "Invalid ID")
	}
	var req dto.BankRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, 400, "Invalid request body")
	}
	if validationErrors := utils.NewValidator().Validate(&req); len(validationErrors) > 0 {
		return utils.ValidationErrorResponse(c, validationErrors)
	}
	item, err := h.Service.UpdateBank(utils.RequestContext(c), id, req)
	if err != nil {
		return utils.HandleDatabaseError(c, err)
	}
	return utils.SuccessResponse(c, 200, item)
}

// @Summary Delete bank (logical, sets active=false)
// @Tags Anagrafiche
// @Security BearerAuth
// @Param id path string true "Bank ID (UUID)"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]string
// @Router /api/v1/banks/{id} [delete]
func (h *AnagraficheHandler) DeleteBank(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.ErrorResponse(c, 400, "Invalid ID")
	}
	if err := h.Service.DeleteBank(utils.RequestContext(c), id); err != nil {
		return utils.HandleDatabaseError(c, err)
	}
	return c.SendStatus(204)
}

// ── Accounting entries ──────────────────────────────────────────────────

// @Summary List accounting entries
// @Tags Anagrafiche
// @Security BearerAuth
// @Produce json
// @Param search query string false "Filter by codice/descrizione"
// @Param tipo query string false "Filter by tipo (ricavo|costo)"
// @Param page query int false "Page number (default 1)"
// @Param limit query int false "Items per page (default 20, max 100)"
// @Success 200 {object} dto.AccountingEntryListResponse
// @Router /api/v1/accounting-entries [get]
func (h *AnagraficheHandler) ListAccountingEntries(c *fiber.Ctx) error {
	items, total, err := h.Service.ListAccountingEntries(utils.RequestContext(c), c.Query("search"), c.Query("tipo"), utils.ParsePageParams(c))
	if err != nil {
		return utils.HandleDatabaseError(c, err)
	}
	return utils.SuccessResponse(c, 200, dto.AccountingEntryListResponse{Data: items, Total: total})
}

// @Summary Create accounting entry
// @Tags Anagrafiche
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param entry body dto.AccountingEntryRequest true "Accounting entry data"
// @Success 201 {object} dto.AccountingEntryResponse
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Router /api/v1/accounting-entries [post]
func (h *AnagraficheHandler) CreateAccountingEntry(c *fiber.Ctx) error {
	var req dto.AccountingEntryRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, 400, "Invalid request body")
	}
	if validationErrors := utils.NewValidator().Validate(&req); len(validationErrors) > 0 {
		return utils.ValidationErrorResponse(c, validationErrors)
	}
	item, err := h.Service.CreateAccountingEntry(utils.RequestContext(c), req)
	if err != nil {
		return utils.HandleDatabaseError(c, err)
	}
	return utils.SuccessResponse(c, 201, item)
}

// @Summary Update accounting entry (full replace)
// @Tags Anagrafiche
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Accounting entry ID (UUID)"
// @Param entry body dto.AccountingEntryRequest true "Accounting entry data"
// @Success 200 {object} dto.AccountingEntryResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/accounting-entries/{id} [put]
func (h *AnagraficheHandler) UpdateAccountingEntry(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.ErrorResponse(c, 400, "Invalid ID")
	}
	var req dto.AccountingEntryRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, 400, "Invalid request body")
	}
	if validationErrors := utils.NewValidator().Validate(&req); len(validationErrors) > 0 {
		return utils.ValidationErrorResponse(c, validationErrors)
	}
	item, err := h.Service.UpdateAccountingEntry(utils.RequestContext(c), id, req)
	if err != nil {
		return utils.HandleDatabaseError(c, err)
	}
	return utils.SuccessResponse(c, 200, item)
}

// @Summary Delete accounting entry (logical, sets active=false)
// @Tags Anagrafiche
// @Security BearerAuth
// @Param id path string true "Accounting entry ID (UUID)"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]string
// @Router /api/v1/accounting-entries/{id} [delete]
func (h *AnagraficheHandler) DeleteAccountingEntry(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.ErrorResponse(c, 400, "Invalid ID")
	}
	if err := h.Service.DeleteAccountingEntry(utils.RequestContext(c), id); err != nil {
		return utils.HandleDatabaseError(c, err)
	}
	return c.SendStatus(204)
}
