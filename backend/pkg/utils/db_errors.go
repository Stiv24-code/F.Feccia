package utils

import (
	"context"
	"errors"
	"log/slog"
	"regexp"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

// parseForeignKeyError returns a user-friendly message without disclosing DB internals.
func parseForeignKeyError(msg string) string {
	pattern := `(?:update|delete|insert)\s+(?:or\s+(?:update|delete))?\s+on\s+table\s+"([^"]+)"\s+violates\s+foreign\s+key\s+constraint\s+"([^"]+)"(?:\s+on\s+table\s+"([^"]+)")?`
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(msg)

	if len(matches) >= 3 {
		return "Cannot complete the operation: the record is referenced by other data"
	}

	tablePattern := `table\s+"([^"]+)"`
	tableRe := regexp.MustCompile(tablePattern)
	tableMatches := tableRe.FindAllStringSubmatch(msg, -1)
	if len(tableMatches) > 0 {
		return "Cannot complete the operation: the record is referenced by other data"
	}

	return "Cannot complete the operation: the record is referenced by other data"
}

// parseTooLongError turns Postgres' 22001 into something an operator can act
// on. Sending a value longer than its column is a client error, not a server
// failure: before this mapping existed it fell through to the generic 500
// ("Please contact support"), which is what an import of a real PDF order hit
// when the extracted load date came out as a window with a timezone —
// 36 characters into what was then a varchar(20).
//
// The message keeps the declared size, which is the one detail that makes the
// error self-explanatory, but never the offending value: it is caller-supplied
// content and would end up echoed in logs and UI alike.
func parseTooLongError(msg string) string {
	re := regexp.MustCompile(`character varying\((\d+)\)`)
	if m := re.FindStringSubmatch(msg); len(m) == 2 {
		return "A value exceeds the maximum length allowed for its field (" + m[1] + " characters)"
	}
	return "A value exceeds the maximum length allowed for its field"
}

// HandleDatabaseError maps database errors to appropriate HTTP responses.
func HandleDatabaseError(c *fiber.Ctx, err error) error {
	if err == nil {
		return nil
	}

	// 0. Service-level API errors (e.g. validation/domain rules)
	type statusCoder interface {
		StatusCode() int
	}
	if sc, ok := err.(statusCoder); ok {
		return ErrorResponse(c, sc.StatusCode(), err.Error())
	}

	// 1. Context timeout
	if errors.Is(err, context.DeadlineExceeded) {
		return ErrorResponse(c, 504, "Database request timed out")
	}

	// 2. Record not found
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrorResponse(c, 404, "Record not found")
	}

	// 3. PostgreSQL specific errors using pgconn
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505": // unique_violation
			return ErrorResponse(c, 409, "A record with these values already exists")
		case "23503": // foreign_key_violation
			msg := parseForeignKeyError(pgErr.Message)
			return ErrorResponse(c, 400, msg)
		case "23502": // not_null_violation
			return ErrorResponse(c, 400, "A required field cannot be empty")
		case "23514": // check_violation
			return ErrorResponse(c, 400, "Data does not satisfy the required constraints")
		case "22001": // string_data_right_truncation
			return ErrorResponse(c, 400, parseTooLongError(pgErr.Message))
		}
	}

	// 4. Fallback for other drivers (like SQLite) or string-based matching
	errStr := err.Error()
	if strings.Contains(errStr, "SQLSTATE 23505") ||
		strings.Contains(errStr, "duplicate key value violates unique constraint") ||
		strings.Contains(errStr, "UNIQUE constraint failed") {
		return ErrorResponse(c, 409, "A record with these values already exists")
	}

	if strings.Contains(errStr, "SQLSTATE 23503") ||
		strings.Contains(errStr, "violates foreign key constraint") {
		msg := parseForeignKeyError(errStr)
		return ErrorResponse(c, 400, msg)
	}

	if strings.Contains(errStr, "SQLSTATE 23502") ||
		strings.Contains(errStr, "violates not-null constraint") {
		return ErrorResponse(c, 400, "A required field cannot be empty")
	}

	if strings.Contains(errStr, "SQLSTATE 23514") ||
		strings.Contains(errStr, "violates check constraint") {
		return ErrorResponse(c, 400, "Data does not satisfy the required constraints")
	}

	if strings.Contains(errStr, "SQLSTATE 22001") ||
		strings.Contains(errStr, "value too long for type") {
		return ErrorResponse(c, 400, parseTooLongError(errStr))
	}

	// 5. Generic database error
	slog.Error(
		"database operation failed",
		"error", errStr,
		"request_id", c.Locals("requestid"),
	)
	return ErrorResponse(c, 500, "Database operation failed. Please contact support.")
}
