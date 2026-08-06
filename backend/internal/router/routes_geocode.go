package router

import (
	app_handlers "fratelli-feccia/internal/handlers"
	"fratelli-feccia/pkg/middleware"
	"fratelli-feccia/pkg/utils"

	"github.com/gofiber/fiber/v2"
)

// registerGeocodeRoutes: forward-geocoding used by anagrafica forms
// (Destination/Garage/WashStation, and the "cliente" portal's own
// Nuova-Destinazione form) to search an address on the map picker. Open to
// any authenticated role — read-only, no anagrafica-specific ACL needed.
//
// Registered on the ungrouped `api` router with JWTAuthMiddleware inline,
// NOT via authAll: same reasoning as registerDestinationReadRoutes — this
// must run BEFORE authAll's Group("", ...) is created (see routes.go),
// otherwise it would silently re-inherit PermitAllRoles(), which excludes
// "cliente".
func registerGeocodeRoutes(api fiber.Router, jwtCfg utils.JWTConfig, handlers *app_handlers.Handler) {
	api.Get("/geocode/search", middleware.JWTAuthMiddleware(jwtCfg), handlers.Geocode.Search)
}
