package carriers

import (
	"context"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"fratelli-feccia/internal/dto"
	"fratelli-feccia/internal/models"
	"fratelli-feccia/pkg/utils"
)

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}
	if err := db.AutoMigrate(&models.Carrier{}); err != nil {
		t.Fatalf("failed to migrate Carrier: %v", err)
	}
	return db
}

func TestCarrierService_CRUD(t *testing.T) {
	ctx := context.Background()
	svc := NewCarrierService(newTestDB(t))

	created, err := svc.Create(ctx, dto.CarrierRequest{RagioneSociale: "Trasporti Rossi"})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if !created.Active {
		t.Fatalf("expected new carrier to be active")
	}

	allPage := utils.PageParams{Page: 1, Limit: 20}

	list, _, err := svc.List(ctx, "rossi", allPage)
	if err != nil || len(list) != 1 {
		t.Fatalf("expected 1 matching carrier, got %v (err=%v)", list, err)
	}

	updated, err := svc.Update(ctx, created.ID, dto.CarrierRequest{RagioneSociale: "Trasporti Bianchi"})
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
	if updated.RagioneSociale != "Trasporti Bianchi" {
		t.Fatalf("expected updated name, got %q", updated.RagioneSociale)
	}

	if err := svc.Delete(ctx, created.ID); err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}
	list, _, err = svc.List(ctx, "", allPage)
	if err != nil || len(list) != 0 {
		t.Fatalf("expected 0 carriers after delete, got %v (err=%v)", list, err)
	}
}
