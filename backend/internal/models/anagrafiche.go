package models

import (
	"time"

	"github.com/google/uuid"
)

// Country, Bank and AccountingEntry are grouped in one file, mirroring
// backend/routers/anagrafiche_extra.py's own rationale: three small CRUDs
// with the same shape, kept together for discoverability.

// Country has no DB-level unique constraint on Iso2 in the Mongo original
// either — uniqueness is enforced at the application layer (check-then-insert
// in the service), not a schema constraint. See CustomerService for the same
// reasoning applied to partita_iva.
type Country struct {
	ID     uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Iso2   string    `gorm:"type:varchar(2);not null" json:"iso2" validate:"required"`
	Iso3   string    `gorm:"type:varchar(3)" json:"iso3"`
	Nome   string    `gorm:"type:varchar(150);not null;index:idx_countries_active_nome,priority:2" json:"nome" validate:"required"`
	Eu     bool      `gorm:"not null;default:false" json:"eu"`
	Valuta string    `gorm:"type:varchar(10)" json:"valuta"`
	Active bool      `gorm:"not null;default:true;index:idx_countries_active_nome,priority:1" json:"active"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Bank struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Nome       string    `gorm:"type:varchar(255);not null;index:idx_banks_active_nome,priority:2" json:"nome" validate:"required"`
	BicSwift   string    `gorm:"type:varchar(20)" json:"bic_swift"`
	IbanPrefix string    `gorm:"type:varchar(10)" json:"iban_prefix"`
	Indirizzo  string    `gorm:"type:varchar(255)" json:"indirizzo"`
	Citta      string    `gorm:"type:varchar(150)" json:"citta"`
	Note       string    `gorm:"type:text" json:"note"`
	Active     bool      `gorm:"not null;default:true;index:idx_banks_active_nome,priority:1" json:"active"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// AccountingEntry.Codice has no DB-level unique constraint either — same
// check-then-insert pattern as Country.Iso2 (mirrors the Python original).
type AccountingEntry struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Codice         string    `gorm:"type:varchar(50);not null;index:idx_accounting_entries_active_codice,priority:2" json:"codice" validate:"required"`
	Descrizione    string    `gorm:"type:varchar(255);not null" json:"descrizione" validate:"required"`
	Tipo           string    `gorm:"type:varchar(20);not null;default:ricavo" json:"tipo"`
	ContoContabile string    `gorm:"type:varchar(50)" json:"conto_contabile"`
	IvaCodice      string    `gorm:"type:varchar(10);default:N8" json:"iva_codice"`
	Active         bool      `gorm:"not null;default:true;index:idx_accounting_entries_active_codice,priority:1" json:"active"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
