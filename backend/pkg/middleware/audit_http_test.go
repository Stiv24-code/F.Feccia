package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"fratelli-feccia/internal/models"
	"fratelli-feccia/pkg/audit"
)

func newAuditTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}
	if err := db.AutoMigrate(&models.AuditLog{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	return db
}

func newAuditTestApp(db *gorm.DB, method, path string, setLocals bool) *fiber.App {
	app := fiber.New()
	handler := func(c *fiber.Ctx) error {
		if setLocals {
			c.Locals("user_id", int64(9))
			c.Locals("role", "admin")
		}
		return c.SendStatus(fiber.StatusOK)
	}
	app.Add(method, path, AuditHTTPMiddleware(audit.NewLogger(db)), handler)
	return app
}

func TestAuditHTTPMiddleware_LogsMutationWithUserContext(t *testing.T) {
	db := newAuditTestDB(t)
	app := newAuditTestApp(db, http.MethodPost, "/api/v1/customers", true)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/customers", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	var rows []models.AuditLog
	if err := db.Find(&rows).Error; err != nil {
		t.Fatalf("failed to read audit logs: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 audit row, got %d", len(rows))
	}
	row := rows[0]
	if row.Resource != "customer" || row.UserID == nil || *row.UserID != 9 || row.UserRole != "admin" {
		t.Fatalf("unexpected audit row: %+v", row)
	}
	if !row.Success || row.StatusCode != 200 {
		t.Fatalf("expected success=true status=200, got %+v", row)
	}
}

func TestAuditHTTPMiddleware_SkipsGETRequests(t *testing.T) {
	db := newAuditTestDB(t)
	app := newAuditTestApp(db, http.MethodGet, "/api/v1/customers", false)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/customers", nil)
	if _, err := app.Test(req, -1); err != nil {
		t.Fatalf("app.Test error: %v", err)
	}

	var count int64
	db.Model(&models.AuditLog{}).Count(&count)
	if count != 0 {
		t.Fatalf("expected no audit rows for a GET request, got %d", count)
	}
}

func TestAuditHTTPMiddleware_SkipsExcludedAuthEndpoints(t *testing.T) {
	db := newAuditTestDB(t)
	app := newAuditTestApp(db, http.MethodPost, "/api/v1/auth/login", false)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil)
	if _, err := app.Test(req, -1); err != nil {
		t.Fatalf("app.Test error: %v", err)
	}

	var count int64
	db.Model(&models.AuditLog{}).Count(&count)
	if count != 0 {
		t.Fatalf("expected auth/login to be excluded from the generic middleware, got %d rows", count)
	}
}

func TestAuditHTTPMiddleware_LogsWithoutUserContextWhenUnauthenticated(t *testing.T) {
	db := newAuditTestDB(t)
	app := newAuditTestApp(db, http.MethodPost, "/api/v1/webhooks/gps/verizon", false)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/webhooks/gps/verizon", nil)
	if _, err := app.Test(req, -1); err != nil {
		t.Fatalf("app.Test error: %v", err)
	}

	var row models.AuditLog
	if err := db.First(&row).Error; err != nil {
		t.Fatalf("expected an audit row for the public webhook mutation: %v", err)
	}
	if row.UserID != nil {
		t.Fatalf("expected nil user_id for an unauthenticated request, got %v", *row.UserID)
	}
}
