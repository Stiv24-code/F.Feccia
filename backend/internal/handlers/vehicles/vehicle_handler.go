package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"fratelli-feccia/internal/dto"
	"fratelli-feccia/internal/services"
	"fratelli-feccia/pkg/utils"
)

type VehicleHandler struct {
	Service services.Vehicle
}

func NewVehicleHandler(service services.Vehicle) *VehicleHandler {
	return &VehicleHandler{Service: service}
}

// ── CRUD ─────────────────────────────────────────────────────────────────

// @Summary List vehicles
// @Tags Vehicles
// @Security BearerAuth
// @Produce json
// @Param search query string false "Filter by targa"
// @Success 200 {array} dto.VehicleResponse
// @Router /api/v1/vehicles [get]
func (h *VehicleHandler) ListVehicles(c *fiber.Ctx) error {
	items, err := h.Service.List(utils.RequestContext(c), c.Query("search"))
	if err != nil {
		return utils.HandleDatabaseError(c, err)
	}
	return utils.SuccessResponse(c, 200, items)
}

// @Summary Create vehicle
// @Tags Vehicles
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param vehicle body dto.VehicleRequest true "Vehicle data"
// @Success 201 {object} dto.VehicleResponse
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Router /api/v1/vehicles [post]
func (h *VehicleHandler) CreateVehicle(c *fiber.Ctx) error {
	var req dto.VehicleRequest
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

// @Summary Update vehicle (full replace of the create-able fields)
// @Tags Vehicles
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Vehicle ID (UUID)"
// @Param vehicle body dto.VehicleRequest true "Vehicle data"
// @Success 200 {object} dto.VehicleResponse
// @Failure 404 {object} map[string]string
// @Router /api/v1/vehicles/{id} [put]
func (h *VehicleHandler) UpdateVehicle(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.ErrorResponse(c, 400, "Invalid ID")
	}
	var req dto.VehicleRequest
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

// @Summary Delete vehicle (logical, sets active=false)
// @Tags Vehicles
// @Security BearerAuth
// @Param id path string true "Vehicle ID (UUID)"
// @Success 204 "No Content"
// @Router /api/v1/vehicles/{id} [delete]
func (h *VehicleHandler) DeleteVehicle(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.ErrorResponse(c, 400, "Invalid ID")
	}
	if err := h.Service.Delete(utils.RequestContext(c), id); err != nil {
		return utils.HandleDatabaseError(c, err)
	}
	return c.SendStatus(204)
}

// ── GPS ──────────────────────────────────────────────────────────────────

// @Summary Update vehicle GPS position (by id or targa)
// @Tags Vehicles
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Vehicle ID (UUID) or targa"
// @Param position body dto.VehicleGPSUpdateRequest true "GPS position"
// @Success 200 {object} dto.GPSUpdateResult
// @Failure 404 {object} map[string]string
// @Router /api/v1/vehicles/{id}/gps-position [post]
func (h *VehicleHandler) UpdateVehicleGPS(c *fiber.Ctx) error {
	var req dto.VehicleGPSUpdateRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, 400, "Invalid request body")
	}
	result, err := h.Service.UpdateGPSByID(utils.RequestContext(c), c.Params("id"), req)
	if err != nil {
		return utils.HandleDatabaseError(c, err)
	}
	return utils.SuccessResponse(c, 200, result)
}

// @Summary Update vehicle GPS position by targa
// @Tags Vehicles
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param targa path string true "Vehicle targa"
// @Param position body dto.VehicleGPSUpdateRequest true "GPS position"
// @Success 200 {object} dto.GPSUpdateResult
// @Failure 404 {object} map[string]string
// @Router /api/v1/vehicles/gps-position-by-plate/{targa} [post]
func (h *VehicleHandler) UpdateVehicleGPSByPlate(c *fiber.Ctx) error {
	var req dto.VehicleGPSUpdateRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, 400, "Invalid request body")
	}
	result, err := h.Service.UpdateGPSByPlate(utils.RequestContext(c), c.Params("targa"), req)
	if err != nil {
		return utils.HandleDatabaseError(c, err)
	}
	return utils.SuccessResponse(c, 200, result)
}

// @Summary GPS history for a vehicle
// @Tags Vehicles
// @Security BearerAuth
// @Produce json
// @Param id path string true "Vehicle ID (UUID) or targa"
// @Param limit query int false "Max results (default 100)"
// @Success 200 {array} dto.GPSHistoryResponse
// @Router /api/v1/vehicles/{id}/gps-history [get]
func (h *VehicleHandler) GetVehicleGPSHistory(c *fiber.Ctx) error {
	items, err := h.Service.GetGPSHistory(utils.RequestContext(c), c.Params("id"), c.QueryInt("limit", 100))
	if err != nil {
		return utils.HandleDatabaseError(c, err)
	}
	return utils.SuccessResponse(c, 200, items)
}

// @Summary Live GPS positions for all active vehicles
// @Tags Vehicles
// @Security BearerAuth
// @Produce json
// @Success 200 {array} dto.GPSLiveVehicle
// @Router /api/v1/vehicles/gps-live [get]
func (h *VehicleHandler) GetAllGPSLive(c *fiber.Ctx) error {
	items, err := h.Service.GetAllGPSLive(utils.RequestContext(c))
	if err != nil {
		return utils.HandleDatabaseError(c, err)
	}
	return utils.SuccessResponse(c, 200, items)
}

// ── Webhooks (public, token-gated — see pkg/middleware.RequireWebhookToken) ──

// @Summary GPS provider webhook ingestion
// @Tags Vehicles
// @Accept json
// @Produce json
// @Param vendor path string true "GPS vendor"
// @Param payload body dto.GPSWebhookPayload true "Normalized GPS payload"
// @Success 200 {object} dto.GPSUpdateResult
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/webhooks/gps/{vendor} [post]
func (h *VehicleHandler) GPSWebhook(c *fiber.Ctx) error {
	var payload dto.GPSWebhookPayload
	if err := c.BodyParser(&payload); err != nil {
		return utils.ErrorResponse(c, 400, "Invalid request body")
	}
	result, err := h.Service.IngestGPSWebhook(utils.RequestContext(c), c.Params("vendor"), payload)
	if err != nil {
		return utils.HandleDatabaseError(c, err)
	}
	return utils.SuccessResponse(c, 200, result)
}

// ── Temperature ──────────────────────────────────────────────────────────

// @Summary Temperature sensor webhook ingestion
// @Tags Vehicles
// @Accept json
// @Produce json
// @Param vendor path string true "Sensor vendor"
// @Param payload body dto.TemperatureWebhookRequest true "Temperature reading"
// @Success 200 {object} dto.TemperatureWebhookResult
// @Failure 404 {object} map[string]string
// @Router /api/v1/webhooks/temperature/{vendor} [post]
func (h *VehicleHandler) TemperatureWebhook(c *fiber.Ctx) error {
	var payload dto.TemperatureWebhookRequest
	if err := c.BodyParser(&payload); err != nil {
		return utils.ErrorResponse(c, 400, "Invalid request body")
	}
	if validationErrors := utils.NewValidator().Validate(&payload); len(validationErrors) > 0 {
		return utils.ValidationErrorResponse(c, validationErrors)
	}
	result, err := h.Service.IngestTemperatureWebhook(utils.RequestContext(c), c.Params("vendor"), payload)
	if err != nil {
		return utils.HandleDatabaseError(c, err)
	}
	return utils.SuccessResponse(c, 200, result)
}

// @Summary Temperature history for a vehicle
// @Tags Vehicles
// @Security BearerAuth
// @Produce json
// @Param id path string true "Vehicle ID (UUID)"
// @Param limit query int false "Max results (default 200)"
// @Param only_alerts query bool false "Only out-of-range readings"
// @Success 200 {array} dto.TemperatureReadingResponse
// @Router /api/v1/vehicles/{id}/temperature [get]
func (h *VehicleHandler) GetVehicleTemperature(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.ErrorResponse(c, 400, "Invalid ID")
	}
	items, err := h.Service.GetTemperatureHistory(utils.RequestContext(c), id, c.QueryInt("limit", 200), c.QueryBool("only_alerts", false))
	if err != nil {
		return utils.HandleDatabaseError(c, err)
	}
	return utils.SuccessResponse(c, 200, items)
}

// @Summary Set temperature thresholds for a vehicle
// @Tags Vehicles
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Vehicle ID (UUID)"
// @Param thresholds body dto.TemperatureThresholdsRequest true "Thresholds"
// @Success 200 {object} dto.TemperatureThresholdsResult
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/vehicles/{id}/temperature-thresholds [patch]
func (h *VehicleHandler) SetTemperatureThresholds(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.ErrorResponse(c, 400, "Invalid ID")
	}
	var req dto.TemperatureThresholdsRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, 400, "Invalid request body")
	}
	result, err := h.Service.SetTemperatureThresholds(utils.RequestContext(c), id, req)
	if err != nil {
		return utils.HandleDatabaseError(c, err)
	}
	return utils.SuccessResponse(c, 200, result)
}
