import { useState, useEffect, useCallback, useMemo } from 'react';
import { getMapTrips } from '@/lib/api';
import { formatEuro } from '@/lib/format';
import type { DtoMapTripsResponse, DtoMapRoute, DtoMapNamedPoint } from '@/api/data-contracts';
import { Card } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Skeleton } from '@/components/ui/skeleton';
import { StatusBadge } from '@/components/shared/StatusBadge';
import { MapContainer, TileLayer, LayersControl, Marker, Popup, Polyline, CircleMarker, Tooltip, useMap } from 'react-leaflet';
import L from 'leaflet';
import { Truck, RefreshCw, Search, MapPin } from 'lucide-react';
import { logger } from '@/lib/logger';
import { useAppSelector } from '@/store/hooks';

// Fix icone Leaflet
// @ts-expect-error — _getIconUrl esiste a runtime ma non è nei type di leaflet
delete L.Icon.Default.prototype._getIconUrl;
L.Icon.Default.mergeOptions({
  iconRetinaUrl: 'https://cdnjs.cloudflare.com/ajax/libs/leaflet/1.9.4/images/marker-icon-2x.png',
  iconUrl: 'https://cdnjs.cloudflare.com/ajax/libs/leaflet/1.9.4/images/marker-icon.png',
  shadowUrl: 'https://cdnjs.cloudflare.com/ajax/libs/leaflet/1.9.4/images/marker-shadow.png',
});

// Icone custom
const createIcon = (color: string, size = 28) => L.divIcon({
  className: '',
  html: `<div style="width:${size}px;height:${size}px;border-radius:50%;background:${color};border:3px solid white;box-shadow:0 2px 8px rgba(0,0,0,0.3);display:flex;align-items:center;justify-content:center;">
    <svg width="${size * 0.5}" height="${size * 0.5}" viewBox="0 0 24 24" fill="none" stroke="white" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><rect x="1" y="3" width="15" height="13"/><polygon points="16 8 20 8 23 11 23 16 16 16 16 8"/><circle cx="5.5" cy="18.5" r="2.5"/><circle cx="18.5" cy="18.5" r="2.5"/></svg>
  </div>`,
  iconSize: [size, size],
  iconAnchor: [size / 2, size / 2],
  popupAnchor: [0, -size / 2],
});

const garageIcon = L.divIcon({
  className: '',
  html: `<div style="width:32px;height:32px;border-radius:8px;background:#0B1220;border:3px solid #2A6FDB;box-shadow:0 2px 12px rgba(42,111,219,0.4);display:flex;align-items:center;justify-content:center;">
    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="#2A6FDB" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"/><polyline points="9 22 9 12 15 12 15 22"/></svg>
  </div>`,
  iconSize: [32, 32],
  iconAnchor: [16, 16],
  popupAnchor: [0, -16],
});

const washIcon = L.divIcon({
  className: '',
  html: `<div style="width:28px;height:28px;border-radius:8px;background:#0B1220;border:3px solid #38BDF8;box-shadow:0 2px 12px rgba(56,189,248,0.4);display:flex;align-items:center;justify-content:center;">
    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="#38BDF8" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><path d="M12 22a7 7 0 0 0 7-7c0-2-1-3.9-3-5.5s-3.5-4-4-6.5c-.5 2.5-2 4-4 5.5C6 9.1 5 11 5 15a7 7 0 0 0 7 7z"/></svg>
  </div>`,
  iconSize: [28, 28],
  iconAnchor: [14, 14],
  popupAnchor: [0, -14],
});

const destIcon = L.divIcon({
  className: '',
  html: `<div style="width:10px;height:10px;border-radius:50%;background:#2A6FDB;border:2px solid white;box-shadow:0 1px 4px rgba(0,0,0,0.2);"></div>`,
  iconSize: [10, 10],
  iconAnchor: [5, 5],
});

const statusColors: Record<string, string> = {
  VIAGGIO: '#E24A4A',
  PIANIFICABILE: '#F0B429',
  CHIUSO: '#F28B2C',
};

// Stili costanti per pathOptions (evitano re-render)
const CIRCLE_MARKER_BASE = { color: 'white', fillOpacity: 1, weight: 2 };
const SHADOW_PATH_OPTIONS = { color: '#000', weight: 8, opacity: 0.15, lineCap: 'round' as const, lineJoin: 'round' as const };
const MAP_CONTAINER_STYLE = { height: '100%', width: '100%' };
const POPUP_STYLES = {
  container: { minWidth: 220 },
  header: { display: 'flex', alignItems: 'center', gap: 6, marginBottom: 4 },
  title: { margin: 0, fontWeight: 700, fontSize: 14, flex: 1 },
  subtitle: { margin: 0, fontSize: 12, color: '#666' },
  hr: { margin: '6px 0', border: 'none', borderTop: '1px solid #eee' },
  route: { margin: '2px 0', fontSize: 12 },
  detail: { margin: '2px 0', fontSize: 11, color: '#888' },
  progressBg: { marginTop: 6, background: '#f0f0f0', borderRadius: 6, height: 6, overflow: 'hidden' },
  progressLabel: { margin: '4px 0 0', fontSize: 10, color: '#999' },
};

// Componente per auto-fit bounds
const FitBounds = ({ routes, garages, washStations }: { routes: DtoMapRoute[]; garages: DtoMapNamedPoint[]; washStations: DtoMapNamedPoint[] }) => {
  const map = useMap();
  useEffect(() => {
    if (routes.length === 0 && garages.length === 0 && washStations.length === 0) return;
    const allPoints: [number, number][] = [];
    routes.forEach(r => {
      if (r.carico?.lat != null && r.carico?.lng != null) allPoints.push([r.carico.lat, r.carico.lng]);
      if (r.scarico?.lat != null && r.scarico?.lng != null) allPoints.push([r.scarico.lat, r.scarico.lng]);
    });
    garages.forEach(g => { if (g.lat != null && g.lng != null) allPoints.push([g.lat, g.lng]); });
    washStations.forEach(w => { if (w.lat != null && w.lng != null) allPoints.push([w.lat, w.lng]); });
    if (allPoints.length > 0) {
      map.fitBounds(allPoints, { padding: [40, 40], maxZoom: 7 });
    }
  }, [routes, garages, washStations, map]);
  return null;
};

export default function MapPage() {
  // Stesso switch tile di OrderRouteMap/MapPicker — default scuro quando il
  // tema è dark, le altre basemap Esri restano comunque selezionabili a mano.
  const isDark = useAppSelector((s) => s.theme.theme === 'dark');
  const [data, setData] = useState<DtoMapTripsResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [selectedRoute, setSelectedRoute] = useState<DtoMapRoute | null>(null);
  // I tre stati sono ora accendibili/spegnibili tutti e tre: prima VIAGGIO
  // era forzato sempre visibile e non c'era una pill per isolarlo.
  const [showViaggio, setShowViaggio] = useState(true);
  const [showPianificabili, setShowPianificabili] = useState(false);
  const [showChiusi, setShowChiusi] = useState(false);
  const [showPoi, setShowPoi] = useState(true);
  const [filterVeicolo, setFilterVeicolo] = useState('');
  const [panelQuery, setPanelQuery] = useState('');

  const fetchData = useCallback(() => {
    setLoading(true);
    getMapTrips().then((r: { data: DtoMapTripsResponse }) => setData(r.data)).catch((err: unknown) => logger.error('Errore caricamento mappa:', err)).finally(() => setLoading(false));
  }, []);
  useEffect(() => { fetchData(); }, [fetchData]);

  const filteredRoutes = useMemo(() => {
    if (!data) return [];
    return (data.routes || []).filter(r => {
      if (!showViaggio && r.stato === 'VIAGGIO') return false;
      if (!showPianificabili && r.stato === 'PIANIFICABILE') return false;
      if (!showChiusi && r.stato === 'CHIUSO') return false;
      if (filterVeicolo && r.motrice?.targa !== filterVeicolo) return false;
      if (!r.carico || !r.scarico) return false;
      if (!r.carico.lat || !r.carico.lng || !r.scarico.lat || !r.scarico.lng) return false;
      if (isNaN(r.carico.lat) || isNaN(r.scarico.lat)) return false;
      return true;
    });
  }, [data, showViaggio, showPianificabili, showChiusi, filterVeicolo]);

  // Pill di stato della toolbar: etichetta, conteggio (dalle stats del
  // backend, come i badge di prima) e interruttore, in un solo elemento.
  const allShown = showViaggio && showPianificabili && showChiusi;
  const showAll = () => { setShowViaggio(true); setShowPianificabili(true); setShowChiusi(true); };
  const STATUS_CHIPS = [
    { key: 'viaggio', label: 'In viaggio', className: 'status-order-blue', count: data?.stats?.in_viaggio || 0, on: showViaggio, toggle: () => setShowViaggio(v => !v) },
    { key: 'pianificabili', label: 'Da pianificare', className: 'status-order-yellow', count: data?.stats?.pianificabili || 0, on: showPianificabili, toggle: () => setShowPianificabili(v => !v) },
    { key: 'chiusi', label: 'Chiusi', className: 'status-chiuso', count: data?.stats?.chiusi || 0, on: showChiusi, toggle: () => setShowChiusi(v => !v) },
  ];

  const uniqueVehicles = useMemo(() => data ? Array.from(new Set((data.routes || []).map(r => r.motrice?.targa).filter(Boolean))) : [], [data]);

  // Lista del pannello laterale: tutti i viaggi tracciati sulla mappa, non
  // solo quelli in viaggio. Prima era fissa su `inViaggio` e quindi ignorava
  // i filtri della toolbar: attivando "Pianificabili" i percorsi comparivano
  // sulla mappa ma il pannello restava vuoto.
  const panelRoutes = useMemo(() => {
    const q = panelQuery.trim().toLowerCase();
    if (!q) return filteredRoutes;
    return filteredRoutes.filter(r => {
      const hay = `${r.progressivo || ''} ${r.cliente?.ragione_sociale || ''} ${r.autista?.nome || ''} ${r.autista?.cognome || ''} ${r.motrice?.targa || ''} ${r.carico?.nome || ''} ${r.scarico?.nome || ''}`;
      return hay.toLowerCase().includes(q);
    });
  }, [filteredRoutes, panelQuery]);

  if (loading) {
    return (
      <div className="space-y-4">
        <Skeleton className="h-12 rounded-xl" />
        <Skeleton className="h-[calc(100vh-180px)] rounded-xl" />
      </div>
    );
  }

  if (!data) return <p className="text-muted-foreground text-center py-12">Impossibile caricare i dati mappa.</p>;

  return (
    /* Impianto del mockup (ui/tms-unificato.html, blocco MAPPA LIVE): due
       colonne 1.9fr / 1fr, a sinistra la riga di filtri sopra la mappa che
       riempie l'altezza, a destra il pannello viaggi a tutta altezza.
       In tema Glass il mockup manda la sola mappa a fondo di finestra e lascia
       il resto dov'è — quel pezzo è in index.css (".glass [data-map-canvas]"):
       la griglia qui è una, e filtri e pannello non si spostano cambiando
       tema. */
    <div
      data-map-shell
      data-testid="map-page"
      className="grid grid-cols-1 gap-3.5 lg:grid-cols-[1.9fr_1fr] lg:h-[calc(100vh-5.5rem)] lg:min-h-[520px]"
    >
      {/* Colonna sinistra: filtri (altezza propria) + mappa (il resto).
          In Glass questa colonna si dissolve (`display:contents`) e i filtri
          diventano fratelli del pannello — vedi index.css. */}
      <div data-map-col className="flex flex-col gap-2.5 min-w-0 lg:min-h-0">
        <div data-map-toolbar className="shrink-0 flex flex-col gap-2 lg:flex-row lg:items-center">
          {/* Pill contatore: un solo elemento per stato invece di un badge
              col numero più un bottone occhio per accenderlo/spegnerlo. */}
          <div className="flex items-center gap-1.5 flex-wrap">
            <button
              type="button"
              onClick={showAll}
              aria-pressed={allShown}
              data-testid="map-chip-all"
              className={`text-xs font-semibold rounded-full border px-3 py-1 transition border-muted-foreground/30 bg-background text-foreground ${
                allShown ? 'ring-2 ring-offset-1 ring-primary' : 'opacity-70 hover:opacity-100'
              }`}
            >
              Tutti
            </button>
            {STATUS_CHIPS.map(c => (
              <button
                key={c.key}
                type="button"
                onClick={c.toggle}
                aria-pressed={c.on}
                data-testid={`map-chip-${c.key}`}
                className={`text-xs font-semibold rounded-full border px-3 py-1 transition ${c.className} ${
                  c.on ? 'ring-2 ring-offset-1 ring-primary' : 'opacity-70 hover:opacity-100'
                }`}
              >
                {c.label} · <span className="tabular-nums">{c.count}</span>
              </button>
            ))}
          </div>

          <div className="flex items-center gap-2 flex-wrap lg:ml-auto">
            {uniqueVehicles.length > 0 && (
              <select
                className="h-8 px-2 text-xs border rounded-md bg-card"
                value={filterVeicolo}
                onChange={e => setFilterVeicolo(e.target.value)}
                data-testid="map-filter-vehicle"
              >
                <option value="">Tutti i mezzi</option>
                {uniqueVehicles.map(v => <option key={v} value={v}>{v}</option>)}
              </select>
            )}
            <Button variant="outline" size="sm" className="text-xs gap-1.5 h-8" onClick={fetchData}>
              <RefreshCw className="h-3.5 w-3.5" /> Aggiorna
            </Button>
            {/* POI: garage, punti di lavaggio e punti d'interesse sono i
                marker "fissi" della mappa, distinti dai tracciati dei viaggi.
                Prima erano sempre accesi e non c'era modo di pulire la vista. */}
            <button
              type="button"
              onClick={() => setShowPoi(p => !p)}
              aria-pressed={showPoi}
              data-testid="map-toggle-poi"
              className="flex items-center gap-2 text-xs font-semibold text-muted-foreground hover:text-foreground transition-colors"
            >
              <MapPin className="h-3.5 w-3.5" /> POI
              <span className={`relative h-4 w-7 rounded-full transition-colors ${showPoi ? 'bg-primary' : 'bg-muted-foreground/30'}`}>
                <span className={`absolute top-0.5 h-3 w-3 rounded-full bg-white shadow transition-all ${showPoi ? 'left-3.5' : 'left-0.5'}`} />
              </span>
            </button>
          </div>
        </div>

        {/* La mappa riempie l'altezza della colonna. In tema Glass
            `data-map-canvas` la porta a fondo di finestra (vedi index.css). */}
        <Card
          data-map-canvas
          className="h-[60vh] lg:h-auto lg:flex-1 lg:min-h-0 rounded-xl overflow-hidden shadow-sm relative"
          data-testid="map-container"
        >
          <MapContainer
            center={[47.0, 9.0]}
            zoom={5}
            style={MAP_CONTAINER_STYLE}
            zoomControl={true}
          >
            {/*
              Basemap: OpenStreetMap standard come default (stessa mappa
              stradale reale usata dal prototipo shortestPath), non lo stile
              "atlante politico" di Esri National Geographic. Le altre 3
              basemap Esri restano disponibili come alternative via il
              selector in alto a destra.

              Attribution conforme ai ToS di Esri (richiamo del servizio
              specifico) e di OpenStreetMap.
            */}
            <LayersControl position="topright">
              <LayersControl.BaseLayer checked={!isDark} name="OpenStreetMap">
                <TileLayer
                  attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'
                  url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
                  maxZoom={19}
                />
              </LayersControl.BaseLayer>
              <LayersControl.BaseLayer checked={isDark} name="CartoDB — Dark Matter">
                <TileLayer
                  attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors &copy; <a href="https://carto.com/attributions">CARTO</a>'
                  url="https://{s}.basemaps.cartocdn.com/dark_all/{z}/{x}/{y}{r}.png"
                  maxZoom={19}
                />
              </LayersControl.BaseLayer>
              <LayersControl.BaseLayer name="Esri — National Geographic">
                <TileLayer
                  attribution='Tiles &copy; Esri &mdash; National Geographic, DeLorme, NAVTEQ, UNEP-WCMC, USGS, NASA, ESA, METI, NRCAN, GEBCO, NOAA, iPC'
                  url="https://server.arcgisonline.com/ArcGIS/rest/services/NatGeo_World_Map/MapServer/tile/{z}/{y}/{x}"
                  maxZoom={16}
                />
              </LayersControl.BaseLayer>
              <LayersControl.BaseLayer name="Esri — World Topographic">
                <TileLayer
                  attribution='Tiles &copy; Esri &mdash; Esri, DeLorme, NAVTEQ, TomTom, Intermap, iPC, USGS, FAO, NPS, NRCAN, GeoBase, Kadaster NL, Ordnance Survey, Esri Japan, METI, Esri China (Hong Kong), and the GIS User Community'
                  url="https://server.arcgisonline.com/ArcGIS/rest/services/World_Topo_Map/MapServer/tile/{z}/{y}/{x}"
                  maxZoom={19}
                />
              </LayersControl.BaseLayer>
              <LayersControl.BaseLayer name="Esri — World Street Map">
                <TileLayer
                  attribution='Tiles &copy; Esri &mdash; Sources: Esri, DeLorme, HERE, USGS, Intermap, iPC, NRCAN, Esri Japan, METI, Esri China (Hong Kong), NOSTRA, MapmyIndia, © OpenStreetMap contributors, and the GIS user community'
                  url="https://server.arcgisonline.com/ArcGIS/rest/services/World_Street_Map/MapServer/tile/{z}/{y}/{x}"
                  maxZoom={19}
                />
              </LayersControl.BaseLayer>
              <LayersControl.BaseLayer name="Esri — World Imagery (satellite)">
                <TileLayer
                  attribution='Tiles &copy; Esri &mdash; Source: Esri, i-cubed, USDA, USGS, AEX, GeoEye, Getmapping, Aerogrid, IGN, IGP, UPR-EGP, and the GIS User Community'
                  url="https://server.arcgisonline.com/ArcGIS/rest/services/World_Imagery/MapServer/tile/{z}/{y}/{x}"
                  maxZoom={19}
                />
              </LayersControl.BaseLayer>
            </LayersControl>
            <FitBounds routes={filteredRoutes} garages={data.garages || []} washStations={data.wash_stations || []} />

            {/* Garage, lavaggi e destinazioni sono i "POI" del prototipo
                (route-map.html: Carico/Scarico, Officina, Lavaggio): punti
                fissi, distinti dai tracciati dei viaggi, spenti insieme
                dall'interruttore POI in toolbar. */}
            {showPoi && (data.garages || []).map((g, i) => (
              g.lat != null && g.lng != null && (
                <Marker key={`g-${i}`} position={[g.lat, g.lng]} icon={garageIcon}>
                  <Popup><strong>{g.nome}</strong><br />Base operativa</Popup>
                  <Tooltip direction="top" offset={[0, -20]} permanent={false}>{g.nome}</Tooltip>
                </Marker>
              )
            ))}

            {showPoi && (data.wash_stations || []).map((w, i) => (
              w.lat != null && w.lng != null && (
                <Marker key={`wash-${i}`} position={[w.lat, w.lng]} icon={washIcon}>
                  <Popup><strong>{w.nome}</strong><br />Punto di lavaggio</Popup>
                  <Tooltip direction="top" offset={[0, -16]} permanent={false}>{w.nome}</Tooltip>
                </Marker>
              )
            ))}

            {showPoi && (data.poi || []).map((p, i) => (
              p.lat != null && p.lng != null && (
                <Marker key={`poi-${i}`} position={[p.lat, p.lng]} icon={destIcon}>
                  <Tooltip direction="top" offset={[0, -8]}>{p.nome}</Tooltip>
                </Marker>
              )
            ))}

            {/* Percorsi stradali reali */}
            {filteredRoutes.map((route) => {
              if (!route.carico || !route.scarico) return null;
              if (!route.carico.lat || !route.scarico.lat) return null;
              if (isNaN(route.carico.lat) || isNaN(route.scarico.lat)) return null;

              const carico = route.carico;
              const scarico = route.scarico;

              // Usa percorso stradale dal backend, fallback a linea diretta
              const roadPts: [number, number][] = route.road_points && route.road_points.length > 1
                ? route.road_points.map(p => [p.lat || 0, p.lng || 0])
                : [[carico.lat as number, carico.lng as number], [scarico.lat as number, scarico.lng as number]];

              const color = (route.stato && statusColors[route.stato]) || '#888';
              const isSelected = selectedRoute?.id === route.id;
              const isActive = route.stato === 'VIAGGIO';

              return (
                <div key={route.id}>
                  {/* Ombra percorso (bordo) */}
                  {isSelected && (
                    <Polyline
                      positions={roadPts}
                      pathOptions={SHADOW_PATH_OPTIONS}
                    />
                  )}

                  {/* Percorso stradale reale */}
                  <Polyline
                    positions={roadPts}
                    pathOptions={{
                      color: color,
                      weight: isSelected ? 5 : 3,
                      opacity: isSelected ? 0.95 : (isActive ? 0.75 : 0.4),
                      dashArray: isActive ? undefined : '6 4',
                      lineCap: 'round',
                      lineJoin: 'round',
                    }}
                    eventHandlers={{ click: () => setSelectedRoute(route) }}
                  />

                  {/* Marker carico */}
                  <CircleMarker
                    center={[carico.lat as number, carico.lng as number]}
                    radius={isSelected ? 7 : 5}
                    pathOptions={{ ...CIRCLE_MARKER_BASE, fillColor: color }}
                  >
                    <Tooltip direction="top">
                      <strong>Carico</strong><br />
                      {route.progressivo} — {route.cliente?.ragione_sociale}
                    </Tooltip>
                  </CircleMarker>

                  {/* Marker scarico */}
                  <CircleMarker
                    center={[scarico.lat as number, scarico.lng as number]}
                    radius={isSelected ? 7 : 5}
                    pathOptions={{ ...CIRCLE_MARKER_BASE, fillColor: color }}
                  >
                    <Tooltip direction="top">
                      <strong>Scarico</strong>
                    </Tooltip>
                  </CircleMarker>

                  {/* Punto di partenza / lavaggio assegnati all'ordine — mostrati solo
                      per la rotta selezionata, per non affollare la vista di default
                      (i garage/lavaggi attivi sono già tutti visibili come marker globali). */}
                  {isSelected && route.garage?.lat != null && route.garage?.lng != null && (
                    <Marker position={[route.garage.lat, route.garage.lng]} icon={garageIcon}>
                      <Popup><strong>{route.garage.nome}</strong><br />Punto di partenza — {route.progressivo}</Popup>
                      <Tooltip direction="top" offset={[0, -20]}>Partenza: {route.garage.nome}</Tooltip>
                    </Marker>
                  )}
                  {isSelected && route.wash_station?.lat != null && route.wash_station?.lng != null && (
                    <Marker position={[route.wash_station.lat, route.wash_station.lng]} icon={washIcon}>
                      <Popup><strong>{route.wash_station.nome}</strong><br />Punto di lavaggio (prima del carico) — {route.progressivo}</Popup>
                      <Tooltip direction="top" offset={[0, -16]}>Lavaggio: {route.wash_station.nome}</Tooltip>
                    </Marker>
                  )}

                  {/* Posizione attuale del mezzo (solo per VIAGGIO) */}
                  {isActive && route.current_position?.lat != null && route.current_position?.lng != null && (
                    <Marker
                      position={[route.current_position.lat, route.current_position.lng]}
                      icon={createIcon(color, isSelected ? 36 : 30)}
                      eventHandlers={{ click: () => setSelectedRoute(route) }}
                    >
                      <Popup className="custom-popup">
                        <div style={POPUP_STYLES.container}>
                          <div style={POPUP_STYLES.header}>
                            <p style={POPUP_STYLES.title}>{route.motrice?.targa || 'Mezzo'}</p>
                          </div>
                          <p style={POPUP_STYLES.subtitle}>{route.autista ? `${route.autista.nome} ${route.autista.cognome}` : 'N/A'}</p>
                          <hr style={POPUP_STYLES.hr} />
                          <p style={POPUP_STYLES.detail}>{route.cliente?.ragione_sociale} • {route.progressivo}</p>
                          {(route.distance_km || 0) > 0 && <p style={POPUP_STYLES.detail}>{route.distance_km} km totali • {route.duration_hours}h stimate</p>}
                          <div style={POPUP_STYLES.progressBg}>
                            <div style={{ width: `${(route.progress || 0) * 100}%`, height: '100%', background: color, borderRadius: 6 }} />
                          </div>
                          <p style={POPUP_STYLES.progressLabel}>{Math.round((route.progress || 0) * 100)}% completato (stimato)</p>
                        </div>
                      </Popup>
                    </Marker>
                  )}
                </div>
              );
            })}
          </MapContainer>
        </Card>
      </div>

      {/* Colonna destra: pannello viaggi a tutta altezza, la lista scorre al
          suo interno. La larghezza la decide la griglia (1fr), non un w-80
          fisso: è la proporzione del mockup. */}
      <Card
        data-map-panel
        className="h-[70vh] lg:h-auto min-w-0 lg:min-h-0 rounded-xl shadow-sm overflow-hidden flex flex-col"
        data-testid="map-sidebar"
      >
          <div className="px-4 py-3 border-b bg-muted/30">
            <h3 className="font-display text-sm font-semibold">
              Viaggi sulla mappa ({panelRoutes.length})
            </h3>
          </div>
          {/* Ricerca nel pannello: c'è nel prototipo e qui mancava del tutto.
              Filtra solo la lista laterale (progressivo, cliente, autista,
              targa, tratta) — i tracciati sulla mappa restano quelli scelti
              con i filtri della toolbar. */}
          <div className="px-3 py-2.5 border-b shrink-0">
            <div className="relative">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-muted-foreground" />
              <Input
                value={panelQuery}
                onChange={(e) => setPanelQuery(e.target.value)}
                placeholder="Cerca viaggio, autista, targa..."
                className="pl-9 h-9 text-sm"
                data-testid="map-panel-search"
              />
            </div>
          </div>
          <div className="flex-1 overflow-y-auto">
            {panelRoutes.length === 0 ? (
              <p className="text-sm text-muted-foreground p-4 text-center">
                {panelQuery ? 'Nessun viaggio corrisponde alla ricerca.' : 'Nessun viaggio da mostrare con i filtri attivi.'}
              </p>
            ) : (
              <div className="divide-y">
                {panelRoutes.map(route => (
                  <button
                    key={route.id}
                    className={`w-full text-left px-4 py-3 transition-colors duration-150 hover:bg-muted/50 ${selectedRoute?.id === route.id ? 'bg-accent/50 border-l-2 border-l-primary' : ''}`}
                    onClick={() => setSelectedRoute(route)}
                    data-testid="map-route-item"
                  >
                    <div className="flex items-center justify-between mb-1">
                      <span className="font-mono text-xs font-medium">{route.progressivo}</span>
                      <StatusBadge stato={route.stato} />
                    </div>
                    {/* La tratta: era un "→" nudo perché i punti carico/
                        scarico dell'API non portavano il nome (ora sì, vedi
                        dto.MapNamedPoint lato backend). */}
                    <p className="text-xs font-medium truncate">
                      {route.carico?.nome || '?'} → {route.scarico?.nome || '?'}
                    </p>
                    <div className="flex items-center gap-2 mt-1.5">
                      {route.motrice?.targa && (
                        <span className="inline-flex items-center gap-1 text-[10px] px-1.5 py-0.5 rounded bg-muted font-mono">
                          <Truck className="h-2.5 w-2.5" /> {route.motrice.targa}
                        </span>
                      )}
                      {route.autista && (
                        <span className="text-[10px] text-muted-foreground truncate">{route.autista.nome} {route.autista.cognome}</span>
                      )}
                    </div>
                    {(route.distance_km || 0) > 0 && (
                      <p className="text-[10px] text-muted-foreground mt-1 tabular-nums">{route.distance_km} km • {route.duration_hours}h stimate</p>
                    )}
                    {/* Barra progresso */}
                    <div className="mt-2 h-1.5 bg-muted rounded-full overflow-hidden">
                      <div
                        className="h-full rounded-full transition-all duration-500"
                        style={{ width: `${(route.progress || 0) * 100}%`, backgroundColor: statusColors.VIAGGIO }}
                      />
                    </div>
                    <div className="flex justify-between mt-1">
                      <span className="text-[10px] text-muted-foreground">
                        {(route.remaining_km || 0) > 0 ? `${route.remaining_km} km rimasti` : route.cliente?.ragione_sociale}
                        {(route.eta_hours || 0) > 0 ? ` • ETA ~${route.eta_hours}h` : ''}
                      </span>
                      <span className="text-[10px] text-muted-foreground tabular-nums">€ {formatEuro(route.tariffa || 0)}</span>
                    </div>
                  </button>
                ))}
              </div>
            )}
          </div>

          {/* Sezione info selezionata */}
          {selectedRoute && (
            <div className="border-t px-4 py-3 bg-card shrink-0">
              <div className="flex items-center justify-between mb-2">
                <span className="font-display text-xs font-semibold">Dettaglio</span>
                <Button variant="ghost" size="sm" className="h-6 text-[10px] px-2" onClick={() => setSelectedRoute(null)}>Chiudi</Button>
              </div>
              <div className="space-y-1 text-xs">
                <div className="flex justify-between"><span className="text-muted-foreground">Ordine:</span><span className="font-mono">{selectedRoute.progressivo}</span></div>
                <div className="flex justify-between"><span className="text-muted-foreground">Cliente:</span><span className="truncate ml-2">{selectedRoute.cliente?.ragione_sociale}</span></div>
                <div className="flex justify-between"><span className="text-muted-foreground">Tratta:</span><span className="truncate ml-2">{selectedRoute.carico?.nome || '?'} → {selectedRoute.scarico?.nome || '?'}</span></div>
                {selectedRoute.garage && <div className="flex justify-between"><span className="text-muted-foreground">Partenza:</span><span className="truncate ml-2">{selectedRoute.garage.nome}</span></div>}
                {selectedRoute.wash_station && <div className="flex justify-between"><span className="text-muted-foreground">Lavaggio:</span><span className="truncate ml-2">{selectedRoute.wash_station.nome}</span></div>}
                <div className="flex justify-between"><span className="text-muted-foreground">Mezzo:</span><span className="font-mono">{selectedRoute.motrice?.targa || '—'}</span></div>
                <div className="flex justify-between"><span className="text-muted-foreground">Autista:</span><span>{selectedRoute.autista ? `${selectedRoute.autista.nome} ${selectedRoute.autista.cognome}` : '—'}</span></div>
                <div className="flex justify-between"><span className="text-muted-foreground">Distanza:</span><span className="tabular-nums">{selectedRoute.distance_km} km • {selectedRoute.duration_hours}h</span></div>
                {(selectedRoute.remaining_km || 0) > 0 && <div className="flex justify-between"><span className="text-muted-foreground">Rimanenti:</span><span className="tabular-nums font-medium">{selectedRoute.remaining_km} km • ETA ~{selectedRoute.eta_hours}h</span></div>}
                <div className="flex justify-between"><span className="text-muted-foreground">Tariffa:</span><span className="font-medium">€ {selectedRoute.tariffa?.toLocaleString('it-IT')}</span></div>
              </div>
            </div>
          )}
      </Card>
    </div>
  );
}
