package models

import (
	"time"

	"github.com/google/uuid"
)

// Driver represents a truck driver (anagrafica autisti).
// Deletion is logical (Active=false), mirroring backend/routers/drivers.py.
type Driver struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Nome            string    `gorm:"type:varchar(150);not null" json:"nome" validate:"required"`
	Cognome         string    `gorm:"type:varchar(150);not null;index:idx_drivers_active_cognome,priority:2" json:"cognome" validate:"required"`
	CodiceFiscale   string    `gorm:"type:varchar(50)" json:"codice_fiscale"`
	Patente         string    `gorm:"type:varchar(50)" json:"patente"`
	ScadenzaPatente *string   `gorm:"type:varchar(20)" json:"scadenza_patente"`
	Telefono        string    `gorm:"type:varchar(50)" json:"telefono"`
	Email           string    `gorm:"type:varchar(255)" json:"email"`
	Note            string    `gorm:"type:text" json:"note"`
	Active          bool      `gorm:"not null;default:true;index:idx_drivers_active_cognome,priority:1" json:"active"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
