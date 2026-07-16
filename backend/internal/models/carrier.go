package models

import (
	"time"

	"github.com/google/uuid"
)

// Carrier represents a subcontracted haulier (anagrafica vettori).
// Deletion is logical (Active=false), mirroring backend/routers/carriers.py.
type Carrier struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	RagioneSociale string    `gorm:"type:varchar(255);not null;index:idx_carriers_active_ragione,priority:2" json:"ragione_sociale" validate:"required"`
	PartitaIva     string    `gorm:"type:varchar(50)" json:"partita_iva"`
	Indirizzo      string    `gorm:"type:varchar(255)" json:"indirizzo"`
	Citta          string    `gorm:"type:varchar(150)" json:"citta"`
	Telefono       string    `gorm:"type:varchar(50)" json:"telefono"`
	Email          string    `gorm:"type:varchar(255)" json:"email"`
	Note           string    `gorm:"type:text" json:"note"`
	Active         bool      `gorm:"not null;default:true;index:idx_carriers_active_ragione,priority:1" json:"active"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
