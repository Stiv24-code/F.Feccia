// Package availability ports backend/routers/availability.py: two read-only
// endpoints cross-referencing vehicles/drivers against active orders (and,
// for drivers, driver_unavailability) for a given date range.
package availability

import (
	"context"

	"gorm.io/gorm"

	"fratelli-feccia/internal/dto"
	"fratelli-feccia/internal/models"
	"fratelli-feccia/internal/services/drivers"
	"fratelli-feccia/internal/services/vehicles"
)

type AvailabilityService struct {
	db *gorm.DB
}

func NewAvailabilityService(db *gorm.DB) *AvailabilityService {
	return &AvailabilityService{db: db}
}

// VehicleAvailability mirrors vehicle_availability: a vehicle is "busy" if
// its targa (motrice or rimorchio) appears on a VIAGGIO order whose
// ritiro/consegna window overlaps [dataDa, dataA].
func (s *AvailabilityService) VehicleAvailability(ctx context.Context, dataDa, dataA string) ([]dto.VehicleAvailabilityResponse, error) {
	var allVehicles []models.Vehicle
	if err := s.db.WithContext(ctx).Where("active = ?", true).Find(&allVehicles).Error; err != nil {
		return nil, err
	}

	var busyOrders []models.Order
	if err := s.db.WithContext(ctx).
		Where("stato = ? AND targa_motrice <> ? AND data_ritiro <= ? AND data_consegna >= ?", "VIAGGIO", "", dataA, dataDa).
		Find(&busyOrders).Error; err != nil {
		return nil, err
	}

	busyPlates := make(map[string]bool, len(busyOrders)*2)
	for _, o := range busyOrders {
		if o.TargaMotrice != "" {
			busyPlates[o.TargaMotrice] = true
		}
		if o.TargaRimorchio != "" {
			busyPlates[o.TargaRimorchio] = true
		}
	}

	result := make([]dto.VehicleAvailabilityResponse, len(allVehicles))
	for i, v := range allVehicles {
		status := "available"
		if busyPlates[v.Targa] {
			status = "busy"
		}
		result[i] = dto.VehicleAvailabilityResponse{VehicleResponse: vehicles.ToResponse(v), Disponibilita: status}
	}
	return result, nil
}

// DriverAvailability mirrors driver_availability: unavailability (leave/sick)
// takes priority over "busy" (an active VIAGGIO order), which takes priority
// over "available".
func (s *AvailabilityService) DriverAvailability(ctx context.Context, dataDa, dataA string) ([]dto.DriverAvailabilityResponse, error) {
	var allDrivers []models.Driver
	if err := s.db.WithContext(ctx).Where("active = ?", true).Find(&allDrivers).Error; err != nil {
		return nil, err
	}

	var busyOrders []models.Order
	if err := s.db.WithContext(ctx).
		Where("stato = ? AND autista_id IS NOT NULL AND data_ritiro <= ? AND data_consegna >= ?", "VIAGGIO", dataA, dataDa).
		Find(&busyOrders).Error; err != nil {
		return nil, err
	}
	busyDriverIDs := make(map[string]bool, len(busyOrders))
	for _, o := range busyOrders {
		if o.AutistaID != nil {
			busyDriverIDs[o.AutistaID.String()] = true
		}
	}

	var unavail []models.DriverUnavailability
	if err := s.db.WithContext(ctx).
		Where("data_da <= ? AND data_a >= ?", dataA, dataDa).
		Find(&unavail).Error; err != nil {
		return nil, err
	}
	unavailableReasons := make(map[string]string, len(unavail))
	for _, u := range unavail {
		reason := u.Motivo
		if reason == "" {
			reason = "ferie"
		}
		unavailableReasons[u.AutistaID.String()] = reason
	}

	result := make([]dto.DriverAvailabilityResponse, len(allDrivers))
	for i, d := range allDrivers {
		idStr := d.ID.String()
		var status, motivo string
		if reason, ok := unavailableReasons[idStr]; ok {
			status, motivo = "unavailable", reason
		} else if busyDriverIDs[idStr] {
			status, motivo = "busy", ""
		} else {
			status, motivo = "available", ""
		}
		result[i] = dto.DriverAvailabilityResponse{DriverResponse: drivers.ToResponse(d), Disponibilita: status, MotivoIndisponibilita: motivo}
	}
	return result, nil
}
