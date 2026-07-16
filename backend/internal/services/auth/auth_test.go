package auth

import (
	"testing"
	"time"

	"fratelli-feccia/internal/dto"
	"fratelli-feccia/internal/models"
	"fratelli-feccia/pkg/utils"

	"github.com/glebarez/sqlite"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// testUser is a minimal user model for auth tests.
// We rely on GORM's NamingStrategy to add "public." prefix.
type testUser struct {
	ID           int64  `gorm:"primaryKey;autoIncrement"`
	Login        string `gorm:"uniqueIndex;not null"`
	PasswordHash string `gorm:"not null"`
	Role         string `gorm:"not null"`
	LastLoginAt  *time.Time
}

// newAuthTestDB creates an in-memory SQLite DB and prepares the users table.
func newAuthTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   "", // removed "public." for SQLite
			SingularTable: false,
		},
	})
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	if err := db.AutoMigrate(&testUser{}); err != nil {
		t.Fatalf("failed to migrate testUser: %v", err)
	}

	return db
}

func newAuthServiceForTest(t *testing.T, db *gorm.DB) *AuthService {
	t.Helper()

	// Ensure JWT secrets are set so utils.NewJWTConfig doesn't panic.
	t.Setenv("JWT_ACCESS_SECRET", "test-access-secret")
	t.Setenv("JWT_REFRESH_SECRET", "test-refresh-secret")

	cfg := utils.NewJWTConfig("test-access-secret", "test-refresh-secret", "60", "24")
	return &AuthService{
		db:      db,
		jwtConf: cfg,
	}
}

func seedTestUser(t *testing.T, db *gorm.DB, login, password, role string) testUser {
	t.Helper()

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	u := testUser{
		Login:        login,
		PasswordHash: string(hash),
		Role:         role,
	}
	if err := db.Create(&u).Error; err != nil {
		t.Fatalf("failed to seed user: %v", err)
	}
	return u
}

func TestAuthService_Login_InvalidLogin(t *testing.T) {
	db := newAuthTestDB(t)
	svc := newAuthServiceForTest(t, db)

	tokens, err := svc.Login("nonexistent", "secret")
	if err == nil || err.Error() != "invalid credentials" {
		t.Fatalf("expected invalid credentials error, got tokens=%+v err=%v", tokens, err)
	}
}

func TestAuthService_Login_WrongPassword(t *testing.T) {
	db := newAuthTestDB(t)
	svc := newAuthServiceForTest(t, db)

	_ = seedTestUser(t, db, "user1", "correct", "user")

	tokens, err := svc.Login("user1", "wrong")
	if err == nil || err.Error() != "invalid credentials" {
		t.Fatalf("expected invalid credentials error, got tokens=%+v err=%v", tokens, err)
	}
}

func TestAuthService_Refresh_InvalidToken(t *testing.T) {
	db := newAuthTestDB(t)
	svc := newAuthServiceForTest(t, db)

	tokens, err := svc.Refresh("invalid-token")
	if err == nil || err.Error() != "invalid refresh token" {
		t.Fatalf("expected invalid refresh token error, got tokens=%+v err=%v", tokens, err)
	}
}

// newModelsTestDB migrates the real models.User (now table-name-portable,
// see models/user.go) so positive-path Login/Register scenarios can be
// exercised against SQLite, unlike the testUser-based negative-path tests above.
func newModelsTestDB(t *testing.T) *gorm.DB {
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

func seedModelsUser(t *testing.T, db *gorm.DB, login, password, role string, active bool) models.User {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}
	u := models.User{Login: login, PasswordHash: string(hash), Role: role}
	if err := db.Create(&u).Error; err != nil {
		t.Fatalf("failed to seed user: %v", err)
	}
	// Active has a `default:true` GORM tag, so Create() silently skips it
	// when the Go zero value (false) is set — a well-known GORM footgun.
	// Force it via an explicit UPDATE instead.
	if !active {
		if err := db.Model(&u).Update("active", false).Error; err != nil {
			t.Fatalf("failed to set active=false: %v", err)
		}
		u.Active = false
	}
	return u
}

func TestAuthService_Login_Success(t *testing.T) {
	db := newModelsTestDB(t)
	svc := newAuthServiceForTest(t, db)
	seedModelsUser(t, db, "active@example.it", "correct-password", "operatore", true)

	result, err := svc.Login("active@example.it", "correct-password")
	if err != nil {
		t.Fatalf("expected successful login, got error: %v", err)
	}
	if result.User.Email != "active@example.it" || !result.User.Active {
		t.Fatalf("expected active user in response, got %+v", result.User)
	}
}

func TestAuthService_Login_DisabledUserRejected(t *testing.T) {
	db := newModelsTestDB(t)
	svc := newAuthServiceForTest(t, db)
	seedModelsUser(t, db, "disabled@example.it", "correct-password", "operatore", false)

	result, err := svc.Login("disabled@example.it", "correct-password")
	if result != nil {
		t.Fatalf("expected nil result for disabled user, got %+v", result)
	}
	apiErr, ok := err.(utils.APIError)
	if !ok || apiErr.StatusCode() != 403 {
		t.Fatalf("expected a 403 APIError, got %v (%T)", err, err)
	}
}

func TestAuthService_Register_CreatesActiveUser(t *testing.T) {
	db := newModelsTestDB(t)
	svc := newAuthServiceForTest(t, db)

	resp, err := svc.Register(dto.RegisterRequest{
		Email: "new@example.it", Name: "New User", Password: "supersecretpw12", Role: "operatore",
	})
	if err != nil {
		t.Fatalf("Register returned error: %v", err)
	}
	if resp.Email != "new@example.it" || resp.Role != "operatore" || !resp.Active {
		t.Fatalf("unexpected register response: %+v", resp)
	}

	// The new user must actually be able to log in afterwards.
	if _, err := svc.Login("new@example.it", "supersecretpw12"); err != nil {
		t.Fatalf("expected the newly registered user to be able to log in, got error: %v", err)
	}
}

func TestAuthService_Register_DuplicateEmailRejected(t *testing.T) {
	db := newModelsTestDB(t)
	svc := newAuthServiceForTest(t, db)
	seedModelsUser(t, db, "dup@example.it", "whatever12345", "operatore", true)

	resp, err := svc.Register(dto.RegisterRequest{
		Email: "dup@example.it", Name: "Someone Else", Password: "anotherpassword12", Role: "operatore",
	})
	if resp != nil {
		t.Fatalf("expected nil response for duplicate email, got %+v", resp)
	}
	apiErr, ok := err.(utils.APIError)
	if !ok || apiErr.StatusCode() != 400 {
		t.Fatalf("expected a 400 APIError, got %v (%T)", err, err)
	}
}
