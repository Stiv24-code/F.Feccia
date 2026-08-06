package utils

import (
	"context"
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
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

// RequestCustomerID reads the "customer_id" JWT claim JWTAuthMiddleware
// stores in Locals — only ever non-empty for RoleCliente. Client-portal
// handlers (routes_client_portal.go) use this instead of trusting any
// customer/cliente id supplied in the request body or path, so a client can
// never read or write another customer's data.
func RequestCustomerID(c *fiber.Ctx) (uuid.UUID, error) {
	raw, _ := c.Locals("customer_id").(string)
	if raw == "" {
		return uuid.UUID{}, errors.New("missing customer_id claim")
	}
	return uuid.Parse(raw)
}
