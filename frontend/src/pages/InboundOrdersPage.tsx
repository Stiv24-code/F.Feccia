import { Fragment, useCallback, useEffect, useMemo, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { toast } from 'sonner';
import {
  Check, X, FileText, RefreshCw, RotateCcw, Search, Settings2, Loader2,
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
import { STATUS_BADGE, fmtKg, getInboundApiError } from '@/components/inbound/constants';

const STATUS_FILTERS: Array<[string, string]> = [
  ['all', 'Tutti'],
  ['pending', 'Da confermare'],
  ['accepted', 'Accettati'],
  ['modify', 'In modifica'],
  ['mail', 'Da e-mail'],
  ['pdf', 'Da PDF'],
];

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
// ordini arrivati via e-mail o importati da PDF, con filtri per cliente e
// stato, scansione della casella e import PDF guidato.
export default function InboundOrdersPage() {
  const navigate = useNavigate();
  const [orders, setOrders] = useState<DtoInboundOrderResponse[]>([]);
  const [config, setConfig] = useState<DtoInboundConfigResponse>({});
  const [loading, setLoading] = useState(true);
  const [fClient, setFClient] = useState('all');
  const [fStatus, setFStatus] = useState('all');
  const [fText, setFText] = useState('');
  const [expanded, setExpanded] = useState<string | null>(null);
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

  const clients = useMemo(() => {
    const map = new Map<string, { name: string; email: string; n: number; acc: number }>();
    for (const o of orders) {
      const name = o.client ?? '';
      if (!map.has(name)) map.set(name, { name, email: o.sender_email ?? '', n: 0, acc: 0 });
      const c = map.get(name)!;
      c.n++;
      if (o.status === 'accepted') c.acc++;
    }
    return [...map.values()].sort((a, b) => a.name.localeCompare(b.name));
  }, [orders]);

  const visible = useMemo(() => orders.filter((o) => {
    if (fClient !== 'all' && o.client !== fClient) return false;
    if (fStatus === 'mail') { if (o.source !== 'mail') return false; }
    else if (fStatus === 'pdf') { if (o.source !== 'pdf') return false; }
    else if (fStatus !== 'all' && o.status !== fStatus) return false;
    if (fText) {
      const hay = `${o.ref} ${o.product} ${o.load_place} ${o.delivery_place} ${o.load_date} ${o.delivery_date} ${o.client} ${o.sender_email}`.toLowerCase();
      if (!hay.includes(fText.toLowerCase())) return false;
    }
    return true;
  }), [orders, fClient, fStatus, fText]);

  const nByStatus = (s: string) => orders.filter((o) => o.status === s).length;

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
      {/* ── Header: badge test + stat + azioni pagina ── */}
      <div className="flex flex-wrap items-end justify-between gap-4">
        <div className="flex items-center gap-3">
          {config.accept_mode && config.accept_mode !== 'production' && (
            <Badge variant="secondary" className="bg-amber-100 text-amber-800 dark:bg-amber-500/15 dark:text-amber-300">
              TEST → {config.test_recipient || 'destinatario non impostato'}
            </Badge>
          )}
          <Button
            variant="ghost"
            size="sm"
            className="gap-1.5 text-xs"
            onClick={() => navigate('/ordini-in-ingresso/template')}
          >
            <Settings2 className="h-3.5 w-3.5" /> Template PDF
          </Button>
        </div>
        <div className="grid grid-cols-4 gap-2">
          <StatCard label="clienti" value={clients.length} />
          <StatCard label="ordini" value={orders.length} />
          <StatCard label="da confermare" value={nByStatus('pending')} tone="amber" />
          <StatCard label="accettati" value={nByStatus('accepted')} tone="emerald" />
        </div>
      </div>

      <div className="grid gap-4 items-start md:grid-cols-[260px_1fr]">
        {/* ── Sidebar clienti ── */}
        <Card className="p-3">
          <p className="text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">Clienti</p>
          <div className="mt-2 flex flex-col gap-1">
            <ClientButton
              active={fClient === 'all'}
              name="Tutti i clienti"
              sub={`${orders.length} ordini`}
              count={`${orders.length}`}
              onClick={() => setFClient('all')}
            />
            {clients.map((c) => (
              <ClientButton
                key={c.name}
                active={fClient === c.name}
                name={c.name}
                sub={c.email}
                count={c.acc ? `${c.acc}✓/${c.n}` : `${c.n}`}
                onClick={() => setFClient(c.name)}
              />
            ))}
          </div>
        </Card>

        {/* ── Main ── */}
        <div className="space-y-3">
          <div className="flex flex-wrap items-center gap-2">
            <div className="relative flex-1 min-w-[220px] max-w-sm">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-muted-foreground" />
              <Input
                placeholder="Cerca prodotto, località, riferimento…"
                value={fText}
                onChange={(e) => setFText(e.target.value)}
                className="pl-9 h-9 text-sm"
              />
            </div>
            <div className="flex rounded-lg border p-0.5 gap-0.5">
              {STATUS_FILTERS.map(([value, label]) => (
                <Button
                  key={value}
                  variant={fStatus === value ? 'secondary' : 'ghost'}
                  size="sm"
                  className="h-7 px-2 text-xs"
                  onClick={() => setFStatus(value)}
                >
                  {label}
                </Button>
              ))}
            </div>
            <Button
              variant="outline"
              size="sm"
              className="gap-1.5 text-xs"
              onClick={doScrape}
              disabled={scraping || !config.mailbox_ready}
              title={config.mailbox_ready ? 'Legge subito la casella di posta' : 'Casella non configurata o non autenticata (vedi backend)'}
            >
              {scraping ? <Loader2 className="h-3.5 w-3.5 animate-spin" /> : <RefreshCw className="h-3.5 w-3.5" />}
              Scansiona casella
            </Button>
            <Button size="sm" className="gap-1.5 text-xs" onClick={() => setImportOpen(true)}>
              <FileText className="h-3.5 w-3.5" /> Importa PDF
            </Button>
          </div>

          <Card className="rounded-xl border shadow-sm">
            <div className="overflow-x-auto">
              <Table className="text-xs md:text-sm" style={{ minWidth: 1000 }}>
                <TableHeader>
                  <TableRow>
                    <TableHead className="py-2 text-xs">Stato</TableHead>
                    <TableHead className="py-2 text-xs">Cliente</TableHead>
                    <TableHead className="py-2 text-xs">Riferimento</TableHead>
                    <TableHead className="py-2 text-xs">Prodotto</TableHead>
                    <TableHead className="py-2 text-xs text-right">Kg</TableHead>
                    <TableHead className="py-2 text-xs">Carico</TableHead>
                    <TableHead className="py-2 text-xs">Consegna</TableHead>
                    <TableHead className="py-2 text-xs">Nolo</TableHead>
                    <TableHead className="py-2 text-xs">Azioni</TableHead>
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
                    const badge = STATUS_BADGE[o.status ?? 'pending'] ?? STATUS_BADGE.pending;
                    return (
                      <Fragment key={o.id}>
                        <TableRow
                          className="cursor-pointer hover:bg-muted/60"
                          onClick={() => setExpanded(expanded === o.id ? null : (o.id ?? null))}
                        >
                          <TableCell className="py-2">
                            <div className="flex flex-col items-start gap-1">
                              <Badge variant="secondary" className={badge.className}>{badge.label}</Badge>
                              {o.portal && <Badge variant="secondary" className="bg-violet-100 text-violet-800 dark:bg-violet-500/15 dark:text-violet-300">portale</Badge>}
                              {o.source === 'mail' && <Badge variant="outline">da e-mail</Badge>}
                              {o.source === 'pdf' && <Badge variant="outline">da PDF</Badge>}
                            </div>
                          </TableCell>
                          <TableCell className="py-2">
                            <div className="font-semibold">{o.client}</div>
                            <div className="text-xs text-muted-foreground">{o.sender_email}</div>
                          </TableCell>
                          <TableCell className="py-2 font-mono text-xs">{o.ref}</TableCell>
                          <TableCell className="py-2">{o.product}</TableCell>
                          <TableCell className="py-2 text-right font-mono text-xs">{fmtKg(o.kg)}</TableCell>
                          <TableCell className="py-2">
                            <div className="font-mono text-xs">{o.load_date}</div>
                            <div className="text-xs text-muted-foreground">{o.load_place}</div>
                          </TableCell>
                          <TableCell className="py-2">
                            <div className="font-mono text-xs">{o.delivery_date}</div>
                            <div className="text-xs text-muted-foreground">{o.delivery_place}</div>
                          </TableCell>
                          <TableCell className="py-2 text-xs">{o.rate}</TableCell>
                          <TableCell className="py-2" onClick={(e) => e.stopPropagation()}>
                            <div className="flex flex-col gap-1">
                              {o.status === 'accepted' ? (
                                <Button
                                  variant="outline" size="sm" className="h-7 gap-1 text-xs"
                                  disabled={busy === o.id}
                                  onClick={() => doAction(o, 'reset')}
                                >
                                  <RotateCcw className="h-3 w-3" /> Riporta in attesa
                                </Button>
                              ) : (
                                <>
                                  <Button
                                    size="sm" className="h-7 gap-1 text-xs"
                                    disabled={busy === o.id}
                                    onClick={() => doAction(o, 'accept')}
                                  >
                                    <Check className="h-3 w-3" /> Accetta
                                  </Button>
                                  <Button
                                    variant="outline" size="sm"
                                    className="h-7 gap-1 text-xs text-destructive border-destructive/40 hover:bg-destructive/10"
                                    disabled={busy === o.id}
                                    onClick={() => doAction(o, 'modify')}
                                  >
                                    <X className="h-3 w-3" /> Non accettare / Modifica
                                  </Button>
                                </>
                              )}
                            </div>
                          </TableCell>
                        </TableRow>
                        {expanded === o.id && (
                          <TableRow>
                            <TableCell colSpan={9} className="bg-muted/50 py-2 text-xs">
                              <b>Mittente:</b> {o.sender_email} · <b>Origine:</b>{' '}
                              {o.source === 'mail' ? 'e-mail' : o.source === 'pdf' ? 'import PDF (template)' : o.source}
                              <br />
                              {o.notes || 'Nessuna nota.'}
                            </TableCell>
                          </TableRow>
                        )}
                      </Fragment>
                    );
                  })}
                </TableBody>
              </Table>
            </div>
          </Card>

          <p className="text-xs text-muted-foreground">
            Clic su una riga per note lavaggio / attrezzatura. «Accetta» invia l’email di conferma;
            «Modifica» apre il client di posta con il mittente precompilato.
          </p>
        </div>
      </div>

      <ImportPdfDialog
        open={importOpen}
        onOpenChange={setImportOpen}
        orders={orders}
        onImported={loadAll}
      />
    </div>
  );
}

function StatCard({ label, value, tone }: { label: string; value: number; tone?: 'amber' | 'emerald' }) {
  const toneClass = tone === 'amber'
    ? 'text-amber-600 dark:text-amber-400'
    : tone === 'emerald'
      ? 'text-emerald-600 dark:text-emerald-400'
      : '';
  return (
    <Card className="px-3 py-2">
      <div className={`text-xl font-bold tabular-nums ${toneClass}`}>{value}</div>
      <div className="text-[10px] uppercase tracking-wider text-muted-foreground">{label}</div>
    </Card>
  );
}

function ClientButton({ active, name, sub, count, onClick }: {
  active: boolean; name: string; sub: string; count: string; onClick: () => void;
}) {
  return (
    <button
      onClick={onClick}
      className={`flex w-full items-center justify-between gap-2 rounded-lg px-2 py-1.5 text-left transition-colors ${
        active ? 'bg-accent text-accent-foreground' : 'hover:bg-muted'
      }`}
    >
      <span className="min-w-0">
        <span className={`block truncate text-sm ${active ? 'font-semibold' : ''}`}>{name}</span>
        <span className="block truncate text-xs text-muted-foreground">{sub}</span>
      </span>
      <Badge variant="outline" className="shrink-0">{count}</Badge>
    </button>
  );
}
