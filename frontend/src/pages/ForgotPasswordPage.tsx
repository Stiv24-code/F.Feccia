import { useState, type FormEvent } from 'react';
import { Link } from 'react-router-dom';
import { forgotPassword } from '@/lib/api';
import { getApiErrorMessage } from '@/lib/apiError';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Button } from '@/components/ui/button';
import { Mail, Loader2, CheckCircle2 } from 'lucide-react';

// Punto d'ingresso per "password dimenticata" — POST /auth/forgot-password
// risponde sempre con lo stesso messaggio generico (non rivela se l'email
// esiste), quindi qui non c'è nessun ramo di errore da gestire sul "non
// trovato": solo un eventuale errore di rete/validazione del form.
export default function ForgotPasswordPage() {
  const [email, setEmail] = useState('');
  const [loading, setLoading] = useState(false);
  const [sent, setSent] = useState(false);
  const [error, setError] = useState('');

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError('');
    try {
      await forgotPassword(email);
      setSent(true);
    } catch (err) {
      setError(getApiErrorMessage(err) || 'Errore durante l\'invio della richiesta.');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center p-6 bg-background">
      <div className="w-full max-w-sm">
        <Card className="border shadow-sm">
          <CardHeader className="pb-4">
            <CardTitle className="text-xl" style={{ fontFamily: "'Space Grotesk', sans-serif" }}>Password dimenticata</CardTitle>
            <CardDescription>Ti mandiamo un link per reimpostarla</CardDescription>
          </CardHeader>
          <CardContent>
            {sent ? (
              <div className="text-center space-y-3 py-2">
                <CheckCircle2 className="h-10 w-10 mx-auto text-primary" />
                <p className="text-sm text-muted-foreground">
                  Se l&apos;indirizzo è registrato, riceverai un&apos;email con le istruzioni per reimpostare la password.
                </p>
                <Button variant="outline" asChild className="mt-2"><Link to="/login">Torna al login</Link></Button>
              </div>
            ) : (
              <form onSubmit={handleSubmit} className="flex flex-col gap-4">
                <div className="space-y-1.5">
                  <Label htmlFor="email">Email</Label>
                  <Input
                    id="email"
                    type="email"
                    value={email}
                    onChange={(e) => setEmail(e.target.value)}
                    placeholder="nome@azienda.it"
                    autoComplete="email"
                    required
                  />
                </div>
                {error && <p className="text-sm text-destructive">{error}</p>}
                <Button type="submit" className="w-full mt-2" disabled={loading}>
                  {loading ? <Loader2 className="h-4 w-4 animate-spin mr-2" /> : <Mail className="h-4 w-4 mr-2" />}
                  Invia link di reset
                </Button>
                <p className="text-sm text-center text-muted-foreground">
                  <Link to="/login" className="underline hover:text-primary">Torna al login</Link>
                </p>
              </form>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
