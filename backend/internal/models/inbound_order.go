package models

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

// InboundOrderStatus is the closed set of values InboundOrder.Status can
// hold — a named type instead of a bare string so the compiler catches typos
// in the acceptance state machine, mirroring OrderStato.
type InboundOrderStatus string

const (
	InboundOrderStatusPending  InboundOrderStatus = "pending"
	InboundOrderStatusAccepted InboundOrderStatus = "accepted"
	InboundOrderStatusModify   InboundOrderStatus = "modify"
)

// InboundOrder source values (where the draft came from).
const (
	InboundOrderSourceSeed   = "seed"
	InboundOrderSourceMail   = "mail"
	InboundOrderSourcePDF    = "pdf"
	InboundOrderSourcePortal = "portal"
)

// InboundOrder is a transport-order draft ingested from the mailbox scraper
// or imported from a client PDF (ported from OrderMesh). It is deliberately
// NOT merged with Order: an inbound order is free text as received from the
// customer (no FKs to Customer/Destination/...), waiting for an operator to
// accept it — converting it into a real models.Order is a separate, explicit
// step.
//
// Dedup rule: one order per (ref, client), case/space-insensitive, so
// re-reading the mailbox never duplicates rows. Enforced by the functional
// unique index inbound_orders_ref_client_key created in
// pkg/database.Migrate — AutoMigrate cannot express expression indexes.
type InboundOrder struct {
	ID            uuid.UUID          `gorm:"type:uuid;primaryKey" json:"id"`
	Client        string             `gorm:"type:varchar(200);not null" json:"client" validate:"required"`
	SenderEmail   string             `gorm:"type:varchar(200);not null;default:''" json:"sender_email"`
	Ref           string             `gorm:"type:varchar(100);not null;default:''" json:"ref"`
	Product       string             `gorm:"type:varchar(200);not null;default:''" json:"product"`
	Kg            int                `gorm:"not null;default:0" json:"kg"`
	LoadDate      string             `gorm:"type:varchar(20)" json:"load_date"`
	LoadPlace     string             `gorm:"type:varchar(200)" json:"load_place"`
	DeliveryDate  string             `gorm:"type:varchar(20)" json:"delivery_date"`
	DeliveryPlace string             `gorm:"type:varchar(200)" json:"delivery_place"`
	Rate          string             `gorm:"type:varchar(50)" json:"rate"`
	Notes         string             `gorm:"type:text" json:"notes"`
	Portal        bool               `gorm:"not null;default:false" json:"portal"`
	Status        InboundOrderStatus `gorm:"type:varchar(20);not null;default:pending;index" json:"status"`
	Source        string             `gorm:"type:varchar(20);not null;default:mail" json:"source"`
	TemplateID    *uuid.UUID         `gorm:"type:uuid" json:"template_id,omitempty"`
	ReceivedAt    time.Time          `json:"received_at"`
	// ClienteID is set only for Source == InboundOrderSourcePortal — the
	// authenticated customer that submitted the request via the self-service
	// portal, used to scope "my pending requests" (GET /me/inbound-orders).
	// Left nil for mail/pdf/seed drafts, which have no such account to tie to.
	ClienteID *uuid.UUID `gorm:"type:uuid;index" json:"cliente_id,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Key identifies an inbound order across scrapes — the in-Go mirror of the
// inbound_orders_ref_client_key unique index, used by the mail scraper to
// skip already-seen orders without hitting the database.
func (o InboundOrder) Key() string {
	return strings.ToLower(strings.TrimSpace(o.Ref) + "|" + strings.TrimSpace(o.Client))
}
