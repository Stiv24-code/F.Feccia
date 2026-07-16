package garages

import (
	"context"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"fratelli-feccia/internal/dto"
	"fratelli-feccia/internal/models"
)

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

	created, err := svc.Create(ctx, dto.GarageRequest{Nome: "Deposito Centrale"})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	list, err := svc.List(ctx)
	if err != nil || len(list) != 1 {
		t.Fatalf("expected 1 garage, got %v (err=%v)", list, err)
	}

	updated, err := svc.Update(ctx, created.ID, dto.GarageRequest{Nome: "Deposito Rinominato"})
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
	if updated.Nome != "Deposito Rinominato" {
		t.Fatalf("expected updated name, got %q", updated.Nome)
	}

	if err := svc.Delete(ctx, created.ID); err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}
	list, err = svc.List(ctx)
	if err != nil || len(list) != 0 {
		t.Fatalf("expected 0 garages after delete, got %v (err=%v)", list, err)
	}
}
