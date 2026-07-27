import { useEffect } from 'react';
import { MapContainer, TileLayer, Marker, Tooltip, Polyline, useMap } from 'react-leaflet';
import L from 'leaflet';

// Fix icone Leaflet di default (stesso workaround di MapPage.js — necessario
// qui perché MapPage è una rotta lazy separata e potrebbe non essere mai
// stata caricata quando questo componente viene montato per la prima volta).
delete L.Icon.Default.prototype._getIconUrl;
L.Icon.Default.mergeOptions({
  iconRetinaUrl: 'https://cdnjs.cloudflare.com/ajax/libs/leaflet/1.9.4/images/marker-icon-2x.png',
  iconUrl: 'https://cdnjs.cloudflare.com/ajax/libs/leaflet/1.9.4/images/marker-icon.png',
  shadowUrl: 'https://cdnjs.cloudflare.com/ajax/libs/leaflet/1.9.4/images/marker-shadow.png',
});

const pointIcon = (color) => L.divIcon({
  className: '',
  html: `<div style="width:18px;height:18px;border-radius:50%;background:${color};border:3px solid white;box-shadow:0 2px 6px rgba(0,0,0,0.35)"></div>`,
  iconSize: [18, 18],
  iconAnchor: [9, 9],
});

const FitPoints = ({ points }) => {
  const map = useMap();
  useEffect(() => {
    if (points.length === 0) return;
    if (points.length === 1) { map.setView(points[0], 11); return; }
    map.fitBounds(points, { padding: [36, 36], maxZoom: 10 });
  }, [points, map]);
  return null;
};

const hasCoords = (p) => p?.lat != null && p?.lng != null;

// Mappa minimale: senza un percorso reale calcolato, la linea tratteggiata
// collega solo carico → scarico; garage e punto di lavaggio, quando
// assegnati, compaiono come marker aggiuntivi (inclusi nel fit dei bounds)
// ma non fanno parte della linea — non sappiamo in che ordine il mezzo li
// tocchi realmente. Quando è disponibile un percorso calcolato (routePoints,
// da OrderRoute/RouteAlternative — vera geometria stradale truck-aware via
// ORS) si disegna quello al suo posto, una polilinea piena invece che tratteggiata.
export default function OrderRouteMap({ carico, scarico, garage, washStation, routePoints: roadPoints, height = 220 }) {
  const routePoints = [];
  if (hasCoords(carico)) routePoints.push([carico.lat, carico.lng]);
  if (hasCoords(scarico)) routePoints.push([scarico.lat, scarico.lng]);

  const allPoints = roadPoints?.length ? [...roadPoints] : [...routePoints];
  if (hasCoords(garage)) allPoints.push([garage.lat, garage.lng]);
  if (hasCoords(washStation)) allPoints.push([washStation.lat, washStation.lng]);

  if (allPoints.length === 0) {
    return (
      <div className="flex items-center justify-center rounded-lg border bg-muted/30 text-xs text-muted-foreground" style={{ height }}>
        Coordinate non disponibili per la mappa
      </div>
    );
  }

  return (
    <div className="rounded-lg overflow-hidden border" style={{ height }} data-testid="order-route-map">
      <MapContainer center={allPoints[0]} zoom={6} style={{ height: '100%', width: '100%' }} zoomControl={true} scrollWheelZoom={true}>
        <TileLayer
          attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'
          url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
          maxZoom={19}
        />
        <FitPoints points={allPoints} />
        {roadPoints?.length > 1 ? (
          <Polyline positions={roadPoints} pathOptions={{ color: '#2a6fdb', weight: 4, opacity: 0.9 }} />
        ) : routePoints.length === 2 && (
          <Polyline positions={routePoints} pathOptions={{ color: '#2a6fdb', weight: 3, dashArray: '6 4', opacity: 0.85 }} />
        )}
        {hasCoords(carico) && (
          <Marker position={[carico.lat, carico.lng]} icon={pointIcon('#2a6fdb')}>
            <Tooltip direction="top" offset={[0, -10]} permanent>Carico · {carico.nome}</Tooltip>
          </Marker>
        )}
        {hasCoords(scarico) && (
          <Marker position={[scarico.lat, scarico.lng]} icon={pointIcon('#1f7a4d')}>
            <Tooltip direction="top" offset={[0, -10]} permanent>Scarico · {scarico.nome}</Tooltip>
          </Marker>
        )}
        {hasCoords(garage) && (
          <Marker position={[garage.lat, garage.lng]} icon={pointIcon('#0B1220')}>
            <Tooltip direction="top" offset={[0, -10]} permanent>Partenza · {garage.nome}</Tooltip>
          </Marker>
        )}
        {hasCoords(washStation) && (
          <Marker position={[washStation.lat, washStation.lng]} icon={pointIcon('#38BDF8')}>
            <Tooltip direction="top" offset={[0, -10]} permanent>Lavaggio · {washStation.nome}</Tooltip>
          </Marker>
        )}
      </MapContainer>
    </div>
  );
}
