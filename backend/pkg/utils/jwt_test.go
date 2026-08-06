package utils

import (
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestNewJWTConfig(t *testing.T) {
	cfg := NewJWTConfig("access-secret", "refresh-secret", "15", "48")

	if cfg.AccessSecret != "access-secret" {
		t.Fatalf("expected AccessSecret %q, got %q", "access-secret", cfg.AccessSecret)
	}
	if cfg.RefreshSecret != "refresh-secret" {
		t.Fatalf("expected RefreshSecret %q, got %q", "refresh-secret", cfg.RefreshSecret)
	}
	if cfg.AccessTTL != 15*time.Minute {
		t.Fatalf("expected AccessTTL %v, got %v", 15*time.Minute, cfg.AccessTTL)
	}
	if cfg.RefreshTTL != 48*time.Hour {
		t.Fatalf("expected RefreshTTL %v, got %v", 48*time.Hour, cfg.RefreshTTL)
	}
}

func TestNewJWTConfigPanicsWithoutSecrets(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic when secrets are missing, got none")
		}
	}()

	_ = NewJWTConfig("", "", "", "")
}

func TestGenerateTokenPairAndParse(t *testing.T) {
	t.Parallel()
	cfg := JWTConfig{
		AccessSecret:  "access-secret",
		RefreshSecret: "refresh-secret",
		AccessTTL:     time.Minute,
		RefreshTTL:    time.Hour,
	}

	userID := int64(42)
	role := "admin"

	pair, err := GenerateTokenPair(userID, role, "", cfg)
	if err != nil {
		t.Fatalf("GenerateTokenPair returned error: %v", err)
	}
	if pair.AccessToken == "" || pair.RefreshToken == "" {
		t.Fatalf("expected non-empty tokens, got access=%q refresh=%q", pair.AccessToken, pair.RefreshToken)
	}

	accessClaims, err := ParseAccessToken(pair.AccessToken, cfg)
	if err != nil {
		t.Fatalf("ParseAccessToken returned error: %v", err)
	}
	if accessClaims.UserID != userID || accessClaims.Role != role {
		t.Fatalf("unexpected access claims: %+v", accessClaims)
	}

	refreshClaims, err := ParseRefreshToken(pair.RefreshToken, cfg)
	if err != nil {
		t.Fatalf("ParseRefreshToken returned error: %v", err)
	}
	if refreshClaims.UserID != userID || refreshClaims.Role != role {
		t.Fatalf("unexpected refresh claims: %+v", refreshClaims)
	}
}

func TestGenerateTokenPairCarriesCustomerIDForCliente(t *testing.T) {
	t.Parallel()
	cfg := JWTConfig{
		AccessSecret:  "access-secret",
		RefreshSecret: "refresh-secret",
		AccessTTL:     time.Minute,
		RefreshTTL:    time.Hour,
	}

	customerID := "11111111-1111-1111-1111-111111111111"
	pair, err := GenerateTokenPair(7, RoleCliente, customerID, cfg)
	if err != nil {
		t.Fatalf("GenerateTokenPair returned error: %v", err)
	}

	accessClaims, err := ParseAccessToken(pair.AccessToken, cfg)
	if err != nil {
		t.Fatalf("ParseAccessToken returned error: %v", err)
	}
	if accessClaims.Role != RoleCliente || accessClaims.CustomerID != customerID {
		t.Fatalf("expected role=%q customer_id=%q, got %+v", RoleCliente, customerID, accessClaims)
	}
}

func TestParseAccessTokenExpired(t *testing.T) {
	t.Parallel()
	cfg := JWTConfig{
		AccessSecret:  "access-secret",
		RefreshSecret: "refresh-secret",
		AccessTTL:     -1 * time.Minute, // already expired
		RefreshTTL:    time.Hour,
	}

	pair, err := GenerateTokenPair(1, RoleAdmin, "", cfg)
	if err != nil {
		t.Fatalf("GenerateTokenPair returned error: %v", err)
	}

	_, err = ParseAccessToken(pair.AccessToken, cfg)
	if err == nil {
		t.Fatalf("expected error for expired access token, got nil")
	}
	if !errors.Is(err, jwt.ErrTokenExpired) {
		t.Fatalf("expected ErrTokenExpired, got %v", err)
	}
}

func TestParseAccessTokenInvalidSignature(t *testing.T) {
	t.Parallel()
	cfg := JWTConfig{
		AccessSecret:  "access-secret",
		RefreshSecret: "refresh-secret",
		AccessTTL:     time.Minute,
		RefreshTTL:    time.Hour,
	}

	pair, err := GenerateTokenPair(1, RoleAdmin, "", cfg)
	if err != nil {
		t.Fatalf("GenerateTokenPair returned error: %v", err)
	}

	badCfg := JWTConfig{
		AccessSecret:  "wrong-secret",
		RefreshSecret: cfg.RefreshSecret,
		AccessTTL:     cfg.AccessTTL,
		RefreshTTL:    cfg.RefreshTTL,
	}

	_, err = ParseAccessToken(pair.AccessToken, badCfg)
	if err == nil {
		t.Fatalf("expected error for invalid signature, got nil")
	}
	if !errors.Is(err, jwt.ErrTokenSignatureInvalid) {
		t.Fatalf("expected ErrTokenSignatureInvalid, got %v", err)
	}
}

func TestParseRefreshTokenInvalidSignature(t *testing.T) {
	t.Parallel()
	cfg := JWTConfig{
		AccessSecret:  "access-secret",
		RefreshSecret: "refresh-secret",
		AccessTTL:     time.Minute,
		RefreshTTL:    time.Hour,
	}

	pair, err := GenerateTokenPair(1, RoleAdmin, "", cfg)
	if err != nil {
		t.Fatalf("GenerateTokenPair returned error: %v", err)
	}

	badCfg := JWTConfig{
		AccessSecret:  cfg.AccessSecret,
		RefreshSecret: "wrong-secret",
		AccessTTL:     cfg.AccessTTL,
		RefreshTTL:    cfg.RefreshTTL,
	}

	_, err = ParseRefreshToken(pair.RefreshToken, badCfg)
	if err == nil {
		t.Fatalf("expected error for invalid signature, got nil")
	}
	if !errors.Is(err, jwt.ErrTokenSignatureInvalid) {
		t.Fatalf("expected ErrTokenSignatureInvalid, got %v", err)
	}
}
