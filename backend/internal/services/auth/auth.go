package auth

import (
	"errors"
	"time"

	"fratelli-feccia/internal/dto"
	"fratelli-feccia/internal/models"
	"fratelli-feccia/pkg/utils"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService struct {
	db      *gorm.DB
	jwtConf utils.JWTConfig
}

func NewAuthService(db *gorm.DB, jwtConf utils.JWTConfig) *AuthService {
	return &AuthService{
		db:      db,
		jwtConf: jwtConf,
	}
}

func (s *AuthService) Login(login, password string) (*dto.LoginResult, error) {
	var user models.User
	if err := s.db.Where("login = ?", login).First(&user).Error; err != nil {
		return nil, errors.New("invalid credentials")
	}

	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		return nil, errors.New("invalid credentials")
	}

	if !user.Active {
		return nil, utils.NewAPIError(403, "Utente disattivato")
	}

	result, err := s.buildLoginResult(user)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	s.db.Model(&user).Update("last_login_at", &now)

	return result, nil
}

func (s *AuthService) Refresh(refreshToken string) (*dto.LoginResult, error) {
	claims, err := utils.ParseRefreshToken(refreshToken, s.jwtConf)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

	var user models.User
	if err := s.db.First(&user, claims.UserID).Error; err != nil {
		return nil, errors.New("invalid user")
	}
	if !user.Active {
		return nil, errors.New("invalid user")
	}

	return s.buildLoginResult(user)
}

func (s *AuthService) Me(userID int64) (*dto.AuthUserResponse, error) {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return nil, errors.New("invalid user")
	}
	if !user.Active {
		return nil, errors.New("invalid user")
	}
	resp := toAuthUserResponse(user)
	return &resp, nil
}

// Register mirrors POST /auth/register (admin-only): creates a new user.
// Unlike CreateUser (the pre-existing, unused-by-frontend admin panel path),
// this uses email/min-12-char-password to match the real frontend contract.
func (s *AuthService) Register(req dto.RegisterRequest) (*dto.AuthUserResponse, error) {
	var count int64
	s.db.Model(&models.User{}).Where("login = ?", req.Email).Count(&count)
	if count > 0 {
		return nil, utils.NewAPIError(400, "Email già registrata")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := models.User{
		Login: req.Email, Name: req.Name, PasswordHash: string(hash), Role: req.Role, Active: true,
	}
	if err := s.db.Create(&user).Error; err != nil {
		return nil, err
	}

	resp := toAuthUserResponse(user)
	return &resp, nil
}

// RegisterClient mirrors POST /auth/register-cliente (public, unauthenticated):
// atomically creates a Customer (anagrafica) and a RoleCliente User linked to
// it, then logs the new account in immediately (no approval step) — same
// LoginResult shape as Login/Refresh, so the frontend can go straight into
// the client portal.
func (s *AuthService) RegisterClient(req dto.ClientRegisterRequest) (*dto.LoginResult, error) {
	var count int64
	s.db.Model(&models.User{}).Where("login = ?", req.Email).Count(&count)
	if count > 0 {
		return nil, utils.NewAPIError(400, "Email già registrata")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	var user models.User
	err = s.db.Transaction(func(tx *gorm.DB) error {
		customer := models.Customer{
			ID:             uuid.New(),
			RagioneSociale: req.RagioneSociale,
			Indirizzo:      req.Indirizzo,
			Citta:          req.Citta,
			Cap:            req.Cap,
			Provincia:      req.Provincia,
			Lat:            req.Lat,
			Lng:            req.Lng,
			PartitaIva:     req.PartitaIva,
			CodiceFiscale:  req.CodiceFiscale,
			Telefono:       req.Telefono,
			Email:          req.Email,
			Active:         true,
		}
		if err := tx.Create(&customer).Error; err != nil {
			return err
		}

		user = models.User{
			Login: req.Email, Name: req.Name, PasswordHash: string(hash),
			Role: utils.RoleCliente, CustomerID: &customer.ID, Active: true,
		}
		return tx.Create(&user).Error
	})
	if err != nil {
		return nil, err
	}

	return s.buildLoginResult(user)
}

func (s *AuthService) buildLoginResult(user models.User) (*dto.LoginResult, error) {
	customerID := ""
	if user.CustomerID != nil {
		customerID = user.CustomerID.String()
	}
	tokens, err := utils.GenerateTokenPair(user.ID, user.Role, customerID, s.jwtConf)
	if err != nil {
		return nil, err
	}

	return &dto.LoginResult{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		RefreshTTL:   s.jwtConf.RefreshTTL,
		TokenType:    "Bearer",
		ExpiresIn:    int64(s.jwtConf.AccessTTL.Seconds()),
		User:         toAuthUserResponse(user),
	}, nil
}

func toAuthUserResponse(u models.User) dto.AuthUserResponse {
	var customerID *string
	if u.CustomerID != nil {
		id := u.CustomerID.String()
		customerID = &id
	}
	return dto.AuthUserResponse{
		ID:         u.ID,
		Email:      u.Login,
		Name:       u.Name,
		Role:       u.Role,
		ProfileID:  nil,
		CustomerID: customerID,
		Active:     u.Active,
	}
}
