package orders

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	"fratelli-feccia/internal/dto"
	"fratelli-feccia/internal/models"
	"fratelli-feccia/pkg/database"
	"fratelli-feccia/pkg/pdfgen"
	"fratelli-feccia/pkg/utils"
)

const defaultListLimit = 500

type OrderService struct {
	db *gorm.DB
}

func NewOrderService(db *gorm.DB) *OrderService {
	return &OrderService{db: db}
}

// ListFilters mirrors the query params of GET /orders in backend/routers/orders.py.
type ListFilters struct {
	Stato     string
	ClienteID string
	DataDa    string
	DataA     string
	Search    string
	Tipologia string
	Limit     int
}

func escapeLike(term string) string {
	if len(term) > 100 {
		term = term[:100]
	}
	replacer := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return replacer.Replace(term)
}

func (s *OrderService) List(ctx context.Context, f ListFilters) ([]dto.OrderResponse, error) {
	limit := f.Limit
	if limit <= 0 {
		limit = defaultListLimit
	}

	query := s.db.WithContext(ctx).Model(&models.Order{}).Preload("Items")
	if f.Stato != "" {
		query = query.Where("stato = ?", f.Stato)
	}
	if f.ClienteID != "" {
		query = query.Where("cliente_id = ?", f.ClienteID)
	}
	if f.Tipologia != "" {
		query = query.Where("tipologia = ?", f.Tipologia)
	}
	if f.DataDa != "" {
		query = query.Where("data_ritiro >= ?", f.DataDa)
	}
	if f.DataA != "" {
		query = query.Where("data_ritiro <= ?", f.DataA)
	}
	if f.Search != "" {
		term := "%" + strings.ToLower(escapeLike(f.Search)) + "%"
		query = query.Where(
			"LOWER(cliente_nome) LIKE ? OR LOWER(progressivo) LIKE ? OR LOWER(rif_ordine_cliente) LIKE ? OR LOWER(destinazione_carico_nome) LIKE ? OR LOWER(destinazione_scarico_nome) LIKE ?",
			term, term, term, term, term,
		)
	}

	var orders []models.Order
	if err := query.Order("created_at DESC").Limit(limit).Find(&orders).Error; err != nil {
		return nil, err
	}

	result := make([]dto.OrderResponse, len(orders))
	for i, o := range orders {
		result[i] = ToResponse(o)
	}
	return result, nil
}

func (s *OrderService) Create(ctx context.Context, req dto.OrderRequest) (*dto.OrderResponse, error) {
	seq, err := database.NextSequence(s.db.WithContext(ctx), "orders")
	if err != nil {
		return nil, err
	}
	progressivo := fmt.Sprintf("%s/%04d", time.Now().Format("06"), seq)

	order := models.Order{
		ID:                      uuid.New(),
		Progressivo:             progressivo,
		ClienteID:               req.ClienteID,
		ClienteNome:             req.ClienteNome,
		DestinazioneCaricoID:    req.DestinazioneCaricoID,
		DestinazioneCaricoNome:  req.DestinazioneCaricoNome,
		DestinazioneScaricoID:   req.DestinazioneScaricoID,
		DestinazioneScaricoNome: req.DestinazioneScaricoNome,
		DataRitiro:              req.DataRitiro,
		OraRitiroDa:             req.OraRitiroDa,
		OraRitiroA:              req.OraRitiroA,
		DataConsegna:            req.DataConsegna,
		OraConsegnaDa:           req.OraConsegnaDa,
		OraConsegnaA:            req.OraConsegnaA,
		Tariffa:                 req.Tariffa,
		TipoTariffa:             defaultString(req.TipoTariffa, "forfait"),
		Tipologia:               defaultString(req.Tipologia, "nazionale"),
		CategoriaTrasporto:      req.CategoriaTrasporto,
		RifOrdineCliente:        req.RifOrdineCliente,
		AndataRitorno:           req.AndataRitorno,
		Note:                    req.Note,
		Items:                   toOrderItems(req.Items),
		ServiziAccessori:        marshalJSON(req.ServiziAccessori),
		CostiAccessori:          marshalJSON(req.CostiAccessori),
		Stato:                   "PIANIFICABILE",
	}

	if err := s.db.WithContext(ctx).Create(&order).Error; err != nil {
		return nil, err
	}

	resp := ToResponse(order)
	return &resp, nil
}

func (s *OrderService) GetByID(ctx context.Context, id uuid.UUID) (*dto.OrderResponse, error) {
	var order models.Order
	if err := s.db.WithContext(ctx).Preload("Items").First(&order, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	resp := ToResponse(order)
	return &resp, nil
}

// Update is a full replace of the "create-able" fields only — it never
// touches stato/targa_motrice/autista_id/vettore_id/viaggio_id/fattura_id/
// progressivo, exactly like Python's update_order (which parses the request
// body as OrderCreate, a schema that doesn't include those fields at all).
func (s *OrderService) Update(ctx context.Context, id uuid.UUID, req dto.OrderRequest) (*dto.OrderResponse, error) {
	var order models.Order
	if err := s.db.WithContext(ctx).First(&order, "id = ?", id).Error; err != nil {
		return nil, err
	}

	order.ClienteID = req.ClienteID
	order.ClienteNome = req.ClienteNome
	order.DestinazioneCaricoID = req.DestinazioneCaricoID
	order.DestinazioneCaricoNome = req.DestinazioneCaricoNome
	order.DestinazioneScaricoID = req.DestinazioneScaricoID
	order.DestinazioneScaricoNome = req.DestinazioneScaricoNome
	order.DataRitiro = req.DataRitiro
	order.OraRitiroDa = req.OraRitiroDa
	order.OraRitiroA = req.OraRitiroA
	order.DataConsegna = req.DataConsegna
	order.OraConsegnaDa = req.OraConsegnaDa
	order.OraConsegnaA = req.OraConsegnaA
	order.Tariffa = req.Tariffa
	order.TipoTariffa = defaultString(req.TipoTariffa, "forfait")
	order.Tipologia = defaultString(req.Tipologia, "nazionale")
	order.CategoriaTrasporto = req.CategoriaTrasporto
	order.RifOrdineCliente = req.RifOrdineCliente
	order.AndataRitorno = req.AndataRitorno
	order.Note = req.Note
	order.ServiziAccessori = marshalJSON(req.ServiziAccessori)
	order.CostiAccessori = marshalJSON(req.CostiAccessori)

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("order_id = ?", order.ID).Delete(&models.OrderItem{}).Error; err != nil {
			return err
		}
		order.Items = toOrderItems(req.Items)
		return tx.Save(&order).Error
	})
	if err != nil {
		return nil, err
	}

	resp := ToResponse(order)
	return &resp, nil
}

// Assign mirrors PATCH /orders/{id}/assign: only valid from PIANIFICABILE,
// moves to PIANIFICATO (driver/vehicle attached, but not yet departed —
// the same target state used when orders are grouped into a Trip, see
// trips.TripService.Create/AddOrder).
func (s *OrderService) Assign(ctx context.Context, id uuid.UUID, req dto.OrderAssignRequest) (*dto.OrderResponse, error) {
	var order models.Order
	if err := s.db.WithContext(ctx).Preload("Items").First(&order, "id = ?", id).Error; err != nil {
		return nil, err
	}
	if order.Stato != "PIANIFICABILE" {
		return nil, utils.NewAPIError(400, fmt.Sprintf("L'ordine in stato %s non può essere assegnato", order.Stato))
	}

	order.TargaMotrice = req.TargaMotrice
	order.TargaRimorchio = req.TargaRimorchio
	order.AutistaID = req.AutistaID
	order.AutistaNome = req.AutistaNome
	order.VettoreID = req.VettoreID
	order.VettoreNome = req.VettoreNome
	order.Stato = "PIANIFICATO"

	if err := s.db.WithContext(ctx).Save(&order).Error; err != nil {
		return nil, err
	}
	resp := ToResponse(order)
	return &resp, nil
}

// Start mirrors PATCH /orders/{id}/start: only valid from PIANIFICATO, moves
// to VIAGGIO. Only for orders NOT attached to a Trip — an order with
// ViaggioID set must depart together with its trip (trips.TripService.Start),
// otherwise the order and its trip would go out of sync.
func (s *OrderService) Start(ctx context.Context, id uuid.UUID) (*dto.OrderResponse, error) {
	var order models.Order
	if err := s.db.WithContext(ctx).Preload("Items").First(&order, "id = ?", id).Error; err != nil {
		return nil, err
	}
	if order.ViaggioID != "" {
		return nil, utils.NewAPIError(400, "L'ordine fa parte di un viaggio: avvialo dal modulo Viaggi")
	}
	if order.Stato != "PIANIFICATO" {
		return nil, utils.NewAPIError(400, fmt.Sprintf("L'ordine in stato %s non può essere avviato. Deve essere in stato PIANIFICATO.", order.Stato))
	}

	order.Stato = "VIAGGIO"
	if err := s.db.WithContext(ctx).Save(&order).Error; err != nil {
		return nil, err
	}
	resp := ToResponse(order)
	return &resp, nil
}

// Close mirrors PATCH /orders/{id}/close: only valid from VIAGGIO, moves to CHIUSO.
func (s *OrderService) Close(ctx context.Context, id uuid.UUID) (*dto.OrderResponse, error) {
	var order models.Order
	if err := s.db.WithContext(ctx).Preload("Items").First(&order, "id = ?", id).Error; err != nil {
		return nil, err
	}
	if order.Stato != "VIAGGIO" {
		return nil, utils.NewAPIError(400, fmt.Sprintf("L'ordine in stato %s non può essere chiuso. Deve essere in stato VIAGGIO.", order.Stato))
	}

	order.Stato = "CHIUSO"
	if err := s.db.WithContext(ctx).Save(&order).Error; err != nil {
		return nil, err
	}
	resp := ToResponse(order)
	return &resp, nil
}

// Discard mirrors PATCH /orders/{id}/discard: valid from PIANIFICABILE or
// PIANIFICATO, moves to SCARTATO (terminal, cancelled). Only for orders NOT
// attached to a Trip — cancelling an order already grouped into a viaggio
// must happen from the Trip itself (out of scope here).
func (s *OrderService) Discard(ctx context.Context, id uuid.UUID) (*dto.OrderResponse, error) {
	var order models.Order
	if err := s.db.WithContext(ctx).Preload("Items").First(&order, "id = ?", id).Error; err != nil {
		return nil, err
	}
	if order.ViaggioID != "" {
		return nil, utils.NewAPIError(400, "L'ordine fa parte di un viaggio: non può essere scartato da qui")
	}
	if order.Stato != "PIANIFICABILE" && order.Stato != "PIANIFICATO" {
		return nil, utils.NewAPIError(400, fmt.Sprintf("L'ordine in stato %s non può essere scartato", order.Stato))
	}

	order.Stato = "SCARTATO"
	if err := s.db.WithContext(ctx).Save(&order).Error; err != nil {
		return nil, err
	}
	resp := ToResponse(order)
	return &resp, nil
}

// Delete mirrors DELETE /orders/{id}: only PIANIFICABILE orders can be deleted, hard delete.
func (s *OrderService) Delete(ctx context.Context, id uuid.UUID) error {
	var order models.Order
	if err := s.db.WithContext(ctx).First(&order, "id = ?", id).Error; err != nil {
		return err
	}
	if order.Stato != "PIANIFICABILE" {
		return utils.NewAPIError(400, "Solo ordini in stato PIANIFICABILE possono essere eliminati")
	}
	return s.db.WithContext(ctx).Delete(&order).Error
}

// GetCMRPDF mirrors GET /orders/{id}/cmr/pdf: resolves the consignee
// (customer) and, if assigned, the vehicle, then renders the CMR waybill.
func (s *OrderService) GetCMRPDF(ctx context.Context, id uuid.UUID) ([]byte, string, error) {
	var order models.Order
	if err := s.db.WithContext(ctx).Preload("Items").First(&order, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "", utils.NewAPIError(404, "Ordine non trovato")
		}
		return nil, "", err
	}

	var consignee models.Customer
	if err := s.db.WithContext(ctx).First(&consignee, "id = ?", order.ClienteID).Error; err != nil {
		consignee = models.Customer{RagioneSociale: order.ClienteNome}
		if consignee.RagioneSociale == "" {
			consignee.RagioneSociale = "-"
		}
	}

	var vehicle *models.Vehicle
	if order.TargaMotrice != "" {
		var v models.Vehicle
		if err := s.db.WithContext(ctx).Where("targa = ?", order.TargaMotrice).First(&v).Error; err == nil {
			vehicle = &v
		}
	}

	pdfBytes, err := pdfgen.BuildCMRPDF(order, consignee, nil, vehicle)
	if err != nil {
		return nil, "", err
	}
	return pdfBytes, pdfgen.MakeCMRFilename(order), nil
}

// ReturnSuggestions ports backend/return_orders.py's find_return_candidates
// scoring algorithm verbatim (same point values/thresholds).
func (s *OrderService) ReturnSuggestions(ctx context.Context, id uuid.UUID, maxDaysGap, limit int) (*dto.OrderReturnSuggestionsResponse, error) {
	if maxDaysGap < 0 || maxDaysGap > 14 {
		return nil, utils.NewAPIError(400, "max_days_gap deve essere tra 0 e 14")
	}
	if limit < 1 || limit > 100 {
		return nil, utils.NewAPIError(400, "limit deve essere tra 1 e 100")
	}

	var orderA models.Order
	if err := s.db.WithContext(ctx).First(&orderA, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.NewAPIError(404, "Ordine non trovato")
		}
		return nil, err
	}

	source := dto.OrderSourceSummary{
		ID:                      orderA.ID,
		Progressivo:             orderA.Progressivo,
		ClienteNome:             orderA.ClienteNome,
		DestinazioneScaricoNome: orderA.DestinazioneScaricoNome,
		DataConsegna:            orderA.DataConsegna,
	}

	scarico := strings.TrimSpace(orderA.DestinazioneScaricoNome)
	dataConsegna := strings.TrimSpace(orderA.DataConsegna)
	if scarico == "" || dataConsegna == "" {
		return &dto.OrderReturnSuggestionsResponse{Count: 0, Candidates: []dto.OrderReturnSuggestion{}, SourceOrder: source}, nil
	}

	dateTo, ok := addDays(dataConsegna, maxDaysGap)
	if !ok {
		return &dto.OrderReturnSuggestionsResponse{Count: 0, Candidates: []dto.OrderReturnSuggestion{}, SourceOrder: source}, nil
	}

	var candidates []models.Order
	term := "%" + strings.ToLower(escapeLike(scarico)) + "%"
	err := s.db.WithContext(ctx).Preload("Items").
		Where("id <> ?", id).
		Where("stato = ?", "PIANIFICABILE").
		Where("LOWER(destinazione_carico_nome) LIKE ?", term).
		Where("data_ritiro >= ? AND data_ritiro <= ?", dataConsegna, dateTo).
		Limit(limit * 3).
		Find(&candidates).Error
	if err != nil {
		return nil, err
	}

	scored := make([]dto.OrderReturnSuggestion, 0, len(candidates))
	for _, b := range candidates {
		score, reasons := scoreCandidate(orderA, b)
		if score <= 0 {
			continue
		}
		scored = append(scored, dto.OrderReturnSuggestion{Order: ToResponse(b), Score: score, Reasons: reasons})
	}
	sort.SliceStable(scored, func(i, j int) bool { return scored[i].Score > scored[j].Score })
	if len(scored) > limit {
		scored = scored[:limit]
	}

	return &dto.OrderReturnSuggestionsResponse{Count: len(scored), Candidates: scored, SourceOrder: source}, nil
}

func scoreCandidate(a, b models.Order) (int, []string) {
	score := 0
	var reasons []string

	diff, ok := dateDiffDays(a.DataConsegna, b.DataRitiro)
	switch {
	case !ok:
		reasons = append(reasons, "Date non confrontabili")
	case diff == 0:
		score += 50
		reasons = append(reasons, "Carico lo stesso giorno dello scarico")
	case diff == 1:
		score += 30
		reasons = append(reasons, "Carico il giorno dopo lo scarico")
	case diff == 2:
		score += 15
		reasons = append(reasons, "Carico due giorni dopo lo scarico")
	case diff < 0:
		return 0, []string{"Date incompatibili (carico prima dello scarico)"}
	}

	if a.ClienteID != b.ClienteID {
		score += 20
		reasons = append(reasons, "Cliente diverso (ritorno commerciale)")
	} else {
		reasons = append(reasons, "Stesso cliente (round-trip)")
	}

	if a.Tariffa > 0 && b.Tariffa >= a.Tariffa*0.7 {
		score += 10
		reasons = append(reasons, fmt.Sprintf("Tariffa ritorno EUR %.2f (>=70%% andata)", b.Tariffa))
	}

	if a.Tipologia != "" && a.Tipologia == b.Tipologia {
		score += 10
		reasons = append(reasons, fmt.Sprintf("Stessa tipologia: %s", a.Tipologia))
	}

	if score > 100 {
		score = 100
	}
	return score, reasons
}

func dateDiffDays(a, b string) (int, bool) {
	da, ok1 := parseISODate(a)
	db, ok2 := parseISODate(b)
	if !ok1 || !ok2 {
		return 0, false
	}
	return int(db.Sub(da).Hours() / 24), true
}

func addDays(dateStr string, days int) (string, bool) {
	d, ok := parseISODate(dateStr)
	if !ok {
		return "", false
	}
	return d.AddDate(0, 0, days).Format("2006-01-02"), true
}

func parseISODate(s string) (time.Time, bool) {
	if len(s) < 10 {
		return time.Time{}, false
	}
	t, err := time.Parse("2006-01-02", s[:10])
	if err != nil {
		return time.Time{}, false
	}
	return t, true
}

func defaultString(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func toOrderItems(items []dto.OrderItemDTO) []models.OrderItem {
	result := make([]models.OrderItem, len(items))
	for i, it := range items {
		result[i] = models.OrderItem{
			ID:                  uuid.New(),
			ProdottoID:          it.ProdottoID,
			ProdottoCodice:      it.ProdottoCodice,
			ProdottoDescrizione: it.ProdottoDescrizione,
			Quantita:            it.Quantita,
			Peso:                it.Peso,
		}
	}
	return result
}

func marshalJSON(v interface{}) datatypes.JSON {
	if v == nil {
		return datatypes.JSON("[]")
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

func unmarshalMaps(raw datatypes.JSON) []map[string]interface{} {
	out := []map[string]interface{}{}
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &out)
	}
	if out == nil {
		out = []map[string]interface{}{}
	}
	return out
}

func ToResponse(o models.Order) dto.OrderResponse {
	items := make([]dto.OrderItemDTO, len(o.Items))
	for i, it := range o.Items {
		items[i] = dto.OrderItemDTO{
			ProdottoID:          it.ProdottoID,
			ProdottoCodice:      it.ProdottoCodice,
			ProdottoDescrizione: it.ProdottoDescrizione,
			Quantita:            it.Quantita,
			Peso:                it.Peso,
		}
	}

	return dto.OrderResponse{
		ID:                      o.ID,
		Progressivo:             o.Progressivo,
		ClienteID:               o.ClienteID,
		ClienteNome:             o.ClienteNome,
		DestinazioneCaricoID:    o.DestinazioneCaricoID,
		DestinazioneCaricoNome:  o.DestinazioneCaricoNome,
		DestinazioneScaricoID:   o.DestinazioneScaricoID,
		DestinazioneScaricoNome: o.DestinazioneScaricoNome,
		DataRitiro:              o.DataRitiro,
		OraRitiroDa:             o.OraRitiroDa,
		OraRitiroA:              o.OraRitiroA,
		DataConsegna:            o.DataConsegna,
		OraConsegnaDa:           o.OraConsegnaDa,
		OraConsegnaA:            o.OraConsegnaA,
		Tariffa:                 o.Tariffa,
		TipoTariffa:             o.TipoTariffa,
		Tipologia:               o.Tipologia,
		CategoriaTrasporto:      o.CategoriaTrasporto,
		RifOrdineCliente:        o.RifOrdineCliente,
		AndataRitorno:           o.AndataRitorno,
		Note:                    o.Note,
		Items:                   items,
		ServiziAccessori:        unmarshalStrings(o.ServiziAccessori),
		CostiAccessori:          unmarshalMaps(o.CostiAccessori),
		Stato:                   o.Stato,
		TargaMotrice:            o.TargaMotrice,
		TargaRimorchio:          o.TargaRimorchio,
		AutistaID:               o.AutistaID,
		AutistaNome:             o.AutistaNome,
		VettoreID:               o.VettoreID,
		VettoreNome:             o.VettoreNome,
		ViaggioID:               o.ViaggioID,
		FatturaID:               o.FatturaID,
		CreatedAt:               o.CreatedAt,
		UpdatedAt:               o.UpdatedAt,
	}
}
