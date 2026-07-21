import { useState } from 'react';
import {
  useGetDestinationsQuery,
  useCreateDestinationMutation,
  useUpdateDestinationMutation,
  useDeleteDestinationMutation,
} from '@/store/api/appApi';
import { DataTable } from '@/components/shared/DataTable';
import { FormDialog } from '@/components/shared/FormDialog';
import { MapPicker } from '@/components/shared/MapPicker';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import { Button } from '@/components/ui/button';
import { TableRow, TableCell } from '@/components/ui/table';
import { toast } from 'sonner';
import { Pencil, Trash2 } from 'lucide-react';

const emptyForm = { nome: '', indirizzo: '', citta: '', cap: '', provincia: '', nazione: 'Italia', lat: null, lng: null, vincoli_scarico: '', note: '' };

export default function DestinationsPage() {
  const [search, setSearch] = useState('');
  const [dialogOpen, setDialogOpen] = useState(false);
  const [form, setForm] = useState(emptyForm);
  const [editId, setEditId] = useState(null);

  const { data = [], isLoading: loading } = useGetDestinationsQuery(search);
  const [createDestination, { isLoading: creating }] = useCreateDestinationMutation();
  const [updateDestination, { isLoading: updating }] = useUpdateDestinationMutation();
  const [deleteDestination] = useDeleteDestinationMutation();
  const saving = creating || updating;

  const openNew = () => { setForm(emptyForm); setEditId(null); setDialogOpen(true); };
  const openEdit = (item) => { setForm({ nome: item.nome, indirizzo: item.indirizzo || '', citta: item.citta || '', cap: item.cap || '', provincia: item.provincia || '', nazione: item.nazione || 'Italia', lat: item.lat ?? null, lng: item.lng ?? null, vincoli_scarico: item.vincoli_scarico || '', note: item.note || '' }); setEditId(item.id); setDialogOpen(true); };

  const handleSave = async () => {
    if (form.lat == null || form.lng == null) { toast.error('Seleziona un punto sulla mappa'); return; }
    try {
      if (editId) { await updateDestination({ id: editId, body: form }).unwrap(); toast.success('Destinazione aggiornata'); }
      else { await createDestination(form).unwrap(); toast.success('Destinazione creata'); }
      setDialogOpen(false);
    } catch (e) { toast.error('Errore'); }
  };

  const handleDelete = async (id) => {
    if (!window.confirm('Eliminare questa destinazione?')) return;
    try { await deleteDestination(id).unwrap(); toast.success('Eliminata'); } catch(e) { toast.error('Errore'); }
  };

  const columns = [
    { key: 'nome', label: 'Nome' },
    { key: 'citta', label: 'Città' },
    { key: 'nazione', label: 'Nazione' },
    { key: 'vincoli', label: 'Vincoli Scarico' },
    { key: 'actions', label: '', className: 'w-20' },
  ];

  return (
    <div data-testid="destinations-page">
      <DataTable
        columns={columns} data={data} loading={loading} searchValue={search}
        onSearchChange={setSearch} onAdd={openNew} addLabel="Nuova Destinazione" testId="masterdata-table"
        renderRow={(item) => (
          <TableRow key={item.id} className="hover:bg-muted/60">
            <TableCell className="py-2 font-medium">{item.nome}</TableCell>
            <TableCell className="py-2">{item.citta}</TableCell>
            <TableCell className="py-2">{item.nazione}</TableCell>
            <TableCell className="py-2 text-xs">{item.vincoli_scarico}</TableCell>
            <TableCell className="py-2">
              <div className="flex gap-1">
                <Button variant="ghost" size="icon" className="h-7 w-7" onClick={() => openEdit(item)}><Pencil className="h-3 w-3" /></Button>
                <Button variant="ghost" size="icon" className="h-7 w-7 text-destructive" onClick={() => handleDelete(item.id)}><Trash2 className="h-3 w-3" /></Button>
              </div>
            </TableCell>
          </TableRow>
        )}
      />
      <FormDialog open={dialogOpen} onClose={setDialogOpen} title={editId ? 'Modifica Destinazione' : 'Nuova Destinazione'} onSubmit={handleSave} loading={saving}>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
          <div className="md:col-span-2 space-y-1.5"><Label>Nome *</Label><Input value={form.nome} onChange={e => setForm({...form, nome: e.target.value})} required /></div>
          <div className="space-y-1.5"><Label>Indirizzo</Label><Input value={form.indirizzo} onChange={e => setForm({...form, indirizzo: e.target.value})} /></div>
          <div className="space-y-1.5"><Label>Città</Label><Input value={form.citta} onChange={e => setForm({...form, citta: e.target.value})} /></div>
          <div className="space-y-1.5"><Label>CAP</Label><Input value={form.cap} onChange={e => setForm({...form, cap: e.target.value})} /></div>
          <div className="space-y-1.5"><Label>Provincia</Label><Input value={form.provincia} onChange={e => setForm({...form, provincia: e.target.value})} /></div>
          <div className="space-y-1.5"><Label>Nazione</Label><Input value={form.nazione} onChange={e => setForm({...form, nazione: e.target.value})} /></div>
          <div className="md:col-span-2 space-y-1.5">
            <Label>Posizione *</Label>
            <MapPicker lat={form.lat} lng={form.lng} onChange={(lat, lng) => setForm({...form, lat, lng})} />
          </div>
          <div className="md:col-span-2 space-y-1.5"><Label>Vincoli Scarico</Label><Textarea value={form.vincoli_scarico} onChange={e => setForm({...form, vincoli_scarico: e.target.value})} rows={2} /></div>
          <div className="md:col-span-2 space-y-1.5"><Label>Note</Label><Textarea value={form.note} onChange={e => setForm({...form, note: e.target.value})} rows={2} /></div>
        </div>
      </FormDialog>
    </div>
  );
}
