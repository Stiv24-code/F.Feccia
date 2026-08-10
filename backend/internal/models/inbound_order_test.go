package models

import (
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Key must normalize exactly like the inbound_orders_ref_client_key index
// (lower + trim on both parts): the scraper's in-memory dedup and the DB
// constraint have to agree on what "the same order" means.
func TestInboundOrder_Key_Normalizes(t *testing.T) {
	a := InboundOrder{Ref: "  ORD-123 ", Client: "ACME S.r.l."}
	b := InboundOrder{Ref: "ord-123", Client: "  acme s.R.L.  "}
	if a.Key() != b.Key() {
		t.Fatalf("expected equal keys, got %q vs %q", a.Key(), b.Key())
	}

	c := InboundOrder{Ref: "ord-124", Client: "acme s.r.l."}
	if a.Key() == c.Key() {
		t.Fatalf("expected different keys for different refs, both %q", a.Key())
	}
}

// AutoMigrate smoke test: the service-layer tests (next integration step)
// will migrate these models on in-memory SQLite exactly like this — a broken
// gorm tag should fail here, not there.
func TestInboundModels_AutoMigrateAndRoundTrip(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&InboundOrder{}, &PdfTemplate{}); err != nil {
		t.Fatalf("AutoMigrate failed: %v", err)
	}

	o := InboundOrder{
		ID:         uuid.New(),
		Client:     "ACME S.r.l.",
		Ref:        "ORD-123",
		Status:     InboundOrderStatusPending,
		Source:     InboundOrderSourceMail,
		ReceivedAt: time.Now(),
	}
	if err := db.Create(&o).Error; err != nil {
		t.Fatalf("failed to create inbound order: %v", err)
	}

	tpl := PdfTemplate{
		ID:      uuid.New(),
		Name:    "ACME layout",
		Client:  "ACME S.r.l.",
		Senders: []byte(`["ordini@acme.it","@acme.it"]`),
		Fields:  []byte(`[{"id":"fld-1","target":"ref","label":"Riferimento","page":0,"x":0.1,"y":0.1,"w":0.2,"h":0.05}]`),
	}
	if err := db.Create(&tpl).Error; err != nil {
		t.Fatalf("failed to create pdf template: %v", err)
	}

	var got PdfTemplate
	if err := db.First(&got, "id = ?", tpl.ID).Error; err != nil {
		t.Fatalf("failed to read back template: %v", err)
	}
	if string(got.Senders) == "" || string(got.Fields) == "" {
		t.Fatalf("expected JSON columns to round-trip, got senders=%q fields=%q", got.Senders, got.Fields)
	}
}
