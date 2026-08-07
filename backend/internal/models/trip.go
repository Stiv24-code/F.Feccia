package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// TripSegment is a genuine child table (FK on TripID), mirroring
// TripSegmentBase's nested-object array in backend/models.py — unlike
// Order's servizi_accessori/costi_accessori, segments have a real fixed
// schema, so they get real columns, not a JSON blob.
type TripSegment struct {
	ID               uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	TripID           uuid.UUID `gorm:"type:uuid;not null;index" json:"trip_id"`
	Ordine           int       `json:"ordine"`
	Tipo             string    `gorm:"type:varchar(50)" json:"tipo"`
	OrigineNome      string    `gorm:"type:varchar(255)" json:"origine_nome"`
	OrigineLat       float64   `json:"origine_lat"`
	OrigineLng       float64   `json:"origine_lng"`
	DestinazioneNome string    `gorm:"type:varchar(255)" json:"destinazione_nome"`
	DestinazioneLat  float64   `json:"destinazione_lat"`
	DestinazioneLng  float64   `json:"destinazione_lng"`
	Km               float64   `json:"km"`
	TempoStimatoMin  int       `json:"tempo_stimato_min"`
	OrdineID         *string   `gorm:"type:varchar(64)" json:"ordine_id"`
}

// Trip mirrors backend/routers/trips.py + TripBase/TripCreate in
// backend/models.py. Reference fields (Motrice/Semirimorchio, AutistaID,
// VettoreID, GarageID) are real belongs-to FKs (*uuid.UUID, nil until
// assigned) — see Order for the same reasoning/history (these used to be
// untyped Mongo-style strings). Associations are Preloaded by the trips
// service and mapped to nested Response DTOs; there is no stored *Nome
// snapshot column anymore.
//
// OrdiniIds is a JSON string array — Mongo has no child-table equivalent for
// a plain id list, and turning it into a join table would add join overhead
// for no query benefit over the orders.viaggio_id back-reference already
// used to fetch a trip's orders.
type Trip struct {
	ID              uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	OrdiniIds       datatypes.JSON `json:"ordini_ids"`
	MotriceID       *uuid.UUID     `gorm:"type:uuid" json:"motrice_id"`
	Motrice         *Motrice       `gorm:"foreignKey:MotriceID;references:ID" json:"-"`
	SemirimorchioID *uuid.UUID     `gorm:"type:uuid" json:"semirimorchio_id"`
	Semirimorchio   *Semirimorchio `gorm:"foreignKey:SemirimorchioID;references:ID" json:"-"`
	AutistaID       *uuid.UUID     `gorm:"type:uuid" json:"autista_id"`
	Autista         *Driver        `gorm:"foreignKey:AutistaID;references:ID" json:"-"`
	VettoreID       *uuid.UUID     `gorm:"type:uuid" json:"vettore_id"`
	Vettore         *Carrier       `gorm:"foreignKey:VettoreID;references:ID" json:"-"`
	GarageID        *uuid.UUID     `gorm:"type:uuid" json:"garage_id"`
	Garage          *Garage        `gorm:"foreignKey:GarageID;references:ID" json:"-"`
	Segments        []TripSegment  `gorm:"foreignKey:TripID;constraint:OnDelete:CASCADE" json:"segmenti"`
	KmTotali        float64        `gorm:"not null;default:0" json:"km_totali"`
	CostoStimato    float64        `gorm:"not null;default:0" json:"costo_stimato"`
	Stato           string         `gorm:"type:varchar(20);not null;default:IN_CORSO;index" json:"stato"`
	Note            string         `gorm:"type:text" json:"note"`
	DataPartenza    string         `gorm:"type:varchar(20)" json:"data_partenza"`
	DataArrivo      string         `gorm:"type:varchar(20)" json:"data_arrivo"`

	CreatedAt time.Time `json:"created_at"`
}

// RouteCache mirrors Mongo's `route_cache` collection: OSRM route results
// cached by rounded from/to coordinate pair, so repeated segment
// (re)computations don't re-hit the OSRM demo server.
type RouteCache struct {
	Key           string         `gorm:"type:varchar(64);primaryKey" json:"key"`
	Points        datatypes.JSON `json:"points"`
	DistanceKm    float64        `json:"distance_km"`
	DurationHours float64        `json:"duration_hours"`
	NumPoints     int            `json:"num_points"`
}
