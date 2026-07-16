package carriers

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"fratelli-feccia/internal/dto"
	"fratelli-feccia/internal/models"
)

const listLimit = 1000

type CarrierService struct {
	db *gorm.DB
}

func NewCarrierService(db *gorm.DB) *CarrierService {
	return &CarrierService{db: db}
}

func escapeLike(term string) string {
	if len(term) > 100 {
		term = term[:100]
	}
	replacer := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return replacer.Replace(term)
}

// List and Create/Update/Delete mirror backend/routers/carriers.py — note
// there is no GET-by-id endpoint there, so this service doesn't expose one.

func (s *CarrierService) List(ctx context.Context, search string) ([]dto.CarrierResponse, error) {
	query := s.db.WithContext(ctx).Model(&models.Carrier{}).Where("active = ?", true)
	if search != "" {
		query = query.Where("LOWER(ragione_sociale) LIKE ?", "%"+strings.ToLower(escapeLike(search))+"%")
	}

	var carriers []models.Carrier
	if err := query.Order("ragione_sociale ASC").Limit(listLimit).Find(&carriers).Error; err != nil {
		return nil, err
	}

	result := make([]dto.CarrierResponse, len(carriers))
	for i, c := range carriers {
		result[i] = toResponse(c)
	}
	return result, nil
}

func (s *CarrierService) Create(ctx context.Context, req dto.CarrierRequest) (*dto.CarrierResponse, error) {
	c := models.Carrier{
		ID:             uuid.New(),
		RagioneSociale: req.RagioneSociale,
		PartitaIva:     req.PartitaIva,
		Indirizzo:      req.Indirizzo,
		Citta:          req.Citta,
		Telefono:       req.Telefono,
		Email:          req.Email,
		Note:           req.Note,
		Active:         true,
	}

	if err := s.db.WithContext(ctx).Create(&c).Error; err != nil {
		return nil, err
	}

	resp := toResponse(c)
	return &resp, nil
}

func (s *CarrierService) Update(ctx context.Context, id uuid.UUID, req dto.CarrierRequest) (*dto.CarrierResponse, error) {
	var c models.Carrier
	if err := s.db.WithContext(ctx).First(&c, "id = ?", id).Error; err != nil {
		return nil, err
	}

	c.RagioneSociale = req.RagioneSociale
	c.PartitaIva = req.PartitaIva
	c.Indirizzo = req.Indirizzo
	c.Citta = req.Citta
	c.Telefono = req.Telefono
	c.Email = req.Email
	c.Note = req.Note

	if err := s.db.WithContext(ctx).Save(&c).Error; err != nil {
		return nil, err
	}

	resp := toResponse(c)
	return &resp, nil
}

func (s *CarrierService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.db.WithContext(ctx).Model(&models.Carrier{}).Where("id = ?", id).Update("active", false).Error
}

func toResponse(c models.Carrier) dto.CarrierResponse {
	return dto.CarrierResponse{
		ID:             c.ID,
		RagioneSociale: c.RagioneSociale,
		PartitaIva:     c.PartitaIva,
		Indirizzo:      c.Indirizzo,
		Citta:          c.Citta,
		Telefono:       c.Telefono,
		Email:          c.Email,
		Note:           c.Note,
		Active:         c.Active,
		CreatedAt:      c.CreatedAt,
		UpdatedAt:      c.UpdatedAt,
	}
}
