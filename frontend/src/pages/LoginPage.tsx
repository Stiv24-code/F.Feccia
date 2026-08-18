import { useState, type FormEvent } from 'react';
import { Link } from 'react-router-dom';
import { useAuth } from '@/lib/auth-context';
import { getApiErrorMessage, getApiErrorStatus } from '@/lib/apiError';
import { resendVerification } from '@/lib/api';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import PasswordInput from '@/components/shared/PasswordInput';
import { Label } from '@/components/ui/label';
import { Button } from '@/components/ui/button';
import { toast } from 'sonner';
import { Truck, LogIn, Loader2 } from 'lucide-react';

export default function LoginPage() {
  const { login } = useAuth();
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [loading, setLoading] = useState(false);
  // Non-null quando l'ultimo tentativo di login è stato respinto perché
  // l'account cliente non ha ancora confermato l'email — mostra un link per
  // richiedere un nuovo link di conferma senza ripresentare la registrazione.
  const [unverifiedEmail, setUnverifiedEmail] = useState<string | null>(null);
  const [resending, setResending] = useState(false);

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setUnverifiedEmail(null);
    try {
      await login(email, password);
      toast.success('Accesso effettuato');
    } catch (err) {
      const message = getApiErrorMessage(err);
      if (getApiErrorStatus(err) === 403 && message?.includes('non confermata')) {
        setUnverifiedEmail(email);
      }
      toast.error(message || 'Credenziali non valide');
    } finally {
      setLoading(false);
    }
  };

  const handleResend = async () => {
    if (!unverifiedEmail) return;
    setResending(true);
    try {
      const res = await resendVerification(unverifiedEmail);
      toast.success(res.data.message || 'Email di conferma inviata di nuovo');
    } catch (err) {
      toast.error(getApiErrorMessage(err) || 'Errore durante l\'invio');
    } finally {
      setResending(false);
    }
  };

  return (
    <div className="min-h-screen flex">
      {/* Left side - Brand */}
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
            Transport Management<br />System
          </h2>
          <p className="text-base leading-relaxed max-w-md" style={{ color: 'var(--sidebar-muted)' }}>
            Gestione trasporti e pianificazione operativa per FECCIA F.lli. Ordini, pianificazione, viaggi e fatturazione in un unico sistema.
          </p>
        </div>
      </div>

      {/* Right side - Form */}
      <div className="flex-1 flex items-center justify-center p-6 bg-background">
        <div className="w-full max-w-sm">
          <div className="lg:hidden flex items-center gap-3 mb-8">
            <div className="w-9 h-9 rounded-xl flex items-center justify-center" style={{ background: 'var(--sidebar-accent)' }}>
              <Truck className="h-4 w-4 text-white" />
            </div>
            <span className="text-lg font-bold" style={{ fontFamily: "'Space Grotesk', sans-serif" }}>TMS · F.lli Feccia</span>
          </div>

          <Card className="border shadow-sm">
            <CardHeader className="pb-4">
              <CardTitle className="text-xl" style={{ fontFamily: "'Space Grotesk', sans-serif" }}>Accedi</CardTitle>
              <CardDescription>Gestione trasporti e pianificazione operativa</CardDescription>
            </CardHeader>
            <CardContent>
              <form onSubmit={handleSubmit} className="flex flex-col gap-4">
                <div className="space-y-1.5">
                  <Label htmlFor="email">Email</Label>
                  <Input
                    id="email"
                    data-testid="login-email-input"
                    type="email"
                    value={email}
                    onChange={(e) => setEmail(e.target.value)}
                    placeholder="nome@azienda.it"
                    autoComplete="email"
                    required
                  />
                </div>
                <div className="space-y-1.5">
                  <Label htmlFor="password">Password</Label>
                  <PasswordInput
                    id="password"
                    data-testid="login-password-input"
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                    placeholder="Password"
                    autoComplete="current-password"
                    required
                  />
                </div>
                <Button type="submit" data-testid="login-submit-button" className="w-full mt-2" disabled={loading}>
                  {loading ? <Loader2 className="h-4 w-4 animate-spin mr-2" /> : <LogIn className="h-4 w-4 mr-2" />}
                  Entra
                </Button>
                {unverifiedEmail && (
                  <p className="text-sm text-center text-muted-foreground">
                    Non hai confermato l&apos;email?{' '}
                    <button type="button" onClick={handleResend} disabled={resending} className="underline hover:text-primary disabled:opacity-50">
                      {resending ? 'Invio…' : 'Invia di nuovo il link di conferma'}
                    </button>
                  </p>
                )}
                <p className="text-sm text-center text-muted-foreground">
                  Sei un cliente? <Link to="/registrati" className="underline hover:text-primary">Registrati</Link>
                </p>
              </form>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  );
}
