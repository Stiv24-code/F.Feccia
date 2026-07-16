package vehicles

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"fratelli-feccia/internal/dto"
	"fratelli-feccia/internal/models"
	"fratelli-feccia/pkg/utils"
)

const (
	listLimit    = 1000
	gpsHistLimit = 100
	tempLimit    = 200
)

type VehicleService struct {
	db *gorm.DB
}

func NewVehicleService(db *gorm.DB) *VehicleService {
	return &VehicleService{db: db}
}

func escapeLike(term string) string {
	if len(term) > 100 {
		term = term[:100]
	}
	replacer := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return replacer.Replace(term)
}

// ── CRUD ─────────────────────────────────────────────────────────────────

func (s *VehicleService) List(ctx context.Context, search string) ([]dto.VehicleResponse, error) {
	query := s.db.WithContext(ctx).Model(&models.Vehicle{}).Where("active = ?", true)
	if search != "" {
		query = query.Where("LOWER(targa) LIKE ?", "%"+strings.ToLower(escapeLike(search))+"%")
	}
	var vehicles []models.Vehicle
	if err := query.Order("targa ASC").Limit(listLimit).Find(&vehicles).Error; err != nil {
		return nil, err
	}
	result := make([]dto.VehicleResponse, len(vehicles))
	for i, v := range vehicles {
		result[i] = ToResponse(v)
	}
	return result, nil
}

func (s *VehicleService) Create(ctx context.Context, req dto.VehicleRequest) (*dto.VehicleResponse, error) {
	v := models.Vehicle{
		ID:             uuid.New(),
		Targa:          req.Targa,
		TipoVeicolo:    defaultString(req.TipoVeicolo, "motrice"),
		Marca:          req.Marca,
		Modello:        req.Modello,
		Anno:           req.Anno,
		Scompartature:  defaultInt(req.Scompartature, 1),
		PortataKg:      req.PortataKg,
		Note:           req.Note,
		GpsTrackerUrl:  req.GpsTrackerUrl,
		GpsTrackerTipo: req.GpsTrackerTipo,
		GpsApiKey:      req.GpsApiKey,
		Active:         true,
	}
	if err := s.db.WithContext(ctx).Create(&v).Error; err != nil {
		return nil, err
	}
	resp := ToResponse(v)
	return &resp, nil
}

// Update replaces only the "create-able" fields, exactly like Python's
// update_vehicle (VehicleCreate schema) — GPS/temperature telemetry fields
// are never touched here.
func (s *VehicleService) Update(ctx context.Context, id uuid.UUID, req dto.VehicleRequest) (*dto.VehicleResponse, error) {
	var v models.Vehicle
	if err := s.db.WithContext(ctx).First(&v, "id = ?", id).Error; err != nil {
		return nil, err
	}
	v.Targa = req.Targa
	v.TipoVeicolo = defaultString(req.TipoVeicolo, "motrice")
	v.Marca = req.Marca
	v.Modello = req.Modello
	v.Anno = req.Anno
	v.Scompartature = defaultInt(req.Scompartature, 1)
	v.PortataKg = req.PortataKg
	v.Note = req.Note
	v.GpsTrackerUrl = req.GpsTrackerUrl
	v.GpsTrackerTipo = req.GpsTrackerTipo
	v.GpsApiKey = req.GpsApiKey

	if err := s.db.WithContext(ctx).Save(&v).Error; err != nil {
		return nil, err
	}
	resp := ToResponse(v)
	return &resp, nil
}

func (s *VehicleService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.db.WithContext(ctx).Model(&models.Vehicle{}).Where("id = ?", id).Update("active", false).Error
}

// ── GPS ──────────────────────────────────────────────────────────────────

// UpdateGPSByID mirrors POST /vehicles/{vehicle_id}/gps-position: looks up
// by id first, falls back to targa (no case-folding here, matching Python).
func (s *VehicleService) UpdateGPSByID(ctx context.Context, idOrTarga string, req dto.VehicleGPSUpdateRequest) (*dto.GPSUpdateResult, error) {
	var v models.Vehicle
	found := false
	if id, err := uuid.Parse(idOrTarga); err == nil {
		if err := s.db.WithContext(ctx).First(&v, "id = ?", id).Error; err == nil {
			found = true
		}
	}
	if !found {
		if err := s.db.WithContext(ctx).First(&v, "targa = ?", idOrTarga).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, utils.NewAPIError(404, "Veicolo non trovato")
			}
			return nil, err
		}
	}
	return s.applyGPSUpdate(ctx, v, req, "")
}

// UpdateGPSByPlate mirrors POST /vehicles/gps-position-by-plate/{targa}.
func (s *VehicleService) UpdateGPSByPlate(ctx context.Context, targa string, req dto.VehicleGPSUpdateRequest) (*dto.GPSUpdateResult, error) {
	var v models.Vehicle
	if err := s.db.WithContext(ctx).First(&v, "targa = ?", targa).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.NewAPIError(404, "Veicolo con targa "+targa+" non trovato")
		}
		return nil, err
	}
	return s.applyGPSUpdate(ctx, v, req, "")
}

// applyGPSUpdate updates the vehicle's live position + appends a gps_history
// row. gpsSourceOverride is used by the vendor webhook path (source derived
// from the URL vendor); when empty, gps_source is derived from the vehicle's
// own gps_tracker_tipo, matching the JWT-authenticated endpoints in
// backend/routers/vehicles.py.
func (s *VehicleService) applyGPSUpdate(ctx context.Context, v models.Vehicle, req dto.VehicleGPSUpdateRequest, gpsSourceOverride string) (*dto.GPSUpdateResult, error) {
	ts := req.Timestamp
	if ts == "" {
		ts = time.Now().UTC().Format(time.RFC3339)
	}

	gpsSource := gpsSourceOverride
	if gpsSource == "" {
		provider := v.GpsTrackerTipo
		if provider == "" {
			provider = "custom"
		}
		gpsSource = "provider_" + strings.ToLower(strings.TrimSpace(provider))
	}

	v.LastLat = req.Lat
	v.LastLng = req.Lng
	v.LastSpeedKmh = req.SpeedKmh
	v.LastHeading = req.Heading
	v.LastGpsUpdate = ts
	v.GpsActive = true
	v.GpsSource = gpsSource

	if err := s.db.WithContext(ctx).Save(&v).Error; err != nil {
		return nil, err
	}

	history := models.GPSHistoryEntry{
		ID:        uuid.New(),
		VehicleID: v.ID,
		Targa:     v.Targa,
		Lat:       req.Lat,
		Lng:       req.Lng,
		SpeedKmh:  req.SpeedKmh,
		Heading:   req.Heading,
		Timestamp: ts,
		GpsSource: gpsSource,
	}
	if err := s.db.WithContext(ctx).Create(&history).Error; err != nil {
		return nil, err
	}

	return &dto.GPSUpdateResult{
		OK:        true,
		Targa:     v.Targa,
		GpsSource: gpsSource,
		Position:  dto.GPSPositionShort{Lat: req.Lat, Lng: req.Lng},
	}, nil
}

func (s *VehicleService) GetGPSHistory(ctx context.Context, vehicleIDOrTarga string, limit int) ([]dto.GPSHistoryResponse, error) {
	if limit <= 0 {
		limit = gpsHistLimit
	}
	query := s.db.WithContext(ctx).Model(&models.GPSHistoryEntry{})
	if id, err := uuid.Parse(vehicleIDOrTarga); err == nil {
		query = query.Where("vehicle_id = ? OR targa = ?", id, vehicleIDOrTarga)
	} else {
		query = query.Where("targa = ?", vehicleIDOrTarga)
	}
	var entries []models.GPSHistoryEntry
	if err := query.Order("timestamp DESC").Limit(limit).Find(&entries).Error; err != nil {
		return nil, err
	}
	result := make([]dto.GPSHistoryResponse, len(entries))
	for i, e := range entries {
		result[i] = dto.GPSHistoryResponse{
			VehicleID: e.VehicleID, Targa: e.Targa, Lat: e.Lat, Lng: e.Lng,
			SpeedKmh: e.SpeedKmh, Heading: e.Heading, Timestamp: e.Timestamp, GpsSource: e.GpsSource,
		}
	}
	return result, nil
}

func (s *VehicleService) GetAllGPSLive(ctx context.Context) ([]dto.GPSLiveVehicle, error) {
	var vehicles []models.Vehicle
	err := s.db.WithContext(ctx).
		Where("active = ? AND last_lat <> 0", true).
		Limit(200).
		Find(&vehicles).Error
	if err != nil {
		return nil, err
	}
	result := make([]dto.GPSLiveVehicle, len(vehicles))
	for i, v := range vehicles {
		result[i] = dto.GPSLiveVehicle{
			ID: v.ID, Targa: v.Targa, Marca: v.Marca, Modello: v.Modello, TipoVeicolo: v.TipoVeicolo,
			LastLat: v.LastLat, LastLng: v.LastLng, LastSpeedKmh: v.LastSpeedKmh, LastHeading: v.LastHeading,
			LastGpsUpdate: v.LastGpsUpdate, GpsActive: v.GpsActive, GpsTrackerUrl: v.GpsTrackerUrl, GpsSource: v.GpsSource,
		}
	}
	return result, nil
}

// ── Webhooks ─────────────────────────────────────────────────────────────

// IngestGPSWebhook mirrors POST /webhooks/gps/{vendor} (public, token-gated).
func (s *VehicleService) IngestGPSWebhook(ctx context.Context, vendor string, payload dto.GPSWebhookPayload) (*dto.GPSUpdateResult, error) {
	if payload.Targa == "" && payload.VehicleID == "" {
		return nil, utils.NewAPIError(400, "`targa` o `vehicle_id` obbligatorio")
	}
	if payload.Lat < -90 || payload.Lat > 90 {
		return nil, utils.NewAPIError(400, "lat fuori range [-90,90]")
	}
	if payload.Lng < -180 || payload.Lng > 180 {
		return nil, utils.NewAPIError(400, "lng fuori range [-180,180]")
	}

	var v models.Vehicle
	var err error
	if payload.VehicleID != "" {
		err = s.db.WithContext(ctx).First(&v, "id = ?", payload.VehicleID).Error
	} else {
		err = s.db.WithContext(ctx).First(&v, "targa = ?", strings.ToUpper(payload.Targa)).Error
	}
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.NewAPIError(404, "Veicolo non trovato")
		}
		return nil, err
	}

	// gps_webhook.py derives gps_source from the vendor in the URL path, not
	// from the vehicle's own gps_tracker_tipo (unlike the JWT-authenticated
	// GPS endpoints).
	gpsSource := "provider_" + strings.ToLower(strings.TrimSpace(vendor))
	return s.applyGPSUpdate(ctx, v, dto.VehicleGPSUpdateRequest{
		Lat: payload.Lat, Lng: payload.Lng, SpeedKmh: payload.SpeedKmh, Heading: payload.Heading, Timestamp: payload.Timestamp,
	}, gpsSource)
}

// IngestTemperatureWebhook mirrors POST /webhooks/temperature/{vendor}.
func (s *VehicleService) IngestTemperatureWebhook(ctx context.Context, vendor string, payload dto.TemperatureWebhookRequest) (*dto.TemperatureWebhookResult, error) {
	var v models.Vehicle
	var err error
	if payload.VehicleID != "" {
		err = s.db.WithContext(ctx).First(&v, "id = ?", payload.VehicleID).Error
	} else if payload.Targa != "" {
		err = s.db.WithContext(ctx).First(&v, "targa = ?", strings.ToUpper(payload.Targa)).Error
	} else {
		return nil, utils.NewAPIError(404, "Veicolo non trovato")
	}
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.NewAPIError(404, "Veicolo non trovato")
		}
		return nil, err
	}

	ts := payload.Ts
	if ts == "" {
		ts = time.Now().UTC().Format(time.RFC3339)
	}
	outOfRange := isOutOfRange(payload.TempCelsius, v.TempMin, v.TempMax)
	source := "webhook_" + strings.ToLower(vendor)

	reading := models.TemperatureReading{
		ID: uuid.New(), VehicleID: v.ID, Targa: v.Targa, TempCelsius: payload.TempCelsius,
		SensorID: payload.SensorID, Ts: ts, Source: source, OutOfRange: outOfRange, CreatedAt: time.Now().UTC(),
	}
	if err := s.db.WithContext(ctx).Create(&reading).Error; err != nil {
		return nil, err
	}

	v.LastTempCelsius = &payload.TempCelsius
	v.LastTempTs = ts
	v.LastTempSensorID = payload.SensorID
	v.LastTempAlert = outOfRange
	if err := s.db.WithContext(ctx).Save(&v).Error; err != nil {
		return nil, err
	}

	return &dto.TemperatureWebhookResult{OK: true, OutOfRange: outOfRange, Alert: outOfRange}, nil
}

func (s *VehicleService) GetTemperatureHistory(ctx context.Context, vehicleID uuid.UUID, limit int, onlyAlerts bool) ([]dto.TemperatureReadingResponse, error) {
	if limit <= 0 {
		limit = tempLimit
	}
	query := s.db.WithContext(ctx).Model(&models.TemperatureReading{}).Where("vehicle_id = ?", vehicleID)
	if onlyAlerts {
		query = query.Where("out_of_range = ?", true)
	}
	var readings []models.TemperatureReading
	if err := query.Order("ts DESC").Limit(limit).Find(&readings).Error; err != nil {
		return nil, err
	}
	result := make([]dto.TemperatureReadingResponse, len(readings))
	for i, r := range readings {
		result[i] = dto.TemperatureReadingResponse{
			VehicleID: r.VehicleID, Targa: r.Targa, TempCelsius: r.TempCelsius,
			SensorID: r.SensorID, Ts: r.Ts, Source: r.Source, OutOfRange: r.OutOfRange,
		}
	}
	return result, nil
}

func (s *VehicleService) SetTemperatureThresholds(ctx context.Context, vehicleID uuid.UUID, req dto.TemperatureThresholdsRequest) (*dto.TemperatureThresholdsResult, error) {
	if req.TempMin == nil && req.TempMax == nil {
		return nil, utils.NewAPIError(400, "Nessuna soglia fornita")
	}
	if req.TempMin != nil && req.TempMax != nil && *req.TempMin > *req.TempMax {
		return nil, utils.NewAPIError(400, "temp_min deve essere <= temp_max")
	}

	var v models.Vehicle
	if err := s.db.WithContext(ctx).First(&v, "id = ?", vehicleID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.NewAPIError(404, "Veicolo non trovato")
		}
		return nil, err
	}
	if req.TempMin != nil {
		v.TempMin = req.TempMin
	}
	if req.TempMax != nil {
		v.TempMax = req.TempMax
	}
	if err := s.db.WithContext(ctx).Save(&v).Error; err != nil {
		return nil, err
	}

	return &dto.TemperatureThresholdsResult{OK: true, TempMin: v.TempMin, TempMax: v.TempMax}, nil
}

func isOutOfRange(temp float64, vmin, vmax *float64) bool {
	if vmin != nil && temp < *vmin {
		return true
	}
	if vmax != nil && temp > *vmax {
		return true
	}
	return false
}

func defaultString(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func defaultInt(value, fallback int) int {
	if value == 0 {
		return fallback
	}
	return value
}

func ToResponse(v models.Vehicle) dto.VehicleResponse {
	return dto.VehicleResponse{
		ID: v.ID, Targa: v.Targa, TipoVeicolo: v.TipoVeicolo, Marca: v.Marca, Modello: v.Modello,
		Anno: v.Anno, Scompartature: v.Scompartature, PortataKg: v.PortataKg, Note: v.Note,
		GpsTrackerUrl: v.GpsTrackerUrl, GpsTrackerTipo: v.GpsTrackerTipo, GpsApiKey: v.GpsApiKey,
		LastLat: v.LastLat, LastLng: v.LastLng, LastSpeedKmh: v.LastSpeedKmh, LastHeading: v.LastHeading,
		LastGpsUpdate: v.LastGpsUpdate, GpsActive: v.GpsActive, GpsSource: v.GpsSource,
		TempMin: v.TempMin, TempMax: v.TempMax, LastTempCelsius: v.LastTempCelsius,
		LastTempTs: v.LastTempTs, LastTempSensorID: v.LastTempSensorID, LastTempAlert: v.LastTempAlert,
		Active: v.Active, CreatedAt: v.CreatedAt, UpdatedAt: v.UpdatedAt,
	}
}
