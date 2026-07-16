package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"fratelli-feccia/pkg/utils"

	"github.com/gofiber/fiber/v2"
)

func TestContextMiddleware_SetsContextWithTimeout(t *testing.T) {
	app := fiber.New()

	app.Use(ContextMiddleware(10 * time.Second))
	app.Get("/", func(c *fiber.Ctx) error {
		ctxVal := c.Locals("ctx")
		if ctxVal == nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "missing ctx local"})
		}
		ctx, ok := ctxVal.(context.Context)
		if !ok {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "ctx local is not context.Context"})
		}

		if deadline, ok := ctx.Deadline(); !ok || deadline.IsZero() {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "deadline not set"})
		}

		// Also verify RequestContext reuses this context.
		if got := utils.RequestContext(c); got != ctx {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "RequestContext did not return stored ctx"})
		}

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
