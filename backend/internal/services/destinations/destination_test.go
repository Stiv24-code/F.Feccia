package destinations

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
	if err := db.AutoMigrate(&models.Destination{}); err != nil {
		t.Fatalf("failed to migrate Destination: %v", err)
	}
	return db
}

func TestDestinationService_CreateDefaultsNazione(t *testing.T) {
	ctx := context.Background()
	svc := NewDestinationService(newTestDB(t))

	resp, err := svc.Create(ctx, dto.DestinationRequest{Nome: "Deposito Nord"})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if resp.Nazione != "Italia" {
		t.Fatalf("expected default Nazione %q, got %q", "Italia", resp.Nazione)
	}
	if !resp.Active {
		t.Fatalf("expected new destination to be active")
	}
}

func TestDestinationService_List_FiltersInactiveAndSearches(t *testing.T) {
	ctx := context.Background()
	svc := NewDestinationService(newTestDB(t))

	nord, err := svc.Create(ctx, dto.DestinationRequest{Nome: "Deposito Nord"})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if _, err := svc.Create(ctx, dto.DestinationRequest{Nome: "Deposito Sud"}); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if err := svc.Delete(ctx, nord.ID); err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}

	all, err := svc.List(ctx, "")
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(all) != 1 || all[0].Nome != "Deposito Sud" {
		t.Fatalf("expected only the active Sud destination, got %+v", all)
	}

	filtered, err := svc.List(ctx, "sud")
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(filtered) != 1 {
		t.Fatalf("expected search to match Sud, got %+v", filtered)
	}
}

func TestDestinationService_GetByID_NotFoundReturnsNilNil(t *testing.T) {
	ctx := context.Background()
	svc := NewDestinationService(newTestDB(t))

	resp, err := svc.GetByID(ctx, uuid.New())
	if err != nil {
		t.Fatalf("expected no error for missing destination, got %v", err)
	}
	if resp != nil {
		t.Fatalf("expected nil response for missing destination, got %+v", resp)
	}
}

func TestDestinationService_Update_NotFound(t *testing.T) {
	ctx := context.Background()
	svc := NewDestinationService(newTestDB(t))

	_, err := svc.Update(ctx, uuid.New(), dto.DestinationRequest{Nome: "X"})
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected gorm.ErrRecordNotFound, got %v", err)
	}
}

func TestDestinationService_Delete_IsLogical(t *testing.T) {
	ctx := context.Background()
	svc := NewDestinationService(newTestDB(t))

	created, err := svc.Create(ctx, dto.DestinationRequest{Nome: "Deposito Est"})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if err := svc.Delete(ctx, created.ID); err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}

	resp, err := svc.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetByID returned error: %v", err)
	}
	if resp == nil || resp.Active {
		t.Fatalf("expected destination to still exist but be inactive, got %+v", resp)
	}
}
