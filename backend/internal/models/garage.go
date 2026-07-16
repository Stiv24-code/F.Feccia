package models

import (
	"time"

	"github.com/google/uuid"
)

// Garage represents a depot / start-end point for trips (garage/deposito).
// Deletion is logical (Active=false), mirroring backend/routers/garages.py.
type Garage struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Nome      string    `gorm:"type:varchar(255);not null" json:"nome" validate:"required"`
	Indirizzo string    `gorm:"type:varchar(255)" json:"indirizzo"`
	Citta     string    `gorm:"type:varchar(150)" json:"citta"`
	Note      string    `gorm:"type:text" json:"note"`
	Active    bool      `gorm:"not null;default:true;index" json:"active"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
