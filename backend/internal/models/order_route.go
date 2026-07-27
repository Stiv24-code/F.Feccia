package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// OrderRoute is the single active road route for an order (garage → carico
// → [via-point...] → scarico → wash_station, any of garage/wash_station
// optional). Only the chosen/edited route is persisted — the 2 discarded
// alternatives from POST /orders/{id}/route-alternatives are never written
// to the DB, they're computed on the fly and shown to the manager to pick
// from (see internal/services/orders.OrderService.RouteAlternatives).
type OrderRoute struct {
	ID      uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	OrderID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex" json:"order_id"`
	// Waypoints is the ordered sequence actually routed:
	// [{"tipo":"garage|destinazione|wash_station","ref_id":"...","nome":"...","lat":..,"lng":..}, ...]
	Waypoints datatypes.JSON `json:"waypoints"`
	// Points is the road geometry ([[lat,lng], ...]) for map rendering.
	Points         datatypes.JSON `json:"points"`
	DistanceKm     float64        `json:"distance_km"`
	DurationMin    int            `json:"duration_min"`
	EditedManually bool           `gorm:"not null;default:false" json:"edited_manually"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
