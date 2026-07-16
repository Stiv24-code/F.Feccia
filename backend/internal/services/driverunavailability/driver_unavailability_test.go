package driverunavailability

import (
	"context"
	"errors"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
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
	if err := db.AutoMigrate(&models.DriverUnavailability{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	return db
}

func TestDriverUnavailabilityService_CreateDefaultsMotivo(t *testing.T) {
	ctx := context.Background()
	svc := NewDriverUnavailabilityService(newTestDB(t))
	driverID := uuid.New()

	resp, err := svc.Create(ctx, dto.DriverUnavailabilityRequest{
		AutistaID: driverID, DataDa: "2026-01-01", DataA: "2026-01-05",
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if resp.Motivo != "ferie" {
		t.Fatalf("expected default motivo %q, got %q", "ferie", resp.Motivo)
	}
}

func TestDriverUnavailabilityService_List_FiltersByAutistaAndDateOverlap(t *testing.T) {
	ctx := context.Background()
	svc := NewDriverUnavailabilityService(newTestDB(t))
	driverA := uuid.New()
	driverB := uuid.New()

	if _, err := svc.Create(ctx, dto.DriverUnavailabilityRequest{AutistaID: driverA, DataDa: "2026-01-01", DataA: "2026-01-05"}); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if _, err := svc.Create(ctx, dto.DriverUnavailabilityRequest{AutistaID: driverB, DataDa: "2026-03-01", DataA: "2026-03-05"}); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	byDriver, err := svc.List(ctx, driverA, "", "")
	if err != nil || len(byDriver) != 1 {
		t.Fatalf("expected 1 record for driverA, got %v (err=%v)", byDriver, err)
	}

	overlapping, err := svc.List(ctx, uuid.Nil, "2026-01-03", "2026-01-10")
	if err != nil || len(overlapping) != 1 {
		t.Fatalf("expected 1 overlapping record, got %v (err=%v)", overlapping, err)
	}

	nonOverlapping, err := svc.List(ctx, uuid.Nil, "2026-02-01", "2026-02-10")
	if err != nil || len(nonOverlapping) != 0 {
		t.Fatalf("expected 0 overlapping records in February, got %v (err=%v)", nonOverlapping, err)
	}
}

func TestDriverUnavailabilityService_Delete_HardDeleteAndNotFound(t *testing.T) {
	ctx := context.Background()
	svc := NewDriverUnavailabilityService(newTestDB(t))

	created, err := svc.Create(ctx, dto.DriverUnavailabilityRequest{AutistaID: uuid.New(), DataDa: "2026-01-01", DataA: "2026-01-05"})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	if err := svc.Delete(ctx, created.ID); err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}

	list, err := svc.List(ctx, uuid.Nil, "", "")
	if err != nil || len(list) != 0 {
		t.Fatalf("expected record to be hard-deleted, got %v (err=%v)", list, err)
	}

	err = svc.Delete(ctx, uuid.New())
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected gorm.ErrRecordNotFound for missing id, got %v", err)
	}
}
