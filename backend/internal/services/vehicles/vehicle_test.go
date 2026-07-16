package vehicles

import (
	"context"
	"errors"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"fratelli-feccia/internal/dto"
	"fratelli-feccia/internal/models"
	"fratelli-feccia/pkg/utils"
)

func assertAPIError(t *testing.T, err error, code int) {
	t.Helper()
	var apiErr utils.APIError
	if err == nil {
		t.Fatalf("expected an APIError %d, got nil", code)
	}
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected an APIError %d, got %v (%T)", code, err, err)
	}
	if apiErr.Code != code {
		t.Fatalf("expected APIError code %d, got %d (%s)", code, apiErr.Code, apiErr.Message)
	}
}

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}
	if err := db.AutoMigrate(&models.Vehicle{}, &models.GPSHistoryEntry{}, &models.TemperatureReading{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	return db
}

func TestVehicleService_CreateDefaultsAndSearch(t *testing.T) {
	ctx := context.Background()
	svc := NewVehicleService(newTestDB(t))

	created, err := svc.Create(ctx, dto.VehicleRequest{Targa: "AB123CD"})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if created.TipoVeicolo != "motrice" || created.Scompartature != 1 {
		t.Fatalf("expected defaults motrice/1, got %q/%d", created.TipoVeicolo, created.Scompartature)
	}

	list, err := svc.List(ctx, "ab123")
	if err != nil || len(list) != 1 {
		t.Fatalf("expected search to match, got %v (err=%v)", list, err)
	}
}

func TestVehicleService_UpdateDoesNotTouchTelemetry(t *testing.T) {
	ctx := context.Background()
	svc := NewVehicleService(newTestDB(t))

	created, err := svc.Create(ctx, dto.VehicleRequest{Targa: "AB123CD"})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if _, err := svc.UpdateGPSByID(ctx, created.ID.String(), dto.VehicleGPSUpdateRequest{Lat: 45.1, Lng: 9.2}); err != nil {
		t.Fatalf("UpdateGPSByID returned error: %v", err)
	}

	updated, err := svc.Update(ctx, created.ID, dto.VehicleRequest{Targa: "AB123CD", Marca: "Iveco"})
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
	if updated.Marca != "Iveco" {
		t.Fatalf("expected marca updated, got %q", updated.Marca)
	}
	if updated.LastLat != 45.1 || updated.LastLng != 9.2 {
		t.Fatalf("expected GPS telemetry to survive a CRUD update, got %+v", updated)
	}
}

func TestVehicleService_UpdateGPSByID_FallsBackToTarga(t *testing.T) {
	ctx := context.Background()
	svc := NewVehicleService(newTestDB(t))

	created, err := svc.Create(ctx, dto.VehicleRequest{Targa: "XY999ZZ", GpsTrackerTipo: "Garmin"})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	result, err := svc.UpdateGPSByID(ctx, "XY999ZZ", dto.VehicleGPSUpdateRequest{Lat: 41.9, Lng: 12.5, SpeedKmh: 80})
	if err != nil {
		t.Fatalf("UpdateGPSByID (by targa) returned error: %v", err)
	}
	if result.GpsSource != "provider_garmin" {
		t.Fatalf("expected gps_source derived from gps_tracker_tipo, got %q", result.GpsSource)
	}

	history, err := svc.GetGPSHistory(ctx, created.ID.String(), 10)
	if err != nil || len(history) != 1 {
		t.Fatalf("expected 1 history entry, got %v (err=%v)", history, err)
	}
}

func TestVehicleService_IngestGPSWebhook_ValidatesAndUsesVendorSource(t *testing.T) {
	ctx := context.Background()
	svc := NewVehicleService(newTestDB(t))

	if _, err := svc.Create(ctx, dto.VehicleRequest{Targa: "AB123CD"}); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	_, err := svc.IngestGPSWebhook(ctx, "garmin", dto.GPSWebhookPayload{Targa: "ab123cd", Lat: 999, Lng: 9})
	assertAPIError(t, err, 400)

	result, err := svc.IngestGPSWebhook(ctx, "garmin", dto.GPSWebhookPayload{Targa: "ab123cd", Lat: 45, Lng: 9})
	if err != nil {
		t.Fatalf("IngestGPSWebhook returned error: %v", err)
	}
	if result.GpsSource != "provider_garmin" {
		t.Fatalf("expected vendor-derived gps_source, got %q", result.GpsSource)
	}

	_, err = svc.IngestGPSWebhook(ctx, "garmin", dto.GPSWebhookPayload{Targa: "nonexistent", Lat: 45, Lng: 9})
	assertAPIError(t, err, 404)
}

func TestVehicleService_TemperatureWebhookAndAlert(t *testing.T) {
	ctx := context.Background()
	svc := NewVehicleService(newTestDB(t))

	tmin, tmax := 2.0, 8.0
	created, err := svc.Create(ctx, dto.VehicleRequest{Targa: "AB123CD"})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if _, err := svc.SetTemperatureThresholds(ctx, created.ID, dto.TemperatureThresholdsRequest{TempMin: &tmin, TempMax: &tmax}); err != nil {
		t.Fatalf("SetTemperatureThresholds returned error: %v", err)
	}

	result, err := svc.IngestTemperatureWebhook(ctx, "sensoneo", dto.TemperatureWebhookRequest{Targa: "AB123CD", TempCelsius: 12})
	if err != nil {
		t.Fatalf("IngestTemperatureWebhook returned error: %v", err)
	}
	if !result.OutOfRange || !result.Alert {
		t.Fatalf("expected out-of-range alert for 12C outside [2,8], got %+v", result)
	}

	history, err := svc.GetTemperatureHistory(ctx, created.ID, 10, false)
	if err != nil || len(history) != 1 {
		t.Fatalf("expected 1 temperature reading, got %v (err=%v)", history, err)
	}

	onlyAlerts, err := svc.GetTemperatureHistory(ctx, created.ID, 10, true)
	if err != nil || len(onlyAlerts) != 1 {
		t.Fatalf("expected the alert reading to also match only_alerts filter, got %v (err=%v)", onlyAlerts, err)
	}
}

func TestVehicleService_SetTemperatureThresholds_ValidatesRange(t *testing.T) {
	ctx := context.Background()
	svc := NewVehicleService(newTestDB(t))

	created, err := svc.Create(ctx, dto.VehicleRequest{Targa: "AB123CD"})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	tmin, tmax := 10.0, 5.0
	_, err = svc.SetTemperatureThresholds(ctx, created.ID, dto.TemperatureThresholdsRequest{TempMin: &tmin, TempMax: &tmax})
	assertAPIError(t, err, 400)

	_, err = svc.SetTemperatureThresholds(ctx, created.ID, dto.TemperatureThresholdsRequest{})
	assertAPIError(t, err, 400)

	_, err = svc.SetTemperatureThresholds(ctx, uuid.New(), dto.TemperatureThresholdsRequest{TempMin: &tmin})
	assertAPIError(t, err, 404)
}

func TestVehicleService_Delete_IsLogical(t *testing.T) {
	ctx := context.Background()
	svc := NewVehicleService(newTestDB(t))

	created, err := svc.Create(ctx, dto.VehicleRequest{Targa: "AB123CD"})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if err := svc.Delete(ctx, created.ID); err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}
	list, err := svc.List(ctx, "")
	if err != nil || len(list) != 0 {
		t.Fatalf("expected 0 active vehicles after delete, got %v (err=%v)", list, err)
	}
}
