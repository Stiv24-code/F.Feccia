package app

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"fratelli-feccia/pkg/middleware"
	"fratelli-feccia/pkg/telemetry"
	"fratelli-feccia/pkg/utils"

	"go.opentelemetry.io/otel/trace"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func (a *App) setupObservabilityMiddleware() {
	if a.Telemetry.TraceProvider != nil {
		a.Router.Use(telemetry.NewFiberMiddleware(a.Telemetry))
	}
	if a.Telemetry.MeterProvider != nil {
		a.Router.Use(telemetry.NewFiberMetricsMiddleware())
	}
	if a.Telemetry.TraceProvider != nil {
		a.Router.Use(func(c *fiber.Ctx) error {
			sc := trace.SpanContextFromContext(c.UserContext())
			if sc.IsValid() {
				c.Locals("trace_id", sc.TraceID().String())
			}
			return c.Next()
		})
	}
}

func (a *App) setupRateLimitingMiddleware() {
	rateLimitMax := a.Config.Server.RateLimitMax
	rateLimitWindow := time.Duration(a.Config.Server.RateLimitWindow) * time.Second
	if strings.EqualFold(os.Getenv("IS_LOCAL"), "true") {
		slog.Info("Rate limiting disabled (local mode)")
		return
	}

	a.Router.Use(limiter.New(limiter.Config{
		Max:        rateLimitMax,
		Expiration: rateLimitWindow,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return utils.ErrorResponse(c, 429, fmt.Sprintf("Too many requests (limit: %d per %v), please try again later", rateLimitMax, rateLimitWindow))
		},
	}))
	slog.Info("Rate limiting configured", "max", rateLimitMax, "window", rateLimitWindow)
}

func (a *App) setupRequestLoggingMiddleware() {
	a.Router.Use(logger.New(logger.Config{
		Format: `{"time":"${time}","level":"info","msg":"http_request","request_id":"${locals:requestid}","trace_id":"${locals:trace_id}","status":${status},"latency":"${latency}","method":"${method}","path":"${path}","ip":"${ip}","err":"${error}","app_error":"${locals:app_error}"}` + "\n",
	}))
}

func (a *App) setupContextMiddleware() {
	a.Router.Use(middleware.ContextMiddleware(30 * time.Second))
}

func (a *App) setupCORSMiddleware() {
	corsOrigins := os.Getenv("CORS_ORIGINS")
	if corsOrigins == "" {
		corsOrigins = "http://localhost:8090,http://localhost:8080,http://127.0.0.1:8090,http://127.0.0.1:8080"
	}

	a.Router.Use(cors.New(cors.Config{
		AllowOrigins:     corsOrigins,
		AllowMethods:     "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization, X-CSRF-Token",
		AllowCredentials: true,
		MaxAge:           600,
	}))
	slog.Info("CORS configured", "origins", corsOrigins)
}
