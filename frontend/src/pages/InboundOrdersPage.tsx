import { useCallback, useEffect, useMemo, useState } from 'react';
import { toast } from 'sonner';
import {
  Check, X, Pencil, FileText, RefreshCw, RotateCcw, Search, Loader2,
} from 'lucide-react';
import { apiClient } from '@/lib/apiClient';
import type { DtoInboundConfigResponse, DtoInboundOrderResponse } from '@/api/data-contracts';
import { Card } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import {
  Table, TableBody, TableCell, TableHead, TableHeader, TableRow,
} from '@/components/ui/table';
import ImportPdfDialog from '@/components/inbound/ImportPdfDialog';
import InboundOrderDetail from '@/components/inbound/InboundOrderDetail';
import { STATUS_BADGE, getInboundApiError } from '@/components/inbound/constants';

// I "canale" del mockup (Trimble, Oracle OTM, Transporeon…) non esistono nel
// modello reale: l'unico dato disponibile è la fonte (source) dell'ordine.
function ChannelBadge({ source }: { source?: string }) {
  if (source === 'mail') {
    return (
      <Badge variant="secondary" className="bg-amber-50 text-amber-700 dark:bg-amber-500/10 dark:text-amber-300">
        Email · AI
      </Badge>
    );
  }
  if (source === 'pdf') {
    return (
      <Badge variant="secondary" className="bg-sky-50 text-sky-700 dark:bg-sky-500/10 dark:text-sky-300">
        PDF · AI
      </Badge>
    );
  }
  return <Badge variant="outline">Manuale</Badge>;
}

// "Da approvare" / "Da confermare su portale" sono entrambi status "pending":
// la distinzione è il flag reale `portal` (PO da confermare sul portale del
// cliente, fuori dal nostro flusso di accettazione).
function statoInfo(o: DtoInboundOrderResponse) {
  if (o.status === 'accepted') return STATUS_BADGE.accepted;
  if (o.status === 'modify') return STATUS_BADGE.modify;
  if (o.portal) {
    return { label: 'Da confermare su portale', className: 'bg-blue-100 text-blue-800 dark:bg-blue-500/15 dark:text-blue-300' };
  }
  return { label: 'Da approvare', className: 'bg-rose-100 text-rose-800 dark:bg-rose-500/15 dark:text-rose-300' };
}

function fmtRelative(iso?: string): string {
  if (!iso) return '—';
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return '—';
  const min = Math.floor((Date.now() - d.getTime()) / 60000);
  if (min < 1) return 'ora';
  if (min < 60) return `${min} min fa`;
  const h = Math.floor(min / 60);
  if (h < 24) return `${h} h fa`;
  const time = d.toLocaleTimeString('it-IT', { hour: '2-digit', minute: '2-digit' });
  const today = new Date();
  const isYesterday = d.getDate() === today.getDate() - 1
    && d.getMonth() === today.getMonth() && d.getFullYear() === today.getFullYear();
  return isYesterday ? `ieri ${time}` : d.toLocaleDateString('it-IT');
}

function mailtoModify(o: DtoInboundOrderResponse) {
  const subject = `Ordine ${o.ref} — richiesta di modifica`;
  const body = [
    'Buongiorno,', '',
    `in merito al vostro ordine ${o.ref} (${o.product}, carico ${o.load_date}),`,
    'non possiamo confermarlo alle condizioni attuali e chiediamo una modifica:', '',
    '[ scrivere qui la modifica richiesta ]', '',
    'Cordiali saluti,',
    'Feccia F.lli S.r.l. — Order Desk',
  ].join('\n');
  return `mailto:${o.sender_email}?subject=${encodeURIComponent(subject)}&body=${encodeURIComponent(body)}`;
}

// Dashboard di accettazione degli ordini in ingresso (porting OrderMesh):
// ordini arrivati via e-mail o importati da PDF, con filtri per stato,
// scansione della casella e import PDF guidato.
export default function InboundOrdersPage() {
  const [orders, setOrders] = useState<DtoInboundOrderResponse[]>([]);
  const [config, setConfig] = useState<DtoInboundConfigResponse>({});
  const [loading, setLoading] = useState(true);
  const [fStatus, setFStatus] = useState('queue');
  const [fText, setFText] = useState('');
  const [selected, setSelected] = useState<DtoInboundOrderResponse | null>(null);
  const [busy, setBusy] = useState<string | null>(null); // order id con azione in corso
  const [scraping, setScraping] = useState(false);
  const [importOpen, setImportOpen] = useState(false);

  const loadAll = useCallback(async () => {
    const [o, c] = await Promise.all([
      apiClient.v1InboundOrdersList(),
      apiClient.v1InboundConfigList(),
    ]);
    setOrders(o.data ?? []);
    setConfig(c.data ?? {});
  }, []);

  useEffect(() => {
    loadAll()
      .catch((err) => toast.error(`Errore di caricamento: ${getInboundApiError(err)}`))
      .finally(() => setLoading(false));
  }, [loadAll]);

  const counts = useMemo(() => {
    const queue = orders.filter((o) => o.status !== 'accepted');
    return {
      queue: queue.length,
      due: queue.filter((o) => !o.portal).length,
      portal: queue.filter((o) => o.portal).length,
      accepted: orders.filter((o) => o.status === 'accepted').length,
    };
  }, [orders]);

  const pills: Array<[string, string, number]> = [
    ['queue', 'Tutti · in coda', counts.queue],
    ['due', 'Da approvare', counts.due],
    ['portal', 'Da confermare su portale', counts.portal],
    ['accepted', 'Accettati', counts.accepted],
  ];

  const visible = useMemo(() => orders.filter((o) => {
    if (fStatus === 'due') { if (o.status === 'accepted' || o.portal) return false; }
    else if (fStatus === 'portal') { if (o.status === 'accepted' || !o.portal) return false; }
    else if (fStatus === 'accepted') { if (o.status !== 'accepted') return false; }
    else if (fStatus === 'queue') { if (o.status === 'accepted') return false; }
    if (fText) {
      const hay = `${o.ref} ${o.product} ${o.load_place} ${o.delivery_place} ${o.load_date} ${o.delivery_date} ${o.client} ${o.sender_email}`.toLowerCase();
      if (!hay.includes(fText.toLowerCase())) return false;
    }
    return true;
  }), [orders, fStatus, fText]);

  async function doAction(o: DtoInboundOrderResponse, action: 'accept' | 'modify' | 'reset') {
    if (!o.id) return;
    setBusy(o.id);
    try {
      if (action === 'accept') {
        const res = await apiClient.v1InboundOrdersAcceptCreate(o.id);
        toast.success(`Ordine accettato — ${res.data.mail ?? ''}`);
      } else if (action === 'modify') {
        await apiClient.v1InboundOrdersModifyCreate(o.id);
        window.location.href = mailtoModify(o); // apre il client di posta con mittente precompilato
        toast.success(`Aperto il client di posta verso ${o.sender_email}`);
      } else {
        await apiClient.v1InboundOrdersResetCreate(o.id);
      }
      setSelected(null);
      await loadAll();
    } catch (err) {
      toast.error(getInboundApiError(err));
    } finally {
      setBusy(null);
    }
  }

  async function doScrape() {
    setScraping(true);
    try {
      const res = await apiClient.v1InboundOrdersScrapeCreate();
      toast.success(`Casella letta: ${res.data.scanned ?? 0} mail esaminate, ${res.data.added ?? 0} nuovi ordini`);
      await loadAll();
    } catch (err) {
      toast.error(getInboundApiError(err));
    } finally {
      setScraping(false);
    }
  }

  return (
    <div data-testid="inbound-orders-page" className="space-y-4">
      {config.accept_mode && config.accept_mode !== 'production' && (
        <Badge variant="secondary" className="bg-amber-100 text-amber-800 dark:bg-amber-500/15 dark:text-amber-300">
          TEST → {config.test_recipient || 'destinatario non impostato'}
        </Badge>
      )}

      <div className="flex flex-wrap items-center gap-2">
        <div className="relative flex-1 min-w-[220px] max-w-sm">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-muted-foreground" />
          <Input
            placeholder="Cerca ordine, cliente, prodotto…"
            value={fText}
            onChange={(e) => setFText(e.target.value)}
            className="pl-9 h-9 text-sm"
          />
        </div>
        <div className="flex rounded-lg border p-0.5 gap-0.5">
          {pills.map(([value, label, n]) => (
            <Button
              key={value}
              variant={fStatus === value ? 'secondary' : 'ghost'}
              size="sm"
              className="h-7 gap-1.5 px-2 text-xs"
              onClick={() => setFStatus(value)}
            >
              {label} <span className="font-semibold">{n}</span>
            </Button>
          ))}
        </div>
        <Button
          variant="outline"
          size="sm"
          className="ml-auto gap-1.5 text-xs"
          onClick={doScrape}
          disabled={scraping || !config.mailbox_ready}
          title={config.mailbox_ready ? 'Legge subito la casella di posta' : 'Casella non configurata o non autenticata (vedi backend)'}
        >
          {scraping ? <Loader2 className="h-3.5 w-3.5 animate-spin" /> : <RefreshCw className="h-3.5 w-3.5" />}
          Scansiona casella
        </Button>
        <Button size="sm" className="gap-1.5 text-xs" onClick={() => setImportOpen(true)}>
          <FileText className="h-3.5 w-3.5" /> Carica da PDF
        </Button>
      </div>

      <Card className="rounded-xl border shadow-sm">
        <div className="overflow-x-auto">
          <Table className="text-xs md:text-sm" style={{ minWidth: 1000 }}>
            <TableHeader>
              <TableRow>
                <TableHead className="py-2 text-xs">Canale</TableHead>
                <TableHead className="py-2 text-xs">Ordine</TableHead>
                <TableHead className="py-2 text-xs">Cliente</TableHead>
                <TableHead className="py-2 text-xs">Tratta</TableHead>
                <TableHead className="py-2 text-xs">Ritiro</TableHead>
                <TableHead className="py-2 text-xs">Consegna</TableHead>
                <TableHead className="py-2 text-xs text-right">Tariffa</TableHead>
                <TableHead className="py-2 text-xs">Stato</TableHead>
                <TableHead className="py-2 text-xs" />
              </TableRow>
            </TableHeader>
            <TableBody>
              {loading && (
                <TableRow>
                  <TableCell colSpan={9} className="py-8 text-center text-muted-foreground">
                    Caricamento…
                  </TableCell>
                </TableRow>
              )}
              {!loading && visible.length === 0 && (
                <TableRow>
                  <TableCell colSpan={9} className="py-8 text-center text-muted-foreground">
                    Nessun ordine con i filtri correnti.
                  </TableCell>
                </TableRow>
              )}
              {visible.map((o) => {
                const stato = statoInfo(o);
                const actionable = o.status !== 'accepted' && !o.portal;
                return (
                  <TableRow
                    key={o.id}
                    className="cursor-pointer hover:bg-muted/60"
                    onClick={() => setSelected(o)}
                  >
                    <TableCell className="py-2"><ChannelBadge source={o.source} /></TableCell>
                    <TableCell className="py-2">
                      <div className="font-mono text-xs font-semibold">{o.ref || '—'}</div>
                      <div className="text-xs text-muted-foreground">{fmtRelative(o.received_at)}</div>
                    </TableCell>
                    <TableCell className="py-2 min-w-0">
                      <div className="truncate font-semibold">{o.client}</div>
                      <div className="truncate text-xs text-muted-foreground">{o.product}</div>
                    </TableCell>
                    <TableCell className="py-2 max-w-[220px] truncate text-xs">
                      {o.load_place || '—'} → {o.delivery_place || '—'}
                    </TableCell>
                    <TableCell className="py-2 text-xs">
                      {o.load_date || <span className="italic text-muted-foreground">da definire</span>}
                    </TableCell>
                    <TableCell className="py-2 text-xs">
                      {o.delivery_date || <span className="italic text-muted-foreground">da definire</span>}
                    </TableCell>
                    <TableCell className="py-2 text-right font-mono text-xs font-semibold">{o.rate || '—'}</TableCell>
                    <TableCell className="py-2">
                      <Badge variant="secondary" className={stato.className}>{stato.label}</Badge>
                    </TableCell>
                    <TableCell className="py-2" onClick={(e) => e.stopPropagation()}>
                      <div className="flex items-center justify-end gap-1">
                        {o.status === 'accepted' && (
                          <Button
                            variant="ghost" size="icon" className="h-7 w-7"
                            title="Riporta in attesa" disabled={busy === o.id}
                            onClick={() => doAction(o, 'reset')}
                          >
                            <RotateCcw className="h-3.5 w-3.5" />
                          </Button>
                        )}
                        {actionable && (
                          <>
                            <Button
                              variant="ghost" size="icon"
                              className="h-7 w-7 text-emerald-600 hover:bg-emerald-50 hover:text-emerald-700 dark:text-emerald-400 dark:hover:bg-emerald-500/10"
                              title="Accetta" disabled={busy === o.id}
                              onClick={() => doAction(o, 'accept')}
                            >
                              <Check className="h-3.5 w-3.5" />
                            </Button>
                            <Button
                              variant="ghost" size="icon"
                              className="h-7 w-7 text-destructive hover:bg-destructive/10"
                              title="Non accettare / Modifica" disabled={busy === o.id}
                              onClick={() => doAction(o, 'modify')}
                            >
                              <X className="h-3.5 w-3.5" />
                            </Button>
                            <Button
                              variant="ghost" size="icon" className="h-7 w-7"
                              title="Dettaglio" onClick={() => setSelected(o)}
                            >
                              <Pencil className="h-3.5 w-3.5" />
                            </Button>
                          </>
                        )}
                      </div>
                    </TableCell>
                  </TableRow>
                );
              })}
            </TableBody>
          </Table>
        </div>
      </Card>

      <p className="text-xs text-muted-foreground">
        Clic su una riga per il dettaglio dell’ordine. «Accetta» invia l’email di conferma;
        «Modifica» apre il client di posta con il mittente precompilato. Gli ordini «da confermare
        su portale» richiedono conferma diretta sul portale del cliente.
      </p>

      <ImportPdfDialog
        open={importOpen}
        onOpenChange={setImportOpen}
        orders={orders}
        onImported={loadAll}
      />
      <InboundOrderDetail
        order={selected}
        onOpenChange={(v) => { if (!v) setSelected(null); }}
        busy={busy}
        onAction={doAction}
      />
    </div>
  );
}
