package audit

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"fratelli-feccia/internal/models"
)

func newTestDB(t *testing.T) *gorm.DB {
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

func TestLogger_Log_WritesRecord(t *testing.T) {
	db := newTestDB(t)
	logger := NewLogger(db)

	userID := int64(7)
	logger.Log(context.Background(), Entry{
		Action: "auth.login", UserID: &userID, UserRole: "admin", Resource: "user",
		ResourceID: "7", StatusCode: 200, Success: true, IP: "1.2.3.4", UserAgent: "test-agent",
		Metadata: map[string]interface{}{"email": "admin@example.it"},
	})

	var rows []models.AuditLog
	if err := db.Find(&rows).Error; err != nil {
		t.Fatalf("failed to read back audit logs: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 audit row, got %d", len(rows))
	}
	row := rows[0]
	if row.Action != "auth.login" || row.UserID == nil || *row.UserID != 7 || !row.Success {
		t.Fatalf("unexpected audit row: %+v", row)
	}
	var meta map[string]interface{}
	if err := json.Unmarshal(row.Metadata, &meta); err != nil {
		t.Fatalf("failed to unmarshal metadata: %v", err)
	}
	if meta["email"] != "admin@example.it" {
		t.Fatalf("expected metadata to round-trip, got %+v", meta)
	}
}

func TestLogger_Log_TruncatesLongFields(t *testing.T) {
	db := newTestDB(t)
	logger := NewLogger(db)

	longUA := make([]byte, 1000)
	for i := range longUA {
		longUA[i] = 'a'
	}
	longErr := make([]byte, 1000)
	for i := range longErr {
		longErr[i] = 'b'
	}

	logger.Log(context.Background(), Entry{Action: "test", UserAgent: string(longUA), Error: string(longErr)})

	var row models.AuditLog
	if err := db.First(&row).Error; err != nil {
		t.Fatalf("failed to read back audit log: %v", err)
	}
	if len(row.UserAgent) != 300 {
		t.Fatalf("expected user_agent truncated to 300 chars, got %d", len(row.UserAgent))
	}
	if len(row.Error) != 500 {
		t.Fatalf("expected error truncated to 500 chars, got %d", len(row.Error))
	}
}

func TestLogger_Log_NeverPanicsOnNilDB(t *testing.T) {
	logger := NewLogger(nil)
	// Must not panic even though the underlying Create() call will fail.
	logger.Log(context.Background(), Entry{Action: "test"})
}

func TestClassifyPath(t *testing.T) {
	cases := []struct {
		path      string
		wantLabel string
		wantResID string
	}{
		{"/customers", "customer", ""},
		{"/customers/abc-123", "customer", "abc-123"},
		{"/vehicles/abc/gps-position", "vehicle_gps", "abc"},
		{"/vehicles/gps-position-by-plate/AB123CD", "vehicle_gps", "AB123CD"},
		{"/orders/xyz/assign", "order_assign", "xyz"},
		{"/orders/xyz/close", "order_close", "xyz"},
		{"/orders/xyz", "order", "xyz"},
		{"/pricelists/pl1/items/it1", "pricelist_item", "it1"},
		{"/invoices/inv1/finalize", "invoice_finalize", "inv1"},
		{"/admin/users/5", "user", "5"},
		{"/auth/register", "user", ""},
		{"/something/weird", "unknown", ""},
	}
	for _, tc := range cases {
		label, resID := ClassifyPath(tc.path)
		if label != tc.wantLabel || resID != tc.wantResID {
			t.Errorf("ClassifyPath(%q) = (%q, %q), want (%q, %q)", tc.path, label, resID, tc.wantLabel, tc.wantResID)
		}
	}
}
