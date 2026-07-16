package models

import (
	"time"

	"github.com/google/uuid"
)

// Product represents a transportable goods type (anagrafica prodotti).
// Deletion is logical (Active=false), mirroring backend/routers/products.py.
// Codice is unique — mirrors the "codice_unique" index already enforced on
// the Mongo `products` collection (migrations/001_create_indexes.py:37).
type Product struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Codice      string    `gorm:"type:varchar(100);not null;uniqueIndex" json:"codice" validate:"required"`
	Descrizione string    `gorm:"type:varchar(255);not null" json:"descrizione" validate:"required"`
	UnitaMisura string    `gorm:"type:varchar(20);default:Kg" json:"unita_misura"`
	Note        string    `gorm:"type:text" json:"note"`
	Active      bool      `gorm:"not null;default:true;index" json:"active"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
