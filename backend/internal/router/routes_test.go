package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"fratelli-feccia/config"
	"fratelli-feccia/pkg/database"
	"fratelli-feccia/pkg/utils"

	"github.com/glebarez/sqlite"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// newRoutesTestApp wires the real SetupRoutes against an in-memory SQLite DB
// — this exercises the actual Fiber Group/Use composition, which is exactly
// what a unit test of an individual middleware function (RequireRole,
// PermitAllRoles, ...) in isolation cannot catch: those all passed on their
// own while the *wiring* in this package silently 403'd every role, because
// registering a second/third Group("", ...) leaks its middleware onto every
// other route sharing prefix "" (see the comment above authAll in routes.go).
func newRoutesTestApp(t *testing.T) (*fiber.App, utils.JWTConfig) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}
	if err := database.Migrate(db); err != nil {
		t.Fatalf("failed to migrate test database: %v", err)
	}

	jwtCfg := utils.NewJWTConfig("test-access-secret", "test-refresh-secret", "60", "24")
	app := fiber.New()
	SetupRoutes(app, db, jwtCfg, config.S3Config{}, config.RoutingConfig{})
	return app, jwtCfg
}

func doRoutesRequest(t *testing.T, app *fiber.App, method, path, bearerToken string) int {
	t.Helper()
	req := httptest.NewRequest(method, path, nil)
	if bearerToken != "" {
		req.Header.Set("Authorization", "Bearer "+bearerToken)
	}
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test error: %v", err)
	}
	return resp.StatusCode
}

// TestSetupRoutes_StaffAndClienteAreNotMutuallyExclusive is the regression
// test for the incident: after adding authAny/authClient as extra
// api.Group("", ...) calls, EVERY role (including admin) started getting
// 403 on EVERY route, because PermitAllRoles() and RequireRole(cliente)
// both ended up applying to every request regardless of role. Neither
// assertion below should ever be a 403.
func TestSetupRoutes_StaffAndClienteAreNotMutuallyExclusive(t *testing.T) {
	app, jwtCfg := newRoutesTestApp(t)

	adminTokens, err := utils.GenerateTokenPair(1, utils.RoleAdmin, "", jwtCfg)
	if err != nil {
		t.Fatalf("GenerateTokenPair (admin) returned error: %v", err)
	}
	clienteTokens, err := utils.GenerateTokenPair(2, utils.RoleCliente, "33333333-3333-3333-3333-333333333333", jwtCfg)
	if err != nil {
		t.Fatalf("GenerateTokenPair (cliente) returned error: %v", err)
	}

	// A staff-only route: admin must NOT be 403'd on it.
	if status := doRoutesRequest(t, app, http.MethodGet, "/api/v1/customers", adminTokens.AccessToken); status == http.StatusForbidden {
		t.Fatalf("expected admin to reach a staff route without 403, got %d", status)
	}

	// The same staff-only route: cliente correctly IS 403'd (this is the one
	// case that SHOULD be forbidden).
	if status := doRoutesRequest(t, app, http.MethodGet, "/api/v1/customers", clienteTokens.AccessToken); status != http.StatusForbidden {
		t.Fatalf("expected cliente to be forbidden on a staff route, got %d", status)
	}

	// Client-portal routes: cliente must NOT be 403'd on any of them.
	for _, path := range []string{"/api/v1/me/anagrafica", "/api/v1/me/orders"} {
		if status := doRoutesRequest(t, app, http.MethodGet, path, clienteTokens.AccessToken); status == http.StatusForbidden {
			t.Fatalf("expected cliente to reach %s without 403, got %d", path, status)
		}
	}

	// Destination reads: BOTH admin and cliente must be able to read them
	// (writes stay staff-only, not exercised here).
	for _, tok := range []string{adminTokens.AccessToken, clienteTokens.AccessToken} {
		if status := doRoutesRequest(t, app, http.MethodGet, "/api/v1/destinations", tok); status == http.StatusForbidden {
			t.Fatalf("expected /destinations read to be reachable, got %d", status)
		}
	}

	// Client can create a new shared destination too.
	if status := doRoutesRequest(t, app, http.MethodPost, "/api/v1/me/destinations", clienteTokens.AccessToken); status == http.StatusForbidden {
		t.Fatalf("expected /me/destinations to be reachable by cliente, got %d", status)
	}

	// Geocode search: used by both staff anagrafica forms and the cliente
	// portal's own Nuova-Destinazione form.
	for _, tok := range []string{adminTokens.AccessToken, clienteTokens.AccessToken} {
		if status := doRoutesRequest(t, app, http.MethodGet, "/api/v1/geocode/search", tok); status == http.StatusForbidden {
			t.Fatalf("expected /geocode/search to be reachable, got %d", status)
		}
	}

	// GET /auth/me: reachable by any authenticated role.
	for _, tok := range []string{adminTokens.AccessToken, clienteTokens.AccessToken} {
		if status := doRoutesRequest(t, app, http.MethodGet, "/api/v1/auth/me", tok); status == http.StatusForbidden {
			t.Fatalf("expected /auth/me to be reachable by any role, got %d", status)
		}
	}
}

// TestSetupRoutes_ClientPortalRejectsMissingToken guards the "fails closed"
// side: without a token at all, both a staff route and a client-portal
// route must reject (401), never accidentally 200 due to a misplaced group.
func TestSetupRoutes_ClientPortalRejectsMissingToken(t *testing.T) {
	app, _ := newRoutesTestApp(t)

	if status := doRoutesRequest(t, app, http.MethodGet, "/api/v1/me/anagrafica", ""); status != http.StatusUnauthorized {
		t.Fatalf("expected 401 without a token, got %d", status)
	}
	if status := doRoutesRequest(t, app, http.MethodGet, "/api/v1/customers", ""); status != http.StatusUnauthorized {
		t.Fatalf("expected 401 without a token, got %d", status)
	}
}
