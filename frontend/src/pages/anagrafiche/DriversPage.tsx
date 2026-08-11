import { useState } from 'react';
import {
  useGetDriversQuery,
  useCreateDriverMutation,
  useUpdateDriverMutation,
  useDeleteDriverMutation,
  useGetDriverTripsQuery,
  useGetDriverUnavailabilityQuery,
  useCreateDriverUnavailabilityMutation,
  useDeleteDriverUnavailabilityMutation,
} from '@/store/api/appApi';
import { getMutationErrorMessage } from '@/store/api/rtkQueryHelpers';
import type { DtoDriverRequest, DtoDriverResponse, DtoDriverUnavailabilityRequest } from '@/api/data-contracts';
import { DataTable, type DataTableColumn } from '@/components/shared/DataTable';
import { FormDialog } from '@/components/shared/FormDialog';
import { StatusBadge } from '@/components/shared/StatusBadge';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Checkbox } from '@/components/ui/checkbox';
import { Popover, PopoverTrigger, PopoverContent } from '@/components/ui/popover';
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { TableRow, TableCell } from '@/components/ui/table';
import { toast } from 'sonner';
import { Pencil, Trash2, ChevronDown, Truck, Plane, Plus } from 'lucide-react';

type PatenteCategoria = NonNullable<DtoDriverRequest['patente']>[number];

// Categorie patente italiane (+ qualifiche professionali CQC/ADR) — mirror
// di models.PatenteCategorie sul backend, unica fonte di verità per gli
// `enums` esposti via swagger; qui serve solo l'elenco valori a runtime,
// il generatore non emette array a partire da union type TS.
const PATENTE_CATEGORIE: PatenteCategoria[] = ['AM', 'A1', 'A2', 'A', 'B1', 'B', 'BE', 'C1', 'C1E', 'C', 'CE', 'D1', 'D1E', 'D', 'DE', 'CQC', 'ADR'];

const MESI = ['gen', 'feb', 'mar', 'apr', 'mag', 'giu', 'lug', 'ago', 'set', 'ott', 'nov', 'dic'];

function formatFerieRange(da?: string | null, a?: string | null): string | null {
  if (!da || !a) return null;
  const [, mDa, dDa] = da.split('-');
  const [, mA, dA] = a.split('-');
  const giornoDa = `${parseInt(dDa, 10)}${mDa === mA ? '' : ' ' + MESI[parseInt(mDa, 10) - 1]}`;
  const giornoA = `${parseInt(dA, 10)} ${MESI[parseInt(mA, 10) - 1]}`;
  return `${giornoDa}–${giornoA}`;
}

const emptyForm: DtoDriverRequest = { nome: '', cognome: '', codice_fiscale: '', patente: [], scadenza_patente: '', telefono: '', email: '', note: '' };

function PatenteMultiSelect({ value, onChange }: { value: PatenteCategoria[]; onChange: (v: PatenteCategoria[]) => void }) {
  const [open, setOpen] = useState(false);
  const toggle = (cat: PatenteCategoria) => onChange(value.includes(cat) ? value.filter(c => c !== cat) : [...value, cat]);

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <button type="button" className="flex w-full items-center justify-between rounded-md border border-input bg-background px-3 py-2 text-sm">
          <span className="flex flex-wrap gap-1">
            {value.length === 0 ? <span className="text-muted-foreground">Seleziona categorie...</span> : value.map(c => <Badge key={c} variant="secondary" className="text-[10px]">{c}</Badge>)}
          </span>
          <ChevronDown className="h-3.5 w-3.5 shrink-0 text-muted-foreground" />
        </button>
      </PopoverTrigger>
      <PopoverContent className="w-64 p-2" align="start">
        <div className="grid grid-cols-3 gap-2">
          {PATENTE_CATEGORIE.map(cat => (
            <label key={cat} className="flex items-center gap-1.5 text-sm cursor-pointer">
              <Checkbox checked={value.includes(cat)} onCheckedChange={() => toggle(cat)} />
              {cat}
            </label>
          ))}
        </div>
      </PopoverContent>
    </Popover>
  );
}

type Motivo = NonNullable<DtoDriverUnavailabilityRequest['motivo']>;

const MOTIVO_LABELS: Record<Motivo, string> = { ferie: 'Ferie', malattia: 'Malattia', permesso: 'Permesso', altro: 'Altro' };

const emptyFerieForm = { data_da: '', data_a: '', motivo: 'ferie' as Motivo, note: '' };

function DriverFerieDialog({ driver, onClose }: { driver: DtoDriverResponse; onClose: () => void }) {
  const [form, setForm] = useState(emptyFerieForm);
  const { data: periods = [], isLoading } = useGetDriverUnavailabilityQuery(driver.id || '', { skip: !driver.id });
  const [createPeriod, { isLoading: creating }] = useCreateDriverUnavailabilityMutation();
  const [deletePeriod] = useDeleteDriverUnavailabilityMutation();

  const handleAdd = async () => {
    if (!driver.id || !form.data_da || !form.data_a) return;
    try {
      await createPeriod({
        autista_id: driver.id,
        autista_nome: `${driver.nome || ''} ${driver.cognome || ''}`.trim(),
        data_da: form.data_da,
        data_a: form.data_a,
        motivo: form.motivo,
        note: form.note || undefined,
      }).unwrap();
      toast.success('Periodo aggiunto');
      setForm(emptyFerieForm);
    } catch (e) { toast.error(getMutationErrorMessage(e) || 'Errore'); }
  };

  const handleDelete = async (id: string) => {
    if (!window.confirm('Eliminare questo periodo?')) return;
    try { await deletePeriod(id).unwrap(); toast.success('Eliminato'); } catch (e) { toast.error(getMutationErrorMessage(e) || 'Errore'); }
  };

  const sorted = [...periods].sort((a, b) => (a.data_da || '').localeCompare(b.data_da || ''));

  return (
    <Dialog open onOpenChange={onClose}>
      <DialogContent className="max-w-lg max-h-[85vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle style={{ fontFamily: "'Space Grotesk', sans-serif" }}>Ferie e assenze — {driver.nome} {driver.cognome}</DialogTitle>
        </DialogHeader>

        <div className="grid grid-cols-2 gap-2 items-end border-b pb-4 mb-2">
          <div className="space-y-1.5"><Label>Dal</Label><Input type="date" value={form.data_da} onChange={e => setForm({ ...form, data_da: e.target.value })} /></div>
          <div className="space-y-1.5"><Label>Al</Label><Input type="date" value={form.data_a} onChange={e => setForm({ ...form, data_a: e.target.value })} /></div>
          <div className="space-y-1.5">
            <Label>Motivo</Label>
            <Select value={form.motivo} onValueChange={(v) => setForm({ ...form, motivo: v as Motivo })}>
              <SelectTrigger className="h-9 text-sm"><SelectValue /></SelectTrigger>
              <SelectContent>
                {Object.entries(MOTIVO_LABELS).map(([value, label]) => <SelectItem key={value} value={value}>{label}</SelectItem>)}
              </SelectContent>
            </Select>
          </div>
          <div className="space-y-1.5"><Label>Note</Label><Input value={form.note} onChange={e => setForm({ ...form, note: e.target.value })} /></div>
          <div className="col-span-2">
            <Button size="sm" className="w-full gap-1.5" disabled={!form.data_da || !form.data_a || creating} onClick={handleAdd}>
              <Plus className="h-3.5 w-3.5" /> Aggiungi periodo
            </Button>
          </div>
        </div>

        {isLoading ? (
          <p className="text-sm text-muted-foreground">Caricamento...</p>
        ) : sorted.length === 0 ? (
          <p className="text-sm text-muted-foreground">Nessun periodo registrato.</p>
        ) : (
          <div className="space-y-2">
            {sorted.map(p => (
              <div key={p.id} className="flex items-center justify-between rounded-md border px-3 py-2 text-sm">
                <div>
                  <div className="font-medium">{MOTIVO_LABELS[(p.motivo as Motivo) || 'altro']}</div>
                  <div className="text-muted-foreground text-xs">{p.data_da || '-'} → {p.data_a || '-'}{p.note ? ` · ${p.note}` : ''}</div>
                </div>
                <Button variant="ghost" size="icon" className="h-7 w-7 text-destructive" onClick={() => p.id && handleDelete(p.id)}><Trash2 className="h-3 w-3" /></Button>
              </div>
            ))}
          </div>
        )}
      </DialogContent>
    </Dialog>
  );
}

function DriverTripsDialog({ driver, onClose }: { driver: DtoDriverResponse; onClose: () => void }) {
  const { data: trips = [], isLoading } = useGetDriverTripsQuery(driver.id || '', { skip: !driver.id });
  return (
    <Dialog open onOpenChange={onClose}>
      <DialogContent className="max-w-lg max-h-[85vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle style={{ fontFamily: "'Space Grotesk', sans-serif" }}>Viaggi assegnati — {driver.nome} {driver.cognome}</DialogTitle>
        </DialogHeader>
        {isLoading ? (
          <p className="text-sm text-muted-foreground">Caricamento...</p>
        ) : trips.length === 0 ? (
          <p className="text-sm text-muted-foreground">Nessun viaggio assegnato.</p>
        ) : (
          <div className="space-y-2">
            {trips.map(t => (
              <div key={t.id} className="flex items-center justify-between rounded-md border px-3 py-2 text-sm">
                <div>
                  <div className="font-medium">{t.motrice?.targa || '-'}</div>
                  <div className="text-muted-foreground text-xs">{t.data_partenza || '-'} → {t.data_arrivo || '-'}</div>
                </div>
                <StatusBadge stato={t.stato} />
              </div>
            ))}
          </div>
        )}
      </DialogContent>
    </Dialog>
  );
}

export default function DriversPage() {
  const [search, setSearch] = useState('');
  const [dialogOpen, setDialogOpen] = useState(false);
  const [form, setForm] = useState<DtoDriverRequest>(emptyForm);
  const [editId, setEditId] = useState<string | null>(null);
  const [tripsDriver, setTripsDriver] = useState<DtoDriverResponse | null>(null);
  const [ferieDriver, setFerieDriver] = useState<DtoDriverResponse | null>(null);

  const { data = [], isLoading: loading } = useGetDriversQuery(search);
  const [createDriver, { isLoading: creating }] = useCreateDriverMutation();
  const [updateDriver, { isLoading: updating }] = useUpdateDriverMutation();
  const [deleteDriver] = useDeleteDriverMutation();
  const saving = creating || updating;

  const openNew = () => { setForm(emptyForm); setEditId(null); setDialogOpen(true); };
  const openEdit = (item: DtoDriverResponse) => { setForm({ nome: item.nome || '', cognome: item.cognome || '', codice_fiscale: item.codice_fiscale || '', patente: item.patente || [], scadenza_patente: item.scadenza_patente || '', telefono: item.telefono || '', email: item.email || '', note: item.note || '' }); setEditId(item.id || null); setDialogOpen(true); };

  const handleSave = async () => {
    try {
      if (editId) { await updateDriver({ id: editId, body: form }).unwrap(); toast.success('Autista aggiornato'); } else { await createDriver(form).unwrap(); toast.success('Autista creato'); }
      setDialogOpen(false);
    } catch (e) { toast.error(getMutationErrorMessage(e) || 'Errore'); }
  };

  const handleDelete = async (id: string) => {
    if (!window.confirm('Eliminare questo autista?')) return;
    try { await deleteDriver(id).unwrap(); toast.success('Eliminato'); } catch (e) { toast.error(getMutationErrorMessage(e) || 'Errore'); }
  };

  const columns: DataTableColumn[] = [{ key: 'cognome', label: 'Cognome' }, { key: 'nome', label: 'Nome' }, { key: 'patente', label: 'Patente' }, { key: 'ferie', label: 'Prossime ferie' }, { key: 'telefono', label: 'Telefono' }, { key: 'actions', label: '', className: 'w-28' }];

  return (
    <div data-testid="drivers-page">
      <DataTable columns={columns} data={data} loading={loading} searchValue={search} onSearchChange={setSearch} onAdd={openNew} addLabel="Nuovo Autista" testId="masterdata-table"
        renderRow={(item) => (
          <TableRow key={item.id} className="hover:bg-muted/60">
            <TableCell className="py-2 font-medium">{item.cognome}</TableCell>
            <TableCell className="py-2">{item.nome}</TableCell>
            <TableCell className="py-2"><div className="flex flex-wrap gap-1">{(item.patente || []).map(c => <Badge key={c} variant="secondary" className="text-[10px]">{c}</Badge>)}</div></TableCell>
            <TableCell className="py-2 text-muted-foreground">{formatFerieRange(item.prossime_ferie_da, item.prossime_ferie_a) || '-'}</TableCell>
            <TableCell className="py-2">{item.telefono}</TableCell>
            <TableCell className="py-2">
              <div className="flex gap-1">
                <Button variant="ghost" size="icon" className="h-7 w-7" title="Viaggi assegnati" onClick={() => setTripsDriver(item)}><Truck className="h-3 w-3" /></Button>
                <Button variant="ghost" size="icon" className="h-7 w-7" title="Ferie e assenze" onClick={() => setFerieDriver(item)}><Plane className="h-3 w-3" /></Button>
                <Button variant="ghost" size="icon" className="h-7 w-7" onClick={() => openEdit(item)}><Pencil className="h-3 w-3" /></Button>
                <Button variant="ghost" size="icon" className="h-7 w-7 text-destructive" onClick={() => item.id && handleDelete(item.id)}><Trash2 className="h-3 w-3" /></Button>
              </div>
            </TableCell>
          </TableRow>
        )}
      />
      <FormDialog open={dialogOpen} onClose={setDialogOpen} title={editId ? 'Modifica Autista' : 'Nuovo Autista'} onSubmit={handleSave} loading={saving}>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
          <div className="space-y-1.5"><Label>Nome *</Label><Input value={form.nome} onChange={e => setForm({ ...form, nome: e.target.value })} required /></div>
          <div className="space-y-1.5"><Label>Cognome *</Label><Input value={form.cognome} onChange={e => setForm({ ...form, cognome: e.target.value })} required /></div>
          <div className="space-y-1.5 md:col-span-2"><Label>Patente</Label><PatenteMultiSelect value={form.patente || []} onChange={v => setForm({ ...form, patente: v })} /></div>
          <div className="space-y-1.5"><Label>Scadenza Patente</Label><Input type="date" value={form.scadenza_patente} onChange={e => setForm({ ...form, scadenza_patente: e.target.value })} /></div>
          <div className="space-y-1.5"><Label>Telefono</Label><Input value={form.telefono} onChange={e => setForm({ ...form, telefono: e.target.value })} /></div>
          <div className="space-y-1.5"><Label>Email</Label><Input value={form.email} onChange={e => setForm({ ...form, email: e.target.value })} /></div>
          <div className="space-y-1.5"><Label>Codice Fiscale</Label><Input value={form.codice_fiscale} onChange={e => setForm({ ...form, codice_fiscale: e.target.value })} /></div>
        </div>
      </FormDialog>
      {tripsDriver && <DriverTripsDialog driver={tripsDriver} onClose={() => setTripsDriver(null)} />}
      {ferieDriver && <DriverFerieDialog driver={ferieDriver} onClose={() => setFerieDriver(null)} />}
    </div>
  );
}
