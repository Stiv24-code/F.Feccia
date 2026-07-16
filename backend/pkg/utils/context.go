package utils

import (
	"context"

	"github.com/gofiber/fiber/v2"
)

// RequestContext retrieves the context from Fiber context locals.
// Falls back to UserContext() if no custom context is set.
func RequestContext(c *fiber.Ctx) context.Context {
	if c == nil {
		return context.Background()
	}
	if ctx, ok := c.Locals("ctx").(context.Context); ok {
		return ctx
	}
	return c.UserContext()
}
