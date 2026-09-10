package app

import (
	"encoding/json"
	"errors"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
)

// Stringa che imita un errore interno reale: nome utente del DB e un path.
// È esattamente il genere di dettaglio che l'ErrorHandler non deve far
// uscire dal server.
const leakyInternalError = `pq: password authentication failed for user "tms_app" (/var/lib/postgresql/data)`

// newTestApp replica il minimo di setupMiddleware che conta per
// l'ErrorHandler: recover (i panic diventano error) e requestid (il body del
// 5xx deve riportarlo).
func newTestApp() *fiber.App {
	app := newFiberApp()
	app.Use(recover.New())
	app.Use(requestid.New())
	return app
}

func doRequest(t *testing.T, app *fiber.App, path string) (int, map[string]any, string) {
	t.Helper()
	resp, err := app.Test(httptest.NewRequest("GET", path, nil))
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	var body map[string]any
	if err := json.Unmarshal(raw, &body); err != nil {
		t.Fatalf("body non JSON: %q", raw)
	}
	return resp.StatusCode, body, string(raw)
}

func TestErrorHandler_5xxDoesNotLeakInternalError(t *testing.T) {
	app := newTestApp()
	app.Get("/boom", func(c *fiber.Ctx) error {
		return errors.New(leakyInternalError)
	})

	status, body, raw := doRequest(t, app, "/boom")

	if status != fiber.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", status)
	}
	if strings.Contains(raw, "tms_app") || strings.Contains(raw, "/var/lib") || strings.Contains(raw, "pq:") {
		t.Fatalf("il body espone l'errore interno: %s", raw)
	}
	if body["error"] != genericServerError {
		t.Fatalf("error = %q, want il testo generico", body["error"])
	}
	// Il request_id è l'unico appiglio dell'utente verso l'assistenza: deve
	// esserci e non essere vuoto.
	if rid, _ := body["request_id"].(string); rid == "" {
		t.Fatalf("request_id mancante nel body: %s", raw)
	}
}

func TestErrorHandler_PanicIsRecoveredAndNotLeaked(t *testing.T) {
	app := newTestApp()
	app.Get("/panic", func(c *fiber.Ctx) error {
		panic(leakyInternalError)
	})

	status, body, raw := doRequest(t, app, "/panic")

	if status != fiber.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", status)
	}
	if strings.Contains(raw, "tms_app") || strings.Contains(raw, "/var/lib") {
		t.Fatalf("il body espone il messaggio del panic: %s", raw)
	}
	if body["error"] != genericServerError {
		t.Fatalf("error = %q, want il testo generico", body["error"])
	}
}

func TestErrorHandler_4xxKeepsFrameworkMessage(t *testing.T) {
	app := newTestApp()
	app.Get("/too-big", func(c *fiber.Ctx) error {
		return fiber.NewError(fiber.StatusRequestEntityTooLarge, "Request Entity Too Large")
	})

	// 404 di route inesistente: il messaggio è di Fiber, controllato e utile.
	status, body, _ := doRequest(t, app, "/questa-route-non-esiste")
	if status != fiber.StatusNotFound {
		t.Fatalf("status = %d, want 404", status)
	}
	if msg, _ := body["error"].(string); !strings.Contains(msg, "Cannot GET") {
		t.Fatalf("messaggio 404 = %q, atteso quello di Fiber", msg)
	}
	if _, has := body["request_id"]; has {
		t.Fatalf("un 4xx non deve portare request_id nel body: %v", body)
	}

	// *fiber.Error 4xx esplicito: stesso trattamento.
	status, body, _ = doRequest(t, app, "/too-big")
	if status != fiber.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want 413", status)
	}
	if body["error"] != "Request Entity Too Large" {
		t.Fatalf("error = %q, want il messaggio del *fiber.Error", body["error"])
	}
}

func TestErrorHandler_FiberError5xxIsAlsoGeneric(t *testing.T) {
	// Un *fiber.Error con codice 5xx (es. fiber.ErrBadGateway con dettagli
	// interni nel messaggio) va trattato come qualunque altro 5xx: il codice
	// resta, il messaggio no.
	app := newTestApp()
	app.Get("/upstream", func(c *fiber.Ctx) error {
		return fiber.NewError(fiber.StatusBadGateway, "dial tcp 10.0.3.7:5432: connect: connection refused")
	})

	status, body, raw := doRequest(t, app, "/upstream")
	if status != fiber.StatusBadGateway {
		t.Fatalf("status = %d, want 502 (il codice va conservato)", status)
	}
	if strings.Contains(raw, "10.0.3.7") {
		t.Fatalf("il body espone l'indirizzo interno: %s", raw)
	}
	if body["error"] != genericServerError {
		t.Fatalf("error = %q, want il testo generico", body["error"])
	}
}
