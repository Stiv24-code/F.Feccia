import { useState, useEffect } from 'react';
import { useGetMyAnagraficaQuery, useUpdateMyAnagraficaMutation } from '@/store/api/appApi';
import { getMutationErrorMessage } from '@/store/api/rtkQueryHelpers';
import type { DtoCustomerRequest } from '@/api/data-contracts';
import { Card } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import { Switch } from '@/components/ui/switch';
import { Button } from '@/components/ui/button';
import { Skeleton } from '@/components/ui/skeleton';
import { toast } from 'sonner';
import { Loader2 } from 'lucide-react';

const emptyForm: DtoCustomerRequest = { ragione_sociale: '', indirizzo: '', citta: '', cap: '', provincia: '', nazione: 'Italia', partita_iva: '', codice_fiscale: '', telefono: '', email: '', pec: '', condizioni_pagamento: '', note: '', richiede_rif_ordine: false };

export default function ClientAnagraficaPage() {
  const { data, isLoading: loading } = useGetMyAnagraficaQuery();
  const [updateMyAnagrafica, { isLoading: saving }] = useUpdateMyAnagraficaMutation();
  const [form, setForm] = useState<DtoCustomerRequest>(emptyForm);

  // Riempie il form solo quando arrivano i dati (mai sovrascrivere quello
  // che l'utente sta digitando con un refetch in background).
  useEffect(() => {
    if (!data) return;
    setForm({
      ragione_sociale: data.ragione_sociale || '', indirizzo: data.indirizzo || '', citta: data.citta || '',
      cap: data.cap || '', provincia: data.provincia || '', nazione: data.nazione || 'Italia',
      partita_iva: data.partita_iva || '', codice_fiscale: data.codice_fiscale || '', telefono: data.telefono || '',
      email: data.email || '', pec: data.pec || '', condizioni_pagamento: data.condizioni_pagamento || '',
      note: data.note || '', richiede_rif_ordine: data.richiede_rif_ordine || false,
    });
  }, [data]);

  const handleSave = async () => {
    try {
      await updateMyAnagrafica(form).unwrap();
      toast.success('Anagrafica aggiornata');
    } catch (e) { toast.error(getMutationErrorMessage(e) || 'Errore'); }
  };

  if (loading) {
    return (
      <div className="space-y-3">
        <Skeleton className="h-8 w-64" />
        <Skeleton className="h-96 rounded-xl" />
      </div>
    );
  }

  return (
    <div data-testid="client-anagrafica-page">
      <Card className="rounded-xl border shadow-sm p-5 space-y-4 max-w-3xl">
        <h2 className="text-xl font-bold" style={{ fontFamily: "'Space Grotesk', sans-serif" }}>La mia anagrafica</h2>
        <form onSubmit={(e) => { e.preventDefault(); handleSave(); }} className="space-y-4">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
            <div className="md:col-span-2 space-y-1.5"><Label>Ragione Sociale *</Label><Input value={form.ragione_sociale} onChange={e => setForm({ ...form, ragione_sociale: e.target.value })} required /></div>
            <div className="space-y-1.5"><Label>Indirizzo</Label><Input value={form.indirizzo} onChange={e => setForm({ ...form, indirizzo: e.target.value })} /></div>
            <div className="space-y-1.5"><Label>Città</Label><Input value={form.citta} onChange={e => setForm({ ...form, citta: e.target.value })} /></div>
            <div className="space-y-1.5"><Label>CAP</Label><Input value={form.cap} onChange={e => setForm({ ...form, cap: e.target.value })} /></div>
            <div className="space-y-1.5"><Label>Provincia</Label><Input value={form.provincia} onChange={e => setForm({ ...form, provincia: e.target.value })} /></div>
            <div className="space-y-1.5"><Label>Nazione</Label><Input value={form.nazione} onChange={e => setForm({ ...form, nazione: e.target.value })} /></div>
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
          <div className="flex justify-end">
            <Button type="submit" disabled={saving} data-testid="client-anagrafica-submit">
              {saving && <Loader2 className="h-4 w-4 animate-spin mr-2" />} Salva
            </Button>
          </div>
        </form>
      </Card>
    </div>
  );
}
