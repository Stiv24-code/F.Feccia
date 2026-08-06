import { useState } from 'react';
import {
  useGetAdminUsersQuery,
  useCreateAdminUserMutation,
  useUpdateAdminUserMutation,
} from '@/store/api/appApi';
import type { DtoAuthUserResponse } from '@/api/data-contracts';
import { getMutationErrorMessage } from '@/store/api/rtkQueryHelpers';
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
import { Plus, Pencil } from 'lucide-react';

type Role = 'admin' | 'amministrazione' | 'planner' | 'operatore';

const ROLE_LABEL: Record<Role, string> = {
  admin: 'Amministratore',
  amministrazione: 'Amministrazione',
  planner: 'Planner',
  operatore: 'Operatore',
};

const ROLES = Object.keys(ROLE_LABEL) as Role[];

type EditForm = { id: number; name: string; role: Role; active: boolean };
type CreateForm = { email: string; name: string; password: string; role: Role };

const emptyEditForm: EditForm = { id: 0, name: '', role: 'operatore', active: true };
const emptyCreateForm: CreateForm = { email: '', name: '', password: '', role: 'operatore' };

export default function UsersPage() {
  const { data: users = [], isLoading: loading } = useGetAdminUsersQuery();
  const [createAdminUser, { isLoading: creating }] = useCreateAdminUserMutation();
  const [updateAdminUser, { isLoading: updating }] = useUpdateAdminUserMutation();
  const saving = creating || updating;

  const [editOpen, setEditOpen] = useState(false);
  const [editForm, setEditForm] = useState<EditForm>(emptyEditForm);

  const [createOpen, setCreateOpen] = useState(false);
  const [createForm, setCreateForm] = useState<CreateForm>(emptyCreateForm);

  const openEdit = (u: DtoAuthUserResponse) => {
    setEditForm({
      id: u.id!,
      name: u.name || '',
      role: (u.role as Role) || 'operatore',
      active: u.active !== false,
    });
    setEditOpen(true);
  };

  const handleEditSave = async () => {
    const original = users.find((u) => u.id === editForm.id);
    if (!original?.email) return;
    try {
      await updateAdminUser({
        id: editForm.id,
        login: original.email,
        name: editForm.name,
        role: editForm.role,
        active: editForm.active,
      }).unwrap();
      toast.success('Utente aggiornato');
      setEditOpen(false);
    } catch (e) {
      toast.error(getMutationErrorMessage(e) || 'Errore salvataggio');
    }
  };

  const handleCreate = async () => {
    if (!createForm.email || !createForm.name || createForm.password.length < 12) {
      toast.error('Compila email, nome e password (>=12 caratteri)');
      return;
    }
    try {
      await createAdminUser(createForm).unwrap();
      toast.success('Utente creato');
      setCreateOpen(false);
      setCreateForm(emptyCreateForm);
    } catch (e) {
      toast.error(getMutationErrorMessage(e) || 'Errore creazione utente');
    }
  };

  return (
    <div className="space-y-3" data-testid="users-page">
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-xl md:text-2xl font-bold tracking-tight" style={{ fontFamily: "'Space Grotesk', sans-serif" }}>
            Utenti
          </h1>
          <p className="text-xs text-muted-foreground">Gestisci utenti e il loro ruolo.</p>
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
                <TableHead>Ruolo</TableHead>
                <TableHead>Stato</TableHead>
                <TableHead className="w-16">Azioni</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {loading ? (
                Array.from({ length: 4 }).map((_, i) => (
                  <TableRow key={`s${i}`}>{Array.from({ length: 5 }).map((_, j) => <TableCell key={j} className="py-2"><Skeleton className="h-4 w-full" /></TableCell>)}</TableRow>
                ))
              ) : users.length === 0 ? (
                <TableRow><TableCell colSpan={5} className="text-center py-8 text-muted-foreground">Nessun utente</TableCell></TableRow>
              ) : users.map((u) => (
                <TableRow key={u.id} className="hover:bg-muted/60">
                  <TableCell className="py-2 font-mono">{u.email}</TableCell>
                  <TableCell className="py-2">{u.name}</TableCell>
                  <TableCell className="py-2">
                    <Badge variant="outline" className="text-[10px]">{ROLE_LABEL[u.role as Role] || u.role}</Badge>
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
              <Input value={editForm.name} onChange={(e) => setEditForm({ ...editForm, name: e.target.value })} />
            </div>
            <div>
              <Label>Ruolo</Label>
              <Select value={editForm.role} onValueChange={(v) => setEditForm({ ...editForm, role: v as Role })}>
                <SelectTrigger><SelectValue /></SelectTrigger>
                <SelectContent>
                  {ROLES.map((r) => <SelectItem key={r} value={r}>{ROLE_LABEL[r]}</SelectItem>)}
                </SelectContent>
              </Select>
            </div>
            <div className="flex items-center gap-2">
              <Switch checked={editForm.active} onCheckedChange={(v) => setEditForm({ ...editForm, active: v })} />
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
              <Input type="email" value={createForm.email} onChange={(e) => setCreateForm({ ...createForm, email: e.target.value })} />
            </div>
            <div>
              <Label>Nome *</Label>
              <Input value={createForm.name} onChange={(e) => setCreateForm({ ...createForm, name: e.target.value })} />
            </div>
            <div>
              <Label>Password * (min 12 caratteri)</Label>
              <Input type="password" value={createForm.password} onChange={(e) => setCreateForm({ ...createForm, password: e.target.value })} />
            </div>
            <div>
              <Label>Ruolo *</Label>
              <Select value={createForm.role} onValueChange={(v) => setCreateForm({ ...createForm, role: v as Role })}>
                <SelectTrigger><SelectValue /></SelectTrigger>
                <SelectContent>
                  {ROLES.map((r) => <SelectItem key={r} value={r}>{ROLE_LABEL[r]}</SelectItem>)}
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
