package app

import (
	"context"
	"errors"
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
		// 26 MB: the PDF import endpoints (/api/v1/pdf/*) accept uploads up
		// to 25 MB (see internal/handlers/pdfimport), plus multipart overhead.
		BodyLimit: 26 * 1024 * 1024,
		ErrorHandler: fiberErrorHandler,
	})
}

// genericServerError è l'unico testo che un 5xx può portare al client.
// Contiene il request_id (nel body, non solo nel log) perché è la cosa che
// l'utente può leggere a voce all'assistenza e che lascia risalire alla riga
// di log con l'errore vero.
const genericServerError = "Errore interno del server. Se il problema persiste contatta l'assistenza indicando il codice richiesta."

// fiberErrorHandler è l'ultimo anello: ci arrivano gli errori che i singoli
// handler non hanno gestito (utils.ErrorResponse/HandleDatabaseError non
// passano di qui: rispondono da soli) e i panic, che middleware/recover
// converte in error. Prima rimandava `err.Error()` così com'era — e per un
// 5xx quel testo è il messaggio del driver Postgres, un path del filesystem
// o il "runtime error: invalid memory address" di un nil pointer: dettagli
// interni che dicono a chi attacca com'è fatto il sistema e all'utente
// niente di utile.
//
// La regola:
//   - *fiber.Error 4xx (404 route inesistente, 405, 413 body troppo grande):
//     il messaggio è del framework, controllato e utile → passa.
//   - tutto il resto (5xx, error generici, panic): testo fisso + request_id
//     nel body; l'errore vero solo nel log, con lo stesso request_id.
func fiberErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	var fe *fiber.Error
	if errors.As(err, &fe) {
		code = fe.Code
	}

	requestID, _ := c.Locals("requestid").(string)
	attrs := []any{
		"method", c.Method(),
		"path", c.Path(),
		"status", code,
		"error", err.Error(),
		"request_id", requestID,
	}

	if code < 500 {
		slog.Warn("request rejected", attrs...)
		return c.Status(code).JSON(fiber.Map{"error": fe.Message})
	}

	slog.Error("request failed", attrs...)
	return c.Status(code).JSON(fiber.Map{
		"error":      genericServerError,
		"request_id": requestID,
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
