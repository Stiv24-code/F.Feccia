// Package mapview ports backend/routers/map.py: aggregates active orders,
// resolved coordinates, OSRM road routes and live/simulated GPS positions
// for the live map view. Named "mapview" (not "map") to avoid shadowing the
// Go builtin.
package mapview

import (
	"context"
	"math"
	"sort"

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

func NewMapService(db *gorm.DB) *MapService {
	return &MapService{db: db, geo: geo.NewGeoService(db)}
}

// Trips mirrors GET /map/trips.
func (s *MapService) Trips(ctx context.Context) (*dto.MapTripsResponse, error) {
	var activeOrders []models.Order
	if err := s.db.WithContext(ctx).
		Where("stato IN ?", []string{"VIAGGIO", "PIANIFICABILE", "CHIUSO"}).
		Order("created_at DESC").Limit(activeOrdersLimit).Find(&activeOrders).Error; err != nil {
		return nil, err
	}

	destMap, err := s.buildDestinationMap(ctx)
	if err != nil {
		return nil, err
	}

	var vehicles []models.Vehicle
	if err := s.db.WithContext(ctx).Where("active = ? AND last_lat <> 0", true).Find(&vehicles).Error; err != nil {
		return nil, err
	}
	gpsVehicles := make(map[string]models.Vehicle, len(vehicles))
	for _, v := range vehicles {
		gpsVehicles[v.Targa] = v
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

		route := buildMapRoute(o, carico, scarico, roadPoints, totalKm, totalHours, gpsVehicles)
		routes = append(routes, route)
	}

	poi := make([]dto.MapPOI, 0, len(destMap))
	for id, p := range destMap {
		poi = append(poi, dto.MapPOI{ID: id, Nome: p.Name, Lat: p.Lat, Lng: p.Lng})
	}
	sort.Slice(poi, func(i, j int) bool { return poi[i].ID < poi[j].ID })

	garages := make([]dto.MapGarage, 0, len(geo.GarageCoords))
	for name, coords := range geo.GarageCoords {
		garages = append(garages, dto.MapGarage{Nome: name, Lat: coords[0], Lng: coords[1]})
	}
	sort.Slice(garages, func(i, j int) bool { return garages[i].Nome < garages[j].Nome })

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
		if r.GpsLive {
			stats.GpsLive++
		}
	}

	return &dto.MapTripsResponse{Routes: routes, POI: poi, Garages: garages, Stats: stats}, nil
}

// buildDestinationMap mirrors build_destination_map, minus the "persist
// lat/lng back to the destination doc" cache optimization — this port's
// Destination model has no lat/lng columns (see internal/models/destination.go),
// matching the fact that the CRUD API never exposes them either way; skipping
// it has no observable effect, just no cross-request memoization here.
func (s *MapService) buildDestinationMap(ctx context.Context) (map[string]geo.NamedPoint, error) {
	var destinations []models.Destination
	if err := s.db.WithContext(ctx).Where("active = ?", true).Limit(200).Find(&destinations).Error; err != nil {
		return nil, err
	}
	destMap := make(map[string]geo.NamedPoint, len(destinations))
	for _, d := range destinations {
		if p, ok := geo.ResolveDestination(d.Nome); ok {
			destMap[d.ID.String()] = p
		}
	}
	return destMap, nil
}

// resolveOrderEndpoints mirrors resolve_order_endpoints: destMap lookup by
// id first, falls back to name-based coordinate resolution.
func resolveOrderEndpoints(o models.Order, destMap map[string]geo.NamedPoint) (carico, scarico geo.NamedPoint, ok bool) {
	carico, found := destMap[o.DestinazioneCaricoID]
	if !found {
		carico, found = geo.ResolveDestination(o.DestinazioneCaricoNome)
	}
	if !found {
		return geo.NamedPoint{}, geo.NamedPoint{}, false
	}
	scarico, found = destMap[o.DestinazioneScaricoID]
	if !found {
		scarico, found = geo.ResolveDestination(o.DestinazioneScaricoNome)
	}
	if !found {
		return geo.NamedPoint{}, geo.NamedPoint{}, false
	}
	return carico, scarico, true
}

func buildMapRoute(o models.Order, carico, scarico geo.NamedPoint, roadPoints [][2]float64, totalKm, totalHours float64, gpsVehicles map[string]models.Vehicle) dto.MapRoute {
	targa := o.TargaMotrice
	gpsData, hasGps := gpsVehicles[targa]

	var gpsLive bool
	var gpsSpeed, gpsHeading, curLat, curLng, progress, remainingKm, etaHours float64
	var gpsTrackerUrl, gpsLastUpdate, gpsSource string
	var lastTempCelsius *float64
	var lastTempAlert bool

	if hasGps {
		gpsSource = gpsData.GpsSource
		lastTempCelsius = gpsData.LastTempCelsius
		lastTempAlert = gpsData.LastTempAlert
	}

	if hasGps && gpsData.LastLat != 0 && gpsData.GpsActive && o.Stato == "VIAGGIO" {
		curLat, curLng = gpsData.LastLat, gpsData.LastLng
		gpsLive = true
		gpsSpeed = gpsData.LastSpeedKmh
		gpsHeading = gpsData.LastHeading
		gpsTrackerUrl = gpsData.GpsTrackerUrl
		gpsLastUpdate = gpsData.LastGpsUpdate

		if len(roadPoints) > 1 {
			minDist := math.Inf(1)
			bestIdx := 0
			for idx, pt := range roadPoints {
				d := math.Hypot(pt[0]-curLat, pt[1]-curLng)
				if d < minDist {
					minDist = d
					bestIdx = idx
				}
			}
			denom := float64(len(roadPoints) - 1)
			if denom < 1 {
				denom = 1
			}
			progress = float64(bestIdx) / denom
			remainingKm = roundTo1(totalKm * (1 - progress))
			avgSpeed := gpsSpeed
			if avgSpeed <= 10 {
				avgSpeed = 70
			}
			if avgSpeed > 0 {
				etaHours = roundTo1(remainingKm / avgSpeed)
			}
		} else {
			progress = 0.5
		}
	} else {
		switch o.Stato {
		case "VIAGGIO":
			progress = 0.6
		case "CHIUSO":
			progress = 1.0
		default:
			progress = 0.0
		}
		curLat, curLng = geo.ComputeVehiclePosition(roadPoints, carico, scarico, progress)
		remainingKm = roundTo1(totalKm * (1 - progress))
		etaHours = roundTo1(totalHours * (1 - progress))
		if hasGps {
			gpsTrackerUrl = gpsData.GpsTrackerUrl
		}
	}

	mapRoadPoints := make([]dto.MapPoint, len(roadPoints))
	for i, p := range roadPoints {
		mapRoadPoints[i] = dto.MapPoint{Lat: p[0], Lng: p[1]}
	}

	garageCoords := geo.GarageCoords[geo.DefaultGarageName]

	return dto.MapRoute{
		ID: o.ID, Progressivo: o.Progressivo, ClienteNome: o.ClienteNome, Stato: o.Stato,
		Tipologia: o.Tipologia, TargaMotrice: targa, AutistaNome: o.AutistaNome,
		DataRitiro: o.DataRitiro, DataConsegna: o.DataConsegna, Tariffa: o.Tariffa,
		Carico: dto.MapPoint{Lat: carico.Lat, Lng: carico.Lng}, Scarico: dto.MapPoint{Lat: scarico.Lat, Lng: scarico.Lng},
		CurrentPosition: dto.MapPoint{Lat: curLat, Lng: curLng}, Progress: progress,
		RoadPoints: mapRoadPoints, DistanceKm: totalKm, DurationHours: totalHours,
		RemainingKm: remainingKm, EtaHours: etaHours,
		GpsLive: gpsLive, GpsSpeedKmh: gpsSpeed, GpsHeading: gpsHeading,
		GpsTrackerUrl: gpsTrackerUrl, GpsLastUpdate: gpsLastUpdate, GpsSource: gpsSource,
		LastTempCelsius: lastTempCelsius, LastTempAlert: lastTempAlert,
		Garage: dto.MapPoint{Lat: garageCoords[0], Lng: garageCoords[1]},
	}
}

func roundTo1(v float64) float64 {
	return float64(int(v*10+0.5)) / 10
}
