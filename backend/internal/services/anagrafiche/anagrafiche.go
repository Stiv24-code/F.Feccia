package anagrafiche

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"fratelli-feccia/internal/dto"
	"fratelli-feccia/internal/models"
	"fratelli-feccia/pkg/utils"
)

// AnagraficheService groups Country/Bank/AccountingEntry, mirroring
// backend/routers/anagrafiche_extra.py's own rationale: three small CRUDs
// with the same shape, kept together for discoverability. Uniqueness
// (Country.Iso2, AccountingEntry.Codice) is enforced at the application
// layer via check-then-insert, exactly like the Python original — there is
// no DB-level unique constraint on either field in Mongo.
type AnagraficheService struct {
	db *gorm.DB
}

func NewAnagraficheService(db *gorm.DB) *AnagraficheService {
	return &AnagraficheService{db: db}
}

const listLimit = 500

func escapeLike(term string) string {
	if len(term) > 100 {
		term = term[:100]
	}
	replacer := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return replacer.Replace(term)
}

// ── Countries ────────────────────────────────────────────────────────────

func (s *AnagraficheService) ListCountries(ctx context.Context, search string) ([]dto.CountryResponse, error) {
	query := s.db.WithContext(ctx).Model(&models.Country{}).Where("active = ?", true)
	if search != "" {
		term := "%" + strings.ToLower(escapeLike(search)) + "%"
		query = query.Where("LOWER(nome) LIKE ? OR LOWER(iso2) LIKE ?", term, term)
	}
	var countries []models.Country
	if err := query.Order("nome ASC").Limit(listLimit).Find(&countries).Error; err != nil {
		return nil, err
	}
	result := make([]dto.CountryResponse, len(countries))
	for i, c := range countries {
		result[i] = toCountryResponse(c)
	}
	return result, nil
}

func (s *AnagraficheService) CreateCountry(ctx context.Context, req dto.CountryRequest) (*dto.CountryResponse, error) {
	iso2 := strings.ToUpper(strings.TrimSpace(req.Iso2))
	if len(iso2) != 2 {
		return nil, utils.NewAPIError(400, "Codice ISO 3166-1 alpha-2 richiesto (2 lettere)")
	}

	var count int64
	if err := s.db.WithContext(ctx).Model(&models.Country{}).Where("iso2 = ? AND active = ?", iso2, true).Count(&count).Error; err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, utils.NewAPIError(409, "Nazione "+iso2+" già presente")
	}

	c := models.Country{
		ID:     uuid.New(),
		Iso2:   iso2,
		Iso3:   strings.ToUpper(req.Iso3),
		Nome:   req.Nome,
		Eu:     req.Eu,
		Valuta: req.Valuta,
		Active: true,
	}
	if err := s.db.WithContext(ctx).Create(&c).Error; err != nil {
		return nil, err
	}
	resp := toCountryResponse(c)
	return &resp, nil
}

func (s *AnagraficheService) UpdateCountry(ctx context.Context, id uuid.UUID, req dto.CountryRequest) (*dto.CountryResponse, error) {
	var c models.Country
	if err := s.db.WithContext(ctx).First(&c, "id = ?", id).Error; err != nil {
		return nil, err
	}

	c.Iso2 = strings.ToUpper(strings.TrimSpace(req.Iso2))
	c.Iso3 = strings.ToUpper(strings.TrimSpace(req.Iso3))
	c.Nome = req.Nome
	c.Eu = req.Eu
	c.Valuta = req.Valuta

	if err := s.db.WithContext(ctx).Save(&c).Error; err != nil {
		return nil, err
	}
	resp := toCountryResponse(c)
	return &resp, nil
}

func (s *AnagraficheService) DeleteCountry(ctx context.Context, id uuid.UUID) error {
	return s.db.WithContext(ctx).Model(&models.Country{}).Where("id = ?", id).Update("active", false).Error
}

// ── Banks ────────────────────────────────────────────────────────────────

func (s *AnagraficheService) ListBanks(ctx context.Context, search string) ([]dto.BankResponse, error) {
	query := s.db.WithContext(ctx).Model(&models.Bank{}).Where("active = ?", true)
	if search != "" {
		term := "%" + strings.ToLower(escapeLike(search)) + "%"
		query = query.Where("LOWER(nome) LIKE ? OR LOWER(bic_swift) LIKE ?", term, term)
	}
	var banks []models.Bank
	if err := query.Order("nome ASC").Limit(listLimit).Find(&banks).Error; err != nil {
		return nil, err
	}
	result := make([]dto.BankResponse, len(banks))
	for i, b := range banks {
		result[i] = toBankResponse(b)
	}
	return result, nil
}

func (s *AnagraficheService) CreateBank(ctx context.Context, req dto.BankRequest) (*dto.BankResponse, error) {
	b := models.Bank{
		ID:         uuid.New(),
		Nome:       req.Nome,
		BicSwift:   req.BicSwift,
		IbanPrefix: req.IbanPrefix,
		Indirizzo:  req.Indirizzo,
		Citta:      req.Citta,
		Note:       req.Note,
		Active:     true,
	}
	if err := s.db.WithContext(ctx).Create(&b).Error; err != nil {
		return nil, err
	}
	resp := toBankResponse(b)
	return &resp, nil
}

func (s *AnagraficheService) UpdateBank(ctx context.Context, id uuid.UUID, req dto.BankRequest) (*dto.BankResponse, error) {
	var b models.Bank
	if err := s.db.WithContext(ctx).First(&b, "id = ?", id).Error; err != nil {
		return nil, err
	}

	b.Nome = req.Nome
	b.BicSwift = req.BicSwift
	b.IbanPrefix = req.IbanPrefix
	b.Indirizzo = req.Indirizzo
	b.Citta = req.Citta
	b.Note = req.Note

	if err := s.db.WithContext(ctx).Save(&b).Error; err != nil {
		return nil, err
	}
	resp := toBankResponse(b)
	return &resp, nil
}

func (s *AnagraficheService) DeleteBank(ctx context.Context, id uuid.UUID) error {
	return s.db.WithContext(ctx).Model(&models.Bank{}).Where("id = ?", id).Update("active", false).Error
}

// ── Accounting entries ──────────────────────────────────────────────────

func (s *AnagraficheService) ListAccountingEntries(ctx context.Context, search, tipo string) ([]dto.AccountingEntryResponse, error) {
	query := s.db.WithContext(ctx).Model(&models.AccountingEntry{}).Where("active = ?", true)
	if tipo != "" {
		query = query.Where("tipo = ?", tipo)
	}
	if search != "" {
		term := "%" + strings.ToLower(escapeLike(search)) + "%"
		query = query.Where("LOWER(codice) LIKE ? OR LOWER(descrizione) LIKE ?", term, term)
	}
	var entries []models.AccountingEntry
	if err := query.Order("codice ASC").Limit(listLimit).Find(&entries).Error; err != nil {
		return nil, err
	}
	result := make([]dto.AccountingEntryResponse, len(entries))
	for i, e := range entries {
		result[i] = toAccountingEntryResponse(e)
	}
	return result, nil
}

func (s *AnagraficheService) CreateAccountingEntry(ctx context.Context, req dto.AccountingEntryRequest) (*dto.AccountingEntryResponse, error) {
	tipo := defaultString(req.Tipo, "ricavo")
	if tipo != "ricavo" && tipo != "costo" {
		return nil, utils.NewAPIError(400, "tipo deve essere 'ricavo' o 'costo'")
	}

	var count int64
	if err := s.db.WithContext(ctx).Model(&models.AccountingEntry{}).Where("codice = ? AND active = ?", req.Codice, true).Count(&count).Error; err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, utils.NewAPIError(409, "Voce contabile "+req.Codice+" già presente")
	}

	e := models.AccountingEntry{
		ID:             uuid.New(),
		Codice:         req.Codice,
		Descrizione:    req.Descrizione,
		Tipo:           tipo,
		ContoContabile: req.ContoContabile,
		IvaCodice:      defaultString(req.IvaCodice, "N8"),
		Active:         true,
	}
	if err := s.db.WithContext(ctx).Create(&e).Error; err != nil {
		return nil, err
	}
	resp := toAccountingEntryResponse(e)
	return &resp, nil
}

func (s *AnagraficheService) UpdateAccountingEntry(ctx context.Context, id uuid.UUID, req dto.AccountingEntryRequest) (*dto.AccountingEntryResponse, error) {
	tipo := defaultString(req.Tipo, "ricavo")
	if tipo != "ricavo" && tipo != "costo" {
		return nil, utils.NewAPIError(400, "tipo deve essere 'ricavo' o 'costo'")
	}

	var e models.AccountingEntry
	if err := s.db.WithContext(ctx).First(&e, "id = ?", id).Error; err != nil {
		return nil, err
	}

	e.Codice = req.Codice
	e.Descrizione = req.Descrizione
	e.Tipo = tipo
	e.ContoContabile = req.ContoContabile
	e.IvaCodice = defaultString(req.IvaCodice, "N8")

	if err := s.db.WithContext(ctx).Save(&e).Error; err != nil {
		return nil, err
	}
	resp := toAccountingEntryResponse(e)
	return &resp, nil
}

func (s *AnagraficheService) DeleteAccountingEntry(ctx context.Context, id uuid.UUID) error {
	return s.db.WithContext(ctx).Model(&models.AccountingEntry{}).Where("id = ?", id).Update("active", false).Error
}

func defaultString(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func toCountryResponse(c models.Country) dto.CountryResponse {
	return dto.CountryResponse{
		ID: c.ID, Iso2: c.Iso2, Iso3: c.Iso3, Nome: c.Nome, Eu: c.Eu, Valuta: c.Valuta,
		Active: c.Active, CreatedAt: c.CreatedAt, UpdatedAt: c.UpdatedAt,
	}
}

func toBankResponse(b models.Bank) dto.BankResponse {
	return dto.BankResponse{
		ID: b.ID, Nome: b.Nome, BicSwift: b.BicSwift, IbanPrefix: b.IbanPrefix,
		Indirizzo: b.Indirizzo, Citta: b.Citta, Note: b.Note,
		Active: b.Active, CreatedAt: b.CreatedAt, UpdatedAt: b.UpdatedAt,
	}
}

func toAccountingEntryResponse(e models.AccountingEntry) dto.AccountingEntryResponse {
	return dto.AccountingEntryResponse{
		ID: e.ID, Codice: e.Codice, Descrizione: e.Descrizione, Tipo: e.Tipo,
		ContoContabile: e.ContoContabile, IvaCodice: e.IvaCodice,
		Active: e.Active, CreatedAt: e.CreatedAt, UpdatedAt: e.UpdatedAt,
	}
}
