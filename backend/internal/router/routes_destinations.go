package router

import (
	app_handlers "fratelli-feccia/internal/handlers"
	"fratelli-feccia/pkg/middleware"
	"fratelli-feccia/pkg/utils"

	"github.com/gofiber/fiber/v2"
)

// registerDestinationRoutes mirrors backend/routers/destinations.py.
func registerDestinationRoutes(authAll fiber.Router, handlers *app_handlers.Handler) {
	authAll.Get("/destinations", handlers.Destinations.ListDestinations)
	authAll.Get("/destinations/:id", handlers.Destinations.GetDestinationByID)

	writeGroup := authAll.Group("/destinations", middleware.RequireRole(utils.RoleAdmin, utils.RoleOperatore))
	writeGroup.Post("", handlers.Destinations.CreateDestination)
	writeGroup.Put("/:id", handlers.Destinations.UpdateDestination)
	writeGroup.Delete("/:id", handlers.Destinations.DeleteDestination)
}
