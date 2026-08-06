package router

import (
	app_handlers "fratelli-feccia/internal/handlers"
	"fratelli-feccia/pkg/middleware"
	"fratelli-feccia/pkg/utils"

	"github.com/gofiber/fiber/v2"
)

// registerClientPortalRoutes mounts the "cliente" self-service portal's own
// scoped routes on the ungrouped `api` router, JWT + RequireRole(cliente)
// applied inline per-route — deliberately NOT via a Group("", ...): that's
// the Fiber footgun documented across this package (see
// registerAuthMeRoute/routes_masterdata.go) where Use(prefix, mw) leaks onto
// every route sharing that prefix, not just the ones chained off the
// returned group. With prefix "" that means EVERY route in the app, which
// would stack RequireRole(cliente) on top of authAll's PermitAllRoles() for
// every staff request too — nobody satisfies both, 403 across the board
// (this exact regression happened once already, see git history).
//
// Every handler here reads utils.RequestCustomerID(c) from the JWT claim
// instead of trusting any id in the path/body — a client can only ever
// read/write its own anagrafica and orders.
func registerClientPortalRoutes(api fiber.Router, handlers *app_handlers.Handler, jwtCfg utils.JWTConfig) {
	jwtAuth := middleware.JWTAuthMiddleware(jwtCfg)
	requireCliente := middleware.RequireRole(utils.RoleCliente)

	api.Get("/me/anagrafica", jwtAuth, requireCliente, handlers.Customers.GetMyAnagrafica)
	api.Put("/me/anagrafica", jwtAuth, requireCliente, handlers.Customers.UpdateMyAnagrafica)

	api.Get("/me/orders", jwtAuth, requireCliente, handlers.Orders.ListMyOrders)
	api.Get("/me/orders/:id", jwtAuth, requireCliente, handlers.Orders.GetMyOrderByID)
	api.Post("/me/orders", jwtAuth, requireCliente, handlers.Orders.CreateMyOrder)
	api.Delete("/me/orders/:id", jwtAuth, requireCliente, handlers.Orders.DeleteMyOrder)

	// Shared pool with staff (registerDestinationReadRoutes) — a client can
	// add a new pickup/delivery address to pick from, but not edit/delete
	// existing ones (those stay staff-only under authAll).
	api.Post("/me/destinations", jwtAuth, requireCliente, handlers.Destinations.CreateMyDestination)
}
