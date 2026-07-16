package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// AuditLog mirrors backend/audit.py's db.audit_logs collection. Retention is
// enforced by pkg/jobs' audit retention job (10 years, matching Python's
// Mongo TTL index) rather than a DB-native TTL, since Postgres has none.
type AuditLog struct {
	ID         uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	Timestamp  time.Time      `gorm:"not null;index" json:"timestamp"`
	Action     string         `gorm:"type:varchar(150);not null" json:"action"`
	UserID     *int64         `gorm:"index" json:"user_id"`
	UserRole   string         `gorm:"type:varchar(50)" json:"user_role"`
	Resource   string         `gorm:"type:varchar(100);index" json:"resource"`
	ResourceID string         `gorm:"type:varchar(100)" json:"resource_id"`
	StatusCode int            `json:"status_code"`
	Success    bool           `gorm:"index" json:"success"`
	IP         string         `gorm:"type:varchar(64)" json:"ip"`
	UserAgent  string         `gorm:"type:varchar(300)" json:"user_agent"`
	Error      string         `gorm:"type:varchar(500)" json:"error"`
	Metadata   datatypes.JSON `json:"metadata"`
}
