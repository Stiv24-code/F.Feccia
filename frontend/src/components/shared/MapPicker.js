import { useCallback, useEffect, useRef, useState } from 'react';
import { MapContainer, TileLayer, Marker, useMap, useMapEvents } from 'react-leaflet';
import L from 'leaflet';
import { Popover, PopoverTrigger, PopoverContent } from '@/components/ui/popover';
import { Input } from '@/components/ui/input';
import { Search, Loader2 } from 'lucide-react';
import { geocodeSearch } from '@/lib/api';
import { logger } from '@/lib/logger';

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
const SEARCH_ZOOM = 15;
const SEARCH_DEBOUNCE_MS = 400;
const MIN_QUERY_LENGTH = 3;

function ClickHandler({ onPick }) {
  useMapEvents({
    click(e) {
      onPick(e.latlng.lat, e.latlng.lng);
    },
  });
  return null;
}

// Sposta la vista sul punto trovato da ricerca indirizzo — solo un flyTo
// mirato quando cambia targetKey, non ricentra ad ogni render.
function FlyToOnSearch({ position, targetKey }) {
  const map = useMap();
  useEffect(() => {
    if (position) map.flyTo(position, SEARCH_ZOOM);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [targetKey]);
  return null;
}

// Mappa cliccabile per impostare lat/lng di un punto (destinazione, garage,
// stazione di lavaggio...): click per posizionare il marker, drag per
// aggiustarlo, oppure cerca un indirizzo per testo (geocoding via
// OpenRouteService, vedi backend GET /geocode/search) — la qualità del
// match dipende dalla copertura dati OSM per quell'indirizzo, non è mai
// garantito esatto, per questo si mostrano fino a 5 candidati da scegliere
// invece del solo primo risultato. Non ricentra la mappa ad ogni click per
// non spostare la vista sotto il cursore dell'utente mentre esplora.
export function MapPicker({ lat, lng, onChange, onAddressSelect }) {
  const hasPoint = typeof lat === 'number' && typeof lng === 'number' && !Number.isNaN(lat) && !Number.isNaN(lng);
  const center = hasPoint ? [lat, lng] : DEFAULT_CENTER;

  const [query, setQuery] = useState('');
  const [results, setResults] = useState([]);
  const [searching, setSearching] = useState(false);
  const [open, setOpen] = useState(false);
  const [flyTarget, setFlyTarget] = useState(null);
  const [flyKey, setFlyKey] = useState(0);
  const debounceRef = useRef(null);

  const handlePick = useCallback((newLat, newLng) => {
    onChange(Math.round(newLat * 1e6) / 1e6, Math.round(newLng * 1e6) / 1e6);
  }, [onChange]);

  useEffect(() => () => clearTimeout(debounceRef.current), []);

  const handleQueryChange = (value) => {
    setQuery(value);
    setOpen(true);
    clearTimeout(debounceRef.current);
    if (value.trim().length < MIN_QUERY_LENGTH) {
      setResults([]);
      setSearching(false);
      return;
    }
    setSearching(true);
    debounceRef.current = setTimeout(() => {
      geocodeSearch(value)
        .then(r => setResults(r.data || []))
        .catch(err => { logger.error('Errore ricerca indirizzo:', err); setResults([]); })
        .finally(() => setSearching(false));
    }, SEARCH_DEBOUNCE_MS);
  };

  const handleSelectResult = (result) => {
    handlePick(result.lat, result.lng);
    setFlyTarget([result.lat, result.lng]);
    setFlyKey(k => k + 1);
    setQuery(result.label);
    setOpen(false);
    if (onAddressSelect) onAddressSelect({ indirizzo: result.indirizzo, citta: result.citta });
  };

  return (
    <div className="space-y-1">
      <Popover open={open && (searching || results.length > 0)} onOpenChange={(o) => { if (!o) setOpen(false); }}>
        <PopoverTrigger asChild>
          <div className="relative">
            <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-muted-foreground pointer-events-none" />
            <Input
              value={query}
              onChange={(e) => handleQueryChange(e.target.value)}
              onFocus={() => setOpen(true)}
              placeholder="Cerca un indirizzo per posizionare il punto..."
              className="pl-8"
              data-testid="map-picker-search"
            />
            {searching && <Loader2 className="absolute right-2.5 top-1/2 -translate-y-1/2 h-3.5 w-3.5 animate-spin text-muted-foreground" />}
          </div>
        </PopoverTrigger>
        <PopoverContent
          className="w-[--radix-popover-trigger-width] p-1 z-[1000]"
          align="start"
          onOpenAutoFocus={(e) => e.preventDefault()}
        >
          {searching ? (
            <div className="flex items-center justify-center gap-2 py-3 text-xs text-muted-foreground">
              <Loader2 className="h-3 w-3 animate-spin" /> Ricerca in corso...
            </div>
          ) : results.length === 0 ? (
            <p className="py-3 text-center text-xs text-muted-foreground">Nessun risultato.</p>
          ) : (
            <div className="flex flex-col gap-0.5 max-h-56 overflow-auto">
              {results.map((r, i) => (
                <button
                  key={i}
                  type="button"
                  onClick={() => handleSelectResult(r)}
                  className="rounded-sm px-2 py-1.5 text-left text-sm hover:bg-muted/60"
                >
                  {r.label}
                </button>
              ))}
            </div>
          )}
        </PopoverContent>
      </Popover>

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
          <FlyToOnSearch position={flyTarget} targetKey={flyKey} />
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
          : 'Cerca un indirizzo o clicca sulla mappa per impostare il punto'}
      </p>
    </div>
  );
}
