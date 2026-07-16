package router

import (
	app_handlers "fratelli-feccia/internal/handlers"
	"fratelli-feccia/pkg/middleware"
	"fratelli-feccia/pkg/utils"

	"github.com/gofiber/fiber/v2"
)

// registerProductRoutes mirrors backend/routers/products.py.
func registerProductRoutes(authAll fiber.Router, handlers *app_handlers.Handler) {
	authAll.Get("/products", handlers.Products.ListProducts)

	writeGroup := authAll.Group("/products", middleware.RequireRole(utils.RoleAdmin, utils.RoleOperatore))
	writeGroup.Post("", handlers.Products.CreateProduct)
	writeGroup.Put("/:id", handlers.Products.UpdateProduct)
	writeGroup.Delete("/:id", handlers.Products.DeleteProduct)
}
