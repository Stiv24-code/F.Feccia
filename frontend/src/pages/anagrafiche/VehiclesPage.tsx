import { useMemo, useState } from 'react';
import {
  useGetMotriciQuery, useCreateMotriceMutation, useUpdateMotriceMutation, useDeleteMotriceMutation,
  useGetSemirimorchiQuery, useCreateSemirimorchioMutation, useUpdateSemirimorchioMutation, useDeleteSemirimorchioMutation,
} from '@/store/api/appApi';
import { getMutationErrorMessage } from '@/store/api/rtkQueryHelpers';
import type { DtoMotriceRequest, DtoMotriceResponse, DtoSemirimorchioRequest, DtoSemirimorchioResponse } from '@/api/data-contracts';
import { formatEuro } from '@/lib/format';
import { DataTable, type DataTableColumn } from '@/components/shared/DataTable';
import { FormDialog } from '@/components/shared/FormDialog';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { TableRow, TableCell } from '@/components/ui/table';
import { toast } from 'sonner';
import { Pencil, Trash2, Plus } from 'lucide-react';

type MezzoTipo = 'Motrice' | 'Semirimorchio';

interface MezzoRow {
  id: string;
  targa: string;
  tipo: MezzoTipo;
  descrizione: string;
  portataKg: number;
}

const emptyMotriceForm: DtoMotriceRequest = { targa: '', marca: '', modello: '', anno: 0, portata_kg: 0, note: '' };
const emptySemirimorchioForm: DtoSemirimorchioRequest = { targa: '', tipo: '', scompartature: 1, portata_kg: 0, note: '' };

const tipoBadgeClass: Record<MezzoTipo, string> = {
  Motrice: 'bg-blue-50 text-blue-700 border-blue-200 dark:bg-blue-500/15 dark:text-blue-300 dark:border-blue-500/30',
  Semirimorchio: 'bg-purple-50 text-purple-700 border-purple-200 dark:bg-purple-500/15 dark:text-purple-300 dark:border-purple-500/30',
};

function motriceToRow(m: DtoMotriceResponse): MezzoRow {
  return {
    id: m.id || '',
    targa: m.targa || '',
    tipo: 'Motrice',
    descrizione: [m.marca, m.modello].filter(Boolean).join(' ') || '—',
    portataKg: m.portata_kg || 0,
  };
}

function semirimorchioToRow(s: DtoSemirimorchioResponse): MezzoRow {
  return {
    id: s.id || '',
    targa: s.targa || '',
    tipo: 'Semirimorchio',
    descrizione: s.tipo ? `${s.tipo}${s.scompartature ? ' · ' + s.scompartature + ' scomp.' : ''}` : '—',
    portataKg: s.portata_kg || 0,
  };
}

function MotriceFormDialog({ open, onClose, editItem }: { open: boolean; onClose: () => void; editItem: DtoMotriceResponse | null }) {
  const [form, setForm] = useState<DtoMotriceRequest>(editItem ? {
    targa: editItem.targa || '', marca: editItem.marca || '', modello: editItem.modello || '',
    anno: editItem.anno || 0, portata_kg: editItem.portata_kg || 0, note: editItem.note || '',
  } : emptyMotriceForm);
  const [createMotrice, { isLoading: creating }] = useCreateMotriceMutation();
  const [updateMotrice, { isLoading: updating }] = useUpdateMotriceMutation();

  const handleSave = async () => {
    try {
      if (editItem?.id) { await updateMotrice({ id: editItem.id, body: form }).unwrap(); toast.success('Motrice aggiornata'); }
      else { await createMotrice(form).unwrap(); toast.success('Motrice creata'); }
      onClose();
    } catch (e) { toast.error(getMutationErrorMessage(e) || 'Errore'); }
  };

  return (
    <FormDialog open={open} onClose={onClose} title={editItem ? 'Modifica Motrice' : 'Nuova Motrice'} onSubmit={handleSave} loading={creating || updating}>
      <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
        <div className="space-y-1.5"><Label>Targa *</Label><Input value={form.targa} onChange={e => setForm({ ...form, targa: e.target.value })} required /></div>
        <div className="space-y-1.5"><Label>Anno</Label><Input type="number" value={form.anno} onChange={e => setForm({ ...form, anno: Number(e.target.value) })} /></div>
        <div className="space-y-1.5"><Label>Marca</Label><Input value={form.marca} onChange={e => setForm({ ...form, marca: e.target.value })} /></div>
        <div className="space-y-1.5"><Label>Modello</Label><Input value={form.modello} onChange={e => setForm({ ...form, modello: e.target.value })} /></div>
        <div className="space-y-1.5"><Label>Portata (Kg)</Label><Input type="number" value={form.portata_kg} onChange={e => setForm({ ...form, portata_kg: Number(e.target.value) })} /></div>
        <div className="space-y-1.5 md:col-span-2"><Label>Note</Label><Input value={form.note} onChange={e => setForm({ ...form, note: e.target.value })} /></div>
      </div>
    </FormDialog>
  );
}

function SemirimorchioFormDialog({ open, onClose, editItem }: { open: boolean; onClose: () => void; editItem: DtoSemirimorchioResponse | null }) {
  const [form, setForm] = useState<DtoSemirimorchioRequest>(editItem ? {
    targa: editItem.targa || '', tipo: editItem.tipo || '', scompartature: editItem.scompartature || 1,
    portata_kg: editItem.portata_kg || 0, note: editItem.note || '',
  } : emptySemirimorchioForm);
  const [createSemirimorchio, { isLoading: creating }] = useCreateSemirimorchioMutation();
  const [updateSemirimorchio, { isLoading: updating }] = useUpdateSemirimorchioMutation();

  const handleSave = async () => {
    try {
      if (editItem?.id) { await updateSemirimorchio({ id: editItem.id, body: form }).unwrap(); toast.success('Semirimorchio aggiornato'); }
      else { await createSemirimorchio(form).unwrap(); toast.success('Semirimorchio creato'); }
      onClose();
    } catch (e) { toast.error(getMutationErrorMessage(e) || 'Errore'); }
  };

  return (
    <FormDialog open={open} onClose={onClose} title={editItem ? 'Modifica Semirimorchio' : 'Nuovo Semirimorchio'} onSubmit={handleSave} loading={creating || updating}>
      <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
        <div className="space-y-1.5"><Label>Targa *</Label><Input value={form.targa} onChange={e => setForm({ ...form, targa: e.target.value })} required /></div>
        <div className="space-y-1.5"><Label>Tipo</Label><Input value={form.tipo} onChange={e => setForm({ ...form, tipo: e.target.value })} placeholder="es. Frigo, Centinato" /></div>
        <div className="space-y-1.5"><Label>Scompartature</Label><Input type="number" value={form.scompartature} onChange={e => setForm({ ...form, scompartature: Number(e.target.value) })} /></div>
        <div className="space-y-1.5"><Label>Portata (Kg)</Label><Input type="number" value={form.portata_kg} onChange={e => setForm({ ...form, portata_kg: Number(e.target.value) })} /></div>
        <div className="space-y-1.5 md:col-span-2"><Label>Note</Label><Input value={form.note} onChange={e => setForm({ ...form, note: e.target.value })} /></div>
      </div>
    </FormDialog>
  );
}

export default function VehiclesPage() {
  const [search, setSearch] = useState('');
  const [tipoFilter, setTipoFilter] = useState<'' | MezzoTipo>('');
  const [motriceDialog, setMotriceDialog] = useState<{ open: boolean; item: DtoMotriceResponse | null }>({ open: false, item: null });
  const [semirimorchioDialog, setSemirimorchioDialog] = useState<{ open: boolean; item: DtoSemirimorchioResponse | null }>({ open: false, item: null });

  const { data: motrici = [], isLoading: loadingMotrici } = useGetMotriciQuery(search, { skip: tipoFilter === 'Semirimorchio' });
  const { data: semirimorchi = [], isLoading: loadingSemirimorchi } = useGetSemirimorchiQuery(search, { skip: tipoFilter === 'Motrice' });
  const [deleteMotrice] = useDeleteMotriceMutation();
  const [deleteSemirimorchio] = useDeleteSemirimorchioMutation();

  const loading = (tipoFilter !== 'Semirimorchio' && loadingMotrici) || (tipoFilter !== 'Motrice' && loadingSemirimorchi);

  const rows = useMemo<MezzoRow[]>(() => {
    const list: MezzoRow[] = [];
    if (tipoFilter !== 'Semirimorchio') list.push(...motrici.map(motriceToRow));
    if (tipoFilter !== 'Motrice') list.push(...semirimorchi.map(semirimorchioToRow));
    return list.sort((a, b) => a.targa.localeCompare(b.targa));
  }, [motrici, semirimorchi, tipoFilter]);

  const handleEdit = (row: MezzoRow) => {
    if (row.tipo === 'Motrice') setMotriceDialog({ open: true, item: motrici.find(m => m.id === row.id) || null });
    else setSemirimorchioDialog({ open: true, item: semirimorchi.find(s => s.id === row.id) || null });
  };

  const handleDelete = async (row: MezzoRow) => {
    if (!window.confirm(`Eliminare il mezzo ${row.targa}?`)) return;
    try {
      if (row.tipo === 'Motrice') { await deleteMotrice(row.id).unwrap(); }
      else { await deleteSemirimorchio(row.id).unwrap(); }
      toast.success('Mezzo eliminato');
    } catch (e) { toast.error(getMutationErrorMessage(e) || 'Errore'); }
  };

  const columns: DataTableColumn[] = [
    { key: 'targa', label: 'Targa' },
    { key: 'tipo', label: 'Tipo' },
    { key: 'descrizione', label: 'Descrizione' },
    { key: 'portata', label: 'Portata (Kg)', className: 'text-right' },
    { key: 'actions', label: '', className: 'w-20' },
  ];

  return (
    <div data-testid="vehicles-page">
      <DataTable
        columns={columns}
        data={rows}
        loading={loading}
        searchValue={search}
        onSearchChange={setSearch}
        testId="masterdata-table"
        filters={
          <Select value={tipoFilter || 'all'} onValueChange={(v) => setTipoFilter(v === 'all' ? '' : (v as MezzoTipo))}>
            <SelectTrigger className="h-9 w-[170px] text-sm" data-testid="vehicles-tipo-filter">
              <SelectValue placeholder="Tutti i mezzi" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">Tutti i mezzi</SelectItem>
              <SelectItem value="Motrice">Motrici</SelectItem>
              <SelectItem value="Semirimorchio">Semirimorchi</SelectItem>
            </SelectContent>
          </Select>
        }
        addSlot={
          <>
            <Button size="sm" variant="outline" className="text-xs gap-1.5" onClick={() => setMotriceDialog({ open: true, item: null })} data-testid="masterdata-new-motrice-button">
              <Plus className="h-3.5 w-3.5" /> Motrice
            </Button>
            <Button size="sm" className="text-xs gap-1.5" onClick={() => setSemirimorchioDialog({ open: true, item: null })} data-testid="masterdata-new-semirimorchio-button">
              <Plus className="h-3.5 w-3.5" /> Semirimorchio
            </Button>
          </>
        }
        renderRow={(item) => (
          <TableRow key={item.id} className="hover:bg-muted/60">
            <TableCell className="py-2 font-mono font-medium">{item.targa}</TableCell>
            <TableCell className="py-2"><Badge variant="outline" className={tipoBadgeClass[item.tipo]}>{item.tipo}</Badge></TableCell>
            <TableCell className="py-2">{item.descrizione}</TableCell>
            <TableCell className="py-2 text-right tabular-nums">{formatEuro(item.portataKg)}</TableCell>
            <TableCell className="py-2">
              <div className="flex gap-1">
                <Button variant="ghost" size="icon" className="h-7 w-7" onClick={() => handleEdit(item)}><Pencil className="h-3 w-3" /></Button>
                <Button variant="ghost" size="icon" className="h-7 w-7 text-destructive" onClick={() => handleDelete(item)}><Trash2 className="h-3 w-3" /></Button>
              </div>
            </TableCell>
          </TableRow>
        )}
      />

      {motriceDialog.open && (
        <MotriceFormDialog open={motriceDialog.open} onClose={() => setMotriceDialog({ open: false, item: null })} editItem={motriceDialog.item} />
      )}
      {semirimorchioDialog.open && (
        <SemirimorchioFormDialog open={semirimorchioDialog.open} onClose={() => setSemirimorchioDialog({ open: false, item: null })} editItem={semirimorchioDialog.item} />
      )}
    </div>
  );
}
