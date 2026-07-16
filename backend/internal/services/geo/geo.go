// Package geo ports backend/services.py's OSRM + coordinate-lookup helpers
// (used by both trips' segment computation and the live map). Shared here
// rather than duplicated so both packages hit the same route cache.
package geo

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"fratelli-feccia/internal/models"
)

// DestinationCoords/GarageCoords mirror backend/services.py's hardcoded
// DESTINATION_COORDS/GARAGE_COORDS fallback tables. Destinations/garages in
// this port (like the Python original in practice) have no lat/lng columns
// of their own, so this table-based lookup is the only resolution path —
// not a fallback for a DB lookup that never actually succeeds.
var DestinationCoords = map[string][2]float64{
	"Laives (BZ)":                 {46.4283, 11.3394},
	"Calvörde (DE)":               {52.3833, 11.3000},
	"Bruxelles (BE)":              {50.8503, 4.3517},
	"Zurigo (CH)":                 {47.3769, 8.5417},
	"Milano (MI)":                 {45.4642, 9.1900},
	"Lodi (LO)":                   {45.3138, 9.5032},
	"Cuneo (CN)":                  {44.3845, 7.5427},
	"Oosterhout (NL)":             {51.6439, 4.8600},
	"Saint Denis de l'Hôtel (FR)": {47.9167, 2.1333},
	"Rinteln (DE)":                {52.1867, 9.0794},
	"Zeebrugge (BE)":              {51.3333, 3.1833},
	"Pevestorf (DE)":              {53.0500, 11.4333},
}

var GarageCoords = map[string][2]float64{
	"Garage Feccia F.lli - Lodi": {45.3138, 9.5032},
	"Deposito Milano":            {45.4642, 9.1900},
}

const DefaultGarageName = "Garage Feccia F.lli - Lodi"

type NamedPoint struct {
	Name string
	Lat  float64
	Lng  float64
}

// GetCoordsForDestination: exact then case-insensitive substring match,
// mirroring get_coords_for_destination.
func GetCoordsForDestination(nome string) ([2]float64, bool) {
	if c, ok := DestinationCoords[nome]; ok {
		return c, true
	}
	lower := strings.ToLower(nome)
	for key, coords := range DestinationCoords {
		keyLower := strings.ToLower(key)
		if strings.Contains(lower, keyLower) || strings.Contains(keyLower, lower) {
			return coords, true
		}
	}
	return [2]float64{}, false
}

func ResolveDestination(name string) (NamedPoint, bool) {
	if name == "" {
		return NamedPoint{}, false
	}
	coords, ok := GetCoordsForDestination(name)
	if !ok {
		return NamedPoint{}, false
	}
	return NamedPoint{Name: name, Lat: coords[0], Lng: coords[1]}, true
}

// ComputeVehiclePosition mirrors compute_vehicle_position: interpolates a
// simulated position along the road polyline (or a straight line fallback).
func ComputeVehiclePosition(roadPoints [][2]float64, carico, scarico NamedPoint, progress float64) (lat, lng float64) {
	if len(roadPoints) > 1 {
		idx := int(progress * float64(len(roadPoints)-1))
		if idx >= len(roadPoints) {
			idx = len(roadPoints) - 1
		}
		return roadPoints[idx][0], roadPoints[idx][1]
	}
	return carico.Lat + (scarico.Lat-carico.Lat)*progress, carico.Lng + (scarico.Lng-carico.Lng)*progress
}

type GeoService struct {
	db *gorm.DB
	// OsrmBaseURL/HTTPClient are exported so tests can point them at a
	// local httptest.Server instead of the real OSRM demo server.
	OsrmBaseURL string
	HTTPClient  *http.Client
}

func NewGeoService(db *gorm.DB) *GeoService {
	return &GeoService{
		db:          db,
		OsrmBaseURL: "https://router.project-osrm.org",
		HTTPClient:  &http.Client{Timeout: 10 * time.Second},
	}
}

// ResolveGarage mirrors _resolve_garage: garage_id lookup (by name, since
// garages have no lat/lng columns either) then garage_nome, then the
// hardcoded default.
func (s *GeoService) ResolveGarage(ctx context.Context, garageID, garageNome string) NamedPoint {
	if garageID != "" {
		var g models.Garage
		if err := s.db.WithContext(ctx).First(&g, "id = ?", garageID).Error; err == nil {
			if coords, ok := GarageCoords[g.Nome]; ok {
				return NamedPoint{Name: g.Nome, Lat: coords[0], Lng: coords[1]}
			}
		}
	}
	if garageNome != "" {
		if coords, ok := GarageCoords[garageNome]; ok {
			return NamedPoint{Name: garageNome, Lat: coords[0], Lng: coords[1]}
		}
	}
	coords := GarageCoords[DefaultGarageName]
	return NamedPoint{Name: DefaultGarageName, Lat: coords[0], Lng: coords[1]}
}

type RouteResult struct {
	Points        [][2]float64 `json:"points"`
	DistanceKm    float64      `json:"distance_km"`
	DurationHours float64      `json:"duration_hours"`
	NumPoints     int          `json:"num_points"`
}

// GetRoadRoute mirrors get_road_route: Postgres-backed cache keyed by
// rounded coordinates, then a best-effort OSRM call. Any failure (timeout,
// non-200, malformed response) degrades to nil — never an error — exactly
// like the Python original swallowing all exceptions.
func (s *GeoService) GetRoadRoute(ctx context.Context, fromLat, fromLng, toLat, toLng float64) *RouteResult {
	cacheKey := fmt.Sprintf("%.3f,%.3f_%.3f,%.3f", fromLat, fromLng, toLat, toLng)

	var cached models.RouteCache
	if err := s.db.WithContext(ctx).First(&cached, "key = ?", cacheKey).Error; err == nil {
		var points [][2]float64
		_ = json.Unmarshal(cached.Points, &points)
		return &RouteResult{Points: points, DistanceKm: cached.DistanceKm, DurationHours: cached.DurationHours, NumPoints: cached.NumPoints}
	}

	url := fmt.Sprintf("%s/route/v1/driving/%f,%f;%f,%f?overview=full&geometries=geojson",
		s.OsrmBaseURL, fromLng, fromLat, toLng, toLat)

	reqCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, url, nil)
	if err != nil {
		return nil
	}
	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		slog.Warn("osrm_routing_failed", "error", err.Error(), "cache_key", cacheKey)
		return nil
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil
	}

	var body struct {
		Code   string `json:"code"`
		Routes []struct {
			Distance float64 `json:"distance"`
			Duration float64 `json:"duration"`
			Geometry struct {
				Coordinates [][2]float64 `json:"coordinates"`
			} `json:"geometry"`
		} `json:"routes"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil || body.Code != "Ok" || len(body.Routes) == 0 {
		return nil
	}

	route := body.Routes[0]
	raw := route.Geometry.Coordinates
	step := len(raw) / 80
	if step < 1 {
		step = 1
	}
	simplified := make([][2]float64, 0, len(raw)/step+1)
	for i := 0; i < len(raw); i += step {
		simplified = append(simplified, raw[i])
	}
	if len(raw) > 0 && (len(simplified) == 0 || simplified[len(simplified)-1] != raw[len(raw)-1]) {
		simplified = append(simplified, raw[len(raw)-1])
	}
	points := make([][2]float64, len(simplified))
	for i, c := range simplified {
		points[i] = [2]float64{c[1], c[0]} // [lng,lat] -> [lat,lng]
	}

	result := RouteResult{
		Points:        points,
		DistanceKm:    roundTo1(route.Distance / 1000),
		DurationHours: roundTo1(route.Duration / 3600),
		NumPoints:     len(points),
	}

	pointsJSON, _ := json.Marshal(result.Points)
	cacheRow := models.RouteCache{
		Key: cacheKey, Points: pointsJSON, DistanceKm: result.DistanceKm,
		DurationHours: result.DurationHours, NumPoints: result.NumPoints,
	}
	s.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "key"}},
		DoUpdates: clause.AssignmentColumns([]string{"points", "distance_km", "duration_hours", "num_points"}),
	}).Create(&cacheRow)

	return &result
}

func roundTo1(v float64) float64 {
	return float64(int(v*10+0.5)) / 10
}
