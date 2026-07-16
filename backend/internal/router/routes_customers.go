package router

import (
	app_handlers "fratelli-feccia/internal/handlers"
	"fratelli-feccia/pkg/middleware"
	"fratelli-feccia/pkg/utils"

	"github.com/gofiber/fiber/v2"
)

// registerCustomerRoutes mirrors backend/routers/customers.py: read is open to
// any authenticated role (authAll already enforces that), write is restricted
// to admin/operatore (require_roles("admin", "operatore") in Python).
func registerCustomerRoutes(authAll fiber.Router, handlers *app_handlers.Handler) {
	authAll.Get("/customers", handlers.Customers.ListCustomers)
	authAll.Get("/customers/:id", handlers.Customers.GetCustomerByID)

	writeGroup := authAll.Group("/customers", middleware.RequireRole(utils.RoleAdmin, utils.RoleOperatore))
	writeGroup.Post("", handlers.Customers.CreateCustomer)
	writeGroup.Put("/:id", handlers.Customers.UpdateCustomer)
	writeGroup.Delete("/:id", handlers.Customers.DeleteCustomer)
}
