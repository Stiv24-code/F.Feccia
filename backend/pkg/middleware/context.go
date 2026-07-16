package middleware

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
)

// ContextMiddleware creates a context with timeout for the request
func ContextMiddleware(timeout time.Duration) fiber.Handler {
	return func(c *fiber.Ctx) error {
		parent := c.UserContext()
		if parent == nil {
			parent = context.Background()
		}

		ctx := parent
		cancel := func() {}
		if timeout > 0 {
			ctx, cancel = context.WithTimeout(parent, timeout)
			defer cancel()
		}

		c.Locals("ctx", ctx)

		return c.Next()
	}
}
