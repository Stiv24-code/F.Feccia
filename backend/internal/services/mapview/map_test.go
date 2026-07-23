package mapview

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"fratelli-feccia/internal/models"
)

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}
	if err := db.AutoMigrate(&models.Order{}, &models.OrderItem{}, &models.Customer{}, &models.Destination{}, &models.Product{}, &models.Driver{}, &models.Carrier{}, &models.Vehicle{}, &models.RouteCache{}, &models.Garage{}, &models.WashStation{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	return db
}

func fakeOSRM(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code": "Ok",
			"routes": []map[string]interface{}{
				{
					"distance": 100000.0,
					"duration": 7200.0,
					"geometry": map[string]interface{}{
						"coordinates": [][2]float64{{9.19, 45.46}, {9.35, 45.39}, {9.50, 45.31}},
					},
				},
			},
		})
	}))
}

// makeDestination seeds a real Destination row so an order can reference it
// via FK — active controls whether it shows up in the map's active-POI list,
// independent of whether the order's carico/scarico association resolves.
func makeDestination(t *testing.T, db *gorm.DB, nome string, lat, lng float64, active bool) uuid.UUID {
	t.Helper()
	d := models.Destination{ID: uuid.New(), Nome: nome, Lat: &lat, Lng: &lng, Active: active}
	if err := db.Create(&d).Error; err != nil {
		t.Fatalf("failed to seed destination %q: %v", nome, err)
	}
	if !active {
		if err := db.Model(&models.Destination{}).Where("id = ?", d.ID).Update("active", false).Error; err != nil {
			t.Fatalf("failed to deactivate destination %q: %v", nome, err)
		}
	}
	return d.ID
}

func createOrder(t *testing.T, db *gorm.DB, stato string, caricoID, scaricoID *uuid.UUID, targa string) models.Order {
	t.Helper()
	o := models.Order{
		ID: uuid.New(), ClienteID: uuid.New(), Stato: models.OrderStato(stato),
		DestinazioneCaricoID: caricoID, DestinazioneScaricoID: scaricoID, TargaMotrice: targa,
		DataRitiro: "2026-01-10", ServiziAccessori: []byte("[]"), CostiAccessori: []byte("[]"),
	}
	if err := db.Create(&o).Error; err != nil {
		t.Fatalf("failed to seed order: %v", err)
	}
	return o
}

func TestMapService_Trips_SimulatedPositionWhenNoGPS(t *testing.T) {
	osrm := fakeOSRM(t)
	defer osrm.Close()

	db := newTestDB(t)
	svc := NewMapService(db)
	svc.geo.OsrmBaseURL = osrm.URL

	carico := makeDestination(t, db, "Milano (MI)", 45.4642, 9.19, false)
	scarico := makeDestination(t, db, "Lodi (LO)", 45.3138, 9.5032, false)
	createOrder(t, db, "VIAGGIO", &carico, &scarico, "AB123CD")

	resp, err := svc.Trips(context.Background())
	if err != nil {
		t.Fatalf("Trips returned error: %v", err)
	}
	if len(resp.Routes) != 1 {
		t.Fatalf("expected 1 route, got %d", len(resp.Routes))
	}
	r := resp.Routes[0]
	if r.GpsLive {
		t.Fatal("expected simulated (non-live) position when no matching GPS vehicle exists")
	}
	if r.Progress != 0.6 {
		t.Fatalf("expected progress 0.6 for VIAGGIO without GPS, got %v", r.Progress)
	}
	if r.DistanceKm != 100 {
		t.Fatalf("expected distance_km 100 from OSRM stub, got %v", r.DistanceKm)
	}
	if resp.Stats.InViaggio != 1 || resp.Stats.GpsLive != 0 {
		t.Fatalf("unexpected stats: %+v", resp.Stats)
	}
}

func TestMapService_Trips_UsesLiveGPSWhenActiveAndInViaggio(t *testing.T) {
	osrm := fakeOSRM(t)
	defer osrm.Close()

	db := newTestDB(t)
	svc := NewMapService(db)
	svc.geo.OsrmBaseURL = osrm.URL

	carico := makeDestination(t, db, "Milano (MI)", 45.4642, 9.19, false)
	scarico := makeDestination(t, db, "Lodi (LO)", 45.3138, 9.5032, false)
	createOrder(t, db, "VIAGGIO", &carico, &scarico, "AB123CD")
	vehicle := models.Vehicle{
		ID: uuid.New(), Targa: "AB123CD", Active: true,
		LastLat: 45.39, LastLng: 9.35, LastSpeedKmh: 80, GpsActive: true, GpsSource: "webhook",
	}
	if err := db.Create(&vehicle).Error; err != nil {
		t.Fatalf("failed to seed vehicle: %v", err)
	}

	resp, err := svc.Trips(context.Background())
	if err != nil {
		t.Fatalf("Trips returned error: %v", err)
	}
	if len(resp.Routes) != 1 {
		t.Fatalf("expected 1 route, got %d", len(resp.Routes))
	}
	r := resp.Routes[0]
	if !r.GpsLive {
		t.Fatal("expected GPS-live position when an active vehicle matches the order's targa")
	}
	if r.CurrentPosition.Lat != 45.39 || r.CurrentPosition.Lng != 9.35 {
		t.Fatalf("expected current position to match live GPS coords, got %+v", r.CurrentPosition)
	}
	if resp.Stats.GpsLive != 1 {
		t.Fatalf("expected gps_live stat 1, got %d", resp.Stats.GpsLive)
	}
}

func TestMapService_Trips_SkipsOrderWithUnresolvableDestination(t *testing.T) {
	db := newTestDB(t)
	svc := NewMapService(db)

	createOrder(t, db, "PIANIFICABILE", nil, nil, "")

	resp, err := svc.Trips(context.Background())
	if err != nil {
		t.Fatalf("Trips returned error: %v", err)
	}
	if len(resp.Routes) != 0 {
		t.Fatalf("expected order with no destination reference to be skipped, got %d routes", len(resp.Routes))
	}
}

func TestMapService_Trips_DegradesToStraightLineOnOSRMFailure(t *testing.T) {
	osrm := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer osrm.Close()

	db := newTestDB(t)
	svc := NewMapService(db)
	svc.geo.OsrmBaseURL = osrm.URL

	carico := makeDestination(t, db, "Milano (MI)", 45.4642, 9.19, false)
	scarico := makeDestination(t, db, "Lodi (LO)", 45.3138, 9.5032, false)
	createOrder(t, db, "CHIUSO", &carico, &scarico, "")

	resp, err := svc.Trips(context.Background())
	if err != nil {
		t.Fatalf("Trips returned error: %v", err)
	}
	if len(resp.Routes) != 1 {
		t.Fatalf("expected 1 route, got %d", len(resp.Routes))
	}
	r := resp.Routes[0]
	if len(r.RoadPoints) != 2 {
		t.Fatalf("expected 2-point straight line fallback, got %d points", len(r.RoadPoints))
	}
	if r.Progress != 1.0 {
		t.Fatalf("expected progress 1.0 for CHIUSO, got %v", r.Progress)
	}
	if r.DistanceKm != 0 {
		t.Fatalf("expected distance_km 0 when OSRM fails, got %v", r.DistanceKm)
	}
}

func TestMapService_Trips_IncludesPOIAndGarages(t *testing.T) {
	osrm := fakeOSRM(t)
	defer osrm.Close()
	db := newTestDB(t)
	svc := NewMapService(db)
	svc.geo.OsrmBaseURL = osrm.URL

	carico := makeDestination(t, db, "Milano (MI)", 45.4642, 9.1900, true)
	scarico := makeDestination(t, db, "Lodi (LO)", 45.3138, 9.5032, false)
	createOrder(t, db, "VIAGGIO", &carico, &scarico, "")

	lodiLat, lodiLng := 45.3138, 9.5032
	rhoLat, rhoLng := 45.5306, 9.0393
	inactiveID := uuid.New()
	garages := []models.Garage{
		{ID: uuid.New(), Nome: "Garage Feccia F.lli - Lodi", Lat: &lodiLat, Lng: &lodiLng, Active: true},
		{ID: inactiveID, Nome: "Deposito Milano Rho", Lat: &rhoLat, Lng: &rhoLng, Active: true},
	}
	if err := db.Create(&garages).Error; err != nil {
		t.Fatalf("failed to seed garages: %v", err)
	}
	// Active:false at struct-literal time is silently overwritten by GORM's
	// `default:true` on create (can't tell "false" from "unset" for a plain
	// bool) — flip it with a targeted update instead, same as Delete() does.
	if err := db.Model(&models.Garage{}).Where("id = ?", inactiveID).Update("active", false).Error; err != nil {
		t.Fatalf("failed to deactivate garage: %v", err)
	}

	washLat, washLng := 45.3852, 10.9296
	inactiveWashID := uuid.New()
	washStations := []models.WashStation{
		{ID: uuid.New(), Nome: "Lavaggio Cisterne Verona", Lat: &washLat, Lng: &washLng, Active: true},
		{ID: inactiveWashID, Nome: "Lavaggio Dismesso", Lat: &washLat, Lng: &washLng, Active: true},
	}
	if err := db.Create(&washStations).Error; err != nil {
		t.Fatalf("failed to seed wash stations: %v", err)
	}
	if err := db.Model(&models.WashStation{}).Where("id = ?", inactiveWashID).Update("active", false).Error; err != nil {
		t.Fatalf("failed to deactivate wash station: %v", err)
	}

	resp, err := svc.Trips(context.Background())
	if err != nil {
		t.Fatalf("Trips returned error: %v", err)
	}
	if len(resp.POI) != 1 || resp.POI[0].Nome != "Milano (MI)" {
		t.Fatalf("expected 1 POI resolved from active destination, got %+v", resp.POI)
	}
	if len(resp.Garages) != 1 || resp.Garages[0].Nome != "Garage Feccia F.lli - Lodi" {
		t.Fatalf("expected only the active garage from the DB, got %+v", resp.Garages)
	}
	if len(resp.WashStations) != 1 || resp.WashStations[0].Nome != "Lavaggio Cisterne Verona" {
		t.Fatalf("expected only the active wash station from the DB, got %+v", resp.WashStations)
	}
}
