package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// PatenteCategorie is the fixed set of Italian driving-license categories (+
// CQC/ADR professional qualifications) a Driver.Patente value may contain.
// Kept as the single source of truth for both backend validation (dto.go's
// `dive,oneof=...` tag) and the swagger `enums` tag the frontend multiselect
// is generated from.
var PatenteCategorie = []string{
	"AM", "A1", "A2", "A", "B1", "B", "BE",
	"C1", "C1E", "C", "CE", "D1", "D1E", "D", "DE",
	"CQC", "ADR",
}

// Driver represents a truck driver (anagrafica autisti).
// Deletion is logical (Active=false), mirroring backend/routers/drivers.py.
// Patente is a JSON string array (categorie multiple, es. ["CE","ADR"]) —
// same storage idiom as Trip.OrdiniIds, no dedicated child table needed for
// a plain list of fixed-vocabulary values.
type Driver struct {
	ID              uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	Nome            string         `gorm:"type:varchar(150);not null" json:"nome" validate:"required"`
	Cognome         string         `gorm:"type:varchar(150);not null;index:idx_drivers_active_cognome,priority:2" json:"cognome" validate:"required"`
	CodiceFiscale   string         `gorm:"type:varchar(50)" json:"codice_fiscale"`
	Patente         datatypes.JSON `json:"patente"`
	ScadenzaPatente *string        `gorm:"type:varchar(20)" json:"scadenza_patente"`
	Telefono        string         `gorm:"type:varchar(50)" json:"telefono"`
	Email           string         `gorm:"type:varchar(255)" json:"email"`
	Note            string         `gorm:"type:text" json:"note"`
	Active          bool           `gorm:"not null;default:true;index:idx_drivers_active_cognome,priority:1" json:"active"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
