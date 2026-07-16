package middleware

import (
	"time"

	"fratelli-feccia/pkg/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

// NewLimiterIP creates a limiter keyed by client IP
func NewLimiterIP(max int, window time.Duration) fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        max,
		Expiration: window,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return utils.ErrorResponse(c, fiber.StatusTooManyRequests, "Too many requests")
		},
	})
}

// NewLimiterIPPath creates a limiter keyed by client IP and request path
func NewLimiterIPPath(max int, window time.Duration) fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        max,
		Expiration: window,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP() + "|" + c.Path()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return utils.ErrorResponse(c, fiber.StatusTooManyRequests, "Too many requests")
		},
	})
}
