package utils

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestRequestContext_NilCtxReturnsBackground(t *testing.T) {
	ctx := RequestContext(nil)
	if ctx == nil {
		t.Fatalf("expected non-nil context")
	}
	if ctx != context.Background() {
		t.Fatalf("expected context.Background, got different context")
	}
}

func TestRequestContext_ReturnsStoredContext(t *testing.T) {
	app := fiber.New()
	testCtx := context.Background()

	app.Get("/", func(c *fiber.Ctx) error {
		c.Locals("ctx", testCtx)
		ctx := RequestContext(c)
		if ctx != testCtx {
			t.Errorf("expected stored context, got different context")
		}
		return c.SendStatus(200)
	})

	req := httptest.NewRequest("GET", "/", nil)
	app.Test(req)
}
