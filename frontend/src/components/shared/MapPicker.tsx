import { useCallback, useEffect, useRef } from 'react';
import { MapContainer, TileLayer, Marker, useMap, useMapEvents } from 'react-leaflet';
import L from 'leaflet';

// Stesso fix icone di MapPage.js — necessario ovunque venga montata una
// MapContainer, altrimenti il marker di default risulta senza immagine
// (Leaflet risolve l'URL delle icone rispetto al bundle, non funziona con
// webpack senza questo override).
// @ts-expect-error — _getIconUrl esiste a runtime ma non è nei type di leaflet
delete L.Icon.Default.prototype._getIconUrl;
L.Icon.Default.mergeOptions({
  iconRetinaUrl: 'https://cdnjs.cloudflare.com/ajax/libs/leaflet/1.9.4/images/marker-icon-2x.png',
  iconUrl: 'https://cdnjs.cloudflare.com/ajax/libs/leaflet/1.9.4/images/marker-icon.png',
  shadowUrl: 'https://cdnjs.cloudflare.com/ajax/libs/leaflet/1.9.4/images/marker-shadow.png',
});

const DEFAULT_CENTER: [number, number] = [42.5, 12.5]; // centro Italia, usato finché non è impostato un punto
const DEFAULT_ZOOM = 5;
const POINT_ZOOM = 12;
const FLY_ZOOM = 15;

function ClickHandler({ onPick }: { onPick: (lat: number, lng: number) => void }) {
  useMapEvents({
    click(e) {
      onPick(e.latlng.lat, e.latlng.lng);
    },
  });
  return null;
}

// Vola verso il punto quando flyToSignal cambia (selezione da un indirizzo
// cercato altrove — vedi AddressSearchInput, usato per il campo Indirizzo
// nei form Destinazioni/Garage/Punti di Lavaggio) — non ad ogni click/drag
// sulla mappa stessa, quello sposta già il marker senza bisogno di volare.
function FlyToOnSignal({ lat, lng, flyToSignal }: { lat: number | null; lng: number | null; flyToSignal: number | null | undefined }) {
  const map = useMap();
  const mounted = useRef(false);
  useEffect(() => {
    if (!mounted.current) { mounted.current = true; return; }
    if (typeof lat === 'number' && typeof lng === 'number') map.flyTo([lat, lng], FLY_ZOOM);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [flyToSignal]);
  return null;
}

export interface MapPickerProps {
  lat: number | null;
  lng: number | null;
  onChange: (lat: number, lng: number) => void;
  flyToSignal?: number | null;
}

// Mappa cliccabile per impostare lat/lng di un punto (destinazione, garage,
// stazione di lavaggio...): click per posizionare il marker, drag per
// aggiustarlo. La ricerca per indirizzo vive nel campo Indirizzo del form
// (AddressSearchInput) — passare flyToSignal (un valore che cambia ad ogni
// selezione) per far volare la mappa lì. Non ricentra la mappa ad ogni click
// per non spostare la vista sotto il cursore dell'utente mentre esplora.
export function MapPicker({ lat, lng, onChange, flyToSignal }: MapPickerProps) {
  const hasPoint = typeof lat === 'number' && typeof lng === 'number' && !Number.isNaN(lat) && !Number.isNaN(lng);
  const center: [number, number] = hasPoint ? [lat as number, lng as number] : DEFAULT_CENTER;

  const handlePick = useCallback((newLat: number, newLng: number) => {
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
          {flyToSignal != null && <FlyToOnSignal lat={lat} lng={lng} flyToSignal={flyToSignal} />}
          {hasPoint && (
            <Marker
              position={[lat as number, lng as number]}
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
          ? `Punto selezionato: ${(lat as number).toFixed(5)}, ${(lng as number).toFixed(5)} — trascina il marker per correggere`
          : 'Cerca un indirizzo nel campo qui sopra o clicca sulla mappa per impostare il punto'}
      </p>
    </div>
  );
}
