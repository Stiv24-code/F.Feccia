import { useState } from 'react';
import {
  useGetWashStationsQuery,
  useCreateWashStationMutation,
  useUpdateWashStationMutation,
  useDeleteWashStationMutation,
} from '@/store/api/appApi';
import { getMutationErrorMessage } from '@/store/api/rtkQueryHelpers';
import type { DtoWashStationRequest, DtoWashStationResponse, DtoGeocodeResultDTO } from '@/api/data-contracts';
import { usePagination } from '@/hooks/use-pagination';
import { DataTable, type DataTableColumn } from '@/components/shared/DataTable';
import { FormDialog } from '@/components/shared/FormDialog';
import { MapPicker } from '@/components/shared/MapPicker';
import { AddressSearchInput } from '@/components/shared/AddressSearchInput';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Button } from '@/components/ui/button';
import { Switch } from '@/components/ui/switch';
import { Badge } from '@/components/ui/badge';
import { TableRow, TableCell } from '@/components/ui/table';
import { toast } from 'sonner';
import { Pencil, Trash2 } from 'lucide-react';

type WashStationForm = Omit<DtoWashStationRequest, 'lat' | 'lng'> & { lat: number | null; lng: number | null };

const PAGE_SIZE = 20;

const emptyForm: WashStationForm = { nome: '', tipo: '', indirizzo: '', citta: '', lat: null, lng: null, orario_da: '', orario_a: '', note: '', active: true };

export default function WashStationsPage() {
  const [dialogOpen, setDialogOpen] = useState(false);
  const [form, setForm] = useState<WashStationForm>(emptyForm);
  const [editId, setEditId] = useState<string | null>(null);
  const [flySignal, setFlySignal] = useState(0);

  const [page, setPage] = usePagination(null);
  const { data: result, isLoading: loading } = useGetWashStationsQuery({ includeInactive: true, page, limit: PAGE_SIZE });
  const data = result?.items ?? [];
  const totalPages = Math.max(1, Math.ceil((result?.total ?? 0) / PAGE_SIZE));
  const [createWashStation, { isLoading: creating }] = useCreateWashStationMutation();
  const [updateWashStation, { isLoading: updating }] = useUpdateWashStationMutation();
  const [deleteWashStation] = useDeleteWashStationMutation();
  const saving = creating || updating;

  const openNew = () => { setForm(emptyForm); setEditId(null); setDialogOpen(true); };
  const openEdit = (item: DtoWashStationResponse) => { setForm({ nome: item.nome || '', tipo: item.tipo || '', indirizzo: item.indirizzo || '', citta: item.citta || '', lat: item.lat ?? null, lng: item.lng ?? null, orario_da: item.orario_da || '', orario_a: item.orario_a || '', note: item.note || '', active: item.active }); setEditId(item.id || null); setDialogOpen(true); };

  const handleSave = async () => {
    if (form.lat == null || form.lng == null) { toast.error('Cerca e seleziona un indirizzo per impostare la posizione'); return; }
    try {
      const body: DtoWashStationRequest = { ...form, lat: form.lat, lng: form.lng };
      if (editId) { await updateWashStation({ id: editId, body }).unwrap(); toast.success('Punto di lavaggio aggiornato'); } else { await createWashStation(body).unwrap(); toast.success('Punto di lavaggio creato'); }
      setDialogOpen(false);
    } catch (e) { toast.error(getMutationErrorMessage(e) || 'Errore'); }
  };

  const handleDelete = async (id: string) => {
    if (!window.confirm('Eliminare?')) return;
    try { await deleteWashStation(id).unwrap(); toast.success('Eliminato'); } catch (e) { toast.error(getMutationErrorMessage(e) || 'Errore'); }
  };

  const columns: DataTableColumn[] = [{ key: 'nome', label: 'Nome' }, { key: 'tipo', label: 'Tipo' }, { key: 'citta', label: 'Città' }, { key: 'stato', label: 'Stato' }, { key: 'actions', label: '', className: 'w-20' }];

  return (
    <div data-testid="wash-stations-page">
      <DataTable columns={columns} data={data} loading={loading} searchValue="" onSearchChange={() => {}} onAdd={openNew} addLabel="Nuovo Punto di Lavaggio" testId="masterdata-table"
        page={page} totalPages={totalPages} onPageChange={setPage}
        renderRow={(item) => (
          <TableRow key={item.id} className={`hover:bg-muted/60 ${!item.active ? 'opacity-60' : ''}`}>
            <TableCell className="py-2 font-medium">{item.nome}</TableCell>
            <TableCell className="py-2 text-muted-foreground">{item.tipo || '—'}</TableCell>
            <TableCell className="py-2">{item.citta}</TableCell>
            <TableCell className="py-2">{item.active ? <Badge className="bg-emerald-100 text-emerald-800 text-[10px]">Attivo</Badge> : <Badge variant="outline" className="text-[10px]">Inattivo</Badge>}</TableCell>
            <TableCell className="py-2"><div className="flex gap-1"><Button variant="ghost" size="icon" className="h-7 w-7" onClick={() => openEdit(item)}><Pencil className="h-3 w-3" /></Button><Button variant="ghost" size="icon" className="h-7 w-7 text-destructive" onClick={() => item.id && handleDelete(item.id)}><Trash2 className="h-3 w-3" /></Button></div></TableCell>
          </TableRow>
        )}
      />
      <FormDialog open={dialogOpen} onClose={setDialogOpen} title={editId ? 'Modifica Punto di Lavaggio' : 'Nuovo Punto di Lavaggio'} onSubmit={handleSave} loading={saving}>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
          <div className="md:col-span-2 space-y-1.5"><Label>Nome *</Label><Input value={form.nome} onChange={e => setForm({ ...form, nome: e.target.value })} required /></div>
          <div className="md:col-span-2 space-y-1.5"><Label>Tipo di lavaggio</Label><Input value={form.tipo} onChange={e => setForm({ ...form, tipo: e.target.value })} placeholder="Es. Lavaggio alimentare EFTCO" /></div>
          <div className="md:col-span-2 space-y-1.5">
            <Label>Indirizzo *</Label>
            <AddressSearchInput
              value={form.indirizzo || ''}
              onChange={(v) => setForm(f => ({ ...f, indirizzo: v }))}
              onSelect={(r: DtoGeocodeResultDTO) => {
                setForm(f => ({ ...f, indirizzo: r.indirizzo || f.indirizzo, citta: r.citta || f.citta, lat: r.lat ?? f.lat, lng: r.lng ?? f.lng }));
                setFlySignal(s => s + 1);
              }}
              required
            />
          </div>
          <div className="space-y-1.5"><Label>Città</Label><Input value={form.citta} readOnly className="bg-muted/50" /></div>
          <div className="md:col-span-2 space-y-1.5">
            <Label>Posizione *</Label>
            <MapPicker
              lat={form.lat} lng={form.lng}
              flyToSignal={flySignal}
            />
          </div>
          <div className="space-y-1.5"><Label>Orario apertura dalle</Label><Input type="time" value={form.orario_da} onChange={e => setForm({ ...form, orario_da: e.target.value })} /></div>
          <div className="space-y-1.5"><Label>Orario apertura alle</Label><Input type="time" value={form.orario_a} onChange={e => setForm({ ...form, orario_a: e.target.value })} /></div>
          {editId && (
            <div className="md:col-span-2 flex items-center gap-2">
              <Switch checked={form.active} onCheckedChange={active => setForm({ ...form, active })} id="wash-station-active" />
              <Label htmlFor="wash-station-active">Attivo</Label>
            </div>
          )}
        </div>
      </FormDialog>
    </div>
  );
}
