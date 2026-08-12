package destinations

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"fratelli-feccia/internal/dto"
	"fratelli-feccia/internal/models"
	"fratelli-feccia/pkg/utils"
)

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

func (s *DestinationService) List(ctx context.Context, search string, includeInactive bool, page utils.PageParams) ([]dto.DestinationResponse, int64, error) {
	query := s.db.WithContext(ctx).Model(&models.Destination{})
	if !includeInactive {
		query = query.Where("active = ?", true)
	}
	if search != "" {
		query = query.Where("LOWER(nome) LIKE ?", "%"+strings.ToLower(escapeLike(search))+"%")
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var destinations []models.Destination
	if err := query.Order("nome ASC").Offset(page.Offset()).Limit(page.Limit).Find(&destinations).Error; err != nil {
		return nil, 0, err
	}

	result := make([]dto.DestinationResponse, len(destinations))
	for i, d := range destinations {
		result[i] = toResponse(d)
	}
	return result, total, nil
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
		Lat:            req.Lat,
		Lng:            req.Lng,
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
	d.Lat = req.Lat
	d.Lng = req.Lng
	d.VincoliScarico = req.VincoliScarico
	d.Note = req.Note

	if err := s.db.WithContext(ctx).Save(&d).Error; err != nil {
		return nil, err
	}
	// Bool column with a `default:true` tag: GORM's create-time zero-skip
	// logic doesn't apply here (Save/Update, not Create), but a plain field
	// assignment on Save was already a source of surprises in this codebase
	// (see the batch-create gotcha in map_test.go) — an explicit targeted
	// update, same as Delete() below, removes any doubt.
	if err := s.db.WithContext(ctx).Model(&models.Destination{}).Where("id = ?", id).Update("active", req.Active).Error; err != nil {
		return nil, err
	}
	d.Active = req.Active

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
		Lat:            d.Lat,
		Lng:            d.Lng,
		VincoliScarico: d.VincoliScarico,
		Note:           d.Note,
		Active:         d.Active,
		CreatedAt:      d.CreatedAt,
		UpdatedAt:      d.UpdatedAt,
	}
}
