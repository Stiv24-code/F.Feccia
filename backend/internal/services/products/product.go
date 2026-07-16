package products

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"fratelli-feccia/internal/dto"
	"fratelli-feccia/internal/models"
)

const listLimit = 1000

type ProductService struct {
	db *gorm.DB
}

func NewProductService(db *gorm.DB) *ProductService {
	return &ProductService{db: db}
}

func escapeLike(term string) string {
	if len(term) > 100 {
		term = term[:100]
	}
	replacer := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return replacer.Replace(term)
}

// List searches codice OR descrizione, mirroring backend/routers/products.py's `$or`.
func (s *ProductService) List(ctx context.Context, search string) ([]dto.ProductResponse, error) {
	query := s.db.WithContext(ctx).Model(&models.Product{}).Where("active = ?", true)
	if search != "" {
		term := "%" + strings.ToLower(escapeLike(search)) + "%"
		query = query.Where("LOWER(codice) LIKE ? OR LOWER(descrizione) LIKE ?", term, term)
	}

	var products []models.Product
	if err := query.Order("codice ASC").Limit(listLimit).Find(&products).Error; err != nil {
		return nil, err
	}

	result := make([]dto.ProductResponse, len(products))
	for i, p := range products {
		result[i] = toResponse(p)
	}
	return result, nil
}

func (s *ProductService) Create(ctx context.Context, req dto.ProductRequest) (*dto.ProductResponse, error) {
	p := models.Product{
		ID:          uuid.New(),
		Codice:      req.Codice,
		Descrizione: req.Descrizione,
		UnitaMisura: defaultString(req.UnitaMisura, "Kg"),
		Note:        req.Note,
		Active:      true,
	}

	if err := s.db.WithContext(ctx).Create(&p).Error; err != nil {
		return nil, err
	}

	resp := toResponse(p)
	return &resp, nil
}

func (s *ProductService) Update(ctx context.Context, id uuid.UUID, req dto.ProductRequest) (*dto.ProductResponse, error) {
	var p models.Product
	if err := s.db.WithContext(ctx).First(&p, "id = ?", id).Error; err != nil {
		return nil, err
	}

	p.Codice = req.Codice
	p.Descrizione = req.Descrizione
	p.UnitaMisura = defaultString(req.UnitaMisura, "Kg")
	p.Note = req.Note

	if err := s.db.WithContext(ctx).Save(&p).Error; err != nil {
		return nil, err
	}

	resp := toResponse(p)
	return &resp, nil
}

func (s *ProductService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.db.WithContext(ctx).Model(&models.Product{}).Where("id = ?", id).Update("active", false).Error
}

func defaultString(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func toResponse(p models.Product) dto.ProductResponse {
	return dto.ProductResponse{
		ID:          p.ID,
		Codice:      p.Codice,
		Descrizione: p.Descrizione,
		UnitaMisura: p.UnitaMisura,
		Note:        p.Note,
		Active:      p.Active,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}
