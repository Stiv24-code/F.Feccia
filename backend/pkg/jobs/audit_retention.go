package jobs

import (
	"context"
	"log/slog"
	"time"

	"gorm.io/gorm"

	"fratelli-feccia/internal/models"
)

// auditRetention mirrors backend/audit.py's AUDIT_RETENTION_SECONDS: 10
// years, matching the Italian fiscal obligation to retain invoice records
// (the longest-lived data in the system). Postgres has no native TTL index
// like Mongo's, so this is enforced by a periodic purge instead.
const auditRetention = 10 * 365.25 * 24 * time.Hour

// StartAuditRetentionJob starts a background ticker that hard-deletes audit
// log rows older than the retention window.
func StartAuditRetentionJob(ctx context.Context, db *gorm.DB) {
	go runAuditRetentionJob(ctx, db)
}

func runAuditRetentionJob(ctx context.Context, db *gorm.DB) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("Panic recovered in audit retention job", "error", r)
		}
	}()

	purgeOldAuditLogs(ctx, db)

	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			purgeOldAuditLogs(ctx, db)
		}
	}
}

func purgeOldAuditLogs(ctx context.Context, db *gorm.DB) {
	cutoff := time.Now().Add(-auditRetention)
	result := db.WithContext(ctx).Where("timestamp < ?", cutoff).Delete(&models.AuditLog{})
	if result.Error != nil {
		slog.Error("Error purging old audit logs", "error", result.Error)
		return
	}
	if result.RowsAffected > 0 {
		slog.Info("Purged expired audit logs", "count", result.RowsAffected, "retention", auditRetention)
	}
}
