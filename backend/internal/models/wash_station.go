package models

import (
	"time"

	"github.com/google/uuid"
)

// WashStation represents a tank/trailer washing point (punto di lavaggio):
// a location a trip can route through, or start from, but never a pickup or
// discharge point (those live on Destination). Deletion is logical
// (Active=false), mirroring Garage/Destination.
type WashStation struct {
	ID   uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Nome string    `gorm:"type:varchar(255);not null" json:"nome" validate:"required"`
	// Tipo di lavaggio offerto (es. "Lavaggio alimentare EFTCO", "Chimico"),
	// testo libero — non un enum, la varietà tra i punti di lavaggio reali è
	// troppo ampia per un set fisso di valori.
	Tipo      string   `gorm:"type:varchar(150)" json:"tipo"`
	Indirizzo string   `gorm:"type:varchar(255)" json:"indirizzo"`
	Citta     string   `gorm:"type:varchar(150)" json:"citta"`
	Lat       *float64 `gorm:"not null" json:"lat"`
	Lng       *float64 `gorm:"not null" json:"lng"`
	Note      string   `gorm:"type:text" json:"note"`
	Active    bool     `gorm:"not null;default:true;index" json:"active"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
