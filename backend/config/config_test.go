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

func TestLoad_InboundDefaults(t *testing.T) {
	// Manipulates many env vars; avoid running in parallel with other env tests.
	t.Setenv("DB_PASSWORD", "secret")
	t.Setenv("JWT_ACCESS_SECRET", "access_secret")
	t.Setenv("JWT_REFRESH_SECRET", "refresh_secret")

	for _, k := range []string{
		"SMTP_HOST", "SMTP_PORT", "SMTP_USER", "SMTP_PASS", "MAIL_FROM",
		"ACCEPT_MODE", "TEST_RECIPIENT",
		"IMAP_HOST", "IMAP_PORT", "IMAP_USER", "IMAP_PASS",
		"MAIL_BACKEND", "GRAPH_CLIENT_ID", "GRAPH_TENANT", "GRAPH_CLIENT_SECRET", "GRAPH_MAILBOX",
		"SUBJECT_FILTER", "SCRAPE_INTERVAL_MIN", "DATA_DIR", "ANTHROPIC_API_KEY",
	} {
		_ = os.Unsetenv(k)
	}

	cfg := Load()

	in := cfg.Inbound
	if in.SMTPPort != "587" {
		t.Fatalf("expected SMTP port %q, got %q", "587", in.SMTPPort)
	}
	if in.IMAPPort != "993" {
		t.Fatalf("expected IMAP port %q, got %q", "993", in.IMAPPort)
	}
	if in.AcceptMode != "test" {
		t.Fatalf("expected accept mode %q, got %q", "test", in.AcceptMode)
	}
	if in.MailBackend != "auto" {
		t.Fatalf("expected mail backend %q, got %q", "auto", in.MailBackend)
	}
	if in.GraphTenant != "organizations" {
		t.Fatalf("expected graph tenant %q, got %q", "organizations", in.GraphTenant)
	}
	if in.SubjectFilter != "[ORDINE]" {
		t.Fatalf("expected subject filter %q, got %q", "[ORDINE]", in.SubjectFilter)
	}
	if in.ScrapeIntervalMin != 5 {
		t.Fatalf("expected scrape interval 5, got %d", in.ScrapeIntervalMin)
	}
	if in.DataDir != "data" {
		t.Fatalf("expected data dir %q, got %q", "data", in.DataDir)
	}
	if in.SMTPConfigured() {
		t.Fatalf("expected SMTPConfigured()==false with no SMTP_HOST")
	}
	if in.IMAPConfigured() {
		t.Fatalf("expected IMAPConfigured()==false with no IMAP_HOST")
	}
}

func TestLoad_InboundOverridesAndSecretUnset(t *testing.T) {
	// Manipulates many env vars; avoid running in parallel with other env tests.
	t.Setenv("DB_PASSWORD", "secret")
	t.Setenv("JWT_ACCESS_SECRET", "access_secret")
	t.Setenv("JWT_REFRESH_SECRET", "refresh_secret")

	t.Setenv("SMTP_HOST", "smtp.example.com")
	t.Setenv("SMTP_PASS", "smtp_secret")
	t.Setenv("IMAP_HOST", "imap.example.com")
	t.Setenv("IMAP_USER", "orders@example.com")
	_ = os.Unsetenv("GRAPH_MAILBOX")
	t.Setenv("SCRAPE_INTERVAL_MIN", "0") // below floor -> falls back to 5

	cfg := Load()

	in := cfg.Inbound
	if in.SMTPHost != "smtp.example.com" || !in.SMTPConfigured() {
		t.Fatalf("expected SMTP host override, got %q", in.SMTPHost)
	}
	if in.SMTPPass != "smtp_secret" {
		t.Fatalf("expected SMTP pass to be loaded, got %q", in.SMTPPass)
	}
	if in.GraphMailbox != "orders@example.com" {
		t.Fatalf("expected graph mailbox to default to IMAP_USER, got %q", in.GraphMailbox)
	}
	if in.ScrapeIntervalMin != 5 {
		t.Fatalf("expected scrape interval floor 5, got %d", in.ScrapeIntervalMin)
	}
	// Secrets are wiped from the process env after load, like DB_PASSWORD.
	if v := os.Getenv("SMTP_PASS"); v != "" {
		t.Fatalf("expected SMTP_PASS to be unset after Load, got %q", v)
	}
}

func TestInboundConfig_BackendResolution(t *testing.T) {
	cases := []struct {
		name     string
		cfg      InboundConfig
		expected string
	}{
		{"explicit imap wins over ms host", InboundConfig{MailBackend: "imap", IMAPHost: "outlook.office365.com"}, "imap"},
		{"explicit graph", InboundConfig{MailBackend: "graph"}, "graph"},
		{"auto with office365 host", InboundConfig{MailBackend: "auto", IMAPHost: "outlook.office365.com"}, "graph"},
		{"auto with outlook host", InboundConfig{MailBackend: "auto", IMAPHost: "imap.OUTLOOK.com"}, "graph"},
		{"auto with generic host", InboundConfig{MailBackend: "auto", IMAPHost: "imaps.aruba.it"}, "imap"},
		{"auto with no host", InboundConfig{MailBackend: "auto"}, "graph"},
	}
	for _, tc := range cases {
		if got := tc.cfg.Backend(); got != tc.expected {
			t.Errorf("%s: expected backend %q, got %q", tc.name, tc.expected, got)
		}
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
