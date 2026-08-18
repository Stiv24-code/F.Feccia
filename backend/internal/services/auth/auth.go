package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"html"
	"time"

	"fratelli-feccia/internal/dto"
	"fratelli-feccia/internal/models"
	"fratelli-feccia/pkg/utils"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Mailer is the seam AuthService needs to send the registration-confirmation
// mail — implemented by mailer.MailerService, same posture as
// inboundorders.AcceptanceMailer: a nil Mailer means "SMTP non configurato",
// and RegisterClient falls back to the pre-verification behaviour (immediate
// login) rather than issuing a token nobody could ever confirm. SendHTML
// (not Send) so the mailed link renders as a real clickable <a>, not a bare
// URL some mail clients show as plain text.
type Mailer interface {
	SendHTML(ctx context.Context, to, subject, htmlBody string) error
}

// TODO: 2 minuti è un valore di test (per verificare comodamente la scadenza
// del link) — riportare a qualcosa come 24h prima di produzione.
const verificationTTL = 2 * time.Minute

type AuthService struct {
	db         *gorm.DB
	jwtConf    utils.JWTConfig
	mailer     Mailer
	appBaseURL string
}

func NewAuthService(db *gorm.DB, jwtConf utils.JWTConfig, mailer Mailer, appBaseURL string) *AuthService {
	return &AuthService{
		db:         db,
		jwtConf:    jwtConf,
		mailer:     mailer,
		appBaseURL: appBaseURL,
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

	// VerificationToken is only ever non-nil for an account that actually had
	// a confirmation mail issued (RegisterClient, SMTP configured) — accounts
	// created before this feature existed, or created while SMTP was
	// unconfigured, never got one and are never gated here.
	if user.VerificationToken != nil && user.EmailVerifiedAt == nil {
		return nil, utils.NewAPIError(403, "Email non confermata: controlla la posta e clicca il link di conferma")
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
// When req.Role is "cliente" the account is scoped to an EXISTING
// Customer/anagrafica the admin picks (req.CustomerID, required in that
// case) — unlike self-service RegisterClient, which has no existing
// Customer to pick from and creates one instead.
func (s *AuthService) Register(req dto.RegisterRequest) (*dto.AuthUserResponse, error) {
	var customerID *uuid.UUID
	if req.Role == utils.RoleCliente {
		if req.CustomerID == nil || *req.CustomerID == "" {
			return nil, utils.NewAPIError(400, "customer_id obbligatorio per il ruolo cliente")
		}
		id, err := uuid.Parse(*req.CustomerID)
		if err != nil {
			return nil, utils.NewAPIError(400, "customer_id non valido")
		}
		var customer models.Customer
		if err := s.db.First(&customer, "id = ?", id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, utils.NewAPIError(400, "Cliente non trovato")
			}
			return nil, err
		}
		customerID = &id
	}

	var count int64
	s.db.Model(&models.User{}).Where("login = ?", req.Email).Count(&count)
	if count > 0 {
		return nil, utils.NewAPIError(400, "Email già registrata")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	user := models.User{
		Login: req.Email, Name: req.Name, PasswordHash: string(hash), Role: req.Role, Active: true,
		CustomerID: customerID,
		// Admin-provisioned — the admin is already an authenticated,
		// trusted actor, so there's no "prove you own this mailbox" step to
		// gate here (unlike self-service RegisterClient below).
		EmailVerifiedAt: &now,
	}
	if err := s.db.Create(&user).Error; err != nil {
		return nil, err
	}

	resp := toAuthUserResponse(user)
	return &resp, nil
}

// RegisterClient mirrors POST /auth/register-cliente (public, unauthenticated):
// atomically creates a Customer (anagrafica) and a RoleCliente User linked to
// it. When SMTP is configured a confirmation link is mailed and the account
// stays unusable (VerificationToken set, EmailVerifiedAt nil) until
// VerifyEmail confirms it — closing the "anyone can sign up under someone
// else's email and get straight into the portal" gap the previous
// immediate-login behaviour had. Without SMTP there is no way to verify
// anything, so it falls back to that old immediate-login behaviour instead
// of minting a token nobody could ever confirm.
func (s *AuthService) RegisterClient(req dto.ClientRegisterRequest) (*dto.RegisterClientResult, *dto.LoginResult, error) {
	var existing models.User
	err := s.db.Where("login = ?", req.Email).First(&existing).Error
	switch {
	case err == nil && existing.Role == utils.RoleCliente && existing.VerificationToken != nil && existing.EmailVerifiedAt == nil:
		// Genuinely stuck on a previous attempt (mail lost/expired, or SMTP
		// was down at the time) — resend rather than error, so it isn't a
		// permanent dead end. Doesn't touch the Customer row already
		// created on the first attempt. Any OTHER existing account at this
		// email (staff, already-verified client, or one from before this
		// feature existed) falls through to the plain duplicate-email
		// rejection below instead — never mailed a "confirm your account"
		// link to an account that never asked for one.
		return s.resendVerification(existing)
	case err == nil:
		return nil, nil, utils.NewAPIError(400, "Email già registrata")
	case !errors.Is(err, gorm.ErrRecordNotFound):
		return nil, nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, nil, err
	}

	var token string
	var expiresAt *time.Time
	verifying := s.mailer != nil
	if verifying {
		t, err := generateToken()
		if err != nil {
			return nil, nil, err
		}
		token = t
		exp := time.Now().Add(verificationTTL)
		expiresAt = &exp
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
		if verifying {
			user.VerificationToken = &token
			user.VerificationExpiresAt = expiresAt
		} else {
			now := time.Now()
			user.EmailVerifiedAt = &now
		}
		return tx.Create(&user).Error
	})
	if err != nil {
		return nil, nil, err
	}

	if !verifying {
		login, err := s.buildLoginResult(user)
		return nil, login, err
	}

	if err := s.sendVerificationMail(user.Login, user.Name, token); err != nil {
		return nil, nil, utils.NewAPIError(502, "registrazione creata ma invio email fallito: "+err.Error())
	}
	return &dto.RegisterClientResult{
		Verified: false,
		Message:  "Registrazione ricevuta: controlla la tua email per confermare l'account.",
		Email:    user.Login,
	}, nil, nil
}

// resendVerification regenerates and re-mails the confirmation link for an
// existing, not-yet-verified account — see RegisterClient.
func (s *AuthService) resendVerification(user models.User) (*dto.RegisterClientResult, *dto.LoginResult, error) {
	if s.mailer == nil {
		// SMTP got disabled after the account was created unverified —
		// nothing to resend, and Login's gate only applies when a token was
		// actually issued, so clearing it unblocks the account outright.
		now := time.Now()
		s.db.Model(&user).Updates(map[string]interface{}{"email_verified_at": &now, "verification_token": nil, "verification_expires_at": nil})
		return nil, nil, utils.NewAPIError(400, "Email già registrata: account confermato automaticamente (verifica email non disponibile), accedi con la password scelta in fase di registrazione")
	}

	token, err := generateToken()
	if err != nil {
		return nil, nil, err
	}
	exp := time.Now().Add(verificationTTL)
	if err := s.db.Model(&user).Updates(map[string]interface{}{"verification_token": token, "verification_expires_at": &exp}).Error; err != nil {
		return nil, nil, err
	}

	if err := s.sendVerificationMail(user.Login, user.Name, token); err != nil {
		return nil, nil, utils.NewAPIError(502, "invio email fallito: "+err.Error())
	}
	return &dto.RegisterClientResult{
		Verified: false,
		Message:  "Un link di conferma era già stato inviato: te ne abbiamo mandato uno nuovo.",
		Email:    user.Login,
	}, nil, nil
}

// VerifyEmail confirms a registration link's token (POST /auth/verify-email)
// — success behaves exactly like Login, minting a fresh token pair.
func (s *AuthService) VerifyEmail(token string) (*dto.LoginResult, error) {
	var user models.User
	if err := s.db.Where("verification_token = ?", token).First(&user).Error; err != nil {
		return nil, utils.NewAPIError(400, "Link di conferma non valido")
	}
	if user.VerificationExpiresAt == nil || user.VerificationExpiresAt.Before(time.Now()) {
		return nil, utils.NewAPIError(400, "Link di conferma scaduto: registrati di nuovo per ricevere un nuovo link")
	}

	now := time.Now()
	if err := s.db.Model(&user).Updates(map[string]interface{}{
		"email_verified_at": &now, "verification_token": nil, "verification_expires_at": nil,
	}).Error; err != nil {
		return nil, err
	}
	user.EmailVerifiedAt = &now
	user.VerificationToken = nil

	return s.buildLoginResult(user)
}

func (s *AuthService) sendVerificationMail(to, name, token string) error {
	link := fmt.Sprintf("%s/verifica-email?token=%s", s.appBaseURL, token)
	subject := "Confermazione registrazione — Feccia F.lli"
	body := fmt.Sprintf(`<!DOCTYPE html>
<html><body style="font-family:Arial,sans-serif;color:#1a1a1a;line-height:1.5;">
<p>Buongiorno %s,</p>
<p>per confermare la registrazione al portale clienti, clicca il link qui sotto (valido %s):</p>
<p><a href="%s" style="display:inline-block;padding:10px 22px;background:#2a6fdb;color:#fff;text-decoration:none;border-radius:6px;">Confermo il mio account</a></p>
<p style="font-size:13px;color:#666;">Se il bottone non funziona, copia questo indirizzo nel browser:<br><a href="%s">%s</a></p>
<p>Se non hai richiesto questa registrazione, ignora questa email.</p>
<p>Cordiali saluti,<br>Feccia F.lli S.r.l.</p>
</body></html>`,
		html.EscapeString(name), formatTTL(verificationTTL), link, link, link,
	)
	return s.mailer.SendHTML(context.Background(), to, subject, body)
}

// formatTTL renders a duration in Italian for the verification-mail copy —
// verificationTTL is deliberately short during testing, so this must reflect
// whatever it's currently set to rather than a hardcoded "24 ore".
func formatTTL(d time.Duration) string {
	if d < time.Hour {
		return fmt.Sprintf("%d minuti", int(d.Round(time.Minute).Minutes()))
	}
	return fmt.Sprintf("%d ore", int(d.Round(time.Hour).Hours()))
}

func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
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
