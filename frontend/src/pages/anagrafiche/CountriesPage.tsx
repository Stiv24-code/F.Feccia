import { useState } from 'react';
import {
  useGetCountriesQuery,
  useCreateCountryMutation,
  useUpdateCountryMutation,
  useDeleteCountryMutation,
} from '@/store/api/appApi';
import { getMutationErrorMessage } from '@/store/api/rtkQueryHelpers';
import type { DtoCountryRequest, DtoCountryResponse } from '@/api/data-contracts';
import { DataTable, type DataTableColumn } from '@/components/shared/DataTable';
import { FormDialog } from '@/components/shared/FormDialog';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Button } from '@/components/ui/button';
import { Checkbox } from '@/components/ui/checkbox';
import { TableRow, TableCell } from '@/components/ui/table';
import { Badge } from '@/components/ui/badge';
import { toast } from 'sonner';
import { Pencil, Trash2 } from 'lucide-react';

const emptyForm: DtoCountryRequest = { iso2: '', iso3: '', nome: '', eu: false, valuta: 'EUR' };

export default function CountriesPage() {
  const [search, setSearch] = useState('');
  const [dialogOpen, setDialogOpen] = useState(false);
  const [form, setForm] = useState<DtoCountryRequest>(emptyForm);
  const [editId, setEditId] = useState<string | null>(null);

  const { data = [], isLoading: loading } = useGetCountriesQuery(search);
  const [createCountry, { isLoading: creating }] = useCreateCountryMutation();
  const [updateCountry, { isLoading: updating }] = useUpdateCountryMutation();
  const [deleteCountry] = useDeleteCountryMutation();
  const saving = creating || updating;

  const openNew = () => { setForm(emptyForm); setEditId(null); setDialogOpen(true); };
  const openEdit = (item: DtoCountryResponse) => {
    setForm({
      iso2: item.iso2 || '', iso3: item.iso3 || '', nome: item.nome || '',
      eu: !!item.eu, valuta: item.valuta || 'EUR',
    });
    setEditId(item.id || null); setDialogOpen(true);
  };

  const handleSave = async () => {
    try {
      if (editId) { await updateCountry({ id: editId, body: form }).unwrap(); toast.success('Nazione aggiornata'); }
      else { await createCountry(form).unwrap(); toast.success('Nazione creata'); }
      setDialogOpen(false);
    } catch (e) { toast.error(getMutationErrorMessage(e) || 'Errore'); }
  };

  const handleDelete = async (id: string) => {
    if (!window.confirm('Eliminare questa nazione?')) return;
    try { await deleteCountry(id).unwrap(); toast.success('Eliminata'); }
    catch (e) { toast.error(getMutationErrorMessage(e) || 'Errore'); }
  };

  const columns: DataTableColumn[] = [
    { key: 'iso2', label: 'ISO2', className: 'font-mono w-16' },
    { key: 'iso3', label: 'ISO3', className: 'font-mono w-16' },
    { key: 'nome', label: 'Nome' },
    { key: 'eu', label: 'UE', className: 'w-16' },
    { key: 'valuta', label: 'Valuta', className: 'font-mono w-20' },
    { key: 'actions', label: '', className: 'w-20' },
  ];

  return (
    <div data-testid="countries-page">
      <DataTable
        columns={columns} data={data} loading={loading}
        searchValue={search} onSearchChange={setSearch}
        onAdd={openNew} addLabel="Nuova Nazione" testId="masterdata-table"
        renderRow={(item) => (
          <TableRow key={item.id} className="hover:bg-muted/60">
            <TableCell className="py-2 font-mono font-medium">{item.iso2}</TableCell>
            <TableCell className="py-2 font-mono text-muted-foreground">{item.iso3}</TableCell>
            <TableCell className="py-2">{item.nome}</TableCell>
            <TableCell className="py-2">{item.eu ? <Badge variant="outline" className="text-[10px]">UE</Badge> : null}</TableCell>
            <TableCell className="py-2 font-mono text-xs">{item.valuta}</TableCell>
            <TableCell className="py-2">
              <div className="flex gap-1">
                <Button variant="ghost" size="icon" className="h-7 w-7" aria-label="Modifica nazione" onClick={() => openEdit(item)}><Pencil className="h-3 w-3" /></Button>
                <Button variant="ghost" size="icon" className="h-7 w-7 text-destructive" aria-label="Elimina nazione" onClick={() => item.id && handleDelete(item.id)}><Trash2 className="h-3 w-3" /></Button>
              </div>
            </TableCell>
          </TableRow>
        )}
      />
      <FormDialog open={dialogOpen} onClose={setDialogOpen} title={editId ? 'Modifica Nazione' : 'Nuova Nazione'} onSubmit={handleSave} loading={saving}>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
          <div className="space-y-1.5"><Label>ISO 3166-1 alpha-2 *</Label><Input value={form.iso2} onChange={e => setForm({ ...form, iso2: e.target.value.toUpperCase() })} maxLength={2} required /></div>
          <div className="space-y-1.5"><Label>ISO 3166-1 alpha-3</Label><Input value={form.iso3} onChange={e => setForm({ ...form, iso3: e.target.value.toUpperCase() })} maxLength={3} /></div>
          <div className="space-y-1.5 md:col-span-2"><Label>Nome *</Label><Input value={form.nome} onChange={e => setForm({ ...form, nome: e.target.value })} required /></div>
          <div className="space-y-1.5"><Label>Valuta (ISO 4217)</Label><Input value={form.valuta} onChange={e => setForm({ ...form, valuta: e.target.value.toUpperCase() })} maxLength={3} /></div>
          <div className="flex items-center gap-2 pt-6">
            <Checkbox id="country-eu" checked={form.eu} onCheckedChange={(v) => setForm({ ...form, eu: !!v })} />
            <Label htmlFor="country-eu" className="cursor-pointer">Stato membro UE</Label>
          </div>
        </div>
      </FormDialog>
    </div>
  );
}
