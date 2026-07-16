import { useState, useEffect, useCallback } from 'react';
import { getDrivers, createDriver, updateDriver, deleteDriver } from '@/lib/api';
import { DataTable } from '@/components/shared/DataTable';
import { FormDialog } from '@/components/shared/FormDialog';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Button } from '@/components/ui/button';
import { TableRow, TableCell } from '@/components/ui/table';
import { toast } from 'sonner';
import { Pencil, Trash2 } from 'lucide-react';

const emptyForm = { nome: '', cognome: '', codice_fiscale: '', patente: '', scadenza_patente: '', telefono: '', email: '', note: '' };

export default function DriversPage() {
  const [data, setData] = useState([]);
  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState('');
  const [dialogOpen, setDialogOpen] = useState(false);
  const [form, setForm] = useState(emptyForm);
  const [editId, setEditId] = useState(null);
  const [saving, setSaving] = useState(false);

  // API imports e state setters sono riferimenti stabili in React
  const fetchData = useCallback(() => { setLoading(true); getDrivers(search).then(r => setData(r.data)).finally(() => setLoading(false)); }, [search]);
  useEffect(() => { fetchData(); }, [fetchData]);

  const openNew = () => { setForm(emptyForm); setEditId(null); setDialogOpen(true); };
  const openEdit = (item) => { setForm({ nome: item.nome, cognome: item.cognome, codice_fiscale: item.codice_fiscale || '', patente: item.patente || '', scadenza_patente: item.scadenza_patente || '', telefono: item.telefono || '', email: item.email || '', note: item.note || '' }); setEditId(item.id); setDialogOpen(true); };

  const handleSave = async () => {
    setSaving(true);
    try {
      if (editId) { await updateDriver(editId, form); toast.success('Autista aggiornato'); } else { await createDriver(form); toast.success('Autista creato'); }
      setDialogOpen(false); fetchData();
    } catch (e) { toast.error('Errore'); } finally { setSaving(false); }
  };

  const handleDelete = async (id) => {
    if (!window.confirm('Eliminare questo autista?')) return;
    try { await deleteDriver(id); toast.success('Eliminato'); fetchData(); } catch(e) { toast.error('Errore'); }
  };

  const columns = [{ key: 'cognome', label: 'Cognome' }, { key: 'nome', label: 'Nome' }, { key: 'patente', label: 'Patente' }, { key: 'telefono', label: 'Telefono' }, { key: 'actions', label: '', className: 'w-20' }];

  return (
    <div data-testid="drivers-page">
      <DataTable columns={columns} data={data} loading={loading} searchValue={search} onSearchChange={setSearch} onAdd={openNew} addLabel="Nuovo Autista" testId="masterdata-table"
        renderRow={(item) => (
          <TableRow key={item.id} className="hover:bg-muted/60">
            <TableCell className="py-2 font-medium">{item.cognome}</TableCell>
            <TableCell className="py-2">{item.nome}</TableCell>
            <TableCell className="py-2">{item.patente}</TableCell>
            <TableCell className="py-2">{item.telefono}</TableCell>
            <TableCell className="py-2"><div className="flex gap-1"><Button variant="ghost" size="icon" className="h-7 w-7" onClick={() => openEdit(item)}><Pencil className="h-3 w-3" /></Button><Button variant="ghost" size="icon" className="h-7 w-7 text-destructive" onClick={() => handleDelete(item.id)}><Trash2 className="h-3 w-3" /></Button></div></TableCell>
          </TableRow>
        )}
      />
      <FormDialog open={dialogOpen} onClose={setDialogOpen} title={editId ? 'Modifica Autista' : 'Nuovo Autista'} onSubmit={handleSave} loading={saving}>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
          <div className="space-y-1.5"><Label>Nome *</Label><Input value={form.nome} onChange={e => setForm({...form, nome: e.target.value})} required /></div>
          <div className="space-y-1.5"><Label>Cognome *</Label><Input value={form.cognome} onChange={e => setForm({...form, cognome: e.target.value})} required /></div>
          <div className="space-y-1.5"><Label>Patente</Label><Input value={form.patente} onChange={e => setForm({...form, patente: e.target.value})} /></div>
          <div className="space-y-1.5"><Label>Scadenza Patente</Label><Input type="date" value={form.scadenza_patente} onChange={e => setForm({...form, scadenza_patente: e.target.value})} /></div>
          <div className="space-y-1.5"><Label>Telefono</Label><Input value={form.telefono} onChange={e => setForm({...form, telefono: e.target.value})} /></div>
          <div className="space-y-1.5"><Label>Email</Label><Input value={form.email} onChange={e => setForm({...form, email: e.target.value})} /></div>
          <div className="space-y-1.5"><Label>Codice Fiscale</Label><Input value={form.codice_fiscale} onChange={e => setForm({...form, codice_fiscale: e.target.value})} /></div>
        </div>
      </FormDialog>
    </div>
  );
}
