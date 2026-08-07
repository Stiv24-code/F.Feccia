import { useState, useEffect, useCallback, useMemo } from 'react';
import { useNavigate } from 'react-router-dom';
import { getOrders, startOrder, closeOrder, discardOrder } from '@/lib/api';
import { getApiErrorMessage } from '@/lib/apiError';
import type { DtoOrderResponse } from '@/api/data-contracts';
import { format, parseISO, isValid, startOfWeek, endOfWeek, startOfMonth, endOfMonth } from 'date-fns';
import { it } from 'date-fns/locale';
import { Card } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Select, SelectTrigger, SelectValue, SelectContent, SelectItem } from '@/components/ui/select';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Badge } from '@/components/ui/badge';
import { Skeleton } from '@/components/ui/skeleton';
import { StatusBadge } from '@/components/shared/StatusBadge';
import PlannerCalendar from '@/components/planner/PlannerCalendar';
import AssignOrderDialog from '@/components/planner/AssignOrderDialog';
import { toast } from 'sonner';
import { Search, CalendarRange, Truck, CheckCircle, PlayCircle, Ban, List as ListIcon, CalendarDays } from 'lucide-react';
import { logger } from '@/lib/logger';

type OrderStatus = 'PIANIFICABILE' | 'PIANIFICATO' | 'VIAGGIO' | 'CHIUSO' | 'SCARTATO';
type PlannerView = 'calendar' | 'list';

// Nome completo autista, o stringa vuota se non ancora assegnato — fonte
// unica per cella tabella, filtro dropdown e ricerca libera.
const driverFullName = (o: DtoOrderResponse) => o.autista ? `${o.autista.nome} ${o.autista.cognome}` : '';

interface OrderGridProps {
  orders: DtoOrderResponse[];
  loading: boolean;
  onAssign: (order: DtoOrderResponse) => void;
  onStart: (order: DtoOrderResponse) => void;
  onClose: (order: DtoOrderResponse) => void;
  onDiscard: (order: DtoOrderResponse) => void;
  onOpenDetail: (order: DtoOrderResponse) => void;
  emptyMsg?: string;
}

// ============================
// Componente griglia riutilizzabile
// ============================
const OrderGrid = ({ orders, loading, onAssign, onStart, onClose, onDiscard, onOpenDetail, emptyMsg }: OrderGridProps) => (
  <Card className="rounded-xl border shadow-sm">
    <div className="overflow-x-auto">
      <Table className="text-xs md:text-sm">
        <TableHeader>
          <TableRow>
            <TableHead className="py-2 text-xs">Prog.</TableHead>
            <TableHead className="py-2 text-xs">Data Carico</TableHead>
            <TableHead className="py-2 text-xs">Data Scarico</TableHead>
            <TableHead className="py-2 text-xs">Tipo</TableHead>
            <TableHead className="py-2 text-xs">Partenza</TableHead>
            <TableHead className="py-2 text-xs">Destino</TableHead>
            <TableHead className="py-2 text-xs">Cliente</TableHead>
            <TableHead className="py-2 text-xs">Mezzo</TableHead>
            <TableHead className="py-2 text-xs">Autista</TableHead>
            <TableHead className="py-2 text-xs">Stato</TableHead>
            <TableHead className="py-2 text-xs">Azioni</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {loading ? Array.from({ length: 3 }).map((_, i) => (
            <TableRow key={`skel-row-${i}`}>{Array.from({ length: 11 }).map((_, j) => <TableCell key={`skel-col-${j}`} className="py-2"><Skeleton className="h-4 w-full" /></TableCell>)}</TableRow>
          )) : orders.length === 0 ? (
            <TableRow><TableCell colSpan={11} className="text-center py-6 text-muted-foreground text-xs">{emptyMsg || 'Nessun ordine'}</TableCell></TableRow>
          ) : orders.map(o => (
            <TableRow key={o.id} className="hover:bg-muted/60 cursor-pointer" onClick={() => onOpenDetail(o)} data-testid="planner-row">
              <TableCell className="py-2 font-mono font-medium">{o.progressivo}</TableCell>
              <TableCell className="py-2 whitespace-nowrap">{o.data_ritiro} {o.ora_ritiro_da}</TableCell>
              <TableCell className="py-2 whitespace-nowrap">{o.data_consegna} {o.ora_consegna_da}</TableCell>
              <TableCell className="py-2"><Badge variant="outline" className="text-[10px]">{o.tipologia}</Badge></TableCell>
              <TableCell className="py-2 max-w-[100px] truncate">{o.destinazione_carico?.nome}</TableCell>
              <TableCell className="py-2 max-w-[100px] truncate">{o.destinazione_scarico?.nome}</TableCell>
              <TableCell className="py-2 max-w-[120px] truncate">{o.cliente?.ragione_sociale}</TableCell>
              <TableCell className="py-2 font-mono">{o.motrice?.targa || '—'}</TableCell>
              <TableCell className="py-2">{driverFullName(o) || '—'}</TableCell>
              <TableCell className="py-2"><StatusBadge stato={o.stato} /></TableCell>
              <TableCell className="py-2" onClick={e => e.stopPropagation()}>
                <div className="flex items-center gap-1">
                  {o.stato === 'PIANIFICABILE' && (
                    <>
                      <Button size="sm" variant="outline" className="h-7 text-xs gap-1" onClick={() => onAssign(o)} data-testid="planner-assign-button">
                        <Truck className="h-3 w-3" /> Assegna
                      </Button>
                      <Button size="icon" variant="ghost" className="h-7 w-7 text-destructive" onClick={() => onDiscard(o)} title="Scarta ordine" data-testid="planner-discard-button">
                        <Ban className="h-3.5 w-3.5" />
                      </Button>
                    </>
                  )}
                  {o.stato === 'PIANIFICATO' && (
                    o.viaggio_id ? (
                      <span className="text-xs text-muted-foreground">Nel viaggio {o.viaggio_id.slice(0, 8)}</span>
                    ) : (
                      <>
                        <Button size="sm" variant="outline" className="h-7 text-xs gap-1" onClick={() => onStart(o)} data-testid="planner-start-button">
                          <PlayCircle className="h-3 w-3" /> Avvia viaggio
                        </Button>
                        <Button size="icon" variant="ghost" className="h-7 w-7 text-destructive" onClick={() => onDiscard(o)} title="Scarta ordine" data-testid="planner-discard-button">
                          <Ban className="h-3.5 w-3.5" />
                        </Button>
                      </>
                    )
                  )}
                  {o.stato === 'VIAGGIO' && (
                    <Button size="sm" variant="outline" className="h-7 text-xs gap-1" onClick={() => onClose(o)} data-testid="planner-close-button">
                      <CheckCircle className="h-3 w-3" /> Chiudi
                    </Button>
                  )}
                </div>
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  </Card>
);

const STATUS_CHIPS: { key: OrderStatus | null; label: string; className: string }[] = [
  { key: null, label: 'Tutti', className: 'border-muted-foreground/30 bg-background text-foreground' },
  { key: 'PIANIFICABILE', label: 'Da pianificare', className: 'status-order-red' },
  { key: 'PIANIFICATO', label: 'Pianificati', className: 'status-order-yellow' },
  { key: 'VIAGGIO', label: 'In viaggio', className: 'status-order-blue' },
  { key: 'CHIUSO', label: 'Consegnati', className: 'status-order-green' },
  { key: 'SCARTATO', label: 'Scartati', className: 'status-order-gray' },
];

// ============================
// Pagina Planner
// ============================
export default function PlannerPage() {
  const navigate = useNavigate();
  const [orders, setOrders] = useState<DtoOrderResponse[]>([]);
  const [loading, setLoading] = useState(true);
  const [assignDialogOpen, setAssignDialogOpen] = useState(false);
  const [selectedOrder, setSelectedOrder] = useState<DtoOrderResponse | null>(null);

  const [search, setSearch] = useState('');
  const [driverFilter, setDriverFilter] = useState('');
  const [statusFilter, setStatusFilter] = useState<OrderStatus | null>(null);
  const [dateFrom, setDateFrom] = useState('');
  const [dateTo, setDateTo] = useState('');
  const [dateOpen, setDateOpen] = useState(false);
  const [view, setView] = useState<PlannerView>('calendar');

  const fetchOrders = useCallback(() => {
    setLoading(true);
    const params: { data_da?: string; data_a?: string } = {};
    if (dateFrom) params.data_da = dateFrom;
    if (dateTo) params.data_a = dateTo;
    getOrders(params).then((r: { data: DtoOrderResponse[] }) => setOrders(r.data)).catch((err: unknown) => logger.error('Errore caricamento ordini planner:', err)).finally(() => setLoading(false));
  }, [dateFrom, dateTo]);

  useEffect(() => { fetchOrders(); }, [fetchOrders]);

  const openDetail = (order: DtoOrderResponse) => navigate(`/planner/ordini/${order.id}`);

  // --- Assegna ---
  const openAssign = (order: DtoOrderResponse) => {
    setSelectedOrder(order);
    setAssignDialogOpen(true);
  };
  const handleClose = async (order: DtoOrderResponse) => {
    if (!order.id || !window.confirm(`Chiudere l'ordine ${order.progressivo}?`)) return;
    try { await closeOrder(order.id); toast.success(`Ordine ${order.progressivo} chiuso`); fetchOrders(); }
    catch (e) { toast.error(getApiErrorMessage(e) || 'Errore'); }
  };
  const handleStart = async (order: DtoOrderResponse) => {
    if (!order.id || !window.confirm(`Avviare il viaggio per l'ordine ${order.progressivo}?`)) return;
    try { await startOrder(order.id); toast.success(`Ordine ${order.progressivo} avviato`); fetchOrders(); }
    catch (e) { toast.error(getApiErrorMessage(e) || 'Errore'); }
  };
  const handleDiscard = async (order: DtoOrderResponse) => {
    if (!order.id || !window.confirm(`Scartare l'ordine ${order.progressivo}? L'operazione non è reversibile.`)) return;
    try { await discardOrder(order.id); toast.success(`Ordine ${order.progressivo} scartato`); fetchOrders(); }
    catch (e) { toast.error(getApiErrorMessage(e) || 'Errore'); }
  };

  // --- Filtri toolbar: ricerca + autista + stato (il periodo è già filtrato server-side) ---
  const driverOptions = useMemo(() => {
    const names = new Set(orders.map(driverFullName).filter(Boolean));
    return Array.from(names).sort((a, b) => a.localeCompare(b));
  }, [orders]);

  const searchAndDriverFiltered = useMemo(() => {
    const q = search.trim().toLowerCase();
    return orders.filter(o => {
      const autistaNome = driverFullName(o);
      if (driverFilter && autistaNome !== driverFilter) return false;
      if (q) {
        const hay = `${o.progressivo || ''} ${o.cliente?.ragione_sociale || ''} ${o.destinazione_carico?.nome || ''} ${o.destinazione_scarico?.nome || ''} ${autistaNome} ${o.motrice?.targa || ''}`.toLowerCase();
        if (!hay.includes(q)) return false;
      }
      return true;
    });
  }, [orders, driverFilter, search]);

  const chipCounts = useMemo(() => {
    const counts: Record<string, number> = { all: searchAndDriverFiltered.length };
    (['PIANIFICABILE', 'PIANIFICATO', 'VIAGGIO', 'CHIUSO', 'SCARTATO'] as OrderStatus[]).forEach(s => {
      counts[s] = searchAndDriverFiltered.filter(o => o.stato === s).length;
    });
    return counts;
  }, [searchAndDriverFiltered]);

  const filtered = useMemo(
    () => statusFilter ? searchAndDriverFiltered.filter(o => o.stato === statusFilter) : searchAndDriverFiltered,
    [searchAndDriverFiltered, statusFilter]
  );

  // --- Chip periodo (Partenza: dal → al) ---
  const dateLabel = useMemo(() => {
    if (!dateFrom && !dateTo) return 'tutte le date';
    const from = dateFrom && isValid(parseISO(dateFrom)) ? format(parseISO(dateFrom), 'd MMM', { locale: it }) : '…';
    const to = dateTo && isValid(parseISO(dateTo)) ? format(parseISO(dateTo), 'd MMM', { locale: it }) : '…';
    return `${from} → ${to}`;
  }, [dateFrom, dateTo]);

  const datePresets = useMemo(() => {
    const today = new Date();
    const iso = (d: Date) => format(d, 'yyyy-MM-dd');
    return [
      { label: 'Oggi', from: iso(today), to: iso(today) },
      { label: 'Questa settimana', from: iso(startOfWeek(today, { weekStartsOn: 1 })), to: iso(endOfWeek(today, { weekStartsOn: 1 })) },
      { label: 'Questo mese', from: iso(startOfMonth(today)), to: iso(endOfMonth(today)) },
    ];
  }, []);

  return (
    <div className="space-y-3" data-testid="planner-page">
      {/* Toolbar */}
      <div className="flex flex-col gap-2 lg:flex-row lg:items-center lg:justify-between flex-wrap">
        <div className="flex items-center gap-2 flex-wrap">
          <div className="relative">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-muted-foreground" />
            <Input
              value={search} onChange={e => setSearch(e.target.value)}
              placeholder="Cerca ordine, cliente, autista, mezzo..."
              className="pl-9 h-9 w-64 text-sm" data-testid="planner-search-input"
            />
          </div>
          <Select value={driverFilter || 'all'} onValueChange={v => setDriverFilter(v === 'all' ? '' : v)}>
            <SelectTrigger className="h-9 w-44 text-sm" data-testid="planner-driver-filter"><SelectValue placeholder="Tutti gli autisti" /></SelectTrigger>
            <SelectContent>
              <SelectItem value="all">Tutti gli autisti</SelectItem>
              {driverOptions.map(name => <SelectItem key={name} value={name}>{name}</SelectItem>)}
            </SelectContent>
          </Select>
          {view === 'list' && (
            <div className="relative">
              <button
                type="button" onClick={() => setDateOpen(o => !o)} data-testid="planner-date-chip"
                className="flex items-center gap-1.5 text-xs border rounded-md px-3 h-9 bg-background hover:bg-muted/50"
              >
                <CalendarRange className="h-3.5 w-3.5 text-muted-foreground" />
                <span className="text-muted-foreground">Partenza:</span>
                <span className="font-medium">{dateLabel}</span>
              </button>
              {dateOpen && (
                <>
                  <button type="button" aria-label="Chiudi selezione periodo" className="fixed inset-0 z-10 cursor-default bg-transparent border-0 p-0" onClick={() => setDateOpen(false)} />
                  <div className="absolute z-20 mt-1 w-72 rounded-lg border bg-popover text-popover-foreground p-3 shadow-md">
                    <div className="grid grid-cols-2 gap-2">
                      <div className="space-y-1">
                        <Label className="text-[10px] text-muted-foreground">Dal</Label>
                        <Input type="date" value={dateFrom} onChange={e => setDateFrom(e.target.value)} className="h-8 text-xs text-center" data-testid="planner-date-from" />
                      </div>
                      <div className="space-y-1">
                        <Label className="text-[10px] text-muted-foreground">Al</Label>
                        <Input type="date" value={dateTo} onChange={e => setDateTo(e.target.value)} className="h-8 text-xs text-center" data-testid="planner-date-to" />
                      </div>
                    </div>
                    <div className="flex flex-wrap gap-1.5 mt-2">
                      {datePresets.map(p => (
                        <button
                          key={p.label} type="button"
                          onClick={() => { setDateFrom(p.from); setDateTo(p.to); }}
                          className="text-[10px] font-medium px-2.5 py-1 rounded-full border hover:border-primary hover:text-primary"
                        >
                          {p.label}
                        </button>
                      ))}
                    </div>
                    <div className="flex justify-between items-center mt-3 pt-2 border-t">
                      <button type="button" onClick={() => { setDateFrom(''); setDateTo(''); }} className="text-xs text-muted-foreground hover:text-destructive">Azzera</button>
                      <Button size="sm" className="h-7 text-xs" onClick={() => setDateOpen(false)}>Applica</Button>
                    </div>
                  </div>
                </>
              )}
            </div>
          )}
        </div>

        <div className="flex items-center gap-2 w-full lg:w-auto">
          <div className="flex items-center gap-1.5 flex-wrap flex-1 min-w-0">
            {STATUS_CHIPS.map(c => {
              const active = statusFilter === c.key;
              const count = c.key ? chipCounts[c.key] : chipCounts.all;
              return (
                <button
                  key={c.key ?? 'all'} type="button" onClick={() => setStatusFilter(c.key)}
                  data-testid={`planner-chip-${c.key ?? 'all'}`}
                  className={`text-xs font-semibold rounded-full border px-3 py-1 transition ${c.className} ${active ? 'ring-2 ring-offset-1 ring-primary' : 'opacity-70 hover:opacity-100'}`}
                >
                  {c.label} {count}
                </button>
              );
            })}
          </div>
          <div className="flex items-center gap-0.5 border rounded-md p-0.5 shrink-0 ml-auto">
            <Button variant={view === 'calendar' ? 'default' : 'ghost'} size="sm" className="h-7 px-2 text-xs gap-1" onClick={() => setView('calendar')} data-testid="planner-view-calendar">
              <CalendarDays className="h-3.5 w-3.5" />
            </Button>
            <Button variant={view === 'list' ? 'default' : 'ghost'} size="sm" className="h-7 px-2 text-xs gap-1" onClick={() => setView('list')} data-testid="planner-view-list">
              <ListIcon className="h-3.5 w-3.5" />
            </Button>
          </div>
        </div>
      </div>

      {view === 'calendar' ? (
        <PlannerCalendar orders={filtered} onOpen={openDetail} />
      ) : (
        <OrderGrid orders={filtered} loading={loading} onAssign={openAssign} onStart={handleStart} onClose={handleClose} onDiscard={handleDiscard} onOpenDetail={openDetail} emptyMsg="Nessun ordine trovato" />
      )}

      <AssignOrderDialog open={assignDialogOpen} onOpenChange={setAssignDialogOpen} order={selectedOrder} onAssigned={fetchOrders} />
    </div>
  );
}
