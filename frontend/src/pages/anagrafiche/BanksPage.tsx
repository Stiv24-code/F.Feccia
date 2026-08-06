import { useState } from 'react';
import {
  useGetBanksQuery,
  useCreateBankMutation,
  useUpdateBankMutation,
  useDeleteBankMutation,
} from '@/store/api/appApi';
import { getMutationErrorMessage } from '@/store/api/rtkQueryHelpers';
import type { DtoBankRequest, DtoBankResponse } from '@/api/data-contracts';
import { DataTable, type DataTableColumn } from '@/components/shared/DataTable';
import { FormDialog } from '@/components/shared/FormDialog';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Button } from '@/components/ui/button';
import { Textarea } from '@/components/ui/textarea';
import { TableRow, TableCell } from '@/components/ui/table';
import { toast } from 'sonner';
import { Pencil, Trash2 } from 'lucide-react';

const emptyForm: DtoBankRequest = { nome: '', bic_swift: '', iban_prefix: '', indirizzo: '', citta: '', note: '' };

export default function BanksPage() {
  const [search, setSearch] = useState('');
  const [dialogOpen, setDialogOpen] = useState(false);
  const [form, setForm] = useState<DtoBankRequest>(emptyForm);
  const [editId, setEditId] = useState<string | null>(null);

  const { data = [], isLoading: loading } = useGetBanksQuery(search);
  const [createBank, { isLoading: creating }] = useCreateBankMutation();
  const [updateBank, { isLoading: updating }] = useUpdateBankMutation();
  const [deleteBank] = useDeleteBankMutation();
  const saving = creating || updating;

  const openNew = () => { setForm(emptyForm); setEditId(null); setDialogOpen(true); };
  const openEdit = (item: DtoBankResponse) => {
    setForm({
      nome: item.nome || '', bic_swift: item.bic_swift || '',
      iban_prefix: item.iban_prefix || '', indirizzo: item.indirizzo || '',
      citta: item.citta || '', note: item.note || '',
    });
    setEditId(item.id || null); setDialogOpen(true);
  };

  const handleSave = async () => {
    try {
      if (editId) { await updateBank({ id: editId, body: form }).unwrap(); toast.success('Banca aggiornata'); }
      else { await createBank(form).unwrap(); toast.success('Banca creata'); }
      setDialogOpen(false);
    } catch (e) { toast.error(getMutationErrorMessage(e) || 'Errore'); }
  };

  const handleDelete = async (id: string) => {
    if (!window.confirm('Eliminare questa banca?')) return;
    try { await deleteBank(id).unwrap(); toast.success('Eliminata'); }
    catch (e) { toast.error(getMutationErrorMessage(e) || 'Errore'); }
  };

  const columns: DataTableColumn[] = [
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
                <Button variant="ghost" size="icon" className="h-7 w-7 text-destructive" aria-label="Elimina banca" onClick={() => item.id && handleDelete(item.id)}><Trash2 className="h-3 w-3" /></Button>
              </div>
            </TableCell>
          </TableRow>
        )}
      />
      <FormDialog open={dialogOpen} onClose={setDialogOpen} title={editId ? 'Modifica Banca' : 'Nuova Banca'} onSubmit={handleSave} loading={saving}>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
          <div className="space-y-1.5 md:col-span-2"><Label>Nome *</Label><Input value={form.nome} onChange={e => setForm({ ...form, nome: e.target.value })} required /></div>
          <div className="space-y-1.5"><Label>BIC / SWIFT</Label><Input value={form.bic_swift} onChange={e => setForm({ ...form, bic_swift: e.target.value.toUpperCase() })} maxLength={11} /></div>
          <div className="space-y-1.5"><Label>Prefisso IBAN</Label><Input value={form.iban_prefix} onChange={e => setForm({ ...form, iban_prefix: e.target.value.toUpperCase() })} maxLength={6} /></div>
          <div className="space-y-1.5"><Label>Indirizzo</Label><Input value={form.indirizzo} onChange={e => setForm({ ...form, indirizzo: e.target.value })} /></div>
          <div className="space-y-1.5"><Label>Città</Label><Input value={form.citta} onChange={e => setForm({ ...form, citta: e.target.value })} /></div>
          <div className="space-y-1.5 md:col-span-2"><Label>Note</Label><Textarea value={form.note} onChange={e => setForm({ ...form, note: e.target.value })} rows={2} /></div>
        </div>
      </FormDialog>
    </div>
  );
}
