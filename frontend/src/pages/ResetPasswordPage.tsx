import { useState, type FormEvent } from 'react';
import { useNavigate, useSearchParams, Link } from 'react-router-dom';
import { resetPassword } from '@/lib/api';
import { getApiErrorMessage } from '@/lib/apiError';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import PasswordInput from '@/components/shared/PasswordInput';
import { Label } from '@/components/ui/label';
import { Button } from '@/components/ui/button';
import { toast } from 'sonner';
import { KeyRound, Loader2, XCircle } from 'lucide-react';

// Destinazione del link mandato da AuthService.sendPasswordResetMail
// (backend/internal/services/auth/auth.go) — legge ?token=, poi POST
// /auth/reset-password. A differenza di VerifyEmailPage il successo non
// apre una sessione: si torna al login e si accede con la password nuova.
export default function ResetPasswordPage() {
  const [params] = useSearchParams();
  const navigate = useNavigate();
  const token = params.get('token');
  const [password, setPassword] = useState('');
  const [confirm, setConfirm] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    if (password !== confirm) {
      setError('Le due password non coincidono.');
      return;
    }
    if (!token) return;
    setLoading(true);
    setError('');
    try {
      await resetPassword(token, password);
      toast.success('Password aggiornata: accedi con la nuova password');
      navigate('/login');
    } catch (err) {
      setError(getApiErrorMessage(err) || 'Link di reset non valido o scaduto.');
    } finally {
      setLoading(false);
    }
  };

  if (!token) {
    return (
      <div className="min-h-screen flex items-center justify-center p-6 bg-background">
        <Card className="w-full max-w-md border shadow-sm">
          <CardContent className="pt-8 pb-8 text-center space-y-3">
            <XCircle className="h-10 w-10 mx-auto text-destructive" />
            <p className="text-lg font-semibold" style={{ fontFamily: "'Space Grotesk', sans-serif" }}>Link non valido</p>
            <p className="text-sm text-muted-foreground">Il link non contiene un token di reset.</p>
            <Button asChild><Link to="/password-dimenticata">Richiedi un nuovo link</Link></Button>
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className="min-h-screen flex items-center justify-center p-6 bg-background">
      <div className="w-full max-w-sm">
        <Card className="border shadow-sm">
          <CardHeader className="pb-4">
            <CardTitle className="text-xl" style={{ fontFamily: "'Space Grotesk', sans-serif" }}>Reimposta password</CardTitle>
            <CardDescription>Scegli una nuova password per il tuo account</CardDescription>
          </CardHeader>
          <CardContent>
            <form onSubmit={handleSubmit} className="flex flex-col gap-4">
              <div className="space-y-1.5">
                <Label htmlFor="password">Nuova password</Label>
                <PasswordInput
                  id="password"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  placeholder="Almeno 12 caratteri"
                  autoComplete="new-password"
                  minLength={12}
                  required
                />
              </div>
              <div className="space-y-1.5">
                <Label htmlFor="confirm">Conferma password</Label>
                <PasswordInput
                  id="confirm"
                  value={confirm}
                  onChange={(e) => setConfirm(e.target.value)}
                  placeholder="Ripeti la password"
                  autoComplete="new-password"
                  minLength={12}
                  required
                />
              </div>
              {error && <p className="text-sm text-destructive">{error}</p>}
              <Button type="submit" className="w-full mt-2" disabled={loading}>
                {loading ? <Loader2 className="h-4 w-4 animate-spin mr-2" /> : <KeyRound className="h-4 w-4 mr-2" />}
                Reimposta password
              </Button>
            </form>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
