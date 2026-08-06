import { useState } from 'react';
import {
  useGetMyOrdersQuery,
  useCreateMyOrderMutation,
  useDeleteMyOrderMutation,
  useGetDestinationsQuery,
  useCreateMyDestinationMutation,
} from '@/store/api/appApi';
import { getMutationErrorMessage } from '@/store/api/rtkQueryHelpers';
import type { DtoOrderRequest, DtoOrderResponse, DtoDestinationRequest } from '@/api/data-contracts';
import { DataTable, type DataTableColumn } from '@/components/shared/DataTable';
import { FormDialog } from '@/components/shared/FormDialog';
import { StatusBadge } from '@/components/shared/StatusBadge';
import { AddressSearchInput } from '@/components/shared/AddressSearchInput';
import { MapPicker } from '@/components/shared/MapPicker';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import { Select, SelectTrigger, SelectValue, SelectContent, SelectItem } from '@/components/ui/select';
import { TableRow, TableCell } from '@/components/ui/table';
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { formatEuro } from '@/lib/format';
import { toast } from 'sonner';
import { Eye, Trash2, Plus } from 'lucide-react';

// cliente_id è richiesto dal tipo generato ma il backend lo sovrascrive
// sempre con l'id del cliente autenticato (vedi OrderHandler.CreateMyOrder)
// — qui resta vuoto, nessun campo lo mostra in form.
const emptyForm: DtoOrderRequest = {
  cliente_id: '',
  destinazione_carico_id: '', destinazione_scarico_id: '',
  data_ritiro: '', ora_ritiro_da: '', ora_ritiro_a: '',
  data_consegna: '', ora_consegna_da: '', ora_consegna_a: '',
  tariffa: 0, rif_ordine_cliente: '', note: '',
};

type DestinationTarget = 'carico' | 'scarico';

const emptyDestinationForm: DtoDestinationRequest = { nome: '', indirizzo: '', citta: '', lat: null, lng: null };

export default function ClientOrdersPage() {
  const [dialogOpen, setDialogOpen] = useState(false);
  const [form, setForm] = useState<DtoOrderRequest>(emptyForm);
  const [detailOrder, setDetailOrder] = useState<DtoOrderResponse | null>(null);

  const [destDialogOpen, setDestDialogOpen] = useState(false);
  const [destTarget, setDestTarget] = useState<DestinationTarget | null>(null);
  const [destForm, setDestForm] = useState<DtoDestinationRequest>(emptyDestinationForm);
  const [destFlySignal, setDestFlySignal] = useState(0);

  const { data = [], isLoading: loading } = useGetMyOrdersQuery();
  // Pool condiviso con lo staff: la mutation createMyDestination invalida il
  // tag 'Destination', questa query lo fornisce — nessun refetch manuale
  // necessario dopo aver creato una nuova destinazione qui sotto.
  const { data: destinations = [] } = useGetDestinationsQuery();
  const [createMyOrder, { isLoading: saving }] = useCreateMyOrderMutation();
  const [deleteMyOrder] = useDeleteMyOrderMutation();
  const [createMyDestination, { isLoading: destSaving }] = useCreateMyDestinationMutation();

  const openNew = () => { setForm(emptyForm); setDialogOpen(true); };

  const openNewDestination = (target: DestinationTarget) => {
    setDestTarget(target);
    setDestForm(emptyDestinationForm);
    setDestDialogOpen(true);
  };

  const handleSaveDestination = async () => {
    if (destForm.lat == null || destForm.lng == null) { toast.error('Seleziona un punto sulla mappa'); return; }
    try {
      const created = await createMyDestination(destForm).unwrap();
      toast.success('Destinazione creata');
      if (destTarget === 'carico') setForm(f => ({ ...f, destinazione_carico_id: created.id || '' }));
      else if (destTarget === 'scarico') setForm(f => ({ ...f, destinazione_scarico_id: created.id || '' }));
      setDestDialogOpen(false);
    } catch (e) { toast.error(getMutationErrorMessage(e) || 'Errore'); }
  };

  const handleSave = async () => {
    try {
      await createMyOrder(form).unwrap();
      toast.success('Ordine creato');
      setDialogOpen(false);
    } catch (e) { toast.error(getMutationErrorMessage(e) || 'Errore'); }
  };

  // Il backend accetta la delete solo se l'ordine è ancora PIANIFICABILE
  // ("da pianificare") — una volta assegnato, il cliente non può più
  // ritirarlo da qui (vedi OrderHandler.DeleteMyOrder).
  const handleDelete = async (id?: string) => {
    if (!id) return;
    if (!window.confirm('Eliminare questo ordine?')) return;
    try { await deleteMyOrder(id).unwrap(); toast.success('Ordine eliminato'); }
    catch (e) { toast.error(getMutationErrorMessage(e) || 'Errore'); }
  };

  const columns: DataTableColumn[] = [
    { key: 'progressivo', label: 'Numero' },
    { key: 'carico', label: 'Carico' },
    { key: 'scarico', label: 'Scarico' },
    { key: 'data_ritiro', label: 'Data Ritiro' },
    { key: 'tariffa', label: 'Tariffa', className: 'text-right' },
    { key: 'stato', label: 'Stato' },
    { key: 'actions', label: '', className: 'w-12' },
  ];

  return (
    <div data-testid="client-orders-page">
      <DataTable
        columns={columns}
        data={data}
        loading={loading}
        searchValue=""
        onSearchChange={() => {}}
        onAdd={openNew}
        addLabel="Nuovo Ordine"
        testId="client-orders-table"
        emptyMessage="Nessun ordine ancora creato."
        renderRow={(item: DtoOrderResponse) => (
          <TableRow key={item.id} className="hover:bg-muted/60">
            <TableCell className="py-2 font-mono font-medium">{item.progressivo}</TableCell>
            <TableCell className="py-2">{item.destinazione_carico?.nome || '—'}</TableCell>
            <TableCell className="py-2">{item.destinazione_scarico?.nome || '—'}</TableCell>
            <TableCell className="py-2 whitespace-nowrap">{item.data_ritiro || '—'}</TableCell>
            <TableCell className="py-2 text-right tabular-nums">€ {formatEuro(item.tariffa || 0)}</TableCell>
            <TableCell className="py-2"><StatusBadge stato={item.stato} /></TableCell>
            <TableCell className="py-2">
              <div className="flex gap-1">
                <Button variant="ghost" size="icon" className="h-7 w-7" onClick={() => setDetailOrder(item)}>
                  <Eye className="h-3 w-3" />
                </Button>
                {item.stato === 'PIANIFICABILE' && (
                  <Button variant="ghost" size="icon" className="h-7 w-7 text-destructive" onClick={() => handleDelete(item.id)}>
                    <Trash2 className="h-3 w-3" />
                  </Button>
                )}
              </div>
            </TableCell>
          </TableRow>
        )}
      />

      <Dialog open={!!detailOrder} onOpenChange={(open) => !open && setDetailOrder(null)}>
        <DialogContent className="max-w-lg">
          <DialogHeader><DialogTitle style={{ fontFamily: "'Space Grotesk', sans-serif" }}>Dettaglio Ordine</DialogTitle></DialogHeader>
          {detailOrder && (
            <div className="space-y-3 text-sm">
              <div className="flex justify-between"><span className="text-muted-foreground">Progressivo:</span><span className="font-mono font-medium">{detailOrder.progressivo}</span></div>
              <div className="flex justify-between"><span className="text-muted-foreground">Carico:</span><span>{detailOrder.destinazione_carico?.nome || '—'}</span></div>
              <div className="flex justify-between"><span className="text-muted-foreground">Scarico:</span><span>{detailOrder.destinazione_scarico?.nome || '—'}</span></div>
              <div className="flex justify-between"><span className="text-muted-foreground">Data Ritiro:</span><span>{detailOrder.data_ritiro || '—'} {detailOrder.ora_ritiro_da}{detailOrder.ora_ritiro_a ? `-${detailOrder.ora_ritiro_a}` : ''}</span></div>
              <div className="flex justify-between"><span className="text-muted-foreground">Data Consegna:</span><span>{detailOrder.data_consegna || '—'} {detailOrder.ora_consegna_da}{detailOrder.ora_consegna_a ? `-${detailOrder.ora_consegna_a}` : ''}</span></div>
              <div className="flex justify-between"><span className="text-muted-foreground">Tariffa:</span><span className="font-medium">€ {formatEuro(detailOrder.tariffa || 0)}</span></div>
              <div className="flex justify-between"><span className="text-muted-foreground">Stato:</span><StatusBadge stato={detailOrder.stato} /></div>
              {detailOrder.autista && <div className="flex justify-between"><span className="text-muted-foreground">Autista:</span><span>{detailOrder.autista.nome} {detailOrder.autista.cognome}</span></div>}
              {detailOrder.vettore && <div className="flex justify-between"><span className="text-muted-foreground">Vettore:</span><span>{detailOrder.vettore.ragione_sociale}</span></div>}
              {detailOrder.note && <div><span className="text-muted-foreground">Note:</span><p className="mt-1">{detailOrder.note}</p></div>}
            </div>
          )}
        </DialogContent>
      </Dialog>

      <FormDialog open={dialogOpen} onClose={setDialogOpen} title="Nuovo Ordine" onSubmit={handleSave} loading={saving} submitLabel="Crea Ordine">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
          <div className="space-y-1.5">
            <Label>Destinazione Carico</Label>
            <div className="flex gap-2">
              <Select value={form.destinazione_carico_id} onValueChange={v => setForm({ ...form, destinazione_carico_id: v })}>
                <SelectTrigger><SelectValue placeholder="Seleziona..." /></SelectTrigger>
                <SelectContent>{destinations.map(d => <SelectItem key={d.id} value={d.id || ''}>{d.nome}</SelectItem>)}</SelectContent>
              </Select>
              <Button type="button" variant="outline" size="icon" className="shrink-0" title="Nuova destinazione" onClick={() => openNewDestination('carico')}>
                <Plus className="h-4 w-4" />
              </Button>
            </div>
          </div>
          <div className="space-y-1.5">
            <Label>Destinazione Scarico</Label>
            <div className="flex gap-2">
              <Select value={form.destinazione_scarico_id} onValueChange={v => setForm({ ...form, destinazione_scarico_id: v })}>
                <SelectTrigger><SelectValue placeholder="Seleziona..." /></SelectTrigger>
                <SelectContent>{destinations.map(d => <SelectItem key={d.id} value={d.id || ''}>{d.nome}</SelectItem>)}</SelectContent>
              </Select>
              <Button type="button" variant="outline" size="icon" className="shrink-0" title="Nuova destinazione" onClick={() => openNewDestination('scarico')}>
                <Plus className="h-4 w-4" />
              </Button>
            </div>
          </div>
          <div className="space-y-1.5"><Label>Data Ritiro</Label><Input type="date" value={form.data_ritiro} onChange={e => setForm({ ...form, data_ritiro: e.target.value })} /></div>
          <div className="space-y-1.5"><Label>Data Consegna</Label><Input type="date" value={form.data_consegna} onChange={e => setForm({ ...form, data_consegna: e.target.value })} /></div>
          <div className="space-y-1.5">
            <Label>Orario Ritiro (da-a)</Label>
            <div className="flex gap-2">
              <Input type="time" value={form.ora_ritiro_da} onChange={e => setForm({ ...form, ora_ritiro_da: e.target.value })} />
              <Input type="time" value={form.ora_ritiro_a} onChange={e => setForm({ ...form, ora_ritiro_a: e.target.value })} />
            </div>
          </div>
          <div className="space-y-1.5">
            <Label>Orario Consegna (da-a)</Label>
            <div className="flex gap-2">
              <Input type="time" value={form.ora_consegna_da} onChange={e => setForm({ ...form, ora_consegna_da: e.target.value })} />
              <Input type="time" value={form.ora_consegna_a} onChange={e => setForm({ ...form, ora_consegna_a: e.target.value })} />
            </div>
          </div>
          <div className="space-y-1.5"><Label>Tariffa desiderata (€)</Label><Input type="number" value={form.tariffa} onChange={e => setForm({ ...form, tariffa: Number(e.target.value) })} /></div>
          <div className="space-y-1.5"><Label>Rif. Vostro Ordine</Label><Input value={form.rif_ordine_cliente} onChange={e => setForm({ ...form, rif_ordine_cliente: e.target.value })} /></div>
          <div className="md:col-span-2 space-y-1.5"><Label>Note</Label><Textarea value={form.note} onChange={e => setForm({ ...form, note: e.target.value })} rows={2} /></div>
        </div>
      </FormDialog>

      <FormDialog
        open={destDialogOpen}
        onClose={setDestDialogOpen}
        title="Nuova Destinazione"
        onSubmit={handleSaveDestination}
        loading={destSaving}
        submitLabel="Crea"
      >
        <div className="space-y-3">
          <div className="space-y-1.5"><Label>Nome *</Label><Input value={destForm.nome} onChange={e => setDestForm({ ...destForm, nome: e.target.value })} required /></div>
          <div className="space-y-1.5">
            <Label>Indirizzo</Label>
            <AddressSearchInput
              value={destForm.indirizzo || ''}
              onChange={(v) => setDestForm(f => ({ ...f, indirizzo: v }))}
              onSelect={(r) => {
                setDestForm(f => ({ ...f, indirizzo: r.indirizzo, citta: r.citta || f.citta, lat: r.lat, lng: r.lng }));
                setDestFlySignal(s => s + 1);
              }}
            />
          </div>
          <div className="space-y-1.5"><Label>Città</Label><Input value={destForm.citta || ''} readOnly className="bg-muted/50" /></div>
          <div className="space-y-1.5">
            <Label>Posizione *</Label>
            <MapPicker
              lat={destForm.lat ?? null} lng={destForm.lng ?? null}
              onChange={(lat, lng) => setDestForm({ ...destForm, lat, lng })}
              flyToSignal={destFlySignal}
            />
          </div>
        </div>
      </FormDialog>
    </div>
  );
}
