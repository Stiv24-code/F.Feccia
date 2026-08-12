package router

import (
	"time"

	app_handlers "fratelli-feccia/internal/handlers"
	"fratelli-feccia/pkg/middleware"
	"fratelli-feccia/pkg/utils"

	"github.com/gofiber/fiber/v2"
)

// registerGeocodeRoutes: forward-geocoding used by anagrafica forms
// (Destination/Garage/WashStation, Customer, and the "cliente" portal's own
// Nuova-Destinazione form) to search an address on the map picker. No auth
// required — the public /registrati self-signup form (ClientRegisterPage)
// calls this before any account/token exists, so gating it behind
// JWTAuthMiddleware made address search 401 there. Read-only and no
// anagrafica-specific ACL needed either way; the real concern with opening it
// up is quota abuse of the paid ORS API behind it, guarded by the per-IP
// limiter below (tighter than the app-wide default since this is now
// reachable by anyone, not just logged-in staff/clients).
//
// Registered on the ungrouped `api` router, NOT via authAll: same reasoning
// as registerDestinationReadRoutes — this must run BEFORE authAll's
// Group("", ...) is created (see routes.go), otherwise it would silently
// re-inherit PermitAllRoles(), which excludes "cliente".
func registerGeocodeRoutes(api fiber.Router, jwtCfg utils.JWTConfig, handlers *app_handlers.Handler) {
	api.Get("/geocode/search", middleware.NewLimiterIPPath(30, time.Minute), handlers.Geocode.Search)
}
