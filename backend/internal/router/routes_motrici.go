package router

import (
	app_handlers "fratelli-feccia/internal/handlers"
	"fratelli-feccia/pkg/middleware"
	"fratelli-feccia/pkg/utils"

	"github.com/gofiber/fiber/v2"
)

// registerMotriceRoutes mirrors registerVehicleRoutes' CRUD half — reads
// open to any authenticated role, writes admin+operatore.
func registerMotriceRoutes(authAll fiber.Router, handlers *app_handlers.Handler) {
	requireWrite := middleware.RequireRole(utils.RoleAdmin, utils.RoleOperatore)

	authAll.Get("/motrici", handlers.Motrici.ListMotrici)
	authAll.Post("/motrici", requireWrite, handlers.Motrici.CreateMotrice)
	authAll.Put("/motrici/:id", requireWrite, handlers.Motrici.UpdateMotrice)
	authAll.Delete("/motrici/:id", requireWrite, handlers.Motrici.DeleteMotrice)
}
