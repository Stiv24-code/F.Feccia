package handlers

import (
	"strconv"
	"strings"

	"fratelli-feccia/internal/dto"
	"fratelli-feccia/internal/services"
	"fratelli-feccia/pkg/audit"
	"fratelli-feccia/pkg/utils"

	"github.com/gofiber/fiber/v2"
)

// Login/refresh/logout are excluded from the generic HTTP audit middleware
// (see pkg/middleware/audit_http.go) and logged explicitly here instead,
// mirroring backend/routers/auth.py — a raw status code can't distinguish
// "wrong password" from "user disabled", which the audit trail needs.
type AuthHandler struct {
	Service     services.Auth
	AuditLogger *audit.Logger
	JWTConfig   utils.JWTConfig
}

func NewAuthHandler(service services.Auth, auditLogger *audit.Logger, jwtCfg utils.JWTConfig) *AuthHandler {
	return &AuthHandler{Service: service, AuditLogger: auditLogger, JWTConfig: jwtCfg}
}

// Login godoc
// @Summary User login
// @Description Authenticate user; access token in body, refresh token as httpOnly cookie
// @Tags Auth
// @Accept json
// @Produce json
// @Param credentials body dto.LoginRequest true "Login credentials"
// @Success 200 {object} dto.LoginResult
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req dto.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, 400, "Invalid request body")
	}

	validator := utils.NewValidator()
	if errs := validator.Validate(&req); len(errs) > 0 {
		return utils.ValidationErrorResponse(c, errs)
	}

	ctx := utils.RequestContext(c)
	ip, ua := c.IP(), c.Get("User-Agent")

	result, err := h.Service.Login(req.Email, req.Password)
	if err != nil {
		sc, isAPIError := err.(interface{ StatusCode() int })
		status := 401
		if isAPIError {
			status = sc.StatusCode()
		}
		h.AuditLogger.Log(ctx, audit.Entry{
			Action: "auth.login", Resource: "user", StatusCode: status, Success: false,
			IP: ip, UserAgent: ua, Error: err.Error(), Metadata: map[string]interface{}{"email": req.Email},
		})
		if isAPIError {
			return utils.HandleDatabaseError(c, err)
		}
		return utils.ErrorResponse(c, 401, err.Error())
	}

	userID := result.User.ID
	h.AuditLogger.Log(ctx, audit.Entry{
		Action: "auth.login", UserID: &userID, UserRole: result.User.Role,
		Resource: "user", ResourceID: strconv.FormatInt(userID, 10), StatusCode: 200, Success: true,
		IP: ip, UserAgent: ua,
	})

	utils.SetRefreshCookie(c, result.RefreshToken, result.RefreshTTL)
	return utils.SuccessResponse(c, 200, result)
}

// Refresh godoc
// @Summary Refresh tokens
// @Description Exchange the httpOnly refresh cookie for a new access token (and rotate the cookie)
// @Tags Auth
// @Produce json
// @Success 200 {object} dto.LoginResult
// @Failure 401 {object} map[string]string
// @Router /api/v1/auth/refresh [post]
func (h *AuthHandler) Refresh(c *fiber.Ctx) error {
	ctx := utils.RequestContext(c)
	ip, ua := c.IP(), c.Get("User-Agent")

	refreshToken := c.Cookies(utils.RefreshCookieName)
	if refreshToken == "" {
		h.AuditLogger.Log(ctx, audit.Entry{
			Action: "auth.refresh", StatusCode: 401, Success: false,
			IP: ip, UserAgent: ua, Error: "missing_cookie",
		})
		return utils.ErrorResponse(c, 401, "Refresh token mancante")
	}

	result, err := h.Service.Refresh(refreshToken)
	if err != nil {
		utils.ClearRefreshCookie(c)
		h.AuditLogger.Log(ctx, audit.Entry{
			Action: "auth.refresh", StatusCode: 401, Success: false,
			IP: ip, UserAgent: ua, Error: err.Error(),
		})
		return utils.ErrorResponse(c, 401, err.Error())
	}

	userID := result.User.ID
	h.AuditLogger.Log(ctx, audit.Entry{
		Action: "auth.refresh", UserID: &userID, UserRole: result.User.Role,
		Resource: "user", ResourceID: strconv.FormatInt(userID, 10), StatusCode: 200, Success: true,
		IP: ip, UserAgent: ua,
	})

	utils.SetRefreshCookie(c, result.RefreshToken, result.RefreshTTL)
	return utils.SuccessResponse(c, 200, result)
}

// Me godoc
// @Summary Current user
// @Tags Auth
// @Security BearerAuth
// @Produce json
// @Success 200 {object} dto.AuthUserResponse
// @Failure 401 {object} map[string]string
// @Router /api/v1/auth/me [get]
func (h *AuthHandler) Me(c *fiber.Ctx) error {
	userID, _ := c.Locals("user_id").(int64)
	user, err := h.Service.Me(userID)
	if err != nil {
		return utils.ErrorResponse(c, 401, "Utente non disponibile")
	}
	return utils.SuccessResponse(c, 200, user)
}

// Register godoc
// @Summary Register a new user (admin-only)
// @Tags Auth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param user body dto.RegisterRequest true "New user data"
// @Success 200 {object} dto.AuthUserResponse
// @Failure 400 {object} map[string]string
// @Router /api/v1/auth/register [post]
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req dto.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, 400, "Invalid request body")
	}

	validator := utils.NewValidator()
	if errs := validator.Validate(&req); len(errs) > 0 {
		return utils.ValidationErrorResponse(c, errs)
	}

	result, err := h.Service.Register(req)
	if err != nil {
		return utils.HandleDatabaseError(c, err)
	}
	return utils.SuccessResponse(c, 200, result)
}

// RegisterClient godoc
// @Summary Self-service client registration (public)
// @Description Creates a Customer (anagrafica) + a "cliente" account atomically, then logs it in immediately (no approval step) — access token in body, refresh token as httpOnly cookie, same as Login.
// @Tags Auth
// @Accept json
// @Produce json
// @Param registration body dto.ClientRegisterRequest true "Client registration data"
// @Success 200 {object} dto.LoginResult
// @Failure 400 {object} map[string]string
// @Router /api/v1/auth/register-cliente [post]
func (h *AuthHandler) RegisterClient(c *fiber.Ctx) error {
	var req dto.ClientRegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, 400, "Invalid request body")
	}
	if errs := utils.NewValidator().Validate(&req); len(errs) > 0 {
		return utils.ValidationErrorResponse(c, errs)
	}

	ctx := utils.RequestContext(c)
	ip, ua := c.IP(), c.Get("User-Agent")

	result, err := h.Service.RegisterClient(req)
	if err != nil {
		h.AuditLogger.Log(ctx, audit.Entry{
			Action: "auth.register_cliente", Resource: "user", StatusCode: 400, Success: false,
			IP: ip, UserAgent: ua, Error: err.Error(), Metadata: map[string]interface{}{"email": req.Email},
		})
		return utils.HandleDatabaseError(c, err)
	}

	userID := result.User.ID
	h.AuditLogger.Log(ctx, audit.Entry{
		Action: "auth.register_cliente", UserID: &userID, UserRole: result.User.Role,
		Resource: "user", ResourceID: strconv.FormatInt(userID, 10), StatusCode: 200, Success: true,
		IP: ip, UserAgent: ua,
	})

	utils.SetRefreshCookie(c, result.RefreshToken, result.RefreshTTL)
	return utils.SuccessResponse(c, 200, result)
}

// Logout godoc
// @Summary Logout
// @Description Clears the refresh cookie; best-effort, works without a valid token
// @Tags Auth
// @Produce json
// @Success 200 {object} map[string]bool
// @Router /api/v1/auth/logout [post]
func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	utils.ClearRefreshCookie(c)

	var userID *int64
	var role string
	authz := c.Get("Authorization")
	if strings.HasPrefix(strings.ToLower(authz), "bearer ") {
		if claims, err := utils.ParseAccessToken(strings.TrimSpace(authz[7:]), h.JWTConfig); err == nil {
			userID = &claims.UserID
			role = claims.Role
		}
	}

	h.AuditLogger.Log(utils.RequestContext(c), audit.Entry{
		Action: "auth.logout", UserID: userID, UserRole: role, Resource: "user",
		StatusCode: 200, Success: true, IP: c.IP(), UserAgent: c.Get("User-Agent"),
	})

	return utils.SuccessResponse(c, 200, fiber.Map{"ok": true})
}
