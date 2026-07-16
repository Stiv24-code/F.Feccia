package models

import "github.com/google/uuid"

// VehicleType, AccessoryCost and TransportCategory are grouped in one file,
// mirroring backend/routers/masterdata.py's own rationale: each collection
// only has list+create (no update/delete, no created_at), so three separate
// files/packages aren't worth the indirection.

type VehicleType struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Nome        string    `gorm:"type:varchar(150);not null" json:"nome" validate:"required"`
	Descrizione string    `gorm:"type:varchar(255)" json:"descrizione"`
	Active      bool      `gorm:"not null;default:true;index" json:"active"`
}

type AccessoryCost struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Nome         string    `gorm:"type:varchar(150);not null" json:"nome" validate:"required"`
	Descrizione  string    `gorm:"type:varchar(255)" json:"descrizione"`
	CostoDefault float64   `gorm:"not null;default:0" json:"costo_default"`
	Active       bool      `gorm:"not null;default:true;index" json:"active"`
}

type TransportCategory struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Nome        string    `gorm:"type:varchar(150);not null" json:"nome" validate:"required"`
	Descrizione string    `gorm:"type:varchar(255)" json:"descrizione"`
	Active      bool      `gorm:"not null;default:true;index" json:"active"`
}
