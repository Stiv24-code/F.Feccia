import { ArrowRight, Check, RotateCcw, X } from 'lucide-react';
import type { DtoInboundOrderResponse } from '@/api/data-contracts';
import {
  Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle,
} from '@/components/ui/dialog';
import { Card } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Separator } from '@/components/ui/separator';
import { STATUS_BADGE, fmtKg } from './constants';

const fmtDateTime = (iso?: string) => (iso ? new Date(iso).toLocaleString('it-IT') : '—');

interface InboundOrderDetailProps {
  order: DtoInboundOrderResponse | null;
  onOpenChange: (open: boolean) => void;
  busy: string | null;
  onAction: (order: DtoInboundOrderResponse, action: 'accept' | 'modify' | 'reset') => void;
}

function Field({ label, value, mono }: { label: string; value?: string; mono?: boolean }) {
  return (
    <div>
      <p className="text-[9px] font-semibold uppercase tracking-wider text-muted-foreground">{label}</p>
      <p className={`mt-0.5 text-sm font-medium ${mono ? 'font-mono' : ''}`}>{value || '—'}</p>
    </div>
  );
}

// Dettaglio ordine in arrivo (versione alleggerita rispetto al mockup di
// riferimento): usa solo i campi già presenti in DtoInboundOrderResponse —
// niente mappa, niente statistiche/tariffario cliente, dati non disponibili
// da nessuna API oggi.
export default function InboundOrderDetail({ order, onOpenChange, busy, onAction }: InboundOrderDetailProps) {
  return (
    <Dialog open={!!order} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[720px] max-h-[85vh] overflow-y-auto">
        {order && (() => {
          const badge = STATUS_BADGE[order.status ?? 'pending'] ?? STATUS_BADGE.pending;
          return (
            <>
              <DialogHeader>
                <div className="flex flex-wrap items-center gap-2">
                  <DialogTitle>{order.ref || 'Ordine in arrivo'}</DialogTitle>
                  <Badge variant="secondary" className={badge.className}>{badge.label}</Badge>
                  {order.portal && (
                    <Badge variant="secondary" className="bg-violet-100 text-violet-800 dark:bg-violet-500/15 dark:text-violet-300">
                      portale
                    </Badge>
                  )}
                  {order.source === 'mail' && <Badge variant="outline">da e-mail</Badge>}
                  {order.source === 'pdf' && <Badge variant="outline">da PDF</Badge>}
                  {order.source === 'portal' && <Badge variant="outline">da portale clienti</Badge>}
                </div>
                <DialogDescription>{order.client}</DialogDescription>
              </DialogHeader>

              <Card className="p-4">
                <div className="flex items-center justify-between gap-4">
                  <div className="min-w-0">
                    <p className="text-[10px] font-semibold uppercase tracking-wider text-primary">↑ Carico</p>
                    <p className="mt-1 truncate text-sm font-semibold">{order.load_place || '—'}</p>
                    <p className="text-xs text-muted-foreground">{order.load_date || 'data da definire'}</p>
                  </div>
                  <ArrowRight className="h-4 w-4 shrink-0 text-muted-foreground" />
                  <div className="min-w-0 text-right">
                    <p className="text-[10px] font-semibold uppercase tracking-wider text-primary">↓ Consegna</p>
                    <p className="mt-1 truncate text-sm font-semibold">{order.delivery_place || '—'}</p>
                    <p className="text-xs text-muted-foreground">{order.delivery_date || 'data da definire'}</p>
                  </div>
                </div>
                {order.rate && (
                  <>
                    <Separator className="my-3" />
                    <div className="flex items-baseline justify-between">
                      <span className="text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">
                        Tariffa
                      </span>
                      <span className="text-lg font-bold">{order.rate}</span>
                    </div>
                  </>
                )}
              </Card>

              <Card className="p-4">
                <p className="mb-3 text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">
                  Dettagli ordine
                </p>
                <div className="grid grid-cols-2 gap-3 sm:grid-cols-3">
                  <Field label="Prodotto" value={order.product} />
                  <Field label="Kg" value={fmtKg(order.kg)} />
                  <Field label="Riferimento" value={order.ref} mono />
                  <Field label="Mittente" value={order.sender_email} />
                  <Field
                    label="Origine"
                    value={
                      order.source === 'mail' ? 'e-mail'
                      : order.source === 'pdf' ? 'import PDF'
                      : order.source === 'portal' ? 'portale clienti'
                      : order.source
                    }
                  />
                  <Field label="Ricevuto" value={fmtDateTime(order.received_at)} />
                </div>
              </Card>

              {order.notes && (
                <Card className="p-4">
                  <p className="mb-2 text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">
                    Note
                  </p>
                  <p className="whitespace-pre-wrap text-sm">{order.notes}</p>
                </Card>
              )}

              <div className="flex justify-end gap-2">
                {order.status === 'accepted' ? (
                  <Button
                    variant="outline" size="sm" className="gap-1.5 text-xs"
                    disabled={busy === order.id}
                    onClick={() => onAction(order, 'reset')}
                  >
                    <RotateCcw className="h-3.5 w-3.5" /> Riporta in attesa
                  </Button>
                ) : (
                  <>
                    <Button
                      variant="outline" size="sm"
                      className="gap-1.5 text-xs text-destructive border-destructive/40 hover:bg-destructive/10"
                      disabled={busy === order.id}
                      onClick={() => onAction(order, 'modify')}
                    >
                      <X className="h-3.5 w-3.5" /> Non accettare / Modifica
                    </Button>
                    <Button
                      size="sm" className="gap-1.5 text-xs"
                      disabled={busy === order.id}
                      onClick={() => onAction(order, 'accept')}
                    >
                      <Check className="h-3.5 w-3.5" /> Accetta
                    </Button>
                  </>
                )}
              </div>
            </>
          );
        })()}
      </DialogContent>
    </Dialog>
  );
}
