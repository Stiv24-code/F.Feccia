// Package geo provides truck-aware route computation (OpenRouteService
// driving-hgv profile) plus the coordinate-lookup helpers ported from
// backend/services.py. Shared here rather than duplicated so trip segments,
// the live map, and order routes all hit the same route cache.
package geo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
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

// ResolveDestination resolves a destination's coordinates, preferring the
// lat/lng stored on its own record (when the caller has one — e.g. from
// Destination.Lat/Lng) over the hardcoded DestinationCoords fallback table.
func ResolveDestination(name string, lat, lng *float64) (NamedPoint, bool) {
	if name == "" {
		return NamedPoint{}, false
	}
	if lat != nil && lng != nil {
		return NamedPoint{Name: name, Lat: *lat, Lng: *lng}, true
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
	// ORSApiKey/ORSBaseURL are exported so tests can point them at a local
	// httptest.Server instead of the real OpenRouteService API. Empty
	// ORSApiKey disables routing entirely (GetRoadRoute* degrade to nil),
	// same posture as pkg/s3invoices when its bucket config is empty.
	ORSApiKey  string
	ORSBaseURL string
	HTTPClient *http.Client
}

func NewGeoService(db *gorm.DB, orsApiKey, orsBaseURL string) *GeoService {
	if orsBaseURL == "" {
		orsBaseURL = "https://api.openrouteservice.org"
	}
	return &GeoService{
		db:         db,
		ORSApiKey:  orsApiKey,
		ORSBaseURL: orsBaseURL,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// ResolveGarage resolves a trip's starting point ("partenza"). Despite the
// name (kept for wire compatibility with Trip.GarageID/GarageNome), garageID
// is looked up against Garage, then Destination, then WashStation — a trip
// can start from any registered point, not only a garage. Each lookup
// prefers the row's own Lat/Lng, falling back to the hardcoded GarageCoords
// table for garages created before that column existed. garageNome and the
// hardcoded default are the last resorts, mirroring the original
// name-only _resolve_garage behaviour.
func (s *GeoService) ResolveGarage(ctx context.Context, garageID, garageNome string) NamedPoint {
	if garageID != "" {
		var g models.Garage
		if err := s.db.WithContext(ctx).First(&g, "id = ?", garageID).Error; err == nil {
			if g.Lat != nil && g.Lng != nil {
				return NamedPoint{Name: g.Nome, Lat: *g.Lat, Lng: *g.Lng}
			}
			if coords, ok := GarageCoords[g.Nome]; ok {
				return NamedPoint{Name: g.Nome, Lat: coords[0], Lng: coords[1]}
			}
		}
		var d models.Destination
		if err := s.db.WithContext(ctx).First(&d, "id = ?", garageID).Error; err == nil {
			if p, ok := ResolveDestination(d.Nome, d.Lat, d.Lng); ok {
				return p
			}
		}
		var w models.WashStation
		if err := s.db.WithContext(ctx).First(&w, "id = ?", garageID).Error; err == nil && w.Lat != nil && w.Lng != nil {
			return NamedPoint{Name: w.Nome, Lat: *w.Lat, Lng: *w.Lng}
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
// rounded coordinates, then a best-effort ORS call. Any failure (no API key,
// timeout, non-200, malformed response) degrades to nil — never an error —
// same posture the OSRM-based version this replaces had.
func (s *GeoService) GetRoadRoute(ctx context.Context, fromLat, fromLng, toLat, toLng float64) *RouteResult {
	cacheKey := fmt.Sprintf("%.3f,%.3f_%.3f,%.3f", fromLat, fromLng, toLat, toLng)

	var cached models.RouteCache
	if err := s.db.WithContext(ctx).First(&cached, "key = ?", cacheKey).Error; err == nil {
		var points [][2]float64
		_ = json.Unmarshal(cached.Points, &points)
		return &RouteResult{Points: points, DistanceKm: cached.DistanceKm, DurationHours: cached.DurationHours, NumPoints: cached.NumPoints}
	}

	results := s.callORS(ctx, [][2]float64{{fromLng, fromLat}, {toLng, toLat}}, 0)
	if len(results) == 0 {
		return nil
	}
	result := results[0]

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

// GetRoadRouteAlternatives requests up to count alternative routes between
// two points (no cache — used only at order-assignment time, not in the
// high-frequency map/segment recompute loops). ORS's alternative_routes only
// supports a plain 2-coordinate request (no via-points), which is exactly
// what this takes.
//
// ORS additionally refuses alternative_routes outright (HTTP error) once the
// route's approximate distance exceeds 100km — a real limit on their side,
// not a bug here. Since this app routes international/long-haul freight,
// most real orders exceed that, so on any failure this falls back to a
// single plain route (no alternative_routes) rather than showing nothing.
func (s *GeoService) GetRoadRouteAlternatives(ctx context.Context, fromLat, fromLng, toLat, toLng float64, count int) []RouteResult {
	coords := [][2]float64{{fromLng, fromLat}, {toLng, toLat}}
	if results := s.callORS(ctx, coords, count); len(results) > 0 {
		return results
	}
	return s.callORS(ctx, coords, 0)
}

// GetRoadRouteMultiWaypoint routes through an arbitrary ordered sequence of
// points (garage/destination/wash-station via-points a manager added by
// hand) as a single ORS request. No cache: manual edits are infrequent and
// the waypoint sequence is unique per call.
func (s *GeoService) GetRoadRouteMultiWaypoint(ctx context.Context, points []NamedPoint) *RouteResult {
	coords := make([][2]float64, len(points))
	for i, p := range points {
		coords[i] = [2]float64{p.Lng, p.Lat}
	}
	results := s.callORS(ctx, coords, 0)
	if len(results) == 0 {
		return nil
	}
	return &results[0]
}

type GeocodeResult struct {
	Label     string
	Locality  string
	Postcode  string
	ProvinceA string
	Lat       float64
	Lng       float64
}

// GeocodeSearch forward-geocodes free-text (address, place name...) via ORS's
// Pelias-based /geocode/search — used so anagrafica forms (Destination,
// Garage, WashStation) can type an address and place the marker instead of
// only clicking the map. Match quality depends entirely on ORS/OpenStreetMap
// data coverage for that address (can be street-level in some areas,
// locality-only in others) — never guaranteed exact, hence returning several
// candidates for the user to pick from rather than just the first result.
func (s *GeoService) GeocodeSearch(ctx context.Context, query string, limit int) []GeocodeResult {
	if s.ORSApiKey == "" || strings.TrimSpace(query) == "" {
		return nil
	}
	if limit <= 0 || limit > 10 {
		limit = 5
	}

	reqCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	reqURL := fmt.Sprintf("%s/geocode/search?api_key=%s&text=%s&size=%d",
		s.ORSBaseURL, url.QueryEscape(s.ORSApiKey), url.QueryEscape(query), limit)
	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil
	}

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		slog.Warn("ors_geocode_failed", "error", err.Error())
		return nil
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		slog.Warn("ors_geocode_failed", "status", resp.StatusCode)
		return nil
	}

	var geoJSON struct {
		Features []struct {
			Properties struct {
				Label      string `json:"label"`
				Locality   string `json:"locality"`
				Postalcode string `json:"postalcode"`
				RegionA    string `json:"region_a"`
			} `json:"properties"`
			Geometry struct {
				Coordinates [2]float64 `json:"coordinates"`
			} `json:"geometry"`
		} `json:"features"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&geoJSON); err != nil {
		return nil
	}

	results := make([]GeocodeResult, 0, len(geoJSON.Features))
	for _, f := range geoJSON.Features {
		results = append(results, GeocodeResult{
			Label:     f.Properties.Label,
			Locality:  f.Properties.Locality,
			Postcode:  f.Properties.Postalcode,
			ProvinceA: f.Properties.RegionA,
			Lng:       f.Geometry.Coordinates[0],
			Lat:       f.Geometry.Coordinates[1],
		})
	}
	return results
}

// callORS POSTs to ORS's /v2/directions/driving-hgv/geojson and returns up
// to max(1, altCount) routes. altCount > 0 requests alternative_routes (ORS
// caps target_count at 3, and — per ORS docs — only honors it for a plain
// 2-coordinate request). Any failure (no key configured, timeout, non-200,
// malformed body) degrades to an empty slice, never an error, mirroring the
// Python original's blanket exception-swallowing.
func (s *GeoService) callORS(ctx context.Context, coords [][2]float64, altCount int) []RouteResult {
	if s.ORSApiKey == "" || len(coords) < 2 {
		return nil
	}

	body := map[string]any{"coordinates": coords, "instructions": false, "geometry_simplify": true}
	if altCount > 0 {
		if altCount > 3 {
			altCount = 3
		}
		body["alternative_routes"] = map[string]any{
			"target_count":  altCount,
			"weight_factor": 1.6,
			"share_factor":  0.6,
		}
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return nil
	}

	url := fmt.Sprintf("%s/v2/directions/driving-hgv/geojson", s.ORSBaseURL)
	reqCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(reqCtx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return nil
	}
	req.Header.Set("Authorization", s.ORSApiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		slog.Warn("ors_routing_failed", "error", err.Error())
		return nil
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		slog.Warn("ors_routing_failed", "status", resp.StatusCode)
		return nil
	}

	var geoJSON struct {
		Features []struct {
			Properties struct {
				Summary struct {
					Distance float64 `json:"distance"`
					Duration float64 `json:"duration"`
				} `json:"summary"`
			} `json:"properties"`
			Geometry struct {
				Coordinates [][2]float64 `json:"coordinates"`
			} `json:"geometry"`
		} `json:"features"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&geoJSON); err != nil || len(geoJSON.Features) == 0 {
		return nil
	}

	results := make([]RouteResult, 0, len(geoJSON.Features))
	for _, f := range geoJSON.Features {
		results = append(results, RouteResult{
			Points:        simplifyAndFlip(f.Geometry.Coordinates),
			DistanceKm:    roundTo1(f.Properties.Summary.Distance / 1000),
			DurationHours: roundTo1(f.Properties.Summary.Duration / 3600),
			NumPoints:     len(f.Geometry.Coordinates),
		})
	}
	// NumPoints above is set from the raw geometry before simplification;
	// fix up to reflect what's actually returned.
	for i := range results {
		results[i].NumPoints = len(results[i].Points)
	}
	return results
}

// simplifyAndFlip downsamples a raw [lng,lat] coordinate list to ~80 points
// (always keeping the last point) and converts to [lat,lng] for Leaflet.
func simplifyAndFlip(raw [][2]float64) [][2]float64 {
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
	return points
}

func roundTo1(v float64) float64 {
	return float64(int(v*10+0.5)) / 10
}
