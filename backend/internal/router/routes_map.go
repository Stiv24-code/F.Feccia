package router

import (
	app_handlers "fratelli-feccia/internal/handlers"

	"github.com/gofiber/fiber/v2"
)

// registerMapRoutes mirrors backend/routers/map.py: read-only, open to any
// authenticated role.
func registerMapRoutes(authAll fiber.Router, handlers *app_handlers.Handler) {
	authAll.Get("/map/trips", handlers.Map.Trips)
}
