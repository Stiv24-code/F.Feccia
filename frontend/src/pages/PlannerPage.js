import { useState, useEffect, useCallback, useMemo } from 'react';
import { getOrders, assignOrder, startOrder, closeOrder, discardOrder, getVehicles, getDrivers, getCarriers, getVehicleAvailability, getDriverAvailability, getDriverUnavailability, createDriverUnavailability, deleteDriverUnavailability } from '@/lib/api';
import { Card } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Select, SelectTrigger, SelectValue, SelectContent, SelectItem } from '@/components/ui/select';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog';
import { Badge } from '@/components/ui/badge';
import { Skeleton } from '@/components/ui/skeleton';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { StatusBadge } from '@/components/shared/StatusBadge';
import PlannerCalendar from '@/components/planner/PlannerCalendar';
import { toast } from 'sonner';
import { CalendarRange, Truck, CheckCircle, Loader2, Users, AlertTriangle, Plus, Trash2, PlayCircle, Ban, List as ListIcon, CalendarDays } from 'lucide-react';
import { logger } from '@/lib/logger';

// ============================
// Componente griglia riutilizzabile
// ============================
const OrderGrid = ({ orders, loading, onAssign, onStart, onClose, onDiscard, title, emptyMsg }) => (
  <Card className="rounded-xl border shadow-sm">
    {title && (
      <div className="px-4 py-2.5 border-b bg-muted/30">
        <h3 className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">{title}</h3>
      </div>
    )}
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
            <TableRow key={o.id} className="hover:bg-muted/60">
              <TableCell className="py-2 font-mono font-medium">{o.progressivo}</TableCell>
              <TableCell className="py-2 whitespace-nowrap">{o.data_ritiro} {o.ora_ritiro_da}</TableCell>
              <TableCell className="py-2 whitespace-nowrap">{o.data_consegna} {o.ora_consegna_da}</TableCell>
              <TableCell className="py-2"><Badge variant="outline" className="text-[10px]">{o.tipologia}</Badge></TableCell>
              <TableCell className="py-2 max-w-[100px] truncate">{o.destinazione_carico_nome}</TableCell>
              <TableCell className="py-2 max-w-[100px] truncate">{o.destinazione_scarico_nome}</TableCell>
              <TableCell className="py-2 max-w-[120px] truncate">{o.cliente_nome}</TableCell>
              <TableCell className="py-2 font-mono">{o.targa_motrice || '—'}</TableCell>
              <TableCell className="py-2">{o.autista_nome || '—'}</TableCell>
              <TableCell className="py-2"><StatusBadge stato={o.stato} /></TableCell>
              <TableCell className="py-2">
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

// ============================
// Pagina Planner
// ============================
export default function PlannerPage() {
  const [orders, setOrders] = useState([]);
  const [loading, setLoading] = useState(true);
  const [assignDialogOpen, setAssignDialogOpen] = useState(false);
  const [selectedOrder, setSelectedOrder] = useState(null);
  const [assignForm, setAssignForm] = useState({ targa_motrice: '', targa_rimorchio: '', autista_id: '', autista_nome: '', vettore_id: '', vettore_nome: '' });
  const [saving, setSaving] = useState(false);
  const [vehicles, setVehicles] = useState([]);
  const [drivers, setDrivers] = useState([]);
  const [carriers, setCarriers] = useState([]);
  const [dateFrom, setDateFrom] = useState('');
  const [dateTo, setDateTo] = useState('');
  const [tab, setTab] = useState('all');
  const [subTab, setSubTab] = useState('pianificabili');
  const [view, setView] = useState('list');
  const [availVehicles, setAvailVehicles] = useState([]);
  const [availDrivers, setAvailDrivers] = useState([]);
  const [availOpen, setAvailOpen] = useState(false);
  const [unavailOpen, setUnavailOpen] = useState(false);
  const [unavailList, setUnavailList] = useState([]);
  const [unavailForm, setUnavailForm] = useState({ autista_id: '', autista_nome: '', data_da: '', data_a: '', motivo: 'ferie', note: '' });
  const [savingUnavail, setSavingUnavail] = useState(false);

  const fetchOrders = useCallback(() => {
    setLoading(true);
    const params = {};
    if (dateFrom) params.data_da = dateFrom;
    if (dateTo) params.data_a = dateTo;
    getOrders(params).then(r => setOrders(r.data)).catch(err => logger.error('Errore caricamento ordini planner:', err)).finally(() => setLoading(false));
  }, [dateFrom, dateTo]);

  useEffect(() => { fetchOrders(); }, [fetchOrders]);

  useEffect(() => {
    Promise.all([getVehicles(), getDrivers(), getCarriers()]).then(([v, d, c]) => {
      setVehicles(v.data); setDrivers(d.data); setCarriers(c.data);
    }).catch(err => logger.error('Errore caricamento lookup planner:', err));
  }, []);

  // --- Assegna ---
  const openAssign = async (order) => {
    setSelectedOrder(order);
    setAssignForm({ targa_motrice: '', targa_rimorchio: '', autista_id: '', autista_nome: '', vettore_id: '', vettore_nome: '' });
    const da = order.data_ritiro || '';
    const a = order.data_consegna || order.data_ritiro || '';
    if (da && a) {
      try {
        const [vRes, dRes] = await Promise.all([getVehicleAvailability(da, a), getDriverAvailability(da, a)]);
        setAvailVehicles(vRes.data); setAvailDrivers(dRes.data);
      } catch (e) { logger.error('Disponibilità fetch error:', e); }
    }
    setAssignDialogOpen(true);
  };
  const handleAssign = async () => {
    setSaving(true);
    try {
      await assignOrder(selectedOrder.id, assignForm);
      toast.success(`Ordine ${selectedOrder.progressivo} assegnato`);
      setAssignDialogOpen(false);
      fetchOrders();
    } catch (e) { toast.error(e.response?.data?.detail || 'Errore'); } finally { setSaving(false); }
  };
  const handleClose = async (order) => {
    if (!window.confirm(`Chiudere l'ordine ${order.progressivo}?`)) return;
    try { await closeOrder(order.id); toast.success(`Ordine ${order.progressivo} chiuso`); fetchOrders(); }
    catch (e) { toast.error(e.response?.data?.detail || 'Errore'); }
  };
  const handleStart = async (order) => {
    if (!window.confirm(`Avviare il viaggio per l'ordine ${order.progressivo}?`)) return;
    try { await startOrder(order.id); toast.success(`Ordine ${order.progressivo} avviato`); fetchOrders(); }
    catch (e) { toast.error(e.response?.data?.detail || 'Errore'); }
  };
  const handleDiscard = async (order) => {
    if (!window.confirm(`Scartare l'ordine ${order.progressivo}? L'operazione non è reversibile.`)) return;
    try { await discardOrder(order.id); toast.success(`Ordine ${order.progressivo} scartato`); fetchOrders(); }
    catch (e) { toast.error(e.response?.data?.detail || 'Errore'); }
  };

  // --- Filtri per tab ---
  const filterByStatus = useCallback((list) => {
    if (subTab === 'pianificabili') return list.filter(o => o.stato === 'PIANIFICABILE');
    if (subTab === 'pianificato') return list.filter(o => o.stato === 'PIANIFICATO');
    if (subTab === 'viaggio') return list.filter(o => o.stato === 'VIAGGIO');
    if (subTab === 'chiusi') return list.filter(o => o.stato === 'CHIUSO');
    if (subTab === 'scartati') return list.filter(o => o.stato === 'SCARTATO');
    return list;
  }, [subTab]);

  // Separazione per tipologia (Partenze / Rientri) — memoizzate
  const exportOrders = useMemo(() => filterByStatus(orders.filter(o => o.tipologia === 'export')), [orders, filterByStatus]);
  const nazionaleOrders = useMemo(() => filterByStatus(orders.filter(o => o.tipologia === 'nazionale')), [orders, filterByStatus]);
  const importOrders = useMemo(() => filterByStatus(orders.filter(o => o.tipologia === 'import')), [orders, filterByStatus]);
  const soloEsteroOrders = useMemo(() => filterByStatus(orders.filter(o => o.tipologia === 'solo_estero')), [orders, filterByStatus]);
  const allFiltered = useMemo(() => filterByStatus(orders), [orders, filterByStatus]);

  // Contatori memoizzati per toolbar e tabs
  const pianificabiliCount = useMemo(() => orders.filter(o => o.stato === 'PIANIFICABILE').length, [orders]);
  const pianificatoCount = useMemo(() => orders.filter(o => o.stato === 'PIANIFICATO').length, [orders]);
  const inViaggioCount = useMemo(() => orders.filter(o => o.stato === 'VIAGGIO').length, [orders]);
  const chiusiCount = useMemo(() => orders.filter(o => o.stato === 'CHIUSO').length, [orders]);
  const scartatiCount = useMemo(() => orders.filter(o => o.stato === 'SCARTATO').length, [orders]);
  const partenzeCount = useMemo(() => orders.filter(o => o.tipologia === 'export' || o.tipologia === 'nazionale').length, [orders]);
  const rientriCount = useMemo(() => orders.filter(o => o.tipologia === 'import' || o.tipologia === 'solo_estero').length, [orders]);

  // Liste veicoli/autisti pre-filtrate per il dialog assegna
  const assignMotrici = useMemo(() => (availVehicles.length > 0 ? availVehicles : vehicles).filter(v => v.tipo_veicolo === 'motrice'), [availVehicles, vehicles]);
  const assignRimorchi = useMemo(() => (availVehicles.length > 0 ? availVehicles : vehicles).filter(v => v.tipo_veicolo !== 'motrice'), [availVehicles, vehicles]);
  const assignDriverList = useMemo(() => availDrivers.length > 0 ? availDrivers : drivers, [availDrivers, drivers]);

  // --- Setter ---
  const setDriver = (id) => {
    const allD = availDrivers.length > 0 ? availDrivers : drivers;
    const d = allD.find(x => x.id === id);
    setAssignForm({ ...assignForm, autista_id: id, autista_nome: d ? `${d.nome} ${d.cognome}` : '' });
  };
  const setVettore = (id) => {
    const c = carriers.find(x => x.id === id);
    setAssignForm({ ...assignForm, vettore_id: id, vettore_nome: c?.ragione_sociale || '' });
  };

  // --- Disponibilità ---
  const openAvailPanel = async () => {
    const da = dateFrom || new Date().toISOString().split('T')[0];
    const a = dateTo || new Date(Date.now() + 7 * 86400000).toISOString().split('T')[0];
    try {
      const [vRes, dRes] = await Promise.all([getVehicleAvailability(da, a), getDriverAvailability(da, a)]);
      setAvailVehicles(vRes.data); setAvailDrivers(dRes.data);
    } catch (e) { toast.error('Errore caricamento disponibilità'); }
    setAvailOpen(true);
  };

  // --- Indisponibilità ---
  const openUnavail = async () => {
    try { const res = await getDriverUnavailability({}); setUnavailList(res.data); } catch (e) { logger.error('Indisponibilità fetch error:', e); }
    setUnavailForm({ autista_id: '', autista_nome: '', data_da: '', data_a: '', motivo: 'ferie', note: '' });
    setUnavailOpen(true);
  };
  const handleCreateUnavail = async () => {
    if (!unavailForm.autista_id || !unavailForm.data_da || !unavailForm.data_a) { toast.error('Compilare autista, data inizio e fine'); return; }
    setSavingUnavail(true);
    try {
      await createDriverUnavailability(unavailForm);
      toast.success('Indisponibilità registrata');
      const res = await getDriverUnavailability({});
      setUnavailList(res.data);
      setUnavailForm({ autista_id: '', autista_nome: '', data_da: '', data_a: '', motivo: 'ferie', note: '' });
    } catch (e) { toast.error('Errore'); } finally { setSavingUnavail(false); }
  };
  const handleDeleteUnavail = async (id) => {
    try { await deleteDriverUnavailability(id); toast.success('Rimossa'); const res = await getDriverUnavailability({}); setUnavailList(res.data); }
    catch (e) { toast.error('Errore'); }
  };
  const setUnavailDriver = (id) => {
    const d = drivers.find(x => x.id === id);
    setUnavailForm({ ...unavailForm, autista_id: id, autista_nome: d ? `${d.nome} ${d.cognome}` : '' });
  };

  return (
    <div className="space-y-3" data-testid="planner-page">
      {/* Toolbar */}
      <div className="flex flex-col gap-2 lg:flex-row lg:items-center lg:justify-between">
        <div className="flex items-center gap-2 flex-wrap">
          <CalendarRange className="h-4 w-4 text-muted-foreground" />
          <Input type="date" value={dateFrom} onChange={e => setDateFrom(e.target.value)} className="h-9 w-40 text-sm" data-testid="planner-date-from" />
          <span className="text-sm text-muted-foreground">a</span>
          <Input type="date" value={dateTo} onChange={e => setDateTo(e.target.value)} className="h-9 w-40 text-sm" data-testid="planner-date-to" />
        </div>
        <div className="flex gap-2 items-center flex-wrap">
          <Button variant="outline" size="sm" className="text-xs gap-1.5" onClick={openAvailPanel} data-testid="planner-availability-button">
            <Users className="h-3.5 w-3.5" /> Disponibilità
          </Button>
          <Button variant="outline" size="sm" className="text-xs gap-1.5" onClick={openUnavail} data-testid="planner-unavail-button">
            <AlertTriangle className="h-3.5 w-3.5" /> Indisponibilità
          </Button>
          <Badge className="status-order-red border text-xs">{pianificabiliCount} da pianificare</Badge>
          <Badge className="status-order-yellow border text-xs">{pianificatoCount} pianificati</Badge>
          <Badge className="status-order-blue border text-xs">{inViaggioCount} in viaggio</Badge>
          <Badge className="status-order-green border text-xs">{chiusiCount} consegnati</Badge>
          <Badge className="status-order-gray border text-xs">{scartatiCount} scartati</Badge>
        </div>
      </div>

      {/* Tab principali: Tutti / Partenze / Rientri */}
      <Tabs value={tab} onValueChange={setTab} data-testid="planner-tabs">
        <div className="flex items-center justify-between gap-3 flex-wrap">
          <TabsList>
            <TabsTrigger value="all">Tutti ({orders.length})</TabsTrigger>
            <TabsTrigger value="partenze" data-testid="planner-tab-partenze">Partenze ({partenzeCount})</TabsTrigger>
            <TabsTrigger value="rientri" data-testid="planner-tab-rientri">Rientri ({rientriCount})</TabsTrigger>
          </TabsList>
          {/* Sub-filtro stato */}
          <div className="flex gap-1">
            {['all', 'pianificabili', 'pianificato', 'viaggio', 'chiusi', 'scartati'].map(s => (
              <Button key={s} variant={subTab === s ? 'default' : 'ghost'} size="sm" className="h-7 text-xs" onClick={() => setSubTab(s)}>
                {s === 'all' ? 'Tutti' : s === 'pianificabili' ? 'Da pianificare' : s === 'pianificato' ? 'Pianificati' : s === 'viaggio' ? 'In viaggio' : s === 'chiusi' ? 'Consegnati' : 'Scartati'}
              </Button>
            ))}
          </div>
          {/* Vista Lista ⇄ Calendario */}
          <div className="flex items-center gap-0.5 border rounded-md p-0.5">
            <Button variant={view === 'list' ? 'default' : 'ghost'} size="sm" className="h-7 px-2 text-xs gap-1" onClick={() => setView('list')} data-testid="planner-view-list">
              <ListIcon className="h-3.5 w-3.5" /> Lista
            </Button>
            <Button variant={view === 'calendar' ? 'default' : 'ghost'} size="sm" className="h-7 px-2 text-xs gap-1" onClick={() => setView('calendar')} data-testid="planner-view-calendar">
              <CalendarDays className="h-3.5 w-3.5" /> Calendario
            </Button>
          </div>
        </div>

        {view === 'calendar' ? (
          <div className="mt-3">
            <PlannerCalendar orders={allFiltered} onAssign={openAssign} onStart={handleStart} onClose={handleClose} />
          </div>
        ) : (
          <>
            {/* === TAB TUTTI === */}
            <TabsContent value="all" className="mt-3" data-testid="planner-grid">
              <OrderGrid orders={allFiltered} loading={loading} onAssign={openAssign} onStart={handleStart} onClose={handleClose} onDiscard={handleDiscard} emptyMsg="Nessun ordine trovato" />
            </TabsContent>

            {/* === TAB PARTENZE (export sopra, nazionale sotto) === */}
            <TabsContent value="partenze" className="mt-3 space-y-4" data-testid="planner-partenze">
              <OrderGrid
                orders={exportOrders} loading={loading} onAssign={openAssign} onStart={handleStart} onClose={handleClose} onDiscard={handleDiscard}
                title={`Italia → Estero — Export (${exportOrders.length})`}
                emptyMsg="Nessun ordine export"
              />
              <OrderGrid
                orders={nazionaleOrders} loading={loading} onAssign={openAssign} onStart={handleStart} onClose={handleClose} onDiscard={handleDiscard}
                title={`Italia → Italia — Nazionale (${nazionaleOrders.length})`}
                emptyMsg="Nessun ordine nazionale"
              />
            </TabsContent>

            {/* === TAB RIENTRI (import sopra, solo_estero sotto) === */}
            <TabsContent value="rientri" className="mt-3 space-y-4" data-testid="planner-rientri">
              <OrderGrid
                orders={importOrders} loading={loading} onAssign={openAssign} onStart={handleStart} onClose={handleClose} onDiscard={handleDiscard}
                title={`Estero → Italia — Import (${importOrders.length})`}
                emptyMsg="Nessun ordine import"
              />
              <OrderGrid
                orders={soloEsteroOrders} loading={loading} onAssign={openAssign} onStart={handleStart} onClose={handleClose} onDiscard={handleDiscard}
                title={`Estero → Estero — Solo Estero (${soloEsteroOrders.length})`}
                emptyMsg="Nessun ordine solo estero"
              />
            </TabsContent>
          </>
        )}
      </Tabs>

      {/* ===== DIALOG ASSEGNA ===== */}
      <Dialog open={assignDialogOpen} onOpenChange={setAssignDialogOpen}>
        <DialogContent className="max-w-lg">
          <DialogHeader>
            <DialogTitle style={{ fontFamily: "'Space Grotesk', sans-serif" }}>Assegna Ordine {selectedOrder?.progressivo}</DialogTitle>
          </DialogHeader>
          <div className="space-y-4">
            <div className="p-3 rounded-lg bg-muted/50 text-sm">
              <p><strong>{selectedOrder?.destinazione_carico_nome}</strong> → <strong>{selectedOrder?.destinazione_scarico_nome}</strong></p>
              <p className="text-muted-foreground">{selectedOrder?.cliente_nome} • {selectedOrder?.data_ritiro}</p>
            </div>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
              <div className="space-y-1.5">
                <Label>Targa Motrice</Label>
                <Select value={assignForm.targa_motrice} onValueChange={v => setAssignForm({ ...assignForm, targa_motrice: v })}>
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
                <Select value={assignForm.targa_rimorchio} onValueChange={v => setAssignForm({ ...assignForm, targa_rimorchio: v })}>
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
                <Select value={assignForm.autista_id} onValueChange={setDriver}>
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
                <Select value={assignForm.vettore_id} onValueChange={setVettore}>
                  <SelectTrigger><SelectValue placeholder="Vettore terzo" /></SelectTrigger>
                  <SelectContent>{carriers.map(c => <SelectItem key={c.id} value={c.id}>{c.ragione_sociale}</SelectItem>)}</SelectContent>
                </Select>
              </div>
            </div>
          </div>
          <DialogFooter className="gap-2">
            <Button variant="outline" onClick={() => setAssignDialogOpen(false)}>Annulla</Button>
            <Button onClick={handleAssign} disabled={saving} data-testid="planner-assign-submit">
              {saving && <Loader2 className="h-4 w-4 animate-spin mr-2" />} Assegna Viaggio
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* ===== DISPONIBILITÀ ===== */}
      <Dialog open={availOpen} onOpenChange={setAvailOpen}>
        <DialogContent className="max-w-3xl max-h-[85vh] overflow-y-auto">
          <DialogHeader><DialogTitle style={{ fontFamily: "'Space Grotesk', sans-serif" }}>Disponibilità Mezzi e Autisti</DialogTitle></DialogHeader>
          <div className="space-y-4">
            <p className="text-sm text-muted-foreground">Periodo: {dateFrom || 'oggi'} → {dateTo || '+7gg'}</p>
            <div>
              <Label className="mb-2 block text-sm font-semibold">Mezzi ({availVehicles.length})</Label>
              <div className="border rounded-lg divide-y max-h-48 overflow-y-auto">
                {availVehicles.map(v => (
                  <div key={v.id} className="flex items-center justify-between px-3 py-2 text-sm">
                    <div className="flex items-center gap-2">
                      <span className={`h-2.5 w-2.5 rounded-full ${v.disponibilita === 'busy' ? 'bg-red-500' : 'bg-emerald-500'}`} />
                      <span className="font-mono font-medium">{v.targa}</span>
                      <span className="text-muted-foreground">{v.marca} - {v.tipo_veicolo}</span>
                    </div>
                    <Badge variant="outline" className={`text-[10px] ${v.disponibilita === 'busy' ? 'border-red-300 text-red-700' : 'border-emerald-300 text-emerald-700'}`}>
                      {v.disponibilita === 'busy' ? 'Occupato' : 'Disponibile'}
                    </Badge>
                  </div>
                ))}
              </div>
            </div>
            <div>
              <Label className="mb-2 block text-sm font-semibold">Autisti ({availDrivers.length})</Label>
              <div className="border rounded-lg divide-y max-h-48 overflow-y-auto">
                {availDrivers.map(d => (
                  <div key={d.id} className="flex items-center justify-between px-3 py-2 text-sm">
                    <div className="flex items-center gap-2">
                      <span className={`h-2.5 w-2.5 rounded-full ${d.disponibilita === 'busy' ? 'bg-red-500' : d.disponibilita === 'unavailable' ? 'bg-amber-500' : 'bg-emerald-500'}`} />
                      <span className="font-medium">{d.nome} {d.cognome}</span>
                    </div>
                    <Badge variant="outline" className={`text-[10px] ${d.disponibilita === 'busy' ? 'border-red-300 text-red-700' : d.disponibilita === 'unavailable' ? 'border-amber-300 text-amber-700' : 'border-emerald-300 text-emerald-700'}`}>
                      {d.disponibilita === 'busy' ? 'Occupato' : d.disponibilita === 'unavailable' ? d.motivo_indisponibilita || 'Indisponibile' : 'Disponibile'}
                    </Badge>
                  </div>
                ))}
              </div>
            </div>
          </div>
        </DialogContent>
      </Dialog>

      {/* ===== INDISPONIBILITÀ ===== */}
      <Dialog open={unavailOpen} onOpenChange={setUnavailOpen}>
        <DialogContent className="max-w-2xl max-h-[85vh] overflow-y-auto">
          <DialogHeader><DialogTitle style={{ fontFamily: "'Space Grotesk', sans-serif" }}>Gestione Indisponibilità Autisti</DialogTitle></DialogHeader>
          <div className="space-y-4">
            <Card className="p-4 border-dashed">
              <Label className="mb-2 block text-sm font-semibold">Nuova Indisponibilità</Label>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
                <div className="space-y-1.5">
                  <Label className="text-xs">Autista *</Label>
                  <Select value={unavailForm.autista_id} onValueChange={setUnavailDriver}>
                    <SelectTrigger className="h-9"><SelectValue placeholder="Seleziona" /></SelectTrigger>
                    <SelectContent>{drivers.map(d => <SelectItem key={d.id} value={d.id}>{d.nome} {d.cognome}</SelectItem>)}</SelectContent>
                  </Select>
                </div>
                <div className="space-y-1.5">
                  <Label className="text-xs">Motivo</Label>
                  <Select value={unavailForm.motivo} onValueChange={v => setUnavailForm({ ...unavailForm, motivo: v })}>
                    <SelectTrigger className="h-9"><SelectValue /></SelectTrigger>
                    <SelectContent>
                      <SelectItem value="ferie">Ferie</SelectItem>
                      <SelectItem value="malattia">Malattia</SelectItem>
                      <SelectItem value="permesso">Permesso</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
                <div className="space-y-1.5"><Label className="text-xs">Dal *</Label><Input type="date" className="h-9" value={unavailForm.data_da} onChange={e => setUnavailForm({ ...unavailForm, data_da: e.target.value })} /></div>
                <div className="space-y-1.5"><Label className="text-xs">Al *</Label><Input type="date" className="h-9" value={unavailForm.data_a} onChange={e => setUnavailForm({ ...unavailForm, data_a: e.target.value })} /></div>
              </div>
              <Button size="sm" className="mt-3 text-xs gap-1.5" onClick={handleCreateUnavail} disabled={savingUnavail} data-testid="unavail-create-button">
                {savingUnavail && <Loader2 className="h-3 w-3 animate-spin" />}
                <Plus className="h-3.5 w-3.5" /> Registra
              </Button>
            </Card>
            <div>
              <Label className="mb-2 block text-sm font-semibold">Registrate ({unavailList.length})</Label>
              {unavailList.length === 0 ? (
                <p className="text-sm text-muted-foreground py-4 text-center">Nessuna indisponibilità.</p>
              ) : (
                <div className="border rounded-lg divide-y max-h-56 overflow-y-auto">
                  {unavailList.map(u => (
                    <div key={u.id} className="flex items-center justify-between px-3 py-2 text-sm">
                      <div>
                        <span className="font-medium">{u.autista_nome}</span>
                        <span className="mx-2 text-muted-foreground">{u.data_da} → {u.data_a}</span>
                        <Badge variant="outline" className="text-[10px]">{u.motivo}</Badge>
                      </div>
                      <Button variant="ghost" size="icon" className="h-7 w-7 text-destructive" onClick={() => handleDeleteUnavail(u.id)}><Trash2 className="h-3 w-3" /></Button>
                    </div>
                  ))}
                </div>
              )}
            </div>
          </div>
        </DialogContent>
      </Dialog>
    </div>
  );
}
