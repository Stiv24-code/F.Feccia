import { useState, useEffect, useCallback, useMemo } from 'react';
import { getTrips, createTrip, getTrip, completeTrip, recomputeTripSegments, downloadTripInstructionsPdf, getOrders, getVehicles, getDrivers, getCarriers, getGarages } from '@/lib/api';
import { FileText } from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Label } from '@/components/ui/label';
import { Select, SelectTrigger, SelectValue, SelectContent, SelectItem } from '@/components/ui/select';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog';
import { Badge } from '@/components/ui/badge';
import { Skeleton } from '@/components/ui/skeleton';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import { Checkbox } from '@/components/ui/checkbox';
import { StatusBadge } from '@/components/shared/StatusBadge';
import { toast } from 'sonner';
import { logger } from '@/lib/logger';
import { Plus, Route, Eye, CheckCircle, Loader2 } from 'lucide-react';

export default function TripsPage() {
  const [trips, setTrips] = useState([]);
  const [loading, setLoading] = useState(true);
  const [newDialogOpen, setNewDialogOpen] = useState(false);
  const [detailOpen, setDetailOpen] = useState(false);
  const [selectedTrip, setSelectedTrip] = useState(null);
  const [saving, setSaving] = useState(false);

  // Lookup data
  const [planOrders, setPlanOrders] = useState([]);
  const [vehicles, setVehicles] = useState([]);
  const [drivers, setDrivers] = useState([]);
  const [carriers, setCarriers] = useState([]);
  const [garages, setGarages] = useState([]);
  const [selectedOrderIds, setSelectedOrderIds] = useState([]);

  const [form, setForm] = useState({
    targa_motrice: '', targa_rimorchio: '', autista_id: '', autista_nome: '',
    vettore_id: '', vettore_nome: '', garage_id: '', garage_nome: '',
    note: '', data_partenza: '', data_arrivo: ''
  });

  const fetchTrips = useCallback(() => { setLoading(true); getTrips().then(r => setTrips(r.data)).catch(err => logger.error('Errore caricamento viaggi:', err)).finally(() => setLoading(false)); }, []);
  useEffect(() => { fetchTrips(); }, [fetchTrips]);

  const openNew = () => {
    setForm({ targa_motrice: '', targa_rimorchio: '', autista_id: '', autista_nome: '', vettore_id: '', vettore_nome: '', garage_id: '', garage_nome: '', note: '', data_partenza: '', data_arrivo: '' });
    setSelectedOrderIds([]);
    Promise.all([getOrders({ stato: 'PIANIFICABILE' }), getVehicles(), getDrivers(), getCarriers(), getGarages()]).then(([o, v, d, c, g]) => {
      setPlanOrders(o.data); setVehicles(v.data); setDrivers(d.data); setCarriers(c.data); setGarages(g.data);
    });
    setNewDialogOpen(true);
  };

  const handleCreate = async () => {
    if (selectedOrderIds.length === 0) { toast.error('Selezionare almeno un ordine'); return; }
    setSaving(true);
    try {
      await createTrip({ ...form, ordini_ids: selectedOrderIds });
      toast.success('Viaggio creato');
      setNewDialogOpen(false);
      fetchTrips();
    } catch (e) { toast.error(e.response?.data?.detail || 'Errore'); } finally { setSaving(false); }
  };

  const openDetail = async (trip) => {
    try {
      const res = await getTrip(trip.id);
      setSelectedTrip(res.data);
      setDetailOpen(true);
    } catch (e) { toast.error('Errore caricamento dettaglio'); }
  };

  const handleComplete = async (id) => {
    if (!window.confirm('Completare questo viaggio? Gli ordini verranno chiusi.')) return;
    try {
      await completeTrip(id);
      toast.success('Viaggio completato, ordini chiusi');
      fetchTrips();
      setDetailOpen(false);
    } catch (e) { toast.error(e.response?.data?.detail || 'Errore'); }
  };

  const toggleOrder = (id) => {
    setSelectedOrderIds(prev => prev.includes(id) ? prev.filter(x => x !== id) : [...prev, id]);
  };

  const setDriver = (id) => {
    const d = drivers.find(x => x.id === id);
    setForm({ ...form, autista_id: id, autista_nome: d ? `${d.nome} ${d.cognome}` : '' });
  };
  const setVettore = (id) => {
    const c = carriers.find(x => x.id === id);
    setForm({ ...form, vettore_id: id, vettore_nome: c?.ragione_sociale || '' });
  };
  const setGarage = (id) => {
    const g = garages.find(x => x.id === id);
    setForm({ ...form, garage_id: id, garage_nome: g?.nome || '' });
  };

  const inCorsoCount = useMemo(() => trips.filter(t => t.stato === 'IN_CORSO').length, [trips]);
  const completatiCount = useMemo(() => trips.filter(t => t.stato === 'COMPLETATO').length, [trips]);

  // Liste pre-filtrate per il dialog nuovo viaggio
  const motriciFiltrate = useMemo(() => vehicles.filter(v => v.tipo_veicolo === 'motrice'), [vehicles]);
  const rimorchiFiltrati = useMemo(() => vehicles.filter(v => v.tipo_veicolo !== 'motrice'), [vehicles]);

  return (
    <div className="space-y-3" data-testid="trips-page">
      <div className="flex justify-between items-center">
        <div className="flex gap-2">
          <Badge className="status-viaggio border text-xs">{inCorsoCount} in corso</Badge>
          <Badge className="status-fatturato border text-xs">{completatiCount} completati</Badge>
        </div>
        <Button size="sm" onClick={openNew} className="text-xs gap-1.5" data-testid="trips-new-button">
          <Plus className="h-3.5 w-3.5" /> Nuovo Viaggio
        </Button>
      </div>

      <Card className="rounded-xl border shadow-sm" data-testid="trips-table">
        <div className="overflow-x-auto">
          <Table className="text-xs md:text-sm">
            <TableHeader>
              <TableRow>
                <TableHead className="py-2 text-xs">ID</TableHead>
                <TableHead className="py-2 text-xs">Motrice</TableHead>
                <TableHead className="py-2 text-xs">Rimorchio</TableHead>
                <TableHead className="py-2 text-xs">Autista</TableHead>
                <TableHead className="py-2 text-xs">Garage</TableHead>
                <TableHead className="py-2 text-xs">Ordini</TableHead>
                <TableHead className="py-2 text-xs text-right">Km</TableHead>
                <TableHead className="py-2 text-xs">Stato</TableHead>
                <TableHead className="py-2 text-xs">Azioni</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {loading ? Array.from({length: 3}).map((_, i) => (
                <TableRow key={`skel-row-${i}`}>{Array.from({length: 9}).map((_, j) => <TableCell key={`skel-col-${j}`} className="py-2"><Skeleton className="h-4 w-full" /></TableCell>)}</TableRow>
              )) : trips.length === 0 ? (
                <TableRow><TableCell colSpan={9} className="text-center py-8 text-muted-foreground">Nessun viaggio. Crea un viaggio dal planner.</TableCell></TableRow>
              ) : trips.map(t => (
                <TableRow key={t.id} className="hover:bg-muted/60">
                  <TableCell className="py-2 font-mono text-xs">{t.id?.substring(0, 8)}</TableCell>
                  <TableCell className="py-2 font-mono">{t.targa_motrice || '-'}</TableCell>
                  <TableCell className="py-2 font-mono">{t.targa_rimorchio || '-'}</TableCell>
                  <TableCell className="py-2">{t.autista_nome || '-'}</TableCell>
                  <TableCell className="py-2">{t.garage_nome || '-'}</TableCell>
                  <TableCell className="py-2">{t.ordini_ids?.length || 0}</TableCell>
                  <TableCell className="py-2 text-right tabular-nums">{t.km_totali || 0}</TableCell>
                  <TableCell className="py-2"><StatusBadge stato={t.stato} /></TableCell>
                  <TableCell className="py-2">
                    <div className="flex gap-1">
                      <Button variant="ghost" size="icon" className="h-7 w-7" onClick={() => openDetail(t)}><Eye className="h-3 w-3" /></Button>
                      {t.stato === 'IN_CORSO' && <Button variant="ghost" size="icon" className="h-7 w-7" onClick={() => handleComplete(t.id)} title="Completa"><CheckCircle className="h-3 w-3" /></Button>}
                    </div>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
      </Card>

      {/* New Trip Dialog */}
      <Dialog open={newDialogOpen} onOpenChange={setNewDialogOpen}>
        <DialogContent className="max-w-2xl max-h-[85vh] overflow-y-auto">
          <DialogHeader><DialogTitle style={{ fontFamily: "'Space Grotesk', sans-serif" }}>Nuovo Viaggio</DialogTitle></DialogHeader>
          <div className="space-y-4">
            <div>
              <Label className="mb-2 block">Seleziona Ordini da includere</Label>
              <div className="border rounded-lg max-h-40 overflow-y-auto">
                {planOrders.length === 0 ? (
                  <p className="text-sm text-muted-foreground p-3">Nessun ordine pianificabile disponibile</p>
                ) : planOrders.map(o => (
                  <label key={o.id} className="flex items-center gap-2 px-3 py-2 hover:bg-muted/50 cursor-pointer text-sm">
                    <Checkbox checked={selectedOrderIds.includes(o.id)} onCheckedChange={() => toggleOrder(o.id)} />
                    <span className="font-mono text-xs">{o.progressivo}</span>
                    <span className="truncate">{o.destinazione_carico_nome} → {o.destinazione_scarico_nome}</span>
                    <span className="ml-auto text-xs text-muted-foreground">{o.cliente_nome}</span>
                  </label>
                ))}
              </div>
            </div>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
              <div className="space-y-1.5"><Label>Motrice</Label><Select value={form.targa_motrice} onValueChange={v => setForm({...form, targa_motrice: v})}><SelectTrigger><SelectValue placeholder="Motrice" /></SelectTrigger><SelectContent>{motriciFiltrate.map(v => <SelectItem key={v.id} value={v.targa}>{v.targa}</SelectItem>)}</SelectContent></Select></div>
              <div className="space-y-1.5"><Label>Rimorchio</Label><Select value={form.targa_rimorchio} onValueChange={v => setForm({...form, targa_rimorchio: v})}><SelectTrigger><SelectValue placeholder="Rimorchio" /></SelectTrigger><SelectContent>{rimorchiFiltrati.map(v => <SelectItem key={v.id} value={v.targa}>{v.targa}</SelectItem>)}</SelectContent></Select></div>
              <div className="space-y-1.5"><Label>Autista</Label><Select value={form.autista_id} onValueChange={setDriver}><SelectTrigger><SelectValue placeholder="Autista" /></SelectTrigger><SelectContent>{drivers.map(d => <SelectItem key={d.id} value={d.id}>{d.nome} {d.cognome}</SelectItem>)}</SelectContent></Select></div>
              <div className="space-y-1.5"><Label>Vettore</Label><Select value={form.vettore_id} onValueChange={setVettore}><SelectTrigger><SelectValue placeholder="Vettore" /></SelectTrigger><SelectContent>{carriers.map(c => <SelectItem key={c.id} value={c.id}>{c.ragione_sociale}</SelectItem>)}</SelectContent></Select></div>
              <div className="space-y-1.5"><Label>Garage Base</Label><Select value={form.garage_id} onValueChange={setGarage}><SelectTrigger><SelectValue placeholder="Garage" /></SelectTrigger><SelectContent>{garages.map(g => <SelectItem key={g.id} value={g.id}>{g.nome}</SelectItem>)}</SelectContent></Select></div>
              <div className="space-y-1.5"><Label>Data Partenza</Label><Input type="date" value={form.data_partenza} onChange={e => setForm({...form, data_partenza: e.target.value})} /></div>
            </div>
            <div className="space-y-1.5"><Label>Note</Label><Textarea value={form.note} onChange={e => setForm({...form, note: e.target.value})} rows={2} /></div>
          </div>
          <DialogFooter className="gap-2">
            <Button variant="outline" onClick={() => setNewDialogOpen(false)}>Annulla</Button>
            <Button onClick={handleCreate} disabled={saving} data-testid="trip-create-button">
              {saving && <Loader2 className="h-4 w-4 animate-spin mr-2" />} Crea Viaggio
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Detail Dialog */}
      <Dialog open={detailOpen} onOpenChange={setDetailOpen}>
        <DialogContent className="max-w-lg">
          <DialogHeader><DialogTitle style={{ fontFamily: "'Space Grotesk', sans-serif" }}>Dettaglio Viaggio</DialogTitle></DialogHeader>
          {selectedTrip && (
            <div className="space-y-3 text-sm">
              <div className="flex justify-between"><span className="text-muted-foreground">Motrice:</span><span className="font-mono">{selectedTrip.targa_motrice || '-'}</span></div>
              <div className="flex justify-between"><span className="text-muted-foreground">Rimorchio:</span><span className="font-mono">{selectedTrip.targa_rimorchio || '-'}</span></div>
              <div className="flex justify-between"><span className="text-muted-foreground">Autista:</span><span>{selectedTrip.autista_nome || '-'}</span></div>
              <div className="flex justify-between"><span className="text-muted-foreground">Garage:</span><span>{selectedTrip.garage_nome || '-'}</span></div>
              <div className="flex justify-between"><span className="text-muted-foreground">Stato:</span><StatusBadge stato={selectedTrip.stato} /></div>
              {selectedTrip.ordini && selectedTrip.ordini.length > 0 && (
                <div>
                  <Label className="mb-2 block">Ordini associati:</Label>
                  <div className="border rounded-lg divide-y">
                    {selectedTrip.ordini.map(o => (
                      <div key={o.id} className="px-3 py-2 flex justify-between items-center">
                        <div>
                          <span className="font-mono text-xs">{o.progressivo}</span>
                          <span className="ml-2">{o.destinazione_carico_nome} → {o.destinazione_scarico_nome}</span>
                        </div>
                        <StatusBadge stato={o.stato} />
                      </div>
                    ))}
                  </div>
                </div>
              )}
              {selectedTrip.segmenti && selectedTrip.segmenti.length > 0 && (
                <div>
                  <div className="flex items-center justify-between mb-2">
                    <Label>Segmenti viaggio (#31):</Label>
                    <Button
                      variant="outline" size="sm" className="h-7 text-xs"
                      onClick={async () => {
                        try {
                          await recomputeTripSegments(selectedTrip.id);
                          const refreshed = await getTrip(selectedTrip.id);
                          setSelectedTrip(refreshed.data);
                          toast.success('Segmenti ricalcolati');
                        } catch (e) {
                          toast.error(e.response?.data?.detail || 'Errore ricalcolo');
                        }
                      }}
                    >
                      Ricalcola
                    </Button>
                  </div>
                  <div className="border rounded-lg divide-y text-xs">
                    {selectedTrip.segmenti.map((s, idx) => (
                      <div key={idx} className="px-3 py-2 flex justify-between items-center">
                        <div>
                          <span className="font-mono text-[10px] text-muted-foreground">[{s.tipo}]</span>
                          <span className="ml-2">{s.origine_nome || s.origine} → {s.destinazione_nome || s.destinazione}</span>
                        </div>
                        <span className="tabular-nums font-medium">{(s.km || 0).toFixed(1)} km</span>
                      </div>
                    ))}
                    <div className="px-3 py-2 flex justify-between items-center font-semibold bg-muted/30">
                      <span>Totale</span>
                      <span className="tabular-nums">{(selectedTrip.km_totali || 0).toFixed(1)} km</span>
                    </div>
                  </div>
                </div>
              )}
              <div className="grid grid-cols-1 md:grid-cols-2 gap-2 mt-2">
                <Button
                  variant="outline"
                  className="w-full"
                  onClick={async () => {
                    try {
                      const r = await downloadTripInstructionsPdf(selectedTrip.id);
                      const cd = r.headers?.['content-disposition'] || '';
                      const m = cd.match(/filename="?([^";]+)"?/i);
                      const filename = m ? m[1] : `istruzioni_${selectedTrip.id}.pdf`;
                      const url = window.URL.createObjectURL(new Blob([r.data], { type: 'application/pdf' }));
                      const a = document.createElement('a');
                      a.href = url; a.download = filename;
                      document.body.appendChild(a); a.click(); a.remove();
                      window.URL.revokeObjectURL(url);
                    } catch (e) {
                      toast.error(e.response?.data?.detail || 'Errore download istruzioni');
                    }
                  }}
                >
                  <FileText className="h-4 w-4 mr-2" /> Istruzioni PDF
                </Button>
                {selectedTrip.stato === 'IN_CORSO' && (
                  <Button onClick={() => handleComplete(selectedTrip.id)} data-testid="trip-complete-button">
                    <CheckCircle className="h-4 w-4 mr-2" /> Completa Viaggio
                  </Button>
                )}
              </div>
            </div>
          )}
        </DialogContent>
      </Dialog>
    </div>
  );
}
