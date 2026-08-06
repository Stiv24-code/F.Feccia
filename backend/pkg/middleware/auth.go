package middleware

import (
	"strings"

	"fratelli-feccia/pkg/utils"

	"github.com/gofiber/fiber/v2"
)

// JWTAuthMiddleware validates access tokens and injects user context.
func JWTAuthMiddleware(cfg utils.JWTConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authz := c.Get("Authorization")
		if !strings.HasPrefix(strings.ToLower(authz), "bearer ") {
			return utils.ErrorResponse(c, 401, "Missing or invalid Authorization header")
		}
		token := strings.TrimSpace(authz[7:])
		claims, err := utils.ParseAccessToken(token, cfg)
		if err != nil {
			return utils.ErrorResponse(c, 401, "Invalid or expired token")
		}
		c.Locals("user_id", claims.UserID)
		c.Locals("role", claims.Role)
		c.Locals("customer_id", claims.CustomerID)
		return c.Next()
	}
}

// RequireRole allows access only to users having one of the given roles.
func RequireRole(roles ...string) fiber.Handler {
	allowed := make(map[string]struct{}, len(roles))
	for _, r := range roles {
		allowed[r] = struct{}{}
	}
	return func(c *fiber.Ctx) error {
		role, _ := c.Locals("role").(string)
		if _, ok := allowed[role]; !ok {
			return utils.ErrorResponse(c, fiber.StatusForbidden, "Forbidden")
		}
		return c.Next()
	}
}

// RequireAdmin allows access only to admin.
func RequireAdmin() fiber.Handler {
	return RequireRole(utils.RoleAdmin)
}

// PermitAllRoles allows access to any authenticated user, regardless of role.
func PermitAllRoles() fiber.Handler {
	return RequireRole(utils.RoleAdmin, utils.RoleAmministrazione, utils.RolePlanner, utils.RoleOperatore)
}
