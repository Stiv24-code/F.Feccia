import { useState, useEffect, useCallback } from 'react';
import { getProfiles, createProfile, updateProfile, deleteProfile } from '@/lib/api';
import { Card } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import { Select, SelectTrigger, SelectValue, SelectContent, SelectItem } from '@/components/ui/select';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog';
import { Badge } from '@/components/ui/badge';
import { Skeleton } from '@/components/ui/skeleton';
import { toast } from 'sonner';
import { logger } from '@/lib/logger';
import { Plus, Pencil, Trash2, Lock } from 'lucide-react';

const ROLE_LABEL = {
  admin: 'Amministratore',
  amministrazione: 'Amministrazione',
  planner: 'Planner',
  operatore: 'Operatore',
};

const emptyForm = { nome: '', descrizione: '', inherits_role: 'operatore' };

export default function ProfilesPage() {
  const [profiles, setProfiles] = useState([]);
  const [loading, setLoading] = useState(true);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [saving, setSaving] = useState(false);
  const [editId, setEditId] = useState(null);
  const [editIsSystem, setEditIsSystem] = useState(false);
  const [form, setForm] = useState(emptyForm);

  const fetchProfiles = useCallback(() => {
    setLoading(true);
    getProfiles()
      .then(r => setProfiles(r.data))
      .catch(err => logger.error('Errore caricamento profili:', err))
      .finally(() => setLoading(false));
  }, []);

  useEffect(() => { fetchProfiles(); }, [fetchProfiles]);

  const openNew = () => {
    setForm(emptyForm);
    setEditId(null);
    setEditIsSystem(false);
    setDialogOpen(true);
  };

  const openEdit = (p) => {
    setForm({ nome: p.nome, descrizione: p.descrizione || '', inherits_role: p.inherits_role });
    setEditId(p.id);
    setEditIsSystem(!!p.system);
    setDialogOpen(true);
  };

  const handleSave = async () => {
    if (!form.nome.trim()) { toast.error('Nome obbligatorio'); return; }
    setSaving(true);
    try {
      if (editId) {
        const payload = { nome: form.nome, descrizione: form.descrizione };
        if (!editIsSystem) payload.inherits_role = form.inherits_role;
        await updateProfile(editId, payload);
        toast.success('Profilo aggiornato');
      } else {
        await createProfile(form);
        toast.success('Profilo creato');
      }
      setDialogOpen(false);
      fetchProfiles();
    } catch (e) {
      toast.error(e.response?.data?.detail || 'Errore salvataggio');
    } finally {
      setSaving(false);
    }
  };

  const handleDelete = async (p) => {
    if (!window.confirm(`Eliminare il profilo "${p.nome}"?`)) return;
    try {
      await deleteProfile(p.id);
      toast.success('Eliminato');
      fetchProfiles();
    } catch (e) {
      toast.error(e.response?.data?.detail || 'Errore eliminazione');
    }
  };

  return (
    <div className="space-y-3" data-testid="profiles-page">
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-xl md:text-2xl font-bold tracking-tight" style={{ fontFamily: "'Space Grotesk', sans-serif" }}>
            Profili RBAC
          </h1>
          <p className="text-xs text-muted-foreground">Crea profili custom che ereditano da uno dei 4 ruoli base.</p>
        </div>
        <Button size="sm" onClick={openNew} className="text-xs gap-1.5">
          <Plus className="h-3.5 w-3.5" /> Nuovo profilo
        </Button>
      </div>

      <Card className="rounded-xl border shadow-sm">
        <div className="overflow-x-auto">
          <Table className="text-xs md:text-sm">
            <TableHeader>
              <TableRow>
                <TableHead>Nome</TableHead>
                <TableHead>Descrizione</TableHead>
                <TableHead>Eredita ruolo</TableHead>
                <TableHead>Tipo</TableHead>
                <TableHead className="w-24">Azioni</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {loading ? (
                Array.from({ length: 4 }).map((_, i) => (
                  <TableRow key={`skel-${i}`}>{Array.from({length:5}).map((_, j) => <TableCell key={j} className="py-2"><Skeleton className="h-4 w-full" /></TableCell>)}</TableRow>
                ))
              ) : profiles.length === 0 ? (
                <TableRow><TableCell colSpan={5} className="text-center py-8 text-muted-foreground">Nessun profilo</TableCell></TableRow>
              ) : profiles.map(p => (
                <TableRow key={p.id} className="hover:bg-muted/60">
                  <TableCell className="py-2 font-medium">{p.nome}</TableCell>
                  <TableCell className="py-2 max-w-[400px] truncate">{p.descrizione || '—'}</TableCell>
                  <TableCell className="py-2"><Badge variant="outline" className="text-[10px]">{ROLE_LABEL[p.inherits_role] || p.inherits_role}</Badge></TableCell>
                  <TableCell className="py-2">
                    {p.system ? (
                      <Badge variant="secondary" className="text-[10px] gap-1"><Lock className="h-3 w-3" /> Sistema</Badge>
                    ) : (
                      <Badge variant="default" className="text-[10px]">Custom</Badge>
                    )}
                  </TableCell>
                  <TableCell className="py-2">
                    <div className="flex gap-1">
                      <Button variant="ghost" size="icon" className="h-7 w-7" onClick={() => openEdit(p)} title="Modifica">
                        <Pencil className="h-3 w-3" />
                      </Button>
                      {!p.system && (
                        <Button variant="ghost" size="icon" className="h-7 w-7 text-destructive" onClick={() => handleDelete(p)} title="Elimina">
                          <Trash2 className="h-3 w-3" />
                        </Button>
                      )}
                    </div>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
      </Card>

      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <DialogContent className="max-w-md">
          <DialogHeader>
            <DialogTitle>{editId ? 'Modifica profilo' : 'Nuovo profilo'}</DialogTitle>
          </DialogHeader>
          <div className="space-y-3">
            <div>
              <Label>Nome *</Label>
              <Input value={form.nome} onChange={e => setForm({ ...form, nome: e.target.value })} maxLength={80} />
            </div>
            <div>
              <Label>Descrizione</Label>
              <Textarea value={form.descrizione} onChange={e => setForm({ ...form, descrizione: e.target.value })} rows={3} />
            </div>
            <div>
              <Label>Eredita ruolo *</Label>
              <Select
                value={form.inherits_role}
                onValueChange={v => setForm({ ...form, inherits_role: v })}
                disabled={editIsSystem}
              >
                <SelectTrigger><SelectValue /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="admin">Amministratore</SelectItem>
                  <SelectItem value="amministrazione">Amministrazione</SelectItem>
                  <SelectItem value="planner">Planner</SelectItem>
                  <SelectItem value="operatore">Operatore</SelectItem>
                </SelectContent>
              </Select>
              {editIsSystem && (
                <p className="text-[11px] text-muted-foreground mt-1">
                  Profilo di sistema: il ruolo ereditato non è modificabile.
                </p>
              )}
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDialogOpen(false)}>Annulla</Button>
            <Button onClick={handleSave} disabled={saving}>{saving ? 'Salvataggio…' : 'Salva'}</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
