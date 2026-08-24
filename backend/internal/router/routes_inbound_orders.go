package router

import (
	app_handlers "fratelli-feccia/internal/handlers"
	"fratelli-feccia/pkg/middleware"
	"fratelli-feccia/pkg/utils"

	"github.com/gofiber/fiber/v2"
)

// registerInboundOrderRoutes exposes OrderMesh's acceptance dashboard: the
// inbound-order list plus the confirm/accept/modify/reset actions. Writes
// use the same role set as the TMS orders (admin/planner/operatore) — in
// OrderMesh these lived unauthenticated at /api/orders.
func registerInboundOrderRoutes(authAll fiber.Router, handlers *app_handlers.Handler) {
	requireWrite := middleware.RequireRole(utils.RoleAdmin, utils.RolePlanner, utils.RoleOperatore)

	authAll.Get("/inbound-orders", handlers.InboundOrders.ListInboundOrders)
	authAll.Get("/inbound-config", handlers.InboundOrders.GetInboundConfig)

	authAll.Post("/inbound-orders", requireWrite, handlers.InboundOrders.CreateInboundOrder)
	authAll.Post("/inbound-orders/scrape", requireWrite, handlers.InboundOrders.ScrapeInboundOrders)
	authAll.Post("/inbound-orders/:id/accept", requireWrite, handlers.InboundOrders.AcceptInboundOrder)
	// Conversione draft -> Order: stesso ruolo di scrittura di POST /orders,
	// che e' esattamente il privilegio che serve — chi puo' convertire una
	// richiesta potrebbe comunque creare l'ordine equivalente a mano.
	authAll.Post("/inbound-orders/:id/convert", requireWrite, handlers.InboundOrders.ConvertInboundOrder)
	authAll.Post("/inbound-orders/:id/modify", requireWrite, handlers.InboundOrders.ModifyInboundOrder)
	authAll.Post("/inbound-orders/:id/reset", requireWrite, handlers.InboundOrders.ResetInboundOrder)
}
