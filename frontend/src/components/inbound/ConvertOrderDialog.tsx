import { useEffect, useMemo, useState } from 'react';
import { Link } from 'react-router-dom';
import { toast } from 'sonner';
import { ArrowRightLeft, Info, TriangleAlert } from 'lucide-react';
import { apiClient } from '@/lib/apiClient';
import { useGetCustomersQuery } from '@/store/api/appApi';
import type {
  DtoInboundOrderConvertRequest,
  DtoInboundOrderResponse,
} from '@/api/data-contracts';
import {
  Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle,
} from '@/components/ui/dialog';
import {
  Select, SelectContent, SelectItem, SelectTrigger, SelectValue,
} from '@/components/ui/select';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import { getInboundApiError, fmtKg } from './constants';

type Props = {
  order: DtoInboundOrderResponse | null;
  onClose: () => void;
  onConverted: () => void | Promise<void>;
};

/**
 * ConvertOrderDialog trasforma una richiesta in ingresso in un ordine TMS
 * vero (POST /inbound-orders/{id}/convert).
 *
 * Il campo cliente è il punto delicato: il backend accetta la conversione
 * senza indicarlo solo se la richiesta ne porta già uno affidabile — cioè se
 * è arrivata dal portale, dove l'id viene dal JWT di chi l'ha inviata. Per
 * una richiesta arrivata via mail o PDF il campo `client` è testo libero
 * scritto dal mittente, che il backend si rifiuta di risolvere in
 * un'anagrafica per somiglianza del nome (chiunque scriva alla casella
 * potrebbe farsi intestare ordini altrui). In quel caso il cliente va scelto
 * qui, ed è una scelta dell'operatore, non un'inferenza.
 */
export default function ConvertOrderDialog({ order, onClose, onConverted }: Props) {
  const open = order !== null;
  const trustedCliente = order?.cliente_id ?? '';
  // L'elenco serve solo quando il dialog è aperto e c'è un cliente da scegliere.
  const { data: customersPage } = useGetCustomersQuery({ limit: 500 }, { skip: !open });
  const customers = useMemo(() => customersPage?.items ?? [], [customersPage]);

  const [clienteID, setClienteID] = useState('');
  const [tariffa, setTariffa] = useState('');
  const [note, setNote] = useState('');
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    if (!order) return;
    setClienteID(order.cliente_id ?? '');
    // La tariffa proposta esiste solo per le richieste dal portale ("tariffa
    // desiderata" del form cliente). Precompilarla rende visibile e
    // modificabile ciò che altrimenti il backend applicherebbe per default.
    setTariffa(order.tariffa_proposta ? String(order.tariffa_proposta) : '');
    setNote('');
  }, [order]);

  const clienteName = useMemo(
    () => customers.find((c) => c.id === clienteID)?.ragione_sociale ?? '',
    [customers, clienteID],
  );

  async function submit() {
    if (!order?.id) return;
    if (!clienteID) {
      toast.error('Seleziona il cliente da intestare all’ordine');
      return;
    }
    setSaving(true);
    try {
      const body: DtoInboundOrderConvertRequest = { cliente_id: clienteID };
      if (note.trim()) body.note = note.trim();
      // Solo un valore effettivamente digitato viene inviato: lasciare il
      // campo vuoto significa "applica il default del backend", non "tariffa
      // zero" — per questo la stringa vuota non diventa 0.
      if (tariffa.trim() !== '') {
        const parsed = Number(tariffa);
        if (Number.isNaN(parsed)) {
          toast.error('Tariffa non valida');
          setSaving(false);
          return;
        }
        body.tariffa = parsed;
      }

      const res = await apiClient.v1InboundOrdersConvertCreate(order.id, body);
      const created = res.data.order;
      toast.success(
        `Ordine ${created?.progressivo ?? ''} creato`,
        res.data.tariffa_from_client
          ? { description: 'Tariffa applicata: quella proposta dal cliente, non ancora rinegoziata.' }
          : undefined,
      );
      onClose();
      await onConverted();
    } catch (err) {
      toast.error(getInboundApiError(err));
    } finally {
      setSaving(false);
    }
  }

  return (
    <Dialog open={open} onOpenChange={(v) => { if (!v) onClose(); }}>
      <DialogContent className="max-w-lg">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <ArrowRightLeft className="h-4 w-4" />
            Converti in ordine
          </DialogTitle>
          <DialogDescription>
            {order?.ref ? `Richiesta ${order.ref} — ` : ''}
            {order?.client}
          </DialogDescription>
        </DialogHeader>

        {order && (
          <div className="space-y-4">
            <div className="rounded-md bg-muted/50 p-3 text-xs space-y-1">
              <div className="flex justify-between gap-4">
                <span className="text-muted-foreground">Tratta</span>
                <span className="text-right">{order.load_place || '—'} → {order.delivery_place || '—'}</span>
              </div>
              <div className="flex justify-between gap-4">
                <span className="text-muted-foreground">Date</span>
                <span className="text-right">{order.load_date || '—'} → {order.delivery_date || '—'}</span>
              </div>
              <div className="flex justify-between gap-4">
                <span className="text-muted-foreground">Prodotto</span>
                <span className="text-right">{order.product || '—'} · {fmtKg(order.kg)} kg</span>
              </div>
              {order.rate && (
                <div className="flex justify-between gap-4">
                  <span className="text-muted-foreground">Nolo indicato</span>
                  <span className="text-right font-mono">{order.rate}</span>
                </div>
              )}
            </div>

            <div className="space-y-1.5">
              <Label>Cliente da fatturare</Label>
              <Select value={clienteID} onValueChange={setClienteID}>
                <SelectTrigger>
                  <SelectValue placeholder="Seleziona il cliente" />
                </SelectTrigger>
                <SelectContent>
                  {customers.map((c) => (
                    <SelectItem key={c.id} value={c.id ?? ''}>{c.ragione_sociale}</SelectItem>
                  ))}
                </SelectContent>
              </Select>
              {trustedCliente ? (
                <p className="flex items-start gap-1.5 text-xs text-muted-foreground">
                  <Info className="mt-0.5 h-3 w-3 shrink-0" />
                  Richiesta inviata dal portale da
                  {' '}
                  <span className="font-medium">{clienteName || order.client}</span>
                  {': '}
                  cliente già identificato in fase di invio.
                </p>
              ) : (
                <p className="flex items-start gap-1.5 text-xs text-amber-600 dark:text-amber-400">
                  <TriangleAlert className="mt-0.5 h-3 w-3 shrink-0" />
                  {`«${order.client}» è il nome scritto dal mittente, non un'anagrafica: scegli tu a chi intestare l'ordine.`}
                </p>
              )}
            </div>

            <div className="space-y-1.5">
              <Label>Tariffa (€)</Label>
              <Input
                type="number"
                value={tariffa}
                onChange={(e) => setTariffa(e.target.value)}
                placeholder="0,00"
              />
              {order.tariffa_proposta ? (
                <p className="text-xs text-muted-foreground">
                  Tariffa desiderata dal cliente. Modificala per pattuirne un&apos;altra.
                </p>
              ) : (
                <p className="text-xs text-muted-foreground">
                  Nessuna tariffa strutturata nella richiesta: lasciando vuoto l&apos;ordine nasce a 0.
                </p>
              )}
            </div>

            <div className="space-y-1.5">
              <Label>Note aggiuntive</Label>
              <Textarea
                rows={2}
                value={note}
                onChange={(e) => setNote(e.target.value)}
                placeholder="Opzionale — si aggiunge a prodotto, kg e note della richiesta"
              />
            </div>
          </div>
        )}

        <DialogFooter>
          <Button variant="outline" onClick={onClose} disabled={saving}>Annulla</Button>
          <Button onClick={submit} disabled={saving || !clienteID}>
            {saving ? 'Conversione…' : 'Crea ordine'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

/**
 * ConvertedLink mostra, per una richiesta già convertita, il collegamento
 * all'ordine nato da essa — così la dashboard non lascia il dubbio se una
 * riga accettata sia diventata un ordine o si sia perduta.
 */
export function ConvertedLink({ orderID }: { orderID: string }) {
  return (
    <Link
      to={`/planner/ordini/${orderID}`}
      className="text-xs font-medium text-primary underline-offset-2 hover:underline"
      onClick={(e) => e.stopPropagation()}
    >
      Ordine creato
    </Link>
  );
}
