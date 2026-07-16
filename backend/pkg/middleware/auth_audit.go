package middleware

import (
	"log/slog"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
)

const (
	authFailureWindow   = time.Minute
	authFailureThreshold = 30
)

type ipAuthFailures struct {
	mu   sync.Mutex
	byIP map[string]*authFailWindow
}

type authFailWindow struct {
	start time.Time
	count int
}

var globalAuthFailures ipAuthFailures

func init() {
	globalAuthFailures.byIP = make(map[string]*authFailWindow)
}

// AuthResponseAudit runs after handlers: tracks 401 per IP (401/403 details are logged in utils.ErrorResponse).
func AuthResponseAudit() fiber.Handler {
	return func(c *fiber.Ctx) error {
		err := c.Next()
		if c.Response().StatusCode() == fiber.StatusUnauthorized {
			globalAuthFailures.record401(c.IP())
		}
		return err
	}
}

func (a *ipAuthFailures) record401(ip string) {
	now := time.Now()
	a.mu.Lock()
	defer a.mu.Unlock()
	w := a.byIP[ip]
	if w == nil || now.Sub(w.start) > authFailureWindow {
		a.byIP[ip] = &authFailWindow{start: now, count: 1}
		return
	}
	w.count++
	if w.count == authFailureThreshold {
		slog.Error("possible credential stuffing or token abuse",
			"client_ip", ip,
			"failures_1m", w.count,
			"window", authFailureWindow,
		)
	}
}
