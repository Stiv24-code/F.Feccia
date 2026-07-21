package garages

import (
	"context"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"fratelli-feccia/internal/dto"
	"fratelli-feccia/internal/models"
)

func ptr(v float64) *float64 { return &v }

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}
	if err := db.AutoMigrate(&models.Garage{}); err != nil {
		t.Fatalf("failed to migrate Garage: %v", err)
	}
	return db
}

func TestGarageService_CRUD(t *testing.T) {
	ctx := context.Background()
	svc := NewGarageService(newTestDB(t))

	created, err := svc.Create(ctx, dto.GarageRequest{Nome: "Deposito Centrale", Lat: ptr(45.4642), Lng: ptr(9.1900)})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	list, err := svc.List(ctx, false)
	if err != nil || len(list) != 1 {
		t.Fatalf("expected 1 garage, got %v (err=%v)", list, err)
	}

	updated, err := svc.Update(ctx, created.ID, dto.GarageRequest{Nome: "Deposito Rinominato", Lat: ptr(45.4642), Lng: ptr(9.1900), Active: true})
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
	if updated.Nome != "Deposito Rinominato" {
		t.Fatalf("expected updated name, got %q", updated.Nome)
	}

	if err := svc.Delete(ctx, created.ID); err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}
	list, err = svc.List(ctx, false)
	if err != nil || len(list) != 0 {
		t.Fatalf("expected 0 garages after delete, got %v (err=%v)", list, err)
	}

	withInactive, err := svc.List(ctx, true)
	if err != nil || len(withInactive) != 1 {
		t.Fatalf("expected include_inactive=true to still show the deleted garage, got %v (err=%v)", withInactive, err)
	}

	reactivated, err := svc.Update(ctx, created.ID, dto.GarageRequest{Nome: "Deposito Rinominato", Lat: ptr(45.4642), Lng: ptr(9.1900), Active: true})
	if err != nil {
		t.Fatalf("Update (reactivate) returned error: %v", err)
	}
	if !reactivated.Active {
		t.Fatalf("expected Update with Active:true to reactivate the garage, got %+v", reactivated)
	}
	list, err = svc.List(ctx, false)
	if err != nil || len(list) != 1 {
		t.Fatalf("expected the reactivated garage back in the default list, got %v (err=%v)", list, err)
	}
}
