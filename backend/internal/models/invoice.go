package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// InvoiceLine is a genuine child table (FK on InvoiceID), mirroring
// InvoiceLineBase's nested-object array in backend/models.py.
type InvoiceLine struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	InvoiceID   uuid.UUID `gorm:"type:uuid;not null;index" json:"-"`
	OrdineID    string    `gorm:"type:varchar(64)" json:"ordine_id"`
	Descrizione string    `gorm:"type:varchar(255)" json:"descrizione"`
	Prodotto    string    `gorm:"type:varchar(255)" json:"prodotto"`
	Peso        float64   `gorm:"not null;default:0" json:"peso"`
	Quantita    float64   `gorm:"not null;default:1" json:"quantita"`
	Tariffa     float64   `gorm:"not null;default:0" json:"tariffa"`
	Totale      float64   `gorm:"not null;default:0" json:"totale"`
	IvaCodice   string    `gorm:"type:varchar(10);default:N8" json:"iva_codice"`
}

// Invoice mirrors backend/routers/invoices.py + InvoiceBase/InvoiceCreate.
// PdfS3Key/PdfUploadedAt/PdfRetainUntil stay permanently nil until PDF
// generation + S3 Object Lock archival are ported (deferred, same as the
// CMR/instructions PDF exports — needs a PDF library decision) — this
// matches Python's own "S3 not configured" dev-mode behavior exactly, not a
// capability regression.
type Invoice struct {
	ID                  uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	Numero              string         `gorm:"type:varchar(30);index" json:"numero"`
	ClienteID           uuid.UUID      `gorm:"type:uuid;not null;index" json:"cliente_id" validate:"required"`
	Cliente             Customer       `gorm:"foreignKey:ClienteID;references:ID" json:"-"`
	DataFattura         string         `gorm:"type:varchar(20)" json:"data_fattura"`
	DataScadenza        string         `gorm:"type:varchar(20)" json:"data_scadenza"`
	CondizioniPagamento string         `gorm:"type:varchar(255)" json:"condizioni_pagamento"`
	Righe               []InvoiceLine  `gorm:"foreignKey:InvoiceID;constraint:OnDelete:CASCADE" json:"righe"`
	CostiAccessori      datatypes.JSON `json:"costi_accessori"`
	TotaleImponibile    float64        `gorm:"not null;default:0" json:"totale_imponibile"`
	TotaleIva           float64        `gorm:"not null;default:0" json:"totale_iva"`
	Totale              float64        `gorm:"not null;default:0" json:"totale"`
	Stato               string         `gorm:"type:varchar(20);not null;default:PROFORMA;index" json:"stato"`
	Tipo                string         `gorm:"type:varchar(20);default:ordine" json:"tipo"`
	Note                string         `gorm:"type:text" json:"note"`
	PdfS3Key            *string        `gorm:"type:varchar(255)" json:"pdf_s3_key"`
	PdfUploadedAt       *string        `gorm:"type:varchar(40)" json:"pdf_uploaded_at"`
	PdfRetainUntil      *string        `gorm:"type:varchar(40)" json:"pdf_retain_until"`

	CreatedAt time.Time `json:"created_at"`
}
