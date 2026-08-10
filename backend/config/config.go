package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	Database DatabaseConfig `json:"database"`
	Server   ServerConfig   `json:"server"`
	Security SecurityConfig `json:"security"`
	Swagger  SwaggerConfig  `json:"swagger"`
	S3       S3Config       `json:"s3"`
	Routing  RoutingConfig  `json:"routing"`
	Inbound  InboundConfig  `json:"inbound"`
}

// InboundConfig holds the order-ingestion settings ported from OrderMesh:
// mailbox scraping (IMAP or Microsoft Graph), SMTP acceptance mails, PDF
// template import and the optional Claude vision fallback. Everything is
// optional, following S3Config/RoutingConfig's "disabled if empty" pattern:
// no SMTP host means no acceptance mails, no mailbox means no scraping, no
// Anthropic key means no vision fallback — never a startup blocker.
type InboundConfig struct {
	SMTPHost string `json:"smtp_host"`
	SMTPPort string `json:"smtp_port"`
	SMTPUser string `json:"-"`
	SMTPPass string `json:"-"`
	MailFrom string `json:"mail_from"`

	// AcceptMode: "test" -> acceptance mails go to TestRecipient;
	// "production" -> acceptance mails go to the order's sender address.
	AcceptMode    string `json:"accept_mode"`
	TestRecipient string `json:"test_recipient"`

	IMAPHost string `json:"imap_host"`
	IMAPPort string `json:"imap_port"`
	IMAPUser string `json:"-"`
	IMAPPass string `json:"-"`

	// MailBackend: "imap", "graph" or "auto" (graph for Microsoft 365 hosts,
	// imap otherwise). Exchange Online rejects password auth on IMAP, so
	// office365/outlook mailboxes must go through Microsoft Graph.
	MailBackend   string `json:"mail_backend"`
	GraphClientID string `json:"graph_client_id"`
	GraphTenant   string `json:"graph_tenant"`
	// App-only (client credentials) mode: with GraphClientSecret set the
	// service authenticates as the application itself — no interactive
	// sign-in. GraphMailbox is the mailbox to read (defaults to IMAPUser).
	GraphClientSecret string `json:"-"`
	GraphMailbox      string `json:"graph_mailbox"`

	// SubjectFilter: only mails whose subject contains this marker are
	// parsed as inbound orders.
	SubjectFilter     string `json:"subject_filter"`
	ScrapeIntervalMin int    `json:"scrape_interval_min"`

	// DataDir is where the Microsoft Graph token is persisted (a mounted
	// volume in Docker) — only used by the delegated sign-in flow.
	DataDir string `json:"data_dir"`

	// AnthropicAPIKey enables the Claude vision fallback for scanned PDFs.
	AnthropicAPIKey string `json:"-"`
}

func (c InboundConfig) SMTPConfigured() bool { return c.SMTPHost != "" }
func (c InboundConfig) IMAPConfigured() bool { return c.IMAPHost != "" }

// Backend resolves MailBackend, mapping "auto" onto the mailbox host.
func (c InboundConfig) Backend() string {
	switch c.MailBackend {
	case "imap", "graph":
		return c.MailBackend
	}
	h := strings.ToLower(c.IMAPHost)
	if h == "" || strings.Contains(h, "office365") || strings.Contains(h, "outlook") {
		return "graph"
	}
	return "imap"
}

// RoutingConfig holds the OpenRouteService settings used for truck-aware
// (driving-hgv) route computation — internal/services/geo.GeoService. Empty
// key means routing is disabled: GetRoadRoute degrades to nil, same as an
// unreachable routing backend, mirroring S3Config's "disabled if bucket
// empty" pattern. BaseURL is configurable because api.openrouteservice.org
// and api.heigit.org are the same backend but not always equally reachable
// from every network (seen in practice: one resolves/routes, the other
// doesn't, depending on the host's DNS/egress).
type RoutingConfig struct {
	ORSApiKey  string `json:"-"`
	ORSBaseURL string `json:"ors_base_url"`
}

// S3Config mirrors backend/config.py's aws_region/s3_invoices_* settings.
// Bucket empty (the default) means S3 archival is disabled — see
// pkg/s3invoices.IsEnabled, matching Python's own dev-mode no-op behavior.
type S3Config struct {
	Region                 string `json:"aws_region"`
	InvoicesBucket         string `json:"s3_invoices_bucket"`
	InvoicesRetentionYears int    `json:"s3_invoices_retention_years"`
	PresignedTTLSeconds    int    `json:"s3_presigned_ttl_seconds"`
}

type DatabaseConfig struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	User     string `json:"user"`
	Password string `json:"-"`
	DBName   string `json:"dbname"`
	SSLMode  string `json:"sslmode"`
}

type SecurityConfig struct {
	JWTAccessSecret  string `json:"-"`
	JWTRefreshSecret string `json:"-"`
	JWTAccessTTL     string `json:"jwt_access_ttl"`
	JWTRefreshTTL    string `json:"jwt_refresh_ttl"`
}

type ServerConfig struct {
	Port            string `json:"port"`
	Host            string `json:"host"`
	RateLimitMax    int    `json:"rate_limit_max"`
	RateLimitWindow int    `json:"rate_limit_window"`
}

type SwaggerConfig struct {
	Host    string   `json:"host"`
	Schemes []string `json:"schemes"`
}

func Load() *Config {
	cfg := &Config{
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", ""),
			DBName:   getEnv("DB_NAME", "appdb"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Server: ServerConfig{
			Port:            getEnv("SERVER_PORT", "8080"),
			Host:            getEnv("SERVER_HOST", "0.0.0.0"),
			RateLimitMax:    getEnvInt("RATE_LIMIT_MAX", 300),
			RateLimitWindow: getEnvInt("RATE_LIMIT_WINDOW", 60),
		},
		Security: SecurityConfig{
			JWTAccessSecret:  getEnv("JWT_ACCESS_SECRET", ""),
			JWTRefreshSecret: getEnv("JWT_REFRESH_SECRET", ""),
			JWTAccessTTL:     getEnv("JWT_ACCESS_TTL_MINUTES", "60"),
			JWTRefreshTTL:    getEnv("JWT_REFRESH_TTL_HOURS", "24"),
		},
		Swagger: SwaggerConfig{
			Host:    getSwaggerHost(),
			Schemes: getSwaggerSchemes(),
		},
		S3: S3Config{
			Region:                 getEnv("AWS_REGION", "eu-west-1"),
			InvoicesBucket:         getEnv("S3_INVOICES_BUCKET", ""),
			InvoicesRetentionYears: getEnvInt("S3_INVOICES_RETENTION_YEARS", 10),
			PresignedTTLSeconds:    getEnvInt("S3_PRESIGNED_TTL_SECONDS", 900),
		},
		Routing: RoutingConfig{
			ORSApiKey:  getEnv("ORS_API_KEY", ""),
			ORSBaseURL: getEnv("ORS_BASE_URL", "https://api.openrouteservice.org"),
		},
		Inbound: InboundConfig{
			SMTPHost:          getEnv("SMTP_HOST", ""),
			SMTPPort:          getEnv("SMTP_PORT", "587"),
			SMTPUser:          getEnv("SMTP_USER", ""),
			SMTPPass:          getEnv("SMTP_PASS", ""),
			MailFrom:          getEnv("MAIL_FROM", ""),
			AcceptMode:        getEnv("ACCEPT_MODE", "test"),
			TestRecipient:     getEnv("TEST_RECIPIENT", ""),
			IMAPHost:          getEnv("IMAP_HOST", ""),
			IMAPPort:          getEnv("IMAP_PORT", "993"),
			IMAPUser:          getEnv("IMAP_USER", ""),
			IMAPPass:          getEnv("IMAP_PASS", ""),
			MailBackend:       getEnv("MAIL_BACKEND", "auto"),
			GraphClientID:     getEnv("GRAPH_CLIENT_ID", ""),
			GraphTenant:       getEnv("GRAPH_TENANT", "organizations"),
			GraphClientSecret: getEnv("GRAPH_CLIENT_SECRET", ""),
			GraphMailbox:      getEnv("GRAPH_MAILBOX", os.Getenv("IMAP_USER")),
			SubjectFilter:     getEnv("SUBJECT_FILTER", "[ORDINE]"),
			ScrapeIntervalMin: getEnvInt("SCRAPE_INTERVAL_MIN", 5),
			DataDir:           getEnv("DATA_DIR", "data"),
			AnthropicAPIKey:   getEnv("ANTHROPIC_API_KEY", ""),
		},
	}

	// A scrape interval below one minute makes no sense (and 0 would make the
	// ticker panic later) — fall back to the default like OrderMesh did.
	if cfg.Inbound.ScrapeIntervalMin < 1 {
		cfg.Inbound.ScrapeIntervalMin = 5
	}

	os.Unsetenv("DB_PASSWORD")
	os.Unsetenv("JWT_ACCESS_SECRET")
	os.Unsetenv("JWT_REFRESH_SECRET")
	os.Unsetenv("SMTP_PASS")
	os.Unsetenv("IMAP_PASS")
	os.Unsetenv("GRAPH_CLIENT_SECRET")
	os.Unsetenv("ANTHROPIC_API_KEY")

	validateConfig(cfg)

	return cfg
}

func validateConfig(cfg *Config) {
	required := map[string]string{
		"DB_PASSWORD":        cfg.Database.Password,
		"JWT_ACCESS_SECRET":  cfg.Security.JWTAccessSecret,
		"JWT_REFRESH_SECRET": cfg.Security.JWTRefreshSecret,
	}

	var missing []string
	for k, v := range required {
		if v == "" {
			missing = append(missing, k)
		}
	}

	if len(missing) > 0 {
		fmt.Fprintf(os.Stderr, "Error: required environment variables not set: %v\n", missing)
		os.Exit(1)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		var result int
		if _, err := fmt.Sscanf(value, "%d", &result); err == nil {
			return result
		}
	}
	return defaultValue
}

func getSwaggerHost() string {
	if host := os.Getenv("SWAGGER_HOST"); host != "" {
		return host
	}
	if strings.EqualFold(os.Getenv("IS_LOCAL"), "true") {
		return "localhost:" + getEnv("SERVER_PORT", "8080")
	}
	return "localhost:8080"
}

func getSwaggerSchemes() []string {
	if schemes := os.Getenv("SWAGGER_SCHEMES"); schemes != "" {
		var result []string
		for _, s := range strings.Split(schemes, ",") {
			if s = strings.TrimSpace(s); s != "" {
				result = append(result, s)
			}
		}
		if len(result) > 0 {
			return result
		}
	}
	return []string{"http", "https"}
}
