package router

import (
	app_handlers "fratelli-feccia/internal/handlers"
	"fratelli-feccia/pkg/middleware"
	"fratelli-feccia/pkg/utils"

	"github.com/gofiber/fiber/v2"
)

// registerCarrierRoutes mirrors backend/routers/carriers.py.
func registerCarrierRoutes(authAll fiber.Router, handlers *app_handlers.Handler) {
	authAll.Get("/carriers", handlers.Carriers.ListCarriers)

	writeGroup := authAll.Group("/carriers", middleware.RequireRole(utils.RoleAdmin, utils.RoleOperatore))
	writeGroup.Post("", handlers.Carriers.CreateCarrier)
	writeGroup.Put("/:id", handlers.Carriers.UpdateCarrier)
	writeGroup.Delete("/:id", handlers.Carriers.DeleteCarrier)
}
