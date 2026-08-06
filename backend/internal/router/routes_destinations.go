package router

import (
	app_handlers "fratelli-feccia/internal/handlers"
	"fratelli-feccia/pkg/middleware"
	"fratelli-feccia/pkg/utils"

	"github.com/gofiber/fiber/v2"
)

// registerDestinationReadRoutes mirrors the read side of
// backend/routers/destinations.py. Registered on the ungrouped `api` router,
// BEFORE authAll's Group("", ...) is created (see routes.go) — with
// JWTAuthMiddleware applied inline, no role restriction. Destinations are
// shared, non-sensitive master data (pickup/delivery addresses), and a
// "cliente" account needs to read them to pick carico/scarico when creating
// its own order (see ClientOrdersPage.tsx).
//
// Ordering matters here, not just the inline-vs-Group choice: Fiber's
// Use(prefix, mw) leaks forward onto every route sharing that prefix
// registered afterwards on the same underlying router, regardless of which
// wrapper variable (api/authAll/...) is used to register it (see
// routes_client_portal.go's comment for the full explanation — this exact
// mistake once made every role get 403'd). Registering these reads after
// authAll exists would silently re-inherit PermitAllRoles(), which excludes
// "cliente" — defeating the point.
func registerDestinationReadRoutes(api fiber.Router, jwtCfg utils.JWTConfig, handlers *app_handlers.Handler) {
	api.Get("/destinations", middleware.JWTAuthMiddleware(jwtCfg), handlers.Destinations.ListDestinations)
	api.Get("/destinations/:id", middleware.JWTAuthMiddleware(jwtCfg), handlers.Destinations.GetDestinationByID)
}

// registerDestinationWriteRoutes mirrors the write side — staff-only
// (admin/operatore), registered normally under authAll like every other
// write endpoint.
func registerDestinationWriteRoutes(authAll fiber.Router, handlers *app_handlers.Handler) {
	writeGroup := authAll.Group("/destinations", middleware.RequireRole(utils.RoleAdmin, utils.RoleOperatore))
	writeGroup.Post("", handlers.Destinations.CreateDestination)
	writeGroup.Put("/:id", handlers.Destinations.UpdateDestination)
	writeGroup.Delete("/:id", handlers.Destinations.DeleteDestination)
}
