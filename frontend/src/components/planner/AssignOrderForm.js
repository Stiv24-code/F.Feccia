import { useState, useEffect, useMemo } from 'react';
import { assignOrder, getVehicles, getDrivers, getCarriers, getVehicleAvailability, getDriverAvailability } from '@/lib/api';
import { useGetGaragesQuery, useGetWashStationsQuery } from '@/store/api/appApi';
import { Button } from '@/components/ui/button';
import { Label } from '@/components/ui/label';
import { Select, SelectTrigger, SelectValue, SelectContent, SelectItem } from '@/components/ui/select';
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { toast } from 'sonner';
import { Loader2, Warehouse, Droplets } from 'lucide-react';
import { logger } from '@/lib/logger';

const EMPTY_FORM = {
  garage_id: '',
  targa_motrice: '', targa_rimorchio: '', autista_id: '',
  vettore_id: '',
  wash_station_id: '',
};

const formatDataBreve = (iso) => {
  if (!iso) return '';
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return '';
  return d.toLocaleDateString('it-IT', { day: 'numeric', month: 'short' });
};

// Form "Assegna trasporto" — corpo condiviso tra AssignOrderDialog (modale,
// usata da PlannerPage) e OrderDetailPage (inline, ordine in PIANIFICABILE),
// cosi' la logica di disponibilita' mezzi/autisti e le select punto di
// partenza/lavaggio vivono in un solo posto invece di essere duplicate.
export default function AssignOrderForm({ order, onAssigned, onCancel }) {
  const [form, setForm] = useState(EMPTY_FORM);
  const [transportMode, setTransportMode] = useState('proprio');
  const [saving, setSaving] = useState(false);
  const [vehicles, setVehicles] = useState([]);
  const [drivers, setDrivers] = useState([]);
  const [carriers, setCarriers] = useState([]);
  const [availVehicles, setAvailVehicles] = useState([]);
  const [availDrivers, setAvailDrivers] = useState([]);

  const { data: garages = [] } = useGetGaragesQuery();
  const { data: washStations = [] } = useGetWashStationsQuery();

  useEffect(() => {
    if (!order) return;
    setForm(EMPTY_FORM);
    setTransportMode('proprio');
    setAvailVehicles([]);
    setAvailDrivers([]);
    Promise.all([getVehicles(), getDrivers(), getCarriers()]).then(([v, d, c]) => {
      setVehicles(v.data); setDrivers(d.data); setCarriers(c.data);
    }).catch(err => logger.error('Errore caricamento lookup assegna:', err));
    const da = order.data_ritiro || '';
    const a = order.data_consegna || order.data_ritiro || '';
    if (da && a) {
      Promise.all([getVehicleAvailability(da, a), getDriverAvailability(da, a)]).then(([vRes, dRes]) => {
        setAvailVehicles(vRes.data); setAvailDrivers(dRes.data);
      }).catch(err => logger.error('Disponibilità fetch error:', err));
    }
  }, [order]);

  const setGarage = (id) => setForm(f => ({ ...f, garage_id: id }));
  const setWashStation = (id) => setForm(f => ({ ...f, wash_station_id: id }));
  const setDriver = (id) => setForm(f => ({ ...f, autista_id: id }));
  const setVettore = (id) => setForm(f => ({ ...f, vettore_id: id }));
  const selectTransportMode = (mode) => {
    setTransportMode(mode);
    if (mode === 'proprio') {
      setForm(f => ({ ...f, vettore_id: '' }));
    } else {
      setForm(f => ({ ...f, targa_motrice: '', targa_rimorchio: '', autista_id: '' }));
    }
  };

  const assignMotrici = useMemo(() => (availVehicles.length > 0 ? availVehicles : vehicles).filter(v => v.tipo_veicolo === 'motrice'), [availVehicles, vehicles]);
  const assignRimorchi = useMemo(() => (availVehicles.length > 0 ? availVehicles : vehicles).filter(v => v.tipo_veicolo !== 'motrice'), [availVehicles, vehicles]);
  const assignDriverList = useMemo(() => availDrivers.length > 0 ? availDrivers : drivers, [availDrivers, drivers]);
  const disponibilitaLabel = useMemo(() => formatDataBreve(order?.data_ritiro), [order]);

  const handleAssign = async () => {
    setSaving(true);
    try {
      await assignOrder(order.id, form);
      toast.success(`Ordine ${order.progressivo} assegnato`);
      if (onAssigned) onAssigned();
    } catch (e) { toast.error(e.response?.data?.detail || 'Errore'); } finally { setSaving(false); }
  };

  return (
    <div className="space-y-5">
      <div>
        <p className="text-[10px] uppercase tracking-wide text-muted-foreground font-semibold mb-3">Dettagli trasporto</p>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
          <div className="space-y-1.5">
            <Label>Punto di partenza</Label>
            <Select value={form.garage_id} onValueChange={setGarage}>
              <SelectTrigger className="border-dashed">
                <span className="flex items-center gap-2 overflow-hidden">
                  <Warehouse className="h-4 w-4 text-muted-foreground shrink-0" />
                  <SelectValue placeholder="Seleziona punto di partenza..." />
                </span>
              </SelectTrigger>
              <SelectContent>{garages.map(g => <SelectItem key={g.id} value={g.id}>{g.nome}</SelectItem>)}</SelectContent>
            </Select>
          </div>
          <div className="space-y-1.5">
            <Label>Punto di lavaggio — dopo lo scarico</Label>
            <Select value={form.wash_station_id} onValueChange={setWashStation}>
              <SelectTrigger className="border-dashed">
                <span className="flex items-center gap-2 overflow-hidden">
                  <Droplets className="h-4 w-4 text-muted-foreground shrink-0" />
                  <SelectValue placeholder="Seleziona punto di lavaggio..." />
                </span>
              </SelectTrigger>
              <SelectContent>{washStations.map(w => <SelectItem key={w.id} value={w.id}>{w.nome}</SelectItem>)}</SelectContent>
            </Select>
          </div>
        </div>
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
                <Select value={form.autista_id} onValueChange={setDriver}>
                  <SelectTrigger><SelectValue placeholder="Autista" /></SelectTrigger>
                  <SelectContent>
                    {assignDriverList.map(d => (
                      <SelectItem key={d.id} value={d.id}>
                        <span className="flex items-center gap-2">
                          <span className={`h-2 w-2 rounded-full shrink-0 ${d.disponibilita === 'busy' ? 'bg-red-500' : d.disponibilita === 'unavailable' ? 'bg-amber-500' : 'bg-emerald-500'}`} />
                          {d.nome} {d.cognome}
                          {d.disponibilita === 'busy' && <span className="text-[10px] text-red-600">occupato</span>}
                          {d.disponibilita === 'unavailable' && <span className="text-[10px] text-amber-600">{d.motivo_indisponibilita}</span>}
                        </span>
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
              <div className="space-y-1.5">
                <Label>Motrice</Label>
                <Select value={form.targa_motrice} onValueChange={v => setForm(f => ({ ...f, targa_motrice: v }))}>
                  <SelectTrigger><SelectValue placeholder="Motrice" /></SelectTrigger>
                  <SelectContent>
                    {assignMotrici.map(v => (
                      <SelectItem key={v.id} value={v.targa}>
                        <span className="flex items-center gap-2">
                          <span className={`h-2 w-2 rounded-full shrink-0 ${v.disponibilita === 'busy' ? 'bg-red-500' : 'bg-emerald-500'}`} />
                          {v.targa} - {v.marca}
                          {v.disponibilita === 'busy' && <span className="text-[10px] text-red-600">occupato</span>}
                        </span>
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
              <div className="space-y-1.5">
                <Label>Rimorchio</Label>
                <Select value={form.targa_rimorchio} onValueChange={v => setForm(f => ({ ...f, targa_rimorchio: v }))}>
                  <SelectTrigger><SelectValue placeholder="Rimorchio" /></SelectTrigger>
                  <SelectContent>
                    {assignRimorchi.map(v => (
                      <SelectItem key={v.id} value={v.targa}>
                        <span className="flex items-center gap-2">
                          <span className={`h-2 w-2 rounded-full shrink-0 ${v.disponibilita === 'busy' ? 'bg-red-500' : 'bg-emerald-500'}`} />
                          {v.targa} - {v.tipo_veicolo}
                          {v.disponibilita === 'busy' && <span className="text-[10px] text-red-600">occupato</span>}
                        </span>
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            </>
          ) : (
            <div className="space-y-1.5">
              <Label>Vettore</Label>
              <Select value={form.vettore_id} onValueChange={setVettore}>
                <SelectTrigger><SelectValue placeholder="Vettore terzo" /></SelectTrigger>
                <SelectContent>{carriers.map(c => <SelectItem key={c.id} value={c.id}>{c.ragione_sociale}</SelectItem>)}</SelectContent>
              </Select>
            </div>
          )}
        </div>
      </div>

      <div className="flex justify-end gap-2">
        {onCancel && <Button variant="outline" onClick={onCancel}>Annulla</Button>}
        <Button onClick={handleAssign} disabled={saving} data-testid="assign-order-submit">
          {saving && <Loader2 className="h-4 w-4 animate-spin mr-2" />} Assegna Viaggio
        </Button>
      </div>
    </div>
  );
}
