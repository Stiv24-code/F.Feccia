import { useState, useEffect, useMemo, useCallback } from 'react';
import { assignOrder, getMotrici, getSemirimorchi, getDrivers, getCarriers, getMotriceAvailability, getSemirimorchioAvailability, getDriverAvailability, getOrderRouteAlternatives } from '@/lib/api';
import { useGetGaragesQuery, useGetWashStationsQuery } from '@/store/api/appApi';
import { getApiErrorMessage } from '@/lib/apiError';
import type {
  DtoOrderResponse,
  DtoOrderAssignRequest,
  DtoMotriceResponse,
  DtoSemirimorchioResponse,
  DtoDriverResponse,
  DtoCarrierResponse,
  DtoMotriceAvailabilityResponse,
  DtoSemirimorchioAvailabilityResponse,
  DtoDriverAvailabilityResponse,
  DtoRouteAlternativeDTO,
  DtoRouteWaypointResponseDTO,
} from '@/api/data-contracts';
import { Button } from '@/components/ui/button';
import { Label } from '@/components/ui/label';
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs';
import LocationCombobox from '@/components/shared/LocationCombobox';
import SearchableSelect from '@/components/shared/SearchableSelect';
import OrderRouteMap from '@/components/shared/OrderRouteMap';
import { toast } from 'sonner';
import { Loader2, Warehouse, Droplets, Check, Route } from 'lucide-react';
import { logger } from '@/lib/logger';

type TransportMode = 'proprio' | 'terzo';

// Trova l'n-esimo waypoint di un certo tipo in una route alternativa — ordine
// garantito dal backend (garage?, carico, scarico, wash_station?), quindi
// idx=1 sul tipo "destinazione" è sempre lo scarico.
const waypointByTipo = (waypoints: DtoRouteWaypointResponseDTO[] | undefined, tipo: string, idx = 0) => (waypoints || []).filter(w => w.tipo === tipo)[idx];

const EMPTY_FORM: DtoOrderAssignRequest = {
  garage_id: '',
  motrice_id: '', semirimorchio_id: '', autista_id: '',
  vettore_id: '',
  wash_station_id: '',
};

const formatDataBreve = (iso?: string) => {
  if (!iso) return '';
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return '';
  return d.toLocaleDateString('it-IT', { day: 'numeric', month: 'short' });
};

export interface AssignOrderFormProps {
  order: DtoOrderResponse;
  onAssigned?: () => void;
  onCancel?: () => void;
}

// Form "Assegna trasporto" — corpo condiviso tra AssignOrderDialog (modale,
// usata da PlannerPage) e OrderDetailPage (inline, ordine in PIANIFICABILE),
// cosi' la logica di disponibilita' mezzi/autisti e le select punto di
// partenza/lavaggio vivono in un solo posto invece di essere duplicate.
export default function AssignOrderForm({ order, onAssigned, onCancel }: AssignOrderFormProps) {
  const [form, setForm] = useState<DtoOrderAssignRequest>(EMPTY_FORM);
  const [transportMode, setTransportMode] = useState<TransportMode>('proprio');
  const [saving, setSaving] = useState(false);
  const [motrici, setMotrici] = useState<DtoMotriceResponse[]>([]);
  const [semirimorchi, setSemirimorchi] = useState<DtoSemirimorchioResponse[]>([]);
  const [drivers, setDrivers] = useState<DtoDriverResponse[]>([]);
  const [carriers, setCarriers] = useState<DtoCarrierResponse[]>([]);
  const [availMotrici, setAvailMotrici] = useState<DtoMotriceAvailabilityResponse[]>([]);
  const [availSemirimorchi, setAvailSemirimorchi] = useState<DtoSemirimorchioAvailabilityResponse[]>([]);
  const [availDrivers, setAvailDrivers] = useState<DtoDriverAvailabilityResponse[]>([]);
  const [routeAlternatives, setRouteAlternatives] = useState<DtoRouteAlternativeDTO[]>([]);
  const [selectedRouteIdx, setSelectedRouteIdx] = useState(0);
  const [routeLoading, setRouteLoading] = useState(false);

  // Selettore mezzi/garage: serve l'elenco completo, non una pagina — limit
  // alto per replicare il comportamento pre-paginazione (cap lato backend).
  const { data: garagesPage } = useGetGaragesQuery({ limit: 500 });
  const { data: washStationsPage } = useGetWashStationsQuery({ limit: 500 });
  const garages = garagesPage?.items ?? [];
  const washStations = washStationsPage?.items ?? [];

  useEffect(() => {
    if (!order) return;
    setForm(EMPTY_FORM);
    setTransportMode('proprio');
    setAvailMotrici([]);
    setAvailSemirimorchi([]);
    setAvailDrivers([]);
    setRouteAlternatives([]);
    setSelectedRouteIdx(0);
    Promise.all([getMotrici(), getSemirimorchi(), getDrivers(), getCarriers()]).then(([m, s, d, c]) => {
      setMotrici(m.data.data ?? []); setSemirimorchi(s.data.data ?? []); setDrivers(d.data.data ?? []); setCarriers(c.data.data ?? []);
    }).catch(err => logger.error('Errore caricamento lookup assegna:', err));
    const da = order.data_ritiro || '';
    const a = order.data_consegna || order.data_ritiro || '';
    if (da && a) {
      Promise.all([getMotriceAvailability(da, a), getSemirimorchioAvailability(da, a), getDriverAvailability(da, a)]).then(([mRes, sRes, dRes]) => {
        setAvailMotrici(mRes.data); setAvailSemirimorchi(sRes.data); setAvailDrivers(dRes.data);
      }).catch(err => logger.error('Disponibilità fetch error:', err));
    }
  }, [order]);

  // Ricalcola le alternative ogni volta che cambia garage/punto di lavaggio
  // (carico/scarico sono fissi, vengono dall'ordine) — la prima resta
  // preselezionata di default così un percorso è sempre pronto anche se il
  // manager non interagisce con le card. fetchRouteAlternatives è estratta
  // per essere riutilizzabile anche dal bottone "Riprova" — ORS ogni tanto
  // fallisce in modo transitorio (timeout, rate limit del piano gratuito) e
  // senza un modo per ritentare l'unica opzione sarebbe toccare di nuovo
  // garage/lavaggio solo per ri-triggerare l'effect.
  const fetchRouteAlternatives = useCallback((orderId: string, garageId?: string, washStationId?: string) => {
    setRouteLoading(true);
    getOrderRouteAlternatives(orderId, { garageId, washStationId })
      .then((r: { data: { alternatives?: DtoRouteAlternativeDTO[] } }) => { setRouteAlternatives(r.data.alternatives || []); setSelectedRouteIdx(0); })
      .catch((err: unknown) => { logger.error('Errore calcolo percorso:', err); setRouteAlternatives([]); })
      .finally(() => setRouteLoading(false));
  }, []);

  useEffect(() => {
    if (!order?.id) return;
    fetchRouteAlternatives(order.id, form.garage_id, form.wash_station_id);
  }, [order?.id, form.garage_id, form.wash_station_id, fetchRouteAlternatives]);

  const setGarage = (id: string) => setForm(f => ({ ...f, garage_id: id }));
  const setWashStation = (id: string) => setForm(f => ({ ...f, wash_station_id: id }));
  const setDriver = (id: string) => setForm(f => ({ ...f, autista_id: id }));
  const setVettore = (id: string) => setForm(f => ({ ...f, vettore_id: id }));
  const selectTransportMode = (mode: string) => {
    setTransportMode(mode as TransportMode);
    if (mode === 'proprio') {
      setForm(f => ({ ...f, vettore_id: '' }));
    } else {
      setForm(f => ({ ...f, motrice_id: '', semirimorchio_id: '', autista_id: '' }));
    }
  };

  const assignMotrici = useMemo(() => availMotrici.length > 0 ? availMotrici : motrici, [availMotrici, motrici]);
  const assignRimorchi = useMemo(() => availSemirimorchi.length > 0 ? availSemirimorchi : semirimorchi, [availSemirimorchi, semirimorchi]);
  const assignDriverList = useMemo(() => availDrivers.length > 0 ? availDrivers : drivers, [availDrivers, drivers]);
  const disponibilitaLabel = useMemo(() => formatDataBreve(order?.data_ritiro), [order]);

  // "Assegna Viaggio" richiede: chi effettua il trasporto (autista se mezzo
  // proprio, vettore se terzo) E un percorso calcolato/selezionato — nessuno
  // dei due è opzionale per assegnare davvero l'ordine.
  const hasTransport = transportMode === 'proprio' ? !!form.autista_id : !!form.vettore_id;
  const hasRoute = !routeLoading && !!routeAlternatives[selectedRouteIdx];
  const canSubmit = hasTransport && hasRoute;

  const handleAssign = async () => {
    setSaving(true);
    try {
      const selectedRoute = routeAlternatives[selectedRouteIdx];
      const payload: DtoOrderAssignRequest = selectedRoute
        ? { ...form, route_waypoints: (selectedRoute.waypoints || []).map(w => ({ tipo: w.tipo as 'garage' | 'destinazione' | 'wash_station', ref_id: w.ref_id || '' })) }
        : form;
      await assignOrder(order.id, payload);
      toast.success(`Ordine ${order.progressivo} assegnato`);
      if (onAssigned) onAssigned();
    } catch (e) { toast.error(getApiErrorMessage(e) || 'Errore'); } finally { setSaving(false); }
  };

  return (
    <div className="space-y-5">
      <div>
        <p className="text-[10px] uppercase tracking-wide text-muted-foreground font-semibold mb-3">Dettagli trasporto</p>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
          <div className="space-y-1.5">
            <Label>Punto di partenza</Label>
            <LocationCombobox
              value={form.garage_id}
              onChange={setGarage}
              options={garages}
              placeholder="Seleziona punto di partenza..."
              searchPlaceholder="Cerca garage..."
              icon={Warehouse}
              iconBg="#1c2534"
              iconColor="#fff"
            />
          </div>
          <div className="space-y-1.5">
            <Label>Punto di lavaggio — dopo lo scarico</Label>
            <LocationCombobox
              value={form.wash_station_id}
              onChange={setWashStation}
              options={washStations}
              placeholder="Seleziona punto di lavaggio..."
              searchPlaceholder="Cerca stazione o tipo lavaggio..."
              icon={Droplets}
              iconBg="#e6f4f2"
              iconColor="#0d9488"
              getSubtitle={(w) => w.tipo || w.indirizzo}
            />
          </div>
        </div>
      </div>

      <div>
        <p className="text-[10px] uppercase tracking-wide text-muted-foreground font-semibold mb-3 flex items-center gap-1.5">
          <Route className="h-3 w-3" /> Percorso
        </p>
        {routeLoading ? (
          <div className="flex items-center gap-2 text-xs text-muted-foreground py-4">
            <Loader2 className="h-3.5 w-3.5 animate-spin" /> Calcolo percorso in corso...
          </div>
        ) : routeAlternatives.length === 0 ? (
          <div className="flex items-center justify-between gap-2 py-2">
            <p className="text-xs text-muted-foreground">Percorso non disponibile (routing non configurato, coordinate mancanti, o errore momentaneo del servizio).</p>
            <Button type="button" variant="outline" size="sm" onClick={() => order.id && fetchRouteAlternatives(order.id, form.garage_id, form.wash_station_id)}>
              Riprova
            </Button>
          </div>
        ) : (
          <>
            {routeAlternatives.length === 1 && (routeAlternatives[0].distance_km || 0) > 100 && (
              <p className="text-xs text-muted-foreground mb-2">
                Percorso oltre i 100 km: OpenRouteService non calcola percorsi alternativi multipli oltre questa soglia, mostrato l&apos;unico percorso disponibile.
              </p>
            )}
            <div className={`grid grid-cols-1 gap-3 ${routeAlternatives.length > 1 ? 'md:grid-cols-2' : ''}`} data-testid="route-alternatives">
            {routeAlternatives.map((alt, idx) => {
              const carico = waypointByTipo(alt.waypoints, 'destinazione', 0);
              const scarico = waypointByTipo(alt.waypoints, 'destinazione', 1);
              const garage = waypointByTipo(alt.waypoints, 'garage');
              const wash = waypointByTipo(alt.waypoints, 'wash_station');
              const selected = idx === selectedRouteIdx;
              return (
                // div, non <button>: la mappa ora ha zoom control interattivi
                // (bottoni Leaflet) — annidare bottoni dentro un <button> è
                // HTML non valido e rompe i loro click.
                <div
                  key={idx}
                  role="button"
                  tabIndex={0}
                  onClick={() => setSelectedRouteIdx(idx)}
                  onKeyDown={(e) => { if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); setSelectedRouteIdx(idx); } }}
                  className={`text-left rounded-lg border p-2 space-y-2 transition-colors cursor-pointer ${selected ? 'border-primary ring-1 ring-primary' : 'border-border hover:bg-muted/40'}`}
                >
                  <OrderRouteMap carico={carico} scarico={scarico} garage={garage} washStation={wash} routePoints={alt.points as [number, number][] | undefined} height={260} />
                  <div className="flex items-center justify-between px-1 pb-1">
                    <span className="text-xs font-semibold">{alt.distance_km} km · {Math.floor((alt.duration_min || 0) / 60)}h{(alt.duration_min || 0) % 60}m</span>
                    {selected && <Check className="h-4 w-4 text-primary shrink-0" />}
                  </div>
                </div>
              );
            })}
            </div>
          </>
        )}
      </div>

      <div>
        <p className="text-[10px] uppercase tracking-wide text-muted-foreground font-semibold mb-3">Chi effettua il trasporto?</p>
        <Tabs value={transportMode} onValueChange={selectTransportMode} data-testid="assign-transport-mode">
          <TabsList>
            <TabsTrigger value="proprio">Mezzo proprio</TabsTrigger>
            <TabsTrigger value="terzo">Vettore terzo</TabsTrigger>
          </TabsList>
        </Tabs>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-3 mt-3">
          {transportMode === 'proprio' ? (
            <>
              <div className="space-y-1.5">
                <Label>Autista{disponibilitaLabel && <span className="text-muted-foreground font-normal"> · disponibilità {disponibilitaLabel}</span>}</Label>
                <SearchableSelect
                  value={form.autista_id}
                  onValueChange={setDriver}
                  options={assignDriverList}
                  getValue={(d) => d.id || ''}
                  getLabel={(d) => `${d.nome} ${d.cognome}`}
                  placeholder="Autista"
                  searchPlaceholder="Cerca autista..."
                  renderItem={(d) => {
                    const disponibilita = (d as DtoDriverAvailabilityResponse).disponibilita;
                    return (
                      <span className="flex items-center gap-2 min-w-0">
                        <span className={`h-2 w-2 rounded-full shrink-0 ${disponibilita === 'busy' ? 'bg-red-500' : disponibilita === 'unavailable' ? 'bg-amber-500' : 'bg-emerald-500'}`} />
                        <span className="truncate">{d.nome} {d.cognome}</span>
                        {disponibilita === 'busy' && <span className="text-[10px] text-red-600 shrink-0">occupato</span>}
                        {disponibilita === 'unavailable' && <span className="text-[10px] text-amber-600 shrink-0">{(d as DtoDriverAvailabilityResponse).motivo_indisponibilita}</span>}
                      </span>
                    );
                  }}
                />
              </div>
              <div className="space-y-1.5">
                <Label>Motrice</Label>
                <SearchableSelect
                  value={form.motrice_id}
                  onValueChange={(v) => setForm(f => ({ ...f, motrice_id: v }))}
                  options={assignMotrici}
                  getValue={(v) => v.id || ''}
                  getLabel={(v) => `${v.targa} - ${v.marca}`}
                  placeholder="Motrice"
                  searchPlaceholder="Cerca motrice..."
                  renderItem={(v) => (
                    <span className="flex items-center gap-2 min-w-0">
                      <span className={`h-2 w-2 rounded-full shrink-0 ${(v as DtoMotriceAvailabilityResponse).disponibilita === 'busy' ? 'bg-red-500' : 'bg-emerald-500'}`} />
                      <span className="truncate">{v.targa} - {v.marca}</span>
                      {(v as DtoMotriceAvailabilityResponse).disponibilita === 'busy' && <span className="text-[10px] text-red-600 shrink-0">occupato</span>}
                    </span>
                  )}
                />
              </div>
              <div className="space-y-1.5">
                <Label>Rimorchio</Label>
                <SearchableSelect
                  value={form.semirimorchio_id}
                  onValueChange={(v) => setForm(f => ({ ...f, semirimorchio_id: v }))}
                  options={assignRimorchi}
                  getValue={(v) => v.id || ''}
                  getLabel={(v) => `${v.targa} - ${v.tipo}`}
                  placeholder="Rimorchio"
                  searchPlaceholder="Cerca rimorchio..."
                  renderItem={(v) => (
                    <span className="flex items-center gap-2 min-w-0">
                      <span className={`h-2 w-2 rounded-full shrink-0 ${(v as DtoSemirimorchioAvailabilityResponse).disponibilita === 'busy' ? 'bg-red-500' : 'bg-emerald-500'}`} />
                      <span className="truncate">{v.targa} - {v.tipo}</span>
                      {(v as DtoSemirimorchioAvailabilityResponse).disponibilita === 'busy' && <span className="text-[10px] text-red-600 shrink-0">occupato</span>}
                    </span>
                  )}
                />
              </div>
            </>
          ) : (
            <div className="space-y-1.5">
              <Label>Vettore</Label>
              <SearchableSelect
                value={form.vettore_id}
                onValueChange={setVettore}
                options={carriers}
                getValue={(c) => c.id || ''}
                getLabel={(c) => c.ragione_sociale || ''}
                placeholder="Vettore terzo"
                searchPlaceholder="Cerca vettore..."
              />
            </div>
          )}
        </div>
      </div>

      <div className="flex items-center justify-end gap-2">
        {!canSubmit && (
          <span className="text-xs text-muted-foreground mr-auto">
            {!hasTransport
              ? (transportMode === 'proprio' ? 'Seleziona un autista per assegnare.' : 'Seleziona un vettore per assegnare.')
              : 'Serve un percorso valido per assegnare.'}
          </span>
        )}
        {onCancel && <Button variant="outline" onClick={onCancel}>Annulla</Button>}
        <Button onClick={handleAssign} disabled={saving || !canSubmit} data-testid="assign-order-submit">
          {saving && <Loader2 className="h-4 w-4 animate-spin mr-2" />} Assegna Viaggio
        </Button>
      </div>
    </div>
  );
}
