package router

import (
	app_handlers "fratelli-feccia/internal/handlers"
	"fratelli-feccia/pkg/middleware"
	"fratelli-feccia/pkg/utils"

	"github.com/gofiber/fiber/v2"
)

// registerOrderRoutes mirrors backend/routers/orders.py.
func registerOrderRoutes(authAll fiber.Router, handlers *app_handlers.Handler) {
	requireWrite := middleware.RequireRole(utils.RoleAdmin, utils.RolePlanner, utils.RoleOperatore)
	requirePlanner := middleware.RequireRole(utils.RoleAdmin, utils.RolePlanner)

	authAll.Get("/orders", handlers.Orders.ListOrders)
	authAll.Get("/orders/:id", handlers.Orders.GetOrderByID)
	authAll.Get("/orders/:id/return-suggestions", handlers.Orders.ReturnSuggestions)
	authAll.Get("/orders/:id/cmr/pdf", handlers.Orders.GetCMRPDF)

	authAll.Post("/orders", requireWrite, handlers.Orders.CreateOrder)
	authAll.Put("/orders/:id", requireWrite, handlers.Orders.UpdateOrder)

	authAll.Patch("/orders/:id/assign", requirePlanner, handlers.Orders.AssignOrder)
	authAll.Patch("/orders/:id/start", requirePlanner, handlers.Orders.StartOrder)
	authAll.Patch("/orders/:id/close", requirePlanner, handlers.Orders.CloseOrder)
	authAll.Patch("/orders/:id/discard", requirePlanner, handlers.Orders.DiscardOrder)
	authAll.Delete("/orders/:id", requirePlanner, handlers.Orders.DeleteOrder)
}
