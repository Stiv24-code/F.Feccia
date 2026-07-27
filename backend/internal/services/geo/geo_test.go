package geo

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"fratelli-feccia/internal/models"
)

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}
	if err := db.AutoMigrate(&models.RouteCache{}, &models.Garage{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	return db
}

func orsFeature(distanceM, durationS float64, coords [][2]float64) map[string]interface{} {
	return map[string]interface{}{
		"properties": map[string]interface{}{
			"summary": map[string]interface{}{"distance": distanceM, "duration": durationS},
		},
		"geometry": map[string]interface{}{"coordinates": coords},
	}
}

func TestGetCoordsForDestination_ExactAndSubstringMatch(t *testing.T) {
	if _, ok := GetCoordsForDestination("Milano (MI)"); !ok {
		t.Fatal("expected exact match for Milano (MI)")
	}
	if _, ok := GetCoordsForDestination("qualcosa Milano (MI) qualcosaltro"); !ok {
		t.Fatal("expected substring match to find Milano (MI)")
	}
	if _, ok := GetCoordsForDestination("Nonexistent City"); ok {
		t.Fatal("expected no match for an unknown destination")
	}
}

func TestGetRoadRoute_CachesAcrossCalls(t *testing.T) {
	calls := 0
	ors := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"features": []map[string]interface{}{
				orsFeature(50000.0, 3600.0, [][2]float64{{9.0, 45.0}, {9.5, 45.5}}),
			},
		})
	}))
	defer ors.Close()

	svc := NewGeoService(newTestDB(t), "test-key", "")
	svc.ORSBaseURL = ors.URL
	ctx := context.Background()

	r1 := svc.GetRoadRoute(ctx, 45.0, 9.0, 45.5, 9.5)
	r2 := svc.GetRoadRoute(ctx, 45.0, 9.0, 45.5, 9.5)
	if r1 == nil || r2 == nil {
		t.Fatalf("expected both calls to succeed, got r1=%v r2=%v", r1, r2)
	}
	if calls != 1 {
		t.Fatalf("expected the second call to be served from cache (1 HTTP call), got %d", calls)
	}
	if r1.DistanceKm != 50 {
		t.Fatalf("expected distance_km 50, got %v", r1.DistanceKm)
	}
}

func TestGetRoadRoute_DegradesGracefullyOnFailure(t *testing.T) {
	ors := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ors.Close()

	svc := NewGeoService(newTestDB(t), "test-key", "")
	svc.ORSBaseURL = ors.URL
	route := svc.GetRoadRoute(context.Background(), 45.0, 9.0, 45.5, 9.5)
	if route != nil {
		t.Fatalf("expected nil route on ORS failure, got %+v", route)
	}
}

func TestGetRoadRoute_Timeout(t *testing.T) {
	ors := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(50 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer ors.Close()

	svc := NewGeoService(newTestDB(t), "test-key", "")
	svc.ORSBaseURL = ors.URL
	svc.HTTPClient.Timeout = 10 * time.Millisecond
	route := svc.GetRoadRoute(context.Background(), 45.0, 9.0, 45.5, 9.5)
	if route != nil {
		t.Fatalf("expected nil route on timeout, got %+v", route)
	}
}

func TestGetRoadRoute_NoAPIKeyDegradesToNil(t *testing.T) {
	svc := NewGeoService(newTestDB(t), "", "")
	route := svc.GetRoadRoute(context.Background(), 45.0, 9.0, 45.5, 9.5)
	if route != nil {
		t.Fatalf("expected nil route when no ORS API key is configured, got %+v", route)
	}
}

func TestGetRoadRouteAlternatives_ReturnsAllFeatures(t *testing.T) {
	var capturedBody map[string]interface{}
	ors := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&capturedBody)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"features": []map[string]interface{}{
				orsFeature(50000.0, 3600.0, [][2]float64{{9.0, 45.0}, {9.5, 45.5}}),
				orsFeature(55000.0, 3300.0, [][2]float64{{9.0, 45.0}, {9.6, 45.4}}),
				orsFeature(60000.0, 3000.0, [][2]float64{{9.0, 45.0}, {9.7, 45.3}}),
			},
		})
	}))
	defer ors.Close()

	svc := NewGeoService(newTestDB(t), "test-key", "")
	svc.ORSBaseURL = ors.URL
	results := svc.GetRoadRouteAlternatives(context.Background(), 45.0, 9.0, 45.5, 9.5, 3)
	if len(results) != 3 {
		t.Fatalf("expected 3 alternatives, got %d", len(results))
	}
	if results[0].DistanceKm != 50 || results[1].DistanceKm != 55 || results[2].DistanceKm != 60 {
		t.Fatalf("unexpected distances: %+v", results)
	}
	altReq, ok := capturedBody["alternative_routes"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected alternative_routes in request body, got %+v", capturedBody)
	}
	if int(altReq["target_count"].(float64)) != 3 {
		t.Fatalf("expected target_count 3, got %v", altReq["target_count"])
	}
}

// TestGetRoadRouteAlternatives_FallsBackToSingleRouteOnFailure mirrors ORS's
// real behavior: alternative_routes errors out entirely once the route
// exceeds ~100km (confirmed against the live API), which would otherwise
// leave a long-haul order with zero proposed routes instead of one.
func TestGetRoadRouteAlternatives_FallsBackToSingleRouteOnFailure(t *testing.T) {
	calls := 0
	ors := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		var body map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&body)
		if _, hasAlt := body["alternative_routes"]; hasAlt {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": map[string]interface{}{"code": 2004, "message": "distance exceeds limit"}})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"features": []map[string]interface{}{orsFeature(120000.0, 7200.0, [][2]float64{{9.0, 45.0}, {12.5, 41.9}})},
		})
	}))
	defer ors.Close()

	svc := NewGeoService(newTestDB(t), "test-key", "")
	svc.ORSBaseURL = ors.URL
	results := svc.GetRoadRouteAlternatives(context.Background(), 45.0, 9.0, 41.9, 12.5, 3)
	if len(results) != 1 {
		t.Fatalf("expected a 1-route fallback when alternatives fail, got %d", len(results))
	}
	if results[0].DistanceKm != 120 {
		t.Fatalf("expected distance_km 120 from the fallback route, got %v", results[0].DistanceKm)
	}
	if calls != 2 {
		t.Fatalf("expected 2 calls (alternatives attempt + single-route fallback), got %d", calls)
	}
}

func TestGetRoadRouteMultiWaypoint_SendsAllCoordinates(t *testing.T) {
	var capturedBody map[string]interface{}
	ors := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&capturedBody)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"features": []map[string]interface{}{
				orsFeature(80000.0, 5400.0, [][2]float64{{9.0, 45.0}, {9.3, 45.2}, {9.5, 45.5}}),
			},
		})
	}))
	defer ors.Close()

	svc := NewGeoService(newTestDB(t), "test-key", "")
	svc.ORSBaseURL = ors.URL
	route := svc.GetRoadRouteMultiWaypoint(context.Background(), []NamedPoint{
		{Name: "Garage", Lat: 45.0, Lng: 9.0},
		{Name: "Carico", Lat: 45.2, Lng: 9.3},
		{Name: "Scarico", Lat: 45.5, Lng: 9.5},
	})
	if route == nil {
		t.Fatal("expected a route, got nil")
	}
	if route.DistanceKm != 80 {
		t.Fatalf("expected distance_km 80, got %v", route.DistanceKm)
	}
	coords, ok := capturedBody["coordinates"].([]interface{})
	if !ok || len(coords) != 3 {
		t.Fatalf("expected 3 coordinates sent to ORS, got %+v", capturedBody["coordinates"])
	}
	if _, hasAlt := capturedBody["alternative_routes"]; hasAlt {
		t.Fatalf("did not expect alternative_routes on a multi-waypoint request")
	}
}

func TestGeocodeSearch_ParsesFeatures(t *testing.T) {
	var capturedQuery string
	ors := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedQuery = r.URL.Query().Get("text")
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"features": []map[string]interface{}{
				{
					"properties": map[string]interface{}{
						"label": "Piazza del Duomo, Milan, MI, Italy", "locality": "Milan",
						"postalcode": "20122", "region_a": "MI", "country": "Italy",
					},
					"geometry": map[string]interface{}{"coordinates": [2]float64{9.19, 45.46}},
				},
			},
		})
	}))
	defer ors.Close()

	svc := NewGeoService(newTestDB(t), "test-key", "")
	svc.ORSBaseURL = ors.URL
	results := svc.GeocodeSearch(context.Background(), "Piazza Duomo, Milano", 5)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Label != "Piazza del Duomo, Milan, MI, Italy" || results[0].Locality != "Milan" {
		t.Fatalf("unexpected result: %+v", results[0])
	}
	if results[0].Postcode != "20122" || results[0].ProvinceA != "MI" || results[0].Country != "Italy" {
		t.Fatalf("expected postcode/province/country to be parsed, got %+v", results[0])
	}
	if results[0].Lat != 45.46 || results[0].Lng != 9.19 {
		t.Fatalf("expected coordinates flipped [lng,lat]->(lat,lng), got lat=%v lng=%v", results[0].Lat, results[0].Lng)
	}
	if capturedQuery != "Piazza Duomo, Milano" {
		t.Fatalf("expected query text to be sent as-is, got %q", capturedQuery)
	}
}

func TestGeocodeSearch_NoAPIKeyDegradesToNil(t *testing.T) {
	svc := NewGeoService(newTestDB(t), "", "")
	results := svc.GeocodeSearch(context.Background(), "Milano", 5)
	if results != nil {
		t.Fatalf("expected nil results when no ORS API key is configured, got %+v", results)
	}
}

func TestResolveGarage_DefaultsWhenUnresolved(t *testing.T) {
	svc := NewGeoService(newTestDB(t), "test-key", "")
	p := svc.ResolveGarage(context.Background(), "", "")
	if p.Name != DefaultGarageName {
		t.Fatalf("expected default garage %q, got %q", DefaultGarageName, p.Name)
	}
}
