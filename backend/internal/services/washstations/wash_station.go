package washstations

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"fratelli-feccia/internal/dto"
	"fratelli-feccia/internal/models"
)

const listLimit = 1000

type WashStationService struct {
	db *gorm.DB
}

func NewWashStationService(db *gorm.DB) *WashStationService {
	return &WashStationService{db: db}
}

func (s *WashStationService) List(ctx context.Context, includeInactive bool) ([]dto.WashStationResponse, error) {
	query := s.db.WithContext(ctx)
	if !includeInactive {
		query = query.Where("active = ?", true)
	}
	var stations []models.WashStation
	if err := query.Limit(listLimit).Find(&stations).Error; err != nil {
		return nil, err
	}

	result := make([]dto.WashStationResponse, len(stations))
	for i, w := range stations {
		result[i] = toResponse(w)
	}
	return result, nil
}

func (s *WashStationService) Create(ctx context.Context, req dto.WashStationRequest) (*dto.WashStationResponse, error) {
	w := models.WashStation{
		ID:        uuid.New(),
		Nome:      req.Nome,
		Tipo:      req.Tipo,
		Indirizzo: req.Indirizzo,
		Citta:     req.Citta,
		Lat:       req.Lat,
		Lng:       req.Lng,
		Note:      req.Note,
		Active:    true,
	}

	if err := s.db.WithContext(ctx).Create(&w).Error; err != nil {
		return nil, err
	}

	resp := toResponse(w)
	return &resp, nil
}

func (s *WashStationService) Update(ctx context.Context, id uuid.UUID, req dto.WashStationRequest) (*dto.WashStationResponse, error) {
	var w models.WashStation
	if err := s.db.WithContext(ctx).First(&w, "id = ?", id).Error; err != nil {
		return nil, err
	}

	w.Nome = req.Nome
	w.Tipo = req.Tipo
	w.Indirizzo = req.Indirizzo
	w.Citta = req.Citta
	w.Lat = req.Lat
	w.Lng = req.Lng
	w.Note = req.Note

	if err := s.db.WithContext(ctx).Save(&w).Error; err != nil {
		return nil, err
	}
	if err := s.db.WithContext(ctx).Model(&models.WashStation{}).Where("id = ?", id).Update("active", req.Active).Error; err != nil {
		return nil, err
	}
	w.Active = req.Active

	resp := toResponse(w)
	return &resp, nil
}

func (s *WashStationService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.db.WithContext(ctx).Model(&models.WashStation{}).Where("id = ?", id).Update("active", false).Error
}

func toResponse(w models.WashStation) dto.WashStationResponse {
	return dto.WashStationResponse{
		ID:        w.ID,
		Nome:      w.Nome,
		Tipo:      w.Tipo,
		Indirizzo: w.Indirizzo,
		Citta:     w.Citta,
		Lat:       w.Lat,
		Lng:       w.Lng,
		Note:      w.Note,
		Active:    w.Active,
		CreatedAt: w.CreatedAt,
		UpdatedAt: w.UpdatedAt,
	}
}
