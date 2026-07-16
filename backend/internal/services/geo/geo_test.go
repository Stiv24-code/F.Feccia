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
	osrm := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code": "Ok",
			"routes": []map[string]interface{}{
				{"distance": 50000.0, "duration": 3600.0, "geometry": map[string]interface{}{"coordinates": [][2]float64{{9.0, 45.0}, {9.5, 45.5}}}},
			},
		})
	}))
	defer osrm.Close()

	svc := NewGeoService(newTestDB(t))
	svc.OsrmBaseURL = osrm.URL
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
	osrm := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer osrm.Close()

	svc := NewGeoService(newTestDB(t))
	svc.OsrmBaseURL = osrm.URL
	route := svc.GetRoadRoute(context.Background(), 45.0, 9.0, 45.5, 9.5)
	if route != nil {
		t.Fatalf("expected nil route on OSRM failure, got %+v", route)
	}
}

func TestGetRoadRoute_Timeout(t *testing.T) {
	osrm := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(50 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer osrm.Close()

	svc := NewGeoService(newTestDB(t))
	svc.OsrmBaseURL = osrm.URL
	svc.HTTPClient.Timeout = 10 * time.Millisecond
	route := svc.GetRoadRoute(context.Background(), 45.0, 9.0, 45.5, 9.5)
	if route != nil {
		t.Fatalf("expected nil route on timeout, got %+v", route)
	}
}

func TestResolveGarage_DefaultsWhenUnresolved(t *testing.T) {
	svc := NewGeoService(newTestDB(t))
	p := svc.ResolveGarage(context.Background(), "", "")
	if p.Name != DefaultGarageName {
		t.Fatalf("expected default garage %q, got %q", DefaultGarageName, p.Name)
	}
}
