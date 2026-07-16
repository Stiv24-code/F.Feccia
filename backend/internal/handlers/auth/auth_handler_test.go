package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"fratelli-feccia/internal/dto"
	"fratelli-feccia/internal/services"
	"fratelli-feccia/internal/services/mocks"
	"fratelli-feccia/pkg/audit"
	"fratelli-feccia/pkg/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/golang/mock/gomock"
)

func newAuthTestApp(svc services.Auth) *fiber.App {
	app := fiber.New()
	// nil DB is fine: audit.Logger.Log() recovers from the resulting panic
	// and merely logs the failure — audit persistence isn't under test here.
	h := NewAuthHandler(svc, audit.NewLogger(nil), utils.JWTConfig{})

	app.Post("/login", h.Login)
	app.Post("/refresh", h.Refresh)
	app.Post("/logout", h.Logout)
	app.Post("/register", h.Register)
	app.Get("/me", func(c *fiber.Ctx) error {
		c.Locals("user_id", int64(1))
		return h.Me(c)
	})

	return app
}

func doRequest(t *testing.T, app *fiber.App, method, path string, body interface{}, headers map[string]string) (*http.Response, map[string]interface{}) {
	t.Helper()

	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatalf("failed to encode body: %v", err)
		}
	}

	req := httptest.NewRequest(method, path, &buf)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test error: %v", err)
	}

	var respBody map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	return resp, respBody
}

func sampleLoginResult() *dto.LoginResult {
	return &dto.LoginResult{
		AccessToken:  "access-token",
		RefreshToken: "refresh-token",
		TokenType:    "Bearer",
		ExpiresIn:    900,
		User:         dto.AuthUserResponse{ID: 1, Email: "admin", Name: "Admin", Role: "admin", Active: true},
	}
}

func TestAuthHandler_Login_InvalidJSON(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuth := mocks.NewMockAuth(ctrl)
	app := newAuthTestApp(mockAuth)

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString("not-json"))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test error: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}
}

func TestAuthHandler_Login_ValidationError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuth := mocks.NewMockAuth(ctrl)
	app := newAuthTestApp(mockAuth)

	// Missing both email and password.
	resp, body := doRequest(t, app, http.MethodPost, "/login", map[string]interface{}{}, nil)

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}
	if body["error"] != "Validation failed" {
		t.Fatalf("expected error %q, got %v", "Validation failed", body["error"])
	}
}

func TestAuthHandler_Login_InvalidCredentials(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuth := mocks.NewMockAuth(ctrl)
	mockAuth.EXPECT().
		Login("user@example.it", "wrong").
		Return(nil, errors.New("invalid credentials"))

	app := newAuthTestApp(mockAuth)

	reqBody := map[string]string{"email": "user@example.it", "password": "wrong"}
	resp, body := doRequest(t, app, http.MethodPost, "/login", reqBody, nil)

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, resp.StatusCode)
	}
	if body["error"] != "invalid credentials" {
		t.Fatalf("expected error %q, got %v", "invalid credentials", body["error"])
	}
}

func TestAuthHandler_Login_Success(t *testing.T) {
	t.Parallel()
	expected := sampleLoginResult()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuth := mocks.NewMockAuth(ctrl)
	mockAuth.EXPECT().
		Login("user@example.it", "pass").
		Return(expected, nil)

	app := newAuthTestApp(mockAuth)

	reqBody := map[string]string{"email": "user@example.it", "password": "pass"}
	resp, body := doRequest(t, app, http.MethodPost, "/login", reqBody, nil)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}
	if body["access_token"] != expected.AccessToken {
		t.Fatalf("expected access_token %q, got %v", expected.AccessToken, body["access_token"])
	}
	if _, ok := body["refresh_token"]; ok {
		t.Fatalf("refresh_token must never appear in the JSON body, got %v", body["refresh_token"])
	}
	user, ok := body["user"].(map[string]interface{})
	if !ok || user["role"] != "admin" {
		t.Fatalf("expected embedded user object with role, got %v", body["user"])
	}

	var refreshCookie *http.Cookie
	for _, ck := range resp.Cookies() {
		if ck.Name == "tms_refresh" {
			refreshCookie = ck
		}
	}
	if refreshCookie == nil {
		t.Fatal("expected tms_refresh cookie to be set on successful login")
	}
	if !refreshCookie.HttpOnly {
		t.Fatal("expected tms_refresh cookie to be HttpOnly")
	}
	if refreshCookie.Value != expected.RefreshToken {
		t.Fatalf("expected cookie value %q, got %q", expected.RefreshToken, refreshCookie.Value)
	}
}

func TestAuthHandler_Refresh_MissingCookie(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Service should not be called when the refresh cookie is missing.
	mockAuth := mocks.NewMockAuth(ctrl)
	app := newAuthTestApp(mockAuth)

	resp, body := doRequest(t, app, http.MethodPost, "/refresh", nil, nil)

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, resp.StatusCode)
	}
	if body["error"] != "Refresh token mancante" {
		t.Fatalf("expected error %q, got %v", "Refresh token mancante", body["error"])
	}
}

func TestAuthHandler_Refresh_TokenFromCookie(t *testing.T) {
	t.Parallel()
	expected := sampleLoginResult()
	expected.AccessToken = "new-access"
	expected.RefreshToken = "new-refresh"

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuth := mocks.NewMockAuth(ctrl)
	mockAuth.EXPECT().
		Refresh("cookie-refresh").
		Return(expected, nil)

	app := newAuthTestApp(mockAuth)

	headers := map[string]string{"Cookie": "tms_refresh=cookie-refresh"}
	resp, body := doRequest(t, app, http.MethodPost, "/refresh", nil, headers)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}
	if body["access_token"] != expected.AccessToken {
		t.Fatalf("expected access_token %q, got %v", expected.AccessToken, body["access_token"])
	}
}

func TestAuthHandler_Refresh_InvalidToken(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuth := mocks.NewMockAuth(ctrl)
	mockAuth.EXPECT().
		Refresh("bad-token").
		Return(nil, errors.New("invalid refresh token"))

	app := newAuthTestApp(mockAuth)

	headers := map[string]string{"Cookie": "tms_refresh=bad-token"}
	resp, body := doRequest(t, app, http.MethodPost, "/refresh", nil, headers)

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, resp.StatusCode)
	}
	if body["error"] != "invalid refresh token" {
		t.Fatalf("expected error %q, got %v", "invalid refresh token", body["error"])
	}
}

func TestAuthHandler_Me_Success(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuth := mocks.NewMockAuth(ctrl)
	mockAuth.EXPECT().
		Me(int64(1)).
		Return(&dto.AuthUserResponse{ID: 1, Email: "admin", Name: "Admin", Role: "admin", Active: true}, nil)

	app := newAuthTestApp(mockAuth)
	resp, body := doRequest(t, app, http.MethodGet, "/me", nil, nil)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}
	if body["role"] != "admin" {
		t.Fatalf("expected role admin, got %v", body["role"])
	}
}

func TestAuthHandler_Login_DisabledUserReturns403(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuth := mocks.NewMockAuth(ctrl)
	mockAuth.EXPECT().
		Login("user@example.it", "pass").
		Return(nil, utils.NewAPIError(403, "Utente disattivato"))

	app := newAuthTestApp(mockAuth)

	reqBody := map[string]string{"email": "user@example.it", "password": "pass"}
	resp, body := doRequest(t, app, http.MethodPost, "/login", reqBody, nil)

	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, resp.StatusCode)
	}
	if body["error"] != "Utente disattivato" {
		t.Fatalf("expected error %q, got %v", "Utente disattivato", body["error"])
	}
}

func TestAuthHandler_Register_Success(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expected := &dto.AuthUserResponse{ID: 2, Email: "new@example.it", Name: "New User", Role: "operatore", Active: true}
	mockAuth := mocks.NewMockAuth(ctrl)
	mockAuth.EXPECT().
		Register(dto.RegisterRequest{Email: "new@example.it", Name: "New User", Password: "supersecretpw12", Role: "operatore"}).
		Return(expected, nil)

	app := newAuthTestApp(mockAuth)

	reqBody := map[string]string{"email": "new@example.it", "name": "New User", "password": "supersecretpw12", "role": "operatore"}
	resp, body := doRequest(t, app, http.MethodPost, "/register", reqBody, nil)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}
	if body["email"] != "new@example.it" {
		t.Fatalf("expected email in response, got %v", body)
	}
}

func TestAuthHandler_Register_ValidationError_ShortPassword(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuth := mocks.NewMockAuth(ctrl)
	app := newAuthTestApp(mockAuth)

	reqBody := map[string]string{"email": "new@example.it", "name": "New User", "password": "short", "role": "operatore"}
	resp, _ := doRequest(t, app, http.MethodPost, "/register", reqBody, nil)

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status %d for password under 12 chars, got %d", http.StatusBadRequest, resp.StatusCode)
	}
}

func TestAuthHandler_Register_DuplicateEmail(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuth := mocks.NewMockAuth(ctrl)
	mockAuth.EXPECT().
		Register(gomock.Any()).
		Return(nil, utils.NewAPIError(400, "Email già registrata"))

	app := newAuthTestApp(mockAuth)

	reqBody := map[string]string{"email": "dup@example.it", "name": "Dup", "password": "supersecretpw12", "role": "operatore"}
	resp, body := doRequest(t, app, http.MethodPost, "/register", reqBody, nil)

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}
	if body["error"] != "Email già registrata" {
		t.Fatalf("expected error %q, got %v", "Email già registrata", body["error"])
	}
}

func TestAuthHandler_Logout_ClearsCookie(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuth := mocks.NewMockAuth(ctrl)
	app := newAuthTestApp(mockAuth)

	resp, body := doRequest(t, app, http.MethodPost, "/logout", nil, nil)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}
	if body["ok"] != true {
		t.Fatalf("expected ok:true, got %v", body)
	}

	var refreshCookie *http.Cookie
	for _, ck := range resp.Cookies() {
		if ck.Name == "tms_refresh" {
			refreshCookie = ck
		}
	}
	if refreshCookie == nil || refreshCookie.Value != "" {
		t.Fatalf("expected tms_refresh cookie to be cleared (empty value), got %+v", refreshCookie)
	}
}
