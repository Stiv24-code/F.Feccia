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

func SetupRoutes(app *fiber.App, db *gorm.DB, jwtCfg utils.JWTConfig, s3Cfg config.S3Config) {
	s3Client, err := s3invoices.NewClient(context.Background(), s3Cfg)
	if err != nil {
		// Non-fatal: mirrors Python's own resilience posture (S3 archival is
		// a soft dependency of invoice finalization, never a startup blocker).
		slog.Error("failed to initialize S3 invoices client, archival disabled", "error", err)
		s3Client, _ = s3invoices.NewClient(context.Background(), config.S3Config{})
	}

	svc := services.NewService(db, jwtCfg, s3Client)
	handlers := app_handlers.NewHandler(svc, audit.NewLogger(db), jwtCfg)

	api := app.Group("/api/v1")

	registerAuthRoutes(api, handlers, jwtCfg)
	registerWebhookRoutes(api, handlers)

	authAll := api.Group("", middleware.JWTAuthMiddleware(jwtCfg), middleware.PermitAllRoles())

	registerAuthMeRoute(authAll, handlers)
	registerAuthRegisterRoute(authAll, handlers)
	registerAdminRoutes(authAll, handlers)
	registerCustomerRoutes(authAll, handlers)
	registerDestinationRoutes(authAll, handlers)
	registerCarrierRoutes(authAll, handlers)
	registerGarageRoutes(authAll, handlers)
	registerDriverRoutes(authAll, handlers)
	registerProductRoutes(authAll, handlers)
	registerMasterdataRoutes(authAll, handlers)
	registerAnagraficheRoutes(authAll, handlers)
	registerDriverUnavailabilityRoutes(authAll, handlers)
	registerOrderRoutes(authAll, handlers)
	registerVehicleRoutes(authAll, handlers)
	registerTripRoutes(authAll, handlers)
	registerPriceListRoutes(authAll, handlers)
	registerInvoiceRoutes(authAll, handlers)
	registerDashboardRoutes(authAll, handlers)
	registerMapRoutes(authAll, handlers)
	registerAvailabilityRoutes(authAll, handlers)
	registerExportRoutes(authAll, handlers)
}
