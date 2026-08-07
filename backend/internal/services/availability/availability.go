// Package availability ports backend/routers/availability.py: read-only
// endpoints cross-referencing motrici/semirimorchi/drivers against active
// orders (and, for drivers, driver_unavailability) for a given date range.
package availability

import (
	"context"

	"gorm.io/gorm"

	"fratelli-feccia/internal/dto"
	"fratelli-feccia/internal/models"
	"fratelli-feccia/internal/services/drivers"
	"fratelli-feccia/internal/services/motrici"
	"fratelli-feccia/internal/services/semirimorchi"
)

type AvailabilityService struct {
	db *gorm.DB
}

func NewAvailabilityService(db *gorm.DB) *AvailabilityService {
	return &AvailabilityService{db: db}
}

// MotriceAvailability mirrors the former vehicle_availability for the
// tractor half: a motrice is "busy" if it's referenced by a VIAGGIO order
// whose ritiro/consegna window overlaps [dataDa, dataA].
func (s *AvailabilityService) MotriceAvailability(ctx context.Context, dataDa, dataA string) ([]dto.MotriceAvailabilityResponse, error) {
	var all []models.Motrice
	if err := s.db.WithContext(ctx).Where("active = ?", true).Find(&all).Error; err != nil {
		return nil, err
	}

	var busyOrders []models.Order
	if err := s.db.WithContext(ctx).
		Where("stato = ? AND motrice_id IS NOT NULL AND data_ritiro <= ? AND data_consegna >= ?", "VIAGGIO", dataA, dataDa).
		Find(&busyOrders).Error; err != nil {
		return nil, err
	}
	busyIDs := make(map[string]bool, len(busyOrders))
	for _, o := range busyOrders {
		if o.MotriceID != nil {
			busyIDs[o.MotriceID.String()] = true
		}
	}

	result := make([]dto.MotriceAvailabilityResponse, len(all))
	for i, m := range all {
		status := "available"
		if busyIDs[m.ID.String()] {
			status = "busy"
		}
		result[i] = dto.MotriceAvailabilityResponse{MotriceResponse: motrici.ToResponse(m), Disponibilita: status}
	}
	return result, nil
}

// SemirimorchioAvailability mirrors MotriceAvailability for the trailer half.
func (s *AvailabilityService) SemirimorchioAvailability(ctx context.Context, dataDa, dataA string) ([]dto.SemirimorchioAvailabilityResponse, error) {
	var all []models.Semirimorchio
	if err := s.db.WithContext(ctx).Where("active = ?", true).Find(&all).Error; err != nil {
		return nil, err
	}

	var busyOrders []models.Order
	if err := s.db.WithContext(ctx).
		Where("stato = ? AND semirimorchio_id IS NOT NULL AND data_ritiro <= ? AND data_consegna >= ?", "VIAGGIO", dataA, dataDa).
		Find(&busyOrders).Error; err != nil {
		return nil, err
	}
	busyIDs := make(map[string]bool, len(busyOrders))
	for _, o := range busyOrders {
		if o.SemirimorchioID != nil {
			busyIDs[o.SemirimorchioID.String()] = true
		}
	}

	result := make([]dto.SemirimorchioAvailabilityResponse, len(all))
	for i, r := range all {
		status := "available"
		if busyIDs[r.ID.String()] {
			status = "busy"
		}
		result[i] = dto.SemirimorchioAvailabilityResponse{SemirimorchioResponse: semirimorchi.ToResponse(r), Disponibilita: status}
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
