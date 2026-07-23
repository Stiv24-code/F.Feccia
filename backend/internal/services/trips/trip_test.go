package trips

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"fratelli-feccia/internal/dto"
	"fratelli-feccia/internal/models"
	"fratelli-feccia/pkg/database"
)

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}
	if err := db.AutoMigrate(&database.Counter{}, &models.Order{}, &models.OrderItem{}, &models.Garage{}, &models.Trip{}, &models.TripSegment{}, &models.RouteCache{}, &models.Customer{}, &models.Destination{}, &models.Product{}, &models.Driver{}, &models.Carrier{}, &models.WashStation{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	return db
}

// fakeOSRM stubs the OSRM demo server so tests never hit the real network.
// Always returns a fixed 100km/2h route.
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
						"coordinates": [][2]float64{{9.19, 45.46}, {9.50, 45.31}},
					},
				},
			},
		})
	}))
}

func newTestService(t *testing.T, osrmURL string) *TripService {
	t.Helper()
	svc := NewTripService(newTestDB(t))
	if osrmURL != "" {
		svc.geo.OsrmBaseURL = osrmURL
	}
	return svc
}

func geoPtr(v float64) *float64 { return &v }

func seedDestination(t *testing.T, db *gorm.DB, nome string, lat, lng float64) uuid.UUID {
	t.Helper()
	d := models.Destination{ID: uuid.New(), Nome: nome, Active: true, Lat: geoPtr(lat), Lng: geoPtr(lng)}
	if err := db.Create(&d).Error; err != nil {
		t.Fatalf("failed to seed destination %q: %v", nome, err)
	}
	return d.ID
}

func createOrder(t *testing.T, db *gorm.DB, carico, scarico, dataRitiro string) models.Order {
	t.Helper()
	caricoID := seedDestination(t, db, carico, 45.0, 9.0)
	scaricoID := seedDestination(t, db, scarico, 45.5, 9.5)
	o := models.Order{
		ID: uuid.New(), ClienteID: uuid.New(),
		DestinazioneCaricoID: &caricoID, DestinazioneScaricoID: &scaricoID,
		DataRitiro: dataRitiro, Stato: "PIANIFICABILE",
		ServiziAccessori: []byte("[]"), CostiAccessori: []byte("[]"),
	}
	if err := db.Create(&o).Error; err != nil {
		t.Fatalf("failed to seed order: %v", err)
	}
	return o
}

func TestTripService_Create_SyncsOrdersAndComputesSegments(t *testing.T) {
	osrm := fakeOSRM(t)
	defer osrm.Close()

	db := newTestDB(t)
	svc := NewTripService(db)
	svc.geo.OsrmBaseURL = osrm.URL

	order := createOrder(t, db, "Milano (MI)", "Lodi (LO)", "2026-01-10")

	trip, err := svc.Create(context.Background(), dto.TripRequest{
		OrdiniIds: []string{order.ID.String()}, TargaMotrice: "AB123CD", AutistaID: uuid.New().String(),
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if trip.Stato != "PIANIFICATO" {
		t.Fatalf("expected initial stato PIANIFICATO, got %q", trip.Stato)
	}
	// base_carico + carico_scarico + scarico_base = 3 segments for 1 order.
	if len(trip.Segmenti) != 3 {
		t.Fatalf("expected 3 segments, got %d (%+v)", len(trip.Segmenti), trip.Segmenti)
	}
	if trip.KmTotali != 300 {
		t.Fatalf("expected km_totali 300 (3 segments x 100km stub), got %v", trip.KmTotali)
	}

	var updatedOrder models.Order
	db.First(&updatedOrder, "id = ?", order.ID)
	if updatedOrder.Stato != "PIANIFICATO" || updatedOrder.ViaggioID == nil || *updatedOrder.ViaggioID != trip.ID {
		t.Fatalf("expected order synced to PIANIFICATO with viaggio_id set, got %+v", updatedOrder)
	}
}

func TestTripService_Create_SkipsOrdersNotPianificabile(t *testing.T) {
	osrm := fakeOSRM(t)
	defer osrm.Close()

	db := newTestDB(t)
	svc := NewTripService(db)
	svc.geo.OsrmBaseURL = osrm.URL

	order := createOrder(t, db, "Milano (MI)", "Lodi (LO)", "2026-01-10")
	db.Model(&models.Order{}).Where("id = ?", order.ID).Update("stato", "CHIUSO")

	if _, err := svc.Create(context.Background(), dto.TripRequest{OrdiniIds: []string{order.ID.String()}}); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	var untouched models.Order
	db.First(&untouched, "id = ?", order.ID)
	if untouched.Stato != "CHIUSO" || untouched.ViaggioID != nil {
		t.Fatalf("expected non-PIANIFICABILE order to be left untouched, got %+v", untouched)
	}
}

func TestTripService_GetByID_IncludesLinkedOrders(t *testing.T) {
	osrm := fakeOSRM(t)
	defer osrm.Close()
	db := newTestDB(t)
	svc := NewTripService(db)
	svc.geo.OsrmBaseURL = osrm.URL

	order := createOrder(t, db, "Milano (MI)", "Lodi (LO)", "2026-01-10")
	trip, err := svc.Create(context.Background(), dto.TripRequest{OrdiniIds: []string{order.ID.String()}})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	detail, err := svc.GetByID(context.Background(), trip.ID)
	if err != nil {
		t.Fatalf("GetByID returned error: %v", err)
	}
	if detail == nil || len(detail.Ordini) != 1 {
		t.Fatalf("expected 1 linked order, got %+v", detail)
	}
}

func TestTripService_GetByID_NotFoundReturnsNilNil(t *testing.T) {
	svc := newTestService(t, "")
	detail, err := svc.GetByID(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if detail != nil {
		t.Fatalf("expected nil for missing trip, got %+v", detail)
	}
}

func TestTripService_Complete_RequiresInCorso(t *testing.T) {
	osrm := fakeOSRM(t)
	defer osrm.Close()
	db := newTestDB(t)
	svc := NewTripService(db)
	svc.geo.OsrmBaseURL = osrm.URL

	order := createOrder(t, db, "Milano (MI)", "Lodi (LO)", "2026-01-10")
	trip, err := svc.Create(context.Background(), dto.TripRequest{OrdiniIds: []string{order.ID.String()}})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	// Complete before Start must fail (trip still PIANIFICATO).
	if _, err := svc.Complete(context.Background(), trip.ID); err == nil {
		t.Fatal("expected error completing a PIANIFICATO trip")
	}
}

func TestTripService_StartThenComplete_FullLifecycle(t *testing.T) {
	osrm := fakeOSRM(t)
	defer osrm.Close()
	db := newTestDB(t)
	svc := NewTripService(db)
	svc.geo.OsrmBaseURL = osrm.URL

	order := createOrder(t, db, "Milano (MI)", "Lodi (LO)", "2026-01-10")
	trip, err := svc.Create(context.Background(), dto.TripRequest{OrdiniIds: []string{order.ID.String()}})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	// Start before... nothing to guard against here except wrong stato, tested below.
	startResult, err := svc.Start(context.Background(), trip.ID)
	if err != nil || !startResult.OK {
		t.Fatalf("Start returned error: %v (result=%+v)", err, startResult)
	}

	var startedOrder models.Order
	db.First(&startedOrder, "id = ?", order.ID)
	if startedOrder.Stato != "VIAGGIO" {
		t.Fatalf("expected order stato VIAGGIO after trip Start, got %q", startedOrder.Stato)
	}

	// Start again must fail (no longer PIANIFICATO).
	if _, err := svc.Start(context.Background(), trip.ID); err == nil {
		t.Fatal("expected error starting an already IN_CORSO trip")
	}

	result, err := svc.Complete(context.Background(), trip.ID)
	if err != nil || !result.OK {
		t.Fatalf("Complete returned error: %v (result=%+v)", err, result)
	}

	var closedOrder models.Order
	db.First(&closedOrder, "id = ?", order.ID)
	if closedOrder.Stato != "CHIUSO" {
		t.Fatalf("expected order stato CHIUSO after complete, got %q", closedOrder.Stato)
	}

	var completedTrip models.Trip
	db.First(&completedTrip, "id = ?", trip.ID)
	if completedTrip.Stato != "COMPLETATO" {
		t.Fatalf("expected trip stato COMPLETATO, got %q", completedTrip.Stato)
	}

	// Complete again must fail (no longer IN_CORSO).
	if _, err := svc.Complete(context.Background(), trip.ID); err == nil {
		t.Fatal("expected error completing an already COMPLETATO trip")
	}
}

func TestTripService_AddOrder_ValidatesStateAndRecomputes(t *testing.T) {
	osrm := fakeOSRM(t)
	defer osrm.Close()
	db := newTestDB(t)
	svc := NewTripService(db)
	svc.geo.OsrmBaseURL = osrm.URL

	first := createOrder(t, db, "Milano (MI)", "Lodi (LO)", "2026-01-10")
	trip, err := svc.Create(context.Background(), dto.TripRequest{OrdiniIds: []string{first.ID.String()}})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	second := createOrder(t, db, "Cuneo (CN)", "Milano (MI)", "2026-01-11")
	result, err := svc.AddOrder(context.Background(), trip.ID, second.ID)
	if err != nil || !result.OK {
		t.Fatalf("AddOrder returned error: %v (result=%+v)", err, result)
	}

	var addedOrder models.Order
	db.First(&addedOrder, "id = ?", second.ID)
	if addedOrder.Stato != "PIANIFICATO" {
		t.Fatalf("expected order added to a PIANIFICATO trip to become PIANIFICATO, got %q", addedOrder.Stato)
	}

	detail, err := svc.GetByID(context.Background(), trip.ID)
	if err != nil {
		t.Fatalf("GetByID returned error: %v", err)
	}
	if len(detail.OrdiniIds) != 2 {
		t.Fatalf("expected 2 orders linked to trip after AddOrder, got %v", detail.OrdiniIds)
	}

	// Adding a non-PIANIFICABILE order must fail.
	db.Model(&models.Order{}).Where("id = ?", second.ID).Update("stato", "CHIUSO")
	_, err = svc.AddOrder(context.Background(), trip.ID, second.ID)
	if err == nil {
		t.Fatal("expected error when adding a non-PIANIFICABILE order")
	}
}

func TestTripService_AddOrder_ToInCorsoTripGoesStraightToViaggio(t *testing.T) {
	osrm := fakeOSRM(t)
	defer osrm.Close()
	db := newTestDB(t)
	svc := NewTripService(db)
	svc.geo.OsrmBaseURL = osrm.URL

	first := createOrder(t, db, "Milano (MI)", "Lodi (LO)", "2026-01-10")
	trip, err := svc.Create(context.Background(), dto.TripRequest{OrdiniIds: []string{first.ID.String()}})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if _, err := svc.Start(context.Background(), trip.ID); err != nil {
		t.Fatalf("Start returned error: %v", err)
	}

	second := createOrder(t, db, "Cuneo (CN)", "Milano (MI)", "2026-01-11")
	if _, err := svc.AddOrder(context.Background(), trip.ID, second.ID); err != nil {
		t.Fatalf("AddOrder returned error: %v", err)
	}

	var addedOrder models.Order
	db.First(&addedOrder, "id = ?", second.ID)
	if addedOrder.Stato != "VIAGGIO" {
		t.Fatalf("expected order added to an already-departed (IN_CORSO) trip to become VIAGGIO directly, got %q", addedOrder.Stato)
	}
}
