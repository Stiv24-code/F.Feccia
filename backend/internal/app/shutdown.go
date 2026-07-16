package app

import (
	"context"
	"log/slog"
	"os"
	"time"
)

func (a *App) shutdown(sig os.Signal, cancelBg context.CancelFunc) {
	slog.Info("Termination signal received", "signal", sig.String())
	slog.Info("Shutting down...")

	if cancelBg != nil {
		cancelBg()
	}

	_ = a.Router.Shutdown()

	if a.TelemetryShutdown != nil {
		telemetryCtx, cancelTelemetry := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelTelemetry()
		if err := a.TelemetryShutdown(telemetryCtx); err != nil {
			slog.Warn("Failed to shutdown telemetry providers cleanly", "error", err)
		}
	}

	sqlDB, _ := a.DB.DB()
	if sqlDB != nil {
		if err := sqlDB.Close(); err != nil {
			slog.Warn("Failed to close database connection pool", "error", err)
		} else {
			slog.Info("Database connection pool closed")
		}
	}

	slog.Info("Shutdown complete")
}
