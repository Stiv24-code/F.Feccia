package middleware

import (
	"fratelli-feccia/pkg/utils"

	"github.com/gofiber/fiber/v2"
)

// LimitBodySize limits the maximum size of the request body for specific routes.
func LimitBodySize(maxBytes int) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if len(c.Body()) > maxBytes {
			return utils.ErrorResponse(c, fiber.StatusRequestEntityTooLarge, "Request body too large")
		}
		return c.Next()
	}
}
