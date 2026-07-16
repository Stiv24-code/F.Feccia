import { useState, useEffect, useCallback } from 'react';
import { getCarriers, createCarrier, updateCarrier, deleteCarrier } from '@/lib/api';
import { DataTable } from '@/components/shared/DataTable';
import { FormDialog } from '@/components/shared/FormDialog';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Button } from '@/components/ui/button';
import { TableRow, TableCell } from '@/components/ui/table';
import { toast } from 'sonner';
import { Pencil, Trash2 } from 'lucide-react';

const emptyForm = { ragione_sociale: '', partita_iva: '', indirizzo: '', citta: '', telefono: '', email: '', note: '' };

export default function CarriersPage() {
  const [data, setData] = useState([]);
  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState('');
  const [dialogOpen, setDialogOpen] = useState(false);
  const [form, setForm] = useState(emptyForm);
  const [editId, setEditId] = useState(null);
  const [saving, setSaving] = useState(false);

  // API imports e state setters sono riferimenti stabili in React
  const fetchData = useCallback(() => { setLoading(true); getCarriers(search).then(r => setData(r.data)).finally(() => setLoading(false)); }, [search]);
  useEffect(() => { fetchData(); }, [fetchData]);

  const openNew = () => { setForm(emptyForm); setEditId(null); setDialogOpen(true); };
  const openEdit = (item) => { setForm({ ragione_sociale: item.ragione_sociale, partita_iva: item.partita_iva || '', indirizzo: item.indirizzo || '', citta: item.citta || '', telefono: item.telefono || '', email: item.email || '', note: item.note || '' }); setEditId(item.id); setDialogOpen(true); };

  const handleSave = async () => {
    setSaving(true);
    try {
      if (editId) { await updateCarrier(editId, form); toast.success('Vettore aggiornato'); } else { await createCarrier(form); toast.success('Vettore creato'); }
      setDialogOpen(false); fetchData();
    } catch (e) { toast.error('Errore'); } finally { setSaving(false); }
  };

  const handleDelete = async (id) => {
    if (!window.confirm('Eliminare questo vettore?')) return;
    try { await deleteCarrier(id); toast.success('Eliminato'); fetchData(); } catch(e) { toast.error('Errore'); }
  };

  const columns = [{ key: 'ragione_sociale', label: 'Ragione Sociale' }, { key: 'citta', label: 'Città' }, { key: 'telefono', label: 'Telefono' }, { key: 'actions', label: '', className: 'w-20' }];

  return (
    <div data-testid="carriers-page">
      <DataTable columns={columns} data={data} loading={loading} searchValue={search} onSearchChange={setSearch} onAdd={openNew} addLabel="Nuovo Vettore" testId="masterdata-table"
        renderRow={(item) => (
          <TableRow key={item.id} className="hover:bg-muted/60">
            <TableCell className="py-2 font-medium">{item.ragione_sociale}</TableCell>
            <TableCell className="py-2">{item.citta}</TableCell>
            <TableCell className="py-2">{item.telefono}</TableCell>
            <TableCell className="py-2"><div className="flex gap-1"><Button variant="ghost" size="icon" className="h-7 w-7" onClick={() => openEdit(item)}><Pencil className="h-3 w-3" /></Button><Button variant="ghost" size="icon" className="h-7 w-7 text-destructive" onClick={() => handleDelete(item.id)}><Trash2 className="h-3 w-3" /></Button></div></TableCell>
          </TableRow>
        )}
      />
      <FormDialog open={dialogOpen} onClose={setDialogOpen} title={editId ? 'Modifica Vettore' : 'Nuovo Vettore'} onSubmit={handleSave} loading={saving}>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
          <div className="md:col-span-2 space-y-1.5"><Label>Ragione Sociale *</Label><Input value={form.ragione_sociale} onChange={e => setForm({...form, ragione_sociale: e.target.value})} required /></div>
          <div className="space-y-1.5"><Label>P.IVA</Label><Input value={form.partita_iva} onChange={e => setForm({...form, partita_iva: e.target.value})} /></div>
          <div className="space-y-1.5"><Label>Città</Label><Input value={form.citta} onChange={e => setForm({...form, citta: e.target.value})} /></div>
          <div className="space-y-1.5"><Label>Telefono</Label><Input value={form.telefono} onChange={e => setForm({...form, telefono: e.target.value})} /></div>
          <div className="space-y-1.5"><Label>Email</Label><Input value={form.email} onChange={e => setForm({...form, email: e.target.value})} /></div>
        </div>
      </FormDialog>
    </div>
  );
}
