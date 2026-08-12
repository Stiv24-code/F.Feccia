package products

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
	if err := db.AutoMigrate(&models.Product{}); err != nil {
		t.Fatalf("failed to migrate Product: %v", err)
	}
	return db
}

func TestProductService_CreateDefaultsUnitaMisura(t *testing.T) {
	ctx := context.Background()
	svc := NewProductService(newTestDB(t))

	resp, err := svc.Create(ctx, dto.ProductRequest{Codice: "P001", Descrizione: "Cemento"})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if resp.UnitaMisura != "Kg" {
		t.Fatalf("expected default unita_misura %q, got %q", "Kg", resp.UnitaMisura)
	}
}

func TestProductService_DuplicateCodiceReturnsConflict(t *testing.T) {
	ctx := context.Background()
	svc := NewProductService(newTestDB(t))

	if _, err := svc.Create(ctx, dto.ProductRequest{Codice: "P001", Descrizione: "Cemento"}); err != nil {
		t.Fatalf("first Create returned error: %v", err)
	}

	_, err := svc.Create(ctx, dto.ProductRequest{Codice: "P001", Descrizione: "Duplicato"})
	if err == nil {
		t.Fatal("expected a unique constraint error for duplicate codice")
	}
}

func TestProductService_SearchByCodiceOrDescrizione(t *testing.T) {
	ctx := context.Background()
	svc := NewProductService(newTestDB(t))

	if _, err := svc.Create(ctx, dto.ProductRequest{Codice: "P001", Descrizione: "Cemento"}); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	allPage := utils.PageParams{Page: 1, Limit: 20}

	byCodice, _, err := svc.List(ctx, "p001", allPage)
	if err != nil || len(byCodice) != 1 {
		t.Fatalf("expected search by codice to match, got %v (err=%v)", byCodice, err)
	}
	byDescrizione, _, err := svc.List(ctx, "cemento", allPage)
	if err != nil || len(byDescrizione) != 1 {
		t.Fatalf("expected search by descrizione to match, got %v (err=%v)", byDescrizione, err)
	}
}
