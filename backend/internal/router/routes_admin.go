package router

import (
	app_handlers "fratelli-feccia/internal/handlers"
	"fratelli-feccia/pkg/middleware"

	"github.com/gofiber/fiber/v2"
)

func registerAdminRoutes(authAll fiber.Router, handlers *app_handlers.Handler) {
	authAll.Get("/admin/users/:id", handlers.Admin.GetUserByID)

	adminGroup := authAll.Group("/admin", middleware.RequireAdmin())
	adminGroup.Get("/users-list", handlers.Admin.ListUsers)
	adminGroup.Get("/users", handlers.Admin.ListAllUsers)
	adminGroup.Patch("/users/:id", handlers.Admin.PatchUser)
	adminGroup.Post("/users", handlers.Admin.CreateUser)
	adminGroup.Put("/users/:id", handlers.Admin.UpdateUser)
	adminGroup.Delete("/users/:id", handlers.Admin.DeleteUser)
}
