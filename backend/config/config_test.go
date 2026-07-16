package config

import (
	"os"
	"os/exec"
	"testing"
)

func TestGetEnv_ReturnsEnvValue(t *testing.T) {
	// Uses process env; keep sequential to avoid interference.
	t.Setenv("TEST_KEY", "value")

	got := getEnv("TEST_KEY", "default")
	if got != "value" {
		t.Fatalf("expected %q, got %q", "value", got)
	}
}

func TestGetEnv_ReturnsDefaultWhenUnset(t *testing.T) {
	// Uses process env; keep sequential to avoid interference.
	const key = "TEST_KEY_UNSET"
	_ = os.Unsetenv(key)

	got := getEnv(key, "default")
	if got != "default" {
		t.Fatalf("expected default %q, got %q", "default", got)
	}
}

func TestLoad_UsesDefaultsWhenEnvMissing(t *testing.T) {
	// Manipulates many env vars; avoid running in parallel with other env tests.
	// Ensure only required variables are set to avoid os.Exit(1).
	t.Setenv("DB_PASSWORD", "secret")
	t.Setenv("JWT_ACCESS_SECRET", "access_secret")
	t.Setenv("JWT_REFRESH_SECRET", "refresh_secret")

	// Unset others to rely on defaults.
	_ = os.Unsetenv("DB_HOST")
	_ = os.Unsetenv("DB_PORT")
	_ = os.Unsetenv("DB_USER")
	_ = os.Unsetenv("DB_NAME")
	_ = os.Unsetenv("DB_SSLMODE")
	_ = os.Unsetenv("SERVER_PORT")
	_ = os.Unsetenv("SERVER_HOST")

	cfg := Load()

	if cfg.Database.Host != "localhost" {
		t.Fatalf("expected DB host %q, got %q", "localhost", cfg.Database.Host)
	}
	if cfg.Database.Port != "5432" {
		t.Fatalf("expected DB port %q, got %q", "5432", cfg.Database.Port)
	}
	if cfg.Database.User != "postgres" {
		t.Fatalf("expected DB user %q, got %q", "postgres", cfg.Database.User)
	}
	if cfg.Database.DBName != "appdb" {
		t.Fatalf("expected DB name %q, got %q", "appdb", cfg.Database.DBName)
	}
	if cfg.Database.SSLMode != "disable" {
		t.Fatalf("expected DB sslmode %q, got %q", "disable", cfg.Database.SSLMode)
	}

	if cfg.Server.Port != "8080" {
		t.Fatalf("expected server port %q, got %q", "8080", cfg.Server.Port)
	}
	if cfg.Server.Host != "0.0.0.0" {
		t.Fatalf("expected server host %q, got %q", "0.0.0.0", cfg.Server.Host)
	}
}

func TestLoad_UsesEnvironmentOverrides(t *testing.T) {
	// Manipulates many env vars; avoid running in parallel with other env tests.
	t.Setenv("DB_HOST", "db.example.com")
	t.Setenv("DB_PORT", "15432")
	t.Setenv("DB_USER", "user1")
	t.Setenv("DB_PASSWORD", "secret")
	t.Setenv("JWT_ACCESS_SECRET", "access_secret_override")
	t.Setenv("JWT_REFRESH_SECRET", "refresh_secret_override")
	t.Setenv("DB_NAME", "customdb")
	t.Setenv("DB_SSLMODE", "require")
	t.Setenv("SERVER_PORT", "9000")
	t.Setenv("SERVER_HOST", "127.0.0.1")

	cfg := Load()

	if cfg.Database.Host != "db.example.com" {
		t.Fatalf("expected DB host %q, got %q", "db.example.com", cfg.Database.Host)
	}
	if cfg.Database.Port != "15432" {
		t.Fatalf("expected DB port %q, got %q", "15432", cfg.Database.Port)
	}
	if cfg.Database.User != "user1" {
		t.Fatalf("expected DB user %q, got %q", "user1", cfg.Database.User)
	}
	if cfg.Database.Password != "secret" {
		t.Fatalf("expected DB password %q, got %q", "secret", cfg.Database.Password)
	}
	if cfg.Security.JWTAccessSecret != "access_secret_override" {
		t.Fatalf("expected JWT access secret %q, got %q", "access_secret_override", cfg.Security.JWTAccessSecret)
	}
	if cfg.Security.JWTRefreshSecret != "refresh_secret_override" {
		t.Fatalf("expected JWT refresh secret %q, got %q", "refresh_secret_override", cfg.Security.JWTRefreshSecret)
	}
	if cfg.Database.DBName != "customdb" {
		t.Fatalf("expected DB name %q, got %q", "customdb", cfg.Database.DBName)
	}
	if cfg.Database.SSLMode != "require" {
		t.Fatalf("expected DB sslmode %q, got %q", "require", cfg.Database.SSLMode)
	}
	if cfg.Server.Port != "9000" {
		t.Fatalf("expected server port %q, got %q", "9000", cfg.Server.Port)
	}
	if cfg.Server.Host != "127.0.0.1" {
		t.Fatalf("expected server host %q, got %q", "127.0.0.1", cfg.Server.Host)
	}
}

// Test that validateConfig/os.Exit(1) is called when required envs (DB_PASSWORD) are missing.
// We use a subprocess so that os.Exit(1) doesn't stop the whole test suite.
func TestValidateConfig_MissingPassword_Exits(t *testing.T) {
	const envKey = "TEST_VALIDATECONFIG_MISSINGPWD"

	if os.Getenv(envKey) == "1" {
		// Child process: run the code that should call os.Exit(1).
		cfg := &Config{
			Database: DatabaseConfig{
				Password: "",
			},
		}
		validateConfig(cfg)
		// If we reach this line, os.Exit wasn't called.
		t.Fatalf("expected os.Exit(1) to be called for missing DB_PASSWORD")
	}

	// Parent process: run tests in a subprocess.
	cmd := exec.Command(os.Args[0], "-test.run=TestValidateConfig_MissingPassword_Exits")
	cmd.Env = append(os.Environ(), envKey+"=1")

	err := cmd.Run()
	if err == nil {
		t.Fatalf("expected subprocess to exit with error")
	}

	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("expected *exec.ExitError, got %T: %v", err, err)
	}

	if exitErr.ExitCode() != 1 {
		t.Fatalf("expected exit code 1, got %d", exitErr.ExitCode())
	}
}
