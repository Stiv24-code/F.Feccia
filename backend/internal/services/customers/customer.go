package customers

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"fratelli-feccia/internal/dto"
	"fratelli-feccia/internal/models"
)

// listLimit mirrors backend/routers/customers.py's `.to_list(1000)` cap —
// this endpoint has no pagination in the Python original.
const listLimit = 1000

type CustomerService struct {
	db *gorm.DB
}

func NewCustomerService(db *gorm.DB) *CustomerService {
	return &CustomerService{db: db}
}

// escapeLike escapes ILIKE wildcards and truncates the term, mirroring the
// intent of backend/database.py's safe_regex (bounded, non-abusable search input).
func escapeLike(term string) string {
	if len(term) > 100 {
		term = term[:100]
	}
	replacer := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return replacer.Replace(term)
}

func (s *CustomerService) List(ctx context.Context, search string) ([]dto.CustomerResponse, error) {
	query := s.db.WithContext(ctx).Model(&models.Customer{}).Where("active = ?", true)
	if search != "" {
		// LOWER()+LIKE (not ILIKE) so this runs identically on Postgres and the
		// SQLite used by unit tests.
		query = query.Where("LOWER(ragione_sociale) LIKE ?", "%"+strings.ToLower(escapeLike(search))+"%")
	}

	var customers []models.Customer
	if err := query.Order("ragione_sociale ASC").Limit(listLimit).Find(&customers).Error; err != nil {
		return nil, err
	}

	result := make([]dto.CustomerResponse, len(customers))
	for i, c := range customers {
		result[i] = toCustomerResponse(c)
	}
	return result, nil
}

func (s *CustomerService) GetByID(ctx context.Context, id uuid.UUID) (*dto.CustomerResponse, error) {
	var customer models.Customer
	if err := s.db.WithContext(ctx).First(&customer, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	resp := toCustomerResponse(customer)
	return &resp, nil
}

func (s *CustomerService) Create(ctx context.Context, req dto.CustomerRequest) (*dto.CustomerResponse, error) {
	customer := models.Customer{
		ID:                  uuid.New(),
		RagioneSociale:      req.RagioneSociale,
		Indirizzo:           req.Indirizzo,
		Citta:               req.Citta,
		Cap:                 req.Cap,
		Provincia:           req.Provincia,
		Nazione:             defaultString(req.Nazione, "Italia"),
		PartitaIva:          req.PartitaIva,
		CodiceFiscale:       req.CodiceFiscale,
		Telefono:            req.Telefono,
		Email:               req.Email,
		Pec:                 req.Pec,
		CondizioniPagamento: req.CondizioniPagamento,
		Note:                req.Note,
		RichiedeRifOrdine:   req.RichiedeRifOrdine,
		Active:              true,
	}

	if err := s.db.WithContext(ctx).Create(&customer).Error; err != nil {
		return nil, err
	}

	resp := toCustomerResponse(customer)
	return &resp, nil
}

func (s *CustomerService) Update(ctx context.Context, id uuid.UUID, req dto.CustomerRequest) (*dto.CustomerResponse, error) {
	var customer models.Customer
	if err := s.db.WithContext(ctx).First(&customer, "id = ?", id).Error; err != nil {
		return nil, err
	}

	customer.RagioneSociale = req.RagioneSociale
	customer.Indirizzo = req.Indirizzo
	customer.Citta = req.Citta
	customer.Cap = req.Cap
	customer.Provincia = req.Provincia
	customer.Nazione = defaultString(req.Nazione, "Italia")
	customer.PartitaIva = req.PartitaIva
	customer.CodiceFiscale = req.CodiceFiscale
	customer.Telefono = req.Telefono
	customer.Email = req.Email
	customer.Pec = req.Pec
	customer.CondizioniPagamento = req.CondizioniPagamento
	customer.Note = req.Note
	customer.RichiedeRifOrdine = req.RichiedeRifOrdine

	if err := s.db.WithContext(ctx).Save(&customer).Error; err != nil {
		return nil, err
	}

	resp := toCustomerResponse(customer)
	return &resp, nil
}

// Delete is a logical delete (active=false), mirroring backend/routers/customers.py's
// delete_customer, which sets active=False and returns ok regardless of match count.
func (s *CustomerService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.db.WithContext(ctx).Model(&models.Customer{}).Where("id = ?", id).Update("active", false).Error
}

func defaultString(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func toCustomerResponse(c models.Customer) dto.CustomerResponse {
	return dto.CustomerResponse{
		ID:                  c.ID,
		RagioneSociale:      c.RagioneSociale,
		Indirizzo:           c.Indirizzo,
		Citta:               c.Citta,
		Cap:                 c.Cap,
		Provincia:           c.Provincia,
		Nazione:             c.Nazione,
		PartitaIva:          c.PartitaIva,
		CodiceFiscale:       c.CodiceFiscale,
		Telefono:            c.Telefono,
		Email:               c.Email,
		Pec:                 c.Pec,
		CondizioniPagamento: c.CondizioniPagamento,
		Note:                c.Note,
		RichiedeRifOrdine:   c.RichiedeRifOrdine,
		Active:              c.Active,
		CreatedAt:           c.CreatedAt,
		UpdatedAt:           c.UpdatedAt,
	}
}
