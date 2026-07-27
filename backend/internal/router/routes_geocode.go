package router

import (
	app_handlers "fratelli-feccia/internal/handlers"

	"github.com/gofiber/fiber/v2"
)

// registerGeocodeRoutes: forward-geocoding used by anagrafica forms
// (Destination/Garage/WashStation) to search an address on the map picker.
// Open to any authenticated role — read-only, no anagrafica-specific ACL needed.
func registerGeocodeRoutes(authAll fiber.Router, handlers *app_handlers.Handler) {
	authAll.Get("/geocode/search", handlers.Geocode.Search)
}
