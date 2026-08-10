package app

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"fratelli-feccia/config"
	"fratelli-feccia/internal/services"
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
	Services          *services.Service
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
	a.startMailScrapeJob(bgCtx)

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

// startMailScrapeJob mirrors OrderMesh's startup: the periodic mailbox read
// runs only when the scraping backend is actually usable; otherwise say why,
// so a missing `graphlogin` shows up in the logs instead of failing silently
// every N minutes.
func (a *App) startMailScrapeJob(ctx context.Context) {
	scraper := a.Services.MailScraperGroup.MailScraper
	if scraper.MailboxReady() {
		interval := time.Duration(a.Config.Inbound.ScrapeIntervalMin) * time.Minute
		jobs.StartMailScrapeJob(ctx, scraper, interval)
		slog.Info("scrape automatico attivo", "interval", interval, "backend", scraper.Backend())
	} else if scraper.Backend() == "graph" {
		slog.Info("backend Graph non autenticato: esegui `go run ./cmd/graphlogin` (o imposta GRAPH_CLIENT_SECRET) per attivare lo scraping")
	} else {
		slog.Info("IMAP non configurato: scraping disattivato")
	}
}
