import { useState, useEffect, useCallback, useMemo } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { getOrder, startOrder, closeOrder, discardOrder, unassignOrder } from '@/lib/api';
import { getApiErrorMessage } from '@/lib/apiError';
import type { DtoOrderResponse } from '@/api/data-contracts';
import { Card } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Label } from '@/components/ui/label';
import { Skeleton } from '@/components/ui/skeleton';
import { StatusBadge } from '@/components/shared/StatusBadge';
import OrderRouteMap from '@/components/shared/OrderRouteMap';
import AssignOrderForm from '@/components/planner/AssignOrderForm';
import { toast } from 'sonner';
import { ArrowLeft, PlayCircle, CheckCircle, Ban, Warehouse, Droplets, RotateCcw } from 'lucide-react';
import { logger } from '@/lib/logger';
import { haversineKm } from '@/lib/geo';

export default function OrderDetailPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [order, setOrder] = useState<DtoOrderResponse | null>(null);
  const [loading, setLoading] = useState(true);

  const fetchOrder = useCallback(() => {
    if (!id) return;
    setLoading(true);
    getOrder(id).then((r: { data: DtoOrderResponse }) => setOrder(r.data)).catch((err: unknown) => { logger.error('Errore caricamento ordine:', err); toast.error('Ordine non trovato'); }).finally(() => setLoading(false));
  }, [id]);
  useEffect(() => { fetchOrder(); }, [fetchOrder]);

  const carico = order?.destinazione_carico;
  const scarico = order?.destinazione_scarico;
  const haversineDistanceKm = useMemo(() => haversineKm(carico, scarico), [carico, scarico]);
  const distanceKm = order?.route?.distance_km ?? haversineDistanceKm;

  const handleUnassign = async () => {
    if (!order?.id || !window.confirm(`Riportare l'ordine ${order.progressivo} a "da pianificare"? Mezzo, autista/vettore e percorso assegnati verranno azzerati.`)) return;
    try { await unassignOrder(order.id); toast.success(`Ordine ${order.progressivo} riportato a da pianificare`); fetchOrder(); }
    catch (e) { toast.error(getApiErrorMessage(e) || 'Errore'); }
  };

  const handleStart = async () => {
    if (!order?.id || !window.confirm(`Avviare il viaggio per l'ordine ${order.progressivo}?`)) return;
    try { await startOrder(order.id); toast.success(`Ordine ${order.progressivo} avviato`); fetchOrder(); }
    catch (e) { toast.error(getApiErrorMessage(e) || 'Errore'); }
  };
  const handleClose = async () => {
    if (!order?.id || !window.confirm(`Chiudere l'ordine ${order.progressivo}?`)) return;
    try { await closeOrder(order.id); toast.success(`Ordine ${order.progressivo} chiuso`); fetchOrder(); }
    catch (e) { toast.error(getApiErrorMessage(e) || 'Errore'); }
  };
  const handleDiscard = async () => {
    if (!order?.id || !window.confirm(`Scartare l'ordine ${order.progressivo}? L'operazione non è reversibile.`)) return;
    try { await discardOrder(order.id); toast.success(`Ordine ${order.progressivo} scartato`); navigate('/planner'); }
    catch (e) { toast.error(getApiErrorMessage(e) || 'Errore'); }
  };

  if (loading) {
    return (
      <div className="space-y-3">
        <Skeleton className="h-8 w-64" />
        <div className="grid grid-cols-1 lg:grid-cols-[1.6fr_1fr] gap-4">
          <Skeleton className="h-96 rounded-xl" />
          <Skeleton className="h-96 rounded-xl" />
        </div>
      </div>
    );
  }

  if (!order) {
    return <p className="text-muted-foreground text-center py-12">Ordine non trovato.</p>;
  }

  return (
    <div className="space-y-3" data-testid="order-detail-page">
      {/* Header */}
      <div className="flex items-center gap-3 flex-wrap">
        <button
          type="button" onClick={() => navigate('/planner')}
          className="flex items-center gap-1 text-sm text-muted-foreground hover:text-primary"
          data-testid="order-detail-back"
        >
          <ArrowLeft className="h-4 w-4" /> Planner
        </button>
        <span className="text-muted-foreground">/</span>
        <span className="font-mono text-sm font-semibold">{order.progressivo}</span>
        <div className="ml-auto flex items-center gap-2">
          {order.stato === 'PIANIFICABILE' && (
            <Button variant="outline" className="text-destructive" onClick={handleDiscard} data-testid="order-detail-discard">
              <Ban className="h-4 w-4 mr-2" /> Scarta
            </Button>
          )}
          {order.stato === 'PIANIFICATO' && !order.viaggio_id && (
            <>
              <Button variant="outline" className="text-destructive" onClick={handleDiscard} data-testid="order-detail-discard">
                <Ban className="h-4 w-4 mr-2" /> Scarta
              </Button>
              <Button variant="outline" onClick={handleUnassign} data-testid="order-detail-unassign">
                <RotateCcw className="h-4 w-4 mr-2" /> Riporta a da pianificare
              </Button>
              <Button onClick={handleStart} data-testid="order-detail-start">
                <PlayCircle className="h-4 w-4 mr-2" /> Avvia viaggio
              </Button>
            </>
          )}
          {order.stato === 'VIAGGIO' && (
            <Button onClick={handleClose} data-testid="order-detail-close">
              <CheckCircle className="h-4 w-4 mr-2" /> Chiudi
            </Button>
          )}
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-[1.6fr_1fr] gap-4 items-start">
        {/* Card principale: cliente, tratta, mappa, itinerario */}
        <Card className="rounded-xl border shadow-sm p-5 space-y-4">
          <div>
            <div className="flex items-center gap-2 flex-wrap">
              <h2 className="text-xl font-bold" style={{ fontFamily: "'Space Grotesk', sans-serif" }}>{order.cliente?.ragione_sociale}</h2>
              <StatusBadge stato={order.stato} />
            </div>
            <p className="text-sm text-muted-foreground mt-1">{order.destinazione_carico?.nome} → {order.destinazione_scarico?.nome}</p>
          </div>

          <OrderRouteMap carico={carico} scarico={scarico} garage={order.garage} washStation={order.wash_station} routePoints={order.route?.points as [number, number][] | undefined} height={280} />

          {/* Itinerario: 2 punti (carico/scarico) — nessun percorso stradale calcolato per il singolo ordine */}
          <div>
            <p className="text-[10px] uppercase tracking-wide text-muted-foreground font-semibold mb-4">Itinerario</p>
            <div className="flex items-start gap-3">
              <div className="flex flex-col gap-1 w-36 shrink-0">
                <span className="text-[10px] font-bold text-primary">↑ CARICO</span>
                <span className="text-sm font-bold leading-tight">{order.destinazione_carico?.nome}</span>
                <span className="text-xs text-muted-foreground">{order.data_ritiro}</span>
                {(order.ora_ritiro_da || order.ora_ritiro_a) && (
                  <span className="inline-flex w-fit items-center text-[11px] font-medium border rounded px-1.5 py-0.5 bg-muted/50">
                    {order.ora_ritiro_da}{order.ora_ritiro_a ? `–${order.ora_ritiro_a}` : ''}
                  </span>
                )}
              </div>
              <div className="flex-1 relative h-0.5 bg-primary mt-2.5">
                {distanceKm != null && (
                  <span className="absolute -top-5 left-1/2 -translate-x-1/2 text-xs font-bold text-primary whitespace-nowrap">
                    ~{distanceKm} km
                  </span>
                )}
              </div>
              <div className="flex flex-col items-end gap-1 w-36 shrink-0 text-right">
                <span className="text-[10px] font-bold text-primary">↓ SCARICO</span>
                <span className="text-sm font-bold leading-tight">{order.destinazione_scarico?.nome}</span>
                <span className="text-xs text-muted-foreground">{order.data_consegna}</span>
                {(order.ora_consegna_da || order.ora_consegna_a) && (
                  <span className="inline-flex w-fit items-center text-[11px] font-medium border rounded px-1.5 py-0.5 bg-muted/50">
                    {order.ora_consegna_da}{order.ora_consegna_a ? `–${order.ora_consegna_a}` : ''}
                  </span>
                )}
              </div>
            </div>
          </div>

          {order.viaggio_id && (
            <div className="rounded-lg border p-3 flex items-center gap-3 text-sm flex-wrap">
              <span>Ordine assegnato al viaggio <b className="font-mono">{order.viaggio_id.slice(0, 8)}</b></span>
            </div>
          )}

          {order.note && (
            <div>
              <p className="text-[10px] uppercase text-muted-foreground mb-1">Note</p>
              <p className="text-sm whitespace-pre-wrap">{order.note}</p>
            </div>
          )}
        </Card>

        {/* Card laterale: dettagli + link cruscotto cliente */}
        <Card className="rounded-xl border shadow-sm p-5 space-y-4">
          <p className="text-[10px] uppercase tracking-wide text-muted-foreground font-semibold">Dettagli</p>
          <div className="grid grid-cols-2 gap-3 text-sm">
            <div><p className="text-[10px] uppercase text-muted-foreground">Tipologia</p><p className="font-medium">{order.tipologia || '—'}</p></div>
            <div><p className="text-[10px] uppercase text-muted-foreground">Categoria</p><p className="font-medium">{order.categoria_trasporto || '—'}</p></div>
            <div><p className="text-[10px] uppercase text-muted-foreground">Tariffa</p><p className="font-medium">{order.tariffa ? `€ ${order.tariffa.toLocaleString('it-IT')}` : '—'}</p></div>
            <div><p className="text-[10px] uppercase text-muted-foreground">Rif. cliente</p><p className="font-medium">{order.rif_ordine_cliente || '—'}</p></div>
            <div><p className="text-[10px] uppercase text-muted-foreground">Punto di partenza</p><p className="font-medium">{order.garage?.nome || '—'}</p></div>
            <div><p className="text-[10px] uppercase text-muted-foreground">Punto di lavaggio</p><p className="font-medium">{order.wash_station?.nome || '—'}</p></div>
            <div><p className="text-[10px] uppercase text-muted-foreground">Mezzo</p><p className="font-medium font-mono">{order.motrice?.targa || '—'}</p></div>
            <div><p className="text-[10px] uppercase text-muted-foreground">Autista</p><p className="font-medium">{order.autista ? `${order.autista.nome} ${order.autista.cognome}` : '—'}</p></div>
            <div><p className="text-[10px] uppercase text-muted-foreground">Vettore</p><p className="font-medium">{order.vettore?.ragione_sociale || '—'}</p></div>
          </div>
          {order.cliente_id && (
            <Button variant="outline" size="sm" className="w-full text-xs" onClick={() => navigate(`/anagrafiche/clienti/${order.cliente_id}/cruscotto`)}>
              Vai al cruscotto cliente →
            </Button>
          )}
        </Card>
      </div>

      {order.stato === 'PIANIFICABILE' ? (
        <Card className="rounded-xl border shadow-sm p-5" data-testid="order-detail-assign-form">
          <AssignOrderForm order={order} onAssigned={fetchOrder} />
        </Card>
      ) : (order.garage || order.wash_station || order.autista || order.vettore) && (
        <Card className="rounded-xl border shadow-sm p-5" data-testid="order-detail-transport-summary">
          <div className="space-y-5">
            <div>
              <p className="text-[10px] uppercase tracking-wide text-muted-foreground font-semibold mb-3">Dettagli trasporto</p>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
                <div className="space-y-1.5">
                  <Label>Punto di partenza</Label>
                  <div className="flex items-center gap-2 text-sm border border-dashed rounded-md px-3 py-2">
                    <Warehouse className="h-4 w-4 text-muted-foreground shrink-0" />
                    <span>{order.garage?.nome || '—'}</span>
                  </div>
                </div>
                <div className="space-y-1.5">
                  <Label>Punto di lavaggio — dopo lo scarico</Label>
                  <div className="flex items-center gap-2 text-sm border border-dashed rounded-md px-3 py-2">
                    <Droplets className="h-4 w-4 text-muted-foreground shrink-0" />
                    <span>{order.wash_station?.nome || '—'}</span>
                  </div>
                </div>
              </div>
            </div>
            <div>
              <p className="text-[10px] uppercase tracking-wide text-muted-foreground font-semibold mb-3">Chi effettua il trasporto?</p>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
                {order.vettore ? (
                  <div className="space-y-1.5">
                    <Label>Vettore</Label>
                    <p className="text-sm font-medium">{order.vettore.ragione_sociale}</p>
                  </div>
                ) : (
                  <>
                    <div className="space-y-1.5">
                      <Label>Autista</Label>
                      <p className="text-sm font-medium">{order.autista ? `${order.autista.nome} ${order.autista.cognome}` : '—'}</p>
                    </div>
                    <div className="space-y-1.5">
                      <Label>Motrice</Label>
                      <p className="text-sm font-medium font-mono">{order.motrice?.targa || '—'}</p>
                    </div>
                    {order.semirimorchio?.targa && (
                      <div className="space-y-1.5">
                        <Label>Rimorchio</Label>
                        <p className="text-sm font-medium font-mono">{order.semirimorchio.targa}</p>
                      </div>
                    )}
                  </>
                )}
              </div>
            </div>
          </div>
        </Card>
      )}
    </div>
  );
}
