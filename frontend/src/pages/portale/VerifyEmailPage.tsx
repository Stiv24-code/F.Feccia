import { useEffect, useRef, useState } from 'react';
import { useNavigate, useSearchParams, Link } from 'react-router-dom';
import { useAuth } from '@/lib/auth-context';
import { getApiErrorMessage } from '@/lib/apiError';
import { Card, CardContent } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { CheckCircle2, XCircle, Loader2 } from 'lucide-react';

// Destinazione del link mandato da AuthService.sendVerificationMail
// (backend/internal/services/auth/auth.go) — legge ?token=, conferma via
// POST /auth/verify-email, poi apre la sessione come un login.
export default function VerifyEmailPage() {
  const [params] = useSearchParams();
  const { verifyEmail } = useAuth();
  const navigate = useNavigate();
  const [status, setStatus] = useState<'loading' | 'success' | 'error'>('loading');
  const [error, setError] = useState('');
  const ranOnce = useRef(false);

  useEffect(() => {
    if (ranOnce.current) return;
    ranOnce.current = true;
    const token = params.get('token');
    if (!token) {
      setStatus('error');
      setError('Link non valido: token mancante.');
      return;
    }
    verifyEmail(token)
      .then(() => {
        setStatus('success');
        setTimeout(() => navigate('/portale'), 1500);
      })
      .catch((err) => {
        setStatus('error');
        setError(getApiErrorMessage(err) || 'Link di conferma non valido o scaduto.');
      });
  }, [params, verifyEmail, navigate]);

  return (
    <div className="min-h-screen flex items-center justify-center p-6 bg-background">
      <Card className="w-full max-w-md border shadow-sm">
        <CardContent className="pt-8 pb-8 text-center space-y-3">
          {status === 'loading' && (
            <>
              <Loader2 className="h-10 w-10 mx-auto animate-spin text-primary" />
              <p className="text-sm text-muted-foreground">Confermo il tuo account...</p>
            </>
          )}
          {status === 'success' && (
            <>
              <CheckCircle2 className="h-10 w-10 mx-auto text-primary" />
              <p className="text-lg font-semibold" style={{ fontFamily: "'Space Grotesk', sans-serif" }}>Account confermato</p>
              <p className="text-sm text-muted-foreground">Ti stiamo portando al portale...</p>
            </>
          )}
          {status === 'error' && (
            <>
              <XCircle className="h-10 w-10 mx-auto text-destructive" />
              <p className="text-lg font-semibold" style={{ fontFamily: "'Space Grotesk', sans-serif" }}>Verifica non riuscita</p>
              <p className="text-sm text-muted-foreground">{error}</p>
              <div className="flex justify-center gap-2 pt-2">
                <Button variant="outline" asChild><Link to="/registrati">Registrati di nuovo</Link></Button>
                <Button asChild><Link to="/login">Vai al login</Link></Button>
              </div>
            </>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
