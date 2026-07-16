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
	ID                        uuid.UUID `gorm:"type:uuid;primaryKey" json:"item_id"`
	PriceListID               uuid.UUID `gorm:"type:uuid;not null;index" json:"-"`
	ProdottoID                string    `gorm:"type:varchar(64)" json:"prodotto_id"`
	ProdottoNome              string    `gorm:"type:varchar(255)" json:"prodotto_nome"`
	DestinazioneCaricoID      string    `gorm:"type:varchar(64)" json:"destinazione_carico_id"`
	DestinazioneCaricoNome    string    `gorm:"type:varchar(255)" json:"destinazione_carico_nome"`
	DestinazioneScaricoID     string    `gorm:"type:varchar(64)" json:"destinazione_scarico_id"`
	DestinazioneScaricoNome   string    `gorm:"type:varchar(255)" json:"destinazione_scarico_nome"`
	Tariffa                   float64   `gorm:"not null;default:0" json:"tariffa"`
	TipoTariffa               string    `gorm:"type:varchar(20);default:forfait" json:"tipo_tariffa"`
	RangePesoMin              float64   `gorm:"not null;default:0" json:"range_peso_min"`
	RangePesoMax              float64   `gorm:"not null;default:0" json:"range_peso_max"`
	UnitaPeso                 string    `gorm:"type:varchar(10);default:Kg" json:"unita_peso"`
	MinimoTassabile           float64   `gorm:"not null;default:0" json:"minimo_tassabile"`
	TipoTrasporto             string    `gorm:"type:varchar(20);default:stradale" json:"tipo_trasporto"`
	PercAdeguamentoCarburante float64   `gorm:"not null;default:0" json:"perc_adeguamento_carburante"`
}

// PriceList mirrors backend/routers/pricelists.py + PriceListBase/PriceListCreate.
type PriceList struct {
	ID          uuid.UUID       `gorm:"type:uuid;primaryKey" json:"id"`
	ClienteID   string          `gorm:"type:varchar(64);not null;index" json:"cliente_id" validate:"required"`
	ClienteNome string          `gorm:"type:varchar(255)" json:"cliente_nome"`
	DataInizio  string          `gorm:"type:varchar(20);index" json:"data_inizio"`
	DataFine    string          `gorm:"type:varchar(20);index" json:"data_fine"`
	Items       []PriceListItem `gorm:"foreignKey:PriceListID;constraint:OnDelete:CASCADE" json:"items"`
	Note        string          `gorm:"type:text" json:"note"`
	InUso       bool            `gorm:"not null;default:false" json:"in_uso"`
	Active      bool            `gorm:"not null;default:true;index" json:"active"`

	CreatedAt time.Time `json:"created_at"`
}
