// Package mapview ports backend/routers/map.py: aggregates active orders,
// resolved coordinates and OSRM road routes for the live map view. Named
// "mapview" (not "map") to avoid shadowing the Go builtin.
package mapview

import (
	"context"
	"encoding/json"
	"sort"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	"fratelli-feccia/internal/dto"
	"fratelli-feccia/internal/models"
	"fratelli-feccia/internal/services/geo"
)

const activeOrdersLimit = 50

type MapService struct {
	db  *gorm.DB
	geo *geo.GeoService
}

func NewMapService(db *gorm.DB, orsApiKey, orsBaseURL string) *MapService {
	return &MapService{db: db, geo: geo.NewGeoService(db, orsApiKey, orsBaseURL)}
}

// Trips mirrors GET /map/trips.
func (s *MapService) Trips(ctx context.Context) (*dto.MapTripsResponse, error) {
	var activeOrders []models.Order
	if err := s.db.WithContext(ctx).
		Preload("Cliente").Preload("Autista").Preload("Motrice").Preload("DestinazioneCarico").Preload("DestinazioneScarico").
		Preload("Garage").Preload("WashStation").
		Where("stato IN ?", []string{"VIAGGIO", "PIANIFICABILE", "CHIUSO"}).
		Order("created_at DESC").Limit(activeOrdersLimit).Find(&activeOrders).Error; err != nil {
		return nil, err
	}

	destMap, err := s.buildDestinationMap(ctx)
	if err != nil {
		return nil, err
	}

	routes := make([]dto.MapRoute, 0, len(activeOrders))
	for _, o := range activeOrders {
		carico, scarico, ok := resolveOrderEndpoints(o, destMap)
		if !ok {
			continue
		}

		roadRoute := s.geo.GetRoadRoute(ctx, carico.Lat, carico.Lng, scarico.Lat, scarico.Lng)
		var roadPoints [][2]float64
		var totalKm, totalHours float64
		if roadRoute != nil {
			roadPoints = roadRoute.Points
			totalKm = roadRoute.DistanceKm
			totalHours = roadRoute.DurationHours
		} else {
			roadPoints = [][2]float64{{carico.Lat, carico.Lng}, {scarico.Lat, scarico.Lng}}
		}

		route := buildMapRoute(o, carico, scarico, roadPoints, totalKm, totalHours)
		routes = append(routes, route)
	}

	poi := make([]dto.MapPOI, 0, len(destMap))
	for id, p := range destMap {
		poi = append(poi, dto.MapPOI{ID: id, Nome: p.Name, Lat: p.Lat, Lng: p.Lng})
	}
	sort.Slice(poi, func(i, j int) bool { return poi[i].ID < poi[j].ID })

	var garageRows []models.Garage
	if err := s.db.WithContext(ctx).Where("active = ?", true).Find(&garageRows).Error; err != nil {
		return nil, err
	}
	garages := make([]dto.MapNamedPoint, len(garageRows))
	for i, g := range garageRows {
		garages[i] = dto.MapNamedPoint{Nome: g.Nome, Lat: *g.Lat, Lng: *g.Lng}
	}
	sort.Slice(garages, func(i, j int) bool { return garages[i].Nome < garages[j].Nome })

	var washStationRows []models.WashStation
	if err := s.db.WithContext(ctx).Where("active = ?", true).Find(&washStationRows).Error; err != nil {
		return nil, err
	}
	washStations := make([]dto.MapNamedPoint, len(washStationRows))
	for i, w := range washStationRows {
		washStations[i] = dto.MapNamedPoint{Nome: w.Nome, Lat: *w.Lat, Lng: *w.Lng}
	}
	sort.Slice(washStations, func(i, j int) bool { return washStations[i].Nome < washStations[j].Nome })

	stats := dto.MapStats{}
	for _, r := range routes {
		switch r.Stato {
		case "VIAGGIO":
			stats.InViaggio++
		case "PIANIFICABILE":
			stats.Pianificabili++
		case "CHIUSO":
			stats.Chiusi++
		}
	}

	return &dto.MapTripsResponse{Routes: routes, POI: poi, Garages: garages, WashStations: washStations, Stats: stats}, nil
}

// buildDestinationMap mirrors build_destination_map: resolves each active
// destination's coordinates, preferring its own Lat/Lng column over the
// hardcoded DestinationCoords fallback table.
func (s *MapService) buildDestinationMap(ctx context.Context) (map[string]geo.NamedPoint, error) {
	var destinations []models.Destination
	if err := s.db.WithContext(ctx).Where("active = ?", true).Limit(200).Find(&destinations).Error; err != nil {
		return nil, err
	}
	destMap := make(map[string]geo.NamedPoint, len(destinations))
	for _, d := range destinations {
		if p, ok := geo.ResolveDestination(d.Nome, d.Lat, d.Lng); ok {
			destMap[d.ID.String()] = p
		}
	}
	return destMap, nil
}

// resolveOrderEndpoints mirrors resolve_order_endpoints: destMap lookup by
// id first, falls back to name-based coordinate resolution.
func resolveOrderEndpoints(o models.Order, destMap map[string]geo.NamedPoint) (carico, scarico geo.NamedPoint, ok bool) {
	var found bool
	if o.DestinazioneCaricoID != nil {
		carico, found = destMap[o.DestinazioneCaricoID.String()]
	}
	if !found && o.DestinazioneCarico != nil {
		carico, found = geo.ResolveDestination(o.DestinazioneCarico.Nome, nil, nil)
	}
	if !found {
		return geo.NamedPoint{}, geo.NamedPoint{}, false
	}
	found = false
	if o.DestinazioneScaricoID != nil {
		scarico, found = destMap[o.DestinazioneScaricoID.String()]
	}
	if !found && o.DestinazioneScarico != nil {
		scarico, found = geo.ResolveDestination(o.DestinazioneScarico.Nome, nil, nil)
	}
	if !found {
		return geo.NamedPoint{}, geo.NamedPoint{}, false
	}
	return carico, scarico, true
}

func buildMapRoute(o models.Order, carico, scarico geo.NamedPoint, roadPoints [][2]float64, totalKm, totalHours float64) dto.MapRoute {
	var progress float64
	switch o.Stato {
	case "VIAGGIO":
		progress = 0.6
	case "CHIUSO":
		progress = 1.0
	default:
		progress = 0.0
	}
	curLat, curLng := geo.ComputeVehiclePosition(roadPoints, carico, scarico, progress)
	remainingKm := roundTo1(totalKm * (1 - progress))
	etaHours := roundTo1(totalHours * (1 - progress))

	mapRoadPoints := make([]dto.MapPoint, len(roadPoints))
	for i, p := range roadPoints {
		mapRoadPoints[i] = dto.MapPoint{Lat: p[0], Lng: p[1]}
	}

	return dto.MapRoute{
		ID: o.ID, Progressivo: o.Progressivo, Cliente: customerResponse(o.Cliente), Stato: string(o.Stato),
		Tipologia: o.Tipologia, Motrice: motriceResponse(o.Motrice), Autista: driverResponse(o.Autista),
		DataRitiro: o.DataRitiro, DataConsegna: o.DataConsegna, Tariffa: o.Tariffa,
		Carico: dto.MapPoint{Lat: carico.Lat, Lng: carico.Lng}, Scarico: dto.MapPoint{Lat: scarico.Lat, Lng: scarico.Lng},
		CurrentPosition: dto.MapPoint{Lat: curLat, Lng: curLng}, Progress: progress,
		RoadPoints: mapRoadPoints, DistanceKm: totalKm, DurationHours: totalHours,
		RemainingKm: remainingKm, EtaHours: etaHours,
		Garage: garageNamedPoint(o.Garage), WashStation: washStationNamedPoint(o.WashStation),
	}
}

func garageNamedPoint(g *models.Garage) *dto.MapNamedPoint {
	if g == nil || g.Lat == nil || g.Lng == nil {
		return nil
	}
	return &dto.MapNamedPoint{Nome: g.Nome, Lat: *g.Lat, Lng: *g.Lng}
}

func washStationNamedPoint(w *models.WashStation) *dto.MapNamedPoint {
	if w == nil || w.Lat == nil || w.Lng == nil {
		return nil
	}
	return &dto.MapNamedPoint{Nome: w.Nome, Lat: *w.Lat, Lng: *w.Lng}
}

func roundTo1(v float64) float64 {
	return float64(int(v*10+0.5)) / 10
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

func motriceResponse(m *models.Motrice) *dto.MotriceResponse {
	if m == nil {
		return nil
	}
	return &dto.MotriceResponse{
		ID: m.ID, Targa: m.Targa, Marca: m.Marca, Modello: m.Modello, Anno: m.Anno,
		PortataKg: m.PortataKg, Note: m.Note, Active: m.Active, CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt,
	}
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
