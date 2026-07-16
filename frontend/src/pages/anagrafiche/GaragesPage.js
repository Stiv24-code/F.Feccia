import { useState, useEffect, useCallback } from 'react';
import { getGarages, createGarage, updateGarage, deleteGarage } from '@/lib/api';
import { DataTable } from '@/components/shared/DataTable';
import { FormDialog } from '@/components/shared/FormDialog';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Button } from '@/components/ui/button';
import { TableRow, TableCell } from '@/components/ui/table';
import { toast } from 'sonner';
import { Pencil, Trash2 } from 'lucide-react';

const emptyForm = { nome: '', indirizzo: '', citta: '', note: '' };

export default function GaragesPage() {
  const [data, setData] = useState([]);
  const [loading, setLoading] = useState(true);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [form, setForm] = useState(emptyForm);
  const [editId, setEditId] = useState(null);
  const [saving, setSaving] = useState(false);

  // API imports e state setters sono riferimenti stabili in React
  const fetchData = useCallback(() => { setLoading(true); getGarages().then(r => setData(r.data)).finally(() => setLoading(false)); }, []);
  useEffect(() => { fetchData(); }, [fetchData]);

  const openNew = () => { setForm(emptyForm); setEditId(null); setDialogOpen(true); };
  const openEdit = (item) => { setForm({ nome: item.nome, indirizzo: item.indirizzo || '', citta: item.citta || '', note: item.note || '' }); setEditId(item.id); setDialogOpen(true); };

  const handleSave = async () => {
    setSaving(true);
    try {
      if (editId) { await updateGarage(editId, form); toast.success('Garage aggiornato'); } else { await createGarage(form); toast.success('Garage creato'); }
      setDialogOpen(false); fetchData();
    } catch (e) { toast.error('Errore'); } finally { setSaving(false); }
  };

  const handleDelete = async (id) => {
    if (!window.confirm('Eliminare?')) return;
    try { await deleteGarage(id); toast.success('Eliminato'); fetchData(); } catch(e) { toast.error('Errore'); }
  };

  const columns = [{ key: 'nome', label: 'Nome' }, { key: 'citta', label: 'Città' }, { key: 'actions', label: '', className: 'w-20' }];

  return (
    <div data-testid="garages-page">
      <DataTable columns={columns} data={data} loading={loading} searchValue="" onSearchChange={() => {}} onAdd={openNew} addLabel="Nuovo Garage" testId="masterdata-table"
        renderRow={(item) => (
          <TableRow key={item.id} className="hover:bg-muted/60">
            <TableCell className="py-2 font-medium">{item.nome}</TableCell>
            <TableCell className="py-2">{item.citta}</TableCell>
            <TableCell className="py-2"><div className="flex gap-1"><Button variant="ghost" size="icon" className="h-7 w-7" onClick={() => openEdit(item)}><Pencil className="h-3 w-3" /></Button><Button variant="ghost" size="icon" className="h-7 w-7 text-destructive" onClick={() => handleDelete(item.id)}><Trash2 className="h-3 w-3" /></Button></div></TableCell>
          </TableRow>
        )}
      />
      <FormDialog open={dialogOpen} onClose={setDialogOpen} title={editId ? 'Modifica Garage' : 'Nuovo Garage'} onSubmit={handleSave} loading={saving}>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
          <div className="md:col-span-2 space-y-1.5"><Label>Nome *</Label><Input value={form.nome} onChange={e => setForm({...form, nome: e.target.value})} required /></div>
          <div className="space-y-1.5"><Label>Indirizzo</Label><Input value={form.indirizzo} onChange={e => setForm({...form, indirizzo: e.target.value})} /></div>
          <div className="space-y-1.5"><Label>Città</Label><Input value={form.citta} onChange={e => setForm({...form, citta: e.target.value})} /></div>
        </div>
      </FormDialog>
    </div>
  );
}
