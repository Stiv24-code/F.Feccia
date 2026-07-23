package models

import (
	"time"

	"github.com/google/uuid"
)

// PriceListItem is a genuine child table (FK on PriceListID), mirroring
// PriceListItemBase's nested-object array in backend/models.py. Its own ID
// doubles as Python's item_id (the identifier used by the item-scoped
// sub-resource endpoints POST/PUT/DELETE /pricelists/{id}/items/{item_id}).
type PriceListItem struct {
	ID                        uuid.UUID    `gorm:"type:uuid;primaryKey" json:"item_id"`
	PriceListID               uuid.UUID    `gorm:"type:uuid;not null;index" json:"-"`
	ProdottoID                *uuid.UUID   `gorm:"type:uuid" json:"prodotto_id"`
	Prodotto                  *Product     `gorm:"foreignKey:ProdottoID;references:ID" json:"-"`
	DestinazioneCaricoID      *uuid.UUID   `gorm:"type:uuid" json:"destinazione_carico_id"`
	DestinazioneCarico        *Destination `gorm:"foreignKey:DestinazioneCaricoID;references:ID" json:"-"`
	DestinazioneScaricoID     *uuid.UUID   `gorm:"type:uuid" json:"destinazione_scarico_id"`
	DestinazioneScarico       *Destination `gorm:"foreignKey:DestinazioneScaricoID;references:ID" json:"-"`
	Tariffa                   float64      `gorm:"not null;default:0" json:"tariffa"`
	TipoTariffa               string       `gorm:"type:varchar(20);default:forfait" json:"tipo_tariffa"`
	RangePesoMin              float64      `gorm:"not null;default:0" json:"range_peso_min"`
	RangePesoMax              float64      `gorm:"not null;default:0" json:"range_peso_max"`
	UnitaPeso                 string       `gorm:"type:varchar(10);default:Kg" json:"unita_peso"`
	MinimoTassabile           float64      `gorm:"not null;default:0" json:"minimo_tassabile"`
	TipoTrasporto             string       `gorm:"type:varchar(20);default:stradale" json:"tipo_trasporto"`
	PercAdeguamentoCarburante float64      `gorm:"not null;default:0" json:"perc_adeguamento_carburante"`
}

// PriceList mirrors backend/routers/pricelists.py + PriceListBase/PriceListCreate.
type PriceList struct {
	ID         uuid.UUID       `gorm:"type:uuid;primaryKey" json:"id"`
	ClienteID  uuid.UUID       `gorm:"type:uuid;not null;index" json:"cliente_id" validate:"required"`
	Cliente    Customer        `gorm:"foreignKey:ClienteID;references:ID" json:"-"`
	DataInizio string          `gorm:"type:varchar(20);index" json:"data_inizio"`
	DataFine   string          `gorm:"type:varchar(20);index" json:"data_fine"`
	Items      []PriceListItem `gorm:"foreignKey:PriceListID;constraint:OnDelete:CASCADE" json:"items"`
	Note       string          `gorm:"type:text" json:"note"`
	InUso      bool            `gorm:"not null;default:false" json:"in_uso"`
	Active     bool            `gorm:"not null;default:true;index" json:"active"`

	CreatedAt time.Time `json:"created_at"`
}
