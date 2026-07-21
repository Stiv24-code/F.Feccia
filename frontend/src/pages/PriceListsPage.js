import { useState, useEffect, useCallback } from 'react';
import { getPriceLists, getPriceList, createPriceList, deletePriceList, addPriceListItem, updatePriceListItem, deletePriceListItem, getDestinations, getProducts } from '@/lib/api';
import { useGetCustomersQuery } from '@/store/api/appApi';
import { formatEuro } from '@/lib/format';
import { Card } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Label } from '@/components/ui/label';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import { Select, SelectTrigger, SelectValue, SelectContent, SelectItem } from '@/components/ui/select';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog';
import { Skeleton } from '@/components/ui/skeleton';
import { Badge } from '@/components/ui/badge';
import { Separator } from '@/components/ui/separator';
import { toast } from 'sonner';
import { logger } from '@/lib/logger';
import { Plus, Trash2, Loader2, Search, Eye, ArrowLeft, Pencil } from 'lucide-react';

const emptyRule = {
  prodotto_id: '', prodotto_nome: '',
  destinazione_carico_id: '', destinazione_carico_nome: '',
  destinazione_scarico_id: '', destinazione_scarico_nome: '',
  tariffa: 0, tipo_tariffa: 'forfait',
  range_peso_min: 0, range_peso_max: 0, unita_peso: 'Kg',
  minimo_tassabile: 0, tipo_trasporto: 'stradale',
  perc_adeguamento_carburante: 0
};

export default function PriceListsPage() {
  const [lists, setLists] = useState([]);
  const [loading, setLoading] = useState(true);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [saving, setSaving] = useState(false);
  const { data: customers = [] } = useGetCustomersQuery();
  const [destinations, setDestinations] = useState([]);
  const [products, setProducts] = useState([]);
  const [search, setSearch] = useState('');
  const [form, setForm] = useState({ cliente_id: '', cliente_nome: '', data_inizio: '', data_fine: '', items: [], note: '' });

  // Dettaglio listino
  const [detailOpen, setDetailOpen] = useState(false);
  const [selectedList, setSelectedList] = useState(null);
  const [ruleDialogOpen, setRuleDialogOpen] = useState(false);
  const [ruleForm, setRuleForm] = useState(emptyRule);
  const [editingItemId, setEditingItemId] = useState(null);
  const [savingRule, setSavingRule] = useState(false);

  const fetchData = useCallback(() => { setLoading(true); getPriceLists().then(r => setLists(r.data)).catch(err => logger.error('Errore caricamento listini:', err)).finally(() => setLoading(false)); }, []);
  useEffect(() => { fetchData(); }, [fetchData]);
  useEffect(() => {
    Promise.all([getDestinations(), getProducts()]).then(([d, p]) => {
      setDestinations(d.data); setProducts(p.data);
    }).catch(err => logger.error('Errore caricamento lookup listini:', err));
  }, []);

  // --- Lista principale ---
  const openNew = () => {
    setForm({ cliente_id: '', cliente_nome: '', data_inizio: '', data_fine: '', items: [], note: '' });
    setDialogOpen(true);
  };
  const setCustomer = (id) => {
    const c = customers.find(x => x.id === id);
    setForm({ ...form, cliente_id: id, cliente_nome: c?.ragione_sociale || '' });
  };
  const handleSave = async () => {
    if (!form.cliente_id) { toast.error('Selezionare un cliente'); return; }
    setSaving(true);
    try {
      await createPriceList(form);
      toast.success('Listino creato');
      setDialogOpen(false);
      fetchData();
    } catch (e) { toast.error('Errore'); } finally { setSaving(false); }
  };
  const handleDelete = async (id) => {
    if (!window.confirm('Eliminare questo listino?')) return;
    try { await deletePriceList(id); toast.success('Eliminato'); fetchData(); } catch (e) { toast.error('Errore'); }
  };

  // --- Dettaglio ---
  const openDetail = async (listino) => {
    try {
      const res = await getPriceList(listino.id);
      setSelectedList(res.data);
      setDetailOpen(true);
    } catch (e) { toast.error('Errore caricamento dettaglio'); }
  };
  const refreshDetail = async () => {
    if (!selectedList) return;
    try { const res = await getPriceList(selectedList.id); setSelectedList(res.data); } catch (e) { logger.error('Errore refresh dettaglio listino:', e); }
  };

  // --- Regole CRUD ---
  const openAddRule = () => {
    setRuleForm({ ...emptyRule });
    setEditingItemId(null);
    setRuleDialogOpen(true);
  };
  const openEditRule = (item) => {
    setRuleForm({
      prodotto_id: item.prodotto_id || '',
      prodotto_nome: item.prodotto_nome || '',
      destinazione_carico_id: item.destinazione_carico_id || '',
      destinazione_carico_nome: item.destinazione_carico_nome || '',
      destinazione_scarico_id: item.destinazione_scarico_id || '',
      destinazione_scarico_nome: item.destinazione_scarico_nome || '',
      tariffa: item.tariffa || 0,
      tipo_tariffa: item.tipo_tariffa || 'forfait',
      range_peso_min: item.range_peso_min || 0,
      range_peso_max: item.range_peso_max || 0,
      unita_peso: item.unita_peso || 'Kg',
      minimo_tassabile: item.minimo_tassabile || 0,
      tipo_trasporto: item.tipo_trasporto || 'stradale',
      perc_adeguamento_carburante: item.perc_adeguamento_carburante || 0,
    });
    setEditingItemId(item.item_id);
    setRuleDialogOpen(true);
  };

  const handleSaveRule = async () => {
    if (ruleForm.tariffa <= 0) { toast.error('Inserire una tariffa valida'); return; }
    setSavingRule(true);
    try {
      // Pulisci "any" dai select
      const cleanedForm = { ...ruleForm };
      if (cleanedForm.prodotto_id === 'any') { cleanedForm.prodotto_id = ''; cleanedForm.prodotto_nome = ''; }
      if (cleanedForm.destinazione_carico_id === 'any') { cleanedForm.destinazione_carico_id = ''; cleanedForm.destinazione_carico_nome = ''; }
      if (cleanedForm.destinazione_scarico_id === 'any') { cleanedForm.destinazione_scarico_id = ''; cleanedForm.destinazione_scarico_nome = ''; }

      if (editingItemId) {
        await updatePriceListItem(selectedList.id, editingItemId, cleanedForm);
        toast.success('Regola aggiornata');
      } else {
        await addPriceListItem(selectedList.id, cleanedForm);
        toast.success('Regola aggiunta');
      }
      setRuleDialogOpen(false);
      await refreshDetail();
    } catch (e) { toast.error(e.response?.data?.detail || 'Errore'); } finally { setSavingRule(false); }
  };

  const handleDeleteRule = async (itemId) => {
    if (!window.confirm('Eliminare questa regola?')) return;
    try {
      await deletePriceListItem(selectedList.id, itemId);
      toast.success('Regola eliminata');
      await refreshDetail();
    } catch (e) { toast.error('Errore'); }
  };

  // Setter per select regola
  const setRuleCarico = (id) => {
    const d = id === 'any' ? null : destinations.find(x => x.id === id);
    setRuleForm({ ...ruleForm, destinazione_carico_id: id, destinazione_carico_nome: d?.nome || '' });
  };
  const setRuleScarico = (id) => {
    const d = id === 'any' ? null : destinations.find(x => x.id === id);
    setRuleForm({ ...ruleForm, destinazione_scarico_id: id, destinazione_scarico_nome: d?.nome || '' });
  };
  const setRuleProdotto = (id) => {
    const p = id === 'any' ? null : products.find(x => x.id === id);
    setRuleForm({ ...ruleForm, prodotto_id: id, prodotto_nome: p ? `${p.codice} - ${p.descrizione}` : '' });
  };

  const filtered = search ? lists.filter(l => l.cliente_nome?.toLowerCase().includes(search.toLowerCase())) : lists;

  // ===========================
  // VISTA DETTAGLIO LISTINO
  // ===========================
  if (detailOpen && selectedList) {
    return (
      <div className="space-y-4" data-testid="pricelist-detail-page">
        <Button variant="ghost" size="sm" onClick={() => { setDetailOpen(false); fetchData(); }} className="gap-1.5 text-xs">
          <ArrowLeft className="h-3.5 w-3.5" /> Torna ai listini
        </Button>

        <Card className="p-4 lg:p-5 shadow-sm">
          <div className="flex flex-col lg:flex-row lg:items-center lg:justify-between gap-3 mb-4">
            <div>
              <h2 className="text-lg font-semibold" style={{ fontFamily: "'Space Grotesk', sans-serif" }}>
                Listino: {selectedList.cliente_nome}
              </h2>
              <p className="text-sm text-muted-foreground">
                Validità: {selectedList.data_inizio} → {selectedList.data_fine}
                {selectedList.in_uso && <Badge className="ml-2 bg-emerald-100 text-emerald-800 text-[10px]">In uso</Badge>}
              </p>
            </div>
            <Button size="sm" onClick={openAddRule} className="text-xs gap-1.5" data-testid="pricelist-add-rule-button">
              <Plus className="h-3.5 w-3.5" /> Aggiungi Regola
            </Button>
          </div>

          <div className="overflow-x-auto rounded-lg border">
            <Table className="text-xs md:text-sm">
              <TableHeader>
                <TableRow>
                  <TableHead className="py-2 text-xs">Prodotto</TableHead>
                  <TableHead className="py-2 text-xs">Tratta</TableHead>
                  <TableHead className="py-2 text-xs text-right">Tariffa</TableHead>
                  <TableHead className="py-2 text-xs">Tipo</TableHead>
                  <TableHead className="py-2 text-xs">Range Peso</TableHead>
                  <TableHead className="py-2 text-xs text-right">Min. Tass.</TableHead>
                  <TableHead className="py-2 text-xs">Trasporto</TableHead>
                  <TableHead className="py-2 text-xs text-right">% Carb.</TableHead>
                  <TableHead className="py-2 text-xs w-20"></TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {(!selectedList.items || selectedList.items.length === 0) ? (
                  <TableRow>
                    <TableCell colSpan={9} className="text-center py-8 text-muted-foreground">
                      Nessuna regola tariffaria. Aggiungi la prima regola.
                    </TableCell>
                  </TableRow>
                ) : selectedList.items.map((item) => (
                  <TableRow key={item.item_id} className="hover:bg-muted/60">
                    <TableCell className="py-2">{item.prodotto_nome || <span className="text-muted-foreground italic">Qualsiasi</span>}</TableCell>
                    <TableCell className="py-2">
                      {(item.destinazione_carico_nome || item.destinazione_scarico_nome) ? (
                        <span className="text-xs">{item.destinazione_carico_nome || '*'} → {item.destinazione_scarico_nome || '*'}</span>
                      ) : <span className="text-muted-foreground italic">Qualsiasi</span>}
                    </TableCell>
                    <TableCell className="py-2 text-right tabular-nums font-medium">€ {formatEuro(item.tariffa)}</TableCell>
                    <TableCell className="py-2">
                      <Badge variant="outline" className="text-[10px]">{item.tipo_tariffa === 'euro_kg' ? '€/Kg' : 'Forfait'}</Badge>
                    </TableCell>
                    <TableCell className="py-2 tabular-nums">
                      {(item.range_peso_min > 0 || item.range_peso_max > 0)
                        ? `${formatEuro(item.range_peso_min)} - ${formatEuro(item.range_peso_max)} ${item.unita_peso}`
                        : <span className="text-muted-foreground">—</span>}
                    </TableCell>
                    <TableCell className="py-2 text-right tabular-nums">
                      {item.minimo_tassabile > 0 ? `${formatEuro(item.minimo_tassabile)} ${item.unita_peso}` : '—'}
                    </TableCell>
                    <TableCell className="py-2">
                      <Badge variant="outline" className="text-[10px]">{item.tipo_trasporto === 'intermodale' ? 'Intermod.' : 'Stradale'}</Badge>
                    </TableCell>
                    <TableCell className="py-2 text-right tabular-nums">
                      {item.perc_adeguamento_carburante > 0 ? `${item.perc_adeguamento_carburante}%` : '—'}
                    </TableCell>
                    <TableCell className="py-2">
                      <div className="flex gap-0.5">
                        <Button variant="ghost" size="icon" className="h-7 w-7" onClick={() => openEditRule(item)} data-testid="pricelist-edit-rule-button">
                          <Pencil className="h-3 w-3" />
                        </Button>
                        <Button variant="ghost" size="icon" className="h-7 w-7 text-destructive" onClick={() => handleDeleteRule(item.item_id)} data-testid="pricelist-delete-rule-button">
                          <Trash2 className="h-3 w-3" />
                        </Button>
                      </div>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
        </Card>

        {/* Dialog Regola (Aggiungi / Modifica) */}
        <Dialog open={ruleDialogOpen} onOpenChange={setRuleDialogOpen}>
          <DialogContent className="max-w-2xl max-h-[85vh] overflow-y-auto">
            <DialogHeader>
              <DialogTitle style={{ fontFamily: "'Space Grotesk', sans-serif" }}>
                {editingItemId ? 'Modifica Regola Tariffaria' : 'Aggiungi Regola Tariffaria'}
              </DialogTitle>
            </DialogHeader>
            <div className="space-y-4">
              <p className="text-xs text-muted-foreground">Lascia vuoti Prodotto / Tratta per regola generica (wildcard).</p>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
                <div className="space-y-1.5">
                  <Label>Prodotto (opzionale)</Label>
                  <Select value={ruleForm.prodotto_id || 'any'} onValueChange={setRuleProdotto}>
                    <SelectTrigger><SelectValue placeholder="Qualsiasi" /></SelectTrigger>
                    <SelectContent>
                      <SelectItem value="any">Qualsiasi</SelectItem>
                      {products.map(p => <SelectItem key={p.id} value={p.id}>{p.codice} - {p.descrizione}</SelectItem>)}
                    </SelectContent>
                  </Select>
                </div>
                <div className="space-y-1.5">
                  <Label>Tipo Trasporto</Label>
                  <Select value={ruleForm.tipo_trasporto} onValueChange={v => setRuleForm({...ruleForm, tipo_trasporto: v})}>
                    <SelectTrigger><SelectValue /></SelectTrigger>
                    <SelectContent>
                      <SelectItem value="stradale">Stradale</SelectItem>
                      <SelectItem value="intermodale">Intermodale</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
                <div className="space-y-1.5">
                  <Label>Dest. Carico (opzionale)</Label>
                  <Select value={ruleForm.destinazione_carico_id || 'any'} onValueChange={setRuleCarico}>
                    <SelectTrigger><SelectValue placeholder="Qualsiasi" /></SelectTrigger>
                    <SelectContent>
                      <SelectItem value="any">Qualsiasi</SelectItem>
                      {destinations.map(d => <SelectItem key={d.id} value={d.id}>{d.nome}</SelectItem>)}
                    </SelectContent>
                  </Select>
                </div>
                <div className="space-y-1.5">
                  <Label>Dest. Scarico (opzionale)</Label>
                  <Select value={ruleForm.destinazione_scarico_id || 'any'} onValueChange={setRuleScarico}>
                    <SelectTrigger><SelectValue placeholder="Qualsiasi" /></SelectTrigger>
                    <SelectContent>
                      <SelectItem value="any">Qualsiasi</SelectItem>
                      {destinations.map(d => <SelectItem key={d.id} value={d.id}>{d.nome}</SelectItem>)}
                    </SelectContent>
                  </Select>
                </div>
              </div>
              <Separator />
              <div className="grid grid-cols-1 md:grid-cols-3 gap-3">
                <div className="space-y-1.5">
                  <Label>Tariffa (€) *</Label>
                  <Input type="number" step="0.01" value={ruleForm.tariffa} onChange={e => setRuleForm({...ruleForm, tariffa: Number(e.target.value)})} />
                </div>
                <div className="space-y-1.5">
                  <Label>Tipo Tariffa</Label>
                  <Select value={ruleForm.tipo_tariffa} onValueChange={v => setRuleForm({...ruleForm, tipo_tariffa: v})}>
                    <SelectTrigger><SelectValue /></SelectTrigger>
                    <SelectContent>
                      <SelectItem value="forfait">Forfait</SelectItem>
                      <SelectItem value="euro_kg">€/Kg</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
                <div className="space-y-1.5">
                  <Label>% Adeg. Carburante</Label>
                  <Input type="number" step="0.1" value={ruleForm.perc_adeguamento_carburante} onChange={e => setRuleForm({...ruleForm, perc_adeguamento_carburante: Number(e.target.value)})} />
                </div>
              </div>
              <div className="grid grid-cols-1 md:grid-cols-4 gap-3">
                <div className="space-y-1.5"><Label>Peso Min.</Label><Input type="number" value={ruleForm.range_peso_min} onChange={e => setRuleForm({...ruleForm, range_peso_min: Number(e.target.value)})} /></div>
                <div className="space-y-1.5"><Label>Peso Max.</Label><Input type="number" value={ruleForm.range_peso_max} onChange={e => setRuleForm({...ruleForm, range_peso_max: Number(e.target.value)})} /></div>
                <div className="space-y-1.5">
                  <Label>Unità Peso</Label>
                  <Select value={ruleForm.unita_peso} onValueChange={v => setRuleForm({...ruleForm, unita_peso: v})}>
                    <SelectTrigger><SelectValue /></SelectTrigger>
                    <SelectContent>
                      <SelectItem value="Kg">Kg</SelectItem>
                      <SelectItem value="quintali">Quintali</SelectItem>
                      <SelectItem value="ton">Tonnellate</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
                <div className="space-y-1.5"><Label>Min. Tassabile</Label><Input type="number" value={ruleForm.minimo_tassabile} onChange={e => setRuleForm({...ruleForm, minimo_tassabile: Number(e.target.value)})} /></div>
              </div>
            </div>
            <DialogFooter className="gap-2">
              <Button variant="outline" onClick={() => setRuleDialogOpen(false)}>Annulla</Button>
              <Button onClick={handleSaveRule} disabled={savingRule} data-testid="pricelist-rule-submit">
                {savingRule && <Loader2 className="h-4 w-4 animate-spin mr-2" />}
                {editingItemId ? 'Salva Modifiche' : 'Aggiungi Regola'}
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </div>
    );
  }

  // ===========================
  // VISTA LISTA LISTINI
  // ===========================
  return (
    <div className="space-y-3" data-testid="pricelists-page">
      <div className="flex flex-col gap-2 lg:flex-row lg:items-center lg:justify-between">
        <div className="relative max-w-sm">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-muted-foreground" />
          <Input placeholder="Cerca per cliente..." value={search} onChange={e => setSearch(e.target.value)} className="pl-9 h-9 text-sm" />
        </div>
        <Button size="sm" onClick={openNew} className="text-xs gap-1.5" data-testid="pricelist-new-button">
          <Plus className="h-3.5 w-3.5" /> Nuovo Listino
        </Button>
      </div>

      <Card className="rounded-xl border shadow-sm">
        <div className="overflow-x-auto">
          <Table className="text-xs md:text-sm">
            <TableHeader>
              <TableRow>
                <TableHead className="py-2 text-xs">Cliente</TableHead>
                <TableHead className="py-2 text-xs">Da</TableHead>
                <TableHead className="py-2 text-xs">A</TableHead>
                <TableHead className="py-2 text-xs">Regole</TableHead>
                <TableHead className="py-2 text-xs">In uso</TableHead>
                <TableHead className="py-2 text-xs w-24"></TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {loading ? Array.from({ length: 3 }).map((_, i) => (
                <TableRow key={`skel-row-${i}`}>{Array.from({ length: 6 }).map((_, j) => <TableCell key={`skel-col-${j}`} className="py-2"><Skeleton className="h-4 w-full" /></TableCell>)}</TableRow>
              )) : filtered.length === 0 ? (
                <TableRow><TableCell colSpan={6} className="text-center py-8 text-muted-foreground">Nessun listino</TableCell></TableRow>
              ) : filtered.map(l => (
                <TableRow key={l.id} className="hover:bg-muted/60 cursor-pointer" onClick={() => openDetail(l)}>
                  <TableCell className="py-2 font-medium">{l.cliente_nome}</TableCell>
                  <TableCell className="py-2 whitespace-nowrap">{l.data_inizio}</TableCell>
                  <TableCell className="py-2 whitespace-nowrap">{l.data_fine}</TableCell>
                  <TableCell className="py-2"><Badge variant="outline" className="text-[10px]">{l.items?.length || 0} regole</Badge></TableCell>
                  <TableCell className="py-2">{l.in_uso ? <Badge className="bg-emerald-100 text-emerald-800 text-[10px]">In uso</Badge> : <Badge variant="outline" className="text-[10px]">No</Badge>}</TableCell>
                  <TableCell className="py-2">
                    <div className="flex gap-1">
                      <Button variant="ghost" size="icon" className="h-7 w-7" onClick={e => { e.stopPropagation(); openDetail(l); }} data-testid="pricelist-detail-button"><Eye className="h-3 w-3" /></Button>
                      <Button variant="ghost" size="icon" className="h-7 w-7 text-destructive" onClick={e => { e.stopPropagation(); handleDelete(l.id); }}><Trash2 className="h-3 w-3" /></Button>
                    </div>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
      </Card>

      {/* Nuovo Listino */}
      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <DialogContent className="max-w-lg">
          <DialogHeader><DialogTitle style={{ fontFamily: "'Space Grotesk', sans-serif" }}>Nuovo Listino</DialogTitle></DialogHeader>
          <div className="space-y-4">
            <div className="space-y-1.5">
              <Label>Cliente *</Label>
              <Select value={form.cliente_id} onValueChange={setCustomer}>
                <SelectTrigger><SelectValue placeholder="Seleziona cliente" /></SelectTrigger>
                <SelectContent>{customers.map(c => <SelectItem key={c.id} value={c.id}>{c.ragione_sociale}</SelectItem>)}</SelectContent>
              </Select>
            </div>
            <div className="grid grid-cols-2 gap-3">
              <div className="space-y-1.5"><Label>Data Inizio</Label><Input type="date" value={form.data_inizio} onChange={e => setForm({ ...form, data_inizio: e.target.value })} /></div>
              <div className="space-y-1.5"><Label>Data Fine</Label><Input type="date" value={form.data_fine} onChange={e => setForm({ ...form, data_fine: e.target.value })} /></div>
            </div>
            <div className="space-y-1.5"><Label>Note</Label><Textarea value={form.note} onChange={e => setForm({ ...form, note: e.target.value })} rows={2} /></div>
          </div>
          <DialogFooter className="gap-2">
            <Button variant="outline" onClick={() => setDialogOpen(false)}>Annulla</Button>
            <Button onClick={handleSave} disabled={saving}>{saving && <Loader2 className="h-4 w-4 animate-spin mr-2" />} Crea Listino</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
