import { useState, useEffect, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import { getOrders, createOrder, deleteOrder, downloadOrderCmrPdf, getReturnSuggestions, getDestinations, getProducts, getTransportCategories, exportOrdersExcel, lookupTariff } from '@/lib/api';
import { getApiErrorMessage } from '@/lib/apiError';
import { useGetCustomersQuery } from '@/store/api/appApi';
import type { DtoOrderRequest, DtoOrderResponse, DtoDestinationResponse, DtoTransportCategoryResponse, DtoOrderReturnSuggestionsResponse } from '@/api/data-contracts';
import { formatEuro } from '@/lib/format';
import { Card } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import { Select, SelectTrigger, SelectValue, SelectContent, SelectItem } from '@/components/ui/select';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog';
import { Badge } from '@/components/ui/badge';
import { Checkbox } from '@/components/ui/checkbox';
import { Skeleton } from '@/components/ui/skeleton';
import { StatusBadge } from '@/components/shared/StatusBadge';
import SearchableSelect from '@/components/shared/SearchableSelect';
import { toast } from 'sonner';
import { logger } from '@/lib/logger';
import { Plus, Search, Download, Loader2, Eye, Trash2, FileText, ArrowLeftRight } from 'lucide-react';

const emptyForm: DtoOrderRequest = {
  cliente_id: '',
  committente_id: '',
  destinazione_carico_id: '',
  destinazione_scarico_id: '',
  data_ritiro: '', ora_ritiro_da: '06:00', ora_ritiro_a: '08:00',
  data_consegna: '', ora_consegna_da: '14:00', ora_consegna_a: '18:00',
  tariffa: 0, tipo_tariffa: 'forfait',
  tipologia: 'nazionale', categoria_trasporto: '', rif_ordine_cliente: '',
  rif_carico: '', note_carico: '', rif_scarico: '', note_scarico: '',
  andata_ritorno: false, provvisorio: false, note: '', items: [],
};

export default function OrdersPage() {
  const navigate = useNavigate();
  const [orders, setOrders] = useState<DtoOrderResponse[]>([]);
  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState('');
  const [statoFilter, setStatoFilter] = useState('');
  const [tipologiaFilter, setTipologiaFilter] = useState('');
  const [dialogOpen, setDialogOpen] = useState(false);
  const [saving, setSaving] = useState(false);
  // Ritorni (#32): dialog separato con candidati di ritorno per il riempimento
  // viaggi a vuoto. Score e motivi calcolati lato backend.
  const [returnsOpen, setReturnsOpen] = useState(false);
  const [returnsLoading, setReturnsLoading] = useState(false);
  const [returnsData, setReturnsData] = useState<DtoOrderReturnSuggestionsResponse | null>(null);

  // Lookup data — serve l'elenco completo, non una pagina.
  const { data: customersPage } = useGetCustomersQuery({ limit: 500 });
  const customers = customersPage?.items ?? [];
  const [destinations, setDestinations] = useState<DtoDestinationResponse[]>([]);
  const [categories, setCategories] = useState<DtoTransportCategoryResponse[]>([]);

  const [form, setForm] = useState<DtoOrderRequest>(emptyForm);

  const fetchOrders = useCallback(() => {
    setLoading(true);
    getOrders({ stato: statoFilter === 'all' ? '' : statoFilter, search, tipologia: tipologiaFilter === 'all' ? '' : tipologiaFilter })
      .then((r: { data: DtoOrderResponse[] }) => setOrders(r.data))
      .catch((err: unknown) => logger.error('Errore caricamento ordini:', err))
      .finally(() => setLoading(false));
  }, [statoFilter, search, tipologiaFilter]);

  useEffect(() => { fetchOrders(); }, [fetchOrders]);

  useEffect(() => {
    Promise.all([getDestinations(), getProducts(), getTransportCategories()])
      .then(([d, , cat]) => {
        setDestinations(d.data.data ?? []);
        setCategories(cat.data);
      });
  }, []);

  const openReturns = async (order: DtoOrderResponse) => {
    if (!order.id) return;
    setReturnsOpen(true);
    setReturnsLoading(true);
    setReturnsData(null);
    try {
      const r = await getReturnSuggestions(order.id, { max_days_gap: 2, limit: 20 });
      setReturnsData(r.data);
    } catch (e) {
      toast.error(getApiErrorMessage(e) || 'Errore caricamento ritorni');
      setReturnsOpen(false);
    } finally {
      setReturnsLoading(false);
    }
  };

  const handleExport = async () => {
    try {
      const res = await exportOrdersExcel({ stato: statoFilter });
      const url = URL.createObjectURL(new Blob([res.data]));
      const a = document.createElement('a'); a.href = url; a.download = 'ordini.xlsx'; a.click();
      toast.success('Esportazione completata');
    } catch (e) { toast.error('Errore esportazione'); }
  };

  const openNew = () => {
    setForm(emptyForm);
    setDialogOpen(true);
  };

  const handleSave = async () => {
    if (!form.cliente_id) { toast.error('Selezionare un cliente'); return; }
    setSaving(true);
    try {
      await createOrder(form);
      toast.success('Ordine creato');
      setDialogOpen(false);
      fetchOrders();
    } catch (e) { toast.error(getApiErrorMessage(e) || 'Errore'); } finally { setSaving(false); }
  };

  const handleDelete = async (id?: string) => {
    if (!id) return;
    if (!window.confirm('Eliminare questo ordine?')) return;
    try { await deleteOrder(id); toast.success('Eliminato'); fetchOrders(); } catch (e) { toast.error(getApiErrorMessage(e) || 'Errore'); }
  };

  const setCustomer = (id: string) => {
    const newForm = { ...form, cliente_id: id };
    setForm(newForm);
    tryLookupTariff(id, newForm.destinazione_carico_id, newForm.destinazione_scarico_id);
  };
  const setCarico = (id: string) => {
    const newForm = { ...form, destinazione_carico_id: id };
    setForm(newForm);
    tryLookupTariff(newForm.cliente_id, id, newForm.destinazione_scarico_id);
  };
  const setScarico = (id: string) => {
    const newForm = { ...form, destinazione_scarico_id: id };
    setForm(newForm);
    tryLookupTariff(newForm.cliente_id, newForm.destinazione_carico_id, id);
  };

  const tryLookupTariff = async (clienteId?: string, caricoId?: string, scaricoId?: string) => {
    if (!clienteId || !caricoId || !scaricoId) return;
    try {
      const res = await lookupTariff({ cliente_id: clienteId, carico_id: caricoId, scarico_id: scaricoId });
      if (res.data.found) {
        setForm(prev => ({ ...prev, tariffa: res.data.tariffa, tipo_tariffa: res.data.tipo_tariffa }));
        toast.info(`Tariffa proposta dal listino: € ${res.data.tariffa.toLocaleString('it-IT')} (${res.data.tipo_tariffa})`);
      }
    } catch (e) { logger.error('Tariff lookup error:', e); /* non-blocking */ }
  };

  return (
    <div className="space-y-3" data-testid="orders-page">
      {/* Filters */}
      <div className="flex flex-col gap-2 lg:flex-row lg:items-center lg:justify-between" data-testid="filter-bar">
        <div className="flex flex-wrap gap-2 items-center">
          <div className="relative">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-muted-foreground" />
            <Input data-testid="orders-search-input" placeholder="Cerca ordini..." value={search} onChange={e => setSearch(e.target.value)} className="pl-9 h-9 text-sm w-64" />
          </div>
          <Select value={statoFilter} onValueChange={setStatoFilter}>
            <SelectTrigger className="h-9 w-44 text-sm"><SelectValue placeholder="Tutti gli stati" /></SelectTrigger>
            <SelectContent>
              <SelectItem value="all">Tutti gli stati</SelectItem>
              <SelectItem value="PIANIFICABILE">Da pianificare</SelectItem>
              <SelectItem value="PIANIFICATO">Pianificato</SelectItem>
              <SelectItem value="VIAGGIO">In viaggio</SelectItem>
              <SelectItem value="CHIUSO">Consegnato</SelectItem>
              <SelectItem value="SCARTATO">Scartato</SelectItem>
            </SelectContent>
          </Select>
          <Select value={tipologiaFilter} onValueChange={setTipologiaFilter}>
            <SelectTrigger className="h-9 w-40 text-sm"><SelectValue placeholder="Tipologia" /></SelectTrigger>
            <SelectContent>
              <SelectItem value="all">Tutte</SelectItem>
              <SelectItem value="import">Import</SelectItem>
              <SelectItem value="export">Export</SelectItem>
              <SelectItem value="nazionale">Nazionale</SelectItem>
              <SelectItem value="solo_estero">Solo Estero</SelectItem>
            </SelectContent>
          </Select>
        </div>
        <div className="flex gap-2">
          <Button variant="outline" size="sm" onClick={handleExport} className="text-xs gap-1.5" data-testid="orders-export-button">
            <Download className="h-3.5 w-3.5" /> Esporta Excel
          </Button>
          <Button size="sm" onClick={openNew} className="text-xs gap-1.5" data-testid="orders-new-button">
            <Plus className="h-3.5 w-3.5" /> Nuovo Ordine
          </Button>
        </div>
      </div>

      {/* Table */}
      <Card className="rounded-xl border shadow-sm" data-testid="orders-table">
        <div className="overflow-x-auto">
          <Table className="text-xs md:text-sm">
            <TableHeader>
              <TableRow>
                <TableHead className="py-2 text-xs">Prog.</TableHead>
                <TableHead className="py-2 text-xs">Cliente</TableHead>
                <TableHead className="py-2 text-xs">Carico</TableHead>
                <TableHead className="py-2 text-xs">Scarico</TableHead>
                <TableHead className="py-2 text-xs">Data Ritiro</TableHead>
                <TableHead className="py-2 text-xs text-right">Tariffa</TableHead>
                <TableHead className="py-2 text-xs">Tipo</TableHead>
                <TableHead className="py-2 text-xs">Stato</TableHead>
                <TableHead className="py-2 text-xs w-20"></TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {loading ? Array.from({ length: 5 }).map((_, i) => (
                <TableRow key={`skel-row-${i}`}>{Array.from({ length: 9 }).map((_, j) => <TableCell key={`skel-col-${j}`} className="py-2"><Skeleton className="h-4 w-full" /></TableCell>)}</TableRow>
              )) : orders.length === 0 ? (
                <TableRow><TableCell colSpan={9} className="text-center py-8 text-muted-foreground">Nessun ordine trovato</TableCell></TableRow>
              ) : orders.map(o => (
                <TableRow
                  key={o.id}
                  className="hover:bg-muted/60 cursor-pointer"
                  onClick={() => navigate(`/planner/ordini/${o.id}`, { state: { from: '/ordini', fromLabel: 'Ordini', readOnly: true } })}
                >
                  <TableCell className="py-2 font-mono font-medium">{o.progressivo}</TableCell>
                  <TableCell className="py-2 max-w-[150px] truncate">{o.cliente?.ragione_sociale}</TableCell>
                  <TableCell className="py-2 max-w-[120px] truncate">{o.destinazione_carico?.nome}</TableCell>
                  <TableCell className="py-2 max-w-[120px] truncate">{o.destinazione_scarico?.nome}</TableCell>
                  <TableCell className="py-2 whitespace-nowrap">{o.data_ritiro}</TableCell>
                  <TableCell className="py-2 text-right tabular-nums">€ {formatEuro(o.tariffa || 0)}</TableCell>
                  <TableCell className="py-2"><Badge variant="outline" className="text-[10px]">{o.tipologia}</Badge></TableCell>
                  <TableCell className="py-2">
                    <div className="flex flex-col items-start gap-1">
                      <StatusBadge stato={o.stato} />
                      {o.provvisorio && <Badge variant="secondary" className="text-[10px] bg-amber-100 text-amber-800 dark:bg-amber-500/15 dark:text-amber-300">provvisorio</Badge>}
                    </div>
                  </TableCell>
                  <TableCell className="py-2" onClick={(e) => e.stopPropagation()}>
                    <div className="flex gap-1">
                      <Button variant="ghost" size="icon" className="h-7 w-7" onClick={() => navigate(`/planner/ordini/${o.id}`, { state: { from: '/ordini', fromLabel: 'Ordini', readOnly: true } })}><Eye className="h-3 w-3" /></Button>
                      {o.data_consegna && o.destinazione_scarico?.nome && (
                        <Button
                          variant="ghost" size="icon" className="h-7 w-7"
                          title="Suggerimenti ritorni" aria-label="Suggerimenti ordini di ritorno"
                          onClick={() => openReturns(o)}
                        >
                          <ArrowLeftRight className="h-3 w-3" />
                        </Button>
                      )}
                      {o.tipologia === 'internazionale' && (
                        <Button
                          variant="ghost" size="icon" className="h-7 w-7"
                          title="Scarica CMR" aria-label="Scarica CMR"
                          onClick={async () => {
                            if (!o.id) return;
                            try {
                              const r = await downloadOrderCmrPdf(o.id);
                              const cd = r.headers?.['content-disposition'] || '';
                              const m = cd.match(/filename="?([^";]+)"?/i);
                              const filename = m ? m[1] : `cmr_${o.progressivo || o.id}.pdf`;
                              const url = window.URL.createObjectURL(new Blob([r.data], { type: 'application/pdf' }));
                              const a = document.createElement('a');
                              a.href = url; a.download = filename;
                              document.body.appendChild(a); a.click(); a.remove();
                              window.URL.revokeObjectURL(url);
                            } catch (e) {
                              toast.error(getApiErrorMessage(e) || 'Errore download CMR');
                            }
                          }}
                        >
                          <FileText className="h-3 w-3" />
                        </Button>
                      )}
                      {o.stato === 'PIANIFICABILE' && <Button variant="ghost" size="icon" className="h-7 w-7 text-destructive" onClick={() => handleDelete(o.id)}><Trash2 className="h-3 w-3" /></Button>}
                    </div>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
      </Card>

      {/* New Order Dialog */}
      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <DialogContent className="max-w-2xl max-h-[85vh] overflow-y-auto">
          <DialogHeader><DialogTitle style={{ fontFamily: "'Space Grotesk', sans-serif" }}>Nuovo Ordine</DialogTitle></DialogHeader>
          <form onSubmit={(e) => { e.preventDefault(); handleSave(); }} className="space-y-4">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
              <div className="space-y-1.5">
                <Label>Cliente *</Label>
                <SearchableSelect
                  value={form.cliente_id}
                  onValueChange={setCustomer}
                  options={customers}
                  getValue={(c) => c.id || ''}
                  getLabel={(c) => c.ragione_sociale || ''}
                  placeholder="Seleziona cliente"
                  searchPlaceholder="Cerca cliente..."
                />
              </div>
              <div className="space-y-1.5">
                <Label>Committente</Label>
                {/* Opzionale (vuoto = coincide col cliente): ri-cliccare il
                    committente selezionato lo deseleziona. */}
                <SearchableSelect
                  value={form.committente_id}
                  onValueChange={v => setForm({ ...form, committente_id: v === form.committente_id ? '' : v })}
                  options={customers}
                  getValue={(c) => c.id || ''}
                  getLabel={(c) => c.ragione_sociale || ''}
                  placeholder="— uguale al cliente —"
                  searchPlaceholder="Cerca committente..."
                />
              </div>
              <div className="space-y-1.5">
                <Label>Destinazione Carico</Label>
                <SearchableSelect
                  value={form.destinazione_carico_id}
                  onValueChange={setCarico}
                  options={destinations}
                  getValue={(d) => d.id || ''}
                  getLabel={(d) => d.nome || ''}
                  placeholder="Carico"
                  searchPlaceholder="Cerca destinazione..."
                />
              </div>
              <div className="space-y-1.5">
                <Label>Destinazione Scarico</Label>
                <SearchableSelect
                  value={form.destinazione_scarico_id}
                  onValueChange={setScarico}
                  options={destinations}
                  getValue={(d) => d.id || ''}
                  getLabel={(d) => d.nome || ''}
                  placeholder="Scarico"
                  searchPlaceholder="Cerca destinazione..."
                />
              </div>
              <div className="space-y-1.5"><Label>Rif. Carico</Label><Input value={form.rif_carico} onChange={e => setForm({ ...form, rif_carico: e.target.value })} /></div>
              <div className="space-y-1.5"><Label>Rif. Scarico</Label><Input value={form.rif_scarico} onChange={e => setForm({ ...form, rif_scarico: e.target.value })} /></div>
              <div className="space-y-1.5"><Label>Note Carico</Label><Input value={form.note_carico} onChange={e => setForm({ ...form, note_carico: e.target.value })} /></div>
              <div className="space-y-1.5"><Label>Note Scarico</Label><Input value={form.note_scarico} onChange={e => setForm({ ...form, note_scarico: e.target.value })} /></div>
              <div className="space-y-1.5"><Label>Data Ritiro</Label><Input type="date" value={form.data_ritiro} onChange={e => setForm({ ...form, data_ritiro: e.target.value })} /></div>
              <div className="space-y-1.5"><Label>Data Consegna</Label><Input type="date" value={form.data_consegna} onChange={e => setForm({ ...form, data_consegna: e.target.value })} /></div>
              <div className="space-y-1.5"><Label>Orario Ritiro (da-a)</Label><div className="flex gap-2"><Input type="time" value={form.ora_ritiro_da} onChange={e => setForm({ ...form, ora_ritiro_da: e.target.value })} /><Input type="time" value={form.ora_ritiro_a} onChange={e => setForm({ ...form, ora_ritiro_a: e.target.value })} /></div></div>
              <div className="space-y-1.5"><Label>Orario Consegna (da-a)</Label><div className="flex gap-2"><Input type="time" value={form.ora_consegna_da} onChange={e => setForm({ ...form, ora_consegna_da: e.target.value })} /><Input type="time" value={form.ora_consegna_a} onChange={e => setForm({ ...form, ora_consegna_a: e.target.value })} /></div></div>
              <div className="space-y-1.5"><Label>Tariffa (€)</Label><Input type="number" value={form.tariffa} onChange={e => setForm({ ...form, tariffa: Number(e.target.value) })} /></div>
              <div className="space-y-1.5">
                <Label>Tipo Tariffa</Label>
                <Select value={form.tipo_tariffa} onValueChange={v => setForm({ ...form, tipo_tariffa: v })}>
                  <SelectTrigger><SelectValue /></SelectTrigger>
                  <SelectContent><SelectItem value="forfait">Forfait</SelectItem><SelectItem value="euro_kg">€/Kg</SelectItem></SelectContent>
                </Select>
              </div>
              <div className="space-y-1.5">
                <Label>Tipologia</Label>
                <Select value={form.tipologia} onValueChange={v => setForm({ ...form, tipologia: v })}>
                  <SelectTrigger><SelectValue /></SelectTrigger>
                  <SelectContent><SelectItem value="import">Import</SelectItem><SelectItem value="export">Export</SelectItem><SelectItem value="nazionale">Nazionale</SelectItem><SelectItem value="solo_estero">Solo Estero</SelectItem></SelectContent>
                </Select>
              </div>
              <div className="space-y-1.5">
                <Label>Categoria Trasporto</Label>
                <Select value={form.categoria_trasporto} onValueChange={v => setForm({ ...form, categoria_trasporto: v })}>
                  <SelectTrigger><SelectValue placeholder="Categoria" /></SelectTrigger>
                  <SelectContent>{categories.map(c => <SelectItem key={c.id} value={c.nome || ''}>{c.nome}</SelectItem>)}</SelectContent>
                </Select>
              </div>
              <div className="space-y-1.5"><Label>Rif. Ordine Cliente</Label><Input value={form.rif_ordine_cliente} onChange={e => setForm({ ...form, rif_ordine_cliente: e.target.value })} /></div>
              <div className="flex items-end gap-2 pb-2">
                <Checkbox id="order-provvisorio" checked={!!form.provvisorio} onCheckedChange={v => setForm({ ...form, provvisorio: !!v })} />
                <Label htmlFor="order-provvisorio" className="cursor-pointer">Ordine provvisorio</Label>
              </div>
              <div className="md:col-span-2 space-y-1.5"><Label>Note</Label><Textarea value={form.note} onChange={e => setForm({ ...form, note: e.target.value })} rows={2} /></div>
            </div>
            <DialogFooter className="gap-2">
              <Button type="button" variant="outline" onClick={() => setDialogOpen(false)}>Annulla</Button>
              <Button type="submit" disabled={saving} data-testid="order-submit-button">
                {saving && <Loader2 className="h-4 w-4 animate-spin mr-2" />} Crea Ordine
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      {/* Ritorni (#32) */}
      <Dialog open={returnsOpen} onOpenChange={setReturnsOpen}>
        <DialogContent className="max-w-3xl">
          <DialogHeader>
            <DialogTitle>Suggerimenti ordini di ritorno</DialogTitle>
          </DialogHeader>
          {returnsLoading ? (
            <div className="py-12 flex items-center justify-center"><Loader2 className="h-5 w-5 animate-spin text-muted-foreground" /></div>
          ) : returnsData ? (
            <div className="space-y-3 text-sm">
              <div className="rounded-lg border bg-muted/40 p-3">
                <p className="text-xs text-muted-foreground mb-1">Ordine di andata</p>
                <p className="font-medium">{returnsData.source_order?.cliente?.ragione_sociale} — {returnsData.source_order?.progressivo}</p>
                <p className="text-xs text-muted-foreground">
                  Scarico a <span className="font-medium">{returnsData.source_order?.destinazione_scarico?.nome}</span> il {returnsData.source_order?.data_consegna || '—'}
                </p>
              </div>
              {!returnsData.count ? (
                <div className="py-8 text-center text-muted-foreground">
                  Nessun ordine di ritorno disponibile in questo periodo.<br />
                  <span className="text-xs">Cerco PIANIFICABILE che partono da {returnsData.source_order?.destinazione_scarico?.nome} entro 2 giorni dallo scarico.</span>
                </div>
              ) : (
                <div className="space-y-2 max-h-[420px] overflow-y-auto">
                  {(returnsData.candidates || []).map((c) => (
                    <div key={c.order?.id} className="rounded-lg border p-3 hover:bg-muted/30">
                      <div className="flex items-start justify-between gap-3">
                        <div className="flex-1 min-w-0">
                          <div className="flex items-center gap-2">
                            <Badge variant={(c.score || 0) >= 60 ? 'default' : 'outline'} className="text-[10px] tabular-nums">{c.score}</Badge>
                            <span className="font-mono text-xs">{c.order?.progressivo}</span>
                            <span className="text-xs text-muted-foreground truncate">{c.order?.cliente?.ragione_sociale}</span>
                          </div>
                          <p className="text-xs mt-1 truncate">
                            <span className="text-muted-foreground">Da</span> {c.order?.destinazione_carico?.nome}
                            <span className="text-muted-foreground"> a </span> {c.order?.destinazione_scarico?.nome}
                          </p>
                          <p className="text-xs text-muted-foreground">
                            Ritiro {c.order?.data_ritiro} · € {formatEuro(c.order?.tariffa || 0)} · {c.order?.tipologia}
                          </p>
                          <ul className="text-[11px] text-muted-foreground mt-1 list-disc list-inside">
                            {(c.reasons || []).map((r, i) => <li key={i}>{r}</li>)}
                          </ul>
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              )}
              <p className="text-[11px] text-muted-foreground pt-2 border-t">
                Assegna lo stesso mezzo/autista a entrambi gli ordini dal Planner per pianificarli come un unico viaggio di andata e ritorno.
              </p>
            </div>
          ) : null}
          <DialogFooter>
            <Button variant="outline" onClick={() => setReturnsOpen(false)}>Chiudi</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
