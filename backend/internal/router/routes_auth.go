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
	// Public self-service signup for the "cliente" role — deliberately a
	// different path from the admin-only /auth/register below, tighter rate
	// limit (anonymous + writes a Customer row, more abuse-prone than login).
	api.Post("/auth/register-cliente", middleware.NewLimiterIPPath(5, time.Minute), handlers.Auth.RegisterClient)
}

// registerAuthMeRoute mirrors GET /auth/me — valid access token required,
// but no role restriction (any of the 4 staff roles or "cliente" can call
// it). Middleware is applied inline on the ungrouped `api` router, not via a
// second Group(""), per the Fiber Group("", ...) footgun documented in the
// roadmap plan — a Group("", JWTAuthMiddleware) here would leak that
// middleware onto every route registered afterwards on `api`, including the
// unauthenticated /auth/login etc. if ever reordered, and (worse) would
// stack with authAll's own Group("", ...) for every already-registered route.
func registerAuthMeRoute(api fiber.Router, handlers *app_handlers.Handler, jwtCfg utils.JWTConfig) {
	api.Get("/auth/me", middleware.JWTAuthMiddleware(jwtCfg), handlers.Auth.Me)
}

// registerAuthRegisterRoute mirrors POST /auth/register (admin-only in
// Python). Middleware is applied inline, not via a Group, per the Fiber
// Group("", ...) footgun documented in the roadmap plan.
func registerAuthRegisterRoute(authAll fiber.Router, handlers *app_handlers.Handler) {
	authAll.Post("/auth/register", middleware.RequireAdmin(), handlers.Auth.Register)
}
