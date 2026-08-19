package app

import (
	"fratelli-feccia/internal/router"
	"fratelli-feccia/pkg/swagger"
	"fratelli-feccia/pkg/telemetry"

	"github.com/gofiber/fiber/v2"
)

func (a *App) registerRoutes() {
	swagger.SetupSwagger(a.Router, a.Config)

	a.Router.Get("/api/v1/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok", "service": a.Name})
	})

	// Prometheus scrape endpoint — only meaningful (and only mounted) when
	// telemetry actually initialized (TELEMETRY_ENABLED=true), same gate as
	// the tracing/metrics middleware in setupObservabilityMiddleware.
	if a.Telemetry.MeterProvider != nil {
		a.Router.Get("/metrics", telemetry.NewPrometheusHandler())
	}

	// Mirrors backend/routers/meta.py's root endpoint.
	a.Router.Get("/api/v1/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "LoginBusiness TMS API", "version": "1.0.0"})
	})

	a.Services = router.SetupRoutes(a.Router, a.DB, a.jwtCfg, a.Config.S3, a.Config.Routing, a.Config.Inbound, a.Config.Server.AppBaseURL)
}
