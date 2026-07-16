package models

import (
	"time"

	"github.com/google/uuid"
)

// DriverUnavailability represents a driver leave/sick/permit period.
// Unlike the master-data entities, deletion here is a real hard delete
// (no `active` flag), mirroring backend/routers/driver_unavailability.py.
// AutistaID has no DB-level FK constraint on drivers.id — the Mongo original
// enforces no referential integrity either.
type DriverUnavailability struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	AutistaID   uuid.UUID `gorm:"type:uuid;not null;index" json:"autista_id" validate:"required"`
	AutistaNome string    `gorm:"type:varchar(255)" json:"autista_nome"`
	DataDa      string    `gorm:"type:varchar(20);not null" json:"data_da" validate:"required"`
	DataA       string    `gorm:"type:varchar(20);not null" json:"data_a" validate:"required"`
	Motivo      string    `gorm:"type:varchar(50);default:ferie" json:"motivo"`
	Note        string    `gorm:"type:text" json:"note"`

	CreatedAt time.Time `json:"created_at"`
}
