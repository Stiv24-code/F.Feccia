package jobs

import (
	"context"
	"fmt"
	"fratelli-feccia/internal/models"
	"log/slog"
	"time"

	"gorm.io/gorm"
)

// StartCleanupJob starts a background ticker that hard-deletes soft-deleted records older than 7 days.
func StartCleanupJob(ctx context.Context, db *gorm.DB) {
	go runCleanupJob(ctx, db)
}

func runCleanupJob(ctx context.Context, db *gorm.DB) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("Panic recovered in cleanup job", "error", r)
		}
	}()

	slog.Info("Initializing cleanup job...")
	cleanup(ctx, db)

	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.Info("Cleanup job stopped", "reason", ctx.Err())
			return
		case <-ticker.C:
			cleanup(ctx, db)
		}
	}
}

func cleanup(ctx context.Context, db *gorm.DB) {
	slog.Info("Starting cleanup of soft-deleted records older than 7 days...")

	cutoff := time.Now().Add(-7 * 24 * time.Hour)
	batchSize := 100

	modelsToClean := []interface{}{
		&models.User{},
	}

	for _, model := range modelsToClean {
		totalDeleted := 0
		for {
			select {
			case <-ctx.Done():
				slog.Info("Cleanup job cancelled during execution", "reason", ctx.Err())
				return
			default:
			}

			subQuery := db.Unscoped().
				Model(model).
				Select("id").
				Where("deleted_at IS NOT NULL AND deleted_at < ?", cutoff).
				Limit(batchSize)

			result := db.Unscoped().
				Where("id IN (?)", subQuery).
				Delete(model)

			if result.Error != nil {
				slog.Error("Error cleaning up model", "model", fmt.Sprintf("%T", model), "error", result.Error)
				break
			}

			if result.RowsAffected == 0 {
				break
			}

			totalDeleted += int(result.RowsAffected)
			time.Sleep(50 * time.Millisecond)
		}

		if totalDeleted > 0 {
			slog.Info("Permanently deleted records", "count", totalDeleted, "model", fmt.Sprintf("%T", model))
		}
	}

	slog.Info("Cleanup job finished.")
}
