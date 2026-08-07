package models

import (
	"time"

	"github.com/google/uuid"
)

// Motrice is a tractor unit (trattore/motrice) — one half of the
// tractor+trailer pair a Trip/Order can be assigned to, split out of the
// former single `Vehicle` table so each half has its own anagrafica.
type Motrice struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Targa     string    `gorm:"type:varchar(20);not null;uniqueIndex" json:"targa" validate:"required"`
	Marca     string    `gorm:"type:varchar(100)" json:"marca"`
	Modello   string    `gorm:"type:varchar(100)" json:"modello"`
	Anno      int       `gorm:"not null;default:0" json:"anno"`
	PortataKg float64   `gorm:"not null;default:0" json:"portata_kg"`
	Note      string    `gorm:"type:text" json:"note"`
	Active    bool      `gorm:"not null;default:true;index:idx_motrici_active_targa,priority:1" json:"active"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
