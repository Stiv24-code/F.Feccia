package services

import (
	"context"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"fratelli-feccia/internal/dto"
	"fratelli-feccia/internal/models"
	"fratelli-feccia/pkg/utils"
)

func newAdminTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}
	if err := db.AutoMigrate(&models.User{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	return db
}

func createAdminTestUser(t *testing.T, db *gorm.DB, login, role string) models.User {
	t.Helper()
	u := models.User{Login: login, PasswordHash: "hash", Role: role, Name: "Test"}
	if err := db.Create(&u).Error; err != nil {
		t.Fatalf("failed to seed user: %v", err)
	}
	return u
}

func TestListAllUsers_ReturnsEmailShapedResponse(t *testing.T) {
	db := newAdminTestDB(t)
	svc := NewAdminService(db, utils.JWTConfig{})

	createAdminTestUser(t, db, "admin@example.it", utils.RoleAdmin)
	createAdminTestUser(t, db, "op@example.it", utils.RoleOperatore)

	result, err := svc.ListAllUsers(context.Background())
	if err != nil {
		t.Fatalf("ListAllUsers returned error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 users, got %d", len(result))
	}
	if result[0].Email != "admin@example.it" || !result[0].Active {
		t.Fatalf("expected email-shaped, active-by-default response, got %+v", result[0])
	}
	if result[0].ProfileID != nil {
		t.Fatalf("expected ProfileID nil (profiles out of scope), got %v", result[0].ProfileID)
	}
}

func TestListAllUsers_IncludesCustomerIDForClienteAccounts(t *testing.T) {
	db := newAdminTestDB(t)
	svc := NewAdminService(db, utils.JWTConfig{})

	customerID := uuid.New()
	u := models.User{Login: "client@example.it", PasswordHash: "hash", Role: utils.RoleCliente, Name: "Client", CustomerID: &customerID}
	if err := db.Create(&u).Error; err != nil {
		t.Fatalf("failed to seed cliente user: %v", err)
	}

	result, err := svc.ListAllUsers(context.Background())
	if err != nil {
		t.Fatalf("ListAllUsers returned error: %v", err)
	}
	if len(result) != 1 || result[0].CustomerID == nil || *result[0].CustomerID != customerID.String() {
		t.Fatalf("expected CustomerID %q in the response, got %+v", customerID, result)
	}
}

func TestPatchUser_UpdatesNameAndActive(t *testing.T) {
	db := newAdminTestDB(t)
	svc := NewAdminService(db, utils.JWTConfig{})

	u := createAdminTestUser(t, db, "op@example.it", utils.RoleOperatore)
	newName := "Renamed"
	active := false

	resp, err := svc.PatchUser(context.Background(), u.ID, dto.PatchUserRequest{Name: &newName, Active: &active})
	if err != nil {
		t.Fatalf("PatchUser returned error: %v", err)
	}
	if resp.Name != "Renamed" || resp.Active {
		t.Fatalf("expected updated name and active=false, got %+v", resp)
	}

	var reloaded models.User
	if err := db.First(&reloaded, u.ID).Error; err != nil {
		t.Fatalf("failed to reload user: %v", err)
	}
	if reloaded.Active {
		t.Fatal("expected active=false to actually persist to the database")
	}
}

func TestPatchUser_RejectsNonEmptyProfileID(t *testing.T) {
	db := newAdminTestDB(t)
	svc := NewAdminService(db, utils.JWTConfig{})

	u := createAdminTestUser(t, db, "op@example.it", utils.RoleOperatore)
	profileID := "some-profile-id"

	_, err := svc.PatchUser(context.Background(), u.ID, dto.PatchUserRequest{ProfileID: &profileID})
	apiErr, ok := err.(utils.APIError)
	if !ok || apiErr.StatusCode() != 400 {
		t.Fatalf("expected a 400 APIError for a non-empty profile_id, got %v (%T)", err, err)
	}
}

func TestPatchUser_RejectsEmptyPatch(t *testing.T) {
	db := newAdminTestDB(t)
	svc := NewAdminService(db, utils.JWTConfig{})

	u := createAdminTestUser(t, db, "op@example.it", utils.RoleOperatore)

	_, err := svc.PatchUser(context.Background(), u.ID, dto.PatchUserRequest{})
	apiErr, ok := err.(utils.APIError)
	if !ok || apiErr.StatusCode() != 400 {
		t.Fatalf("expected a 400 APIError for an empty patch, got %v (%T)", err, err)
	}
}

func TestPatchUser_CannotDeactivateLastAdmin(t *testing.T) {
	db := newAdminTestDB(t)
	svc := NewAdminService(db, utils.JWTConfig{})

	admin := createAdminTestUser(t, db, "admin@example.it", utils.RoleAdmin)
	active := false

	_, err := svc.PatchUser(context.Background(), admin.ID, dto.PatchUserRequest{Active: &active})
	apiErr, ok := err.(utils.APIError)
	if !ok || apiErr.StatusCode() != 400 {
		t.Fatalf("expected a 400 APIError when deactivating the last admin, got %v (%T)", err, err)
	}
}

func TestPatchUser_CanDeactivateAdminWhenAnotherAdminExists(t *testing.T) {
	db := newAdminTestDB(t)
	svc := NewAdminService(db, utils.JWTConfig{})

	admin1 := createAdminTestUser(t, db, "admin1@example.it", utils.RoleAdmin)
	createAdminTestUser(t, db, "admin2@example.it", utils.RoleAdmin)
	active := false

	resp, err := svc.PatchUser(context.Background(), admin1.ID, dto.PatchUserRequest{Active: &active})
	if err != nil {
		t.Fatalf("expected deactivation to succeed with another admin present, got error: %v", err)
	}
	if resp.Active {
		t.Fatalf("expected active=false in response, got %+v", resp)
	}
}
