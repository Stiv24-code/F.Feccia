package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	"fratelli-feccia/pkg/audit"
	"fratelli-feccia/pkg/utils"
)

// authExcludedPaths mirrors backend/audit.py's _AUTH_EXCLUDED_PATHS: these
// endpoints have explicit, more precise audit calls in the auth handler
// (distinguishing "wrong password" from "user disabled", which a raw status
// code can't), so the generic mutation logger skips them.
var authExcludedPaths = map[string]struct{}{
	"/api/v1/auth/login":   {},
	"/api/v1/auth/logout":  {},
	"/api/v1/auth/refresh": {},
}

// AuditHTTPMiddleware mirrors build_http_audit_middleware: logs every
// mutating (POST/PUT/PATCH/DELETE) /api/v1/* request. Registered globally,
// before JWT auth runs — it reads user_id/role from c.Locals *after*
// calling c.Next(), by which point JWTAuthMiddleware (if this route requires
// auth) has already populated them.
func AuditHTTPMiddleware(logger *audit.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		method := c.Method()
		path := c.Path()

		isMutation := method == fiber.MethodPost || method == fiber.MethodPut ||
			method == fiber.MethodPatch || method == fiber.MethodDelete
		isAPI := strings.HasPrefix(path, "/api/v1/")
		_, isAuthEndpoint := authExcludedPaths[path]

		if !isMutation || !isAPI || isAuthEndpoint {
			return c.Next()
		}

		userAgent := c.Get("User-Agent")
		ip := c.IP()

		err := c.Next()

		var userID *int64
		if uid, ok := c.Locals("user_id").(int64); ok {
			userID = &uid
		}
		role, _ := c.Locals("role").(string)

		resource, resourceID := audit.ClassifyPath(strings.TrimPrefix(path, "/api/v1"))
		status := c.Response().StatusCode()

		logger.Log(utils.RequestContext(c), audit.Entry{
			Action: method + " " + path, UserID: userID, UserRole: role,
			Resource: resource, ResourceID: resourceID, StatusCode: status,
			Success: status >= 200 && status < 400, IP: ip, UserAgent: userAgent,
		})

		return err
	}
}
