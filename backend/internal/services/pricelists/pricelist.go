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

func (s *PriceListService) List(ctx context.Context, clienteID string) ([]dto.PriceListResponse, error) {
	query := s.db.WithContext(ctx).Preload("Items").Where("active = ?", true)
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
	pl := models.PriceList{
		ID: uuid.New(), ClienteID: req.ClienteID, ClienteNome: req.ClienteNome,
		DataInizio: req.DataInizio, DataFine: req.DataFine, Note: req.Note,
		Items: toItems(req.Items), Active: true,
	}
	if err := s.db.WithContext(ctx).Create(&pl).Error; err != nil {
		return nil, err
	}
	resp := toResponse(pl)
	return &resp, nil
}

func (s *PriceListService) GetByID(ctx context.Context, id uuid.UUID) (*dto.PriceListResponse, error) {
	var pl models.PriceList
	if err := s.db.WithContext(ctx).Preload("Items").First(&pl, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	resp := toResponse(pl)
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

	if existing.InUso {
		newPL := models.PriceList{
			ID: uuid.New(), ClienteID: req.ClienteID, ClienteNome: req.ClienteNome,
			DataInizio: req.DataInizio, DataFine: req.DataFine, Note: req.Note,
			Items: toItems(req.Items), Active: true, InUso: false,
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

	existing.ClienteID = req.ClienteID
	existing.ClienteNome = req.ClienteNome
	existing.DataInizio = req.DataInizio
	existing.DataFine = req.DataFine
	existing.Note = req.Note

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("price_list_id = ?", id).Delete(&models.PriceListItem{}).Error; err != nil {
			return err
		}
		existing.Items = toItems(req.Items)
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
func (s *PriceListService) AddItem(ctx context.Context, plID uuid.UUID, item dto.PriceListItemDTO) (*dto.PriceListItemAddResult, error) {
	var pl models.PriceList
	if err := s.db.WithContext(ctx).Preload("Items").First(&pl, "id = ?", plID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.NewAPIError(404, "Listino non trovato")
		}
		return nil, err
	}

	newItem := toItem(item)
	newItem.PriceListID = plID
	if err := s.db.WithContext(ctx).Create(&newItem).Error; err != nil {
		return nil, err
	}

	return &dto.PriceListItemAddResult{OK: true, ItemID: newItem.ID, ItemsCount: len(pl.Items) + 1}, nil
}

// UpdateItem mirrors PUT /pricelists/{id}/items/{item_id} — item_id is
// preserved even if the request body includes a different one.
func (s *PriceListService) UpdateItem(ctx context.Context, plID, itemID uuid.UUID, item dto.PriceListItemDTO) (*dto.PriceListItemUpdateResult, error) {
	var count int64
	if err := s.db.WithContext(ctx).Model(&models.PriceList{}).Where("id = ?", plID).Count(&count).Error; err != nil {
		return nil, err
	}
	if count == 0 {
		return nil, utils.NewAPIError(404, "Listino non trovato")
	}

	updated := toItem(item)
	result := s.db.WithContext(ctx).Model(&models.PriceListItem{}).
		Where("id = ? AND price_list_id = ?", itemID, plID).
		Updates(map[string]interface{}{
			"prodotto_id": updated.ProdottoID, "prodotto_nome": updated.ProdottoNome,
			"destinazione_carico_id": updated.DestinazioneCaricoID, "destinazione_carico_nome": updated.DestinazioneCaricoNome,
			"destinazione_scarico_id": updated.DestinazioneScaricoID, "destinazione_scarico_nome": updated.DestinazioneScaricoNome,
			"tariffa": updated.Tariffa, "tipo_tariffa": updated.TipoTariffa,
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

	var lists []models.PriceList
	err := s.db.WithContext(ctx).Preload("Items").
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
			score := scoreRuleMatch(item, caricoID, scaricoID, prodottoID, peso)
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

func scoreTratta(item models.PriceListItem, caricoID, scaricoID string) int {
	itemCarico := item.DestinazioneCaricoID
	itemScarico := item.DestinazioneScaricoID
	if itemCarico != "" && itemScarico != "" {
		if itemCarico == caricoID && itemScarico == scaricoID {
			return 10
		}
		return -1
	}
	if itemCarico != "" || itemScarico != "" {
		if itemCarico != "" && itemCarico != caricoID {
			return -1
		}
		if itemScarico != "" && itemScarico != scaricoID {
			return -1
		}
		return 5
	}
	return 0
}

func scoreProdotto(item models.PriceListItem, prodottoID string) int {
	if item.ProdottoID == "" {
		return 0
	}
	if item.ProdottoID == prodottoID {
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

func scoreRuleMatch(item models.PriceListItem, caricoID, scaricoID, prodottoID string, peso float64) int {
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

func toItem(dtoItem dto.PriceListItemDTO) models.PriceListItem {
	id := uuid.New()
	if dtoItem.ItemID != nil {
		id = *dtoItem.ItemID
	}
	return models.PriceListItem{
		ID: id, ProdottoID: dtoItem.ProdottoID, ProdottoNome: dtoItem.ProdottoNome,
		DestinazioneCaricoID: dtoItem.DestinazioneCaricoID, DestinazioneCaricoNome: dtoItem.DestinazioneCaricoNome,
		DestinazioneScaricoID: dtoItem.DestinazioneScaricoID, DestinazioneScaricoNome: dtoItem.DestinazioneScaricoNome,
		Tariffa: dtoItem.Tariffa, TipoTariffa: defaultString(dtoItem.TipoTariffa, "forfait"),
		RangePesoMin: dtoItem.RangePesoMin, RangePesoMax: dtoItem.RangePesoMax,
		UnitaPeso: defaultString(dtoItem.UnitaPeso, "Kg"), MinimoTassabile: dtoItem.MinimoTassabile,
		TipoTrasporto: defaultString(dtoItem.TipoTrasporto, "stradale"), PercAdeguamentoCarburante: dtoItem.PercAdeguamentoCarburante,
	}
}

func toItems(items []dto.PriceListItemDTO) []models.PriceListItem {
	result := make([]models.PriceListItem, len(items))
	for i, it := range items {
		result[i] = toItem(it)
	}
	return result
}

func defaultString(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func toItemDTO(it models.PriceListItem) dto.PriceListItemDTO {
	id := it.ID
	return dto.PriceListItemDTO{
		ItemID: &id, ProdottoID: it.ProdottoID, ProdottoNome: it.ProdottoNome,
		DestinazioneCaricoID: it.DestinazioneCaricoID, DestinazioneCaricoNome: it.DestinazioneCaricoNome,
		DestinazioneScaricoID: it.DestinazioneScaricoID, DestinazioneScaricoNome: it.DestinazioneScaricoNome,
		Tariffa: it.Tariffa, TipoTariffa: it.TipoTariffa,
		RangePesoMin: it.RangePesoMin, RangePesoMax: it.RangePesoMax,
		UnitaPeso: it.UnitaPeso, MinimoTassabile: it.MinimoTassabile,
		TipoTrasporto: it.TipoTrasporto, PercAdeguamentoCarburante: it.PercAdeguamentoCarburante,
	}
}

func toResponse(pl models.PriceList) dto.PriceListResponse {
	items := make([]dto.PriceListItemDTO, len(pl.Items))
	for i, it := range pl.Items {
		items[i] = toItemDTO(it)
	}
	return dto.PriceListResponse{
		ID: pl.ID, ClienteID: pl.ClienteID, ClienteNome: pl.ClienteNome,
		DataInizio: pl.DataInizio, DataFine: pl.DataFine, Items: items,
		Note: pl.Note, InUso: pl.InUso, Active: pl.Active, CreatedAt: pl.CreatedAt,
	}
}
