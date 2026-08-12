package drivers

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	"fratelli-feccia/internal/dto"
	"fratelli-feccia/internal/models"
	"fratelli-feccia/pkg/utils"
)

type DriverService struct {
	db *gorm.DB
}

func NewDriverService(db *gorm.DB) *DriverService {
	return &DriverService{db: db}
}

func escapeLike(term string) string {
	if len(term) > 100 {
		term = term[:100]
	}
	replacer := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return replacer.Replace(term)
}

// List searches nome OR cognome, mirroring backend/routers/drivers.py's `$or`.
// Also attaches, per driver, the nearest upcoming motivo=ferie
// DriverUnavailability window (if any) as ProssimeFerie{Da,A}.
func (s *DriverService) List(ctx context.Context, search string, page utils.PageParams) ([]dto.DriverResponse, int64, error) {
	query := s.db.WithContext(ctx).Model(&models.Driver{}).Where("active = ?", true)
	if search != "" {
		term := "%" + strings.ToLower(escapeLike(search)) + "%"
		query = query.Where("LOWER(nome) LIKE ? OR LOWER(cognome) LIKE ?", term, term)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var drivers []models.Driver
	if err := query.Order("cognome ASC").Offset(page.Offset()).Limit(page.Limit).Find(&drivers).Error; err != nil {
		return nil, 0, err
	}

	nextFerie, err := s.nextFeriePerDriver(ctx, drivers)
	if err != nil {
		return nil, 0, err
	}

	result := make([]dto.DriverResponse, len(drivers))
	for i, d := range drivers {
		result[i] = ToResponse(d)
		if next, ok := nextFerie[d.ID]; ok {
			result[i].ProssimeFerieDa = &next.DataDa
			result[i].ProssimeFerieA = &next.DataA
		}
	}
	return result, total, nil
}

// nextFeriePerDriver batch-fetches the earliest still-upcoming motivo=ferie
// unavailability window per driver in one query, rather than one query per
// row from the List loop.
func (s *DriverService) nextFeriePerDriver(ctx context.Context, drivers []models.Driver) (map[uuid.UUID]models.DriverUnavailability, error) {
	result := map[uuid.UUID]models.DriverUnavailability{}
	if len(drivers) == 0 {
		return result, nil
	}
	ids := make([]uuid.UUID, len(drivers))
	for i, d := range drivers {
		ids[i] = d.ID
	}

	today := time.Now().UTC().Format("2006-01-02")
	var windows []models.DriverUnavailability
	if err := s.db.WithContext(ctx).
		Where("autista_id IN ? AND motivo = ? AND data_a >= ?", ids, "ferie", today).
		Order("data_da ASC").Find(&windows).Error; err != nil {
		return nil, err
	}
	for _, w := range windows {
		if _, seen := result[w.AutistaID]; !seen {
			result[w.AutistaID] = w
		}
	}
	return result, nil
}

func (s *DriverService) Create(ctx context.Context, req dto.DriverRequest) (*dto.DriverResponse, error) {
	d := models.Driver{
		ID:              uuid.New(),
		Nome:            req.Nome,
		Cognome:         req.Cognome,
		CodiceFiscale:   req.CodiceFiscale,
		Patente:         marshalStrings(req.Patente),
		ScadenzaPatente: req.ScadenzaPatente,
		Telefono:        req.Telefono,
		Email:           req.Email,
		Note:            req.Note,
		Active:          true,
	}

	if err := s.db.WithContext(ctx).Create(&d).Error; err != nil {
		return nil, err
	}

	resp := ToResponse(d)
	return &resp, nil
}

func (s *DriverService) Update(ctx context.Context, id uuid.UUID, req dto.DriverRequest) (*dto.DriverResponse, error) {
	var d models.Driver
	if err := s.db.WithContext(ctx).First(&d, "id = ?", id).Error; err != nil {
		return nil, err
	}

	d.Nome = req.Nome
	d.Cognome = req.Cognome
	d.CodiceFiscale = req.CodiceFiscale
	d.Patente = marshalStrings(req.Patente)
	d.ScadenzaPatente = req.ScadenzaPatente
	d.Telefono = req.Telefono
	d.Email = req.Email
	d.Note = req.Note

	if err := s.db.WithContext(ctx).Save(&d).Error; err != nil {
		return nil, err
	}

	resp := ToResponse(d)
	return &resp, nil
}

func (s *DriverService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.db.WithContext(ctx).Model(&models.Driver{}).Where("id = ?", id).Update("active", false).Error
}

func marshalStrings(v []string) datatypes.JSON {
	if v == nil {
		v = []string{}
	}
	b, err := json.Marshal(v)
	if err != nil {
		return datatypes.JSON("[]")
	}
	return datatypes.JSON(b)
}

func unmarshalStrings(raw datatypes.JSON) []string {
	out := []string{}
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &out)
	}
	if out == nil {
		out = []string{}
	}
	return out
}

func ToResponse(d models.Driver) dto.DriverResponse {
	return dto.DriverResponse{
		ID:              d.ID,
		Nome:            d.Nome,
		Cognome:         d.Cognome,
		CodiceFiscale:   d.CodiceFiscale,
		Patente:         unmarshalStrings(d.Patente),
		ScadenzaPatente: d.ScadenzaPatente,
		Telefono:        d.Telefono,
		Email:           d.Email,
		Note:            d.Note,
		Active:          d.Active,
		CreatedAt:       d.CreatedAt,
		UpdatedAt:       d.UpdatedAt,
	}
}
