import { useCallback } from 'react';
import { MapContainer, TileLayer, Marker, useMapEvents } from 'react-leaflet';
import L from 'leaflet';

// Stesso fix icone di MapPage.js — necessario ovunque venga montata una
// MapContainer, altrimenti il marker di default risulta senza immagine
// (Leaflet risolve l'URL delle icone rispetto al bundle, non funziona con
// webpack senza questo override).
delete L.Icon.Default.prototype._getIconUrl;
L.Icon.Default.mergeOptions({
  iconRetinaUrl: 'https://cdnjs.cloudflare.com/ajax/libs/leaflet/1.9.4/images/marker-icon-2x.png',
  iconUrl: 'https://cdnjs.cloudflare.com/ajax/libs/leaflet/1.9.4/images/marker-icon.png',
  shadowUrl: 'https://cdnjs.cloudflare.com/ajax/libs/leaflet/1.9.4/images/marker-shadow.png',
});

const DEFAULT_CENTER = [42.5, 12.5]; // centro Italia, usato finché non è impostato un punto
const DEFAULT_ZOOM = 5;
const POINT_ZOOM = 12;

function ClickHandler({ onPick }) {
  useMapEvents({
    click(e) {
      onPick(e.latlng.lat, e.latlng.lng);
    },
  });
  return null;
}

// Mappa cliccabile per impostare lat/lng di un punto (destinazione, garage,
// stazione di lavaggio...): click per posizionare il marker, drag per
// aggiustarlo. Non ricentra la mappa ad ogni click per non spostare la vista
// sotto il cursore dell'utente mentre esplora.
export function MapPicker({ lat, lng, onChange }) {
  const hasPoint = typeof lat === 'number' && typeof lng === 'number' && !Number.isNaN(lat) && !Number.isNaN(lng);
  const center = hasPoint ? [lat, lng] : DEFAULT_CENTER;

  const handlePick = useCallback((newLat, newLng) => {
    onChange(Math.round(newLat * 1e6) / 1e6, Math.round(newLng * 1e6) / 1e6);
  }, [onChange]);

  return (
    <div className="space-y-1">
      <div className="rounded-md overflow-hidden border" style={{ height: 220 }}>
        <MapContainer
          center={center}
          zoom={hasPoint ? POINT_ZOOM : DEFAULT_ZOOM}
          style={{ height: '100%', width: '100%' }}
          scrollWheelZoom={false}
        >
          <TileLayer
            url="https://server.arcgisonline.com/ArcGIS/rest/services/World_Street_Map/MapServer/tile/{z}/{y}/{x}"
            attribution="Tiles &copy; Esri"
          />
          <ClickHandler onPick={handlePick} />
          {hasPoint && (
            <Marker
              position={[lat, lng]}
              draggable
              eventHandlers={{
                dragend: (e) => {
                  const { lat: newLat, lng: newLng } = e.target.getLatLng();
                  handlePick(newLat, newLng);
                },
              }}
            />
          )}
        </MapContainer>
      </div>
      <p className="text-xs text-muted-foreground">
        {hasPoint
          ? `Punto selezionato: ${lat.toFixed(5)}, ${lng.toFixed(5)} — trascina il marker per correggere`
          : 'Clicca sulla mappa per impostare il punto'}
      </p>
    </div>
  );
}
