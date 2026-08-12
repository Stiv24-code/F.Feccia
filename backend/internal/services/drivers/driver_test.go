package drivers

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
	if err := db.AutoMigrate(&models.Driver{}, &models.DriverUnavailability{}); err != nil {
		t.Fatalf("failed to migrate Driver: %v", err)
	}
	return db
}

func TestDriverService_CRUD_AndSearchByNomeOrCognome(t *testing.T) {
	ctx := context.Background()
	svc := NewDriverService(newTestDB(t))

	created, err := svc.Create(ctx, dto.DriverRequest{Nome: "Mario", Cognome: "Rossi"})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	allPage := utils.PageParams{Page: 1, Limit: 20}

	byNome, _, err := svc.List(ctx, "mario", allPage)
	if err != nil || len(byNome) != 1 {
		t.Fatalf("expected search by nome to match, got %v (err=%v)", byNome, err)
	}
	byCognome, _, err := svc.List(ctx, "rossi", allPage)
	if err != nil || len(byCognome) != 1 {
		t.Fatalf("expected search by cognome to match, got %v (err=%v)", byCognome, err)
	}

	updated, err := svc.Update(ctx, created.ID, dto.DriverRequest{Nome: "Mario", Cognome: "Verdi"})
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
	if updated.Cognome != "Verdi" {
		t.Fatalf("expected updated cognome, got %q", updated.Cognome)
	}

	if err := svc.Delete(ctx, created.ID); err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}
	list, _, err := svc.List(ctx, "", allPage)
	if err != nil || len(list) != 0 {
		t.Fatalf("expected 0 drivers after delete, got %v (err=%v)", list, err)
	}
}
