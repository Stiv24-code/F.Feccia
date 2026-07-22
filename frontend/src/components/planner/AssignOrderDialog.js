import { useState, useEffect, useMemo } from 'react';
import { assignOrder, getVehicles, getDrivers, getCarriers, getVehicleAvailability, getDriverAvailability } from '@/lib/api';
import { Button } from '@/components/ui/button';
import { Label } from '@/components/ui/label';
import { Select, SelectTrigger, SelectValue, SelectContent, SelectItem } from '@/components/ui/select';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog';
import { toast } from 'sonner';
import { Loader2 } from 'lucide-react';
import { logger } from '@/lib/logger';

const EMPTY_FORM = { targa_motrice: '', targa_rimorchio: '', autista_id: '', autista_nome: '', vettore_id: '', vettore_nome: '' };

// Dialog "Assegna Ordine" — riutilizzato da PlannerPage (calendario/lista) e
// da OrderDetailPage, cosi la logica di disponibilita mezzi/autisti vive in
// un solo posto.
export default function AssignOrderDialog({ open, onOpenChange, order, onAssigned }) {
  const [form, setForm] = useState(EMPTY_FORM);
  const [saving, setSaving] = useState(false);
  const [vehicles, setVehicles] = useState([]);
  const [drivers, setDrivers] = useState([]);
  const [carriers, setCarriers] = useState([]);
  const [availVehicles, setAvailVehicles] = useState([]);
  const [availDrivers, setAvailDrivers] = useState([]);

  useEffect(() => {
    if (!open || !order) return;
    setForm(EMPTY_FORM);
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
  }, [open, order]);

  const setDriver = (id) => {
    const allD = availDrivers.length > 0 ? availDrivers : drivers;
    const d = allD.find(x => x.id === id);
    setForm(f => ({ ...f, autista_id: id, autista_nome: d ? `${d.nome} ${d.cognome}` : '' }));
  };
  const setVettore = (id) => {
    const c = carriers.find(x => x.id === id);
    setForm(f => ({ ...f, vettore_id: id, vettore_nome: c?.ragione_sociale || '' }));
  };
  const assignMotrici = useMemo(() => (availVehicles.length > 0 ? availVehicles : vehicles).filter(v => v.tipo_veicolo === 'motrice'), [availVehicles, vehicles]);
  const assignRimorchi = useMemo(() => (availVehicles.length > 0 ? availVehicles : vehicles).filter(v => v.tipo_veicolo !== 'motrice'), [availVehicles, vehicles]);
  const assignDriverList = useMemo(() => availDrivers.length > 0 ? availDrivers : drivers, [availDrivers, drivers]);

  const handleAssign = async () => {
    setSaving(true);
    try {
      await assignOrder(order.id, form);
      toast.success(`Ordine ${order.progressivo} assegnato`);
      onOpenChange(false);
      if (onAssigned) onAssigned();
    } catch (e) { toast.error(e.response?.data?.detail || 'Errore'); } finally { setSaving(false); }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-lg">
        <DialogHeader>
          <DialogTitle style={{ fontFamily: "'Space Grotesk', sans-serif" }}>Assegna Ordine {order?.progressivo}</DialogTitle>
        </DialogHeader>
        <div className="space-y-4">
          <div className="p-3 rounded-lg bg-muted/50 text-sm">
            <p><strong>{order?.destinazione_carico_nome}</strong> → <strong>{order?.destinazione_scarico_nome}</strong></p>
            <p className="text-muted-foreground">{order?.cliente_nome} • {order?.data_ritiro}</p>
          </div>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
            <div className="space-y-1.5">
              <Label>Targa Motrice</Label>
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
              <Label>Targa Rimorchio</Label>
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
            <div className="space-y-1.5">
              <Label>Autista</Label>
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
              <Label>Vettore (opzionale)</Label>
              <Select value={form.vettore_id} onValueChange={setVettore}>
                <SelectTrigger><SelectValue placeholder="Vettore terzo" /></SelectTrigger>
                <SelectContent>{carriers.map(c => <SelectItem key={c.id} value={c.id}>{c.ragione_sociale}</SelectItem>)}</SelectContent>
              </Select>
            </div>
          </div>
        </div>
        <DialogFooter className="gap-2">
          <Button variant="outline" onClick={() => onOpenChange(false)}>Annulla</Button>
          <Button onClick={handleAssign} disabled={saving} data-testid="planner-assign-submit">
            {saving && <Loader2 className="h-4 w-4 animate-spin mr-2" />} Assegna Viaggio
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
