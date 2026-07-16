import { useState, useEffect, useCallback } from 'react';
import { getAdminUsers, updateAdminUser, getProfiles, register } from '@/lib/api';
import { Card } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Select, SelectTrigger, SelectValue, SelectContent, SelectItem } from '@/components/ui/select';
import { Switch } from '@/components/ui/switch';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog';
import { Badge } from '@/components/ui/badge';
import { Skeleton } from '@/components/ui/skeleton';
import { toast } from 'sonner';
import { logger } from '@/lib/logger';
import { Plus, Pencil } from 'lucide-react';

export default function UsersPage() {
  const [users, setUsers] = useState([]);
  const [profiles, setProfiles] = useState([]);
  const [profilesById, setProfilesById] = useState({});
  const [loading, setLoading] = useState(true);

  const [editOpen, setEditOpen] = useState(false);
  const [editForm, setEditForm] = useState({ id: '', name: '', profile_id: '', active: true });
  const [saving, setSaving] = useState(false);

  const [createOpen, setCreateOpen] = useState(false);
  const [createForm, setCreateForm] = useState({ email: '', name: '', password: '', profile_id: 'operatore', role: 'operatore' });

  const fetchAll = useCallback(() => {
    setLoading(true);
    Promise.all([getAdminUsers(), getProfiles()])
      .then(([u, p]) => {
        setUsers(u.data);
        setProfiles(p.data);
        const map = {};
        p.data.forEach(x => { map[x.id] = x; });
        setProfilesById(map);
      })
      .catch(err => logger.error('Errore caricamento utenti:', err))
      .finally(() => setLoading(false));
  }, []);

  useEffect(() => { fetchAll(); }, [fetchAll]);

  const openEdit = (u) => {
    setEditForm({
      id: u.id,
      name: u.name,
      profile_id: u.profile_id || '',
      active: u.active !== false,
    });
    setEditOpen(true);
  };

  const handleEditSave = async () => {
    setSaving(true);
    try {
      const payload = {};
      if (editForm.name) payload.name = editForm.name;
      if (editForm.profile_id) payload.profile_id = editForm.profile_id;
      payload.active = editForm.active;
      await updateAdminUser(editForm.id, payload);
      toast.success('Utente aggiornato');
      setEditOpen(false);
      fetchAll();
    } catch (e) {
      toast.error(e.response?.data?.detail || 'Errore salvataggio');
    } finally {
      setSaving(false);
    }
  };

  const handleCreate = async () => {
    if (!createForm.email || !createForm.name || createForm.password.length < 12) {
      toast.error('Compila email, nome e password (>=12 caratteri)');
      return;
    }
    setSaving(true);
    try {
      // Lo schema attuale di POST /auth/register usa `role`. Lo deriviamo
      // dal profilo selezionato (lookup inherits_role) e mandiamo entrambi.
      const prof = profilesById[createForm.profile_id];
      const role = prof?.inherits_role || 'operatore';
      await register({
        email: createForm.email,
        name: createForm.name,
        password: createForm.password,
        role,
        profile_id: createForm.profile_id,
      });
      toast.success('Utente creato');
      setCreateOpen(false);
      setCreateForm({ email: '', name: '', password: '', profile_id: 'operatore', role: 'operatore' });
      fetchAll();
    } catch (e) {
      toast.error(e.response?.data?.detail || 'Errore creazione utente');
    } finally {
      setSaving(false);
    }
  };

  return (
    <div className="space-y-3" data-testid="users-page">
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-xl md:text-2xl font-bold tracking-tight" style={{ fontFamily: "'Space Grotesk', sans-serif" }}>
            Utenti
          </h1>
          <p className="text-xs text-muted-foreground">Gestisci utenti e assegna profili RBAC.</p>
        </div>
        <Button size="sm" onClick={() => setCreateOpen(true)} className="text-xs gap-1.5">
          <Plus className="h-3.5 w-3.5" /> Nuovo utente
        </Button>
      </div>

      <Card className="rounded-xl border shadow-sm">
        <div className="overflow-x-auto">
          <Table className="text-xs md:text-sm">
            <TableHeader>
              <TableRow>
                <TableHead>Email</TableHead>
                <TableHead>Nome</TableHead>
                <TableHead>Profilo</TableHead>
                <TableHead>Stato</TableHead>
                <TableHead className="w-16">Azioni</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {loading ? (
                Array.from({ length: 4 }).map((_, i) => (
                  <TableRow key={`s${i}`}>{Array.from({length:5}).map((_, j) => <TableCell key={j} className="py-2"><Skeleton className="h-4 w-full" /></TableCell>)}</TableRow>
                ))
              ) : users.length === 0 ? (
                <TableRow><TableCell colSpan={5} className="text-center py-8 text-muted-foreground">Nessun utente</TableCell></TableRow>
              ) : users.map(u => (
                <TableRow key={u.id} className="hover:bg-muted/60">
                  <TableCell className="py-2 font-mono">{u.email}</TableCell>
                  <TableCell className="py-2">{u.name}</TableCell>
                  <TableCell className="py-2">
                    {u.profile_id ? (
                      <Badge variant="outline" className="text-[10px]">{profilesById[u.profile_id]?.nome || u.profile_id}</Badge>
                    ) : (
                      <Badge variant="secondary" className="text-[10px]">— ({u.role})</Badge>
                    )}
                  </TableCell>
                  <TableCell className="py-2">
                    {u.active ? <Badge variant="default" className="text-[10px]">Attivo</Badge> : <Badge variant="secondary" className="text-[10px]">Disattivato</Badge>}
                  </TableCell>
                  <TableCell className="py-2">
                    <Button variant="ghost" size="icon" className="h-7 w-7" onClick={() => openEdit(u)} title="Modifica">
                      <Pencil className="h-3 w-3" />
                    </Button>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
      </Card>

      {/* Edit dialog */}
      <Dialog open={editOpen} onOpenChange={setEditOpen}>
        <DialogContent className="max-w-md">
          <DialogHeader><DialogTitle>Modifica utente</DialogTitle></DialogHeader>
          <div className="space-y-3">
            <div>
              <Label>Nome</Label>
              <Input value={editForm.name} onChange={e => setEditForm({...editForm, name: e.target.value})} />
            </div>
            <div>
              <Label>Profilo</Label>
              <Select value={editForm.profile_id} onValueChange={v => setEditForm({...editForm, profile_id: v})}>
                <SelectTrigger><SelectValue placeholder="Seleziona profilo" /></SelectTrigger>
                <SelectContent>
                  {profiles.map(p => <SelectItem key={p.id} value={p.id}>{p.nome}</SelectItem>)}
                </SelectContent>
              </Select>
            </div>
            <div className="flex items-center gap-2">
              <Switch checked={editForm.active} onCheckedChange={v => setEditForm({...editForm, active: v})} />
              <Label>Attivo</Label>
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setEditOpen(false)}>Annulla</Button>
            <Button onClick={handleEditSave} disabled={saving}>{saving ? 'Salvataggio…' : 'Salva'}</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Create dialog */}
      <Dialog open={createOpen} onOpenChange={setCreateOpen}>
        <DialogContent className="max-w-md">
          <DialogHeader><DialogTitle>Nuovo utente</DialogTitle></DialogHeader>
          <div className="space-y-3">
            <div>
              <Label>Email *</Label>
              <Input type="email" value={createForm.email} onChange={e => setCreateForm({...createForm, email: e.target.value})} />
            </div>
            <div>
              <Label>Nome *</Label>
              <Input value={createForm.name} onChange={e => setCreateForm({...createForm, name: e.target.value})} />
            </div>
            <div>
              <Label>Password * (min 12 caratteri)</Label>
              <Input type="password" value={createForm.password} onChange={e => setCreateForm({...createForm, password: e.target.value})} />
            </div>
            <div>
              <Label>Profilo *</Label>
              <Select value={createForm.profile_id} onValueChange={v => setCreateForm({...createForm, profile_id: v})}>
                <SelectTrigger><SelectValue /></SelectTrigger>
                <SelectContent>
                  {profiles.map(p => <SelectItem key={p.id} value={p.id}>{p.nome}</SelectItem>)}
                </SelectContent>
              </Select>
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setCreateOpen(false)}>Annulla</Button>
            <Button onClick={handleCreate} disabled={saving}>{saving ? 'Creazione…' : 'Crea'}</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
