package garages

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"fratelli-feccia/internal/dto"
	"fratelli-feccia/internal/models"
)

const listLimit = 1000

type GarageService struct {
	db *gorm.DB
}

func NewGarageService(db *gorm.DB) *GarageService {
	return &GarageService{db: db}
}

// List has no search parameter, mirroring backend/routers/garages.py (no
// search support, no GET-by-id either).
func (s *GarageService) List(ctx context.Context, includeInactive bool) ([]dto.GarageResponse, error) {
	query := s.db.WithContext(ctx)
	if !includeInactive {
		query = query.Where("active = ?", true)
	}
	var garages []models.Garage
	if err := query.Limit(listLimit).Find(&garages).Error; err != nil {
		return nil, err
	}

	result := make([]dto.GarageResponse, len(garages))
	for i, g := range garages {
		result[i] = toResponse(g)
	}
	return result, nil
}

func (s *GarageService) Create(ctx context.Context, req dto.GarageRequest) (*dto.GarageResponse, error) {
	g := models.Garage{
		ID:        uuid.New(),
		Nome:      req.Nome,
		Indirizzo: req.Indirizzo,
		Citta:     req.Citta,
		Lat:       req.Lat,
		Lng:       req.Lng,
		Note:      req.Note,
		Active:    true,
	}

	if err := s.db.WithContext(ctx).Create(&g).Error; err != nil {
		return nil, err
	}

	resp := toResponse(g)
	return &resp, nil
}

func (s *GarageService) Update(ctx context.Context, id uuid.UUID, req dto.GarageRequest) (*dto.GarageResponse, error) {
	var g models.Garage
	if err := s.db.WithContext(ctx).First(&g, "id = ?", id).Error; err != nil {
		return nil, err
	}

	g.Nome = req.Nome
	g.Indirizzo = req.Indirizzo
	g.Citta = req.Citta
	g.Lat = req.Lat
	g.Lng = req.Lng
	g.Note = req.Note

	if err := s.db.WithContext(ctx).Save(&g).Error; err != nil {
		return nil, err
	}
	if err := s.db.WithContext(ctx).Model(&models.Garage{}).Where("id = ?", id).Update("active", req.Active).Error; err != nil {
		return nil, err
	}
	g.Active = req.Active

	resp := toResponse(g)
	return &resp, nil
}

func (s *GarageService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.db.WithContext(ctx).Model(&models.Garage{}).Where("id = ?", id).Update("active", false).Error
}

func toResponse(g models.Garage) dto.GarageResponse {
	return dto.GarageResponse{
		ID:        g.ID,
		Nome:      g.Nome,
		Indirizzo: g.Indirizzo,
		Citta:     g.Citta,
		Lat:       g.Lat,
		Lng:       g.Lng,
		Note:      g.Note,
		Active:    g.Active,
		CreatedAt: g.CreatedAt,
		UpdatedAt: g.UpdatedAt,
	}
}
