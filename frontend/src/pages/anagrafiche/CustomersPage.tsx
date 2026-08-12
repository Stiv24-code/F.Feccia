import { useState } from 'react';
import { Link } from 'react-router-dom';
import {
  useGetCustomersQuery,
  useCreateCustomerMutation,
  useUpdateCustomerMutation,
  useDeleteCustomerMutation,
} from '@/store/api/appApi';
import { getMutationErrorMessage } from '@/store/api/rtkQueryHelpers';
import type { DtoCustomerRequest, DtoCustomerResponse, DtoGeocodeResultDTO } from '@/api/data-contracts';
import { DataTable, type DataTableColumn } from '@/components/shared/DataTable';
import { FormDialog } from '@/components/shared/FormDialog';
import { MapPicker } from '@/components/shared/MapPicker';
import { AddressSearchInput } from '@/components/shared/AddressSearchInput';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Switch } from '@/components/ui/switch';
import { Textarea } from '@/components/ui/textarea';
import { Button } from '@/components/ui/button';
import { TableRow, TableCell } from '@/components/ui/table';
import { toast } from 'sonner';
import { Pencil, Trash2, BarChart3 } from 'lucide-react';

type CustomerForm = Omit<DtoCustomerRequest, 'lat' | 'lng'> & { lat: number | null; lng: number | null };

const emptyForm: CustomerForm = { ragione_sociale: '', indirizzo: '', citta: '', cap: '', provincia: '', nazione: 'Italia', lat: null, lng: null, partita_iva: '', codice_fiscale: '', telefono: '', email: '', pec: '', condizioni_pagamento: '', note: '', richiede_rif_ordine: false };

export default function CustomersPage() {
  const [search, setSearch] = useState('');
  const [dialogOpen, setDialogOpen] = useState(false);
  const [form, setForm] = useState<CustomerForm>(emptyForm);
  const [editId, setEditId] = useState<string | null>(null);
  const [flySignal, setFlySignal] = useState(0);

  const { data = [], isLoading: loading } = useGetCustomersQuery(search);
  const [createCustomer, { isLoading: creating }] = useCreateCustomerMutation();
  const [updateCustomer, { isLoading: updating }] = useUpdateCustomerMutation();
  const [deleteCustomer] = useDeleteCustomerMutation();
  const saving = creating || updating;

  const openNew = () => { setForm(emptyForm); setEditId(null); setDialogOpen(true); };
  const openEdit = (item: DtoCustomerResponse) => { setForm({ ragione_sociale: item.ragione_sociale || '', indirizzo: item.indirizzo || '', citta: item.citta || '', cap: item.cap || '', provincia: item.provincia || '', nazione: item.nazione || 'Italia', lat: item.lat ?? null, lng: item.lng ?? null, partita_iva: item.partita_iva || '', codice_fiscale: item.codice_fiscale || '', telefono: item.telefono || '', email: item.email || '', pec: item.pec || '', condizioni_pagamento: item.condizioni_pagamento || '', note: item.note || '', richiede_rif_ordine: item.richiede_rif_ordine || false }); setEditId(item.id || null); setDialogOpen(true); };

  const handleSave = async () => {
    try {
      const body: DtoCustomerRequest = { ...form, lat: form.lat ?? undefined, lng: form.lng ?? undefined };
      if (editId) { await updateCustomer({ id: editId, body }).unwrap(); toast.success('Cliente aggiornato'); }
      else { await createCustomer(body).unwrap(); toast.success('Cliente creato'); }
      setDialogOpen(false);
    } catch (e) { toast.error(getMutationErrorMessage(e) || 'Errore'); }
  };

  const handleDelete = async (id: string) => {
    if (!window.confirm('Eliminare questo cliente?')) return;
    try { await deleteCustomer(id).unwrap(); toast.success('Cliente eliminato'); } catch (e) { toast.error(getMutationErrorMessage(e) || 'Errore'); }
  };

  const columns: DataTableColumn[] = [
    { key: 'ragione_sociale', label: 'Ragione Sociale' },
    { key: 'citta', label: 'Città' },
    { key: 'provincia', label: 'Prov.' },
    { key: 'partita_iva', label: 'P.IVA', className: 'font-mono' },
    { key: 'telefono', label: 'Telefono' },
    { key: 'actions', label: '', className: 'w-20' },
  ];

  return (
    <div data-testid="customers-page">
      <DataTable
        columns={columns}
        data={data}
        loading={loading}
        searchValue={search}
        onSearchChange={setSearch}
        onAdd={openNew}
        addLabel="Nuovo Cliente"
        testId="masterdata-table"
        renderRow={(item) => (
          <TableRow key={item.id} className="hover:bg-muted/60">
            <TableCell className="py-2 font-medium">{item.ragione_sociale}</TableCell>
            <TableCell className="py-2">{item.citta}</TableCell>
            <TableCell className="py-2">{item.provincia}</TableCell>
            <TableCell className="py-2 font-mono text-xs">{item.partita_iva}</TableCell>
            <TableCell className="py-2">{item.telefono}</TableCell>
            <TableCell className="py-2">
              <div className="flex gap-1">
                <Button asChild variant="ghost" size="icon" className="h-7 w-7" title="Cruscotto commerciale" aria-label="Cruscotto commerciale">
                  <Link to={`/anagrafiche/clienti/${item.id}/cruscotto`}><BarChart3 className="h-3 w-3" /></Link>
                </Button>
                <Button variant="ghost" size="icon" className="h-7 w-7" onClick={() => openEdit(item)}><Pencil className="h-3 w-3" /></Button>
                <Button variant="ghost" size="icon" className="h-7 w-7 text-destructive" onClick={() => item.id && handleDelete(item.id)}><Trash2 className="h-3 w-3" /></Button>
              </div>
            </TableCell>
          </TableRow>
        )}
      />

      <FormDialog open={dialogOpen} onClose={setDialogOpen} title={editId ? 'Modifica Cliente' : 'Nuovo Cliente'} onSubmit={handleSave} loading={saving}>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
          <div className="md:col-span-2 space-y-1.5"><Label>Ragione Sociale *</Label><Input value={form.ragione_sociale} onChange={e => setForm({ ...form, ragione_sociale: e.target.value })} required /></div>
          <div className="md:col-span-2 space-y-1.5">
            <Label>Indirizzo</Label>
            <AddressSearchInput
              value={form.indirizzo || ''}
              onChange={(v) => setForm(f => ({ ...f, indirizzo: v }))}
              onSelect={(r: DtoGeocodeResultDTO) => {
                setForm(f => ({ ...f, indirizzo: r.indirizzo || f.indirizzo, citta: r.citta || f.citta, cap: r.cap || f.cap, provincia: r.provincia || f.provincia, nazione: r.nazione || f.nazione, lat: r.lat ?? f.lat, lng: r.lng ?? f.lng }));
                setFlySignal(s => s + 1);
              }}
            />
          </div>
          <div className="space-y-1.5"><Label>Città</Label><Input value={form.citta} onChange={e => setForm({ ...form, citta: e.target.value })} /></div>
          <div className="space-y-1.5"><Label>CAP</Label><Input value={form.cap} onChange={e => setForm({ ...form, cap: e.target.value })} /></div>
          <div className="space-y-1.5"><Label>Provincia</Label><Input value={form.provincia} onChange={e => setForm({ ...form, provincia: e.target.value })} /></div>
          <div className="space-y-1.5"><Label>Nazione</Label><Input value={form.nazione} onChange={e => setForm({ ...form, nazione: e.target.value })} /></div>
          <div className="md:col-span-2 space-y-1.5">
            <Label>Posizione</Label>
            <MapPicker lat={form.lat} lng={form.lng} flyToSignal={flySignal} />
          </div>
          <div className="space-y-1.5"><Label>Partita IVA</Label><Input value={form.partita_iva} onChange={e => setForm({ ...form, partita_iva: e.target.value })} /></div>
          <div className="space-y-1.5"><Label>Codice Fiscale</Label><Input value={form.codice_fiscale} onChange={e => setForm({ ...form, codice_fiscale: e.target.value })} /></div>
          <div className="space-y-1.5"><Label>Telefono</Label><Input value={form.telefono} onChange={e => setForm({ ...form, telefono: e.target.value })} /></div>
          <div className="space-y-1.5"><Label>Email</Label><Input value={form.email} onChange={e => setForm({ ...form, email: e.target.value })} /></div>
          <div className="space-y-1.5"><Label>PEC</Label><Input value={form.pec} onChange={e => setForm({ ...form, pec: e.target.value })} /></div>
          <div className="space-y-1.5"><Label>Condizioni Pagamento</Label><Input value={form.condizioni_pagamento} onChange={e => setForm({ ...form, condizioni_pagamento: e.target.value })} /></div>
          <div className="md:col-span-2 space-y-1.5"><Label>Note</Label><Textarea value={form.note} onChange={e => setForm({ ...form, note: e.target.value })} rows={2} /></div>
          <div className="md:col-span-2 flex items-center gap-2">
            <Switch checked={form.richiede_rif_ordine} onCheckedChange={v => setForm({ ...form, richiede_rif_ordine: v })} />
            <Label>Richiede n. ordine rif. cliente per fatturazione</Label>
          </div>
        </div>
      </FormDialog>
    </div>
  );
}
