import { useState } from 'react';
import { getVehicleTemperature, setVehicleTemperatureThresholds } from '@/lib/api';
import {
  useGetVehiclesQuery,
  useCreateVehicleMutation,
  useUpdateVehicleMutation,
  useDeleteVehicleMutation,
} from '@/store/api/appApi';
import { formatEuro } from '@/lib/format';
import { DataTable } from '@/components/shared/DataTable';
import { FormDialog } from '@/components/shared/FormDialog';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Button } from '@/components/ui/button';
import { TableRow, TableCell } from '@/components/ui/table';
import { Badge } from '@/components/ui/badge';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog';
import { Select, SelectTrigger, SelectValue, SelectContent, SelectItem } from '@/components/ui/select';
import { toast } from 'sonner';
import { Pencil, Trash2, Satellite, ExternalLink, Thermometer, AlertTriangle } from 'lucide-react';
import { LineChart, Line, XAxis, YAxis, Tooltip, ResponsiveContainer, ReferenceLine, CartesianGrid } from 'recharts';
import { format, parseISO, isValid } from 'date-fns';

const emptyForm = { targa: '', tipo_veicolo: 'motrice', marca: '', modello: '', anno: 0, scompartature: 1, portata_kg: 0, note: '', gps_tracker_url: '', gps_tracker_tipo: '' };

export default function VehiclesPage() {
  const [search, setSearch] = useState('');
  const [dialogOpen, setDialogOpen] = useState(false);
  const [form, setForm] = useState(emptyForm);
  const [editId, setEditId] = useState(null);

  const { data = [], isLoading: loading, refetch } = useGetVehiclesQuery(search);
  const [createVehicle, { isLoading: creating }] = useCreateVehicleMutation();
  const [updateVehicle, { isLoading: updating }] = useUpdateVehicleMutation();
  const [deleteVehicle] = useDeleteVehicleMutation();
  const saving = creating || updating;

  // Temperatura cisterna (#38)
  const [tempDialog, setTempDialog] = useState(null);  // veicolo selezionato
  const [tempReadings, setTempReadings] = useState([]);
  const [tempLoading, setTempLoading] = useState(false);
  const [tempForm, setTempForm] = useState({ temp_min: '', temp_max: '' });
  const [tempSaving, setTempSaving] = useState(false);

  const openNew = () => { setForm(emptyForm); setEditId(null); setDialogOpen(true); };
  const openEdit = (item) => { setForm({ targa: item.targa, tipo_veicolo: item.tipo_veicolo || 'motrice', marca: item.marca || '', modello: item.modello || '', anno: item.anno || 0, scompartature: item.scompartature || 1, portata_kg: item.portata_kg || 0, note: item.note || '', gps_tracker_url: item.gps_tracker_url || '', gps_tracker_tipo: item.gps_tracker_tipo || '' }); setEditId(item.id); setDialogOpen(true); };

  const handleSave = async () => {
    try {
      if (editId) { await updateVehicle({ id: editId, body: form }).unwrap(); toast.success('Veicolo aggiornato'); }
      else { await createVehicle(form).unwrap(); toast.success('Veicolo creato'); }
      setDialogOpen(false);
    } catch (e) { toast.error('Errore'); }
  };

  const handleDelete = async (id) => {
    if (!window.confirm('Eliminare questo veicolo?')) return;
    try { await deleteVehicle(id).unwrap(); toast.success('Eliminato'); } catch(e) { toast.error('Errore'); }
  };

  const openTemp = async (vehicle) => {
    setTempDialog(vehicle);
    setTempForm({
      temp_min: vehicle.temp_min ?? '',
      temp_max: vehicle.temp_max ?? '',
    });
    setTempReadings([]);
    setTempLoading(true);
    try {
      const r = await getVehicleTemperature(vehicle.id, { limit: 200 });
      setTempReadings(r.data);
    } catch (e) {
      toast.error('Errore caricamento letture');
    } finally {
      setTempLoading(false);
    }
  };

  const handleSaveThresholds = async () => {
    if (!tempDialog) return;
    setTempSaving(true);
    try {
      const payload = {};
      if (tempForm.temp_min !== '') payload.temp_min = parseFloat(tempForm.temp_min);
      if (tempForm.temp_max !== '') payload.temp_max = parseFloat(tempForm.temp_max);
      await setVehicleTemperatureThresholds(tempDialog.id, payload);
      toast.success('Soglie aggiornate');
      setTempDialog(null);
      refetch();
    } catch (e) {
      toast.error(e.response?.data?.detail || 'Errore salvataggio soglie');
    } finally {
      setTempSaving(false);
    }
  };

  const tipoLabel = { motrice: 'Motrice', rimorchio: 'Rimorchio', rimorchio_isotermico: 'Rim. Isotermico', container: 'Container' };
  const columns = [
    { key: 'targa', label: 'Targa' },
    { key: 'tipo', label: 'Tipo' },
    { key: 'marca', label: 'Marca' },
    { key: 'portata', label: 'Portata (Kg)', className: 'text-right' },
    { key: 'gps', label: 'GPS' },
    { key: 'actions', label: '', className: 'w-20' },
  ];

  return (
    <div data-testid="vehicles-page">
      <DataTable columns={columns} data={data} loading={loading} searchValue={search} onSearchChange={setSearch} onAdd={openNew} addLabel="Nuovo Veicolo" testId="masterdata-table"
        renderRow={(item) => (
          <TableRow key={item.id} className="hover:bg-muted/60">
            <TableCell className="py-2 font-mono font-medium">{item.targa}</TableCell>
            <TableCell className="py-2">{tipoLabel[item.tipo_veicolo] || item.tipo_veicolo}</TableCell>
            <TableCell className="py-2">{item.marca}</TableCell>
            <TableCell className="py-2 text-right tabular-nums">{formatEuro(item.portata_kg)}</TableCell>
            <TableCell className="py-2">
              {item.gps_tracker_url ? (
                <a href={item.gps_tracker_url} target="_blank" rel="noopener noreferrer" className="inline-flex items-center gap-1 text-xs text-primary hover:underline">
                  <Satellite className="h-3 w-3" />
                  {item.gps_active ? <span className="h-1.5 w-1.5 rounded-full bg-emerald-500 animate-pulse" /> : null}
                  {item.gps_tracker_tipo || 'GPS'}
                  <ExternalLink className="h-2.5 w-2.5" />
                </a>
              ) : <span className="text-xs text-muted-foreground">—</span>}
            </TableCell>
            <TableCell className="py-2">
              <div className="flex gap-1">
                <Button
                  variant="ghost" size="icon"
                  className={`h-7 w-7 ${item.last_temp_alert ? 'text-destructive' : ''}`}
                  title={item.last_temp_alert ? 'Temperatura fuori soglia' : 'Temperatura cisterna'}
                  aria-label="Temperatura cisterna"
                  onClick={() => openTemp(item)}
                >
                  {item.last_temp_alert ? <AlertTriangle className="h-3 w-3" /> : <Thermometer className="h-3 w-3" />}
                </Button>
                <Button variant="ghost" size="icon" className="h-7 w-7" onClick={() => openEdit(item)}><Pencil className="h-3 w-3" /></Button>
                <Button variant="ghost" size="icon" className="h-7 w-7 text-destructive" onClick={() => handleDelete(item.id)}><Trash2 className="h-3 w-3" /></Button>
              </div>
            </TableCell>
          </TableRow>
        )}
      />
      <FormDialog open={dialogOpen} onClose={setDialogOpen} title={editId ? 'Modifica Veicolo' : 'Nuovo Veicolo'} onSubmit={handleSave} loading={saving}>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
          <div className="space-y-1.5"><Label>Targa *</Label><Input value={form.targa} onChange={e => setForm({...form, targa: e.target.value})} required /></div>
          <div className="space-y-1.5">
            <Label>Tipo Veicolo</Label>
            <Select value={form.tipo_veicolo} onValueChange={v => setForm({...form, tipo_veicolo: v})}>
              <SelectTrigger><SelectValue /></SelectTrigger>
              <SelectContent>
                <SelectItem value="motrice">Motrice</SelectItem>
                <SelectItem value="rimorchio">Rimorchio</SelectItem>
                <SelectItem value="rimorchio_isotermico">Rimorchio Isotermico</SelectItem>
                <SelectItem value="container">Container</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <div className="space-y-1.5"><Label>Marca</Label><Input value={form.marca} onChange={e => setForm({...form, marca: e.target.value})} /></div>
          <div className="space-y-1.5"><Label>Modello</Label><Input value={form.modello} onChange={e => setForm({...form, modello: e.target.value})} /></div>
          <div className="space-y-1.5"><Label>Portata (Kg)</Label><Input type="number" value={form.portata_kg} onChange={e => setForm({...form, portata_kg: Number(e.target.value)})} /></div>
          <div className="space-y-1.5"><Label>Scompartature</Label><Input type="number" value={form.scompartature} onChange={e => setForm({...form, scompartature: Number(e.target.value)})} /></div>
          <div className="md:col-span-2 pt-2 border-t">
            <p className="text-xs font-semibold mb-2 flex items-center gap-1.5"><Satellite className="h-3.5 w-3.5" /> Tracking GPS</p>
          </div>
          <div className="md:col-span-2 space-y-1.5">
            <Label>URL Portale GPS Tracker</Label>
            <Input value={form.gps_tracker_url} onChange={e => setForm({...form, gps_tracker_url: e.target.value})} placeholder="https://fleet.garmin.com/vehicle/..." />
            <p className="text-[10px] text-muted-foreground">Link diretto al portale del provider GPS per questo mezzo (Garmin, Verizon Connect, Geotab, ecc.)</p>
          </div>
          <div className="space-y-1.5">
            <Label>Provider GPS</Label>
            <Select value={form.gps_tracker_tipo || 'none'} onValueChange={v => setForm({...form, gps_tracker_tipo: v === 'none' ? '' : v})}>
              <SelectTrigger><SelectValue placeholder="Seleziona provider" /></SelectTrigger>
              <SelectContent>
                <SelectItem value="none">Nessuno</SelectItem>
                <SelectItem value="garmin">Garmin Fleet</SelectItem>
                <SelectItem value="verizon">Verizon Connect</SelectItem>
                <SelectItem value="geotab">Geotab</SelectItem>
                <SelectItem value="teltonika">Teltonika</SelectItem>
                <SelectItem value="webfleet">Webfleet (TomTom)</SelectItem>
                <SelectItem value="samsara">Samsara</SelectItem>
                <SelectItem value="custom">Altro / Custom</SelectItem>
              </SelectContent>
            </Select>
          </div>
        </div>
      </FormDialog>

      {/* Dialog temperatura cisterna (#38) */}
      <Dialog open={!!tempDialog} onOpenChange={(o) => !o && setTempDialog(null)}>
        <DialogContent className="max-w-3xl">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <Thermometer className="h-4 w-4" />
              Temperatura cisterna {tempDialog?.targa ? `· ${tempDialog.targa}` : ''}
              {tempDialog?.last_temp_alert && (
                <Badge variant="destructive" className="text-[10px]">Fuori soglia</Badge>
              )}
            </DialogTitle>
          </DialogHeader>
          {tempDialog && (
            <div className="space-y-4 text-sm">
              {/* Stato corrente */}
              <div className="grid grid-cols-3 gap-3">
                <div className="rounded-lg border bg-muted/30 p-3">
                  <p className="text-[11px] text-muted-foreground">Ultima lettura</p>
                  <p className="text-lg font-bold tabular-nums">
                    {tempDialog.last_temp_celsius != null
                      ? `${tempDialog.last_temp_celsius.toFixed(1)} °C`
                      : '—'}
                  </p>
                  <p className="text-[10px] text-muted-foreground">
                    {tempDialog.last_temp_ts ? new Date(tempDialog.last_temp_ts).toLocaleString('it-IT') : 'Mai'}
                  </p>
                </div>
                <div className="rounded-lg border p-3">
                  <p className="text-[11px] text-muted-foreground">Soglia min</p>
                  <p className="text-lg font-bold tabular-nums">
                    {tempDialog.temp_min != null ? `${tempDialog.temp_min.toFixed(1)} °C` : '—'}
                  </p>
                </div>
                <div className="rounded-lg border p-3">
                  <p className="text-[11px] text-muted-foreground">Soglia max</p>
                  <p className="text-lg font-bold tabular-nums">
                    {tempDialog.temp_max != null ? `${tempDialog.temp_max.toFixed(1)} °C` : '—'}
                  </p>
                </div>
              </div>

              {/* Grafico letture */}
              <div className="border rounded-lg p-3">
                <p className="text-xs font-semibold uppercase text-muted-foreground mb-2">
                  Andamento ultime {tempReadings.length} letture
                </p>
                <div className="h-48">
                  {tempLoading ? (
                    <div className="flex items-center justify-center h-full text-xs text-muted-foreground">Caricamento…</div>
                  ) : tempReadings.length === 0 ? (
                    <div className="flex items-center justify-center h-full text-xs text-muted-foreground">
                      Nessuna lettura. Avviare il sensore IoT o lo script di simulazione.
                    </div>
                  ) : (
                    <ResponsiveContainer width="100%" height="100%">
                      <LineChart data={[...tempReadings].reverse().map(r => ({
                        ...r,
                        label: (() => { const d = parseISO(r.ts); return isValid(d) ? format(d, 'HH:mm') : ''; })(),
                      }))}>
                        <CartesianGrid strokeDasharray="3 3" stroke="hsl(214 18% 88%)" />
                        <XAxis dataKey="label" tick={{ fontSize: 10 }} stroke="hsl(215 16% 38%)" />
                        <YAxis tick={{ fontSize: 10 }} stroke="hsl(215 16% 38%)" unit=" °C" />
                        <Tooltip contentStyle={{ borderRadius: 8, fontSize: 12 }} formatter={(v) => `${v.toFixed(1)} °C`} />
                        {tempDialog.temp_min != null && (
                          <ReferenceLine y={tempDialog.temp_min} stroke="#3B82F6" strokeDasharray="4 4" label={{ value: 'min', fontSize: 10, fill: '#3B82F6' }} />
                        )}
                        {tempDialog.temp_max != null && (
                          <ReferenceLine y={tempDialog.temp_max} stroke="#EF4444" strokeDasharray="4 4" label={{ value: 'max', fontSize: 10, fill: '#EF4444' }} />
                        )}
                        <Line type="monotone" dataKey="temp_celsius" stroke="#0EA5A6" strokeWidth={2} dot={false} />
                      </LineChart>
                    </ResponsiveContainer>
                  )}
                </div>
              </div>

              {/* Form soglie */}
              <div className="border rounded-lg p-3 space-y-2">
                <p className="text-xs font-semibold uppercase text-muted-foreground">Configura soglie</p>
                <div className="grid grid-cols-2 gap-3">
                  <div>
                    <Label>Min (°C)</Label>
                    <Input
                      type="number" step="0.1"
                      value={tempForm.temp_min}
                      onChange={e => setTempForm({...tempForm, temp_min: e.target.value})}
                      placeholder="es. 2"
                    />
                  </div>
                  <div>
                    <Label>Max (°C)</Label>
                    <Input
                      type="number" step="0.1"
                      value={tempForm.temp_max}
                      onChange={e => setTempForm({...tempForm, temp_max: e.target.value})}
                      placeholder="es. 8"
                    />
                  </div>
                </div>
                <p className="text-[11px] text-muted-foreground">
                  Lascia vuoto per disabilitare un lato. Tipica catena fredda: 2-8 °C.
                </p>
              </div>
            </div>
          )}
          <DialogFooter>
            <Button variant="outline" onClick={() => setTempDialog(null)}>Chiudi</Button>
            <Button onClick={handleSaveThresholds} disabled={tempSaving}>
              {tempSaving ? 'Salvataggio…' : 'Salva soglie'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
