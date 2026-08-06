package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"fratelli-feccia/pkg/utils"

	"github.com/gofiber/fiber/v2"
)

func newTestJWTConfig() utils.JWTConfig {
	return utils.JWTConfig{
		AccessSecret:  "access-secret",
		RefreshSecret: "refresh-secret",
		AccessTTL:     time.Minute,
		RefreshTTL:    time.Hour,
	}
}

func newJWTTestApp(cfg utils.JWTConfig, handlers ...fiber.Handler) *fiber.App {
	app := fiber.New()
	routeHandlers := append([]fiber.Handler{JWTAuthMiddleware(cfg)}, handlers...)
	app.Get("/", routeHandlers...)
	return app
}

func doJWTRequest(t *testing.T, app *fiber.App, authHeader string) (int, map[string]any) {
	t.Helper()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test error: %v", err)
	}

	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	return resp.StatusCode, body
}

func TestJWTAuthMiddleware_MissingAuthorizationHeader(t *testing.T) {
	cfg := newTestJWTConfig()
	app := newJWTTestApp(cfg)

	status, body := doJWTRequest(t, app, "")

	if status != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, status)
	}
	if body["error"] != "Missing or invalid Authorization header" {
		t.Fatalf("expected error message %q, got %v", "Missing or invalid Authorization header", body["error"])
	}
}

func TestJWTAuthMiddleware_InvalidToken(t *testing.T) {
	cfg := newTestJWTConfig()
	app := newJWTTestApp(cfg)

	status, body := doJWTRequest(t, app, "Bearer invalid-token")

	if status != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, status)
	}
	if body["error"] != "Invalid or expired token" {
		t.Fatalf("expected error message %q, got %v", "Invalid or expired token", body["error"])
	}
}

func TestJWTAuthMiddleware_ValidTokenSetsLocals(t *testing.T) {
	cfg := newTestJWTConfig()
	app := newJWTTestApp(cfg, func(c *fiber.Ctx) error {
		userID, okID := c.Locals("user_id").(int64)
		role, okRole := c.Locals("role").(string)
		if !okID || !okRole {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "missing locals"})
		}
		return c.JSON(fiber.Map{"user_id": userID, "role": role})
	})

	const wantUserID int64 = 42
	const wantRole = "admin"

	pair, err := utils.GenerateTokenPair(wantUserID, wantRole, "", cfg)
	if err != nil {
		t.Fatalf("GenerateTokenPair error: %v", err)
	}

	status, body := doJWTRequest(t, app, "Bearer "+pair.AccessToken)

	if status != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, status)
	}
	if int64(body["user_id"].(float64)) != wantUserID {
		t.Fatalf("expected user_id %d, got %v", wantUserID, body["user_id"])
	}
	if body["role"] != wantRole {
		t.Fatalf("expected role %q, got %v", wantRole, body["role"])
	}
}

func TestJWTAuthMiddleware_SetsCustomerIDLocalForCliente(t *testing.T) {
	cfg := newTestJWTConfig()
	app := newJWTTestApp(cfg, func(c *fiber.Ctx) error {
		customerID, _ := c.Locals("customer_id").(string)
		return c.JSON(fiber.Map{"customer_id": customerID})
	})

	const wantCustomerID = "11111111-1111-1111-1111-111111111111"
	pair, err := utils.GenerateTokenPair(7, utils.RoleCliente, wantCustomerID, cfg)
	if err != nil {
		t.Fatalf("GenerateTokenPair error: %v", err)
	}

	status, body := doJWTRequest(t, app, "Bearer "+pair.AccessToken)

	if status != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, status)
	}
	if body["customer_id"] != wantCustomerID {
		t.Fatalf("expected customer_id %q, got %v", wantCustomerID, body["customer_id"])
	}
}

func TestRequireRole_Forbidden(t *testing.T) {
	app := fiber.New()
	app.Get("/", func(c *fiber.Ctx) error {
		c.Locals("role", "unknown")
		return c.Next()
	}, RequireRole("admin"), func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test error: %v", err)
	}
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, resp.StatusCode)
	}
}

func TestRequireRole_Allowed(t *testing.T) {
	app := fiber.New()
	app.Get("/", func(c *fiber.Ctx) error {
		c.Locals("role", "admin")
		return c.Next()
	}, RequireRole("admin", "planner"), func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}
}

func TestRequireAdmin_Shorthand(t *testing.T) {
	app := fiber.New()
	app.Get("/", func(c *fiber.Ctx) error {
		c.Locals("role", "admin")
		return c.Next()
	}, RequireAdmin(), func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}
}

func TestPermitAllRoles(t *testing.T) {
	tests := []struct {
		name       string
		role       string
		wantStatus int
	}{
		{"allowed for admin", utils.RoleAdmin, fiber.StatusOK},
		{"allowed for amministrazione", utils.RoleAmministrazione, fiber.StatusOK},
		{"allowed for planner", utils.RolePlanner, fiber.StatusOK},
		{"allowed for operatore", utils.RoleOperatore, fiber.StatusOK},
		// cliente is deliberately excluded — a client account must never
		// get blanket access to the staff-only route group (it only
		// reaches routes_client_portal.go's dedicated group).
		{"forbidden for cliente", utils.RoleCliente, fiber.StatusForbidden},
		{"forbidden for empty role", "", fiber.StatusForbidden},
		{"forbidden for unknown role", "user", fiber.StatusForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()
			app.Get("/api/v1/test", func(c *fiber.Ctx) error {
				if tt.role != "" {
					c.Locals("role", tt.role)
				}
				return c.Next()
			}, PermitAllRoles(), func(c *fiber.Ctx) error {
				return c.SendStatus(fiber.StatusOK)
			})

			req := httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)
			resp, err := app.Test(req, -1)
			if err != nil {
				t.Fatalf("app.Test error: %v", err)
			}
			if resp.StatusCode != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, resp.StatusCode)
			}
		})
	}
}
