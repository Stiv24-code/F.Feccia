import { useState, useEffect, useCallback } from 'react';
import { getAccountingEntries, createAccountingEntry, updateAccountingEntry, deleteAccountingEntry } from '@/lib/api';
import { DataTable } from '@/components/shared/DataTable';
import { FormDialog } from '@/components/shared/FormDialog';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Button } from '@/components/ui/button';
import { Select, SelectTrigger, SelectValue, SelectContent, SelectItem } from '@/components/ui/select';
import { TableRow, TableCell } from '@/components/ui/table';
import { Badge } from '@/components/ui/badge';
import { toast } from 'sonner';
import { Pencil, Trash2 } from 'lucide-react';

const emptyForm = { codice: '', descrizione: '', tipo: 'ricavo', conto_contabile: '', iva_codice: 'N8' };

export default function AccountingEntriesPage() {
  const [data, setData] = useState([]);
  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState('');
  const [dialogOpen, setDialogOpen] = useState(false);
  const [form, setForm] = useState(emptyForm);
  const [editId, setEditId] = useState(null);
  const [saving, setSaving] = useState(false);

  const fetchData = useCallback(() => {
    setLoading(true);
    getAccountingEntries({ search }).then(r => setData(r.data)).finally(() => setLoading(false));
  }, [search]);
  useEffect(() => { fetchData(); }, [fetchData]);

  const openNew = () => { setForm(emptyForm); setEditId(null); setDialogOpen(true); };
  const openEdit = (item) => {
    setForm({
      codice: item.codice || '', descrizione: item.descrizione || '',
      tipo: item.tipo || 'ricavo', conto_contabile: item.conto_contabile || '',
      iva_codice: item.iva_codice || 'N8',
    });
    setEditId(item.id); setDialogOpen(true);
  };

  const handleSave = async () => {
    setSaving(true);
    try {
      if (editId) { await updateAccountingEntry(editId, form); toast.success('Voce aggiornata'); }
      else { await createAccountingEntry(form); toast.success('Voce creata'); }
      setDialogOpen(false); fetchData();
    } catch (e) { toast.error(e.response?.data?.detail || 'Errore'); } finally { setSaving(false); }
  };

  const handleDelete = async (id) => {
    if (!window.confirm('Eliminare questa voce contabile?')) return;
    try { await deleteAccountingEntry(id); toast.success('Eliminata'); fetchData(); }
    catch (e) { toast.error('Errore'); }
  };

  const columns = [
    { key: 'codice', label: 'Codice', className: 'font-mono w-24' },
    { key: 'descrizione', label: 'Descrizione' },
    { key: 'tipo', label: 'Tipo', className: 'w-24' },
    { key: 'conto', label: 'Conto', className: 'font-mono w-32' },
    { key: 'iva', label: 'IVA', className: 'font-mono w-16' },
    { key: 'actions', label: '', className: 'w-20' },
  ];

  return (
    <div data-testid="accounting-entries-page">
      <DataTable
        columns={columns} data={data} loading={loading}
        searchValue={search} onSearchChange={setSearch}
        onAdd={openNew} addLabel="Nuova Voce" testId="masterdata-table"
        renderRow={(item) => (
          <TableRow key={item.id} className="hover:bg-muted/60">
            <TableCell className="py-2 font-mono font-medium">{item.codice}</TableCell>
            <TableCell className="py-2">{item.descrizione}</TableCell>
            <TableCell className="py-2">
              <Badge variant={item.tipo === 'ricavo' ? 'default' : 'secondary'} className="text-[10px] capitalize">{item.tipo}</Badge>
            </TableCell>
            <TableCell className="py-2 font-mono text-xs">{item.conto_contabile}</TableCell>
            <TableCell className="py-2 font-mono text-xs">{item.iva_codice}</TableCell>
            <TableCell className="py-2">
              <div className="flex gap-1">
                <Button variant="ghost" size="icon" className="h-7 w-7" aria-label="Modifica voce contabile" onClick={() => openEdit(item)}><Pencil className="h-3 w-3" /></Button>
                <Button variant="ghost" size="icon" className="h-7 w-7 text-destructive" aria-label="Elimina voce contabile" onClick={() => handleDelete(item.id)}><Trash2 className="h-3 w-3" /></Button>
              </div>
            </TableCell>
          </TableRow>
        )}
      />
      <FormDialog open={dialogOpen} onClose={setDialogOpen} title={editId ? 'Modifica Voce Contabile' : 'Nuova Voce Contabile'} onSubmit={handleSave} loading={saving}>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
          <div className="space-y-1.5"><Label>Codice *</Label><Input value={form.codice} onChange={e => setForm({...form, codice: e.target.value.toUpperCase()})} required /></div>
          <div className="space-y-1.5">
            <Label>Tipo *</Label>
            <Select value={form.tipo} onValueChange={(v) => setForm({...form, tipo: v})}>
              <SelectTrigger><SelectValue /></SelectTrigger>
              <SelectContent>
                <SelectItem value="ricavo">Ricavo</SelectItem>
                <SelectItem value="costo">Costo</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <div className="space-y-1.5 md:col-span-2"><Label>Descrizione *</Label><Input value={form.descrizione} onChange={e => setForm({...form, descrizione: e.target.value})} required /></div>
          <div className="space-y-1.5"><Label>Conto Contabile</Label><Input value={form.conto_contabile} onChange={e => setForm({...form, conto_contabile: e.target.value})} placeholder="es. 70.10.0010" /></div>
          <div className="space-y-1.5"><Label>Codice IVA</Label><Input value={form.iva_codice} onChange={e => setForm({...form, iva_codice: e.target.value.toUpperCase()})} placeholder="es. 22, N8, N3.4" /></div>
        </div>
      </FormDialog>
    </div>
  );
}
