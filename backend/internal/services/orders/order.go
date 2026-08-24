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
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	"fratelli-feccia/internal/dto"
	"fratelli-feccia/internal/models"
	"fratelli-feccia/pkg/database"
	"fratelli-feccia/pkg/pdfgen"
	"fratelli-feccia/pkg/utils"

	"fratelli-feccia/internal/services/geo"
)

const defaultListLimit = 500

type OrderService struct {
	db  *gorm.DB
	geo *geo.GeoService
}

func NewOrderService(db *gorm.DB, orsApiKey, orsBaseURL string) *OrderService {
	return &OrderService{db: db, geo: geo.NewGeoService(db, orsApiKey, orsBaseURL)}
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

// PreloadAssociations loads every belongs-to reference an Order can
// carry (Cliente, le due Destinazioni, Garage, Autista, Vettore, WashStation,
// Items.Prodotto) — the single choke point ToResponse relies on to build the
// nested Response DTOs, reused by every load site instead of repeating the
// same 8-preload chain at each one.
func PreloadAssociations(q *gorm.DB) *gorm.DB {
	return q.
		Preload("Items.Prodotto").
		Preload("Cliente").
		Preload("Committente").
		Preload("DestinazioneCarico").
		Preload("DestinazioneScarico").
		Preload("Garage").
		Preload("Motrice").
		Preload("Semirimorchio").
		Preload("Autista").
		Preload("Vettore").
		Preload("WashStation").
		Preload("Route")
}

func (s *OrderService) reload(ctx context.Context, id uuid.UUID) (*models.Order, error) {
	var order models.Order
	if err := PreloadAssociations(s.db.WithContext(ctx)).First(&order, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

func (s *OrderService) List(ctx context.Context, f ListFilters) ([]dto.OrderResponse, error) {
	ctx, span := otel.Tracer("fratelli-feccia/orders").Start(ctx, "orders.List",
		trace.WithAttributes(
			attribute.String("orders.stato", f.Stato),
			attribute.String("orders.cliente_id", f.ClienteID),
			attribute.String("orders.search", f.Search),
			attribute.Int("orders.limit", f.Limit),
		),
	)
	defer span.End()

	limit := f.Limit
	if limit <= 0 {
		limit = defaultListLimit
	}

	query := PreloadAssociations(s.db.WithContext(ctx).Model(&models.Order{}))
	if f.Stato != "" {
		query = query.Where("orders.stato = ?", f.Stato)
	}
	if f.ClienteID != "" {
		query = query.Where("orders.cliente_id = ?", f.ClienteID)
	}
	if f.Tipologia != "" {
		query = query.Where("orders.tipologia = ?", f.Tipologia)
	}
	if f.DataDa != "" {
		query = query.Where("orders.data_ritiro >= ?", f.DataDa)
	}
	if f.DataA != "" {
		query = query.Where("orders.data_ritiro <= ?", f.DataA)
	}
	if f.Search != "" {
		term := "%" + strings.ToLower(escapeLike(f.Search)) + "%"
		query = query.
			Joins("LEFT JOIN customers ON customers.id = orders.cliente_id").
			Joins("LEFT JOIN destinations dest_carico ON dest_carico.id = orders.destinazione_carico_id").
			Joins("LEFT JOIN destinations dest_scarico ON dest_scarico.id = orders.destinazione_scarico_id").
			Where(
				"LOWER(customers.ragione_sociale) LIKE ? OR LOWER(orders.progressivo) LIKE ? OR LOWER(orders.rif_ordine_cliente) LIKE ? OR LOWER(dest_carico.nome) LIKE ? OR LOWER(dest_scarico.nome) LIKE ?",
				term, term, term, term, term,
			)
	}

	queryCtx, querySpan := otel.Tracer("fratelli-feccia/orders").Start(ctx, "orders.List.query")
	var orders []models.Order
	err := query.WithContext(queryCtx).Order("orders.created_at DESC").Limit(limit).Find(&orders).Error
	if err != nil {
		querySpan.RecordError(err)
		querySpan.SetStatus(codes.Error, err.Error())
	}
	querySpan.SetAttributes(attribute.Int("orders.result_count", len(orders)))
	querySpan.End()
	if err != nil {
		return nil, err
	}

	_, mapSpan := otel.Tracer("fratelli-feccia/orders").Start(ctx, "orders.List.map")
	result := make([]dto.OrderResponse, len(orders))
	for i, o := range orders {
		result[i] = ToResponse(o)
	}
	mapSpan.End()

	span.SetAttributes(attribute.Int("orders.result_count", len(result)))
	return result, nil
}

func (s *OrderService) Create(ctx context.Context, req dto.OrderRequest) (*dto.OrderResponse, error) {
	seq, err := database.NextSequence(s.db.WithContext(ctx), "orders")
	if err != nil {
		return nil, err
	}
	progressivo := fmt.Sprintf("%s/%04d", time.Now().Format("06"), seq)

	clienteID, err := utils.ParseUUID(req.ClienteID)
	if err != nil {
		return nil, err
	}
	committenteID, err := utils.ParseOptionalUUID(req.CommittenteID)
	if err != nil {
		return nil, err
	}
	caricoID, err := utils.ParseOptionalUUID(req.DestinazioneCaricoID)
	if err != nil {
		return nil, err
	}
	scaricoID, err := utils.ParseOptionalUUID(req.DestinazioneScaricoID)
	if err != nil {
		return nil, err
	}
	items, err := toOrderItems(req.Items)
	if err != nil {
		return nil, err
	}

	order := models.Order{
		ID:                    uuid.New(),
		Progressivo:           progressivo,
		ClienteID:             clienteID,
		CommittenteID:         committenteID,
		DestinazioneCaricoID:  caricoID,
		DestinazioneScaricoID: scaricoID,
		DataRitiro:            req.DataRitiro,
		OraRitiroDa:           req.OraRitiroDa,
		OraRitiroA:            req.OraRitiroA,
		DataConsegna:          req.DataConsegna,
		OraConsegnaDa:         req.OraConsegnaDa,
		OraConsegnaA:          req.OraConsegnaA,
		Tariffa:               req.Tariffa,
		TipoTariffa:           defaultString(req.TipoTariffa, "forfait"),
		Tipologia:             defaultString(req.Tipologia, "nazionale"),
		CategoriaTrasporto:    req.CategoriaTrasporto,
		RifOrdineCliente:      req.RifOrdineCliente,
		RifCarico:             req.RifCarico,
		NoteCarico:            req.NoteCarico,
		RifScarico:            req.RifScarico,
		NoteScarico:           req.NoteScarico,
		AndataRitorno:         req.AndataRitorno,
		Provvisorio:           req.Provvisorio,
		Note:                  req.Note,
		Items:                 items,
		ServiziAccessori:      marshalJSON(req.ServiziAccessori),
		CostiAccessori:        marshalJSON(req.CostiAccessori),
		Stato:                 models.OrderStatoPianificabile,
	}

	if err := s.db.WithContext(ctx).Create(&order).Error; err != nil {
		return nil, err
	}

	reloaded, err := s.reload(ctx, order.ID)
	if err != nil {
		return nil, err
	}
	resp := ToResponse(*reloaded)
	return &resp, nil
}

func (s *OrderService) GetByID(ctx context.Context, id uuid.UUID) (*dto.OrderResponse, error) {
	order, err := s.reload(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	resp := ToResponse(*order)
	return &resp, nil
}

// Update is a full replace of the "create-able" fields only — it never
// touches stato/motrice_id/semirimorchio_id/autista_id/vettore_id/viaggio_id/
// fattura_id/progressivo, exactly like Python's update_order (which parses the request
// body as OrderCreate, a schema that doesn't include those fields at all).
func (s *OrderService) Update(ctx context.Context, id uuid.UUID, req dto.OrderRequest) (*dto.OrderResponse, error) {
	var order models.Order
	if err := s.db.WithContext(ctx).First(&order, "id = ?", id).Error; err != nil {
		return nil, err
	}

	clienteID, err := utils.ParseUUID(req.ClienteID)
	if err != nil {
		return nil, err
	}
	committenteID, err := utils.ParseOptionalUUID(req.CommittenteID)
	if err != nil {
		return nil, err
	}
	caricoID, err := utils.ParseOptionalUUID(req.DestinazioneCaricoID)
	if err != nil {
		return nil, err
	}
	scaricoID, err := utils.ParseOptionalUUID(req.DestinazioneScaricoID)
	if err != nil {
		return nil, err
	}
	items, err := toOrderItems(req.Items)
	if err != nil {
		return nil, err
	}

	order.ClienteID = clienteID
	order.CommittenteID = committenteID
	order.DestinazioneCaricoID = caricoID
	order.DestinazioneScaricoID = scaricoID
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
	order.RifCarico = req.RifCarico
	order.NoteCarico = req.NoteCarico
	order.RifScarico = req.RifScarico
	order.NoteScarico = req.NoteScarico
	order.AndataRitorno = req.AndataRitorno
	order.Provvisorio = req.Provvisorio
	order.Note = req.Note
	order.ServiziAccessori = marshalJSON(req.ServiziAccessori)
	order.CostiAccessori = marshalJSON(req.CostiAccessori)

	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("order_id = ?", order.ID).Delete(&models.OrderItem{}).Error; err != nil {
			return err
		}
		order.Items = items
		return tx.Save(&order).Error
	})
	if err != nil {
		return nil, err
	}

	reloaded, err := s.reload(ctx, order.ID)
	if err != nil {
		return nil, err
	}
	resp := ToResponse(*reloaded)
	return &resp, nil
}

// Assign mirrors PATCH /orders/{id}/assign: only valid from PIANIFICABILE,
// moves to PIANIFICATO (driver/vehicle attached, but not yet departed —
// the same target state used when orders are grouped into a Trip, see
// trips.TripService.Create/AddOrder).
func (s *OrderService) Assign(ctx context.Context, id uuid.UUID, req dto.OrderAssignRequest) (*dto.OrderResponse, error) {
	var order models.Order
	if err := s.db.WithContext(ctx).First(&order, "id = ?", id).Error; err != nil {
		return nil, err
	}
	if order.Stato != models.OrderStatoPianificabile {
		return nil, utils.NewAPIError(400, fmt.Sprintf("L'ordine in stato %s non può essere assegnato", order.Stato))
	}

	garageID, err := utils.ParseOptionalUUID(req.GarageID)
	if err != nil {
		return nil, err
	}
	motriceID, err := utils.ParseOptionalUUID(req.MotriceID)
	if err != nil {
		return nil, err
	}
	semirimorchioID, err := utils.ParseOptionalUUID(req.SemirimorchioID)
	if err != nil {
		return nil, err
	}
	autistaID, err := utils.ParseOptionalUUID(req.AutistaID)
	if err != nil {
		return nil, err
	}
	vettoreID, err := utils.ParseOptionalUUID(req.VettoreID)
	if err != nil {
		return nil, err
	}
	washStationID, err := utils.ParseOptionalUUID(req.WashStationID)
	if err != nil {
		return nil, err
	}

	order.GarageID = garageID
	order.MotriceID = motriceID
	order.SemirimorchioID = semirimorchioID
	order.AutistaID = autistaID
	order.VettoreID = vettoreID
	order.WashStationID = washStationID
	order.Stato = models.OrderStatoPianificato

	if err := s.db.WithContext(ctx).Save(&order).Error; err != nil {
		return nil, err
	}

	// Il manager ha già scelto una delle alternative proposte in form: la
	// calcoliamo qui (mai fidandoci della geometria del client) così che
	// l'ordine abbia già un percorso pronto nel momento esatto in cui passa
	// a PIANIFICATO. Se manca (form saltato) o il routing non è disponibile,
	// l'assegnazione resta comunque valida — solo senza route calcolata.
	if len(req.RouteWaypoints) >= 2 {
		resolved := make([]routeWaypoint, len(req.RouteWaypoints))
		namedPoints := make([]geo.NamedPoint, len(req.RouteWaypoints))
		for i, w := range req.RouteWaypoints {
			wp, err := s.resolveWaypoint(ctx, w.Tipo, w.RefID)
			if err != nil {
				return nil, err
			}
			resolved[i] = wp
			namedPoints[i] = geo.NamedPoint{Name: wp.Nome, Lat: wp.Lat, Lng: wp.Lng}
		}
		if route := s.geo.GetRoadRouteMultiWaypoint(ctx, namedPoints); route != nil {
			if err := s.upsertOrderRoute(ctx, &order, resolved, route, false); err != nil {
				return nil, err
			}
		}
	}

	reloaded, err := s.reload(ctx, order.ID)
	if err != nil {
		return nil, err
	}
	resp := ToResponse(*reloaded)
	return &resp, nil
}

// Unassign mirrors PATCH /orders/{id}/unassign: only valid from PIANIFICATO,
// moves back to PIANIFICABILE — the reverse of Assign. A PIANIFICATO order
// is locked (no route/waypoint edits, see UpdateRoute's callers): to change
// garage/mezzo/autista/vettore/wash_station or the route, the manager must
// unassign first, edit via AssignOrderForm, then Assign again. Only for
// orders NOT attached to a Trip, same restriction as Start/Discard.
func (s *OrderService) Unassign(ctx context.Context, id uuid.UUID) (*dto.OrderResponse, error) {
	var order models.Order
	if err := s.db.WithContext(ctx).First(&order, "id = ?", id).Error; err != nil {
		return nil, err
	}
	if order.ViaggioID != nil {
		return nil, utils.NewAPIError(400, "L'ordine fa parte di un viaggio: rimuovilo dal viaggio per modificarlo")
	}
	if order.Stato != models.OrderStatoPianificato {
		return nil, utils.NewAPIError(400, fmt.Sprintf("L'ordine in stato %s non può essere riportato a pianificabile. Deve essere in stato PIANIFICATO.", order.Stato))
	}

	oldRouteID := order.RouteID

	order.Stato = models.OrderStatoPianificabile
	order.GarageID = nil
	order.MotriceID = nil
	order.SemirimorchioID = nil
	order.AutistaID = nil
	order.VettoreID = nil
	order.WashStationID = nil
	order.RouteID = nil

	// Il vecchio OrderRoute va rimosso solo dopo che l'ordine non lo
	// referenzia più (route_id azzerato) — altrimenti la FK fk_orders_route
	// blocca la DELETE (violazione vista su Postgres, non su SQLite dove i
	// test girano perché lì i FK non sono enforced di default).
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&order).Error; err != nil {
			return err
		}
		if oldRouteID != nil {
			return tx.Delete(&models.OrderRoute{}, "id = ?", *oldRouteID).Error
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	reloaded, err := s.reload(ctx, order.ID)
	if err != nil {
		return nil, err
	}
	resp := ToResponse(*reloaded)
	return &resp, nil
}

// Start mirrors PATCH /orders/{id}/start: only valid from PIANIFICATO, moves
// to VIAGGIO. Only for orders NOT attached to a Trip — an order with
// ViaggioID set must depart together with its trip (trips.TripService.Start),
// otherwise the order and its trip would go out of sync.
func (s *OrderService) Start(ctx context.Context, id uuid.UUID) (*dto.OrderResponse, error) {
	var order models.Order
	if err := s.db.WithContext(ctx).First(&order, "id = ?", id).Error; err != nil {
		return nil, err
	}
	if order.ViaggioID != nil {
		return nil, utils.NewAPIError(400, "L'ordine fa parte di un viaggio: avvialo dal modulo Viaggi")
	}
	if order.Stato != models.OrderStatoPianificato {
		return nil, utils.NewAPIError(400, fmt.Sprintf("L'ordine in stato %s non può essere avviato. Deve essere in stato PIANIFICATO.", order.Stato))
	}

	order.Stato = models.OrderStatoViaggio
	if err := s.db.WithContext(ctx).Save(&order).Error; err != nil {
		return nil, err
	}
	reloaded, err := s.reload(ctx, order.ID)
	if err != nil {
		return nil, err
	}
	resp := ToResponse(*reloaded)
	return &resp, nil
}

// Close mirrors PATCH /orders/{id}/close: only valid from VIAGGIO, moves to CHIUSO.
func (s *OrderService) Close(ctx context.Context, id uuid.UUID) (*dto.OrderResponse, error) {
	var order models.Order
	if err := s.db.WithContext(ctx).First(&order, "id = ?", id).Error; err != nil {
		return nil, err
	}
	if order.Stato != models.OrderStatoViaggio {
		return nil, utils.NewAPIError(400, fmt.Sprintf("L'ordine in stato %s non può essere chiuso. Deve essere in stato VIAGGIO.", order.Stato))
	}

	order.Stato = models.OrderStatoChiuso
	if err := s.db.WithContext(ctx).Save(&order).Error; err != nil {
		return nil, err
	}
	reloaded, err := s.reload(ctx, order.ID)
	if err != nil {
		return nil, err
	}
	resp := ToResponse(*reloaded)
	return &resp, nil
}

// Discard mirrors PATCH /orders/{id}/discard: valid from PIANIFICABILE or
// PIANIFICATO, moves to SCARTATO (terminal, cancelled). Only for orders NOT
// attached to a Trip — cancelling an order already grouped into a viaggio
// must happen from the Trip itself (out of scope here).
func (s *OrderService) Discard(ctx context.Context, id uuid.UUID) (*dto.OrderResponse, error) {
	var order models.Order
	if err := s.db.WithContext(ctx).First(&order, "id = ?", id).Error; err != nil {
		return nil, err
	}
	if order.ViaggioID != nil {
		return nil, utils.NewAPIError(400, "L'ordine fa parte di un viaggio: non può essere scartato da qui")
	}
	if order.Stato != models.OrderStatoPianificabile && order.Stato != models.OrderStatoPianificato {
		return nil, utils.NewAPIError(400, fmt.Sprintf("L'ordine in stato %s non può essere scartato", order.Stato))
	}

	order.Stato = models.OrderStatoScartato
	if err := s.db.WithContext(ctx).Save(&order).Error; err != nil {
		return nil, err
	}
	reloaded, err := s.reload(ctx, order.ID)
	if err != nil {
		return nil, err
	}
	resp := ToResponse(*reloaded)
	return &resp, nil
}

// Delete mirrors DELETE /orders/{id}: only PIANIFICABILE orders can be deleted, hard delete.
func (s *OrderService) Delete(ctx context.Context, id uuid.UUID) error {
	var order models.Order
	if err := s.db.WithContext(ctx).First(&order, "id = ?", id).Error; err != nil {
		return err
	}
	if order.Stato != models.OrderStatoPianificabile {
		return utils.NewAPIError(400, "Solo ordini in stato PIANIFICABILE possono essere eliminati")
	}
	return s.db.WithContext(ctx).Delete(&order).Error
}

// GetCMRPDF mirrors GET /orders/{id}/cmr/pdf: resolves the consignee
// (customer) and, if assigned, the vehicle, then renders the CMR waybill.
func (s *OrderService) GetCMRPDF(ctx context.Context, id uuid.UUID) ([]byte, string, error) {
	order, err := s.reload(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "", utils.NewAPIError(404, "Ordine non trovato")
		}
		return nil, "", err
	}

	// Cliente is Preloaded above; customers are soft-deleted (Active flag),
	// never hard-deleted, so a missing association here only happens on a
	// genuine data-integrity gap — fall back to a placeholder rather than a
	// second lookup (there's no denormalized name snapshot left to recover).
	consignee := order.Cliente
	if consignee.ID == uuid.Nil {
		consignee = models.Customer{RagioneSociale: "-"}
	}

	pdfBytes, err := pdfgen.BuildCMRPDF(*order, consignee, nil, order.Motrice, order.Semirimorchio)
	if err != nil {
		return nil, "", err
	}
	return pdfBytes, pdfgen.MakeCMRFilename(*order), nil
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

	orderA, err := s.reload(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.NewAPIError(404, "Ordine non trovato")
		}
		return nil, err
	}

	source := dto.OrderSourceSummary{
		ID:                  orderA.ID,
		Progressivo:         orderA.Progressivo,
		Cliente:             customerResponse(orderA.Cliente),
		DestinazioneScarico: destinationResponse(orderA.DestinazioneScarico),
		DataConsegna:        orderA.DataConsegna,
	}

	scaricoNome := ""
	if orderA.DestinazioneScarico != nil {
		scaricoNome = strings.TrimSpace(orderA.DestinazioneScarico.Nome)
	}
	dataConsegna := strings.TrimSpace(orderA.DataConsegna)
	if scaricoNome == "" || dataConsegna == "" {
		return &dto.OrderReturnSuggestionsResponse{Count: 0, Candidates: []dto.OrderReturnSuggestion{}, SourceOrder: source}, nil
	}

	dateTo, ok := addDays(dataConsegna, maxDaysGap)
	if !ok {
		return &dto.OrderReturnSuggestionsResponse{Count: 0, Candidates: []dto.OrderReturnSuggestion{}, SourceOrder: source}, nil
	}

	var candidates []models.Order
	term := "%" + strings.ToLower(escapeLike(scaricoNome)) + "%"
	query := PreloadAssociations(s.db.WithContext(ctx).Model(&models.Order{})).
		Joins("LEFT JOIN destinations dest_carico ON dest_carico.id = orders.destinazione_carico_id").
		Where("orders.id <> ?", id).
		Where("orders.stato = ?", string(models.OrderStatoPianificabile)).
		Where("LOWER(dest_carico.nome) LIKE ?", term).
		Where("orders.data_ritiro >= ? AND orders.data_ritiro <= ?", dataConsegna, dateTo).
		Limit(limit * 3)
	if err := query.Find(&candidates).Error; err != nil {
		return nil, err
	}

	scored := make([]dto.OrderReturnSuggestion, 0, len(candidates))
	for _, b := range candidates {
		score, reasons := scoreCandidate(*orderA, b)
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

func toOrderItems(items []dto.OrderItemRequestDTO) ([]models.OrderItem, error) {
	result := make([]models.OrderItem, len(items))
	for i, it := range items {
		prodottoID, err := utils.ParseUUID(it.ProdottoID)
		if err != nil {
			return nil, err
		}
		result[i] = models.OrderItem{
			ID:         uuid.New(),
			ProdottoID: prodottoID,
			Quantita:   it.Quantita,
			Peso:       it.Peso,
		}
	}
	return result, nil
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

// uuidPtrString renders an optional FK for the wire: "" when unset, matching
// the pre-migration contract the frontend already expects for these fields.
func uuidPtrString(id *uuid.UUID) string {
	if id == nil {
		return ""
	}
	return id.String()
}

// committenteResponse adapts the nullable Committente association to
// customerResponse (which takes the always-present Cliente by value).
func committenteResponse(c *models.Customer) *dto.CustomerResponse {
	if c == nil {
		return nil
	}
	return customerResponse(*c)
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

func garageResponse(g *models.Garage) *dto.GarageResponse {
	if g == nil {
		return nil
	}
	return &dto.GarageResponse{
		ID: g.ID, Nome: g.Nome, Indirizzo: g.Indirizzo, Citta: g.Citta, Lat: g.Lat, Lng: g.Lng,
		Note: g.Note, Active: g.Active, CreatedAt: g.CreatedAt, UpdatedAt: g.UpdatedAt,
	}
}

func washStationResponse(w *models.WashStation) *dto.WashStationResponse {
	if w == nil {
		return nil
	}
	return &dto.WashStationResponse{
		ID: w.ID, Nome: w.Nome, Tipo: w.Tipo, Indirizzo: w.Indirizzo, Citta: w.Citta, Lat: w.Lat, Lng: w.Lng,
		Note: w.Note, Active: w.Active, CreatedAt: w.CreatedAt, UpdatedAt: w.UpdatedAt,
	}
}

func motriceResponse(m *models.Motrice) *dto.MotriceResponse {
	if m == nil {
		return nil
	}
	return &dto.MotriceResponse{
		ID: m.ID, Targa: m.Targa, Marca: m.Marca, Modello: m.Modello, Anno: m.Anno,
		PortataKg: m.PortataKg, Note: m.Note, Active: m.Active, CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt,
	}
}

func semirimorchioResponse(r *models.Semirimorchio) *dto.SemirimorchioResponse {
	if r == nil {
		return nil
	}
	return &dto.SemirimorchioResponse{
		ID: r.ID, Targa: r.Targa, Tipo: r.Tipo, Scompartature: r.Scompartature,
		PortataKg: r.PortataKg, Note: r.Note, Active: r.Active, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt,
	}
}

func driverResponse(d *models.Driver) *dto.DriverResponse {
	if d == nil {
		return nil
	}
	return &dto.DriverResponse{
		ID: d.ID, Nome: d.Nome, Cognome: d.Cognome, CodiceFiscale: d.CodiceFiscale,
		Patente: unmarshalStrings(d.Patente), ScadenzaPatente: d.ScadenzaPatente, Telefono: d.Telefono,
		Email: d.Email, Note: d.Note, Active: d.Active, CreatedAt: d.CreatedAt, UpdatedAt: d.UpdatedAt,
	}
}

func carrierResponse(c *models.Carrier) *dto.CarrierResponse {
	if c == nil {
		return nil
	}
	return &dto.CarrierResponse{
		ID: c.ID, RagioneSociale: c.RagioneSociale, PartitaIva: c.PartitaIva, Indirizzo: c.Indirizzo,
		Citta: c.Citta, Telefono: c.Telefono, Email: c.Email, Note: c.Note, Active: c.Active,
		CreatedAt: c.CreatedAt, UpdatedAt: c.UpdatedAt,
	}
}

func productResponse(p *models.Product) *dto.ProductResponse {
	if p == nil || p.ID == uuid.Nil {
		return nil
	}
	return &dto.ProductResponse{
		ID: p.ID, Codice: p.Codice, Descrizione: p.Descrizione, UnitaMisura: p.UnitaMisura,
		Note: p.Note, Active: p.Active, CreatedAt: p.CreatedAt, UpdatedAt: p.UpdatedAt,
	}
}

func ToResponse(o models.Order) dto.OrderResponse {
	items := make([]dto.OrderItemResponseDTO, len(o.Items))
	for i, it := range o.Items {
		prod := it.Prodotto
		items[i] = dto.OrderItemResponseDTO{
			Prodotto: productResponse(&prod),
			Quantita: it.Quantita,
			Peso:     it.Peso,
		}
	}

	return dto.OrderResponse{
		ID:                    o.ID,
		Progressivo:           o.Progressivo,
		ClienteID:             o.ClienteID.String(),
		Cliente:               customerResponse(o.Cliente),
		CommittenteID:         uuidPtrString(o.CommittenteID),
		Committente:           committenteResponse(o.Committente),
		DestinazioneCaricoID:  uuidPtrString(o.DestinazioneCaricoID),
		DestinazioneCarico:    destinationResponse(o.DestinazioneCarico),
		DestinazioneScaricoID: uuidPtrString(o.DestinazioneScaricoID),
		DestinazioneScarico:   destinationResponse(o.DestinazioneScarico),
		DataRitiro:            o.DataRitiro,
		OraRitiroDa:           o.OraRitiroDa,
		OraRitiroA:            o.OraRitiroA,
		DataConsegna:          o.DataConsegna,
		OraConsegnaDa:         o.OraConsegnaDa,
		OraConsegnaA:          o.OraConsegnaA,
		Tariffa:               o.Tariffa,
		TipoTariffa:           o.TipoTariffa,
		Tipologia:             o.Tipologia,
		CategoriaTrasporto:    o.CategoriaTrasporto,
		RifOrdineCliente:      o.RifOrdineCliente,
		RifCarico:             o.RifCarico,
		NoteCarico:            o.NoteCarico,
		RifScarico:            o.RifScarico,
		NoteScarico:           o.NoteScarico,
		AndataRitorno:         o.AndataRitorno,
		Provvisorio:           o.Provvisorio,
		Note:                  o.Note,
		Items:                 items,
		ServiziAccessori:      unmarshalStrings(o.ServiziAccessori),
		CostiAccessori:        unmarshalMaps(o.CostiAccessori),
		Stato:                 string(o.Stato),
		GarageID:              uuidPtrString(o.GarageID),
		Garage:                garageResponse(o.Garage),
		MotriceID:             uuidPtrString(o.MotriceID),
		Motrice:               motriceResponse(o.Motrice),
		SemirimorchioID:       uuidPtrString(o.SemirimorchioID),
		Semirimorchio:         semirimorchioResponse(o.Semirimorchio),
		AutistaID:             uuidPtrString(o.AutistaID),
		Autista:               driverResponse(o.Autista),
		VettoreID:             uuidPtrString(o.VettoreID),
		Vettore:               carrierResponse(o.Vettore),
		WashStationID:         uuidPtrString(o.WashStationID),
		WashStation:           washStationResponse(o.WashStation),
		RouteID:               uuidPtrString(o.RouteID),
		Route:                 routeResponse(o.Route),
		ViaggioID:             uuidPtrString(o.ViaggioID),
		FatturaID:             uuidPtrString(o.FatturaID),
		CreatedAt:             o.CreatedAt,
		UpdatedAt:             o.UpdatedAt,
	}
}
