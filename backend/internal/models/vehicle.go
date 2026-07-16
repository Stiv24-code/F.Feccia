package models

import (
	"time"

	"github.com/google/uuid"
)

// Vehicle mirrors backend/routers/vehicles.py + VehicleBase in
// backend/models.py. Telemetry fields (Last*, Gps*, Temp*) are written by
// separate GPS/temperature endpoints (RBAC admin+planner), not by the
// CRUD form (RBAC admin+operatore) — same document, different write-paths,
// exactly like the Python original.
type Vehicle struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Targa          string    `gorm:"type:varchar(20);not null;uniqueIndex" json:"targa" validate:"required"`
	TipoVeicolo    string    `gorm:"type:varchar(20);default:motrice" json:"tipo_veicolo"`
	Marca          string    `gorm:"type:varchar(100)" json:"marca"`
	Modello        string    `gorm:"type:varchar(100)" json:"modello"`
	Anno           int       `gorm:"not null;default:0" json:"anno"`
	Scompartature  int       `gorm:"not null;default:1" json:"scompartature"`
	PortataKg      float64   `gorm:"not null;default:0" json:"portata_kg"`
	Note           string    `gorm:"type:text" json:"note"`
	GpsTrackerUrl  string    `gorm:"type:varchar(255)" json:"gps_tracker_url"`
	GpsTrackerTipo string    `gorm:"type:varchar(50)" json:"gps_tracker_tipo"`
	GpsApiKey      string    `gorm:"type:varchar(255)" json:"gps_api_key"`

	LastLat       float64 `gorm:"not null;default:0" json:"last_lat"`
	LastLng       float64 `gorm:"not null;default:0" json:"last_lng"`
	LastSpeedKmh  float64 `gorm:"not null;default:0" json:"last_speed_kmh"`
	LastHeading   float64 `gorm:"not null;default:0" json:"last_heading"`
	LastGpsUpdate string  `gorm:"type:varchar(40)" json:"last_gps_update"`
	GpsActive     bool    `gorm:"not null;default:false" json:"gps_active"`
	GpsSource     string  `gorm:"type:varchar(50)" json:"gps_source"`

	TempMin          *float64 `json:"temp_min"`
	TempMax          *float64 `json:"temp_max"`
	LastTempCelsius  *float64 `json:"last_temp_celsius"`
	LastTempTs       string   `gorm:"type:varchar(40)" json:"last_temp_ts"`
	LastTempSensorID string   `gorm:"type:varchar(100)" json:"last_temp_sensor_id"`
	LastTempAlert    bool     `gorm:"not null;default:false" json:"last_temp_alert"`

	Active bool `gorm:"not null;default:true;index:idx_vehicles_active_targa,priority:1" json:"active"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// GPSHistoryEntry is an append-only history row, mirroring Mongo's
// `gps_history` collection.
type GPSHistoryEntry struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	VehicleID uuid.UUID `gorm:"type:uuid;not null;index" json:"vehicle_id"`
	Targa     string    `gorm:"type:varchar(20);index" json:"targa"`
	Lat       float64   `json:"lat"`
	Lng       float64   `json:"lng"`
	SpeedKmh  float64   `json:"speed_kmh"`
	Heading   float64   `json:"heading"`
	Timestamp string    `gorm:"type:varchar(40);index" json:"timestamp"`
	GpsSource string    `gorm:"type:varchar(50)" json:"gps_source"`
}

// TemperatureReading mirrors Mongo's `temperature_readings` collection.
type TemperatureReading struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	VehicleID   uuid.UUID `gorm:"type:uuid;not null;index" json:"vehicle_id"`
	Targa       string    `gorm:"type:varchar(20)" json:"targa"`
	TempCelsius float64   `json:"temp_celsius"`
	SensorID    string    `gorm:"type:varchar(100)" json:"sensor_id"`
	Ts          string    `gorm:"type:varchar(40);index" json:"ts"`
	Source      string    `gorm:"type:varchar(50)" json:"source"`
	OutOfRange  bool      `gorm:"not null;default:false;index" json:"out_of_range"`

	CreatedAt time.Time `json:"-"`
}
