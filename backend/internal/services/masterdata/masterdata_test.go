package masterdata

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
	if err := db.AutoMigrate(&models.VehicleType{}, &models.AccessoryCost{}, &models.TransportCategory{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	return db
}

func TestMasterdataService_VehicleTypes_ListAndCreate(t *testing.T) {
	ctx := context.Background()
	svc := NewMasterdataService(newTestDB(t))

	if _, err := svc.CreateVehicleType(ctx, dto.VehicleTypeRequest{Nome: "Motrice"}); err != nil {
		t.Fatalf("CreateVehicleType returned error: %v", err)
	}
	list, err := svc.ListVehicleTypes(ctx)
	if err != nil || len(list) != 1 {
		t.Fatalf("expected 1 vehicle type, got %v (err=%v)", list, err)
	}
}

func TestMasterdataService_AccessoryCosts_ListAndCreate(t *testing.T) {
	ctx := context.Background()
	svc := NewMasterdataService(newTestDB(t))

	created, err := svc.CreateAccessoryCost(ctx, dto.AccessoryCostRequest{Nome: "Sponda idraulica", CostoDefault: 25.5})
	if err != nil {
		t.Fatalf("CreateAccessoryCost returned error: %v", err)
	}
	if created.CostoDefault != 25.5 {
		t.Fatalf("expected CostoDefault 25.5, got %v", created.CostoDefault)
	}
	list, err := svc.ListAccessoryCosts(ctx)
	if err != nil || len(list) != 1 {
		t.Fatalf("expected 1 accessory cost, got %v (err=%v)", list, err)
	}
}

func TestMasterdataService_TransportCategories_ListAndCreate(t *testing.T) {
	ctx := context.Background()
	svc := NewMasterdataService(newTestDB(t))

	if _, err := svc.CreateTransportCategory(ctx, dto.TransportCategoryRequest{Nome: "Frigo"}); err != nil {
		t.Fatalf("CreateTransportCategory returned error: %v", err)
	}
	list, err := svc.ListTransportCategories(ctx)
	if err != nil || len(list) != 1 {
		t.Fatalf("expected 1 transport category, got %v (err=%v)", list, err)
	}
}
