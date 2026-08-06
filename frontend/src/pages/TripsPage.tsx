import { useState, useEffect, useCallback, useMemo } from 'react';
import { getTrips, createTrip, getTrip, startTrip, completeTrip, recomputeTripSegments, downloadTripInstructionsPdf, getOrders, getVehicles, getDrivers, getCarriers, getGarages } from '@/lib/api';
import { getApiErrorMessage } from '@/lib/apiError';
import type { DtoTripResponse, DtoTripDetailResponse, DtoTripRequest, DtoOrderResponse, DtoVehicleResponse, DtoDriverResponse, DtoCarrierResponse, DtoGarageResponse } from '@/api/data-contracts';
import { FileText } from 'lucide-react';
import { Card } from '@/components/ui/card';
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
import { Plus, Eye, PlayCircle, CheckCircle, Loader2 } from 'lucide-react';

const emptyForm: DtoTripRequest = {
  targa_motrice: '', targa_rimorchio: '', autista_id: '',
  vettore_id: '', garage_id: '',
  note: '', data_partenza: '', data_arrivo: '',
};

export default function TripsPage() {
  const [trips, setTrips] = useState<DtoTripResponse[]>([]);
  const [loading, setLoading] = useState(true);
  const [newDialogOpen, setNewDialogOpen] = useState(false);
  const [detailOpen, setDetailOpen] = useState(false);
  const [selectedTrip, setSelectedTrip] = useState<DtoTripDetailResponse | null>(null);
  const [saving, setSaving] = useState(false);

  // Lookup data
  const [planOrders, setPlanOrders] = useState<DtoOrderResponse[]>([]);
  const [vehicles, setVehicles] = useState<DtoVehicleResponse[]>([]);
  const [drivers, setDrivers] = useState<DtoDriverResponse[]>([]);
  const [carriers, setCarriers] = useState<DtoCarrierResponse[]>([]);
  const [garages, setGarages] = useState<DtoGarageResponse[]>([]);
  const [selectedOrderIds, setSelectedOrderIds] = useState<string[]>([]);

  const [form, setForm] = useState<DtoTripRequest>(emptyForm);

  const fetchTrips = useCallback(() => {
    setLoading(true);
    getTrips().then((r: { data: DtoTripResponse[] }) => setTrips(r.data)).catch((err: unknown) => logger.error('Errore caricamento viaggi:', err)).finally(() => setLoading(false));
  }, []);
  useEffect(() => { fetchTrips(); }, [fetchTrips]);

  const openNew = () => {
    setForm(emptyForm);
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
    } catch (e) { toast.error(getApiErrorMessage(e) || 'Errore'); } finally { setSaving(false); }
  };

  const openDetail = async (trip: { id?: string }) => {
    if (!trip.id) return;
    try {
      const res = await getTrip(trip.id);
      setSelectedTrip(res.data);
      setDetailOpen(true);
    } catch (e) { toast.error('Errore caricamento dettaglio'); }
  };

  const handleStart = async (id?: string) => {
    if (!id) return;
    if (!window.confirm('Avviare questo viaggio? I suoi ordini pianificati passeranno a "in viaggio".')) return;
    try {
      await startTrip(id);
      toast.success('Viaggio avviato');
      fetchTrips();
      if (selectedTrip?.id === id) openDetail({ id });
    } catch (e) { toast.error(getApiErrorMessage(e) || 'Errore'); }
  };

  const handleComplete = async (id?: string) => {
    if (!id) return;
    if (!window.confirm('Completare questo viaggio? Gli ordini verranno chiusi.')) return;
    try {
      await completeTrip(id);
      toast.success('Viaggio completato, ordini chiusi');
      fetchTrips();
      setDetailOpen(false);
    } catch (e) { toast.error(getApiErrorMessage(e) || 'Errore'); }
  };

  const toggleOrder = (id: string) => {
    setSelectedOrderIds(prev => prev.includes(id) ? prev.filter(x => x !== id) : [...prev, id]);
  };

  const setDriver = (id: string) => setForm({ ...form, autista_id: id });
  const setVettore = (id: string) => setForm({ ...form, vettore_id: id });
  const setGarage = (id: string) => setForm({ ...form, garage_id: id });

  const pianificatoCount = useMemo(() => trips.filter(t => t.stato === 'PIANIFICATO').length, [trips]);
  const inCorsoCount = useMemo(() => trips.filter(t => t.stato === 'IN_CORSO').length, [trips]);
  const completatiCount = useMemo(() => trips.filter(t => t.stato === 'COMPLETATO').length, [trips]);

  // Liste pre-filtrate per il dialog nuovo viaggio
  const motriciFiltrate = useMemo(() => vehicles.filter(v => v.tipo_veicolo === 'motrice'), [vehicles]);
  const rimorchiFiltrati = useMemo(() => vehicles.filter(v => v.tipo_veicolo !== 'motrice'), [vehicles]);

  return (
    <div className="space-y-3" data-testid="trips-page">
      <div className="flex justify-between items-center">
        <div className="flex gap-2">
          <Badge className="status-order-yellow border text-xs">{pianificatoCount} pianificati</Badge>
          <Badge className="status-order-blue border text-xs">{inCorsoCount} in corso</Badge>
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
              {loading ? Array.from({ length: 3 }).map((_, i) => (
                <TableRow key={`skel-row-${i}`}>{Array.from({ length: 9 }).map((_, j) => <TableCell key={`skel-col-${j}`} className="py-2"><Skeleton className="h-4 w-full" /></TableCell>)}</TableRow>
              )) : trips.length === 0 ? (
                <TableRow><TableCell colSpan={9} className="text-center py-8 text-muted-foreground">Nessun viaggio. Crea un viaggio dal planner.</TableCell></TableRow>
              ) : trips.map(t => (
                <TableRow key={t.id} className="hover:bg-muted/60">
                  <TableCell className="py-2 font-mono text-xs">{t.id?.substring(0, 8)}</TableCell>
                  <TableCell className="py-2 font-mono">{t.targa_motrice || '-'}</TableCell>
                  <TableCell className="py-2 font-mono">{t.targa_rimorchio || '-'}</TableCell>
                  <TableCell className="py-2">{t.autista ? `${t.autista.nome} ${t.autista.cognome}` : '-'}</TableCell>
                  <TableCell className="py-2">{t.garage?.nome || '-'}</TableCell>
                  <TableCell className="py-2">{t.ordini_ids?.length || 0}</TableCell>
                  <TableCell className="py-2 text-right tabular-nums">{t.km_totali || 0}</TableCell>
                  <TableCell className="py-2"><StatusBadge stato={t.stato} /></TableCell>
                  <TableCell className="py-2">
                    <div className="flex gap-1">
                      <Button variant="ghost" size="icon" className="h-7 w-7" onClick={() => openDetail(t)}><Eye className="h-3 w-3" /></Button>
                      {t.stato === 'PIANIFICATO' && <Button variant="ghost" size="icon" className="h-7 w-7" onClick={() => handleStart(t.id)} title="Avvia"><PlayCircle className="h-3 w-3" /></Button>}
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
                    <Checkbox checked={!!o.id && selectedOrderIds.includes(o.id)} onCheckedChange={() => o.id && toggleOrder(o.id)} />
                    <span className="font-mono text-xs">{o.progressivo}</span>
                    <span className="truncate">{o.destinazione_carico?.nome} → {o.destinazione_scarico?.nome}</span>
                    <span className="ml-auto text-xs text-muted-foreground">{o.cliente?.ragione_sociale}</span>
                  </label>
                ))}
              </div>
            </div>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
              <div className="space-y-1.5"><Label>Motrice</Label><Select value={form.targa_motrice} onValueChange={v => setForm({ ...form, targa_motrice: v })}><SelectTrigger><SelectValue placeholder="Motrice" /></SelectTrigger><SelectContent>{motriciFiltrate.map(v => <SelectItem key={v.id} value={v.targa || ''}>{v.targa}</SelectItem>)}</SelectContent></Select></div>
              <div className="space-y-1.5"><Label>Rimorchio</Label><Select value={form.targa_rimorchio} onValueChange={v => setForm({ ...form, targa_rimorchio: v })}><SelectTrigger><SelectValue placeholder="Rimorchio" /></SelectTrigger><SelectContent>{rimorchiFiltrati.map(v => <SelectItem key={v.id} value={v.targa || ''}>{v.targa}</SelectItem>)}</SelectContent></Select></div>
              <div className="space-y-1.5"><Label>Autista</Label><Select value={form.autista_id} onValueChange={setDriver}><SelectTrigger><SelectValue placeholder="Autista" /></SelectTrigger><SelectContent>{drivers.map(d => <SelectItem key={d.id} value={d.id || ''}>{d.nome} {d.cognome}</SelectItem>)}</SelectContent></Select></div>
              <div className="space-y-1.5"><Label>Vettore</Label><Select value={form.vettore_id} onValueChange={setVettore}><SelectTrigger><SelectValue placeholder="Vettore" /></SelectTrigger><SelectContent>{carriers.map(c => <SelectItem key={c.id} value={c.id || ''}>{c.ragione_sociale}</SelectItem>)}</SelectContent></Select></div>
              <div className="space-y-1.5"><Label>Garage Base</Label><Select value={form.garage_id} onValueChange={setGarage}><SelectTrigger><SelectValue placeholder="Garage" /></SelectTrigger><SelectContent>{garages.map(g => <SelectItem key={g.id} value={g.id || ''}>{g.nome}</SelectItem>)}</SelectContent></Select></div>
              <div className="space-y-1.5"><Label>Data Partenza</Label><Input type="date" value={form.data_partenza} onChange={e => setForm({ ...form, data_partenza: e.target.value })} /></div>
            </div>
            <div className="space-y-1.5"><Label>Note</Label><Textarea value={form.note} onChange={e => setForm({ ...form, note: e.target.value })} rows={2} /></div>
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
              <div className="flex justify-between"><span className="text-muted-foreground">Autista:</span><span>{selectedTrip.autista ? `${selectedTrip.autista.nome} ${selectedTrip.autista.cognome}` : '-'}</span></div>
              <div className="flex justify-between"><span className="text-muted-foreground">Garage:</span><span>{selectedTrip.garage?.nome || '-'}</span></div>
              <div className="flex justify-between"><span className="text-muted-foreground">Stato:</span><StatusBadge stato={selectedTrip.stato} /></div>
              {selectedTrip.ordini && selectedTrip.ordini.length > 0 && (
                <div>
                  <Label className="mb-2 block">Ordini associati:</Label>
                  <div className="border rounded-lg divide-y">
                    {selectedTrip.ordini.map(o => (
                      <div key={o.id} className="px-3 py-2 flex justify-between items-center">
                        <div>
                          <span className="font-mono text-xs">{o.progressivo}</span>
                          <span className="ml-2">{o.destinazione_carico?.nome} → {o.destinazione_scarico?.nome}</span>
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
                        if (!selectedTrip.id) return;
                        try {
                          await recomputeTripSegments(selectedTrip.id);
                          const refreshed = await getTrip(selectedTrip.id);
                          setSelectedTrip(refreshed.data);
                          toast.success('Segmenti ricalcolati');
                        } catch (e) {
                          toast.error(getApiErrorMessage(e) || 'Errore ricalcolo');
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
                          <span className="ml-2">{s.origine_nome} → {s.destinazione_nome}</span>
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
                    if (!selectedTrip.id) return;
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
                      toast.error(getApiErrorMessage(e) || 'Errore download istruzioni');
                    }
                  }}
                >
                  <FileText className="h-4 w-4 mr-2" /> Istruzioni PDF
                </Button>
                {selectedTrip.stato === 'PIANIFICATO' && (
                  <Button onClick={() => handleStart(selectedTrip.id)} data-testid="trip-start-button">
                    <PlayCircle className="h-4 w-4 mr-2" /> Avvia Viaggio
                  </Button>
                )}
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
