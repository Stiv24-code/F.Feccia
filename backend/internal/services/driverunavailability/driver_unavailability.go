package driverunavailability

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"fratelli-feccia/internal/dto"
	"fratelli-feccia/internal/models"
)

const listLimit = 500

type DriverUnavailabilityService struct {
	db *gorm.DB
}

func NewDriverUnavailabilityService(db *gorm.DB) *DriverUnavailabilityService {
	return &DriverUnavailabilityService{db: db}
}

// List mirrors backend/routers/driver_unavailability.py: optional filter by
// autista_id, optional date-range overlap filter (existing.data_da <= dataA
// AND existing.data_a >= dataDa), sorted by data_da descending.
func (s *DriverUnavailabilityService) List(ctx context.Context, autistaID uuid.UUID, dataDa, dataA string) ([]dto.DriverUnavailabilityResponse, error) {
	query := s.db.WithContext(ctx).Model(&models.DriverUnavailability{})
	if autistaID != uuid.Nil {
		query = query.Where("autista_id = ?", autistaID)
	}
	if dataDa != "" && dataA != "" {
		query = query.Where("data_da <= ? AND data_a >= ?", dataA, dataDa)
	}

	var items []models.DriverUnavailability
	if err := query.Order("data_da DESC").Limit(listLimit).Find(&items).Error; err != nil {
		return nil, err
	}

	result := make([]dto.DriverUnavailabilityResponse, len(items))
	for i, item := range items {
		result[i] = toResponse(item)
	}
	return result, nil
}

func (s *DriverUnavailabilityService) Create(ctx context.Context, req dto.DriverUnavailabilityRequest) (*dto.DriverUnavailabilityResponse, error) {
	item := models.DriverUnavailability{
		ID:          uuid.New(),
		AutistaID:   req.AutistaID,
		AutistaNome: req.AutistaNome,
		DataDa:      req.DataDa,
		DataA:       req.DataA,
		Motivo:      defaultString(req.Motivo, "ferie"),
		Note:        req.Note,
	}

	if err := s.db.WithContext(ctx).Create(&item).Error; err != nil {
		return nil, err
	}

	resp := toResponse(item)
	return &resp, nil
}

// Delete is a real hard delete, mirroring the Python original (no `active`
// flag on this entity). Returns gorm.ErrRecordNotFound if nothing matched,
// so the handler can translate it to 404 like Python's `deleted_count == 0`.
func (s *DriverUnavailabilityService) Delete(ctx context.Context, id uuid.UUID) error {
	result := s.db.WithContext(ctx).Delete(&models.DriverUnavailability{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func defaultString(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func toResponse(item models.DriverUnavailability) dto.DriverUnavailabilityResponse {
	return dto.DriverUnavailabilityResponse{
		ID:          item.ID,
		AutistaID:   item.AutistaID,
		AutistaNome: item.AutistaNome,
		DataDa:      item.DataDa,
		DataA:       item.DataA,
		Motivo:      item.Motivo,
		Note:        item.Note,
		CreatedAt:   item.CreatedAt,
	}
}
