package utils

import (
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

// RefreshCookieName mirrors backend/services.py's REFRESH_COOKIE_NAME
// ("tms_refresh") — the frontend's auth-context relies on this exact
// cookie to silently restore a session on page load via
// POST /api/v1/auth/refresh.
//
// RefreshCookiePath must match the mount path of the auth routes
// (registerAuthRoutes on the "/api/v1" group, see internal/router/routes.go),
// otherwise the browser never sends the cookie back on refresh — it was
// left as the old Python path ("/api/auth") after the Go port moved the
// routes under "/api/v1/auth", causing every refresh to fail with
// "Refresh token mancante".
const (
	RefreshCookieName = "tms_refresh"
	RefreshCookiePath = "/api/v1/auth"
)

// SetRefreshCookie sets the httpOnly refresh cookie, mirroring
// backend/services.py's set_refresh_cookie.
func SetRefreshCookie(c *fiber.Ctx, token string, ttl time.Duration) {
	c.Cookie(&fiber.Cookie{
		Name:     RefreshCookieName,
		Value:    token,
		Path:     RefreshCookiePath,
		MaxAge:   int(ttl.Seconds()),
		HTTPOnly: true,
		SameSite: fiber.CookieSameSiteLaxMode,
		Secure:   strings.EqualFold(os.Getenv("SECURE_COOKIES"), "true"),
	})
}

// ClearRefreshCookie mirrors backend/services.py's clear_refresh_cookie.
func ClearRefreshCookie(c *fiber.Ctx) {
	c.Cookie(&fiber.Cookie{
		Name:     RefreshCookieName,
		Value:    "",
		Path:     RefreshCookiePath,
		MaxAge:   -1,
		HTTPOnly: true,
		SameSite: fiber.CookieSameSiteLaxMode,
		Secure:   strings.EqualFold(os.Getenv("SECURE_COOKIES"), "true"),
	})
}
