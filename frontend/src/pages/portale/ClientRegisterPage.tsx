import { useState, type FormEvent } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { useAuth } from '@/lib/auth-context';
import { getApiErrorMessage } from '@/lib/apiError';
import type { DtoClientRegisterRequest, DtoGeocodeResultDTO } from '@/api/data-contracts';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import PasswordInput from '@/components/shared/PasswordInput';
import { Label } from '@/components/ui/label';
import { Button } from '@/components/ui/button';
import { MapPicker } from '@/components/shared/MapPicker';
import { AddressSearchInput } from '@/components/shared/AddressSearchInput';
import { toast } from 'sonner';
import { Truck, UserPlus, Loader2, MailCheck } from 'lucide-react';

type RegisterForm = Omit<DtoClientRegisterRequest, 'lat' | 'lng'> & { lat: number | null; lng: number | null };

const emptyForm: RegisterForm = {
  ragione_sociale: '', indirizzo: '', citta: '', cap: '', provincia: '', lat: null, lng: null,
  partita_iva: '', codice_fiscale: '', telefono: '', name: '', email: '', password: '',
};

export default function ClientRegisterPage() {
  const { registerClient } = useAuth();
  const navigate = useNavigate();
  const [form, setForm] = useState<RegisterForm>(emptyForm);
  const [flySignal, setFlySignal] = useState(0);
  const [loading, setLoading] = useState(false);
  // Non-null quando il backend ha mandato un'email di conferma invece di
  // loggare subito (SMTP configurato) — sostituisce il form con un messaggio
  // "controlla la posta" invece di navigare nel portale.
  const [pendingMessage, setPendingMessage] = useState<string | null>(null);

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setLoading(true);
    try {
      const body: DtoClientRegisterRequest = { ...form, lat: form.lat ?? undefined, lng: form.lng ?? undefined };
      const result = await registerClient(body);
      if (result.verified === false) {
        setPendingMessage(result.message || 'Controlla la tua email per confermare l\'account.');
        return;
      }
      toast.success('Registrazione completata');
      navigate('/portale');
    } catch (err) {
      toast.error(getApiErrorMessage(err) || 'Errore durante la registrazione');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen flex">
      <div className="hidden lg:flex lg:w-1/2 flex-col justify-between p-12" style={{ background: 'var(--sidebar-bg)' }}>
        <div>
          <div className="flex items-center gap-3 mb-16">
            <div className="w-10 h-10 rounded-xl flex items-center justify-center" style={{ background: 'var(--sidebar-accent)' }}>
              <Truck className="h-5 w-5 text-white" />
            </div>
            <span className="text-xl font-bold tracking-tight" style={{ color: 'var(--sidebar-text)', fontFamily: "'Space Grotesk', sans-serif" }}>
              TMS <span className="font-normal" style={{ color: 'var(--sidebar-muted)' }}>· F.lli Feccia</span>
            </span>
          </div>
          <h2 className="text-4xl font-bold leading-tight mb-4" style={{ color: 'var(--sidebar-text)', fontFamily: "'Space Grotesk', sans-serif" }}>
            Portale Cliente
          </h2>
          <p className="text-base leading-relaxed max-w-md" style={{ color: 'var(--sidebar-muted)' }}>
            Registrati per creare i tuoi ordini di trasporto e seguirne lo stato — FECCIA F.lli si occupa della pianificazione.
          </p>
        </div>
      </div>

      <div className="flex-1 flex items-center justify-center p-6 bg-background">
        <div className="w-full max-w-lg">
          <div className="lg:hidden flex items-center gap-3 mb-8">
            <div className="w-9 h-9 rounded-xl flex items-center justify-center" style={{ background: 'var(--sidebar-accent)' }}>
              <Truck className="h-4 w-4 text-white" />
            </div>
            <span className="text-lg font-bold" style={{ fontFamily: "'Space Grotesk', sans-serif" }}>TMS · F.lli Feccia</span>
          </div>

          <Card className="border shadow-sm">
            {pendingMessage ? (
              <CardContent className="pt-6 pb-8 text-center space-y-3">
                <MailCheck className="h-10 w-10 mx-auto text-primary" />
                <p className="text-lg font-semibold" style={{ fontFamily: "'Space Grotesk', sans-serif" }}>Controlla la tua email</p>
                <p className="text-sm text-muted-foreground">{pendingMessage}</p>
                <p className="text-sm text-center text-muted-foreground pt-2">
                  Hai già confermato? <Link to="/login" className="underline hover:text-primary">Accedi</Link>
                </p>
              </CardContent>
            ) : (
            <>
            <CardHeader className="pb-4">
              <CardTitle className="text-xl" style={{ fontFamily: "'Space Grotesk', sans-serif" }}>Registrati come cliente</CardTitle>
              <CardDescription>Crea il tuo account per inviare ordini di trasporto</CardDescription>
            </CardHeader>
            <CardContent>
              <form onSubmit={handleSubmit} className="space-y-4">
                <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
                  <div className="md:col-span-2 space-y-1.5">
                    <Label htmlFor="ragione_sociale">Ragione Sociale *</Label>
                    <Input id="ragione_sociale" value={form.ragione_sociale} onChange={e => setForm({ ...form, ragione_sociale: e.target.value })} required />
                  </div>
                  <div className="md:col-span-2 space-y-1.5">
                    <Label htmlFor="indirizzo">Indirizzo</Label>
                    <AddressSearchInput
                      value={form.indirizzo || ''}
                      onChange={(v) => setForm(f => ({ ...f, indirizzo: v }))}
                      onSelect={(r: DtoGeocodeResultDTO) => {
                        setForm(f => ({ ...f, indirizzo: r.indirizzo || f.indirizzo, citta: r.citta || f.citta, cap: r.cap || f.cap, provincia: r.provincia || f.provincia, lat: r.lat ?? f.lat, lng: r.lng ?? f.lng }));
                        setFlySignal(s => s + 1);
                      }}
                    />
                  </div>
                  <div className="space-y-1.5"><Label htmlFor="citta">Città</Label><Input id="citta" value={form.citta} onChange={e => setForm({ ...form, citta: e.target.value })} /></div>
                  <div className="space-y-1.5"><Label htmlFor="cap">CAP</Label><Input id="cap" value={form.cap} onChange={e => setForm({ ...form, cap: e.target.value })} /></div>
                  <div className="space-y-1.5"><Label htmlFor="provincia">Provincia</Label><Input id="provincia" value={form.provincia} onChange={e => setForm({ ...form, provincia: e.target.value })} /></div>
                  <div className="md:col-span-2 space-y-1.5">
                    <Label>Posizione</Label>
                    <MapPicker lat={form.lat} lng={form.lng} flyToSignal={flySignal} />
                  </div>
                  <div className="space-y-1.5"><Label htmlFor="partita_iva">Partita IVA</Label><Input id="partita_iva" value={form.partita_iva} onChange={e => setForm({ ...form, partita_iva: e.target.value })} /></div>
                  <div className="space-y-1.5"><Label htmlFor="codice_fiscale">Codice Fiscale</Label><Input id="codice_fiscale" value={form.codice_fiscale} onChange={e => setForm({ ...form, codice_fiscale: e.target.value })} /></div>
                  <div className="md:col-span-2 space-y-1.5"><Label htmlFor="telefono">Telefono</Label><Input id="telefono" value={form.telefono} onChange={e => setForm({ ...form, telefono: e.target.value })} /></div>
                </div>

                <div className="border-t pt-4 grid grid-cols-1 md:grid-cols-2 gap-3">
                  <div className="md:col-span-2 space-y-1.5">
                    <Label htmlFor="name">Referente *</Label>
                    <Input id="name" value={form.name} onChange={e => setForm({ ...form, name: e.target.value })} placeholder="Nome e cognome" required />
                  </div>
                  <div className="md:col-span-2 space-y-1.5">
                    <Label htmlFor="email">Email *</Label>
                    <Input id="email" type="email" value={form.email} onChange={e => setForm({ ...form, email: e.target.value })} autoComplete="email" required />
                  </div>
                  <div className="md:col-span-2 space-y-1.5">
                    <Label htmlFor="password">Password *</Label>
                    <PasswordInput id="password" value={form.password} onChange={e => setForm({ ...form, password: e.target.value })} autoComplete="new-password" minLength={12} placeholder="Almeno 12 caratteri" required />
                  </div>
                </div>

                <Button type="submit" className="w-full mt-2" disabled={loading} data-testid="client-register-submit">
                  {loading ? <Loader2 className="h-4 w-4 animate-spin mr-2" /> : <UserPlus className="h-4 w-4 mr-2" />}
                  Registrati
                </Button>
                <p className="text-sm text-center text-muted-foreground">
                  Hai già un account? <Link to="/login" className="underline hover:text-primary">Accedi</Link>
                </p>
              </form>
            </CardContent>
            </>
            )}
          </Card>
        </div>
      </div>
    </div>
  );
}
