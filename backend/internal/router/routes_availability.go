package router

import (
	app_handlers "fratelli-feccia/internal/handlers"

	"github.com/gofiber/fiber/v2"
)

// registerAvailabilityRoutes mirrors backend/routers/availability.py:
// read-only, open to any authenticated role.
func registerAvailabilityRoutes(authAll fiber.Router, handlers *app_handlers.Handler) {
	authAll.Get("/availability/motrici", handlers.Availability.Motrici)
	authAll.Get("/availability/semirimorchi", handlers.Availability.Semirimorchi)
	authAll.Get("/availability/drivers", handlers.Availability.Drivers)
}
