package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User represents an application user
type User struct {
	ID           int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	Login        string `gorm:"type:varchar(150);uniqueIndex;not null" json:"login" validate:"required,min=3,max=150"`
	PasswordHash string `gorm:"type:varchar(255);not null" json:"password_hash" validate:"required"`
	Name         string `gorm:"type:varchar(255)" json:"name,omitempty" validate:"max=255"`
	Role         string `gorm:"type:varchar(50);not null;default:admin" json:"role" validate:"required,oneof=admin amministrazione planner operatore cliente"`
	// CustomerID links a RoleCliente account to its Customer/anagrafica —
	// always nil for staff roles. Set once, at registration (see
	// AuthService.RegisterClient), never reassigned.
	CustomerID  *uuid.UUID `gorm:"type:uuid;index" json:"customer_id"`
	Active      bool       `gorm:"not null;default:true" json:"active"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`

	// Email verification (self-service client registration only — see
	// AuthService.RegisterClient). VerificationToken is non-nil exactly
	// while a confirmation is outstanding; Login only refuses access when a
	// token was actually issued and never confirmed, so accounts created
	// before this feature existed (VerificationToken always nil) are never
	// retroactively locked out.
	EmailVerifiedAt       *time.Time `json:"email_verified_at,omitempty"`
	VerificationToken     *string    `gorm:"type:varchar(64);uniqueIndex" json:"-"`
	VerificationExpiresAt *time.Time `json:"-"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty" swaggerignore:"true"`
}

// No TableName() override: like Customer (see customer.go), relies on GORM's
// default naming ("users") so service unit tests can run against SQLite.
// Postgres already defaults new tables to the "public" schema, so behavior
// is unchanged there.
