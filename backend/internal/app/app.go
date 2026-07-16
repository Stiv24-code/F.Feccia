package app

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"fratelli-feccia/config"
	"fratelli-feccia/pkg/audit"
	"fratelli-feccia/pkg/database"
	"fratelli-feccia/pkg/jobs"
	"fratelli-feccia/pkg/middleware"
	"fratelli-feccia/pkg/telemetry"
	"fratelli-feccia/pkg/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"gorm.io/gorm"
)

type App struct {
	Name              string
	DB                *gorm.DB
	Router            *fiber.App
	Config            *config.Config
	jwtCfg            utils.JWTConfig
	Telemetry         telemetry.TelemetryProviders
	TelemetryShutdown telemetry.Shutdown
}

func New(name string) *App {
	initLogger()
	cfg := config.Load()

	utils.InitValidator()

	db := initDatabase(cfg)

	if err := database.Migrate(db); err != nil {
		slog.Error("Failed to migrate database", "error", err)
		os.Exit(1)
	}

	seedSuperAdmin(db)

	app := newFiberApp()
	jwtCfg := initJWTConfig(cfg)
	providers, telemetryShutdown := initTelemetryIfEnabled()

	a := &App{
		Name:              name,
		DB:                db,
		Router:            app,
		Config:            cfg,
		jwtCfg:            jwtCfg,
		Telemetry:         providers,
		TelemetryShutdown: telemetryShutdown,
	}

	a.setupMiddleware()
	a.registerRoutes()
	return a
}

func (a *App) setupMiddleware() {
	a.Router.Use(recover.New())
	a.Router.Use(requestid.New())

	a.setupObservabilityMiddleware()
	a.setupRateLimitingMiddleware()
	a.setupRequestLoggingMiddleware()
	a.setupContextMiddleware()
	a.setupCORSMiddleware()
	a.Router.Use(middleware.AuthResponseAudit())
	a.Router.Use(middleware.AuditHTTPMiddleware(audit.NewLogger(a.DB)))
}

func (a *App) Start() {
	bgCtx, cancelBg := context.WithCancel(context.Background())
	defer cancelBg()

	jobs.StartCleanupJob(bgCtx, a.DB)
	jobs.StartAuditRetentionJob(bgCtx, a.DB)

	addr := fmt.Sprintf("%s:%s", a.Config.Server.Host, a.Config.Server.Port)
	slog.Info("Service starting", "name", a.Name, "address", addr)

	go func() {
		if err := a.Router.Listen(addr); err != nil {
			slog.Info("Fiber stopped", "error", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	a.shutdown(sig, cancelBg)
}
