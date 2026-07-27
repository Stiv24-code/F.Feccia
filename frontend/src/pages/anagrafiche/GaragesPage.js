import { useState } from 'react';
import {
  useGetGaragesQuery,
  useCreateGarageMutation,
  useUpdateGarageMutation,
  useDeleteGarageMutation,
} from '@/store/api/appApi';
import { DataTable } from '@/components/shared/DataTable';
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

const emptyForm = { nome: '', indirizzo: '', citta: '', lat: null, lng: null, note: '', active: true };

export default function GaragesPage() {
  const [dialogOpen, setDialogOpen] = useState(false);
  const [form, setForm] = useState(emptyForm);
  const [editId, setEditId] = useState(null);
  const [flySignal, setFlySignal] = useState(0);

  // include_inactive: la pagina anagrafica deve poter riattivare un garage
  // disattivato, quindi mostra anche gli inattivi (a differenza di dove
  // Garage viene usato solo per scegliere un punto di partenza).
  const { data = [], isLoading: loading } = useGetGaragesQuery(true);
  const [createGarage, { isLoading: creating }] = useCreateGarageMutation();
  const [updateGarage, { isLoading: updating }] = useUpdateGarageMutation();
  const [deleteGarage] = useDeleteGarageMutation();
  const saving = creating || updating;

  const openNew = () => { setForm(emptyForm); setEditId(null); setDialogOpen(true); };
  const openEdit = (item) => { setForm({ nome: item.nome, indirizzo: item.indirizzo || '', citta: item.citta || '', lat: item.lat ?? null, lng: item.lng ?? null, note: item.note || '', active: item.active }); setEditId(item.id); setDialogOpen(true); };

  const handleSave = async () => {
    if (form.lat == null || form.lng == null) { toast.error('Seleziona un punto sulla mappa'); return; }
    try {
      if (editId) { await updateGarage({ id: editId, body: form }).unwrap(); toast.success('Garage aggiornato'); } else { await createGarage(form).unwrap(); toast.success('Garage creato'); }
      setDialogOpen(false);
    } catch (e) { toast.error('Errore'); }
  };

  const handleDelete = async (id) => {
    if (!window.confirm('Eliminare?')) return;
    try { await deleteGarage(id).unwrap(); toast.success('Eliminato'); } catch(e) { toast.error('Errore'); }
  };

  const columns = [{ key: 'nome', label: 'Nome' }, { key: 'citta', label: 'Città' }, { key: 'stato', label: 'Stato' }, { key: 'actions', label: '', className: 'w-20' }];

  return (
    <div data-testid="garages-page">
      <DataTable columns={columns} data={data} loading={loading} searchValue="" onSearchChange={() => {}} onAdd={openNew} addLabel="Nuovo Garage" testId="masterdata-table"
        renderRow={(item) => (
          <TableRow key={item.id} className={`hover:bg-muted/60 ${!item.active ? 'opacity-60' : ''}`}>
            <TableCell className="py-2 font-medium">{item.nome}</TableCell>
            <TableCell className="py-2">{item.citta}</TableCell>
            <TableCell className="py-2">{item.active ? <Badge className="bg-emerald-100 text-emerald-800 text-[10px]">Attivo</Badge> : <Badge variant="outline" className="text-[10px]">Inattivo</Badge>}</TableCell>
            <TableCell className="py-2"><div className="flex gap-1"><Button variant="ghost" size="icon" className="h-7 w-7" onClick={() => openEdit(item)}><Pencil className="h-3 w-3" /></Button><Button variant="ghost" size="icon" className="h-7 w-7 text-destructive" onClick={() => handleDelete(item.id)}><Trash2 className="h-3 w-3" /></Button></div></TableCell>
          </TableRow>
        )}
      />
      <FormDialog open={dialogOpen} onClose={setDialogOpen} title={editId ? 'Modifica Garage' : 'Nuovo Garage'} onSubmit={handleSave} loading={saving}>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
          <div className="md:col-span-2 space-y-1.5"><Label>Nome *</Label><Input value={form.nome} onChange={e => setForm({...form, nome: e.target.value})} required /></div>
          <div className="md:col-span-2 space-y-1.5">
            <Label>Indirizzo</Label>
            <AddressSearchInput
              value={form.indirizzo}
              onChange={(v) => setForm(f => ({...f, indirizzo: v}))}
              onSelect={(r) => {
                setForm(f => ({...f, indirizzo: r.indirizzo, citta: r.citta || f.citta, lat: r.lat, lng: r.lng}));
                setFlySignal(s => s + 1);
              }}
            />
          </div>
          <div className="space-y-1.5"><Label>Città</Label><Input value={form.citta} onChange={e => setForm({...form, citta: e.target.value})} /></div>
          <div className="md:col-span-2 space-y-1.5">
            <Label>Posizione *</Label>
            <MapPicker
              lat={form.lat} lng={form.lng}
              onChange={(lat, lng) => setForm({...form, lat, lng})}
              flyToSignal={flySignal}
            />
          </div>
          {editId && (
            <div className="md:col-span-2 flex items-center gap-2">
              <Switch checked={form.active} onCheckedChange={active => setForm({...form, active})} id="garage-active" />
              <Label htmlFor="garage-active">Attivo</Label>
            </div>
          )}
        </div>
      </FormDialog>
    </div>
  );
}
