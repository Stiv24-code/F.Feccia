package destinations

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"fratelli-feccia/internal/dto"
	"fratelli-feccia/internal/models"
)

const listLimit = 1000

type DestinationService struct {
	db *gorm.DB
}

func NewDestinationService(db *gorm.DB) *DestinationService {
	return &DestinationService{db: db}
}

func escapeLike(term string) string {
	if len(term) > 100 {
		term = term[:100]
	}
	replacer := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return replacer.Replace(term)
}

func (s *DestinationService) List(ctx context.Context, search string) ([]dto.DestinationResponse, error) {
	query := s.db.WithContext(ctx).Model(&models.Destination{}).Where("active = ?", true)
	if search != "" {
		query = query.Where("LOWER(nome) LIKE ?", "%"+strings.ToLower(escapeLike(search))+"%")
	}

	var destinations []models.Destination
	if err := query.Order("nome ASC").Limit(listLimit).Find(&destinations).Error; err != nil {
		return nil, err
	}

	result := make([]dto.DestinationResponse, len(destinations))
	for i, d := range destinations {
		result[i] = toResponse(d)
	}
	return result, nil
}

func (s *DestinationService) GetByID(ctx context.Context, id uuid.UUID) (*dto.DestinationResponse, error) {
	var d models.Destination
	if err := s.db.WithContext(ctx).First(&d, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	resp := toResponse(d)
	return &resp, nil
}

func (s *DestinationService) Create(ctx context.Context, req dto.DestinationRequest) (*dto.DestinationResponse, error) {
	d := models.Destination{
		ID:             uuid.New(),
		Nome:           req.Nome,
		Indirizzo:      req.Indirizzo,
		Citta:          req.Citta,
		Cap:            req.Cap,
		Provincia:      req.Provincia,
		Nazione:        defaultString(req.Nazione, "Italia"),
		VincoliScarico: req.VincoliScarico,
		Note:           req.Note,
		Active:         true,
	}

	if err := s.db.WithContext(ctx).Create(&d).Error; err != nil {
		return nil, err
	}

	resp := toResponse(d)
	return &resp, nil
}

func (s *DestinationService) Update(ctx context.Context, id uuid.UUID, req dto.DestinationRequest) (*dto.DestinationResponse, error) {
	var d models.Destination
	if err := s.db.WithContext(ctx).First(&d, "id = ?", id).Error; err != nil {
		return nil, err
	}

	d.Nome = req.Nome
	d.Indirizzo = req.Indirizzo
	d.Citta = req.Citta
	d.Cap = req.Cap
	d.Provincia = req.Provincia
	d.Nazione = defaultString(req.Nazione, "Italia")
	d.VincoliScarico = req.VincoliScarico
	d.Note = req.Note

	if err := s.db.WithContext(ctx).Save(&d).Error; err != nil {
		return nil, err
	}

	resp := toResponse(d)
	return &resp, nil
}

func (s *DestinationService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.db.WithContext(ctx).Model(&models.Destination{}).Where("id = ?", id).Update("active", false).Error
}

func defaultString(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func toResponse(d models.Destination) dto.DestinationResponse {
	return dto.DestinationResponse{
		ID:             d.ID,
		Nome:           d.Nome,
		Indirizzo:      d.Indirizzo,
		Citta:          d.Citta,
		Cap:            d.Cap,
		Provincia:      d.Provincia,
		Nazione:        d.Nazione,
		VincoliScarico: d.VincoliScarico,
		Note:           d.Note,
		Active:         d.Active,
		CreatedAt:      d.CreatedAt,
		UpdatedAt:      d.UpdatedAt,
	}
}
