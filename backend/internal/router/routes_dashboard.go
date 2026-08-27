package router

import (
	app_handlers "fratelli-feccia/internal/handlers"

	"github.com/gofiber/fiber/v2"
)

// registerDashboardRoutes mirrors backend/routers/dashboard.py: all
// read-only, open to any authenticated role.
func registerDashboardRoutes(authAll fiber.Router, handlers *app_handlers.Handler) {
	authAll.Get("/dashboard/stats", handlers.Dashboard.Stats)
	authAll.Get("/dashboard/customer/:customer_id", handlers.Dashboard.CustomerDashboard)
	authAll.Get("/dashboard/recent-orders", handlers.Dashboard.RecentOrders)
	authAll.Get("/dashboard/nav-counts", handlers.Dashboard.NavCounts)
}
