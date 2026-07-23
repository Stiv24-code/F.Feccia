package pricelists

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"fratelli-feccia/internal/dto"
	"fratelli-feccia/internal/models"
	"fratelli-feccia/pkg/utils"
)

const listLimit = 1000

type PriceListService struct {
	db *gorm.DB
}

func NewPriceListService(db *gorm.DB) *PriceListService {
	return &PriceListService{db: db}
}

// preloadPriceListAssociations loads Cliente plus every item-level belongs-to
// reference (Prodotto, le due Destinazioni) — the choke point toResponse
// relies on to build the nested Response DTOs.
func preloadPriceListAssociations(q *gorm.DB) *gorm.DB {
	return q.
		Preload("Cliente").
		Preload("Items.Prodotto").
		Preload("Items.DestinazioneCarico").
		Preload("Items.DestinazioneScarico")
}

func (s *PriceListService) reload(ctx context.Context, id uuid.UUID) (*models.PriceList, error) {
	var pl models.PriceList
	if err := preloadPriceListAssociations(s.db.WithContext(ctx)).First(&pl, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &pl, nil
}

func (s *PriceListService) List(ctx context.Context, clienteID string) ([]dto.PriceListResponse, error) {
	query := preloadPriceListAssociations(s.db.WithContext(ctx)).Where("active = ?", true)
	if clienteID != "" {
		query = query.Where("cliente_id = ?", clienteID)
	}
	var lists []models.PriceList
	if err := query.Order("created_at DESC").Limit(listLimit).Find(&lists).Error; err != nil {
		return nil, err
	}
	result := make([]dto.PriceListResponse, len(lists))
	for i, pl := range lists {
		result[i] = toResponse(pl)
	}
	return result, nil
}

func (s *PriceListService) Create(ctx context.Context, req dto.PriceListRequest) (*dto.PriceListResponse, error) {
	clienteID, err := utils.ParseUUID(req.ClienteID)
	if err != nil {
		return nil, err
	}
	items, err := toItems(req.Items)
	if err != nil {
		return nil, err
	}

	pl := models.PriceList{
		ID: uuid.New(), ClienteID: clienteID,
		DataInizio: req.DataInizio, DataFine: req.DataFine, Note: req.Note,
		Items: items, Active: true,
	}
	if err := s.db.WithContext(ctx).Create(&pl).Error; err != nil {
		return nil, err
	}
	reloaded, err := s.reload(ctx, pl.ID)
	if err != nil {
		return nil, err
	}
	resp := toResponse(*reloaded)
	return &resp, nil
}

func (s *PriceListService) GetByID(ctx context.Context, id uuid.UUID) (*dto.PriceListResponse, error) {
	pl, err := s.reload(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	resp := toResponse(*pl)
	return &resp, nil
}

// Update mirrors update_pricelist: if the pricelist is in_uso, the update is
// applied to a *new* duplicate (preserving history) instead of in place —
// the original is deactivated, matching the Python original exactly.
func (s *PriceListService) Update(ctx context.Context, id uuid.UUID, req dto.PriceListRequest) (*dto.PriceListUpdateResult, error) {
	var existing models.PriceList
	if err := s.db.WithContext(ctx).First(&existing, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.NewAPIError(404, "Listino non trovato")
		}
		return nil, err
	}

	clienteID, err := utils.ParseUUID(req.ClienteID)
	if err != nil {
		return nil, err
	}
	items, err := toItems(req.Items)
	if err != nil {
		return nil, err
	}

	if existing.InUso {
		newPL := models.PriceList{
			ID: uuid.New(), ClienteID: clienteID,
			DataInizio: req.DataInizio, DataFine: req.DataFine, Note: req.Note,
			Items: items, Active: true, InUso: false,
		}
		err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			if err := tx.Create(&newPL).Error; err != nil {
				return err
			}
			return tx.Model(&models.PriceList{}).Where("id = ?", id).Update("active", false).Error
		})
		if err != nil {
			return nil, err
		}
		return &dto.PriceListUpdateResult{OK: true, NewID: &newPL.ID, Duplicated: true}, nil
	}

	existing.ClienteID = clienteID
	existing.DataInizio = req.DataInizio
	existing.DataFine = req.DataFine
	existing.Note = req.Note

	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("price_list_id = ?", id).Delete(&models.PriceListItem{}).Error; err != nil {
			return err
		}
		existing.Items = items
		return tx.Save(&existing).Error
	})
	if err != nil {
		return nil, err
	}
	return &dto.PriceListUpdateResult{OK: true, Duplicated: false}, nil
}

func (s *PriceListService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.db.WithContext(ctx).Model(&models.PriceList{}).Where("id = ?", id).Update("active", false).Error
}

// AddItem mirrors POST /pricelists/{id}/items.
func (s *PriceListService) AddItem(ctx context.Context, plID uuid.UUID, item dto.PriceListItemRequestDTO) (*dto.PriceListItemAddResult, error) {
	var pl models.PriceList
	if err := s.db.WithContext(ctx).Preload("Items").First(&pl, "id = ?", plID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.NewAPIError(404, "Listino non trovato")
		}
		return nil, err
	}

	newItem, err := toItem(item)
	if err != nil {
		return nil, err
	}
	newItem.PriceListID = plID
	if err := s.db.WithContext(ctx).Create(&newItem).Error; err != nil {
		return nil, err
	}

	return &dto.PriceListItemAddResult{OK: true, ItemID: newItem.ID, ItemsCount: len(pl.Items) + 1}, nil
}

// UpdateItem mirrors PUT /pricelists/{id}/items/{item_id} — item_id is
// preserved even if the request body includes a different one.
func (s *PriceListService) UpdateItem(ctx context.Context, plID, itemID uuid.UUID, item dto.PriceListItemRequestDTO) (*dto.PriceListItemUpdateResult, error) {
	var count int64
	if err := s.db.WithContext(ctx).Model(&models.PriceList{}).Where("id = ?", plID).Count(&count).Error; err != nil {
		return nil, err
	}
	if count == 0 {
		return nil, utils.NewAPIError(404, "Listino non trovato")
	}

	updated, err := toItem(item)
	if err != nil {
		return nil, err
	}
	result := s.db.WithContext(ctx).Model(&models.PriceListItem{}).
		Where("id = ? AND price_list_id = ?", itemID, plID).
		Updates(map[string]interface{}{
			"prodotto_id":             updated.ProdottoID,
			"destinazione_carico_id":  updated.DestinazioneCaricoID,
			"destinazione_scarico_id": updated.DestinazioneScaricoID,
			"tariffa":                 updated.Tariffa, "tipo_tariffa": updated.TipoTariffa,
			"range_peso_min": updated.RangePesoMin, "range_peso_max": updated.RangePesoMax,
			"unita_peso": updated.UnitaPeso, "minimo_tassabile": updated.MinimoTassabile,
			"tipo_trasporto": updated.TipoTrasporto, "perc_adeguamento_carburante": updated.PercAdeguamentoCarburante,
		})
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, utils.NewAPIError(404, "Regola non trovata")
	}

	return &dto.PriceListItemUpdateResult{OK: true, ItemID: itemID}, nil
}

// DeleteItem mirrors DELETE /pricelists/{id}/items/{item_id}. The Python
// index-based legacy fallback is not ported — every item in this schema
// always has a real ID, unlike Mongo documents migrated before item_id
// existed.
func (s *PriceListService) DeleteItem(ctx context.Context, plID, itemID uuid.UUID) (*dto.PriceListItemDeleteResult, error) {
	var count int64
	if err := s.db.WithContext(ctx).Model(&models.PriceList{}).Where("id = ?", plID).Count(&count).Error; err != nil {
		return nil, err
	}
	if count == 0 {
		return nil, utils.NewAPIError(404, "Listino non trovato")
	}

	if err := s.db.WithContext(ctx).Where("id = ? AND price_list_id = ?", itemID, plID).Delete(&models.PriceListItem{}).Error; err != nil {
		return nil, err
	}

	var remaining int64
	s.db.WithContext(ctx).Model(&models.PriceListItem{}).Where("price_list_id = ?", plID).Count(&remaining)

	return &dto.PriceListItemDeleteResult{OK: true, ItemsCount: int(remaining)}, nil
}

// LookupTariff ports services.py's score_rule_match/compute_tariff engine
// verbatim (same point values and tariff formula).
func (s *PriceListService) LookupTariff(ctx context.Context, clienteID, caricoID, scaricoID, prodottoID string, peso float64) (*dto.TariffLookupResult, error) {
	today := time.Now().UTC().Format("2006-01-02")

	caricoUUID, err := utils.ParseOptionalUUID(caricoID)
	if err != nil {
		return nil, err
	}
	scaricoUUID, err := utils.ParseOptionalUUID(scaricoID)
	if err != nil {
		return nil, err
	}
	prodottoUUID, err := utils.ParseOptionalUUID(prodottoID)
	if err != nil {
		return nil, err
	}

	var lists []models.PriceList
	err = s.db.WithContext(ctx).Preload("Items").
		Where("active = ? AND cliente_id = ? AND data_inizio <= ? AND data_fine >= ?", true, clienteID, today, today).
		Limit(100).Find(&lists).Error
	if err != nil {
		return nil, err
	}
	if len(lists) == 0 {
		if err := s.db.WithContext(ctx).Preload("Items").
			Where("active = ? AND cliente_id = ?", true, clienteID).
			Order("created_at DESC").Limit(1).Find(&lists).Error; err != nil {
			return nil, err
		}
	}

	var best *dto.TariffLookupResult
	bestScore := -1

	for _, pl := range lists {
		for _, item := range pl.Items {
			score := scoreRuleMatch(item, caricoUUID, scaricoUUID, prodottoUUID, peso)
			if score < 0 {
				continue
			}
			if score > bestScore {
				bestScore = score
				calc, base, tipo, carburante, minimo := computeTariff(item, peso)
				best = &dto.TariffLookupResult{
					Found: true, Tariffa: calc, TariffaBase: base, TipoTariffa: tipo,
					PercAdeguamentoCarburante: carburante, MinimoTassabile: minimo,
					ListinoID: pl.ID, ItemID: item.ID, Score: bestScore,
				}
			}
		}
	}

	if best == nil {
		return &dto.TariffLookupResult{Found: false, Tariffa: 0, TipoTariffa: "forfait"}, nil
	}
	return best, nil
}

// ── Scoring engine (ports services.py's score_rule_match/compute_tariff) ──

func uuidEq(a, b *uuid.UUID) bool {
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

func scoreTratta(item models.PriceListItem, caricoID, scaricoID *uuid.UUID) int {
	itemCarico := item.DestinazioneCaricoID
	itemScarico := item.DestinazioneScaricoID
	if itemCarico != nil && itemScarico != nil {
		if uuidEq(itemCarico, caricoID) && uuidEq(itemScarico, scaricoID) {
			return 10
		}
		return -1
	}
	if itemCarico != nil || itemScarico != nil {
		if itemCarico != nil && !uuidEq(itemCarico, caricoID) {
			return -1
		}
		if itemScarico != nil && !uuidEq(itemScarico, scaricoID) {
			return -1
		}
		return 5
	}
	return 0
}

func scoreProdotto(item models.PriceListItem, prodottoID *uuid.UUID) int {
	if item.ProdottoID == nil {
		return 0
	}
	if uuidEq(item.ProdottoID, prodottoID) {
		return 5
	}
	return -1
}

func scorePeso(item models.PriceListItem, peso float64) int {
	pesoMin := item.RangePesoMin
	pesoMax := item.RangePesoMax
	if (pesoMin > 0 || pesoMax > 0) && peso > 0 {
		if pesoMin > 0 && peso < pesoMin {
			return -1
		}
		if pesoMax > 0 && peso > pesoMax {
			return -1
		}
		return 3
	}
	return 0
}

func scoreRuleMatch(item models.PriceListItem, caricoID, scaricoID, prodottoID *uuid.UUID, peso float64) int {
	tratta := scoreTratta(item, caricoID, scaricoID)
	if tratta < 0 {
		return -1
	}
	prodotto := scoreProdotto(item, prodottoID)
	if prodotto < 0 {
		return -1
	}
	pesoScore := scorePeso(item, peso)
	if pesoScore < 0 {
		return -1
	}
	return tratta + prodotto + pesoScore
}

func computeTariff(item models.PriceListItem, peso float64) (calcolata, base float64, tipo string, carburante, minimo float64) {
	base = item.Tariffa
	tipo = item.TipoTariffa
	if tipo == "" {
		tipo = "forfait"
	}
	minimo = item.MinimoTassabile
	carburante = item.PercAdeguamentoCarburante

	if tipo == "euro_kg" && peso > 0 {
		pesoEffettivo := peso
		if minimo > 0 && minimo > peso {
			pesoEffettivo = minimo
		}
		calcolata = base * pesoEffettivo
	} else {
		calcolata = base
	}

	if carburante > 0 {
		calcolata *= 1 + carburante/100
	}

	return roundTo2(calcolata), base, tipo, carburante, minimo
}

func roundTo2(v float64) float64 {
	return float64(int(v*100+0.5)) / 100
}

// ── Helpers ──────────────────────────────────────────────────────────────

func toItem(dtoItem dto.PriceListItemRequestDTO) (models.PriceListItem, error) {
	id := uuid.New()
	if dtoItem.ItemID != nil {
		id = *dtoItem.ItemID
	}
	prodottoID, err := utils.ParseOptionalUUID(dtoItem.ProdottoID)
	if err != nil {
		return models.PriceListItem{}, err
	}
	caricoID, err := utils.ParseOptionalUUID(dtoItem.DestinazioneCaricoID)
	if err != nil {
		return models.PriceListItem{}, err
	}
	scaricoID, err := utils.ParseOptionalUUID(dtoItem.DestinazioneScaricoID)
	if err != nil {
		return models.PriceListItem{}, err
	}
	return models.PriceListItem{
		ID: id, ProdottoID: prodottoID,
		DestinazioneCaricoID: caricoID, DestinazioneScaricoID: scaricoID,
		Tariffa: dtoItem.Tariffa, TipoTariffa: defaultString(dtoItem.TipoTariffa, "forfait"),
		RangePesoMin: dtoItem.RangePesoMin, RangePesoMax: dtoItem.RangePesoMax,
		UnitaPeso: defaultString(dtoItem.UnitaPeso, "Kg"), MinimoTassabile: dtoItem.MinimoTassabile,
		TipoTrasporto: defaultString(dtoItem.TipoTrasporto, "stradale"), PercAdeguamentoCarburante: dtoItem.PercAdeguamentoCarburante,
	}, nil
}

func toItems(items []dto.PriceListItemRequestDTO) ([]models.PriceListItem, error) {
	result := make([]models.PriceListItem, len(items))
	for i, it := range items {
		item, err := toItem(it)
		if err != nil {
			return nil, err
		}
		result[i] = item
	}
	return result, nil
}

func defaultString(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func customerResponse(c models.Customer) *dto.CustomerResponse {
	if c.ID == uuid.Nil {
		return nil
	}
	return &dto.CustomerResponse{
		ID: c.ID, RagioneSociale: c.RagioneSociale, Indirizzo: c.Indirizzo, Citta: c.Citta,
		Cap: c.Cap, Provincia: c.Provincia, Nazione: c.Nazione, PartitaIva: c.PartitaIva,
		CodiceFiscale: c.CodiceFiscale, Telefono: c.Telefono, Email: c.Email, Pec: c.Pec,
		CondizioniPagamento: c.CondizioniPagamento, Note: c.Note, RichiedeRifOrdine: c.RichiedeRifOrdine,
		Active: c.Active, CreatedAt: c.CreatedAt, UpdatedAt: c.UpdatedAt,
	}
}

func productResponse(p *models.Product) *dto.ProductResponse {
	if p == nil {
		return nil
	}
	return &dto.ProductResponse{
		ID: p.ID, Codice: p.Codice, Descrizione: p.Descrizione, UnitaMisura: p.UnitaMisura,
		Note: p.Note, Active: p.Active, CreatedAt: p.CreatedAt, UpdatedAt: p.UpdatedAt,
	}
}

func destinationResponse(d *models.Destination) *dto.DestinationResponse {
	if d == nil {
		return nil
	}
	return &dto.DestinationResponse{
		ID: d.ID, Nome: d.Nome, Indirizzo: d.Indirizzo, Citta: d.Citta, Cap: d.Cap,
		Provincia: d.Provincia, Nazione: d.Nazione, Lat: d.Lat, Lng: d.Lng,
		VincoliScarico: d.VincoliScarico, Note: d.Note, Active: d.Active,
		CreatedAt: d.CreatedAt, UpdatedAt: d.UpdatedAt,
	}
}

func toItemDTO(it models.PriceListItem) dto.PriceListItemResponseDTO {
	return dto.PriceListItemResponseDTO{
		ItemID: it.ID, Prodotto: productResponse(it.Prodotto),
		DestinazioneCarico: destinationResponse(it.DestinazioneCarico), DestinazioneScarico: destinationResponse(it.DestinazioneScarico),
		Tariffa: it.Tariffa, TipoTariffa: it.TipoTariffa,
		RangePesoMin: it.RangePesoMin, RangePesoMax: it.RangePesoMax,
		UnitaPeso: it.UnitaPeso, MinimoTassabile: it.MinimoTassabile,
		TipoTrasporto: it.TipoTrasporto, PercAdeguamentoCarburante: it.PercAdeguamentoCarburante,
	}
}

func toResponse(pl models.PriceList) dto.PriceListResponse {
	items := make([]dto.PriceListItemResponseDTO, len(pl.Items))
	for i, it := range pl.Items {
		items[i] = toItemDTO(it)
	}
	return dto.PriceListResponse{
		ID: pl.ID, ClienteID: pl.ClienteID.String(), Cliente: customerResponse(pl.Cliente),
		DataInizio: pl.DataInizio, DataFine: pl.DataFine, Items: items,
		Note: pl.Note, InUso: pl.InUso, Active: pl.Active, CreatedAt: pl.CreatedAt,
	}
}
