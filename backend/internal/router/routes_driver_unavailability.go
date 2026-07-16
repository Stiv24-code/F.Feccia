package router

import (
	app_handlers "fratelli-feccia/internal/handlers"
	"fratelli-feccia/pkg/middleware"
	"fratelli-feccia/pkg/utils"

	"github.com/gofiber/fiber/v2"
)

// registerDriverUnavailabilityRoutes mirrors backend/routers/driver_unavailability.py:
// read is open to any authenticated role, write (create/delete) is admin+planner.
func registerDriverUnavailabilityRoutes(authAll fiber.Router, handlers *app_handlers.Handler) {
	authAll.Get("/driver-unavailability", handlers.DriverUnavailability.List)

	writeGroup := authAll.Group("/driver-unavailability", middleware.RequireRole(utils.RoleAdmin, utils.RolePlanner))
	writeGroup.Post("", handlers.DriverUnavailability.Create)
	writeGroup.Delete("/:id", handlers.DriverUnavailability.Delete)
}
