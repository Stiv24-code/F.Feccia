import { useState, useEffect, useCallback } from 'react';
import { getBanks, createBank, updateBank, deleteBank } from '@/lib/api';
import { DataTable } from '@/components/shared/DataTable';
import { FormDialog } from '@/components/shared/FormDialog';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Button } from '@/components/ui/button';
import { Textarea } from '@/components/ui/textarea';
import { TableRow, TableCell } from '@/components/ui/table';
import { toast } from 'sonner';
import { Pencil, Trash2 } from 'lucide-react';

const emptyForm = { nome: '', bic_swift: '', iban_prefix: '', indirizzo: '', citta: '', note: '' };

export default function BanksPage() {
  const [data, setData] = useState([]);
  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState('');
  const [dialogOpen, setDialogOpen] = useState(false);
  const [form, setForm] = useState(emptyForm);
  const [editId, setEditId] = useState(null);
  const [saving, setSaving] = useState(false);

  const fetchData = useCallback(() => {
    setLoading(true);
    getBanks(search).then(r => setData(r.data)).finally(() => setLoading(false));
  }, [search]);
  useEffect(() => { fetchData(); }, [fetchData]);

  const openNew = () => { setForm(emptyForm); setEditId(null); setDialogOpen(true); };
  const openEdit = (item) => {
    setForm({
      nome: item.nome || '', bic_swift: item.bic_swift || '',
      iban_prefix: item.iban_prefix || '', indirizzo: item.indirizzo || '',
      citta: item.citta || '', note: item.note || '',
    });
    setEditId(item.id); setDialogOpen(true);
  };

  const handleSave = async () => {
    setSaving(true);
    try {
      if (editId) { await updateBank(editId, form); toast.success('Banca aggiornata'); }
      else { await createBank(form); toast.success('Banca creata'); }
      setDialogOpen(false); fetchData();
    } catch (e) { toast.error(e.response?.data?.detail || 'Errore'); } finally { setSaving(false); }
  };

  const handleDelete = async (id) => {
    if (!window.confirm('Eliminare questa banca?')) return;
    try { await deleteBank(id); toast.success('Eliminata'); fetchData(); }
    catch (e) { toast.error('Errore'); }
  };

  const columns = [
    { key: 'nome', label: 'Nome' },
    { key: 'bic', label: 'BIC/SWIFT', className: 'font-mono' },
    { key: 'citta', label: 'Città' },
    { key: 'actions', label: '', className: 'w-20' },
  ];

  return (
    <div data-testid="banks-page">
      <DataTable
        columns={columns} data={data} loading={loading}
        searchValue={search} onSearchChange={setSearch}
        onAdd={openNew} addLabel="Nuova Banca" testId="masterdata-table"
        renderRow={(item) => (
          <TableRow key={item.id} className="hover:bg-muted/60">
            <TableCell className="py-2 font-medium">{item.nome}</TableCell>
            <TableCell className="py-2 font-mono text-xs">{item.bic_swift}</TableCell>
            <TableCell className="py-2">{item.citta}</TableCell>
            <TableCell className="py-2">
              <div className="flex gap-1">
                <Button variant="ghost" size="icon" className="h-7 w-7" aria-label="Modifica banca" onClick={() => openEdit(item)}><Pencil className="h-3 w-3" /></Button>
                <Button variant="ghost" size="icon" className="h-7 w-7 text-destructive" aria-label="Elimina banca" onClick={() => handleDelete(item.id)}><Trash2 className="h-3 w-3" /></Button>
              </div>
            </TableCell>
          </TableRow>
        )}
      />
      <FormDialog open={dialogOpen} onClose={setDialogOpen} title={editId ? 'Modifica Banca' : 'Nuova Banca'} onSubmit={handleSave} loading={saving}>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
          <div className="space-y-1.5 md:col-span-2"><Label>Nome *</Label><Input value={form.nome} onChange={e => setForm({...form, nome: e.target.value})} required /></div>
          <div className="space-y-1.5"><Label>BIC / SWIFT</Label><Input value={form.bic_swift} onChange={e => setForm({...form, bic_swift: e.target.value.toUpperCase()})} maxLength={11} /></div>
          <div className="space-y-1.5"><Label>Prefisso IBAN</Label><Input value={form.iban_prefix} onChange={e => setForm({...form, iban_prefix: e.target.value.toUpperCase()})} maxLength={6} /></div>
          <div className="space-y-1.5"><Label>Indirizzo</Label><Input value={form.indirizzo} onChange={e => setForm({...form, indirizzo: e.target.value})} /></div>
          <div className="space-y-1.5"><Label>Città</Label><Input value={form.citta} onChange={e => setForm({...form, citta: e.target.value})} /></div>
          <div className="space-y-1.5 md:col-span-2"><Label>Note</Label><Textarea value={form.note} onChange={e => setForm({...form, note: e.target.value})} rows={2} /></div>
        </div>
      </FormDialog>
    </div>
  );
}
