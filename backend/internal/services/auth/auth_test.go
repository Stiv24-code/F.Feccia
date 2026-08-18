package auth

import (
	"context"
	"strings"
	"testing"
	"time"

	"fratelli-feccia/internal/dto"
	"fratelli-feccia/internal/models"
	"fratelli-feccia/pkg/utils"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
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
		// mailer left nil: no SMTP in tests, so RegisterClient takes the
		// immediate-login fallback path — same shape as before this test was
		// written, no per-test wiring needed.
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
	if err := db.AutoMigrate(&models.User{}, &models.Customer{}); err != nil {
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

func TestAuthService_Register_ClienteRoleRequiresExistingCustomer(t *testing.T) {
	db := newModelsTestDB(t)
	svc := newAuthServiceForTest(t, db)

	resp, err := svc.Register(dto.RegisterRequest{
		Email: "no-customer@example.it", Name: "Someone", Password: "supersecretpw12", Role: utils.RoleCliente,
	})
	if resp != nil {
		t.Fatalf("expected nil response without a customer_id, got %+v", resp)
	}
	apiErr, ok := err.(utils.APIError)
	if !ok || apiErr.StatusCode() != 400 {
		t.Fatalf("expected a 400 APIError, got %v (%T)", err, err)
	}
}

func TestAuthService_Register_ClienteRoleLinksToExistingCustomer(t *testing.T) {
	db := newModelsTestDB(t)
	svc := newAuthServiceForTest(t, db)

	customer := models.Customer{ID: uuid.New(), RagioneSociale: "Existing S.p.A."}
	if err := db.Create(&customer).Error; err != nil {
		t.Fatalf("failed to seed customer: %v", err)
	}
	customerID := customer.ID.String()

	resp, err := svc.Register(dto.RegisterRequest{
		Email: "admin-created-client@example.it", Name: "Referente", Password: "supersecretpw12",
		Role: utils.RoleCliente, CustomerID: &customerID,
	})
	if err != nil {
		t.Fatalf("Register returned error: %v", err)
	}
	if resp.Role != utils.RoleCliente || resp.CustomerID == nil || *resp.CustomerID != customerID {
		t.Fatalf("expected the new account linked to the given customer, got %+v", resp)
	}

	// Admin-provisioned — must be able to log in immediately, no email
	// verification gate (unlike self-service RegisterClient).
	if _, err := svc.Login("admin-created-client@example.it", "supersecretpw12"); err != nil {
		t.Fatalf("expected the admin-created client account to be able to log in, got error: %v", err)
	}
}

func TestAuthService_RegisterClient_CreatesLinkedCustomerAndLogsIn(t *testing.T) {
	db := newModelsTestDB(t)
	svc := newAuthServiceForTest(t, db)

	pending, result, err := svc.RegisterClient(dto.ClientRegisterRequest{
		RagioneSociale: "Acme S.r.l.", Email: "cliente@example.it",
		Name: "Mario Rossi", Password: "supersecretpw12",
	})
	if err != nil {
		t.Fatalf("RegisterClient returned error: %v", err)
	}
	if pending != nil {
		t.Fatalf("expected no pending verification (no mailer configured in this test), got %+v", pending)
	}
	if result.User.Role != utils.RoleCliente || !result.User.Active {
		t.Fatalf("expected an active cliente account, got %+v", result.User)
	}
	if result.User.CustomerID == nil || *result.User.CustomerID == "" {
		t.Fatalf("expected CustomerID to be set on the response, got %+v", result.User)
	}
	if result.AccessToken == "" {
		t.Fatalf("expected RegisterClient to auto-login (non-empty access token)")
	}

	var customer models.Customer
	if err := db.First(&customer, "id = ?", *result.User.CustomerID).Error; err != nil {
		t.Fatalf("expected a Customer row linked to the new user, got error: %v", err)
	}
	if customer.RagioneSociale != "Acme S.r.l." {
		t.Fatalf("expected the anagrafica fields to be persisted, got %+v", customer)
	}

	// The claim must round-trip through the access token too, not just the
	// response body — routes_client_portal.go relies on it.
	claims, err := utils.ParseAccessToken(result.AccessToken, svc.jwtConf)
	if err != nil {
		t.Fatalf("ParseAccessToken returned error: %v", err)
	}
	if claims.Role != utils.RoleCliente || claims.CustomerID != *result.User.CustomerID {
		t.Fatalf("expected access token claims to carry role=cliente + matching customer_id, got %+v", claims)
	}

	if _, err := svc.Login("cliente@example.it", "supersecretpw12"); err != nil {
		t.Fatalf("expected the newly registered client to be able to log in, got error: %v", err)
	}
}

func TestAuthService_RegisterClient_DuplicateEmailRejected(t *testing.T) {
	db := newModelsTestDB(t)
	svc := newAuthServiceForTest(t, db)
	seedModelsUser(t, db, "dup-client@example.it", "whatever12345", "operatore", true)

	pending, result, err := svc.RegisterClient(dto.ClientRegisterRequest{
		RagioneSociale: "Beta S.p.A.", Email: "dup-client@example.it",
		Name: "Someone", Password: "anotherpassword12",
	})
	if pending != nil || result != nil {
		t.Fatalf("expected no result for duplicate email, got pending=%+v login=%+v", pending, result)
	}
	apiErr, ok := err.(utils.APIError)
	if !ok || apiErr.StatusCode() != 400 {
		t.Fatalf("expected a 400 APIError, got %v (%T)", err, err)
	}

	var count int64
	db.Model(&models.Customer{}).Where("ragione_sociale = ?", "Beta S.p.A.").Count(&count)
	if count != 0 {
		t.Fatalf("expected the Customer creation to be rolled back too, found %d rows", count)
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

// fakeMailer records every mail it's asked to send instead of touching SMTP —
// stands in for mailer.MailerService in tests exercising the
// verification-mail path.
type fakeMailer struct {
	to, subject, body string
	sendErr           error
}

func (m *fakeMailer) SendHTML(_ context.Context, to, subject, body string) error {
	m.to, m.subject, m.body = to, subject, body
	return m.sendErr
}

func TestAuthService_RegisterClient_WithMailerSendsVerificationAndBlocksLogin(t *testing.T) {
	db := newModelsTestDB(t)
	svc := newAuthServiceForTest(t, db)
	fm := &fakeMailer{}
	svc.mailer = fm
	svc.appBaseURL = "https://ffeccia.homes"

	pending, login, err := svc.RegisterClient(dto.ClientRegisterRequest{
		RagioneSociale: "Verify S.r.l.", Email: "pending-client@example.it",
		Name: "Mario Rossi", Password: "supersecretpw12",
	})
	if err != nil {
		t.Fatalf("RegisterClient returned error: %v", err)
	}
	if login != nil {
		t.Fatalf("expected no immediate login while a mailer is configured, got %+v", login)
	}
	if pending == nil || pending.Verified || pending.Email != "pending-client@example.it" {
		t.Fatalf("expected a pending verification result, got %+v", pending)
	}
	if fm.to != "pending-client@example.it" {
		t.Fatalf("expected the verification mail to go to the new account's email, got %q", fm.to)
	}

	// The account exists but must not be able to log in yet.
	if _, err := svc.Login("pending-client@example.it", "supersecretpw12"); err == nil {
		t.Fatalf("expected login to be refused before email verification")
	} else if apiErr, ok := err.(utils.APIError); !ok || apiErr.StatusCode() != 403 {
		t.Fatalf("expected a 403 APIError, got %v (%T)", err, err)
	}

	// Extract the token from the mailed link and confirm it.
	idx := strings.Index(fm.body, "token=")
	if idx == -1 {
		t.Fatalf("expected the mail body to contain a verification link, got %q", fm.body)
	}
	token := fm.body[idx+len("token="):]
	if end := strings.IndexAny(token, "\n\r \""); end != -1 {
		token = token[:end]
	}
	result, err := svc.VerifyEmail(token)
	if err != nil {
		t.Fatalf("VerifyEmail returned error: %v", err)
	}
	if result.User.Email != "pending-client@example.it" || result.AccessToken == "" {
		t.Fatalf("expected VerifyEmail to log the account in, got %+v", result)
	}

	// Now a normal Login must succeed too.
	if _, err := svc.Login("pending-client@example.it", "supersecretpw12"); err != nil {
		t.Fatalf("expected login to succeed after verification, got error: %v", err)
	}
}

func TestAuthService_VerifyEmail_InvalidTokenRejected(t *testing.T) {
	db := newModelsTestDB(t)
	svc := newAuthServiceForTest(t, db)

	if _, err := svc.VerifyEmail("not-a-real-token"); err == nil {
		t.Fatalf("expected an error for an unknown token")
	}
}
