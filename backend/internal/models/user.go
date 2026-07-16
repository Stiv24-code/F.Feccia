package models

import (
	"time"

	"gorm.io/gorm"
)

// User represents an application user
type User struct {
	ID           int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	Login        string     `gorm:"type:varchar(150);uniqueIndex;not null" json:"login" validate:"required,min=3,max=150"`
	PasswordHash string     `gorm:"type:varchar(255);not null" json:"password_hash" validate:"required"`
	Name         string     `gorm:"type:varchar(255)" json:"name,omitempty" validate:"max=255"`
	Role         string     `gorm:"type:varchar(50);not null;default:admin" json:"role" validate:"required,oneof=admin amministrazione planner operatore"`
	Active       bool       `gorm:"not null;default:true" json:"active"`
	LastLoginAt  *time.Time `json:"last_login_at,omitempty"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty" swaggerignore:"true"`
}

// No TableName() override: like Customer (see customer.go), relies on GORM's
// default naming ("users") so service unit tests can run against SQLite.
// Postgres already defaults new tables to the "public" schema, so behavior
// is unchanged there.
