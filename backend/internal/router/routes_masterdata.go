package router

import (
	app_handlers "fratelli-feccia/internal/handlers"
	"fratelli-feccia/pkg/middleware"
	"fratelli-feccia/pkg/utils"

	"github.com/gofiber/fiber/v2"
)

// registerMasterdataRoutes mirrors backend/routers/masterdata.py — write is
// admin-only (require_roles("admin")), unlike the other master-data routers
// which also allow "operatore".
//
// The write middleware is passed inline per-route (not via an empty-prefix
// authAll.Group("", mw)) because Fiber v2's Group(prefix, handlers...) with
// a non-empty handlers list registers those handlers as Use(prefix, ...) —
// which matches every route sharing that prefix registered afterwards, not
// just the routes chained off the returned group. With prefix "" that means
// the middleware would leak onto every route registered later in routes.go.
func registerMasterdataRoutes(authAll fiber.Router, handlers *app_handlers.Handler) {
	requireAdmin := middleware.RequireRole(utils.RoleAdmin)

	authAll.Get("/vehicle-types", handlers.Masterdata.ListVehicleTypes)
	authAll.Post("/vehicle-types", requireAdmin, handlers.Masterdata.CreateVehicleType)

	authAll.Get("/accessory-costs", handlers.Masterdata.ListAccessoryCosts)
	authAll.Post("/accessory-costs", requireAdmin, handlers.Masterdata.CreateAccessoryCost)

	authAll.Get("/transport-categories", handlers.Masterdata.ListTransportCategories)
	authAll.Post("/transport-categories", requireAdmin, handlers.Masterdata.CreateTransportCategory)
}
