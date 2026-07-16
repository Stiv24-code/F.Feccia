// Package audit ports backend/audit.py: a persistent, best-effort audit log
// for sensitive actions, kept for fiscal-compliance retention (see
// pkg/jobs/audit_retention.go). Two channels, matching the Python original:
//   - Log: explicit business events not 1:1 with an HTTP mutation (login
//     success/failure, refresh, logout — see pkg/middleware/audit_http.go's
//     exclusion list).
//   - the HTTP middleware in pkg/middleware/audit_http.go, which auto-logs
//     every other API mutation (POST/PUT/PATCH/DELETE).
package audit

import (
	"context"
	"encoding/json"
	"log/slog"
	"regexp"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"fratelli-feccia/internal/models"
)

// Entry mirrors log_audit's keyword arguments.
type Entry struct {
	Action     string
	UserID     *int64
	UserRole   string
	Resource   string
	ResourceID string
	StatusCode int
	Success    bool
	IP         string
	UserAgent  string
	Error      string
	Metadata   map[string]interface{}
}

type Logger struct {
	db *gorm.DB
}

func NewLogger(db *gorm.DB) *Logger {
	return &Logger{db: db}
}

// Log writes an audit record. Never fails the caller — like the Python
// original, a broken audit sink must not break the request it's auditing.
func (l *Logger) Log(ctx context.Context, e Entry) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("panic recovered while writing audit log", "error", r)
		}
	}()

	userAgent := e.UserAgent
	if len(userAgent) > 300 {
		userAgent = userAgent[:300]
	}
	errMsg := e.Error
	if len(errMsg) > 500 {
		errMsg = errMsg[:500]
	}

	var metadataJSON []byte
	if len(e.Metadata) > 0 {
		metadataJSON, _ = json.Marshal(e.Metadata)
	}

	row := models.AuditLog{
		ID: uuid.New(), Timestamp: time.Now(), Action: e.Action, UserID: e.UserID, UserRole: e.UserRole,
		Resource: e.Resource, ResourceID: e.ResourceID, StatusCode: e.StatusCode,
		Success: e.Success, IP: e.IP, UserAgent: userAgent, Error: errMsg,
		Metadata: metadataJSON,
	}

	if err := l.db.WithContext(ctx).Create(&row).Error; err != nil {
		slog.Error("failed to write audit log", "error", err, "action", e.Action)
	}
}

// pathPatterns mirrors backend/audit.py's _PATH_PATTERNS, translated to
// paths relative to the /api/v1 prefix (Python's are relative to /api).
var pathPatterns = []struct {
	re    *regexp.Regexp
	label string
}{
	{regexp.MustCompile(`^/customers/([^/]+)/?$`), "customer"},
	{regexp.MustCompile(`^/customers/?$`), "customer"},
	{regexp.MustCompile(`^/destinations/([^/]+)/?$`), "destination"},
	{regexp.MustCompile(`^/destinations/?$`), "destination"},
	{regexp.MustCompile(`^/vehicles/([^/]+)/gps-position/?$`), "vehicle_gps"},
	{regexp.MustCompile(`^/vehicles/gps-position-by-plate/([^/]+)/?$`), "vehicle_gps"},
	{regexp.MustCompile(`^/vehicles/([^/]+)/?$`), "vehicle"},
	{regexp.MustCompile(`^/vehicles/?$`), "vehicle"},
	{regexp.MustCompile(`^/drivers/([^/]+)/?$`), "driver"},
	{regexp.MustCompile(`^/drivers/?$`), "driver"},
	{regexp.MustCompile(`^/carriers/([^/]+)/?$`), "carrier"},
	{regexp.MustCompile(`^/carriers/?$`), "carrier"},
	{regexp.MustCompile(`^/products/([^/]+)/?$`), "product"},
	{regexp.MustCompile(`^/products/?$`), "product"},
	{regexp.MustCompile(`^/garages/([^/]+)/?$`), "garage"},
	{regexp.MustCompile(`^/garages/?$`), "garage"},
	{regexp.MustCompile(`^/pricelists/([^/]+)/items/([^/]+)/?$`), "pricelist_item"},
	{regexp.MustCompile(`^/pricelists/([^/]+)/items/?$`), "pricelist_item"},
	{regexp.MustCompile(`^/pricelists/([^/]+)/?$`), "pricelist"},
	{regexp.MustCompile(`^/pricelists/?$`), "pricelist"},
	{regexp.MustCompile(`^/orders/([^/]+)/assign/?$`), "order_assign"},
	{regexp.MustCompile(`^/orders/([^/]+)/close/?$`), "order_close"},
	{regexp.MustCompile(`^/orders/([^/]+)/?$`), "order"},
	{regexp.MustCompile(`^/orders/?$`), "order"},
	{regexp.MustCompile(`^/trips/([^/]+)/complete/?$`), "trip_complete"},
	{regexp.MustCompile(`^/trips/([^/]+)/add-order/?$`), "trip_add_order"},
	{regexp.MustCompile(`^/trips/?$`), "trip"},
	{regexp.MustCompile(`^/invoices/([^/]+)/finalize/?$`), "invoice_finalize"},
	{regexp.MustCompile(`^/invoices/([^/]+)/?$`), "invoice"},
	{regexp.MustCompile(`^/invoices/?$`), "invoice"},
	{regexp.MustCompile(`^/driver-unavailability/([^/]+)/?$`), "driver_unavailability"},
	{regexp.MustCompile(`^/driver-unavailability/?$`), "driver_unavailability"},
	{regexp.MustCompile(`^/admin/users/([^/]+)/?$`), "user"},
	{regexp.MustCompile(`^/admin/users/?$`), "user"},
	{regexp.MustCompile(`^/auth/register/?$`), "user"},
}

// ClassifyPath mirrors _classify_path: returns (resource_label, resource_id).
func ClassifyPath(path string) (string, string) {
	for _, p := range pathPatterns {
		m := p.re.FindStringSubmatch(path)
		if m == nil {
			continue
		}
		resourceID := ""
		if len(m) > 1 {
			resourceID = m[len(m)-1]
		}
		return p.label, resourceID
	}
	return "unknown", ""
}
