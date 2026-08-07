import { useState } from 'react';
import {
  useGetMotriciQuery, useCreateMotriceMutation, useUpdateMotriceMutation, useDeleteMotriceMutation,
  useGetSemirimorchiQuery, useCreateSemirimorchioMutation, useUpdateSemirimorchioMutation, useDeleteSemirimorchioMutation,
} from '@/store/api/appApi';
import { getMutationErrorMessage } from '@/store/api/rtkQueryHelpers';
import type { DtoMotriceRequest, DtoMotriceResponse, DtoSemirimorchioRequest, DtoSemirimorchioResponse } from '@/api/data-contracts';
import { formatEuro } from '@/lib/format';
import { DataTable, type DataTableColumn } from '@/components/shared/DataTable';
import { FormDialog } from '@/components/shared/FormDialog';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Button } from '@/components/ui/button';
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { TableRow, TableCell } from '@/components/ui/table';
import { toast } from 'sonner';
import { Pencil, Trash2 } from 'lucide-react';

const emptyMotriceForm: DtoMotriceRequest = { targa: '', marca: '', modello: '', anno: 0, portata_kg: 0, note: '' };
const emptySemirimorchioForm: DtoSemirimorchioRequest = { targa: '', tipo: '', scompartature: 1, portata_kg: 0, note: '' };

function MotriciTab() {
  const [search, setSearch] = useState('');
  const [dialogOpen, setDialogOpen] = useState(false);
  const [form, setForm] = useState<DtoMotriceRequest>(emptyMotriceForm);
  const [editId, setEditId] = useState<string | null>(null);

  const { data = [], isLoading: loading } = useGetMotriciQuery(search);
  const [createMotrice, { isLoading: creating }] = useCreateMotriceMutation();
  const [updateMotrice, { isLoading: updating }] = useUpdateMotriceMutation();
  const [deleteMotrice] = useDeleteMotriceMutation();
  const saving = creating || updating;

  const openNew = () => { setForm(emptyMotriceForm); setEditId(null); setDialogOpen(true); };
  const openEdit = (item: DtoMotriceResponse) => { setForm({ targa: item.targa || '', marca: item.marca || '', modello: item.modello || '', anno: item.anno || 0, portata_kg: item.portata_kg || 0, note: item.note || '' }); setEditId(item.id || null); setDialogOpen(true); };

  const handleSave = async () => {
    try {
      if (editId) { await updateMotrice({ id: editId, body: form }).unwrap(); toast.success('Motrice aggiornata'); }
      else { await createMotrice(form).unwrap(); toast.success('Motrice creata'); }
      setDialogOpen(false);
    } catch (e) { toast.error(getMutationErrorMessage(e) || 'Errore'); }
  };

  const handleDelete = async (id: string) => {
    if (!window.confirm('Eliminare questa motrice?')) return;
    try { await deleteMotrice(id).unwrap(); toast.success('Eliminata'); } catch (e) { toast.error(getMutationErrorMessage(e) || 'Errore'); }
  };

  const columns: DataTableColumn[] = [
    { key: 'targa', label: 'Targa' },
    { key: 'marca', label: 'Marca' },
    { key: 'modello', label: 'Modello' },
    { key: 'portata', label: 'Portata (Kg)', className: 'text-right' },
    { key: 'actions', label: '', className: 'w-20' },
  ];

  return (
    <>
      <DataTable columns={columns} data={data} loading={loading} searchValue={search} onSearchChange={setSearch} onAdd={openNew} addLabel="Nuova Motrice" testId="masterdata-table"
        renderRow={(item) => (
          <TableRow key={item.id} className="hover:bg-muted/60">
            <TableCell className="py-2 font-mono font-medium">{item.targa}</TableCell>
            <TableCell className="py-2">{item.marca}</TableCell>
            <TableCell className="py-2">{item.modello}</TableCell>
            <TableCell className="py-2 text-right tabular-nums">{formatEuro(item.portata_kg)}</TableCell>
            <TableCell className="py-2">
              <div className="flex gap-1">
                <Button variant="ghost" size="icon" className="h-7 w-7" onClick={() => openEdit(item)}><Pencil className="h-3 w-3" /></Button>
                <Button variant="ghost" size="icon" className="h-7 w-7 text-destructive" onClick={() => item.id && handleDelete(item.id)}><Trash2 className="h-3 w-3" /></Button>
              </div>
            </TableCell>
          </TableRow>
        )}
      />
      <FormDialog open={dialogOpen} onClose={setDialogOpen} title={editId ? 'Modifica Motrice' : 'Nuova Motrice'} onSubmit={handleSave} loading={saving}>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
          <div className="space-y-1.5"><Label>Targa *</Label><Input value={form.targa} onChange={e => setForm({ ...form, targa: e.target.value })} required /></div>
          <div className="space-y-1.5"><Label>Anno</Label><Input type="number" value={form.anno} onChange={e => setForm({ ...form, anno: Number(e.target.value) })} /></div>
          <div className="space-y-1.5"><Label>Marca</Label><Input value={form.marca} onChange={e => setForm({ ...form, marca: e.target.value })} /></div>
          <div className="space-y-1.5"><Label>Modello</Label><Input value={form.modello} onChange={e => setForm({ ...form, modello: e.target.value })} /></div>
          <div className="space-y-1.5"><Label>Portata (Kg)</Label><Input type="number" value={form.portata_kg} onChange={e => setForm({ ...form, portata_kg: Number(e.target.value) })} /></div>
          <div className="space-y-1.5 md:col-span-2"><Label>Note</Label><Input value={form.note} onChange={e => setForm({ ...form, note: e.target.value })} /></div>
        </div>
      </FormDialog>
    </>
  );
}

function SemirimorchiTab() {
  const [search, setSearch] = useState('');
  const [dialogOpen, setDialogOpen] = useState(false);
  const [form, setForm] = useState<DtoSemirimorchioRequest>(emptySemirimorchioForm);
  const [editId, setEditId] = useState<string | null>(null);

  const { data = [], isLoading: loading } = useGetSemirimorchiQuery(search);
  const [createSemirimorchio, { isLoading: creating }] = useCreateSemirimorchioMutation();
  const [updateSemirimorchio, { isLoading: updating }] = useUpdateSemirimorchioMutation();
  const [deleteSemirimorchio] = useDeleteSemirimorchioMutation();
  const saving = creating || updating;

  const openNew = () => { setForm(emptySemirimorchioForm); setEditId(null); setDialogOpen(true); };
  const openEdit = (item: DtoSemirimorchioResponse) => { setForm({ targa: item.targa || '', tipo: item.tipo || '', scompartature: item.scompartature || 1, portata_kg: item.portata_kg || 0, note: item.note || '' }); setEditId(item.id || null); setDialogOpen(true); };

  const handleSave = async () => {
    try {
      if (editId) { await updateSemirimorchio({ id: editId, body: form }).unwrap(); toast.success('Semirimorchio aggiornato'); }
      else { await createSemirimorchio(form).unwrap(); toast.success('Semirimorchio creato'); }
      setDialogOpen(false);
    } catch (e) { toast.error(getMutationErrorMessage(e) || 'Errore'); }
  };

  const handleDelete = async (id: string) => {
    if (!window.confirm('Eliminare questo semirimorchio?')) return;
    try { await deleteSemirimorchio(id).unwrap(); toast.success('Eliminato'); } catch (e) { toast.error(getMutationErrorMessage(e) || 'Errore'); }
  };

  const columns: DataTableColumn[] = [
    { key: 'targa', label: 'Targa' },
    { key: 'tipo', label: 'Tipo' },
    { key: 'scompartature', label: 'Scompartature', className: 'text-right' },
    { key: 'portata', label: 'Portata (Kg)', className: 'text-right' },
    { key: 'actions', label: '', className: 'w-20' },
  ];

  return (
    <>
      <DataTable columns={columns} data={data} loading={loading} searchValue={search} onSearchChange={setSearch} onAdd={openNew} addLabel="Nuovo Semirimorchio" testId="masterdata-table"
        renderRow={(item) => (
          <TableRow key={item.id} className="hover:bg-muted/60">
            <TableCell className="py-2 font-mono font-medium">{item.targa}</TableCell>
            <TableCell className="py-2">{item.tipo}</TableCell>
            <TableCell className="py-2 text-right tabular-nums">{item.scompartature}</TableCell>
            <TableCell className="py-2 text-right tabular-nums">{formatEuro(item.portata_kg)}</TableCell>
            <TableCell className="py-2">
              <div className="flex gap-1">
                <Button variant="ghost" size="icon" className="h-7 w-7" onClick={() => openEdit(item)}><Pencil className="h-3 w-3" /></Button>
                <Button variant="ghost" size="icon" className="h-7 w-7 text-destructive" onClick={() => item.id && handleDelete(item.id)}><Trash2 className="h-3 w-3" /></Button>
              </div>
            </TableCell>
          </TableRow>
        )}
      />
      <FormDialog open={dialogOpen} onClose={setDialogOpen} title={editId ? 'Modifica Semirimorchio' : 'Nuovo Semirimorchio'} onSubmit={handleSave} loading={saving}>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
          <div className="space-y-1.5"><Label>Targa *</Label><Input value={form.targa} onChange={e => setForm({ ...form, targa: e.target.value })} required /></div>
          <div className="space-y-1.5"><Label>Tipo</Label><Input value={form.tipo} onChange={e => setForm({ ...form, tipo: e.target.value })} placeholder="es. Frigo, Centinato" /></div>
          <div className="space-y-1.5"><Label>Scompartature</Label><Input type="number" value={form.scompartature} onChange={e => setForm({ ...form, scompartature: Number(e.target.value) })} /></div>
          <div className="space-y-1.5"><Label>Portata (Kg)</Label><Input type="number" value={form.portata_kg} onChange={e => setForm({ ...form, portata_kg: Number(e.target.value) })} /></div>
          <div className="space-y-1.5 md:col-span-2"><Label>Note</Label><Input value={form.note} onChange={e => setForm({ ...form, note: e.target.value })} /></div>
        </div>
      </FormDialog>
    </>
  );
}

export default function TransportPage() {
  const [tab, setTab] = useState<'motrici' | 'semirimorchi'>('motrici');

  return (
    <div data-testid="transport-page">
      <div className="mb-3">
        <Tabs value={tab} onValueChange={(v) => setTab(v as 'motrici' | 'semirimorchi')} data-testid="transport-tabs">
          <TabsList>
            <TabsTrigger value="motrici">Motrici</TabsTrigger>
            <TabsTrigger value="semirimorchi">Semirimorchi</TabsTrigger>
          </TabsList>
        </Tabs>
      </div>
      {tab === 'motrici' ? <MotriciTab /> : <SemirimorchiTab />}
    </div>
  );
}
