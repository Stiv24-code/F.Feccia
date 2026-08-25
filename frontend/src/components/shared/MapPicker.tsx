import { useEffect, useRef } from 'react';
import { MapContainer, TileLayer, Marker, useMap } from 'react-leaflet';
import L from 'leaflet';
import { useAppSelector } from '@/store/hooks';

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
  flyToSignal?: number | null;
}

// Mappa di sola visualizzazione: il punto (destinazione, garage, stazione di
// lavaggio...) si imposta esclusivamente cercando un indirizzo nel campo
// Indirizzo del form (AddressSearchInput) — niente click né drag del marker,
// così il punto salvato corrisponde sempre esattamente al risultato del
// geocoding scelto. flyToSignal (un valore che cambia ad ogni selezione) fa
// volare la mappa lì.
export function MapPicker({ lat, lng, flyToSignal }: MapPickerProps) {
  // Stesso switch tile di OrderRouteMap — le tile chiare Esri stonano su
  // sfondo scuro, in dark mode si passa a CartoDB Dark Matter.
  const isDark = useAppSelector((s) => s.theme.theme === 'dark');
  const hasPoint = typeof lat === 'number' && typeof lng === 'number' && !Number.isNaN(lat) && !Number.isNaN(lng);
  const center: [number, number] = hasPoint ? [lat as number, lng as number] : DEFAULT_CENTER;

  return (
    <div className="space-y-1">
      <div className="rounded-md overflow-hidden border" style={{ height: 220 }}>
        <MapContainer
          center={center}
          zoom={hasPoint ? POINT_ZOOM : DEFAULT_ZOOM}
          style={{ height: '100%', width: '100%' }}
          scrollWheelZoom={false}
        >
          {isDark ? (
            <TileLayer
              url="https://{s}.basemaps.cartocdn.com/dark_all/{z}/{x}/{y}{r}.png"
              attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors &copy; <a href="https://carto.com/attributions">CARTO</a>'
              maxZoom={19}
            />
          ) : (
            <TileLayer
              url="https://server.arcgisonline.com/ArcGIS/rest/services/World_Street_Map/MapServer/tile/{z}/{y}/{x}"
              attribution="Tiles &copy; Esri"
            />
          )}
          {flyToSignal != null && <FlyToOnSignal lat={lat} lng={lng} flyToSignal={flyToSignal} />}
          {hasPoint && <Marker position={[lat as number, lng as number]} />}
        </MapContainer>
      </div>
      <p className="text-xs text-muted-foreground">
        {hasPoint
          ? `Punto selezionato: ${(lat as number).toFixed(5)}, ${(lng as number).toFixed(5)}`
          : 'Cerca un indirizzo nel campo qui sopra per impostare il punto'}
      </p>
    </div>
  );
}
