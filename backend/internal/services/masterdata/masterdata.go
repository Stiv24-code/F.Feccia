package masterdata

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"fratelli-feccia/internal/dto"
	"fratelli-feccia/internal/models"
)

// MasterdataService groups VehicleType/AccessoryCost/TransportCategory,
// mirroring backend/routers/masterdata.py's own rationale: each collection
// only has list+create (no update/delete), not worth three separate services.
type MasterdataService struct {
	db *gorm.DB
}

func NewMasterdataService(db *gorm.DB) *MasterdataService {
	return &MasterdataService{db: db}
}

const listLimit = 1000

func (s *MasterdataService) ListVehicleTypes(ctx context.Context) ([]dto.VehicleTypeResponse, error) {
	var items []models.VehicleType
	if err := s.db.WithContext(ctx).Where("active = ?", true).Limit(listLimit).Find(&items).Error; err != nil {
		return nil, err
	}
	result := make([]dto.VehicleTypeResponse, len(items))
	for i, v := range items {
		result[i] = dto.VehicleTypeResponse{ID: v.ID, Nome: v.Nome, Descrizione: v.Descrizione, Active: v.Active}
	}
	return result, nil
}

func (s *MasterdataService) CreateVehicleType(ctx context.Context, req dto.VehicleTypeRequest) (*dto.VehicleTypeResponse, error) {
	v := models.VehicleType{ID: uuid.New(), Nome: req.Nome, Descrizione: req.Descrizione, Active: true}
	if err := s.db.WithContext(ctx).Create(&v).Error; err != nil {
		return nil, err
	}
	return &dto.VehicleTypeResponse{ID: v.ID, Nome: v.Nome, Descrizione: v.Descrizione, Active: v.Active}, nil
}

func (s *MasterdataService) ListAccessoryCosts(ctx context.Context) ([]dto.AccessoryCostResponse, error) {
	var items []models.AccessoryCost
	if err := s.db.WithContext(ctx).Where("active = ?", true).Limit(listLimit).Find(&items).Error; err != nil {
		return nil, err
	}
	result := make([]dto.AccessoryCostResponse, len(items))
	for i, a := range items {
		result[i] = dto.AccessoryCostResponse{ID: a.ID, Nome: a.Nome, Descrizione: a.Descrizione, CostoDefault: a.CostoDefault, Active: a.Active}
	}
	return result, nil
}

func (s *MasterdataService) CreateAccessoryCost(ctx context.Context, req dto.AccessoryCostRequest) (*dto.AccessoryCostResponse, error) {
	a := models.AccessoryCost{ID: uuid.New(), Nome: req.Nome, Descrizione: req.Descrizione, CostoDefault: req.CostoDefault, Active: true}
	if err := s.db.WithContext(ctx).Create(&a).Error; err != nil {
		return nil, err
	}
	return &dto.AccessoryCostResponse{ID: a.ID, Nome: a.Nome, Descrizione: a.Descrizione, CostoDefault: a.CostoDefault, Active: a.Active}, nil
}

func (s *MasterdataService) ListTransportCategories(ctx context.Context) ([]dto.TransportCategoryResponse, error) {
	var items []models.TransportCategory
	if err := s.db.WithContext(ctx).Where("active = ?", true).Limit(listLimit).Find(&items).Error; err != nil {
		return nil, err
	}
	result := make([]dto.TransportCategoryResponse, len(items))
	for i, t := range items {
		result[i] = dto.TransportCategoryResponse{ID: t.ID, Nome: t.Nome, Descrizione: t.Descrizione, Active: t.Active}
	}
	return result, nil
}

func (s *MasterdataService) CreateTransportCategory(ctx context.Context, req dto.TransportCategoryRequest) (*dto.TransportCategoryResponse, error) {
	t := models.TransportCategory{ID: uuid.New(), Nome: req.Nome, Descrizione: req.Descrizione, Active: true}
	if err := s.db.WithContext(ctx).Create(&t).Error; err != nil {
		return nil, err
	}
	return &dto.TransportCategoryResponse{ID: t.ID, Nome: t.Nome, Descrizione: t.Descrizione, Active: t.Active}, nil
}
