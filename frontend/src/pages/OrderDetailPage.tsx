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
import RouteItinerary, { type ItineraryStop } from '@/components/shared/RouteItinerary';
import AssignOrderForm from '@/components/planner/AssignOrderForm';
import { toast } from 'sonner';
import { ArrowLeft, PlayCircle, CheckCircle, Ban, RotateCcw } from 'lucide-react';
import { logger } from '@/lib/logger';

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

  // Itinerario a 4 tappe (garage/lavaggio opzionali) per gli stati già
  // assegnati — per PIANIFICABILE l'itinerario vive dentro AssignOrderForm,
  // che ha accesso alle tappe "live" mentre l'utente sceglie garage/lavaggio.
  const fullItineraryStops = useMemo<ItineraryStop[]>(() => {
    if (!order) return [];
    const ritiroChip = order.ora_ritiro_da || order.ora_ritiro_a ? `${order.ora_ritiro_da ?? ''}${order.ora_ritiro_a ? `–${order.ora_ritiro_a}` : ''}` : undefined;
    const consegnaChip = order.ora_consegna_da || order.ora_consegna_a ? `${order.ora_consegna_da ?? ''}${order.ora_consegna_a ? `–${order.ora_consegna_a}` : ''}` : undefined;
    const stops: ItineraryStop[] = [];
    if (order.garage) stops.push({ variant: 'garage', nome: order.garage.nome, lat: order.garage.lat, lng: order.garage.lng });
    stops.push({ variant: 'carico', nome: carico?.nome, sub: order.data_ritiro, chip: ritiroChip, lat: carico?.lat, lng: carico?.lng });
    stops.push({ variant: 'scarico', nome: scarico?.nome, sub: order.data_consegna, chip: consegnaChip, lat: scarico?.lat, lng: scarico?.lng });
    if (order.wash_station) stops.push({ variant: 'wash', nome: order.wash_station.nome, lat: order.wash_station.lat, lng: order.wash_station.lng });
    return stops;
  }, [order, carico, scarico]);

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

      <Card className="rounded-xl border shadow-sm p-5 space-y-4" data-testid="order-detail-assign-form">
        <div className="flex items-start justify-between gap-4 flex-wrap">
          <div className="min-w-0">
            <div className="flex items-center gap-2 flex-wrap">
              <h2 className="text-xl font-bold" style={{ fontFamily: "'Space Grotesk', sans-serif" }}>{order.cliente?.ragione_sociale}</h2>
              <StatusBadge stato={order.stato} />
            </div>
            <p className="text-sm text-muted-foreground mt-1">
              {order.destinazione_carico?.nome} → {order.destinazione_scarico?.nome}
              {order.rif_ordine_cliente && <> · Rif. cliente: {order.rif_ordine_cliente}</>}
            </p>
            {order.tariffa != null && (
              <div className="mt-3">
                <p className="text-[10px] uppercase tracking-wide text-muted-foreground font-semibold">Tariffa</p>
                <p className="text-xl font-bold">€ {order.tariffa.toLocaleString('it-IT')}</p>
              </div>
            )}
          </div>
          <div className="flex flex-col items-end gap-2 shrink-0">
            {(order.tipologia || order.categoria_trasporto) && (
              <div className="flex items-center gap-1.5 flex-wrap justify-end">
                {order.tipologia && <span className="text-[11px] font-semibold border rounded-full bg-muted px-3 py-1.5 whitespace-nowrap">{order.tipologia}</span>}
                {order.categoria_trasporto && <span className="text-[11px] font-semibold border rounded-full bg-muted px-3 py-1.5 whitespace-nowrap">{order.categoria_trasporto}</span>}
              </div>
            )}
            {order.cliente_id && (
              <Button variant="outline" size="sm" className="text-xs" onClick={() => navigate(`/anagrafiche/clienti/${order.cliente_id}/cruscotto`)}>
                Vai al cruscotto cliente →
              </Button>
            )}
          </div>
        </div>

        {order.stato === 'PIANIFICABILE' ? (
          <AssignOrderForm order={order} onAssigned={fetchOrder} />
        ) : (
          <>
            <OrderRouteMap carico={carico} scarico={scarico} garage={order.garage} washStation={order.wash_station} routePoints={order.route?.points as [number, number][] | undefined} height={280} />
            <RouteItinerary stops={fullItineraryStops} />
          </>
        )}

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

      {order.stato !== 'PIANIFICABILE' && (order.autista || order.vettore) && (
        <Card className="rounded-xl border shadow-sm p-5" data-testid="order-detail-transport-summary">
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
        </Card>
      )}
    </div>
  );
}
