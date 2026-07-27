package trips

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"sort"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	"fratelli-feccia/internal/dto"
	"fratelli-feccia/internal/models"
	"fratelli-feccia/internal/services/geo"
	"fratelli-feccia/internal/services/orders"
	"fratelli-feccia/pkg/pdfgen"
	"fratelli-feccia/pkg/utils"
)

const defaultListLimit = 200

type TripService struct {
	db  *gorm.DB
	geo *geo.GeoService
}

func NewTripService(db *gorm.DB, orsApiKey, orsBaseURL string) *TripService {
	return &TripService{db: db, geo: geo.NewGeoService(db, orsApiKey, orsBaseURL)}
}

// preloadTripAssociations loads every belongs-to reference a Trip can carry
// (Autista, Vettore, Garage) plus Segments — the choke point toResponse
// relies on to build the nested Response DTOs.
func preloadTripAssociations(q *gorm.DB) *gorm.DB {
	return q.Preload("Segments").Preload("Autista").Preload("Vettore").Preload("Garage")
}

func (s *TripService) reload(ctx context.Context, id uuid.UUID) (*models.Trip, error) {
	var trip models.Trip
	if err := preloadTripAssociations(s.db.WithContext(ctx)).First(&trip, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &trip, nil
}

func (s *TripService) List(ctx context.Context, stato string, limit int) ([]dto.TripResponse, error) {
	if limit <= 0 {
		limit = defaultListLimit
	}
	query := preloadTripAssociations(s.db.WithContext(ctx))
	if stato != "" {
		query = query.Where("stato = ?", stato)
	}
	var trips []models.Trip
	if err := query.Order("created_at DESC").Limit(limit).Find(&trips).Error; err != nil {
		return nil, err
	}
	result := make([]dto.TripResponse, len(trips))
	for i, t := range trips {
		result[i] = toResponse(t)
	}
	return result, nil
}

// Create mirrors create_trip: syncs any PIANIFICABILE orders in
// req.OrdiniIds to PIANIFICATO (silently skipping ones not in that state,
// matching Python's filtered update_one), then computes segments via OSRM.
// The trip itself starts PIANIFICATO too — orders and trip only move to
// VIAGGIO/IN_CORSO together, via Start.
func (s *TripService) Create(ctx context.Context, req dto.TripRequest) (*dto.TripResponse, error) {
	autistaID, err := utils.ParseOptionalUUID(req.AutistaID)
	if err != nil {
		return nil, err
	}
	vettoreID, err := utils.ParseOptionalUUID(req.VettoreID)
	if err != nil {
		return nil, err
	}
	garageID, err := utils.ParseOptionalUUID(req.GarageID)
	if err != nil {
		return nil, err
	}

	trip := models.Trip{
		ID:             uuid.New(),
		OrdiniIds:      marshalStrings(req.OrdiniIds),
		TargaMotrice:   req.TargaMotrice,
		TargaRimorchio: req.TargaRimorchio,
		AutistaID:      autistaID,
		VettoreID:      vettoreID,
		GarageID:       garageID,
		Note:           req.Note,
		DataPartenza:   req.DataPartenza,
		DataArrivo:     req.DataArrivo,
		Stato:          "PIANIFICATO",
	}

	for _, orderID := range req.OrdiniIds {
		s.db.WithContext(ctx).Model(&models.Order{}).
			Where("id = ? AND stato = ?", orderID, "PIANIFICABILE").
			Updates(map[string]interface{}{
				"stato":           "PIANIFICATO",
				"viaggio_id":      trip.ID,
				"targa_motrice":   req.TargaMotrice,
				"targa_rimorchio": req.TargaRimorchio,
				"autista_id":      autistaID,
				"vettore_id":      vettoreID,
				"updated_at":      time.Now().UTC(),
			})
	}

	trip.Segments = s.buildSegments(ctx, garageID, req.OrdiniIds)
	trip.KmTotali = totalKm(trip.Segments)

	if err := s.db.WithContext(ctx).Create(&trip).Error; err != nil {
		return nil, err
	}
	reloaded, err := s.reload(ctx, trip.ID)
	if err != nil {
		return nil, err
	}
	resp := toResponse(*reloaded)
	return &resp, nil
}

func (s *TripService) GetByID(ctx context.Context, id uuid.UUID) (*dto.TripDetailResponse, error) {
	trip, err := s.reload(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	var linkedOrders []models.Order
	if err := orders.PreloadAssociations(s.db.WithContext(ctx)).Where("viaggio_id = ?", id).Find(&linkedOrders).Error; err != nil {
		return nil, err
	}
	ordini := make([]dto.OrderResponse, len(linkedOrders))
	for i, o := range linkedOrders {
		ordini[i] = orders.ToResponse(o)
	}

	return &dto.TripDetailResponse{TripResponse: toResponse(*trip), Ordini: ordini}, nil
}

// GetInstructionsPDF mirrors GET /trips/{id}/instructions/pdf: renders the
// driver's operational service sheet for the trip. Generated on the fly,
// never archived (the trip changes often until completion).
func (s *TripService) GetInstructionsPDF(ctx context.Context, id uuid.UUID) ([]byte, string, error) {
	trip, err := s.reload(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "", utils.NewAPIError(404, "Viaggio non trovato")
		}
		return nil, "", err
	}

	var linkedOrders []models.Order
	if err := orders.PreloadAssociations(s.db.WithContext(ctx)).Where("viaggio_id = ?", id).
		Order("data_ritiro ASC, created_at ASC").Find(&linkedOrders).Error; err != nil {
		return nil, "", err
	}

	var driver *models.Driver
	if trip.AutistaID != nil {
		var d models.Driver
		if err := s.db.WithContext(ctx).First(&d, "id = ?", *trip.AutistaID).Error; err == nil {
			driver = &d
		}
	}

	pdfBytes, err := pdfgen.BuildInstructionsPDF(*trip, linkedOrders, trip.Segments, driver)
	if err != nil {
		return nil, "", err
	}
	return pdfBytes, pdfgen.MakeInstructionsFilename(*trip), nil
}

// RecomputeSegments mirrors POST /trips/{id}/recompute-segments.
func (s *TripService) RecomputeSegments(ctx context.Context, id uuid.UUID) (*dto.RecomputeSegmentsResult, error) {
	var trip models.Trip
	if err := s.db.WithContext(ctx).First(&trip, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.NewAPIError(404, "Viaggio non trovato")
		}
		return nil, err
	}

	ordiniIds := unmarshalStrings(trip.OrdiniIds)
	segments := s.buildSegments(ctx, trip.GarageID, ordiniIds)
	km := totalKm(segments)

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("trip_id = ?", trip.ID).Delete(&models.TripSegment{}).Error; err != nil {
			return err
		}
		for i := range segments {
			segments[i].TripID = trip.ID
		}
		if len(segments) > 0 {
			if err := tx.Create(&segments).Error; err != nil {
				return err
			}
		}
		return tx.Model(&models.Trip{}).Where("id = ?", trip.ID).Update("km_totali", km).Error
	})
	if err != nil {
		return nil, err
	}

	return &dto.RecomputeSegmentsResult{OK: true, SegmentiCount: len(segments), KmTotali: km}, nil
}

// Start mirrors PATCH /trips/{id}/start: only valid from PIANIFICATO, moves
// the trip to IN_CORSO and bulk-starts its still-PIANIFICATO orders to
// VIAGGIO (trip and its orders always depart together).
func (s *TripService) Start(ctx context.Context, id uuid.UUID) (*dto.OKResult, error) {
	var trip models.Trip
	if err := s.db.WithContext(ctx).First(&trip, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.NewAPIError(404, "Viaggio non trovato")
		}
		return nil, err
	}
	if trip.Stato != "PIANIFICATO" {
		return nil, utils.NewAPIError(400, fmt.Sprintf("Il viaggio in stato %s non può essere avviato. Deve essere in stato PIANIFICATO.", trip.Stato))
	}

	if err := s.db.WithContext(ctx).Model(&models.Order{}).
		Where("viaggio_id = ? AND stato = ?", id, "PIANIFICATO").
		Updates(map[string]interface{}{"stato": "VIAGGIO", "updated_at": time.Now().UTC()}).Error; err != nil {
		return nil, err
	}

	if err := s.db.WithContext(ctx).Model(&models.Trip{}).Where("id = ?", id).Update("stato", "IN_CORSO").Error; err != nil {
		return nil, err
	}

	return &dto.OKResult{OK: true}, nil
}

// Complete mirrors PATCH /trips/{id}/complete: only valid from IN_CORSO,
// bulk-closes the trip's VIAGGIO orders and marks the trip COMPLETATO.
func (s *TripService) Complete(ctx context.Context, id uuid.UUID) (*dto.OKResult, error) {
	var trip models.Trip
	if err := s.db.WithContext(ctx).First(&trip, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.NewAPIError(404, "Viaggio non trovato")
		}
		return nil, err
	}
	if trip.Stato != "IN_CORSO" {
		return nil, utils.NewAPIError(400, fmt.Sprintf("Il viaggio in stato %s non può essere completato. Deve essere in stato IN_CORSO.", trip.Stato))
	}

	if err := s.db.WithContext(ctx).Model(&models.Order{}).
		Where("viaggio_id = ? AND stato = ?", id, "VIAGGIO").
		Updates(map[string]interface{}{"stato": "CHIUSO", "updated_at": time.Now().UTC()}).Error; err != nil {
		return nil, err
	}

	if err := s.db.WithContext(ctx).Model(&models.Trip{}).Where("id = ?", id).Update("stato", "COMPLETATO").Error; err != nil {
		return nil, err
	}

	return &dto.OKResult{OK: true}, nil
}

// AddOrder mirrors PATCH /trips/{id}/add-order: appends an order to the
// trip's ordini_ids, then moves the order to the trip's vehicle/driver (not
// carrier — matching Python's dict, which omits vettore_id here unlike
// Create). Target stato mirrors the trip's own state: if the trip has
// already departed (IN_CORSO) the added order joins directly as VIAGGIO
// (the truck is already moving); otherwise it becomes PIANIFICATO like the
// trip's other orders, and departs together with it via Start.
func (s *TripService) AddOrder(ctx context.Context, tripID, orderID uuid.UUID) (*dto.OKResult, error) {
	var trip models.Trip
	if err := s.db.WithContext(ctx).First(&trip, "id = ?", tripID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.NewAPIError(404, "Viaggio non trovato")
		}
		return nil, err
	}

	var order models.Order
	if err := s.db.WithContext(ctx).First(&order, "id = ?", orderID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.NewAPIError(404, "Ordine non trovato")
		}
		return nil, err
	}
	if order.Stato != models.OrderStatoPianificabile {
		return nil, utils.NewAPIError(400, "L'ordine deve essere in stato PIANIFICABILE")
	}

	ordiniIds := unmarshalStrings(trip.OrdiniIds)
	ordiniIds = append(ordiniIds, orderID.String())
	trip.OrdiniIds = marshalStrings(ordiniIds)
	if err := s.db.WithContext(ctx).Model(&models.Trip{}).Where("id = ?", tripID).Update("ordini_ids", trip.OrdiniIds).Error; err != nil {
		return nil, err
	}

	targetStato := string(models.OrderStatoPianificato)
	if trip.Stato == "IN_CORSO" {
		targetStato = string(models.OrderStatoViaggio)
	}
	if err := s.db.WithContext(ctx).Model(&models.Order{}).Where("id = ?", orderID).Updates(map[string]interface{}{
		"stato":           targetStato,
		"viaggio_id":      tripID,
		"targa_motrice":   trip.TargaMotrice,
		"targa_rimorchio": trip.TargaRimorchio,
		"autista_id":      trip.AutistaID,
		"updated_at":      time.Now().UTC(),
	}).Error; err != nil {
		return nil, err
	}

	segments := s.buildSegments(ctx, trip.GarageID, ordiniIds)
	km := totalKm(segments)
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("trip_id = ?", tripID).Delete(&models.TripSegment{}).Error; err != nil {
			return err
		}
		for i := range segments {
			segments[i].TripID = tripID
		}
		if len(segments) > 0 {
			if err := tx.Create(&segments).Error; err != nil {
				return err
			}
		}
		return tx.Model(&models.Trip{}).Where("id = ?", tripID).Update("km_totali", km).Error
	})
	if err != nil {
		return nil, err
	}

	return &dto.OKResult{OK: true}, nil
}

// ── Segment computation (ports backend/trip_segments.py) ──────────────────

func (s *TripService) buildSegments(ctx context.Context, garageID *uuid.UUID, ordiniIds []string) []models.TripSegment {
	segments := []models.TripSegment{}
	if len(ordiniIds) == 0 {
		return segments
	}
	garageIDStr := ""
	if garageID != nil {
		garageIDStr = garageID.String()
	}
	garage := s.geo.ResolveGarage(ctx, garageIDStr, "")

	var ordersList []models.Order
	s.db.WithContext(ctx).Preload("DestinazioneCarico").Preload("DestinazioneScarico").Where("id IN ?", ordiniIds).Find(&ordersList)
	sort.SliceStable(ordersList, func(i, j int) bool {
		if ordersList[i].DataRitiro != ordersList[j].DataRitiro {
			return ordersList[i].DataRitiro < ordersList[j].DataRitiro
		}
		return ordersList[i].CreatedAt.Before(ordersList[j].CreatedAt)
	})

	counter := 0
	var prevScarico *geo.NamedPoint

	for _, o := range ordersList {
		caricoNome, scaricoNome := "", ""
		var caricoLat, caricoLng, scaricoLat, scaricoLng *float64
		if o.DestinazioneCarico != nil {
			caricoNome, caricoLat, caricoLng = o.DestinazioneCarico.Nome, o.DestinazioneCarico.Lat, o.DestinazioneCarico.Lng
		}
		if o.DestinazioneScarico != nil {
			scaricoNome, scaricoLat, scaricoLng = o.DestinazioneScarico.Nome, o.DestinazioneScarico.Lat, o.DestinazioneScarico.Lng
		}
		carico, ok1 := geo.ResolveDestination(caricoNome, caricoLat, caricoLng)
		scarico, ok2 := geo.ResolveDestination(scaricoNome, scaricoLat, scaricoLng)
		if !ok1 || !ok2 {
			slog.Warn("trip_segments_skip_order", "id", o.ID.String(), "reason", "missing_coords")
			continue
		}
		oid := o.ID.String()
		if prevScarico == nil {
			segments = append(segments, s.segment(ctx, counter, "base_carico", garage, carico, &oid))
		} else {
			segments = append(segments, s.segment(ctx, counter, "scarico_carico_successivo", *prevScarico, carico, &oid))
		}
		counter++
		segments = append(segments, s.segment(ctx, counter, "carico_scarico", carico, scarico, &oid))
		counter++
		prevScarico = &scarico
	}

	if prevScarico != nil {
		segments = append(segments, s.segment(ctx, counter, "scarico_base", *prevScarico, garage, nil))
	}

	return segments
}

func (s *TripService) segment(ctx context.Context, ordine int, tipo string, origine, destinazione geo.NamedPoint, ordineID *string) models.TripSegment {
	route := s.geo.GetRoadRoute(ctx, origine.Lat, origine.Lng, destinazione.Lat, destinazione.Lng)
	km := 0.0
	minuti := 0
	if route != nil {
		km = route.DistanceKm
		minuti = int(math.Round(route.DurationHours * 60))
	}
	return models.TripSegment{
		ID:               uuid.New(),
		Ordine:           ordine,
		Tipo:             tipo,
		OrigineNome:      origine.Name,
		OrigineLat:       origine.Lat,
		OrigineLng:       origine.Lng,
		DestinazioneNome: destinazione.Name,
		DestinazioneLat:  destinazione.Lat,
		DestinazioneLng:  destinazione.Lng,
		Km:               km,
		TempoStimatoMin:  minuti,
		OrdineID:         ordineID,
	}
}

func totalKm(segments []models.TripSegment) float64 {
	total := 0.0
	for _, s := range segments {
		total += s.Km
	}
	return roundTo1(total)
}

func roundTo1(v float64) float64 {
	return float64(int(v*10+0.5)) / 10
}

// ── Helpers ──────────────────────────────────────────────────────────────

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

func uuidPtrString(id *uuid.UUID) string {
	if id == nil {
		return ""
	}
	return id.String()
}

func driverResponse(d *models.Driver) *dto.DriverResponse {
	if d == nil {
		return nil
	}
	return &dto.DriverResponse{
		ID: d.ID, Nome: d.Nome, Cognome: d.Cognome, CodiceFiscale: d.CodiceFiscale,
		Patente: d.Patente, ScadenzaPatente: d.ScadenzaPatente, Telefono: d.Telefono,
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

func garageResponse(g *models.Garage) *dto.GarageResponse {
	if g == nil {
		return nil
	}
	return &dto.GarageResponse{
		ID: g.ID, Nome: g.Nome, Indirizzo: g.Indirizzo, Citta: g.Citta, Lat: g.Lat, Lng: g.Lng,
		Note: g.Note, Active: g.Active, CreatedAt: g.CreatedAt, UpdatedAt: g.UpdatedAt,
	}
}

func toResponse(t models.Trip) dto.TripResponse {
	segmenti := make([]dto.TripSegmentDTO, len(t.Segments))
	for i, seg := range t.Segments {
		segmenti[i] = dto.TripSegmentDTO{
			Ordine: seg.Ordine, Tipo: seg.Tipo,
			OrigineNome: seg.OrigineNome, OrigineLat: seg.OrigineLat, OrigineLng: seg.OrigineLng,
			DestinazioneNome: seg.DestinazioneNome, DestinazioneLat: seg.DestinazioneLat, DestinazioneLng: seg.DestinazioneLng,
			Km: seg.Km, TempoStimatoMin: seg.TempoStimatoMin, OrdineID: seg.OrdineID,
		}
	}
	return dto.TripResponse{
		ID: t.ID, OrdiniIds: unmarshalStrings(t.OrdiniIds),
		TargaMotrice: t.TargaMotrice, TargaRimorchio: t.TargaRimorchio,
		AutistaID: uuidPtrString(t.AutistaID), Autista: driverResponse(t.Autista),
		VettoreID: uuidPtrString(t.VettoreID), Vettore: carrierResponse(t.Vettore),
		GarageID: uuidPtrString(t.GarageID), Garage: garageResponse(t.Garage),
		Segmenti: segmenti, KmTotali: t.KmTotali, CostoStimato: t.CostoStimato,
		Stato: t.Stato, Note: t.Note, DataPartenza: t.DataPartenza, DataArrivo: t.DataArrivo,
		CreatedAt: t.CreatedAt,
	}
}
