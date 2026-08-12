package washstations

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
	if err := db.AutoMigrate(&models.WashStation{}); err != nil {
		t.Fatalf("failed to migrate WashStation: %v", err)
	}
	return db
}

func ptr(v float64) *float64 { return &v }

func TestWashStationService_CRUD(t *testing.T) {
	ctx := context.Background()
	svc := NewWashStationService(newTestDB(t))

	created, err := svc.Create(ctx, dto.WashStationRequest{Nome: "Lavaggio Nord", Lat: ptr(45.4642), Lng: ptr(9.1900)})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if created.Lat == nil || *created.Lat != 45.4642 {
		t.Fatalf("expected lat to round-trip, got %+v", created.Lat)
	}

	allPage := utils.PageParams{Page: 1, Limit: 20}

	list, _, err := svc.List(ctx, false, allPage)
	if err != nil || len(list) != 1 {
		t.Fatalf("expected 1 wash station, got %v (err=%v)", list, err)
	}

	updated, err := svc.Update(ctx, created.ID, dto.WashStationRequest{Nome: "Lavaggio Rinominato", Lat: ptr(45.0), Lng: ptr(9.0), Active: true})
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
	if updated.Nome != "Lavaggio Rinominato" {
		t.Fatalf("expected updated name, got %q", updated.Nome)
	}

	if err := svc.Delete(ctx, created.ID); err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}
	list, _, err = svc.List(ctx, false, allPage)
	if err != nil || len(list) != 0 {
		t.Fatalf("expected 0 wash stations after delete, got %v (err=%v)", list, err)
	}

	withInactive, _, err := svc.List(ctx, true, allPage)
	if err != nil || len(withInactive) != 1 {
		t.Fatalf("expected include_inactive=true to still show the deleted station, got %v (err=%v)", withInactive, err)
	}

	reactivated, err := svc.Update(ctx, created.ID, dto.WashStationRequest{Nome: "Lavaggio Rinominato", Lat: ptr(45.0), Lng: ptr(9.0), Active: true})
	if err != nil {
		t.Fatalf("Update (reactivate) returned error: %v", err)
	}
	if !reactivated.Active {
		t.Fatalf("expected Update with Active:true to reactivate the station, got %+v", reactivated)
	}
	list, _, err = svc.List(ctx, false, allPage)
	if err != nil || len(list) != 1 {
		t.Fatalf("expected the reactivated station back in the default list, got %v (err=%v)", list, err)
	}
}
