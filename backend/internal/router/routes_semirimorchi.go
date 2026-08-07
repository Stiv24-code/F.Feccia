package router

import (
	app_handlers "fratelli-feccia/internal/handlers"
	"fratelli-feccia/pkg/middleware"
	"fratelli-feccia/pkg/utils"

	"github.com/gofiber/fiber/v2"
)

// registerSemirimorchioRoutes mirrors registerMotriceRoutes for the trailer
// half of the former single Vehicle table.
func registerSemirimorchioRoutes(authAll fiber.Router, handlers *app_handlers.Handler) {
	requireWrite := middleware.RequireRole(utils.RoleAdmin, utils.RoleOperatore)

	authAll.Get("/semirimorchi", handlers.Semirimorchi.ListSemirimorchi)
	authAll.Post("/semirimorchi", requireWrite, handlers.Semirimorchi.CreateSemirimorchio)
	authAll.Put("/semirimorchi/:id", requireWrite, handlers.Semirimorchi.UpdateSemirimorchio)
	authAll.Delete("/semirimorchi/:id", requireWrite, handlers.Semirimorchi.DeleteSemirimorchio)
}
