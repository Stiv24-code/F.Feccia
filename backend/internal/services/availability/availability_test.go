package availability

import (
	"context"
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
	if err := db.AutoMigrate(&models.Vehicle{}, &models.Driver{}, &models.Order{}, &models.OrderItem{}, &models.DriverUnavailability{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	return db
}

func TestVehicleAvailability_BusyWhenOverlappingViaggioOrder(t *testing.T) {
	db := newTestDB(t)
	svc := NewAvailabilityService(db)

	v1 := models.Vehicle{ID: uuid.New(), Targa: "AB123CD", Active: true}
	v2 := models.Vehicle{ID: uuid.New(), Targa: "XY999ZZ", Active: true}
	if err := db.Create(&v1).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&v2).Error; err != nil {
		t.Fatal(err)
	}

	busy := models.Order{
		ID: uuid.New(), ClienteID: uuid.New(), Stato: "VIAGGIO", TargaMotrice: "AB123CD",
		DataRitiro: "2026-01-05", DataConsegna: "2026-01-10",
		ServiziAccessori: []byte("[]"), CostiAccessori: []byte("[]"),
	}
	if err := db.Create(&busy).Error; err != nil {
		t.Fatal(err)
	}

	result, err := svc.VehicleAvailability(context.Background(), "2026-01-08", "2026-01-12")
	if err != nil {
		t.Fatalf("VehicleAvailability returned error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 vehicles, got %d", len(result))
	}
	statusByTarga := map[string]string{}
	for _, r := range result {
		statusByTarga[r.Targa] = r.Disponibilita
	}
	if statusByTarga["AB123CD"] != "busy" {
		t.Fatalf("expected AB123CD busy, got %q", statusByTarga["AB123CD"])
	}
	if statusByTarga["XY999ZZ"] != "available" {
		t.Fatalf("expected XY999ZZ available, got %q", statusByTarga["XY999ZZ"])
	}
}

func TestVehicleAvailability_TrailerPlateAlsoMarkedBusy(t *testing.T) {
	db := newTestDB(t)
	svc := NewAvailabilityService(db)

	trailer := models.Vehicle{ID: uuid.New(), Targa: "TR001AA", Active: true}
	if err := db.Create(&trailer).Error; err != nil {
		t.Fatal(err)
	}
	order := models.Order{
		ID: uuid.New(), ClienteID: uuid.New(), Stato: "VIAGGIO", TargaMotrice: "AB123CD", TargaRimorchio: "TR001AA",
		DataRitiro: "2026-01-05", DataConsegna: "2026-01-10",
		ServiziAccessori: []byte("[]"), CostiAccessori: []byte("[]"),
	}
	if err := db.Create(&order).Error; err != nil {
		t.Fatal(err)
	}

	result, err := svc.VehicleAvailability(context.Background(), "2026-01-08", "2026-01-12")
	if err != nil {
		t.Fatalf("VehicleAvailability returned error: %v", err)
	}
	if len(result) != 1 || result[0].Disponibilita != "busy" {
		t.Fatalf("expected the trailer plate to be busy too, got %+v", result)
	}
}

func TestVehicleAvailability_NonOverlappingRangeStaysAvailable(t *testing.T) {
	db := newTestDB(t)
	svc := NewAvailabilityService(db)

	v := models.Vehicle{ID: uuid.New(), Targa: "AB123CD", Active: true}
	if err := db.Create(&v).Error; err != nil {
		t.Fatal(err)
	}
	order := models.Order{
		ID: uuid.New(), ClienteID: uuid.New(), Stato: "VIAGGIO", TargaMotrice: "AB123CD",
		DataRitiro: "2026-01-01", DataConsegna: "2026-01-02",
		ServiziAccessori: []byte("[]"), CostiAccessori: []byte("[]"),
	}
	if err := db.Create(&order).Error; err != nil {
		t.Fatal(err)
	}

	result, err := svc.VehicleAvailability(context.Background(), "2026-02-01", "2026-02-05")
	if err != nil {
		t.Fatalf("VehicleAvailability returned error: %v", err)
	}
	if len(result) != 1 || result[0].Disponibilita != "available" {
		t.Fatalf("expected available for non-overlapping range, got %+v", result)
	}
}

func TestDriverAvailability_UnavailableTakesPriorityOverBusy(t *testing.T) {
	db := newTestDB(t)
	svc := NewAvailabilityService(db)

	driver := models.Driver{ID: uuid.New(), Nome: "Mario", Cognome: "Rossi", Active: true}
	if err := db.Create(&driver).Error; err != nil {
		t.Fatal(err)
	}
	order := models.Order{
		ID: uuid.New(), ClienteID: uuid.New(), Stato: "VIAGGIO", AutistaID: &driver.ID,
		DataRitiro: "2026-01-05", DataConsegna: "2026-01-10",
		ServiziAccessori: []byte("[]"), CostiAccessori: []byte("[]"),
	}
	if err := db.Create(&order).Error; err != nil {
		t.Fatal(err)
	}
	unav := models.DriverUnavailability{
		ID: uuid.New(), AutistaID: driver.ID, DataDa: "2026-01-01", DataA: "2026-01-31", Motivo: "malattia",
	}
	if err := db.Create(&unav).Error; err != nil {
		t.Fatal(err)
	}

	result, err := svc.DriverAvailability(context.Background(), "2026-01-08", "2026-01-12")
	if err != nil {
		t.Fatalf("DriverAvailability returned error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 driver, got %d", len(result))
	}
	if result[0].Disponibilita != "unavailable" || result[0].MotivoIndisponibilita != "malattia" {
		t.Fatalf("expected unavailable/malattia (priority over busy order), got %+v", result[0])
	}
}

func TestDriverAvailability_BusyWhenNoUnavailabilityRecord(t *testing.T) {
	db := newTestDB(t)
	svc := NewAvailabilityService(db)

	driver := models.Driver{ID: uuid.New(), Nome: "Luigi", Cognome: "Verdi", Active: true}
	if err := db.Create(&driver).Error; err != nil {
		t.Fatal(err)
	}
	order := models.Order{
		ID: uuid.New(), ClienteID: uuid.New(), Stato: "VIAGGIO", AutistaID: &driver.ID,
		DataRitiro: "2026-01-05", DataConsegna: "2026-01-10",
		ServiziAccessori: []byte("[]"), CostiAccessori: []byte("[]"),
	}
	if err := db.Create(&order).Error; err != nil {
		t.Fatal(err)
	}

	result, err := svc.DriverAvailability(context.Background(), "2026-01-08", "2026-01-12")
	if err != nil {
		t.Fatalf("DriverAvailability returned error: %v", err)
	}
	if len(result) != 1 || result[0].Disponibilita != "busy" || result[0].MotivoIndisponibilita != "" {
		t.Fatalf("expected busy with empty motivo, got %+v", result[0])
	}
}

func TestDriverAvailability_AvailableWhenNoMatch(t *testing.T) {
	db := newTestDB(t)
	svc := NewAvailabilityService(db)

	driver := models.Driver{ID: uuid.New(), Nome: "Anna", Cognome: "Bianchi", Active: true}
	if err := db.Create(&driver).Error; err != nil {
		t.Fatal(err)
	}

	result, err := svc.DriverAvailability(context.Background(), "2026-01-08", "2026-01-12")
	if err != nil {
		t.Fatalf("DriverAvailability returned error: %v", err)
	}
	if len(result) != 1 || result[0].Disponibilita != "available" {
		t.Fatalf("expected available, got %+v", result[0])
	}
}
