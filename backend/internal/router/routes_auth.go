package router

import (
	app_handlers "fratelli-feccia/internal/handlers"
	"fratelli-feccia/pkg/middleware"
	"fratelli-feccia/pkg/utils"
	"time"

	"github.com/gofiber/fiber/v2"
)

func registerAuthRoutes(api fiber.Router, handlers *app_handlers.Handler, jwtCfg utils.JWTConfig) {
	api.Post("/auth/login", middleware.NewLimiterIPPath(15, time.Minute), handlers.Auth.Login)
	api.Post("/auth/refresh", middleware.NewLimiterIPPath(15, time.Minute), handlers.Auth.Refresh)
	// Logout mirrors backend/routers/auth.py: best-effort, no auth required —
	// it just clears the refresh cookie even if the access token already expired.
	api.Post("/auth/logout", handlers.Auth.Logout)
}

// registerAuthMeRoute is called with the already-authenticated authAll group
// (see routes.go) since GET /auth/me requires a valid access token.
func registerAuthMeRoute(authAll fiber.Router, handlers *app_handlers.Handler) {
	authAll.Get("/auth/me", handlers.Auth.Me)
}

// registerAuthRegisterRoute mirrors POST /auth/register (admin-only in
// Python). Middleware is applied inline, not via a Group, per the Fiber
// Group("", ...) footgun documented in the roadmap plan.
func registerAuthRegisterRoute(authAll fiber.Router, handlers *app_handlers.Handler) {
	authAll.Post("/auth/register", middleware.RequireAdmin(), handlers.Auth.Register)
}
