package models

import (
	"time"

	"github.com/google/uuid"
)

// Semirimorchio is a semi-trailer — the other half of the tractor+trailer
// pair a Trip/Order can be assigned to (see Motrice).
type Semirimorchio struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Targa         string    `gorm:"type:varchar(20);not null;uniqueIndex" json:"targa" validate:"required"`
	Tipo          string    `gorm:"type:varchar(50)" json:"tipo"`
	Scompartature int       `gorm:"not null;default:1" json:"scompartature"`
	PortataKg     float64   `gorm:"not null;default:0" json:"portata_kg"`
	Note          string    `gorm:"type:text" json:"note"`
	Active        bool      `gorm:"not null;default:true;index:idx_semirimorchi_active_targa,priority:1" json:"active"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
