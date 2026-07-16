package router

import (
	app_handlers "fratelli-feccia/internal/handlers"

	"github.com/gofiber/fiber/v2"
)

// registerExportRoutes mirrors backend/routers/export.py: open to any
// authenticated role (rate-limited in Python via slowapi; local-mode rate
// limiting here follows the same pattern already used by the other
// endpoints, see pkg/middleware — no per-route change needed).
func registerExportRoutes(authAll fiber.Router, handlers *app_handlers.Handler) {
	authAll.Get("/export/orders", handlers.Export.Orders)
}
