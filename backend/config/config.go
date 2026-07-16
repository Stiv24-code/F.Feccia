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
	}

	os.Unsetenv("DB_PASSWORD")
	os.Unsetenv("JWT_ACCESS_SECRET")
	os.Unsetenv("JWT_REFRESH_SECRET")

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
