package customers

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

func newCustomerTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}
	if err := db.AutoMigrate(&models.Customer{}); err != nil {
		t.Fatalf("failed to migrate Customer: %v", err)
	}
	return db
}

func validRequest() dto.CustomerRequest {
	return dto.CustomerRequest{
		RagioneSociale: "Acme S.r.l.",
		Citta:          "Milano",
		PartitaIva:     "12345678901",
	}
}

func TestCustomerService_Create_DefaultsNazioneAndActive(t *testing.T) {
	ctx := context.Background()
	svc := NewCustomerService(newCustomerTestDB(t))

	resp, err := svc.Create(ctx, validRequest())
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if resp.Nazione != "Italia" {
		t.Fatalf("expected default Nazione %q, got %q", "Italia", resp.Nazione)
	}
	if !resp.Active {
		t.Fatalf("expected new customer to be active")
	}
	if resp.ID == uuid.Nil {
		t.Fatalf("expected a generated ID")
	}
}

func TestCustomerService_List_FiltersInactiveAndSearches(t *testing.T) {
	ctx := context.Background()
	svc := NewCustomerService(newCustomerTestDB(t))

	acme, err := svc.Create(ctx, dto.CustomerRequest{RagioneSociale: "Acme S.r.l."})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if _, err := svc.Create(ctx, dto.CustomerRequest{RagioneSociale: "Beta S.p.A."}); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if err := svc.Delete(ctx, acme.ID); err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}

	all, err := svc.List(ctx, "")
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(all) != 1 || all[0].RagioneSociale != "Beta S.p.A." {
		t.Fatalf("expected only the active Beta customer, got %+v", all)
	}

	filtered, err := svc.List(ctx, "beta")
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(filtered) != 1 {
		t.Fatalf("expected search to match Beta, got %+v", filtered)
	}

	none, err := svc.List(ctx, "nomatch")
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(none) != 0 {
		t.Fatalf("expected no matches, got %+v", none)
	}
}

func TestCustomerService_GetByID_NotFoundReturnsNilNil(t *testing.T) {
	ctx := context.Background()
	svc := NewCustomerService(newCustomerTestDB(t))

	resp, err := svc.GetByID(ctx, uuid.New())
	if err != nil {
		t.Fatalf("expected no error for missing customer, got %v", err)
	}
	if resp != nil {
		t.Fatalf("expected nil response for missing customer, got %+v", resp)
	}
}

func TestCustomerService_Update_ReplacesFields(t *testing.T) {
	ctx := context.Background()
	svc := NewCustomerService(newCustomerTestDB(t))

	created, err := svc.Create(ctx, validRequest())
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	updated, err := svc.Update(ctx, created.ID, dto.CustomerRequest{
		RagioneSociale: "Acme S.r.l. (aggiornata)",
		Citta:          "Torino",
	})
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
	if updated.RagioneSociale != "Acme S.r.l. (aggiornata)" || updated.Citta != "Torino" {
		t.Fatalf("unexpected updated customer: %+v", updated)
	}
	if updated.PartitaIva != "" {
		t.Fatalf("expected full-replace update to clear partita_iva, got %q", updated.PartitaIva)
	}
}

func TestCustomerService_Update_NotFound(t *testing.T) {
	ctx := context.Background()
	svc := NewCustomerService(newCustomerTestDB(t))

	_, err := svc.Update(ctx, uuid.New(), validRequest())
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected gorm.ErrRecordNotFound, got %v", err)
	}
}

func TestCustomerService_Delete_IsLogical(t *testing.T) {
	ctx := context.Background()
	svc := NewCustomerService(newCustomerTestDB(t))

	created, err := svc.Create(ctx, validRequest())
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
		t.Fatalf("expected customer to still exist but be inactive, got %+v", resp)
	}
}

func TestCustomerService_Delete_MissingIDIsNoop(t *testing.T) {
	ctx := context.Background()
	svc := NewCustomerService(newCustomerTestDB(t))

	if err := svc.Delete(ctx, uuid.New()); err != nil {
		t.Fatalf("expected no error deleting a missing customer, got %v", err)
	}
}
