package router

import (
	app_handlers "fratelli-feccia/internal/handlers"
	"fratelli-feccia/pkg/middleware"
	"fratelli-feccia/pkg/utils"

	"github.com/gofiber/fiber/v2"
)

// registerGarageRoutes mirrors backend/routers/garages.py.
func registerGarageRoutes(authAll fiber.Router, handlers *app_handlers.Handler) {
	authAll.Get("/garages", handlers.Garages.ListGarages)

	writeGroup := authAll.Group("/garages", middleware.RequireRole(utils.RoleAdmin, utils.RoleOperatore))
	writeGroup.Post("", handlers.Garages.CreateGarage)
	writeGroup.Put("/:id", handlers.Garages.UpdateGarage)
	writeGroup.Delete("/:id", handlers.Garages.DeleteGarage)
}
