package utils

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	RoleAdmin           = "admin"
	RoleAmministrazione = "amministrazione"
	RolePlanner         = "planner"
	RoleOperatore       = "operatore"
)

func IsValidRole(role string) bool {
	switch role {
	case RoleAdmin, RoleAmministrazione, RolePlanner, RoleOperatore:
		return true
	default:
		return false
	}
}

type JWTConfig struct {
	AccessSecret  string
	RefreshSecret string
	AccessTTL     time.Duration
	RefreshTTL    time.Duration
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type UserClaims struct {
	UserID int64  `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func NewJWTConfig(accessSecret, refreshSecret, accessTTLMinutes, refreshTTLHours string) JWTConfig {
	accessTTL := 60 * time.Minute
	if v, err := strconv.Atoi(accessTTLMinutes); err == nil {
		accessTTL = time.Duration(v) * time.Minute
	}

	refreshTTL := 24 * time.Hour
	if v, err := strconv.Atoi(refreshTTLHours); err == nil {
		refreshTTL = time.Duration(v) * time.Hour
	}

	cfg := JWTConfig{
		AccessSecret:  accessSecret,
		RefreshSecret: refreshSecret,
		AccessTTL:     accessTTL,
		RefreshTTL:    refreshTTL,
	}

	if cfg.AccessSecret == "" || cfg.RefreshSecret == "" {
		panic("JWT secrets must be provided")
	}

	return cfg
}

func GenerateTokenPair(userID int64, role string, cfg JWTConfig) (TokenPair, error) {
	if !IsValidRole(role) {
		return TokenPair{}, fmt.Errorf("invalid role %q", role)
	}

	accessClaims := UserClaims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(cfg.AccessTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString([]byte(cfg.AccessSecret))
	if err != nil {
		return TokenPair{}, err
	}

	refreshClaims := UserClaims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(cfg.RefreshTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte(cfg.RefreshSecret))
	if err != nil {
		return TokenPair{}, err
	}

	return TokenPair{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}

func ParseAccessToken(tokenStr string, cfg JWTConfig) (*UserClaims, error) {
	return parseToken(tokenStr, cfg.AccessSecret)
}

func ParseRefreshToken(tokenStr string, cfg JWTConfig) (*UserClaims, error) {
	return parseToken(tokenStr, cfg.RefreshSecret)
}

func parseToken(tokenStr, secret string) (*UserClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &UserClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*UserClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	if !IsValidRole(claims.Role) {
		return nil, errors.New("invalid role")
	}
	return claims, nil
}
