package app

import (
	"fratelli-feccia/internal/router"
	"fratelli-feccia/pkg/swagger"

	"github.com/gofiber/fiber/v2"
)

func (a *App) registerRoutes() {
	swagger.SetupSwagger(a.Router, a.Config)

	a.Router.Get("/api/v1/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok", "service": a.Name})
	})

	// Mirrors backend/routers/meta.py's root endpoint.
	a.Router.Get("/api/v1/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "LoginBusiness TMS API", "version": "1.0.0"})
	})

	router.SetupRoutes(a.Router, a.DB, a.jwtCfg, a.Config.S3)
}
