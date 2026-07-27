package orders

import (
	"context"
	"encoding/json"
	"fmt"
	"math"

	"github.com/google/uuid"

	"fratelli-feccia/internal/dto"
	"fratelli-feccia/internal/models"
	"fratelli-feccia/internal/services/geo"
	"fratelli-feccia/pkg/utils"
)

// routeWaypoint is a resolved (nome+coordinate) waypoint, the internal
// counterpart of dto.RouteWaypointDTO (which only ever carries tipo+ref_id
// over the wire — coordinates always come from the DB, never the client).
type routeWaypoint struct {
	Tipo  string
	RefID string
	Nome  string
	Lat   float64
	Lng   float64
}

func (w routeWaypoint) toResponseDTO() dto.RouteWaypointResponseDTO {
	return dto.RouteWaypointResponseDTO{Tipo: w.Tipo, RefID: w.RefID, Nome: w.Nome, Lat: w.Lat, Lng: w.Lng}
}

func waypointResponseDTOs(waypoints []routeWaypoint) []dto.RouteWaypointResponseDTO {
	out := make([]dto.RouteWaypointResponseDTO, len(waypoints))
	for i, w := range waypoints {
		out[i] = w.toResponseDTO()
	}
	return out
}

// resolveWaypoint looks up a waypoint's own coordinates by (tipo, ref_id) —
// never trusts client-supplied lat/lng, mirroring how every other order
// reference (garage, autista, ...) is resolved server-side.
func (s *OrderService) resolveWaypoint(ctx context.Context, tipo, refID string) (routeWaypoint, error) {
	id, err := uuid.Parse(refID)
	if err != nil {
		return routeWaypoint{}, utils.NewAPIError(400, fmt.Sprintf("ref_id waypoint non valido: %s", refID))
	}

	switch tipo {
	case "garage":
		var g models.Garage
		if err := s.db.WithContext(ctx).First(&g, "id = ?", id).Error; err != nil || g.Lat == nil || g.Lng == nil {
			return routeWaypoint{}, utils.NewAPIError(400, "Garage non trovato o senza coordinate")
		}
		return routeWaypoint{Tipo: tipo, RefID: refID, Nome: g.Nome, Lat: *g.Lat, Lng: *g.Lng}, nil
	case "destinazione":
		var d models.Destination
		if err := s.db.WithContext(ctx).First(&d, "id = ?", id).Error; err != nil || d.Lat == nil || d.Lng == nil {
			return routeWaypoint{}, utils.NewAPIError(400, "Destinazione non trovata o senza coordinate")
		}
		return routeWaypoint{Tipo: tipo, RefID: refID, Nome: d.Nome, Lat: *d.Lat, Lng: *d.Lng}, nil
	case "wash_station":
		var w models.WashStation
		if err := s.db.WithContext(ctx).First(&w, "id = ?", id).Error; err != nil || w.Lat == nil || w.Lng == nil {
			return routeWaypoint{}, utils.NewAPIError(400, "Punto di lavaggio non trovato o senza coordinate")
		}
		return routeWaypoint{Tipo: tipo, RefID: refID, Nome: w.Nome, Lat: *w.Lat, Lng: *w.Lng}, nil
	default:
		return routeWaypoint{}, utils.NewAPIError(400, fmt.Sprintf("Tipo waypoint sconosciuto: %s", tipo))
	}
}

// RouteAlternatives proposes up to 3 candidate routes for an order's
// carico→scarico leg (the only part with genuine route diversity — ORS's
// alternative_routes only works for a plain 2-coordinate request), each
// with the fixed garage→carico / scarico→wash_station legs prepended /
// appended when those points are given. Purely computed — nothing is
// written to the DB, the manager picks one before it's persisted (Assign)
// or a caller edits it further (UpdateRoute).
func (s *OrderService) RouteAlternatives(ctx context.Context, orderID uuid.UUID, garageID, washStationID string) ([]dto.RouteAlternativeDTO, error) {
	var order models.Order
	if err := s.db.WithContext(ctx).Preload("DestinazioneCarico").Preload("DestinazioneScarico").
		First(&order, "id = ?", orderID).Error; err != nil {
		return nil, err
	}
	carico := order.DestinazioneCarico
	scarico := order.DestinazioneScarico
	if carico == nil || carico.Lat == nil || carico.Lng == nil || scarico == nil || scarico.Lat == nil || scarico.Lng == nil {
		return nil, utils.NewAPIError(400, "L'ordine non ha coordinate di carico/scarico valide")
	}
	caricoWP := routeWaypoint{Tipo: "destinazione", RefID: carico.ID.String(), Nome: carico.Nome, Lat: *carico.Lat, Lng: *carico.Lng}
	scaricoWP := routeWaypoint{Tipo: "destinazione", RefID: scarico.ID.String(), Nome: scarico.Nome, Lat: *scarico.Lat, Lng: *scarico.Lng}

	var garageWP, washWP *routeWaypoint
	if garageID != "" {
		wp, err := s.resolveWaypoint(ctx, "garage", garageID)
		if err != nil {
			return nil, err
		}
		garageWP = &wp
	}
	if washStationID != "" {
		wp, err := s.resolveWaypoint(ctx, "wash_station", washStationID)
		if err != nil {
			return nil, err
		}
		washWP = &wp
	}

	coreAlts := s.geo.GetRoadRouteAlternatives(ctx, caricoWP.Lat, caricoWP.Lng, scaricoWP.Lat, scaricoWP.Lng, 3)

	var garageLeg, washLeg *geo.RouteResult
	if garageWP != nil {
		garageLeg = s.geo.GetRoadRoute(ctx, garageWP.Lat, garageWP.Lng, caricoWP.Lat, caricoWP.Lng)
	}
	if washWP != nil {
		washLeg = s.geo.GetRoadRoute(ctx, scaricoWP.Lat, scaricoWP.Lng, washWP.Lat, washWP.Lng)
	}

	result := make([]dto.RouteAlternativeDTO, 0, len(coreAlts))
	for _, core := range coreAlts {
		waypoints := make([]routeWaypoint, 0, 4)
		points := make([][2]float64, 0, len(core.Points)+40)
		distanceKm := 0.0
		durationHours := 0.0

		if garageWP != nil {
			waypoints = append(waypoints, *garageWP)
			if garageLeg != nil {
				points = append(points, garageLeg.Points...)
				distanceKm += garageLeg.DistanceKm
				durationHours += garageLeg.DurationHours
			}
		}
		waypoints = append(waypoints, caricoWP)
		points = append(points, core.Points...)
		distanceKm += core.DistanceKm
		durationHours += core.DurationHours
		waypoints = append(waypoints, scaricoWP)
		if washWP != nil {
			waypoints = append(waypoints, *washWP)
			if washLeg != nil {
				points = append(points, washLeg.Points...)
				distanceKm += washLeg.DistanceKm
				durationHours += washLeg.DurationHours
			}
		}

		result = append(result, dto.RouteAlternativeDTO{
			Waypoints:   waypointResponseDTOs(waypoints),
			Points:      points,
			DistanceKm:  distanceKm,
			DurationMin: int(math.Round(durationHours * 60)),
		})
	}
	return result, nil
}

// UpdateRoute recomputes and persists an order's route for an arbitrary
// manager-edited waypoint sequence (add/remove/reorder points from
// Destinations/Garage/WashStation) — always a single ORS multi-waypoint
// request, no alternatives (those only make sense before a route has been
// customized).
func (s *OrderService) UpdateRoute(ctx context.Context, orderID uuid.UUID, waypointsReq []dto.RouteWaypointDTO) (*dto.OrderResponse, error) {
	if len(waypointsReq) < 2 {
		return nil, utils.NewAPIError(400, "Servono almeno 2 waypoint (carico e scarico)")
	}

	var order models.Order
	if err := s.db.WithContext(ctx).First(&order, "id = ?", orderID).Error; err != nil {
		return nil, err
	}

	resolved := make([]routeWaypoint, len(waypointsReq))
	namedPoints := make([]geo.NamedPoint, len(waypointsReq))
	for i, w := range waypointsReq {
		wp, err := s.resolveWaypoint(ctx, w.Tipo, w.RefID)
		if err != nil {
			return nil, err
		}
		resolved[i] = wp
		namedPoints[i] = geo.NamedPoint{Name: wp.Nome, Lat: wp.Lat, Lng: wp.Lng}
	}

	route := s.geo.GetRoadRouteMultiWaypoint(ctx, namedPoints)
	if route == nil {
		return nil, utils.NewAPIError(502, "Impossibile calcolare il percorso (routing non disponibile)")
	}

	if err := s.upsertOrderRoute(ctx, &order, resolved, route, true); err != nil {
		return nil, err
	}

	reloaded, err := s.reload(ctx, order.ID)
	if err != nil {
		return nil, err
	}
	resp := ToResponse(*reloaded)
	return &resp, nil
}

// upsertOrderRoute creates or replaces order's OrderRoute row in place
// (an order has at most one active route — no history of past routes kept).
func (s *OrderService) upsertOrderRoute(ctx context.Context, order *models.Order, waypoints []routeWaypoint, route *geo.RouteResult, editedManually bool) error {
	waypointsJSON, err := json.Marshal(waypointResponseDTOs(waypoints))
	if err != nil {
		return err
	}
	pointsJSON, err := json.Marshal(route.Points)
	if err != nil {
		return err
	}

	orderRoute := models.OrderRoute{
		OrderID:        order.ID,
		Waypoints:      waypointsJSON,
		Points:         pointsJSON,
		DistanceKm:     route.DistanceKm,
		DurationMin:    int(math.Round(route.DurationHours * 60)),
		EditedManually: editedManually,
	}

	if order.RouteID != nil {
		orderRoute.ID = *order.RouteID
		return s.db.WithContext(ctx).Save(&orderRoute).Error
	}

	orderRoute.ID = uuid.New()
	if err := s.db.WithContext(ctx).Create(&orderRoute).Error; err != nil {
		return err
	}
	order.RouteID = &orderRoute.ID
	return s.db.WithContext(ctx).Model(&models.Order{}).Where("id = ?", order.ID).Update("route_id", orderRoute.ID).Error
}

// routeResponse converts a preloaded OrderRoute into its response DTO,
// unmarshaling the stored waypoints/points JSON columns.
func routeResponse(r *models.OrderRoute) *dto.RouteResponseDTO {
	if r == nil {
		return nil
	}
	var waypoints []dto.RouteWaypointResponseDTO
	_ = json.Unmarshal(r.Waypoints, &waypoints)
	var points [][2]float64
	_ = json.Unmarshal(r.Points, &points)
	return &dto.RouteResponseDTO{
		ID: r.ID, Waypoints: waypoints, Points: points,
		DistanceKm: r.DistanceKm, DurationMin: r.DurationMin, EditedManually: r.EditedManually,
	}
}
