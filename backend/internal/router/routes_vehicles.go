package router

import (
	app_handlers "fratelli-feccia/internal/handlers"
	"fratelli-feccia/pkg/middleware"
	"fratelli-feccia/pkg/utils"

	"github.com/gofiber/fiber/v2"
)

// registerVehicleRoutes mirrors backend/routers/vehicles.py + temperature.py's
// vehicle-scoped endpoints. CRUD is admin+operatore, GPS/temperature-threshold
// writes are admin+planner, reads are open to any authenticated role.
func registerVehicleRoutes(authAll fiber.Router, handlers *app_handlers.Handler) {
	requireWrite := middleware.RequireRole(utils.RoleAdmin, utils.RoleOperatore)
	requirePlanner := middleware.RequireRole(utils.RoleAdmin, utils.RolePlanner)

	authAll.Get("/vehicles", handlers.Vehicles.ListVehicles)
	authAll.Get("/vehicles/gps-live", handlers.Vehicles.GetAllGPSLive)
	authAll.Get("/vehicles/:id/gps-history", handlers.Vehicles.GetVehicleGPSHistory)
	authAll.Get("/vehicles/:id/temperature", handlers.Vehicles.GetVehicleTemperature)

	authAll.Post("/vehicles", requireWrite, handlers.Vehicles.CreateVehicle)
	authAll.Put("/vehicles/:id", requireWrite, handlers.Vehicles.UpdateVehicle)
	authAll.Delete("/vehicles/:id", requireWrite, handlers.Vehicles.DeleteVehicle)

	authAll.Post("/vehicles/gps-position-by-plate/:targa", requirePlanner, handlers.Vehicles.UpdateVehicleGPSByPlate)
	authAll.Post("/vehicles/:id/gps-position", requirePlanner, handlers.Vehicles.UpdateVehicleGPS)
	authAll.Patch("/vehicles/:id/temperature-thresholds", requirePlanner, handlers.Vehicles.SetTemperatureThresholds)
}

// registerWebhookRoutes mirrors backend/routers/gps_webhook.py and
// temperature.py: public endpoints (no JWT), gated by a static token
// (see pkg/middleware.RequireWebhookToken).
func registerWebhookRoutes(api fiber.Router, handlers *app_handlers.Handler) {
	api.Post("/webhooks/gps/:vendor", middleware.RequireWebhookToken("GPS_WEBHOOK_TOKEN"), handlers.Vehicles.GPSWebhook)
	api.Post("/webhooks/temperature/:vendor", middleware.RequireWebhookToken("TEMP_WEBHOOK_TOKEN"), handlers.Vehicles.TemperatureWebhook)
}
