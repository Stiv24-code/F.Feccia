package models

import (
	"time"

	"github.com/google/uuid"
)

// Destination represents a loading/unloading location (anagrafica destinazioni).
// Deletion is logical (Active=false), mirroring backend/routers/destinations.py.
type Destination struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Nome           string    `gorm:"type:varchar(255);not null;index:idx_destinations_active_nome,priority:2" json:"nome" validate:"required"`
	Indirizzo      string    `gorm:"type:varchar(255)" json:"indirizzo"`
	Citta          string    `gorm:"type:varchar(150)" json:"citta"`
	Cap            string    `gorm:"type:varchar(20)" json:"cap"`
	Provincia      string    `gorm:"type:varchar(50)" json:"provincia"`
	Nazione        string    `gorm:"type:varchar(100);default:Italia" json:"nazione"`
	VincoliScarico string    `gorm:"type:text" json:"vincoli_scarico"`
	Note           string    `gorm:"type:text" json:"note"`
	Active         bool      `gorm:"not null;default:true;index:idx_destinations_active_nome,priority:1" json:"active"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
