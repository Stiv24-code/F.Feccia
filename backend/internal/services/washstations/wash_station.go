package washstations

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"fratelli-feccia/internal/dto"
	"fratelli-feccia/internal/models"
	"fratelli-feccia/pkg/utils"
)

type WashStationService struct {
	db *gorm.DB
}

func NewWashStationService(db *gorm.DB) *WashStationService {
	return &WashStationService{db: db}
}

func (s *WashStationService) List(ctx context.Context, includeInactive bool, page utils.PageParams) ([]dto.WashStationResponse, int64, error) {
	query := s.db.WithContext(ctx).Model(&models.WashStation{})
	if !includeInactive {
		query = query.Where("active = ?", true)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var stations []models.WashStation
	if err := query.Order("nome ASC").Offset(page.Offset()).Limit(page.Limit).Find(&stations).Error; err != nil {
		return nil, 0, err
	}

	result := make([]dto.WashStationResponse, len(stations))
	for i, w := range stations {
		result[i] = toResponse(w)
	}
	return result, total, nil
}

// ListAll ritorna tutti i punti di lavaggio attivi senza paginazione — usata
// dai picker (es. select "punto di lavaggio" nell'assegnazione trasporto)
// che devono poter ordinare/filtrare per vicinanza sull'elenco intero invece
// di un sottoinsieme troncato a un limite arbitrario.
func (s *WashStationService) ListAll(ctx context.Context) ([]dto.WashStationResponse, error) {
	var stations []models.WashStation
	if err := s.db.WithContext(ctx).Where("active = ?", true).Order("nome ASC").Find(&stations).Error; err != nil {
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
		OrarioDa:  req.OrarioDa,
		OrarioA:   req.OrarioA,
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
	w.OrarioDa = req.OrarioDa
	w.OrarioA = req.OrarioA
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
		OrarioDa:  w.OrarioDa,
		OrarioA:   w.OrarioA,
		Note:      w.Note,
		Active:    w.Active,
		CreatedAt: w.CreatedAt,
		UpdatedAt: w.UpdatedAt,
	}
}
