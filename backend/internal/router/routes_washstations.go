package router

import (
	app_handlers "fratelli-feccia/internal/handlers"
	"fratelli-feccia/pkg/middleware"
	"fratelli-feccia/pkg/utils"

	"github.com/gofiber/fiber/v2"
)

// registerWashStationRoutes wires up the punti-di-lavaggio CRUD.
func registerWashStationRoutes(authAll fiber.Router, handlers *app_handlers.Handler) {
	authAll.Get("/wash-stations", handlers.WashStations.ListWashStations)

	writeGroup := authAll.Group("/wash-stations", middleware.RequireRole(utils.RoleAdmin, utils.RoleOperatore))
	writeGroup.Post("", handlers.WashStations.CreateWashStation)
	writeGroup.Put("/:id", handlers.WashStations.UpdateWashStation)
	writeGroup.Delete("/:id", handlers.WashStations.DeleteWashStation)
}
