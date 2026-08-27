package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// OrderItem is a genuine child table (FK on OrderID), mirroring the
// well-structured OrderItemBase array in backend/models.py. ProdottoID is a
// real belongs-to FK — Prodotto is Preloaded by the orders service and
// mapped to a nested dto.ProductResponse, replacing the old denormalized
// ProdottoCodice/ProdottoDescrizione snapshot columns.
type OrderItem struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	OrderID    uuid.UUID `gorm:"type:uuid;not null;index" json:"order_id"`
	ProdottoID uuid.UUID `gorm:"type:uuid;not null" json:"prodotto_id"`
	Prodotto   Product   `gorm:"foreignKey:ProdottoID;references:ID" json:"-"`
	Quantita   float64   `gorm:"not null;default:0" json:"quantita"`
	Peso       float64   `gorm:"not null;default:0" json:"peso"`
}

// OrderStato is the closed set of values Order.Stato can hold — a named
// type instead of a bare string so the compiler catches typos in the state
// machine (services/orders/order.go's Assign/Start/Close/Discard/Delete),
// the one thing that used to only be caught at runtime (or not at all).
type OrderStato string

const (
	OrderStatoPianificabile OrderStato = "PIANIFICABILE"
	OrderStatoPianificato   OrderStato = "PIANIFICATO"
	OrderStatoViaggio       OrderStato = "VIAGGIO"
	OrderStatoChiuso        OrderStato = "CHIUSO"
	OrderStatoScartato      OrderStato = "SCARTATO"
)

// Order mirrors backend/routers/orders.py + OrderBase/OrderCreate in
// backend/models.py. Reference fields are real belongs-to FKs (uuid.UUID,
// or *uuid.UUID for the ones legitimately empty until the order is
// planned/invoiced, e.g. AutistaID is nil until assign_order runs) —
// replacing the old untyped-string-plus-denormalized-name pairs inherited
// from Mongo. Associations are Preloaded by the orders service
// (preloadOrderAssociations) and mapped to nested Response DTOs; there is
// no stored *Nome snapshot column anymore, the name always comes from the
// live associated row.
//
// ViaggioID/FatturaID stay untyped-*uuid.UUID references without an
// association/Preload — nesting the full Trip/Invoice here would create a
// circular DTO shape (Trip embeds its Orders), so the API keeps exposing
// these as flat ids.
//
// ServiziAccessori ([]string) and CostiAccessori ([]dict) have no fixed
// schema in the Python original (CostiAccessori is List[dict], genuinely
// schema-less and always empty in current usage) — stored as JSON columns
// rather than forcing a fake rigid shape.
type Order struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Progressivo string    `gorm:"type:varchar(20);index" json:"progressivo"`
	ClienteID   uuid.UUID `gorm:"type:uuid;not null;index" json:"cliente_id" validate:"required"`
	Cliente     Customer  `gorm:"foreignKey:ClienteID;references:ID" json:"-"`
	// CommittenteID: la parte ordinante quando diversa dal cliente fatturato
	// (maschera legacy "ordine 2017": Cliente BUNGE LODERS, Committente
	// BUNGE LODERS WORMERVEER). Nil = coincide con Cliente.
	CommittenteID         *uuid.UUID   `gorm:"type:uuid" json:"committente_id"`
	Committente           *Customer    `gorm:"foreignKey:CommittenteID;references:ID" json:"-"`
	DestinazioneCaricoID  *uuid.UUID   `gorm:"type:uuid;index" json:"destinazione_carico_id"`
	DestinazioneCarico    *Destination `gorm:"foreignKey:DestinazioneCaricoID;references:ID" json:"-"`
	DestinazioneScaricoID *uuid.UUID   `gorm:"type:uuid" json:"destinazione_scarico_id"`
	DestinazioneScarico   *Destination `gorm:"foreignKey:DestinazioneScaricoID;references:ID" json:"-"`
	// DataRitiro/DataConsegna seguono la larghezza di InboundOrder.LoadDate:
	// Convert ci copia dentro la data del draft cosi' com'e', quindi una
	// colonna piu' stretta a valle sposterebbe soltanto il 22001 dal momento
	// dell'import a quello della conversione. Restano varchar e non date
	// perche' il valore puo' essere una finestra ("22/07/26 15:00 -
	// 23/07/26 04:00 CEST") che nessun parser normalizza senza perdere meta'
	// dell'informazione.
	DataRitiro         string  `gorm:"type:varchar(100);index" json:"data_ritiro"`
	OraRitiroDa        string  `gorm:"type:varchar(10)" json:"ora_ritiro_da"`
	OraRitiroA         string  `gorm:"type:varchar(10)" json:"ora_ritiro_a"`
	DataConsegna       string  `gorm:"type:varchar(100)" json:"data_consegna"`
	OraConsegnaDa      string  `gorm:"type:varchar(10)" json:"ora_consegna_da"`
	OraConsegnaA       string  `gorm:"type:varchar(10)" json:"ora_consegna_a"`
	Tariffa            float64 `gorm:"not null;default:0" json:"tariffa"`
	TipoTariffa        string  `gorm:"type:varchar(20);default:forfait" json:"tipo_tariffa"`
	Tipologia          string  `gorm:"type:varchar(20);default:nazionale;index" json:"tipologia"`
	CategoriaTrasporto string  `gorm:"type:varchar(100)" json:"categoria_trasporto"`
	RifOrdineCliente   string  `gorm:"type:varchar(100)" json:"rif_ordine_cliente"`
	// Riferimento e note per fermata (maschera legacy: "Rif. carico/Note
	// carico" sul mittente, "Rif. scarico/Note scarico" sul destinatario) —
	// distinti da RifOrdineCliente/Note che restano globali sull'ordine.
	RifCarico     string `gorm:"type:varchar(100)" json:"rif_carico"`
	NoteCarico    string `gorm:"type:text" json:"note_carico"`
	RifScarico    string `gorm:"type:varchar(100)" json:"rif_scarico"`
	NoteScarico   string `gorm:"type:text" json:"note_scarico"`
	AndataRitorno bool   `gorm:"not null;default:false" json:"andata_ritorno"`
	// Provvisorio: l'ordine è inserito ma non ancora confermato dal cliente
	// ("Ord. Provvisorio" della maschera legacy) — resta pianificabile ma la
	// UI lo evidenzia; non è uno stato del ciclo di vita.
	Provvisorio      bool           `gorm:"not null;default:false;index" json:"provvisorio"`
	Note             string         `gorm:"type:text" json:"note"`
	Items            []OrderItem    `gorm:"foreignKey:OrderID;constraint:OnDelete:CASCADE" json:"items"`
	ServiziAccessori datatypes.JSON `json:"servizi_accessori"`
	CostiAccessori   datatypes.JSON `json:"costi_accessori"`
	Stato            OrderStato     `gorm:"type:varchar(20);not null;default:PIANIFICABILE;index" json:"stato"`
	GarageID         *uuid.UUID     `gorm:"type:uuid" json:"garage_id"`
	Garage           *Garage        `gorm:"foreignKey:GarageID;references:ID" json:"-"`
	MotriceID        *uuid.UUID     `gorm:"type:uuid" json:"motrice_id"`
	Motrice          *Motrice       `gorm:"foreignKey:MotriceID;references:ID" json:"-"`
	SemirimorchioID  *uuid.UUID     `gorm:"type:uuid" json:"semirimorchio_id"`
	Semirimorchio    *Semirimorchio `gorm:"foreignKey:SemirimorchioID;references:ID" json:"-"`
	AutistaID        *uuid.UUID     `gorm:"type:uuid" json:"autista_id"`
	Autista          *Driver        `gorm:"foreignKey:AutistaID;references:ID" json:"-"`
	VettoreID        *uuid.UUID     `gorm:"type:uuid" json:"vettore_id"`
	Vettore          *Carrier       `gorm:"foreignKey:VettoreID;references:ID" json:"-"`
	WashStationID    *uuid.UUID     `gorm:"type:uuid" json:"wash_station_id"`
	WashStation      *WashStation   `gorm:"foreignKey:WashStationID;references:ID" json:"-"`
	RouteID          *uuid.UUID     `gorm:"type:uuid" json:"route_id"`
	Route            *OrderRoute    `gorm:"foreignKey:RouteID;references:ID" json:"-"`
	ViaggioID        *uuid.UUID     `gorm:"type:uuid" json:"viaggio_id"`
	FatturaID        *uuid.UUID     `gorm:"type:uuid" json:"fattura_id"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
