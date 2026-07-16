package router

import (
	app_handlers "fratelli-feccia/internal/handlers"
	"fratelli-feccia/pkg/middleware"
	"fratelli-feccia/pkg/utils"

	"github.com/gofiber/fiber/v2"
)

// registerPriceListRoutes mirrors backend/routers/pricelists.py: write
// (create/update/delete pricelist + item CRUD) is admin+amministrazione,
// reads (list/get/lookup-tariff) are open to any authenticated role.
func registerPriceListRoutes(authAll fiber.Router, handlers *app_handlers.Handler) {
	requireAmministrazione := middleware.RequireRole(utils.RoleAdmin, utils.RoleAmministrazione)

	authAll.Get("/pricelists", handlers.PriceLists.ListPriceLists)
	authAll.Get("/pricelists/lookup-tariff", handlers.PriceLists.LookupTariff)
	authAll.Get("/pricelists/:id", handlers.PriceLists.GetPriceListByID)

	authAll.Post("/pricelists", requireAmministrazione, handlers.PriceLists.CreatePriceList)
	authAll.Put("/pricelists/:id", requireAmministrazione, handlers.PriceLists.UpdatePriceList)
	authAll.Delete("/pricelists/:id", requireAmministrazione, handlers.PriceLists.DeletePriceList)

	authAll.Post("/pricelists/:id/items", requireAmministrazione, handlers.PriceLists.AddItem)
	authAll.Put("/pricelists/:id/items/:item_id", requireAmministrazione, handlers.PriceLists.UpdateItem)
	authAll.Delete("/pricelists/:id/items/:item_id", requireAmministrazione, handlers.PriceLists.DeleteItem)
}
