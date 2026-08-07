package motrici

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"fratelli-feccia/internal/dto"
	"fratelli-feccia/internal/models"
)

const listLimit = 1000

type MotriceService struct {
	db *gorm.DB
}

func NewMotriceService(db *gorm.DB) *MotriceService {
	return &MotriceService{db: db}
}

func escapeLike(term string) string {
	if len(term) > 100 {
		term = term[:100]
	}
	replacer := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return replacer.Replace(term)
}

func (s *MotriceService) List(ctx context.Context, search string) ([]dto.MotriceResponse, error) {
	query := s.db.WithContext(ctx).Model(&models.Motrice{}).Where("active = ?", true)
	if search != "" {
		query = query.Where("LOWER(targa) LIKE ?", "%"+strings.ToLower(escapeLike(search))+"%")
	}
	var items []models.Motrice
	if err := query.Order("targa ASC").Limit(listLimit).Find(&items).Error; err != nil {
		return nil, err
	}
	result := make([]dto.MotriceResponse, len(items))
	for i, m := range items {
		result[i] = ToResponse(m)
	}
	return result, nil
}

func (s *MotriceService) Create(ctx context.Context, req dto.MotriceRequest) (*dto.MotriceResponse, error) {
	m := models.Motrice{
		ID:        uuid.New(),
		Targa:     req.Targa,
		Marca:     req.Marca,
		Modello:   req.Modello,
		Anno:      req.Anno,
		PortataKg: req.PortataKg,
		Note:      req.Note,
		Active:    true,
	}
	if err := s.db.WithContext(ctx).Create(&m).Error; err != nil {
		return nil, err
	}
	resp := ToResponse(m)
	return &resp, nil
}

func (s *MotriceService) Update(ctx context.Context, id uuid.UUID, req dto.MotriceRequest) (*dto.MotriceResponse, error) {
	var m models.Motrice
	if err := s.db.WithContext(ctx).First(&m, "id = ?", id).Error; err != nil {
		return nil, err
	}
	m.Targa = req.Targa
	m.Marca = req.Marca
	m.Modello = req.Modello
	m.Anno = req.Anno
	m.PortataKg = req.PortataKg
	m.Note = req.Note

	if err := s.db.WithContext(ctx).Save(&m).Error; err != nil {
		return nil, err
	}
	resp := ToResponse(m)
	return &resp, nil
}

func (s *MotriceService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.db.WithContext(ctx).Model(&models.Motrice{}).Where("id = ?", id).Update("active", false).Error
}

func ToResponse(m models.Motrice) dto.MotriceResponse {
	return dto.MotriceResponse{
		ID:        m.ID,
		Targa:     m.Targa,
		Marca:     m.Marca,
		Modello:   m.Modello,
		Anno:      m.Anno,
		PortataKg: m.PortataKg,
		Note:      m.Note,
		Active:    m.Active,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}
