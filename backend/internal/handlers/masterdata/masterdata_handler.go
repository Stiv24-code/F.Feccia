package handlers

import (
	"github.com/gofiber/fiber/v2"

	"fratelli-feccia/internal/dto"
	"fratelli-feccia/internal/services"
	"fratelli-feccia/pkg/utils"
)

// MasterdataHandler groups vehicle-types/accessory-costs/transport-categories,
// mirroring backend/routers/masterdata.py's own rationale (list+create only,
// no update/delete for any of the three).
type MasterdataHandler struct {
	Service services.Masterdata
}

func NewMasterdataHandler(service services.Masterdata) *MasterdataHandler {
	return &MasterdataHandler{Service: service}
}

// @Summary List vehicle types
// @Tags Masterdata
// @Security BearerAuth
// @Produce json
// @Success 200 {array} dto.VehicleTypeResponse
// @Router /api/v1/vehicle-types [get]
func (h *MasterdataHandler) ListVehicleTypes(c *fiber.Ctx) error {
	items, err := h.Service.ListVehicleTypes(utils.RequestContext(c))
	if err != nil {
		return utils.HandleDatabaseError(c, err)
	}
	return utils.SuccessResponse(c, 200, items)
}

// @Summary Create vehicle type
// @Tags Masterdata
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param item body dto.VehicleTypeRequest true "Vehicle type data"
// @Success 201 {object} dto.VehicleTypeResponse
// @Failure 400 {object} map[string]string
// @Router /api/v1/vehicle-types [post]
func (h *MasterdataHandler) CreateVehicleType(c *fiber.Ctx) error {
	var req dto.VehicleTypeRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, 400, "Invalid request body")
	}
	if validationErrors := utils.NewValidator().Validate(&req); len(validationErrors) > 0 {
		return utils.ValidationErrorResponse(c, validationErrors)
	}
	item, err := h.Service.CreateVehicleType(utils.RequestContext(c), req)
	if err != nil {
		return utils.HandleDatabaseError(c, err)
	}
	return utils.SuccessResponse(c, 201, item)
}

// @Summary List accessory costs
// @Tags Masterdata
// @Security BearerAuth
// @Produce json
// @Success 200 {array} dto.AccessoryCostResponse
// @Router /api/v1/accessory-costs [get]
func (h *MasterdataHandler) ListAccessoryCosts(c *fiber.Ctx) error {
	items, err := h.Service.ListAccessoryCosts(utils.RequestContext(c))
	if err != nil {
		return utils.HandleDatabaseError(c, err)
	}
	return utils.SuccessResponse(c, 200, items)
}

// @Summary Create accessory cost
// @Tags Masterdata
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param item body dto.AccessoryCostRequest true "Accessory cost data"
// @Success 201 {object} dto.AccessoryCostResponse
// @Failure 400 {object} map[string]string
// @Router /api/v1/accessory-costs [post]
func (h *MasterdataHandler) CreateAccessoryCost(c *fiber.Ctx) error {
	var req dto.AccessoryCostRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, 400, "Invalid request body")
	}
	if validationErrors := utils.NewValidator().Validate(&req); len(validationErrors) > 0 {
		return utils.ValidationErrorResponse(c, validationErrors)
	}
	item, err := h.Service.CreateAccessoryCost(utils.RequestContext(c), req)
	if err != nil {
		return utils.HandleDatabaseError(c, err)
	}
	return utils.SuccessResponse(c, 201, item)
}

// @Summary List transport categories
// @Tags Masterdata
// @Security BearerAuth
// @Produce json
// @Success 200 {array} dto.TransportCategoryResponse
// @Router /api/v1/transport-categories [get]
func (h *MasterdataHandler) ListTransportCategories(c *fiber.Ctx) error {
	items, err := h.Service.ListTransportCategories(utils.RequestContext(c))
	if err != nil {
		return utils.HandleDatabaseError(c, err)
	}
	return utils.SuccessResponse(c, 200, items)
}

// @Summary Create transport category
// @Tags Masterdata
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param item body dto.TransportCategoryRequest true "Transport category data"
// @Success 201 {object} dto.TransportCategoryResponse
// @Failure 400 {object} map[string]string
// @Router /api/v1/transport-categories [post]
func (h *MasterdataHandler) CreateTransportCategory(c *fiber.Ctx) error {
	var req dto.TransportCategoryRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, 400, "Invalid request body")
	}
	if validationErrors := utils.NewValidator().Validate(&req); len(validationErrors) > 0 {
		return utils.ValidationErrorResponse(c, validationErrors)
	}
	item, err := h.Service.CreateTransportCategory(utils.RequestContext(c), req)
	if err != nil {
		return utils.HandleDatabaseError(c, err)
	}
	return utils.SuccessResponse(c, 201, item)
}
