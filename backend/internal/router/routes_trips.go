package router

import (
	app_handlers "fratelli-feccia/internal/handlers"
	"fratelli-feccia/pkg/middleware"
	"fratelli-feccia/pkg/utils"

	"github.com/gofiber/fiber/v2"
)

// registerTripRoutes mirrors backend/routers/trips.py: write (create,
// recompute-segments, complete, add-order) is admin+planner, reads are open
// to any authenticated role.
func registerTripRoutes(authAll fiber.Router, handlers *app_handlers.Handler) {
	requirePlanner := middleware.RequireRole(utils.RoleAdmin, utils.RolePlanner)

	authAll.Get("/trips", handlers.Trips.ListTrips)
	authAll.Get("/trips/:id", handlers.Trips.GetTripByID)
	authAll.Get("/trips/:id/instructions/pdf", handlers.Trips.GetInstructionsPDF)

	authAll.Post("/trips", requirePlanner, handlers.Trips.CreateTrip)
	authAll.Post("/trips/:id/recompute-segments", requirePlanner, handlers.Trips.RecomputeSegments)
	authAll.Patch("/trips/:id/complete", requirePlanner, handlers.Trips.CompleteTrip)
	authAll.Patch("/trips/:id/add-order", requirePlanner, handlers.Trips.AddOrderToTrip)
}
