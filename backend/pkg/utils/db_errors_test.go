package utils

import (
	"encoding/json"
	"errors"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgconn"
)

// statusFor runs err through HandleDatabaseError on a throwaway app and
// returns what the caller would actually receive.
func statusFor(t *testing.T, err error) (int, string) {
	t.Helper()
	app := fiber.New()
	app.Get("/", func(c *fiber.Ctx) error { return HandleDatabaseError(c, err) })

	resp, rerr := app.Test(httptest.NewRequest("GET", "/", nil))
	if rerr != nil {
		t.Fatalf("app.Test failed: %v", rerr)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var payload struct {
		Error string `json:"error"`
	}
	if uerr := json.Unmarshal(body, &payload); uerr != nil {
		t.Fatalf("response is not the expected JSON: %s", body)
	}
	return resp.StatusCode, payload.Error
}

// A value longer than its column is the caller's mistake, not a server
// failure. It used to fall through to the generic 500 ("Please contact
// support"), which is what importing a real PDF order hit when the extracted
// load date came out as a window with a timezone.
func TestHandleDatabaseError_TooLongValueIsClientError(t *testing.T) {
	pgErr := &pgconn.PgError{
		Code:    "22001",
		Message: `value too long for type character varying(20)`,
	}

	status, msg := statusFor(t, pgErr)
	if status != 400 {
		t.Fatalf("expected 400 for 22001, got %d (%s)", status, msg)
	}
	// The declared size is the one detail that makes the error actionable.
	if !strings.Contains(msg, "20 characters") {
		t.Fatalf("expected the column size in the message, got %q", msg)
	}
}

// Drivers that do not surface a typed *pgconn.PgError (and GORM's wrapped
// errors) must land on the same answer, like every other code handled here.
func TestHandleDatabaseError_TooLongValueStringFallback(t *testing.T) {
	err := errors.New(`ERROR: value too long for type character varying(100) (SQLSTATE 22001)`)

	status, msg := statusFor(t, err)
	if status != 400 {
		t.Fatalf("expected 400 for the string fallback, got %d (%s)", status, msg)
	}
	if !strings.Contains(msg, "100 characters") {
		t.Fatalf("expected the column size in the message, got %q", msg)
	}
}

// An unrecognised failure must stay a 500: widening the 22001 mapping must
// not turn every database error into a client error.
func TestHandleDatabaseError_UnknownStaysServerError(t *testing.T) {
	status, _ := statusFor(t, errors.New("connection reset by peer"))
	if status != 500 {
		t.Fatalf("expected 500 for an unknown error, got %d", status)
	}
}

func TestParseTooLongError_WithoutSizeInMessage(t *testing.T) {
	msg := parseTooLongError("value too long")
	if msg == "" || strings.Contains(msg, "(") {
		t.Fatalf("expected a size-less fallback message, got %q", msg)
	}
}

// The offending value is caller-supplied content and must never be echoed
// back: it would travel into logs and UI alike.
func TestParseTooLongError_DoesNotEchoTheValue(t *testing.T) {
	msg := parseTooLongError(`value too long for type character varying(20): "22/07/26 15:00 - 23/07/26 04:00 CEST"`)
	if strings.Contains(msg, "22/07/26") {
		t.Fatalf("the offending value leaked into the message: %q", msg)
	}
}
