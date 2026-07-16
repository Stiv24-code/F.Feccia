package models

import (
	"time"

	"github.com/google/uuid"
)

// Customer represents a customer master-data record (anagrafica clienti).
// Deletion is logical (Active=false), mirroring the Python/Mongo behavior —
// there is no gorm.DeletedAt soft-delete here.
type Customer struct {
	ID                  uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	RagioneSociale      string    `gorm:"type:varchar(255);not null;index:idx_customers_active_ragione,priority:2" json:"ragione_sociale" validate:"required"`
	Indirizzo           string    `gorm:"type:varchar(255)" json:"indirizzo"`
	Citta               string    `gorm:"type:varchar(150)" json:"citta"`
	Cap                 string    `gorm:"type:varchar(20)" json:"cap"`
	Provincia           string    `gorm:"type:varchar(50)" json:"provincia"`
	Nazione             string    `gorm:"type:varchar(100);default:Italia" json:"nazione"`
	PartitaIva          string    `gorm:"type:varchar(50)" json:"partita_iva"`
	CodiceFiscale       string    `gorm:"type:varchar(50)" json:"codice_fiscale"`
	Telefono            string    `gorm:"type:varchar(50)" json:"telefono"`
	Email               string    `gorm:"type:varchar(255)" json:"email"`
	Pec                 string    `gorm:"type:varchar(255)" json:"pec"`
	CondizioniPagamento string    `gorm:"type:varchar(255)" json:"condizioni_pagamento"`
	Note                string    `gorm:"type:text" json:"note"`
	RichiedeRifOrdine   bool      `gorm:"not null;default:false" json:"richiede_rif_ordine"`
	Active              bool      `gorm:"not null;default:true;index:idx_customers_active_ragione,priority:1" json:"active"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// No TableName() override: relies on GORM's default naming ("customers") so
// service unit tests can run against SQLite, whose query planner doesn't
// resolve a "public" schema qualifier. Postgres already defaults new tables
// to the "public" schema, so behavior is unchanged there.
