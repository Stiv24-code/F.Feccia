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
	"fratelli-feccia/pkg/utils"
)

func ptr(v float64) *float64 { return &v }

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

	resp, err := svc.Create(ctx, dto.DestinationRequest{Nome: "Deposito Nord", Lat: ptr(45.0), Lng: ptr(9.0)})
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

	nord, err := svc.Create(ctx, dto.DestinationRequest{Nome: "Deposito Nord", Lat: ptr(45.0), Lng: ptr(9.0)})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if _, err := svc.Create(ctx, dto.DestinationRequest{Nome: "Deposito Sud", Lat: ptr(41.0), Lng: ptr(12.0)}); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if err := svc.Delete(ctx, nord.ID); err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}

	allPage := utils.PageParams{Page: 1, Limit: 20}

	all, _, err := svc.List(ctx, "", false, allPage)
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(all) != 1 || all[0].Nome != "Deposito Sud" {
		t.Fatalf("expected only the active Sud destination, got %+v", all)
	}

	filtered, _, err := svc.List(ctx, "sud", false, allPage)
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(filtered) != 1 {
		t.Fatalf("expected search to match Sud, got %+v", filtered)
	}

	withInactive, _, err := svc.List(ctx, "", true, allPage)
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(withInactive) != 2 {
		t.Fatalf("expected include_inactive=true to also return the deactivated Nord, got %+v", withInactive)
	}
}

func TestDestinationService_Update_CanReactivate(t *testing.T) {
	ctx := context.Background()
	svc := NewDestinationService(newTestDB(t))

	created, err := svc.Create(ctx, dto.DestinationRequest{Nome: "Deposito Ovest", Lat: ptr(45.0), Lng: ptr(9.0)})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if err := svc.Delete(ctx, created.ID); err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}

	updated, err := svc.Update(ctx, created.ID, dto.DestinationRequest{Nome: "Deposito Ovest", Lat: ptr(45.0), Lng: ptr(9.0), Active: true})
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
	if !updated.Active {
		t.Fatalf("expected Update with Active:true to reactivate the destination, got %+v", updated)
	}

	all, _, err := svc.List(ctx, "", false, utils.PageParams{Page: 1, Limit: 20})
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(all) != 1 || all[0].ID != created.ID {
		t.Fatalf("expected the reactivated destination to show up in the default (active-only) list, got %+v", all)
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

	created, err := svc.Create(ctx, dto.DestinationRequest{Nome: "Deposito Est", Lat: ptr(45.0), Lng: ptr(9.0)})
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
