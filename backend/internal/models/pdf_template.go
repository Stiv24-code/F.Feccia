package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// InboundOrderFieldTargets are the InboundOrder fields a template zone can
// be mapped to. The keys mirror InboundOrder's JSON names so the template
// editor UI and the PDF import logic speak the same language.
var InboundOrderFieldTargets = []string{
	"client", "sender_email", "ref", "product", "kg",
	"load_date", "load_place", "delivery_date", "delivery_place",
	"rate", "notes",
}

// PdfTemplateField maps a rectangular zone of the PDF onto one InboundOrder
// field. Bounds are normalized 0..1 relative to the page, so the mapping is
// independent from render resolution. This is the element shape of
// PdfTemplate.Fields (stored as a JSON array).
type PdfTemplateField struct {
	ID     string  `json:"id"`
	Target string  `json:"target"` // one of InboundOrderFieldTargets
	Label  string  `json:"label"`  // free label shown in the editor
	Page   int     `json:"page"`   // 0-based
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	W      float64 `json:"w"`
	H      float64 `json:"h"`
}

// PdfTemplate describes how to read one client's PDF order layout (ported
// from OrderMesh). Senders is a JSON array of lowercase addresses or
// "@domain" patterns used to preselect the template from a mail sender;
// the matching happens in Go, not in SQL, so JSON replaces OrderMesh's
// native TEXT[] column. Fields is a JSON array of PdfTemplateField.
type PdfTemplate struct {
	ID      uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	Name    string         `gorm:"type:varchar(200);not null" json:"name" validate:"required"`
	Client  string         `gorm:"type:varchar(200);not null;default:''" json:"client"`
	Senders datatypes.JSON `json:"senders"`
	Fields  datatypes.JSON `json:"fields"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
