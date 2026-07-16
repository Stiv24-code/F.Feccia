package utils

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

// RetryConfig contains settings for retry attempts
type RetryConfig struct {
	MaxAttempts  int           // Maximum number of attempts
	InitialDelay time.Duration // Initial delay before first retry attempt
	MaxDelay     time.Duration // Maximum delay between attempts
	Multiplier   float64       // Multiplier for exponential backoff
}

// DefaultRetryConfig returns default configuration for database operations
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     5 * time.Second,
		Multiplier:   2.0,
	}
}

// DatabaseConnectionRetryConfig returns configuration for database connection
// Uses more attempts and longer delays
func DatabaseConnectionRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts:  10,
		InitialDelay: 1 * time.Second,
		MaxDelay:     30 * time.Second,
		Multiplier:   2.0,
	}
}

// IsRetryableError checks if an operation can be retried for the given error
func IsRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// PostgreSQL errors via pgconn
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		// Deadlock detection (serialization failure)
		if pgErr.Code == "40001" {
			return true
		}
		// Lock not available (could not obtain lock)
		if pgErr.Code == "55P03" {
			return true
		}
		// Connection errors
		// 08003 - connection does not exist
		// 08006 - connection failure
		// 08001 - SQL client unable to establish SQL connection
		// 08000 - connection exception
		if pgErr.Code == "08003" || pgErr.Code == "08006" || pgErr.Code == "08001" || pgErr.Code == "08000" {
			return true
		}
		// Serialization failures (transaction rollback)
		if pgErr.Code == "40001" || pgErr.Code == "40P01" {
			return true
		}
		// Admin shutdown
		if pgErr.Code == "57P01" {
			return true
		}
		// Crash shutdown
		if pgErr.Code == "57P02" {
			return true
		}
		// Cannot connect now
		if pgErr.Code == "57P03" {
			return true
		}
	}

	// GORM errors - don't retry for "not found"
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false
	}

	// Temporary network errors
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}

	// Check string-based errors for network issues
	errStr := strings.ToLower(err.Error())
	retryableStrings := []string{
		"connection refused",
		"connection reset",
		"connection closed",
		"no connection",
		"network is unreachable",
		"timeout",
		"temporary failure",
		"try again",
		"too many connections",
		"server closed the connection",
		"broken pipe",
		"i/o timeout",
	}

	for _, retryableStr := range retryableStrings {
		if strings.Contains(errStr, retryableStr) {
			return true
		}
	}

	return false
}

// Retry executes a function with retry attempts
// Uses exponential backoff between attempts
func Retry(ctx context.Context, fn func() error, config RetryConfig) error {
	var lastErr error
	delay := config.InitialDelay

	for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
		err := fn()
		if err == nil {
			if attempt > 1 {
				slog.Info("Operation succeeded after retry",
					"attempt", attempt,
					"total_attempts", config.MaxAttempts)
			}
			return nil
		}

		lastErr = err

		// Check if we can retry
		if !IsRetryableError(err) {
			slog.Debug("Error is not retryable, aborting retries",
				"attempt", attempt,
				"error", err.Error())
			return err
		}

		// If this is the last attempt, return error
		if attempt == config.MaxAttempts {
			slog.Warn("Retry exhausted, all attempts failed",
				"attempts", attempt,
				"max_attempts", config.MaxAttempts,
				"error", err.Error())
			return err
		}

		// Log retry attempt
		slog.Info("Retrying operation",
			"attempt", attempt,
			"max_attempts", config.MaxAttempts,
			"next_delay", delay,
			"error", err.Error())

		// Wait before next attempt
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}

		// Increase delay exponentially
		delay = time.Duration(float64(delay) * config.Multiplier)
		if delay > config.MaxDelay {
			delay = config.MaxDelay
		}
	}

	return lastErr
}

// RetryWithResult executes a function with retry attempts and returns the result
// Used for operations that return a value
func RetryWithResult[T any](ctx context.Context, fn func() (T, error), config RetryConfig) (T, error) {
	var lastErr error
	var zero T
	delay := config.InitialDelay

	for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
		result, err := fn()
		if err == nil {
			if attempt > 1 {
				slog.Info("Operation succeeded after retry",
					"attempt", attempt,
					"total_attempts", config.MaxAttempts)
			}
			return result, nil
		}

		lastErr = err

		// Check if we can retry
		if !IsRetryableError(err) {
			slog.Debug("Error is not retryable, aborting retries",
				"attempt", attempt,
				"error", err.Error())
			return zero, err
		}

		// If this is the last attempt, return error
		if attempt == config.MaxAttempts {
			slog.Warn("Retry exhausted, all attempts failed",
				"attempts", attempt,
				"max_attempts", config.MaxAttempts,
				"error", err.Error())
			return zero, err
		}

		// Log retry attempt
		slog.Info("Retrying operation",
			"attempt", attempt,
			"max_attempts", config.MaxAttempts,
			"next_delay", delay,
			"error", err.Error())

		// Wait before next attempt
		select {
		case <-ctx.Done():
			return zero, ctx.Err()
		case <-time.After(delay):
		}

		// Increase delay exponentially
		delay = time.Duration(float64(delay) * config.Multiplier)
		if delay > config.MaxDelay {
			delay = config.MaxDelay
		}
	}

	return zero, lastErr
}
