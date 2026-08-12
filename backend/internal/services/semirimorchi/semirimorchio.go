package semirimorchi

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"fratelli-feccia/internal/dto"
	"fratelli-feccia/internal/models"
	"fratelli-feccia/pkg/utils"
)

type SemirimorchioService struct {
	db *gorm.DB
}

func NewSemirimorchioService(db *gorm.DB) *SemirimorchioService {
	return &SemirimorchioService{db: db}
}

func escapeLike(term string) string {
	if len(term) > 100 {
		term = term[:100]
	}
	replacer := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return replacer.Replace(term)
}

func (s *SemirimorchioService) List(ctx context.Context, search string, page utils.PageParams) ([]dto.SemirimorchioResponse, int64, error) {
	query := s.db.WithContext(ctx).Model(&models.Semirimorchio{}).Where("active = ?", true)
	if search != "" {
		query = query.Where("LOWER(targa) LIKE ?", "%"+strings.ToLower(escapeLike(search))+"%")
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []models.Semirimorchio
	if err := query.Order("targa ASC").Offset(page.Offset()).Limit(page.Limit).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	result := make([]dto.SemirimorchioResponse, len(items))
	for i, r := range items {
		result[i] = ToResponse(r)
	}
	return result, total, nil
}

func (s *SemirimorchioService) Create(ctx context.Context, req dto.SemirimorchioRequest) (*dto.SemirimorchioResponse, error) {
	r := models.Semirimorchio{
		ID:            uuid.New(),
		Targa:         req.Targa,
		Tipo:          req.Tipo,
		Scompartature: defaultInt(req.Scompartature, 1),
		PortataKg:     req.PortataKg,
		Note:          req.Note,
		Active:        true,
	}
	if err := s.db.WithContext(ctx).Create(&r).Error; err != nil {
		return nil, err
	}
	resp := ToResponse(r)
	return &resp, nil
}

func (s *SemirimorchioService) Update(ctx context.Context, id uuid.UUID, req dto.SemirimorchioRequest) (*dto.SemirimorchioResponse, error) {
	var r models.Semirimorchio
	if err := s.db.WithContext(ctx).First(&r, "id = ?", id).Error; err != nil {
		return nil, err
	}
	r.Targa = req.Targa
	r.Tipo = req.Tipo
	r.Scompartature = defaultInt(req.Scompartature, 1)
	r.PortataKg = req.PortataKg
	r.Note = req.Note

	if err := s.db.WithContext(ctx).Save(&r).Error; err != nil {
		return nil, err
	}
	resp := ToResponse(r)
	return &resp, nil
}

func (s *SemirimorchioService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.db.WithContext(ctx).Model(&models.Semirimorchio{}).Where("id = ?", id).Update("active", false).Error
}

func defaultInt(value, fallback int) int {
	if value == 0 {
		return fallback
	}
	return value
}

func ToResponse(r models.Semirimorchio) dto.SemirimorchioResponse {
	return dto.SemirimorchioResponse{
		ID:            r.ID,
		Targa:         r.Targa,
		Tipo:          r.Tipo,
		Scompartature: r.Scompartature,
		PortataKg:     r.PortataKg,
		Note:          r.Note,
		Active:        r.Active,
		CreatedAt:     r.CreatedAt,
		UpdatedAt:     r.UpdatedAt,
	}
}
