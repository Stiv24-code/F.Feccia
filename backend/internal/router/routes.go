package router

import (
	"context"
	"log/slog"

	"fratelli-feccia/config"
	app_handlers "fratelli-feccia/internal/handlers"
	"fratelli-feccia/internal/services"
	"fratelli-feccia/pkg/audit"
	"fratelli-feccia/pkg/middleware"
	"fratelli-feccia/pkg/s3invoices"
	"fratelli-feccia/pkg/utils"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func SetupRoutes(app *fiber.App, db *gorm.DB, jwtCfg utils.JWTConfig, s3Cfg config.S3Config, routingCfg config.RoutingConfig) {
	s3Client, err := s3invoices.NewClient(context.Background(), s3Cfg)
	if err != nil {
		// Non-fatal: mirrors Python's own resilience posture (S3 archival is
		// a soft dependency of invoice finalization, never a startup blocker).
		slog.Error("failed to initialize S3 invoices client, archival disabled", "error", err)
		s3Client, _ = s3invoices.NewClient(context.Background(), config.S3Config{})
	}

	svc := services.NewService(db, jwtCfg, s3Client, routingCfg.ORSApiKey, routingCfg.ORSBaseURL)
	handlers := app_handlers.NewHandler(svc, audit.NewLogger(db), jwtCfg)

	api := app.Group("/api/v1")

	registerAuthRoutes(api, handlers, jwtCfg)

	// Everything reachable by "cliente" (or by any authenticated role
	// regardless of which one) MUST be registered here, before authAll's
	// Group("", ...) below. Fiber's Use(prefix, mw) — which is what Group
	// with a non-empty handlers list boils down to — leaks forward onto
	// every route sharing that prefix registered afterwards, on the same
	// underlying router, no matter which wrapper variable is used to
	// register it. With prefix "" that's every route in the app: anything
	// registered after authAll silently re-inherits PermitAllRoles(), which
	// excludes "cliente" — no amount of additional inline middleware on that
	// route can undo an already-applied 403. (This exact mistake happened
	// once already — every role started getting 403'd — see git history on
	// this file.)
	registerAuthMeRoute(api, handlers, jwtCfg)
	registerClientPortalRoutes(api, handlers, jwtCfg)
	registerDestinationReadRoutes(api, jwtCfg, handlers)
	registerGeocodeRoutes(api, jwtCfg, handlers)

	authAll := api.Group("", middleware.JWTAuthMiddleware(jwtCfg), middleware.PermitAllRoles())

	registerAuthRegisterRoute(authAll, handlers)
	registerAdminRoutes(authAll, handlers)
	registerCustomerRoutes(authAll, handlers)
	registerDestinationWriteRoutes(authAll, handlers)
	registerCarrierRoutes(authAll, handlers)
	registerGarageRoutes(authAll, handlers)
	registerWashStationRoutes(authAll, handlers)
	registerDriverRoutes(authAll, handlers)
	registerProductRoutes(authAll, handlers)
	registerMasterdataRoutes(authAll, handlers)
	registerAnagraficheRoutes(authAll, handlers)
	registerDriverUnavailabilityRoutes(authAll, handlers)
	registerOrderRoutes(authAll, handlers)
	registerMotriceRoutes(authAll, handlers)
	registerSemirimorchioRoutes(authAll, handlers)
	registerTripRoutes(authAll, handlers)
	registerPriceListRoutes(authAll, handlers)
	registerInvoiceRoutes(authAll, handlers)
	registerDashboardRoutes(authAll, handlers)
	registerMapRoutes(authAll, handlers)
	registerAvailabilityRoutes(authAll, handlers)
	registerExportRoutes(authAll, handlers)
}
