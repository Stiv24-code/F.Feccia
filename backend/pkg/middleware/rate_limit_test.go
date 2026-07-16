package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
)

func TestNewLimiterIP_LimitsByIP(t *testing.T) {
	app := fiber.New()
	app.Use(NewLimiterIP(1, time.Minute))
	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"ok": true})
	})

	// First request from IP 1.2.3.4 should pass.
	req1 := httptest.NewRequest(http.MethodGet, "/", nil)
	req1.RemoteAddr = "1.2.3.4:1234"
	resp1, err := app.Test(req1, -1)
	if err != nil {
		t.Fatalf("app.Test error: %v", err)
	}
	if resp1.StatusCode != http.StatusOK {
		t.Fatalf("expected first request status %d, got %d", http.StatusOK, resp1.StatusCode)
	}

	// Second request from same IP should be limited.
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.RemoteAddr = "1.2.3.4:5678"
	resp2, err := app.Test(req2, -1)
	if err != nil {
		t.Fatalf("app.Test error: %v", err)
	}
	if resp2.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("expected second request status %d, got %d", http.StatusTooManyRequests, resp2.StatusCode)
	}
}

func TestNewLimiterIPPath_LimitsPerPath(t *testing.T) {
	app := fiber.New()
	app.Use(NewLimiterIPPath(1, time.Minute))
	app.Get("/a", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"ok": "a"})
	})
	app.Get("/b", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"ok": "b"})
	})

	// First request to /a from IP should pass.
	req1 := httptest.NewRequest(http.MethodGet, "/a", nil)
	req1.RemoteAddr = "1.2.3.4:1234"
	resp1, err := app.Test(req1, -1)
	if err != nil {
		t.Fatalf("app.Test error: %v", err)
	}
	if resp1.StatusCode != http.StatusOK {
		t.Fatalf("expected first /a status %d, got %d", http.StatusOK, resp1.StatusCode)
	}

	// Second request to /a from same IP should be limited.
	req2 := httptest.NewRequest(http.MethodGet, "/a", nil)
	req2.RemoteAddr = "1.2.3.4:5678"
	resp2, err := app.Test(req2, -1)
	if err != nil {
		t.Fatalf("app.Test error: %v", err)
	}
	if resp2.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("expected second /a status %d, got %d", http.StatusTooManyRequests, resp2.StatusCode)
	}

	// Request to /b from same IP should still pass (separate key).
	req3 := httptest.NewRequest(http.MethodGet, "/b", nil)
	req3.RemoteAddr = "1.2.3.4:9999"
	resp3, err := app.Test(req3, -1)
	if err != nil {
		t.Fatalf("app.Test error: %v", err)
	}
	if resp3.StatusCode != http.StatusOK {
		t.Fatalf("expected /b status %d, got %d", http.StatusOK, resp3.StatusCode)
	}
}
