import { useState, useEffect, useCallback, useMemo } from 'react';
import { getMapTrips } from '@/lib/api';
import { formatEuro } from '@/lib/format';
import { Card } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Skeleton } from '@/components/ui/skeleton';
import { StatusBadge } from '@/components/shared/StatusBadge';
import { MapContainer, TileLayer, LayersControl, Marker, Popup, Polyline, CircleMarker, Tooltip, useMap } from 'react-leaflet';
import L from 'leaflet';
import { Truck, RefreshCw, Eye, EyeOff } from 'lucide-react';
import { logger } from '@/lib/logger';

// Fix icone Leaflet
delete L.Icon.Default.prototype._getIconUrl;
L.Icon.Default.mergeOptions({
  iconRetinaUrl: 'https://cdnjs.cloudflare.com/ajax/libs/leaflet/1.9.4/images/marker-icon-2x.png',
  iconUrl: 'https://cdnjs.cloudflare.com/ajax/libs/leaflet/1.9.4/images/marker-icon.png',
  shadowUrl: 'https://cdnjs.cloudflare.com/ajax/libs/leaflet/1.9.4/images/marker-shadow.png',
});

// Icone custom
const createIcon = (color, size = 28, isLive = false) => L.divIcon({
  className: '',
  html: `<div style="width:${size}px;height:${size}px;border-radius:50%;background:${color};border:3px solid ${isLive ? '#22D3EE' : 'white'};box-shadow:0 2px 8px rgba(0,0,0,0.3)${isLive ? ',0 0 12px rgba(34,211,238,0.5)' : ''};display:flex;align-items:center;justify-content:center;${isLive ? 'animation:pulse 2s infinite;' : ''}">
    <svg width="${size * 0.5}" height="${size * 0.5}" viewBox="0 0 24 24" fill="none" stroke="white" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><rect x="1" y="3" width="15" height="13"/><polygon points="16 8 20 8 23 11 23 16 16 16 16 8"/><circle cx="5.5" cy="18.5" r="2.5"/><circle cx="18.5" cy="18.5" r="2.5"/></svg>
  </div>${isLive ? '<style>@keyframes pulse{0%,100%{opacity:1}50%{opacity:0.7}}</style>' : ''}`,
  iconSize: [size, size],
  iconAnchor: [size / 2, size / 2],
  popupAnchor: [0, -size / 2],
});

const garageIcon = L.divIcon({
  className: '',
  html: `<div style="width:32px;height:32px;border-radius:8px;background:#0B1220;border:3px solid #22D3EE;box-shadow:0 2px 12px rgba(34,211,238,0.4);display:flex;align-items:center;justify-content:center;">
    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="#22D3EE" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"/><polyline points="9 22 9 12 15 12 15 22"/></svg>
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
  html: `<div style="width:10px;height:10px;border-radius:50%;background:hsl(195,92%,28%);border:2px solid white;box-shadow:0 1px 4px rgba(0,0,0,0.2);"></div>`,
  iconSize: [10, 10],
  iconAnchor: [5, 5],
});

const statusColors = {
  VIAGGIO: '#E24A4A',
  PIANIFICABILE: '#F0B429',
  CHIUSO: '#F28B2C',
  FATTURATO: '#2AA36B',
};

// Stili costanti per pathOptions (evitano re-render)
const CIRCLE_MARKER_BASE = { color: 'white', fillOpacity: 1, weight: 2 };
const SHADOW_PATH_OPTIONS = { color: '#000', weight: 8, opacity: 0.15, lineCap: 'round', lineJoin: 'round' };
const MAP_CONTAINER_STYLE = { height: '100%', width: '100%' };
const POPUP_STYLES = {
  container: { minWidth: 220 },
  header: { display: 'flex', alignItems: 'center', gap: 6, marginBottom: 4 },
  title: { margin: 0, fontWeight: 700, fontSize: 14, flex: 1 },
  gpsBadge: { background: '#22D3EE', color: '#0B1220', fontSize: 9, fontWeight: 700, padding: '1px 6px', borderRadius: 10 },
  subtitle: { margin: 0, fontSize: 12, color: '#666' },
  hr: { margin: '6px 0', border: 'none', borderTop: '1px solid #eee' },
  route: { margin: '2px 0', fontSize: 12 },
  detail: { margin: '2px 0', fontSize: 11, color: '#888' },
  gpsInfo: { margin: '2px 0', fontSize: 11, color: '#0E7A81', fontWeight: 600 },
  progressBg: { marginTop: 6, background: '#f0f0f0', borderRadius: 6, height: 6, overflow: 'hidden' },
  progressLabel: { margin: '4px 0 0', fontSize: 10, color: '#999' },
  trackerLink: { display: 'block', marginTop: 6, fontSize: 10, color: '#0E7A81', textDecoration: 'underline' },
};

// Componente per auto-fit bounds
const FitBounds = ({ routes, garages, washStations }) => {
  const map = useMap();
  useEffect(() => {
    if (routes.length === 0 && garages.length === 0 && washStations.length === 0) return;
    const allPoints = [];
    routes.forEach(r => {
      if (r.carico) allPoints.push([r.carico.lat, r.carico.lng]);
      if (r.scarico) allPoints.push([r.scarico.lat, r.scarico.lng]);
    });
    garages.forEach(g => allPoints.push([g.lat, g.lng]));
    washStations.forEach(w => allPoints.push([w.lat, w.lng]));
    if (allPoints.length > 0) {
      map.fitBounds(allPoints, { padding: [40, 40], maxZoom: 7 });
    }
  }, [routes, garages, washStations, map]);
  return null;
};

export default function MapPage() {
  const [data, setData] = useState(null);
  const [loading, setLoading] = useState(true);
  const [selectedRoute, setSelectedRoute] = useState(null);
  const [showPianificabili, setShowPianificabili] = useState(false);
  const [showChiusi, setShowChiusi] = useState(false);
  const [filterVeicolo, setFilterVeicolo] = useState('');

  const fetchData = useCallback(() => {
    setLoading(true);
    getMapTrips().then(r => setData(r.data)).catch(err => logger.error('Errore caricamento mappa:', err)).finally(() => setLoading(false));
  }, []);
  useEffect(() => { fetchData(); }, [fetchData]);

  const filteredRoutes = useMemo(() => {
    if (!data) return [];
    return data.routes.filter(r => {
      if (!showPianificabili && r.stato === 'PIANIFICABILE') return false;
      if (!showChiusi && r.stato === 'CHIUSO') return false;
      if (filterVeicolo && r.targa_motrice !== filterVeicolo) return false;
      if (!r.carico || !r.scarico) return false;
      if (!r.carico.lat || !r.carico.lng || !r.scarico.lat || !r.scarico.lng) return false;
      if (isNaN(r.carico.lat) || isNaN(r.scarico.lat)) return false;
      return true;
    });
  }, [data, showPianificabili, showChiusi, filterVeicolo]);

  const inViaggio = useMemo(() => filteredRoutes.filter(r => r.stato === 'VIAGGIO'), [filteredRoutes]);
  const uniqueVehicles = useMemo(() => data ? [...new Set(data.routes.map(r => r.targa_motrice).filter(Boolean))] : [], [data]);

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
    <div className="space-y-3" data-testid="map-page">
      {/* Toolbar */}
      <div className="flex flex-col gap-2 lg:flex-row lg:items-center lg:justify-between">
        <div className="flex items-center gap-2 flex-wrap">
          <Badge className="status-viaggio border text-xs font-medium">{data.stats.in_viaggio} in viaggio</Badge>
          {data.stats.gps_live > 0 && <Badge className="bg-emerald-100 text-emerald-800 border-emerald-300 border text-xs font-medium">{data.stats.gps_live} GPS live</Badge>}
          <Badge className="status-pianificabile border text-xs font-medium">{data.stats.pianificabili} da pianificare</Badge>
          <Badge className="status-chiuso border text-xs font-medium">{data.stats.chiusi} chiusi</Badge>
        </div>
        <div className="flex items-center gap-2 flex-wrap">
          <Button
            variant={showPianificabili ? 'default' : 'outline'} size="sm" className="text-xs gap-1.5 h-8"
            onClick={() => setShowPianificabili(!showPianificabili)}
          >
            {showPianificabili ? <Eye className="h-3.5 w-3.5" /> : <EyeOff className="h-3.5 w-3.5" />} Pianificabili
          </Button>
          <Button
            variant={showChiusi ? 'default' : 'outline'} size="sm" className="text-xs gap-1.5 h-8"
            onClick={() => setShowChiusi(!showChiusi)}
          >
            {showChiusi ? <Eye className="h-3.5 w-3.5" /> : <EyeOff className="h-3.5 w-3.5" />} Chiusi
          </Button>
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
        </div>
      </div>

      {/* Layout: mappa + pannello laterale */}
      <div className="flex gap-3 h-[calc(100vh-200px)]">
        {/* Mappa */}
        <Card className="flex-1 rounded-xl overflow-hidden shadow-sm relative" data-testid="map-container">
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
              <LayersControl.BaseLayer checked name="OpenStreetMap">
                <TileLayer
                  attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'
                  url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
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
            <FitBounds routes={filteredRoutes} garages={data.garages} washStations={data.wash_stations || []} />

            {/* Garage */}
            {data.garages.map((g, i) => (
              <Marker key={`g-${i}`} position={[g.lat, g.lng]} icon={garageIcon}>
                <Popup><strong>{g.nome}</strong><br />Base operativa</Popup>
                <Tooltip direction="top" offset={[0, -20]} permanent={false}>{g.nome}</Tooltip>
              </Marker>
            ))}

            {/* Punti di lavaggio */}
            {(data.wash_stations || []).map((w, i) => (
              <Marker key={`wash-${i}`} position={[w.lat, w.lng]} icon={washIcon}>
                <Popup><strong>{w.nome}</strong><br />Punto di lavaggio</Popup>
                <Tooltip direction="top" offset={[0, -16]} permanent={false}>{w.nome}</Tooltip>
              </Marker>
            ))}

            {/* Destinazioni (punti piccoli) */}
            {data.poi.map((p, i) => (
              <Marker key={`poi-${i}`} position={[p.lat, p.lng]} icon={destIcon}>
                <Tooltip direction="top" offset={[0, -8]}>{p.nome}</Tooltip>
              </Marker>
            ))}

            {/* Percorsi stradali reali */}
            {filteredRoutes.map((route) => {
              if (!route.carico || !route.scarico) return null;
              if (!route.carico.lat || !route.scarico.lat) return null;
              if (isNaN(route.carico.lat) || isNaN(route.scarico.lat)) return null;

              // Usa percorso stradale dal backend, fallback a linea diretta
              const roadPts = route.road_points && route.road_points.length > 1
                ? route.road_points
                : [[route.carico.lat, route.carico.lng], [route.scarico.lat, route.scarico.lng]];

              const color = statusColors[route.stato] || '#888';
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
                      dashArray: isActive ? null : '6 4',
                      lineCap: 'round',
                      lineJoin: 'round',
                    }}
                    eventHandlers={{ click: () => setSelectedRoute(route) }}
                  />

                  {/* Marker carico */}
                  <CircleMarker
                    center={[route.carico.lat, route.carico.lng]}
                    radius={isSelected ? 7 : 5}
                    pathOptions={{ ...CIRCLE_MARKER_BASE, fillColor: color }}
                  >
                    <Tooltip direction="top">
                      <strong>Carico:</strong> {route.carico.nome}<br />
                      {route.progressivo} — {route.cliente_nome}
                    </Tooltip>
                  </CircleMarker>

                  {/* Marker scarico */}
                  <CircleMarker
                    center={[route.scarico.lat, route.scarico.lng]}
                    radius={isSelected ? 7 : 5}
                    pathOptions={{ ...CIRCLE_MARKER_BASE, fillColor: color }}
                  >
                    <Tooltip direction="top">
                      <strong>Scarico:</strong> {route.scarico.nome}
                    </Tooltip>
                  </CircleMarker>

                  {/* Posizione attuale del mezzo (solo per VIAGGIO) */}
                  {isActive && route.current_position && (
                    <Marker
                      position={[route.current_position.lat, route.current_position.lng]}
                      icon={createIcon(color, isSelected ? 36 : 30, route.gps_live)}
                      eventHandlers={{ click: () => setSelectedRoute(route) }}
                    >
                      <Popup className="custom-popup">
                        <div style={POPUP_STYLES.container}>
                          <div style={POPUP_STYLES.header}>
                            <p style={POPUP_STYLES.title}>{route.targa_motrice || 'Mezzo'}</p>
                            {route.gps_live && route.gps_source === 'simulated' && (
                              <span style={{ ...POPUP_STYLES.gpsBadge, background: '#FEF2C0', color: '#7A4D00', borderColor: '#E1B768' }}>DEMO</span>
                            )}
                            {route.gps_live && route.gps_source && route.gps_source !== 'simulated' && (
                              <span style={POPUP_STYLES.gpsBadge}>GPS {route.gps_source.replace(/^provider_/, '').toUpperCase()}</span>
                            )}
                            {route.gps_live && !route.gps_source && <span style={POPUP_STYLES.gpsBadge}>GPS LIVE</span>}
                            {route.last_temp_alert && (
                              <span style={{ ...POPUP_STYLES.gpsBadge, background: '#FEE2E2', color: '#991B1B', borderColor: '#FCA5A5' }}>
                                TEMP ALERT
                              </span>
                            )}
                          </div>
                          {route.last_temp_celsius != null && (
                            <p style={POPUP_STYLES.detail}>
                              Temperatura: <strong>{route.last_temp_celsius.toFixed(1)} °C</strong>
                              {route.last_temp_alert ? ' (fuori soglia)' : ''}
                            </p>
                          )}
                          <p style={POPUP_STYLES.subtitle}>{route.autista_nome || 'N/A'}</p>
                          <hr style={POPUP_STYLES.hr} />
                          <p style={POPUP_STYLES.route}><strong>{route.carico.nome}</strong> → <strong>{route.scarico.nome}</strong></p>
                          <p style={POPUP_STYLES.detail}>{route.cliente_nome} • {route.progressivo}</p>
                          {route.distance_km > 0 && <p style={POPUP_STYLES.detail}>{route.distance_km} km totali • {route.duration_hours}h stimate</p>}
                          {route.gps_live && route.gps_speed_kmh > 0 && (
                            <p style={POPUP_STYLES.gpsInfo}>Velocità: {Math.round(route.gps_speed_kmh)} km/h</p>
                          )}
                          {route.remaining_km > 0 && (
                            <p style={POPUP_STYLES.gpsInfo}>Rimanenti: {route.remaining_km} km • ETA ~{route.eta_hours}h</p>
                          )}
                          <div style={POPUP_STYLES.progressBg}>
                            <div style={{ width: `${route.progress * 100}%`, height: '100%', background: route.gps_live ? '#22D3EE' : color, borderRadius: 6 }} />
                          </div>
                          <p style={POPUP_STYLES.progressLabel}>{Math.round(route.progress * 100)}% completato{route.gps_live ? ' (GPS)' : ' (stimato)'}</p>
                          {route.gps_tracker_url && (
                            <a href={route.gps_tracker_url} target="_blank" rel="noopener noreferrer" style={{ display: 'block', marginTop: 6, fontSize: 10, color: '#0E7A81', textDecoration: 'underline' }}>Apri tracker GPS →</a>
                          )}
                        </div>
                      </Popup>
                    </Marker>
                  )}
                </div>
              );
            })}
          </MapContainer>
        </Card>

        {/* Pannello laterale */}
        <Card className="w-80 shrink-0 rounded-xl shadow-sm overflow-hidden flex flex-col" data-testid="map-sidebar">
          <div className="px-4 py-3 border-b bg-muted/30">
            <h3 className="text-sm font-semibold" style={{ fontFamily: "'Space Grotesk', sans-serif" }}>
              Viaggi attivi ({inViaggio.length})
            </h3>
          </div>
          <div className="flex-1 overflow-y-auto">
            {inViaggio.length === 0 ? (
              <p className="text-sm text-muted-foreground p-4 text-center">Nessun viaggio attivo con percorso mappabile.</p>
            ) : (
              <div className="divide-y">
                {inViaggio.map(route => (
                  <button
                    key={route.id}
                    className={`w-full text-left px-4 py-3 transition-colors duration-150 hover:bg-muted/50 ${selectedRoute?.id === route.id ? 'bg-accent/50 border-l-2 border-l-primary' : ''}`}
                    onClick={() => setSelectedRoute(route)}
                    data-testid="map-route-item"
                  >
                    <div className="flex items-center justify-between mb-1">
                      <span className="font-mono text-xs font-medium">{route.progressivo}</span>
                      <div className="flex items-center gap-1.5">
                        {route.gps_live && <span className="inline-flex items-center gap-1 text-[9px] px-1.5 py-0.5 rounded-full bg-cyan-100 text-cyan-800 font-semibold"><span className="h-1.5 w-1.5 rounded-full bg-cyan-500 animate-pulse" />LIVE</span>}
                        <StatusBadge stato={route.stato} />
                      </div>
                    </div>
                    <p className="text-xs font-medium truncate">{route.carico?.nome} → {route.scarico?.nome}</p>
                    <div className="flex items-center gap-2 mt-1.5">
                      {route.targa_motrice && (
                        <span className="inline-flex items-center gap-1 text-[10px] px-1.5 py-0.5 rounded bg-muted font-mono">
                          <Truck className="h-2.5 w-2.5" /> {route.targa_motrice}
                        </span>
                      )}
                      {route.autista_nome && (
                        <span className="text-[10px] text-muted-foreground truncate">{route.autista_nome}</span>
                      )}
                    </div>
                    {route.distance_km > 0 && (
                      <p className="text-[10px] text-muted-foreground mt-1 tabular-nums">{route.distance_km} km • {route.duration_hours}h stimate</p>
                    )}
                    {/* Barra progresso */}
                    <div className="mt-2 h-1.5 bg-muted rounded-full overflow-hidden">
                      <div
                        className="h-full rounded-full transition-all duration-500"
                        style={{ width: `${route.progress * 100}%`, backgroundColor: route.gps_live ? '#22D3EE' : statusColors.VIAGGIO }}
                      />
                    </div>
                    <div className="flex justify-between mt-1">
                      <span className="text-[10px] text-muted-foreground">
                        {route.remaining_km > 0 ? `${route.remaining_km} km rimasti` : route.cliente_nome}
                        {route.eta_hours > 0 ? ` • ETA ~${route.eta_hours}h` : ''}
                        {route.gps_live && route.gps_speed_kmh > 0 ? ` • ${Math.round(route.gps_speed_kmh)} km/h` : ''}
                      </span>
                      <span className="text-[10px] text-muted-foreground tabular-nums">€ {formatEuro(route.tariffa)}</span>
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
                <span className="text-xs font-semibold" style={{ fontFamily: "'Space Grotesk', sans-serif" }}>Dettaglio</span>
                <Button variant="ghost" size="sm" className="h-6 text-[10px] px-2" onClick={() => setSelectedRoute(null)}>Chiudi</Button>
              </div>
              <div className="space-y-1 text-xs">
                <div className="flex justify-between"><span className="text-muted-foreground">Ordine:</span><span className="font-mono">{selectedRoute.progressivo}</span></div>
                <div className="flex justify-between"><span className="text-muted-foreground">Cliente:</span><span className="truncate ml-2">{selectedRoute.cliente_nome}</span></div>
                <div className="flex justify-between"><span className="text-muted-foreground">Tratta:</span><span className="truncate ml-2">{selectedRoute.carico?.nome} → {selectedRoute.scarico?.nome}</span></div>
                <div className="flex justify-between"><span className="text-muted-foreground">Mezzo:</span><span className="font-mono">{selectedRoute.targa_motrice || '—'}</span></div>
                <div className="flex justify-between"><span className="text-muted-foreground">Autista:</span><span>{selectedRoute.autista_nome || '—'}</span></div>
                <div className="flex justify-between"><span className="text-muted-foreground">Distanza:</span><span className="tabular-nums">{selectedRoute.distance_km} km • {selectedRoute.duration_hours}h</span></div>
                {selectedRoute.remaining_km > 0 && <div className="flex justify-between"><span className="text-muted-foreground">Rimanenti:</span><span className="tabular-nums font-medium">{selectedRoute.remaining_km} km • ETA ~{selectedRoute.eta_hours}h</span></div>}
                {selectedRoute.gps_live && <div className="flex justify-between"><span className="text-muted-foreground">GPS:</span><span className="text-cyan-600 font-medium">{Math.round(selectedRoute.gps_speed_kmh)} km/h — LIVE</span></div>}
                <div className="flex justify-between"><span className="text-muted-foreground">Tariffa:</span><span className="font-medium">€ {selectedRoute.tariffa?.toLocaleString('it-IT')}</span></div>
                {selectedRoute.gps_tracker_url && (
                  <a href={selectedRoute.gps_tracker_url} target="_blank" rel="noopener noreferrer" className="block mt-1 text-primary hover:underline text-[10px]">Apri portale GPS tracker →</a>
                )}
              </div>
            </div>
          )}
        </Card>
      </div>
    </div>
  );
}
