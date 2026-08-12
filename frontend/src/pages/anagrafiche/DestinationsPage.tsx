import { useState } from 'react';
import {
  useGetDestinationsQuery,
  useCreateDestinationMutation,
  useUpdateDestinationMutation,
  useDeleteDestinationMutation,
} from '@/store/api/appApi';
import { getMutationErrorMessage } from '@/store/api/rtkQueryHelpers';
import type { DtoDestinationRequest, DtoDestinationResponse, DtoGeocodeResultDTO } from '@/api/data-contracts';
import { usePagination } from '@/hooks/use-pagination';
import { DataTable, type DataTableColumn } from '@/components/shared/DataTable';
import { FormDialog } from '@/components/shared/FormDialog';
import { MapPicker } from '@/components/shared/MapPicker';
import { AddressSearchInput } from '@/components/shared/AddressSearchInput';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import { Button } from '@/components/ui/button';
import { Switch } from '@/components/ui/switch';
import { Badge } from '@/components/ui/badge';
import { TableRow, TableCell } from '@/components/ui/table';
import { toast } from 'sonner';
import { Pencil, Trash2 } from 'lucide-react';

type DestinationForm = Omit<DtoDestinationRequest, 'lat' | 'lng'> & { lat: number | null; lng: number | null };

const PAGE_SIZE = 20;

const emptyForm: DestinationForm = { nome: '', indirizzo: '', citta: '', cap: '', provincia: '', nazione: 'Italia', lat: null, lng: null, vincoli_scarico: '', note: '', active: true };

export default function DestinationsPage() {
  const [search, setSearch] = useState('');
  const [dialogOpen, setDialogOpen] = useState(false);
  const [form, setForm] = useState<DestinationForm>(emptyForm);
  const [editId, setEditId] = useState<string | null>(null);
  const [flySignal, setFlySignal] = useState(0);

  const [page, setPage] = usePagination(search);
  const { data: result, isLoading: loading } = useGetDestinationsQuery({ search, includeInactive: true, page, limit: PAGE_SIZE });
  const data = result?.items ?? [];
  const totalPages = Math.max(1, Math.ceil((result?.total ?? 0) / PAGE_SIZE));
  const [createDestination, { isLoading: creating }] = useCreateDestinationMutation();
  const [updateDestination, { isLoading: updating }] = useUpdateDestinationMutation();
  const [deleteDestination] = useDeleteDestinationMutation();
  const saving = creating || updating;

  const openNew = () => { setForm(emptyForm); setEditId(null); setDialogOpen(true); };
  const openEdit = (item: DtoDestinationResponse) => { setForm({ nome: item.nome || '', indirizzo: item.indirizzo || '', citta: item.citta || '', cap: item.cap || '', provincia: item.provincia || '', nazione: item.nazione || 'Italia', lat: item.lat ?? null, lng: item.lng ?? null, vincoli_scarico: item.vincoli_scarico || '', note: item.note || '', active: item.active }); setEditId(item.id || null); setDialogOpen(true); };

  const handleSave = async () => {
    if (form.lat == null || form.lng == null) { toast.error('Cerca e seleziona un indirizzo per impostare la posizione'); return; }
    try {
      const body: DtoDestinationRequest = { ...form, lat: form.lat, lng: form.lng };
      if (editId) { await updateDestination({ id: editId, body }).unwrap(); toast.success('Destinazione aggiornata'); }
      else { await createDestination(body).unwrap(); toast.success('Destinazione creata'); }
      setDialogOpen(false);
    } catch (e) { toast.error(getMutationErrorMessage(e) || 'Errore'); }
  };

  const handleDelete = async (id: string) => {
    if (!window.confirm('Eliminare questa destinazione?')) return;
    try { await deleteDestination(id).unwrap(); toast.success('Eliminata'); } catch (e) { toast.error(getMutationErrorMessage(e) || 'Errore'); }
  };

  const columns: DataTableColumn[] = [
    { key: 'nome', label: 'Nome' },
    { key: 'citta', label: 'Città' },
    { key: 'nazione', label: 'Nazione' },
    { key: 'vincoli', label: 'Vincoli Scarico' },
    { key: 'stato', label: 'Stato' },
    { key: 'actions', label: '', className: 'w-20' },
  ];

  return (
    <div data-testid="destinations-page">
      <DataTable
        columns={columns} data={data} loading={loading} searchValue={search}
        onSearchChange={setSearch} onAdd={openNew} addLabel="Nuova Destinazione" testId="masterdata-table"
        page={page} totalPages={totalPages} onPageChange={setPage}
        renderRow={(item) => (
          <TableRow key={item.id} className={`hover:bg-muted/60 ${!item.active ? 'opacity-60' : ''}`}>
            <TableCell className="py-2 font-medium">{item.nome}</TableCell>
            <TableCell className="py-2">{item.citta}</TableCell>
            <TableCell className="py-2">{item.nazione}</TableCell>
            <TableCell className="py-2 text-xs">{item.vincoli_scarico}</TableCell>
            <TableCell className="py-2">{item.active ? <Badge className="bg-emerald-100 text-emerald-800 text-[10px]">Attivo</Badge> : <Badge variant="outline" className="text-[10px]">Inattivo</Badge>}</TableCell>
            <TableCell className="py-2">
              <div className="flex gap-1">
                <Button variant="ghost" size="icon" className="h-7 w-7" onClick={() => openEdit(item)}><Pencil className="h-3 w-3" /></Button>
                <Button variant="ghost" size="icon" className="h-7 w-7 text-destructive" onClick={() => item.id && handleDelete(item.id)}><Trash2 className="h-3 w-3" /></Button>
              </div>
            </TableCell>
          </TableRow>
        )}
      />
      <FormDialog open={dialogOpen} onClose={setDialogOpen} title={editId ? 'Modifica Destinazione' : 'Nuova Destinazione'} onSubmit={handleSave} loading={saving}>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
          <div className="md:col-span-2 space-y-1.5"><Label>Nome *</Label><Input value={form.nome} onChange={e => setForm({ ...form, nome: e.target.value })} required /></div>
          <div className="md:col-span-2 space-y-1.5">
            <Label>Indirizzo *</Label>
            <AddressSearchInput
              value={form.indirizzo || ''}
              onChange={(v) => setForm(f => ({ ...f, indirizzo: v }))}
              onSelect={(r: DtoGeocodeResultDTO) => {
                setForm(f => ({ ...f, indirizzo: r.indirizzo || f.indirizzo, citta: r.citta || f.citta, cap: r.cap || f.cap, provincia: r.provincia || f.provincia, nazione: r.nazione || f.nazione, lat: r.lat ?? f.lat, lng: r.lng ?? f.lng }));
                setFlySignal(s => s + 1);
              }}
              required
            />
          </div>
          {/* Sola lettura: valorizzati dalla ricerca su Indirizzo (geocoding ORS), non editabili a mano — coerenza garantita con l'indirizzo scelto. */}
          <div className="space-y-1.5"><Label>Città</Label><Input value={form.citta} readOnly className="bg-muted/50" /></div>
          {/* CAP editabile: il geocoding ORS non sempre restituisce il postalcode (dati OSM incompleti per alcune vie/comuni) — a differenza di città/provincia/nazione va quindi correggibile a mano. */}
          <div className="space-y-1.5"><Label>CAP</Label><Input value={form.cap} onChange={e => setForm({ ...form, cap: e.target.value })} /></div>
          <div className="space-y-1.5"><Label>Provincia</Label><Input value={form.provincia} readOnly className="bg-muted/50" /></div>
          <div className="space-y-1.5"><Label>Nazione</Label><Input value={form.nazione} readOnly className="bg-muted/50" /></div>
          <div className="md:col-span-2 space-y-1.5">
            <Label>Posizione *</Label>
            <MapPicker
              lat={form.lat} lng={form.lng}
              flyToSignal={flySignal}
            />
          </div>
          <div className="md:col-span-2 space-y-1.5"><Label>Vincoli Scarico</Label><Textarea value={form.vincoli_scarico} onChange={e => setForm({ ...form, vincoli_scarico: e.target.value })} rows={2} /></div>
          <div className="md:col-span-2 space-y-1.5"><Label>Note</Label><Textarea value={form.note} onChange={e => setForm({ ...form, note: e.target.value })} rows={2} /></div>
          {editId && (
            <div className="md:col-span-2 flex items-center gap-2">
              <Switch checked={form.active} onCheckedChange={active => setForm({ ...form, active })} id="destination-active" />
              <Label htmlFor="destination-active">Attivo</Label>
            </div>
          )}
        </div>
      </FormDialog>
    </div>
  );
}
