package app

import (
	"context"
	"log/slog"
	"os"
	"strings"
	"time"

	"fratelli-feccia/config"
	"fratelli-feccia/pkg/database"
	"fratelli-feccia/pkg/telemetry"
	"fratelli-feccia/pkg/utils"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func initLogger() {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	slog.SetDefault(slog.New(handler))
}

func newFiberApp() *fiber.App {
	return fiber.New(fiber.Config{
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
		BodyLimit:    10 * 1024 * 1024,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			slog.Error("request failed",
				"method", c.Method(),
				"path", c.Path(),
				"status", code,
				"error", err.Error(),
				"request_id", c.Locals("requestid"),
			)
			return c.Status(code).JSON(fiber.Map{"error": err.Error()})
		},
	})
}

func initJWTConfig(cfg *config.Config) utils.JWTConfig {
	return utils.NewJWTConfig(
		cfg.Security.JWTAccessSecret,
		cfg.Security.JWTRefreshSecret,
		cfg.Security.JWTAccessTTL,
		cfg.Security.JWTRefreshTTL,
	)
}

func initTelemetryIfEnabled() (telemetry.TelemetryProviders, telemetry.Shutdown) {
	var providers telemetry.TelemetryProviders
	var telemetryShutdown telemetry.Shutdown

	telemetryCtx, cancelTelemetry := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelTelemetry()

	if strings.EqualFold(os.Getenv("TELEMETRY_ENABLED"), "true") {
		if p, shutdown, err := telemetry.Init(telemetryCtx); err != nil {
			slog.Warn("OpenTelemetry initialization failed, continuing without telemetry", "error", err)
		} else {
			providers = p
			telemetryShutdown = shutdown
		}
	} else {
		slog.Info("OpenTelemetry disabled (TELEMETRY_ENABLED!=true)")
	}

	return providers, telemetryShutdown
}

func initDatabase(cfg *config.Config) *gorm.DB {
	db, err := database.Connect(cfg)
	if err != nil {
		slog.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	return db
}
