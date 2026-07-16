package utils

import (
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel/trace"
)

// ListResult is the standard paginated list response.
type ListResult struct {
	Data  interface{} `json:"data"`
	Total int64       `json:"total"`
}

// ErrorResponse returns an HTTP JSON error response
func ErrorResponse(c *fiber.Ctx, status int, msg string) error {
	c.Locals("app_error", msg)

	// Controlled errors (we return JSON and no Go error), so they won't appear in Fiber's ErrorHandler.
	// Log them here for observability; treat 5xx as errors, 4xx as warnings.
	traceID := ""
	if sc := trace.SpanContextFromContext(c.UserContext()); sc.IsValid() {
		traceID = sc.TraceID().String()
	}
	attrs := []any{
		"status", status,
		"method", c.Method(),
		"path", c.Path(),
		"request_id", c.Locals("requestid"),
		"trace_id", traceID,
		"error", msg,
	}
	if status == 401 || status == 403 {
		attrs = append(attrs, "client_ip", c.IP(), "user_agent", c.Get("User-Agent"))
	}
	if status >= 500 {
		slog.Error("request failed", attrs...)
	} else if status >= 400 {
		slog.Warn("request rejected", attrs...)
	}

	return c.Status(status).JSON(fiber.Map{"error": msg})
}

// SuccessResponse returns an HTTP JSON success response
func SuccessResponse(c *fiber.Ctx, status int, data interface{}) error {
	return c.Status(status).JSON(data)
}

// ValidationErrorResponse returns a JSON response with validation errors
func ValidationErrorResponse(c *fiber.Ctx, validationErrors []ValidationError) error {
	c.Locals("app_error", "Validation failed")
	return c.Status(400).JSON(fiber.Map{
		"error":   "Validation failed",
		"details": validationErrors,
	})
}
