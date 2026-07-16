package router

import (
	app_handlers "fratelli-feccia/internal/handlers"
	"fratelli-feccia/pkg/middleware"
	"fratelli-feccia/pkg/utils"

	"github.com/gofiber/fiber/v2"
)

// registerAnagraficheRoutes mirrors backend/routers/anagrafiche_extra.py —
// write is admin+amministrazione (not operatore, unlike most master data).
//
// The write middleware is passed inline per-route (not via an empty-prefix
// authAll.Group("", mw)) because Fiber v2's Group(prefix, handlers...) with
// a non-empty handlers list registers those handlers as Use(prefix, ...) —
// which matches every route sharing that prefix registered afterwards, not
// just the routes chained off the returned group. With prefix "" that means
// the middleware would leak onto every route registered later in routes.go.
func registerAnagraficheRoutes(authAll fiber.Router, handlers *app_handlers.Handler) {
	requireAmministrazione := middleware.RequireRole(utils.RoleAdmin, utils.RoleAmministrazione)

	authAll.Get("/countries", handlers.Anagrafiche.ListCountries)
	authAll.Post("/countries", requireAmministrazione, handlers.Anagrafiche.CreateCountry)
	authAll.Put("/countries/:id", requireAmministrazione, handlers.Anagrafiche.UpdateCountry)
	authAll.Delete("/countries/:id", requireAmministrazione, handlers.Anagrafiche.DeleteCountry)

	authAll.Get("/banks", handlers.Anagrafiche.ListBanks)
	authAll.Post("/banks", requireAmministrazione, handlers.Anagrafiche.CreateBank)
	authAll.Put("/banks/:id", requireAmministrazione, handlers.Anagrafiche.UpdateBank)
	authAll.Delete("/banks/:id", requireAmministrazione, handlers.Anagrafiche.DeleteBank)

	authAll.Get("/accounting-entries", handlers.Anagrafiche.ListAccountingEntries)
	authAll.Post("/accounting-entries", requireAmministrazione, handlers.Anagrafiche.CreateAccountingEntry)
	authAll.Put("/accounting-entries/:id", requireAmministrazione, handlers.Anagrafiche.UpdateAccountingEntry)
	authAll.Delete("/accounting-entries/:id", requireAmministrazione, handlers.Anagrafiche.DeleteAccountingEntry)
}
