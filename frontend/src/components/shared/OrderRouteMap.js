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

const FitTwoPoints = ({ points }) => {
  const map = useMap();
  useEffect(() => {
    if (points.length === 0) return;
    if (points.length === 1) { map.setView(points[0], 11); return; }
    map.fitBounds(points, { padding: [36, 36], maxZoom: 10 });
  }, [points, map]);
  return null;
};

// Mappa minimale con al più 2 punti (carico/scarico) e una linea diretta
// tra i due — nessun percorso stradale reale: non abbiamo ancora un
// routing/via point calcolato lato backend per il singolo ordine.
export default function OrderRouteMap({ carico, scarico, height = 220 }) {
  const points = [];
  if (carico?.lat != null && carico?.lng != null) points.push([carico.lat, carico.lng]);
  if (scarico?.lat != null && scarico?.lng != null) points.push([scarico.lat, scarico.lng]);

  if (points.length === 0) {
    return (
      <div className="flex items-center justify-center rounded-lg border bg-muted/30 text-xs text-muted-foreground" style={{ height }}>
        Coordinate non disponibili per la mappa
      </div>
    );
  }

  return (
    <div className="rounded-lg overflow-hidden border" style={{ height }} data-testid="order-route-map">
      <MapContainer center={points[0]} zoom={6} style={{ height: '100%', width: '100%' }} zoomControl={false} scrollWheelZoom={false}>
        <TileLayer
          attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'
          url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
          maxZoom={19}
        />
        <FitTwoPoints points={points} />
        {points.length === 2 && (
          <Polyline positions={points} pathOptions={{ color: '#2a6fdb', weight: 3, dashArray: '6 4', opacity: 0.85 }} />
        )}
        {carico?.lat != null && carico?.lng != null && (
          <Marker position={[carico.lat, carico.lng]} icon={pointIcon('#2a6fdb')}>
            <Tooltip direction="top" offset={[0, -10]} permanent>Carico · {carico.nome}</Tooltip>
          </Marker>
        )}
        {scarico?.lat != null && scarico?.lng != null && (
          <Marker position={[scarico.lat, scarico.lng]} icon={pointIcon('#1f7a4d')}>
            <Tooltip direction="top" offset={[0, -10]} permanent>Scarico · {scarico.nome}</Tooltip>
          </Marker>
        )}
      </MapContainer>
    </div>
  );
}
