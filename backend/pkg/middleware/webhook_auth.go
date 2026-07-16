package middleware

import (
	"os"

	"fratelli-feccia/pkg/utils"

	"github.com/gofiber/fiber/v2"
)

// RequireWebhookToken mirrors backend's gps_webhook.py/_verify_signature and
// temperature.py/_verify_webhook_token: a static X-Webhook-Token header
// checked against the given env var. If the env var isn't set (dev), the
// check is skipped entirely — matching the Python behavior exactly.
func RequireWebhookToken(envVar string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		expected := os.Getenv(envVar)
		if expected == "" {
			return c.Next()
		}
		token := c.Get("X-Webhook-Token")
		if token == "" || token != expected {
			return utils.ErrorResponse(c, 401, "Webhook token non valido")
		}
		return c.Next()
	}
}
