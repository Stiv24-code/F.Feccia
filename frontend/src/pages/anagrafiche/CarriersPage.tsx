import { useState } from 'react';
import {
  useGetCarriersQuery,
  useCreateCarrierMutation,
  useUpdateCarrierMutation,
  useDeleteCarrierMutation,
} from '@/store/api/appApi';
import { getMutationErrorMessage } from '@/store/api/rtkQueryHelpers';
import type { DtoCarrierRequest, DtoCarrierResponse } from '@/api/data-contracts';
import { usePagination } from '@/hooks/use-pagination';
import { DataTable, type DataTableColumn } from '@/components/shared/DataTable';
import { FormDialog } from '@/components/shared/FormDialog';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Button } from '@/components/ui/button';
import { TableRow, TableCell } from '@/components/ui/table';
import { toast } from 'sonner';
import { Pencil, Trash2 } from 'lucide-react';

const PAGE_SIZE = 20;

const emptyForm: DtoCarrierRequest = { ragione_sociale: '', partita_iva: '', indirizzo: '', citta: '', telefono: '', email: '', note: '' };

export default function CarriersPage() {
  const [search, setSearch] = useState('');
  const [dialogOpen, setDialogOpen] = useState(false);
  const [form, setForm] = useState<DtoCarrierRequest>(emptyForm);
  const [editId, setEditId] = useState<string | null>(null);

  const [page, setPage] = usePagination(search);
  const { data: result, isLoading: loading } = useGetCarriersQuery({ search, page, limit: PAGE_SIZE });
  const data = result?.items ?? [];
  const totalPages = Math.max(1, Math.ceil((result?.total ?? 0) / PAGE_SIZE));
  const [createCarrier, { isLoading: creating }] = useCreateCarrierMutation();
  const [updateCarrier, { isLoading: updating }] = useUpdateCarrierMutation();
  const [deleteCarrier] = useDeleteCarrierMutation();
  const saving = creating || updating;

  const openNew = () => { setForm(emptyForm); setEditId(null); setDialogOpen(true); };
  const openEdit = (item: DtoCarrierResponse) => { setForm({ ragione_sociale: item.ragione_sociale || '', partita_iva: item.partita_iva || '', indirizzo: item.indirizzo || '', citta: item.citta || '', telefono: item.telefono || '', email: item.email || '', note: item.note || '' }); setEditId(item.id || null); setDialogOpen(true); };

  const handleSave = async () => {
    try {
      if (editId) { await updateCarrier({ id: editId, body: form }).unwrap(); toast.success('Vettore aggiornato'); } else { await createCarrier(form).unwrap(); toast.success('Vettore creato'); }
      setDialogOpen(false);
    } catch (e) { toast.error(getMutationErrorMessage(e) || 'Errore'); }
  };

  const handleDelete = async (id: string) => {
    if (!window.confirm('Eliminare questo vettore?')) return;
    try { await deleteCarrier(id).unwrap(); toast.success('Eliminato'); } catch (e) { toast.error(getMutationErrorMessage(e) || 'Errore'); }
  };

  const columns: DataTableColumn[] = [{ key: 'ragione_sociale', label: 'Ragione Sociale' }, { key: 'citta', label: 'Città' }, { key: 'telefono', label: 'Telefono' }, { key: 'actions', label: '', className: 'w-20' }];

  return (
    <div data-testid="carriers-page">
      <DataTable columns={columns} data={data} loading={loading} searchValue={search} onSearchChange={setSearch} onAdd={openNew} addLabel="Nuovo Vettore" testId="masterdata-table"
        page={page} totalPages={totalPages} onPageChange={setPage}
        renderRow={(item) => (
          <TableRow key={item.id} className="hover:bg-muted/60">
            <TableCell className="py-2 font-medium">{item.ragione_sociale}</TableCell>
            <TableCell className="py-2">{item.citta}</TableCell>
            <TableCell className="py-2">{item.telefono}</TableCell>
            <TableCell className="py-2"><div className="flex gap-1"><Button variant="ghost" size="icon" className="h-7 w-7" onClick={() => openEdit(item)}><Pencil className="h-3 w-3" /></Button><Button variant="ghost" size="icon" className="h-7 w-7 text-destructive" onClick={() => item.id && handleDelete(item.id)}><Trash2 className="h-3 w-3" /></Button></div></TableCell>
          </TableRow>
        )}
      />
      <FormDialog open={dialogOpen} onClose={setDialogOpen} title={editId ? 'Modifica Vettore' : 'Nuovo Vettore'} onSubmit={handleSave} loading={saving}>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
          <div className="md:col-span-2 space-y-1.5"><Label>Ragione Sociale *</Label><Input value={form.ragione_sociale} onChange={e => setForm({ ...form, ragione_sociale: e.target.value })} required /></div>
          <div className="space-y-1.5"><Label>P.IVA</Label><Input value={form.partita_iva} onChange={e => setForm({ ...form, partita_iva: e.target.value })} /></div>
          <div className="space-y-1.5"><Label>Città</Label><Input value={form.citta} onChange={e => setForm({ ...form, citta: e.target.value })} /></div>
          <div className="space-y-1.5"><Label>Telefono</Label><Input value={form.telefono} onChange={e => setForm({ ...form, telefono: e.target.value })} /></div>
          <div className="space-y-1.5"><Label>Email</Label><Input value={form.email} onChange={e => setForm({ ...form, email: e.target.value })} /></div>
        </div>
      </FormDialog>
    </div>
  );
}
