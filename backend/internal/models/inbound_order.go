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
// customer (client/places/product are strings, not anagrafica FKs), waiting
// for an operator to accept it — converting it into a real models.Order is a
// separate, explicit step (services/inboundorders.Convert, POST
// /inbound-orders/:id/convert), tracked by OrderID.
//
// All FKs are ON DELETE SET NULL: the provenance links (TemplateID,
// ClienteID), the structured portal payload (CommittenteID, Destinazione*ID)
// and the conversion link (OrderID). Losing a referenced row must never
// delete the received order or rewrite what the customer actually asked for.
//
// Dedup rule: one order per (ref, client), case/space-insensitive, so
// re-reading the mailbox never duplicates rows. Enforced by the functional
// unique index inbound_orders_ref_client_key created in
// pkg/database.Migrate — AutoMigrate cannot express expression indexes.
type InboundOrder struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Client      string    `gorm:"type:varchar(200);not null" json:"client" validate:"required"`
	SenderEmail string    `gorm:"type:varchar(200);not null;default:''" json:"sender_email"`
	Ref         string    `gorm:"type:varchar(100);not null;default:''" json:"ref"`
	// Product/LoadPlace/DeliveryPlace/Rate sono text e non varchar(N): sono
	// testo estratto da un documento arbitrario, senza un limite naturale da
	// dichiarare, e in Postgres text e varchar(N) hanno identica
	// rappresentazione — il limite non fa risparmiare spazio, produce solo un
	// 22001 su un ordine legittimo. Restano varchar i campi con un vincolo
	// vero: Ref e Client, che compongono l'indice unico di dedup (una btree
	// ha un tetto alla dimensione della riga indicizzata, quindi li' il
	// limite protegge invece di inciampare), e SenderEmail, che un indirizzo
	// RFC-valido non supera comunque.
	Product string `gorm:"type:text;not null;default:''" json:"product"`
	Kg      int    `gorm:"not null;default:0" json:"kg"`
	// LoadDate/DeliveryDate sono testo libero come il resto del draft, non
	// date normalizzate: quello che l'estrazione tira fuori da un PDF reale
	// e' spesso una finestra con fuso ("22/07/26 15:00 - 23/07/26 04:00
	// CEST", 36 caratteri). I varchar(20) originali erano dimensionati su un
	// "2026-08-10" e facevano fallire l'insert con 22001 sul primo ordine
	// vero — vedi il caso Bunge Loders. Restano stringhe, non date: e'
	// Convert a decidere che farne quando nasce l'ordine.
	LoadDate      string             `gorm:"type:varchar(100)" json:"load_date"`
	LoadPlace     string             `gorm:"type:text" json:"load_place"`
	DeliveryDate  string             `gorm:"type:varchar(100)" json:"delivery_date"`
	DeliveryPlace string             `gorm:"type:text" json:"delivery_place"`
	Rate          string             `gorm:"type:text" json:"rate"`
	Notes         string             `gorm:"type:text" json:"notes"`
	Portal        bool               `gorm:"not null;default:false" json:"portal"`
	Status        InboundOrderStatus `gorm:"type:varchar(20);not null;default:pending;index" json:"status"`
	Source        string             `gorm:"type:varchar(20);not null;default:mail" json:"source"`
	// TemplateID: il template PDF usato per l'import. L'associazione esiste
	// solo per fare generare ad AutoMigrate la FK (ON DELETE SET NULL:
	// eliminare un template non deve toccare gli ordini già importati) —
	// nessun servizio la Preload-a.
	TemplateID *uuid.UUID   `gorm:"type:uuid" json:"template_id,omitempty"`
	Template   *PdfTemplate `gorm:"foreignKey:TemplateID;references:ID;constraint:OnDelete:SET NULL" json:"-"`
	ReceivedAt time.Time    `json:"received_at"`
	// ClienteID is set only for Source == InboundOrderSourcePortal — the
	// authenticated customer that submitted the request via the self-service
	// portal, used to scope "my pending requests" (GET /me/inbound-orders).
	// Left nil for mail/pdf/seed drafts, which have no such account to tie
	// to. Association declared only for the FK (ON DELETE SET NULL), never
	// Preload-ed.
	ClienteID *uuid.UUID `gorm:"type:uuid;index" json:"cliente_id,omitempty"`
	Cliente   *Customer  `gorm:"foreignKey:ClienteID;references:ID;constraint:OnDelete:SET NULL" json:"-"`

	// --- Payload strutturato, valorizzato solo per Source == portal ---
	//
	// Il portale riceve una richiesta gia' in forma di Order (il cliente
	// sceglie le destinazioni da una lista, quindi gli UUID sono validi e
	// autenticati) ma la deve archiviare come draft. Prima questi id
	// venivano risolti in nome e buttati, e Convert non aveva modo di
	// ricostruirli: gli unici dati residui erano LoadPlace/DeliveryPlace,
	// stringhe non risolvibili in FK senza un match sul nome — che per i
	// draft mail/pdf sarebbe spoofabile. Preservarli qui rende la
	// conversione di un draft da portale esatta e senza inferenze.
	//
	// Colonne dedicate e non un unico blob JSON perche' cosi' le FK sono
	// vere: ON DELETE SET NULL azzera l'id se l'anagrafica sparisce fra
	// l'invio e l'accettazione, invece di lasciare nel JSON un UUID morto
	// che Convert scoprirebbe solo al Create dell'ordine.
	CommittenteID         *uuid.UUID   `gorm:"type:uuid" json:"committente_id,omitempty"`
	Committente           *Customer    `gorm:"foreignKey:CommittenteID;references:ID;constraint:OnDelete:SET NULL" json:"-"`
	DestinazioneCaricoID  *uuid.UUID   `gorm:"type:uuid" json:"destinazione_carico_id,omitempty"`
	DestinazioneCarico    *Destination `gorm:"foreignKey:DestinazioneCaricoID;references:ID;constraint:OnDelete:SET NULL" json:"-"`
	DestinazioneScaricoID *uuid.UUID   `gorm:"type:uuid" json:"destinazione_scarico_id,omitempty"`
	DestinazioneScarico   *Destination `gorm:"foreignKey:DestinazioneScaricoID;references:ID;constraint:OnDelete:SET NULL" json:"-"`
	// Orari: duplicano l'omonima riga in Notes ("Orario ritiro: 08:00-12:00"),
	// che resta perche' e' quella che la dashboard di accettazione rende.
	// Qui servono strutturati per poterli passare all'Order senza riparsare
	// una nota scritta per essere letta da un umano.
	OraRitiroDa   string `gorm:"type:varchar(10)" json:"ora_ritiro_da"`
	OraRitiroA    string `gorm:"type:varchar(10)" json:"ora_ritiro_a"`
	OraConsegnaDa string `gorm:"type:varchar(10)" json:"ora_consegna_da"`
	OraConsegnaA  string `gorm:"type:varchar(10)" json:"ora_consegna_a"`
	// TariffaProposta e' la "Tariffa desiderata" del form portale: una
	// proposta del cliente, MAI un prezzo. Sta in un campo distinto da
	// Order.Tariffa proprio per non poter finire per sbaglio in un ordine
	// come importo pattuito — Convert la usa solo come default di un valore
	// che l'operatore ha davanti agli occhi quando converte. Rate (stringa
	// formattata) resta per la dashboard e per i draft mail/pdf, dove la
	// tariffa e' testo libero non parsabile.
	TariffaProposta float64 `gorm:"not null;default:0" json:"tariffa_proposta"`

	// OrderID e' l'unico legame draft -> Order: nil finche' il draft non e'
	// stato convertito. Rende Convert idempotente (secondo tentativo = 409
	// con l'id gia' creato, non un ordine gemello) e permette al portale di
	// mostrare al cliente dove e' finita la sua richiesta. ON DELETE SET
	// NULL: cancellare l'ordine non cancella il draft, lo rende
	// riconvertibile.
	OrderID *uuid.UUID `gorm:"type:uuid;index" json:"order_id,omitempty"`
	Order   *Order     `gorm:"foreignKey:OrderID;references:ID;constraint:OnDelete:SET NULL" json:"-"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Key identifies an inbound order across scrapes — the in-Go mirror of the
// inbound_orders_ref_client_key unique index, used by the mail scraper to
// skip already-seen orders without hitting the database.
func (o InboundOrder) Key() string {
	return strings.ToLower(strings.TrimSpace(o.Ref) + "|" + strings.TrimSpace(o.Client))
}
