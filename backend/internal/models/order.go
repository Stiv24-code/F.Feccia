package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// OrderItem is a genuine child table (FK on OrderID), mirroring the
// well-structured OrderItemBase array in backend/models.py.
type OrderItem struct {
	ID                  uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	OrderID             uuid.UUID `gorm:"type:uuid;not null;index" json:"order_id"`
	ProdottoID          string    `gorm:"type:varchar(64)" json:"prodotto_id"`
	ProdottoCodice      string    `gorm:"type:varchar(100)" json:"prodotto_codice"`
	ProdottoDescrizione string    `gorm:"type:varchar(255)" json:"prodotto_descrizione"`
	Quantita            float64   `gorm:"not null;default:0" json:"quantita"`
	Peso                float64   `gorm:"not null;default:0" json:"peso"`
}

// Order mirrors backend/routers/orders.py + OrderBase/OrderCreate in
// backend/models.py. Reference fields (ClienteID, DestinazioneCaricoID,
// AutistaID, VettoreID, ViaggioID, FatturaID) are plain strings, not
// uuid.UUID/FK — they mirror Mongo's untyped string references and are
// legitimately empty until the order is planned/invoiced (e.g. AutistaID is
// "" until assign_order runs).
//
// ServiziAccessori ([]string) and CostiAccessori ([]dict) have no fixed
// schema in the Python original (CostiAccessori is List[dict], genuinely
// schema-less and always empty in current usage) — stored as JSON columns
// rather than forcing a fake rigid shape.
type Order struct {
	ID                      uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	Progressivo             string         `gorm:"type:varchar(20);index" json:"progressivo"`
	ClienteID               string         `gorm:"type:varchar(64);not null;index" json:"cliente_id" validate:"required"`
	ClienteNome             string         `gorm:"type:varchar(255)" json:"cliente_nome"`
	DestinazioneCaricoID    string         `gorm:"type:varchar(64)" json:"destinazione_carico_id"`
	DestinazioneCaricoNome  string         `gorm:"type:varchar(255);index" json:"destinazione_carico_nome"`
	DestinazioneScaricoID   string         `gorm:"type:varchar(64)" json:"destinazione_scarico_id"`
	DestinazioneScaricoNome string         `gorm:"type:varchar(255)" json:"destinazione_scarico_nome"`
	DataRitiro              string         `gorm:"type:varchar(20);index" json:"data_ritiro"`
	OraRitiroDa             string         `gorm:"type:varchar(10)" json:"ora_ritiro_da"`
	OraRitiroA              string         `gorm:"type:varchar(10)" json:"ora_ritiro_a"`
	DataConsegna            string         `gorm:"type:varchar(20)" json:"data_consegna"`
	OraConsegnaDa           string         `gorm:"type:varchar(10)" json:"ora_consegna_da"`
	OraConsegnaA            string         `gorm:"type:varchar(10)" json:"ora_consegna_a"`
	Tariffa                 float64        `gorm:"not null;default:0" json:"tariffa"`
	TipoTariffa             string         `gorm:"type:varchar(20);default:forfait" json:"tipo_tariffa"`
	Tipologia               string         `gorm:"type:varchar(20);default:nazionale;index" json:"tipologia"`
	CategoriaTrasporto      string         `gorm:"type:varchar(100)" json:"categoria_trasporto"`
	RifOrdineCliente        string         `gorm:"type:varchar(100)" json:"rif_ordine_cliente"`
	AndataRitorno           bool           `gorm:"not null;default:false" json:"andata_ritorno"`
	Note                    string         `gorm:"type:text" json:"note"`
	Items                   []OrderItem    `gorm:"foreignKey:OrderID;constraint:OnDelete:CASCADE" json:"items"`
	ServiziAccessori        datatypes.JSON `json:"servizi_accessori"`
	CostiAccessori          datatypes.JSON `json:"costi_accessori"`
	Stato                   string         `gorm:"type:varchar(20);not null;default:PIANIFICABILE;index" json:"stato"`
	TargaMotrice            string         `gorm:"type:varchar(20)" json:"targa_motrice"`
	TargaRimorchio          string         `gorm:"type:varchar(20)" json:"targa_rimorchio"`
	AutistaID               string         `gorm:"type:varchar(64)" json:"autista_id"`
	AutistaNome             string         `gorm:"type:varchar(255)" json:"autista_nome"`
	VettoreID               string         `gorm:"type:varchar(64)" json:"vettore_id"`
	VettoreNome             string         `gorm:"type:varchar(255)" json:"vettore_nome"`
	ViaggioID               string         `gorm:"type:varchar(64)" json:"viaggio_id"`
	FatturaID               string         `gorm:"type:varchar(64)" json:"fattura_id"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
