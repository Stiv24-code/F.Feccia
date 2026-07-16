package router

import (
	app_handlers "fratelli-feccia/internal/handlers"
	"fratelli-feccia/pkg/middleware"
	"fratelli-feccia/pkg/utils"

	"github.com/gofiber/fiber/v2"
)

// registerDriverRoutes mirrors backend/routers/drivers.py.
func registerDriverRoutes(authAll fiber.Router, handlers *app_handlers.Handler) {
	authAll.Get("/drivers", handlers.Drivers.ListDrivers)

	writeGroup := authAll.Group("/drivers", middleware.RequireRole(utils.RoleAdmin, utils.RoleOperatore))
	writeGroup.Post("", handlers.Drivers.CreateDriver)
	writeGroup.Put("/:id", handlers.Drivers.UpdateDriver)
	writeGroup.Delete("/:id", handlers.Drivers.DeleteDriver)
}
