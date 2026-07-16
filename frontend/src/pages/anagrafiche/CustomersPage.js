import { useState, useEffect, useCallback } from 'react';
import { Link } from 'react-router-dom';
import { getCustomers, createCustomer, updateCustomer, deleteCustomer } from '@/lib/api';
import { DataTable } from '@/components/shared/DataTable';
import { FormDialog } from '@/components/shared/FormDialog';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Switch } from '@/components/ui/switch';
import { Textarea } from '@/components/ui/textarea';
import { Button } from '@/components/ui/button';
import { TableRow, TableCell } from '@/components/ui/table';
import { toast } from 'sonner';
import { Pencil, Trash2, BarChart3 } from 'lucide-react';

const emptyForm = { ragione_sociale: '', indirizzo: '', citta: '', cap: '', provincia: '', nazione: 'Italia', partita_iva: '', codice_fiscale: '', telefono: '', email: '', pec: '', condizioni_pagamento: '', note: '', richiede_rif_ordine: false };

export default function CustomersPage() {
  const [data, setData] = useState([]);
  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState('');
  const [dialogOpen, setDialogOpen] = useState(false);
  const [form, setForm] = useState(emptyForm);
  const [editId, setEditId] = useState(null);
  const [saving, setSaving] = useState(false);

  // API imports e state setters sono riferimenti stabili in React
  const fetchData = useCallback(() => { setLoading(true); getCustomers(search).then(r => setData(r.data)).finally(() => setLoading(false)); }, [search]);
  useEffect(() => { fetchData(); }, [fetchData]);

  const openNew = () => { setForm(emptyForm); setEditId(null); setDialogOpen(true); };
  const openEdit = (item) => { setForm({ ragione_sociale: item.ragione_sociale, indirizzo: item.indirizzo || '', citta: item.citta || '', cap: item.cap || '', provincia: item.provincia || '', nazione: item.nazione || 'Italia', partita_iva: item.partita_iva || '', codice_fiscale: item.codice_fiscale || '', telefono: item.telefono || '', email: item.email || '', pec: item.pec || '', condizioni_pagamento: item.condizioni_pagamento || '', note: item.note || '', richiede_rif_ordine: item.richiede_rif_ordine || false }); setEditId(item.id); setDialogOpen(true); };

  const handleSave = async () => {
    setSaving(true);
    try {
      if (editId) { await updateCustomer(editId, form); toast.success('Cliente aggiornato'); }
      else { await createCustomer(form); toast.success('Cliente creato'); }
      setDialogOpen(false); fetchData();
    } catch (e) { toast.error(e.response?.data?.detail || 'Errore'); } finally { setSaving(false); }
  };

  const handleDelete = async (id) => {
    if (!window.confirm('Eliminare questo cliente?')) return;
    try { await deleteCustomer(id); toast.success('Cliente eliminato'); fetchData(); } catch(e) { toast.error('Errore'); }
  };

  const columns = [
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
                <Button variant="ghost" size="icon" className="h-7 w-7 text-destructive" onClick={() => handleDelete(item.id)}><Trash2 className="h-3 w-3" /></Button>
              </div>
            </TableCell>
          </TableRow>
        )}
      />

      <FormDialog open={dialogOpen} onClose={setDialogOpen} title={editId ? 'Modifica Cliente' : 'Nuovo Cliente'} onSubmit={handleSave} loading={saving}>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
          <div className="md:col-span-2 space-y-1.5"><Label>Ragione Sociale *</Label><Input value={form.ragione_sociale} onChange={e => setForm({...form, ragione_sociale: e.target.value})} required /></div>
          <div className="space-y-1.5"><Label>Indirizzo</Label><Input value={form.indirizzo} onChange={e => setForm({...form, indirizzo: e.target.value})} /></div>
          <div className="space-y-1.5"><Label>Città</Label><Input value={form.citta} onChange={e => setForm({...form, citta: e.target.value})} /></div>
          <div className="space-y-1.5"><Label>CAP</Label><Input value={form.cap} onChange={e => setForm({...form, cap: e.target.value})} /></div>
          <div className="space-y-1.5"><Label>Provincia</Label><Input value={form.provincia} onChange={e => setForm({...form, provincia: e.target.value})} /></div>
          <div className="space-y-1.5"><Label>Nazione</Label><Input value={form.nazione} onChange={e => setForm({...form, nazione: e.target.value})} /></div>
          <div className="space-y-1.5"><Label>Partita IVA</Label><Input value={form.partita_iva} onChange={e => setForm({...form, partita_iva: e.target.value})} /></div>
          <div className="space-y-1.5"><Label>Codice Fiscale</Label><Input value={form.codice_fiscale} onChange={e => setForm({...form, codice_fiscale: e.target.value})} /></div>
          <div className="space-y-1.5"><Label>Telefono</Label><Input value={form.telefono} onChange={e => setForm({...form, telefono: e.target.value})} /></div>
          <div className="space-y-1.5"><Label>Email</Label><Input value={form.email} onChange={e => setForm({...form, email: e.target.value})} /></div>
          <div className="space-y-1.5"><Label>PEC</Label><Input value={form.pec} onChange={e => setForm({...form, pec: e.target.value})} /></div>
          <div className="space-y-1.5"><Label>Condizioni Pagamento</Label><Input value={form.condizioni_pagamento} onChange={e => setForm({...form, condizioni_pagamento: e.target.value})} /></div>
          <div className="md:col-span-2 space-y-1.5"><Label>Note</Label><Textarea value={form.note} onChange={e => setForm({...form, note: e.target.value})} rows={2} /></div>
          <div className="md:col-span-2 flex items-center gap-2">
            <Switch checked={form.richiede_rif_ordine} onCheckedChange={v => setForm({...form, richiede_rif_ordine: v})} />
            <Label>Richiede n. ordine rif. cliente per fatturazione</Label>
          </div>
        </div>
      </FormDialog>
    </div>
  );
}
