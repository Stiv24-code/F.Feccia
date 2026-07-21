import { useState } from 'react';
import {
  useGetProductsQuery,
  useCreateProductMutation,
  useUpdateProductMutation,
  useDeleteProductMutation,
} from '@/store/api/appApi';
import { DataTable } from '@/components/shared/DataTable';
import { FormDialog } from '@/components/shared/FormDialog';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Button } from '@/components/ui/button';
import { TableRow, TableCell } from '@/components/ui/table';
import { toast } from 'sonner';
import { Pencil, Trash2 } from 'lucide-react';

const emptyForm = { codice: '', descrizione: '', unita_misura: 'Kg', note: '' };

export default function ProductsPage() {
  const [search, setSearch] = useState('');
  const [dialogOpen, setDialogOpen] = useState(false);
  const [form, setForm] = useState(emptyForm);
  const [editId, setEditId] = useState(null);

  const { data = [], isLoading: loading } = useGetProductsQuery(search);
  const [createProduct, { isLoading: creating }] = useCreateProductMutation();
  const [updateProduct, { isLoading: updating }] = useUpdateProductMutation();
  const [deleteProduct] = useDeleteProductMutation();
  const saving = creating || updating;

  const openNew = () => { setForm(emptyForm); setEditId(null); setDialogOpen(true); };
  const openEdit = (item) => { setForm({ codice: item.codice, descrizione: item.descrizione, unita_misura: item.unita_misura || 'Kg', note: item.note || '' }); setEditId(item.id); setDialogOpen(true); };

  const handleSave = async () => {
    try {
      if (editId) { await updateProduct({ id: editId, body: form }).unwrap(); toast.success('Prodotto aggiornato'); } else { await createProduct(form).unwrap(); toast.success('Prodotto creato'); }
      setDialogOpen(false);
    } catch (e) { toast.error('Errore'); }
  };

  const handleDelete = async (id) => {
    if (!window.confirm('Eliminare questo prodotto?')) return;
    try { await deleteProduct(id).unwrap(); toast.success('Eliminato'); } catch(e) { toast.error('Errore'); }
  };

  const columns = [{ key: 'codice', label: 'Codice', className: 'font-mono' }, { key: 'descrizione', label: 'Descrizione' }, { key: 'um', label: 'U.M.' }, { key: 'actions', label: '', className: 'w-20' }];

  return (
    <div data-testid="products-page">
      <DataTable columns={columns} data={data} loading={loading} searchValue={search} onSearchChange={setSearch} onAdd={openNew} addLabel="Nuovo Prodotto" testId="masterdata-table"
        renderRow={(item) => (
          <TableRow key={item.id} className="hover:bg-muted/60">
            <TableCell className="py-2 font-mono font-medium">{item.codice}</TableCell>
            <TableCell className="py-2">{item.descrizione}</TableCell>
            <TableCell className="py-2">{item.unita_misura}</TableCell>
            <TableCell className="py-2"><div className="flex gap-1"><Button variant="ghost" size="icon" className="h-7 w-7" onClick={() => openEdit(item)}><Pencil className="h-3 w-3" /></Button><Button variant="ghost" size="icon" className="h-7 w-7 text-destructive" onClick={() => handleDelete(item.id)}><Trash2 className="h-3 w-3" /></Button></div></TableCell>
          </TableRow>
        )}
      />
      <FormDialog open={dialogOpen} onClose={setDialogOpen} title={editId ? 'Modifica Prodotto' : 'Nuovo Prodotto'} onSubmit={handleSave} loading={saving}>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
          <div className="space-y-1.5"><Label>Codice *</Label><Input value={form.codice} onChange={e => setForm({...form, codice: e.target.value})} required /></div>
          <div className="space-y-1.5"><Label>Descrizione *</Label><Input value={form.descrizione} onChange={e => setForm({...form, descrizione: e.target.value})} required /></div>
          <div className="space-y-1.5"><Label>Unità di Misura</Label><Input value={form.unita_misura} onChange={e => setForm({...form, unita_misura: e.target.value})} /></div>
        </div>
      </FormDialog>
    </div>
  );
}
