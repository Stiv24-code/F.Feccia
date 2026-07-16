import { useState, useEffect, useCallback, useMemo } from 'react';
import {
  DndContext, useDraggable, useDroppable, DragOverlay, PointerSensor, useSensor, useSensors,
} from '@dnd-kit/core';
import { format, addDays, parseISO, isValid, startOfDay } from 'date-fns';
import { it } from 'date-fns/locale';
import { getOrders, getVehicles, createTrip, getDrivers } from '@/lib/api';
import { Card } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Badge } from '@/components/ui/badge';
import { Skeleton } from '@/components/ui/skeleton';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog';
import { Select, SelectTrigger, SelectValue, SelectContent, SelectItem } from '@/components/ui/select';
import { toast } from 'sonner';
import { logger } from '@/lib/logger';
import { ChevronLeft, ChevronRight, Truck, Loader2, GripVertical } from 'lucide-react';
import { usePersistentFilters } from '@/hooks/use-persistent-filters';

// ════════════════════════════════════════════════════════════════════════════
// PlannerDnDPage (#40)
// Drag&drop ordini PIANIFICABILE su slot veicolo×giorno → crea viaggio.
// La PlannerPage classica resta accessibile via tab; questa è una vista
// alternativa pensata per il workflow "trascina e assegna".
// ════════════════════════════════════════════════════════════════════════════

const formatDayLabel = (date) => format(date, "EEE d MMM", { locale: it });
const toIso = (date) => format(date, 'yyyy-MM-dd');

const ORDER_DRAG_TYPE = 'order';

// ─── Card ordine draggable ──────────────────────────────────────────────────
const OrderCard = ({ order, isDragging }) => (
  <div
    className={`rounded-lg border bg-card p-2 text-xs shadow-sm cursor-grab active:cursor-grabbing select-none ${
      isDragging ? 'opacity-50' : 'hover:shadow-md'
    }`}
  >
    <div className="flex items-start gap-1.5">
      <GripVertical className="h-3 w-3 text-muted-foreground mt-0.5 shrink-0" />
      <div className="flex-1 min-w-0">
        <div className="flex items-center justify-between gap-2">
          <span className="font-mono text-[10px] font-semibold">{order.progressivo}</span>
          <Badge variant="outline" className="text-[9px] px-1 py-0">{order.tipologia}</Badge>
        </div>
        <p className="font-medium truncate mt-0.5">{order.cliente_nome}</p>
        <p className="text-[11px] text-muted-foreground truncate">
          {order.destinazione_carico_nome} → {order.destinazione_scarico_nome}
        </p>
        <p className="text-[10px] text-muted-foreground tabular-nums">
          {order.data_ritiro || '—'}
        </p>
      </div>
    </div>
  </div>
);

const DraggableOrder = ({ order }) => {
  const { attributes, listeners, setNodeRef, isDragging } = useDraggable({
    id: order.id,
    data: { type: ORDER_DRAG_TYPE, order },
  });
  return (
    <div ref={setNodeRef} {...listeners} {...attributes}>
      <OrderCard order={order} isDragging={isDragging} />
    </div>
  );
};

// ─── Slot droppable veicolo×giorno ──────────────────────────────────────────
const DroppableSlot = ({ vehicleId, day, children, hasContent }) => {
  const id = `${vehicleId}__${day}`;
  const { isOver, setNodeRef } = useDroppable({ id, data: { vehicleId, day } });
  return (
    <div
      ref={setNodeRef}
      className={`min-h-[60px] border border-dashed rounded-md p-1 transition-colors ${
        isOver ? 'bg-primary/10 border-primary' : hasContent ? 'border-border' : 'border-muted'
      }`}
    >
      {children}
    </div>
  );
};

// ─── Slot content: tab(s) trip esistente o vuoto ────────────────────────────
const SlotContent = ({ trips, ordersByTrip }) => {
  if (!trips || trips.length === 0) {
    return <p className="text-[10px] text-muted-foreground text-center py-3">Drop qui</p>;
  }
  return (
    <div className="space-y-1">
      {trips.map(t => (
        <div key={t.id} className="rounded bg-accent/40 px-1.5 py-1 text-[10px]">
          <div className="flex items-center gap-1">
            <Truck className="h-2.5 w-2.5" />
            <span className="font-mono truncate">{t.id.slice(0, 6)}</span>
            <Badge variant="outline" className="text-[9px] px-1 py-0">{t.stato}</Badge>
          </div>
          {(ordersByTrip[t.id] || []).slice(0, 2).map(o => (
            <p key={o.id} className="truncate text-muted-foreground">{o.cliente_nome}</p>
          ))}
        </div>
      ))}
    </div>
  );
};

// ════════════════════════════════════════════════════════════════════════════

export default function PlannerDnDPage() {
  const [filters, setFilters] = usePersistentFilters('planner-dnd', {
    weekStart: toIso(startOfDay(new Date())),
    vettore_id: '',
  });

  const [orders, setOrders] = useState([]);
  const [vehicles, setVehicles] = useState([]);
  const [drivers, setDrivers] = useState([]);
  const [loading, setLoading] = useState(true);
  const [activeOrder, setActiveOrder] = useState(null);

  // Dialog conferma drop
  const [dropDialog, setDropDialog] = useState(null);
  const [dropForm, setDropForm] = useState({ autista_id: '', data_arrivo: '', note: '' });
  const [creating, setCreating] = useState(false);

  const weekStartDate = useMemo(() => {
    const d = parseISO(filters.weekStart);
    return isValid(d) ? d : startOfDay(new Date());
  }, [filters.weekStart]);

  const days = useMemo(
    () => Array.from({ length: 7 }, (_, i) => addDays(weekStartDate, i)),
    [weekStartDate],
  );

  const fetchAll = useCallback(() => {
    setLoading(true);
    Promise.all([
      getOrders({ stato: 'PIANIFICABILE' }),
      getVehicles(),
      getDrivers(),
    ])
      .then(([o, v, d]) => {
        setOrders(o.data);
        setVehicles(v.data.filter(x => x.active !== false));
        setDrivers(d.data.filter(x => x.active !== false));
      })
      .catch(err => logger.error('Errore caricamento planner DnD:', err))
      .finally(() => setLoading(false));
  }, []);

  useEffect(() => { fetchAll(); }, [fetchAll]);

  // Backlog ordini visibili: nel range data_ritiro entro la settimana, oppure
  // senza data (per non perderli). Filtro nullo = mostra tutti.
  const backlog = useMemo(() => {
    const weekEnd = addDays(weekStartDate, 7);
    return orders.filter(o => {
      if (!o.data_ritiro) return true;
      const d = parseISO(o.data_ritiro);
      return isValid(d) && d >= weekStartDate && d < weekEnd;
    });
  }, [orders, weekStartDate]);

  // Trips per slot — V1: non carichiamo trip esistenti nella griglia.
  // Il drop crea sempre un nuovo trip; ammodernamenti (mostrare trip già
  // pianificati per veicolo+giorno) sono potenziale V2.
  const tripsBySlot = {};
  const ordersByTrip = {};

  const sensors = useSensors(useSensor(PointerSensor, { activationConstraint: { distance: 5 } }));

  const handleDragStart = (e) => {
    const order = e.active?.data?.current?.order;
    setActiveOrder(order || null);
  };

  const handleDragEnd = (e) => {
    setActiveOrder(null);
    const orderData = e.active?.data?.current;
    const dropData = e.over?.data?.current;
    if (!orderData?.order || !dropData) return;

    const vehicle = vehicles.find(v => v.id === dropData.vehicleId);
    if (!vehicle) return;

    setDropDialog({ order: orderData.order, vehicle, day: dropData.day });
    setDropForm({
      autista_id: '',
      data_arrivo: dropData.day,  // default: stesso giorno del carico
      note: '',
    });
  };

  const handleConfirmCreate = async () => {
    if (!dropDialog) return;
    setCreating(true);
    try {
      const driver = drivers.find(d => d.id === dropForm.autista_id);
      await createTrip({
        ordini_ids: [dropDialog.order.id],
        targa_motrice: dropDialog.vehicle.targa,
        targa_rimorchio: dropDialog.vehicle.targa_rimorchio || '',
        autista_id: dropForm.autista_id,
        autista_nome: driver ? `${driver.cognome} ${driver.nome}` : '',
        garage_id: '', garage_nome: '',
        data_partenza: dropDialog.day,
        data_arrivo: dropForm.data_arrivo,
        note: dropForm.note,
      });
      toast.success('Viaggio creato');
      setDropDialog(null);
      fetchAll();
    } catch (e) {
      toast.error(e.response?.data?.detail || 'Errore creazione viaggio');
    } finally {
      setCreating(false);
    }
  };

  const shiftWeek = (offset) => {
    setFilters({ weekStart: toIso(addDays(weekStartDate, offset * 7)) });
  };

  return (
    <div className="space-y-3" data-testid="planner-dnd-page">
      <div className="flex flex-wrap items-center gap-2 justify-between">
        <div className="flex items-center gap-2">
          <Button variant="outline" size="icon" className="h-8 w-8" onClick={() => shiftWeek(-1)} aria-label="Settimana precedente">
            <ChevronLeft className="h-4 w-4" />
          </Button>
          <Input
            type="date"
            value={filters.weekStart}
            onChange={e => setFilters({ weekStart: e.target.value })}
            className="h-8 w-40 text-xs"
            data-testid="planner-week-start"
          />
          <Button variant="outline" size="icon" className="h-8 w-8" onClick={() => shiftWeek(1)} aria-label="Settimana successiva">
            <ChevronRight className="h-4 w-4" />
          </Button>
          <span className="text-xs text-muted-foreground hidden md:inline">
            {format(weekStartDate, 'd MMM', { locale: it })} – {format(addDays(weekStartDate, 6), 'd MMM yyyy', { locale: it })}
          </span>
        </div>
        <p className="text-xs text-muted-foreground">
          Trascina un ordine sulla cella veicolo×giorno per creare un viaggio.
        </p>
      </div>

      <DndContext sensors={sensors} onDragStart={handleDragStart} onDragEnd={handleDragEnd}>
        <div className="grid grid-cols-1 lg:grid-cols-[280px_1fr] gap-3">
          {/* Backlog */}
          <Card className="rounded-xl border shadow-sm">
            <div className="px-3 py-2 border-b bg-muted/30">
              <h3 className="text-xs font-semibold uppercase tracking-wider">
                Ordini pianificabili ({backlog.length})
              </h3>
            </div>
            <div className="p-2 space-y-1.5 max-h-[68vh] overflow-y-auto">
              {loading ? (
                Array.from({ length: 4 }).map((_, i) => <Skeleton key={i} className="h-16 w-full rounded-lg" />)
              ) : backlog.length === 0 ? (
                <p className="text-xs text-muted-foreground text-center py-8">Nessun ordine in questa settimana.</p>
              ) : backlog.map(o => <DraggableOrder key={o.id} order={o} />)}
            </div>
          </Card>

          {/* Griglia */}
          <Card className="rounded-xl border shadow-sm overflow-hidden">
            <div className="overflow-x-auto">
              <table className="w-full text-xs">
                <thead>
                  <tr className="bg-muted/30">
                    <th className="text-left px-2 py-2 font-semibold uppercase tracking-wider sticky left-0 bg-muted/30 z-10 w-32">Veicolo</th>
                    {days.map(d => (
                      <th key={toIso(d)} className="text-left px-2 py-2 font-semibold uppercase tracking-wider min-w-[120px]">
                        {formatDayLabel(d)}
                      </th>
                    ))}
                  </tr>
                </thead>
                <tbody>
                  {loading ? (
                    Array.from({ length: 4 }).map((_, i) => (
                      <tr key={`s${i}`}>
                        <td className="px-2 py-2"><Skeleton className="h-6 w-24" /></td>
                        {days.map((_, j) => <td key={j} className="p-1"><Skeleton className="h-14 w-full" /></td>)}
                      </tr>
                    ))
                  ) : vehicles.length === 0 ? (
                    <tr><td colSpan={8} className="text-center py-8 text-muted-foreground">Nessun veicolo attivo</td></tr>
                  ) : vehicles.map(v => (
                    <tr key={v.id} className="border-t">
                      <td className="px-2 py-2 sticky left-0 bg-card z-10 align-top">
                        <div className="flex items-center gap-1.5">
                          <Truck className="h-3 w-3 shrink-0" />
                          <span className="font-mono font-semibold truncate">{v.targa}</span>
                        </div>
                        <p className="text-[10px] text-muted-foreground truncate">{v.tipo_veicolo}</p>
                      </td>
                      {days.map(d => {
                        const slotKey = `${v.id}__${toIso(d)}`;
                        const slotTrips = tripsBySlot[slotKey] || [];
                        return (
                          <td key={toIso(d)} className="p-1 align-top">
                            <DroppableSlot vehicleId={v.id} day={toIso(d)} hasContent={slotTrips.length > 0}>
                              <SlotContent trips={slotTrips} ordersByTrip={ordersByTrip} />
                            </DroppableSlot>
                          </td>
                        );
                      })}
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </Card>
        </div>

        <DragOverlay>
          {activeOrder ? <OrderCard order={activeOrder} /> : null}
        </DragOverlay>
      </DndContext>

      {/* Dialog conferma creazione viaggio */}
      <Dialog open={!!dropDialog} onOpenChange={(o) => !o && setDropDialog(null)}>
        <DialogContent className="max-w-md">
          <DialogHeader><DialogTitle>Crea viaggio</DialogTitle></DialogHeader>
          {dropDialog && (
            <div className="space-y-3 text-sm">
              <div className="rounded-lg border bg-muted/40 p-3">
                <p className="text-[11px] text-muted-foreground mb-1">Ordine</p>
                <p className="font-medium">{dropDialog.order.cliente_nome}</p>
                <p className="text-xs text-muted-foreground truncate">
                  {dropDialog.order.destinazione_carico_nome} → {dropDialog.order.destinazione_scarico_nome}
                </p>
              </div>
              <div className="grid grid-cols-2 gap-2 text-xs">
                <div>
                  <Label>Veicolo</Label>
                  <p className="font-mono py-2">{dropDialog.vehicle.targa}</p>
                </div>
                <div>
                  <Label>Partenza</Label>
                  <p className="font-mono py-2">{dropDialog.day}</p>
                </div>
              </div>
              <div>
                <Label>Autista *</Label>
                <Select value={dropForm.autista_id} onValueChange={v => setDropForm({...dropForm, autista_id: v})}>
                  <SelectTrigger><SelectValue placeholder="Seleziona autista" /></SelectTrigger>
                  <SelectContent>
                    {drivers.map(d => <SelectItem key={d.id} value={d.id}>{d.cognome} {d.nome}</SelectItem>)}
                  </SelectContent>
                </Select>
              </div>
              <div>
                <Label>Data arrivo prevista</Label>
                <Input type="date" value={dropForm.data_arrivo} onChange={e => setDropForm({...dropForm, data_arrivo: e.target.value})} />
              </div>
            </div>
          )}
          <DialogFooter>
            <Button variant="outline" onClick={() => setDropDialog(null)}>Annulla</Button>
            <Button onClick={handleConfirmCreate} disabled={creating || !dropForm.autista_id}>
              {creating ? <><Loader2 className="h-4 w-4 mr-2 animate-spin" /> Creazione…</> : 'Crea viaggio'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
