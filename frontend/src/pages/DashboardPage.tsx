import { useMemo, type ReactNode } from 'react';
import { useNavigate } from 'react-router-dom';
import { formatEuro, formatDayMonth } from '@/lib/format';
import { PageHeaderActions, PageHeaderMeta } from '@/components/layout/PageHeaderActions';
import { Button } from '@/components/ui/button';
import {
  useGetDashboardStatsQuery,
  useGetOrdersListQuery,
  useGetTripsQuery,
  useGetInboundOrdersQuery,
} from '@/store/api/appApi';
import { logger } from '@/lib/logger';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Skeleton } from '@/components/ui/skeleton';
import { TrendingUp, ArrowRight, Upload } from 'lucide-react';
import { BarChart, Bar, XAxis, YAxis, Tooltip, ResponsiveContainer, Cell } from 'recharts';
import { StatusBadge } from '@/components/shared/StatusBadge';
import type { DtoOrderResponse, DtoTripResponse } from '@/api/data-contracts';

interface KPICardProps {
  title: string;
  value: ReactNode;
  description?: ReactNode;
  testId?: string;
  onClick?: () => void;
}

// Nessuna icona: in queste card il soggetto è il numero, e l'icona in cornice
// a destra gli competeva accanto senza aggiungere informazione (rilievo 09
// dell'audit sul tema Glass — nel prototipo è stata rimossa).
const KPICard = ({ title, value, description, testId, onClick }: KPICardProps) => (
  <Card
    data-testid={testId || 'kpi-card'}
    onClick={onClick}
    className={`shadow-sm ${onClick ? 'cursor-pointer transition-shadow hover:shadow-md' : ''}`}
  >
    <CardContent className="p-4 lg:p-5">
      <p className="text-xs font-medium text-muted-foreground mb-1">{title}</p>
      <p className="font-display text-2xl md:text-3xl font-bold tracking-tight">{value}</p>
      {description && <p className="text-xs mt-1">{description}</p>}
    </CardContent>
  </Card>
);

// Badge numerico "a pillola" per le intestazioni Da approvare/Da pianificare
// (rosso/ambra pieno, come nel mockup TMS Unificato).
const CountPill = ({ n, tone }: { n: number; tone: 'red' | 'amber' }) => (
  <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-[10px] font-bold border ${
    tone === 'red'
      ? 'bg-rose-100 text-rose-800 border-rose-200 dark:bg-rose-500/15 dark:text-rose-300 dark:border-rose-500/30'
      : 'bg-amber-100 text-amber-800 border-amber-200 dark:bg-amber-500/15 dark:text-amber-300 dark:border-amber-500/30'
  }`}>
    {n}
  </span>
);

const HeaderLink = ({ label, onClick }: { label: string; onClick: () => void }) => (
  <button type="button" onClick={onClick} className="ml-auto text-xs font-semibold text-primary hover:underline flex items-center gap-1 shrink-0">
    {label} <ArrowRight className="h-3 w-3" />
  </button>
);

// Data corrente per la fascia di testa ("Lunedì 8 settembre 2026"). Calcolata
// al primo render: la pagina non resta aperta a cavallo della mezzanotte in un
// uso normale, e un timer solo per questo non vale.
const todayLabel = (() => {
  const label = new Date().toLocaleDateString('it-IT', {
    weekday: 'long', day: 'numeric', month: 'long', year: 'numeric',
  });
  return label.charAt(0).toUpperCase() + label.slice(1);
})();

const toIso = (d: Date) => d.toISOString().slice(0, 10);
const addDays = (d: Date, n: number) => { const r = new Date(d); r.setDate(r.getDate() + n); return r; };
const capitalize = (s: string) => s.charAt(0).toUpperCase() + s.slice(1);
const monthName = (d: Date) => d.toLocaleDateString('it-IT', { month: 'long' });

// Variazione % vs il mese precedente — mostra "n.d." quando il mese
// precedente non ha dati (divisione per zero, non "+Infinity%").
function pctVsPrevMonth(curr: number, prev: number, prevLabel: string): { text: string; positive: boolean } {
  if (prev === 0) return { text: curr > 0 ? `n.d. su ${prevLabel}` : `invariato su ${prevLabel}`, positive: curr >= 0 };
  const pct = Math.round(((curr - prev) / prev) * 100);
  return { text: `${pct > 0 ? '+' : ''}${pct}% su ${prevLabel}`, positive: pct >= 0 };
}

// Origine → destinazione di un viaggio: primo e ultimo segmento in ordine
// (percorso stradale reale, richiede ORS_API_KEY). Se i segmenti non sono
// disponibili (es. locale senza chiave ORS: il routing degrada a nil, vedi
// docker-compose.yml) si ricava comunque una tratta indicativa dagli ordini
// del viaggio: carico del primo, scarico dell'ultimo.
function tripTratta(t: DtoTripResponse, orderById: Map<string, DtoOrderResponse>): string {
  const segs = (t.segmenti ?? []).slice().sort((a, b) => (a.ordine ?? 0) - (b.ordine ?? 0));
  if (segs.length > 0) {
    return `${segs[0].origine_nome || '?'} → ${segs[segs.length - 1].destinazione_nome || '?'}`;
  }
  const ids = t.ordini_ids ?? [];
  if (ids.length === 0) return '—';
  const first = orderById.get(ids[0]);
  const last = orderById.get(ids[ids.length - 1]);
  const origine = first?.destinazione_carico?.nome;
  const destinazione = last?.destinazione_scarico?.nome;
  if (!origine && !destinazione) return '—';
  return `${origine || '?'} → ${destinazione || '?'}`;
}

function tripMezzo(t: DtoTripResponse): string {
  if (t.motrice?.targa) return t.motrice.targa;
  if (t.vettore?.ragione_sociale) return t.vettore.ragione_sociale;
  return '—';
}

function tripAutista(t: DtoTripResponse): string {
  if (!t.autista) return '—';
  return `${t.autista.nome || ''} ${t.autista.cognome || ''}`.trim() || '—';
}

// Percentuale di avanzamento stimata da partenza/arrivo pianificati (non da
// una posizione GPS reale, che non abbiamo) — solo per i viaggi IN_CORSO.
function tripProgressPct(t: DtoTripResponse): number | null {
  if (!t.data_partenza || !t.data_arrivo) return null;
  const start = new Date(`${t.data_partenza}T00:00:00`).getTime();
  const end = new Date(`${t.data_arrivo}T00:00:00`).getTime();
  if (Number.isNaN(start) || Number.isNaN(end) || end <= start) return null;
  const now = Date.now();
  return Math.max(0, Math.min(100, Math.round(((now - start) / (end - start)) * 100)));
}

// Alert flotta: 5 contatori derivati da viaggi + totali attivi di stats
// (replica il pannello "Flotta" del mockup TMS Unificato, #dashboard).
interface FleetAlert { color: string; label: string; value: number; blink?: boolean }

export default function DashboardPage() {
  const navigate = useNavigate();
  const today = useMemo(() => new Date(), []);
  const todayIso = toIso(today);
  const soonIso = toIso(addDays(today, 2));
  // 8 settimane di storico per il grafico "Andamento ordini" (come il mockup).
  const eightWeeksAgoIso = toIso(addDays(today, -55));

  const statsQuery = useGetDashboardStatsQuery();
  const daPianificareQuery = useGetOrdersListQuery({ stato: 'PIANIFICABILE' });
  const weeklyQuery = useGetOrdersListQuery({ data_da: eightWeeksAgoIso, data_a: todayIso });
  const tripsQuery = useGetTripsQuery();
  const inboundQuery = useGetInboundOrdersQuery();
  // Ordini dei viaggi (In corso/Pianificato) — solo per ricavare origine/
  // destinazione in tabella quando i segmenti del viaggio sono vuoti (vedi
  // tripTratta).
  const viaggioOrdersQuery = useGetOrdersListQuery({ stato: 'VIAGGIO' });
  const pianificatoOrdersQuery = useGetOrdersListQuery({ stato: 'PIANIFICATO' });

  const stats = statsQuery.data;
  const loading = statsQuery.isLoading || daPianificareQuery.isLoading
    || weeklyQuery.isLoading || tripsQuery.isLoading || inboundQuery.isLoading
    || viaggioOrdersQuery.isLoading || pianificatoOrdersQuery.isLoading;

  if (statsQuery.error) logger.error('Dashboard stats error:', statsQuery.error);
  if (daPianificareQuery.error) logger.error('Dashboard da-pianificare error:', daPianificareQuery.error);
  if (weeklyQuery.error) logger.error('Dashboard weekly error:', weeklyQuery.error);
  if (tripsQuery.error) logger.error('Dashboard trips error:', tripsQuery.error);
  if (inboundQuery.error) logger.error('Dashboard inbound error:', inboundQuery.error);
  if (viaggioOrdersQuery.error) logger.error('Dashboard viaggio-orders error:', viaggioOrdersQuery.error);
  if (pianificatoOrdersQuery.error) logger.error('Dashboard pianificato-orders error:', pianificatoOrdersQuery.error);

  if (loading) {
    return (
      <div className="space-y-4">
        <div className="grid grid-cols-2 lg:grid-cols-4 gap-3 lg:gap-4">
          {[1, 2, 3, 4].map(i => <Skeleton key={i} className="h-24 rounded-xl" />)}
        </div>
        <Skeleton className="h-64 rounded-xl" />
      </div>
    );
  }

  const daPianificareOrders = daPianificareQuery.data ?? [];
  const lateN = daPianificareOrders.filter((o) => (o.data_ritiro || '9999') < todayIso).length;
  const soonN = daPianificareOrders.filter((o) => {
    const r = o.data_ritiro || '9999';
    return r >= todayIso && r <= soonIso;
  }).length;
  const planSubColor = lateN > 0 ? 'text-red-600 dark:text-red-400' : soonN > 0 ? 'text-amber-600 dark:text-amber-400' : 'text-emerald-600 dark:text-emerald-400';
  const planSub = lateN > 0 || soonN > 0
    ? [lateN > 0 ? `${lateN} in ritardo` : null, soonN > 0 ? `${soonN} entro 48h` : null].filter(Boolean).join(' · ')
    : 'nessun ritiro urgente';

  const trips = tripsQuery.data ?? [];
  const liveTrips = trips.filter((t) => t.stato === 'IN_CORSO');
  const plannedTrips = trips.filter((t) => t.stato === 'PIANIFICATO');
  const mezziLiveSet = new Set(liveTrips.filter((t) => t.motrice_id).map((t) => t.motrice_id));
  const mezziPlanSet = new Set(plannedTrips.filter((t) => t.motrice_id && !mezziLiveSet.has(t.motrice_id)).map((t) => t.motrice_id));
  const nLive = mezziLiveSet.size;
  const nPlan = mezziPlanSet.size;
  const disponibiliGarage = Math.max((stats?.total_motrici || 0) - nLive - nPlan, 0);
  const terziAttiviN = liveTrips.filter((t) => t.vettore_id).length;
  const autistiBusySet = new Set(liveTrips.filter((t) => t.autista_id).map((t) => t.autista_id));
  const autistiDisponibiliN = Math.max((stats?.total_drivers || 0) - autistiBusySet.size, 0);

  const orderById = new Map<string, DtoOrderResponse>();
  for (const o of viaggioOrdersQuery.data ?? []) if (o.id) orderById.set(o.id, o);
  for (const o of pianificatoOrdersQuery.data ?? []) if (o.id) orderById.set(o.id, o);

  const fleetAlerts: FleetAlert[] = [
    { color: '#1D55AD', label: 'Mezzi in viaggio', value: nLive, blink: true },
    { color: '#8A6508', label: 'Su viaggi pianificati', value: nPlan },
    { color: '#1F7A4D', label: 'Disponibili in garage', value: disponibiliGarage },
    { color: '#7C3AED', label: 'Vettori terzi attivi', value: terziAttiviN },
    { color: '#0D9488', label: 'Autisti disponibili', value: autistiDisponibiliN },
  ];

  // 8 bucket settimanali di 7 giorni, l'ultimo termina oggi (evidenziato).
  const weeklyOrders = weeklyQuery.data ?? [];
  const weeklyBars = Array.from({ length: 8 }, (_, i) => {
    const weeksBack = 7 - i;
    const start = addDays(today, -weeksBack * 7 - 6);
    const end = addDays(today, -weeksBack * 7);
    const startIso = toIso(start);
    const endIso = toIso(end);
    return {
      key: startIso,
      label: capitalize(start.toLocaleDateString('it-IT', { day: 'numeric', month: 'short' }).replace('.', '')),
      count: weeklyOrders.filter((o) => (o.data_ritiro || '') >= startIso && (o.data_ritiro || '') <= endIso).length,
    };
  });

  const daApprovareAll = (inboundQuery.data ?? []).filter((o) => o.status === 'pending' && !o.portal);
  const daApprovare = daApprovareAll
    .slice()
    .sort((a, b) => (a.received_at || a.created_at || '').localeCompare(b.received_at || b.created_at || ''))
    .slice(0, 4);

  const daPianificareTop = daPianificareOrders
    .slice()
    .sort((a, b) => (a.data_ritiro || '9999').localeCompare(b.data_ritiro || '9999'))
    .slice(0, 4);

  // Metà e metà (In corso / Pianificato) così, quando entrambi esistono, non
  // capita che i soli viaggi In corso riempiano tutti gli slot e i
  // Pianificati spariscano dalla lista (visti solo aprendo il planner).
  const inCorsoTop = trips.filter((t) => t.stato === 'IN_CORSO').sort((a, b) => (a.id || '').localeCompare(b.id || '')).slice(0, 3);
  const pianificatoTop = trips.filter((t) => t.stato === 'PIANIFICATO').sort((a, b) => (a.id || '').localeCompare(b.id || '')).slice(0, 3);
  const viaggiTop = [...inCorsoTop, ...pianificatoTop];

  const openOrder = (id?: string) => { if (id) navigate(`/planner/ordini/${id}`); };
  const openTrip = (t: DtoTripResponse) => openOrder(t.ordini_ids?.[0]);
  const goPlanQueue = () => navigate('/planner', { state: { statusFilter: 'PIANIFICABILE' } });

  const curMonthLabel = monthName(today);
  const prevMonthLabel = monthName(addDays(new Date(today.getFullYear(), today.getMonth(), 1), -1));
  const ordiniDelta = pctVsPrevMonth(stats?.orders_this_month || 0, stats?.orders_prev_month || 0, prevMonthLabel);
  const fatturatoDelta = pctVsPrevMonth(stats?.revenue_this_month || 0, stats?.revenue_prev_month || 0, prevMonthLabel);

  return (
    <div className="space-y-4 lg:space-y-6" data-testid="dashboard-page">
      {/* Data corrente accanto al titolo: nel prototipo ce l'ha solo la
          Dashboard, le altre pagine hanno il titolo nudo. */}
      <PageHeaderMeta>
        <span className="text-xs text-muted-foreground whitespace-nowrap" data-testid="page-header-date">
          {todayLabel}
        </span>
      </PageHeaderMeta>

      {/* Azione primaria sulla fascia di testa (vedi PageHeaderActions.tsx):
          il caricamento di un ordine da PDF è il punto d'ingresso più usato
          della Dashboard e prima non c'era, si passava dalla sidebar. */}
      <PageHeaderActions>
        <Button size="sm" className="text-xs gap-1.5" onClick={() => navigate('/ordini-in-ingresso')} data-testid="dashboard-upload-pdf">
          <Upload className="h-3.5 w-3.5" /> Carica PDF
        </Button>
      </PageHeaderActions>

      {/* KPI Row */}
      <div className="grid grid-cols-2 lg:grid-cols-4 gap-3 lg:gap-4" data-testid="dashboard-kpi">
        <KPICard
          title={`Ordini · ${curMonthLabel}`}
          value={stats?.orders_this_month || 0}
          description={<span className={ordiniDelta.positive ? 'text-emerald-600 dark:text-emerald-400' : 'text-red-600 dark:text-red-400'}>{ordiniDelta.text}</span>}
          onClick={() => navigate('/ordini')}
          testId="kpi-orders-month"
        />
        <KPICard
          title="Viaggi in corso"
          value={liveTrips.length}
          description={<span className="text-muted-foreground">{nLive} mezzi propri · {terziAttiviN} {terziAttiviN === 1 ? 'vettore terzo' : 'vettori terzi'}</span>}
          onClick={() => navigate('/planner')}
          testId="kpi-viaggi-in-corso"
        />
        <KPICard
          title="Da pianificare"
          value={stats?.pianificabili || 0}
          description={<span className={planSubColor}>{planSub}</span>}
          onClick={goPlanQueue}
          testId="kpi-da-pianificare"
        />
        <KPICard
          title={`Fatturato · ${curMonthLabel}`}
          value={`€ ${formatEuro(stats?.revenue_this_month || 0)}`}
          description={<span className={fatturatoDelta.positive ? 'text-emerald-600 dark:text-emerald-400' : 'text-red-600 dark:text-red-400'}>{fatturatoDelta.text}</span>}
          onClick={() => navigate('/ordini')}
          testId="kpi-fatturato-month"
        />
      </div>

      <div className="grid grid-cols-1 xl:grid-cols-[1.6fr_1fr] gap-4">
        {/* Andamento ordini — 8 settimane */}
        <Card className="shadow-sm">
          <CardHeader className="pb-2">
            <div className="flex items-baseline gap-2">
              <CardTitle className="font-display text-base flex items-center gap-2">
                <TrendingUp className="h-4 w-4" /> Andamento ordini
              </CardTitle>
              <span className="text-xs text-muted-foreground">ultime 8 settimane</span>
              <span className="ml-auto inline-flex items-center gap-1.5 text-[11px] text-muted-foreground">
                <span className="h-2 w-2 rounded-sm bg-primary" /> settimana corrente
              </span>
            </div>
          </CardHeader>
          <CardContent>
            <div className="h-56">
              <ResponsiveContainer width="100%" height="100%">
                <BarChart data={weeklyBars}>
                  <XAxis dataKey="label" tick={{ fontSize: 11 }} stroke="hsl(var(--muted-foreground))" />
                  <YAxis tick={{ fontSize: 11 }} stroke="hsl(var(--muted-foreground))" allowDecimals={false} />
                  <Tooltip
                    contentStyle={{ borderRadius: 8, fontSize: 12, border: '1px solid hsl(var(--border))' }}
                    formatter={(val: number) => [val, 'Ordini']}
                  />
                  <Bar dataKey="count" radius={[4, 4, 0, 0]}>
                    {weeklyBars.map((b, i) => (
                      <Cell key={b.key} fill={i === weeklyBars.length - 1 ? 'hsl(var(--primary))' : 'hsl(var(--accent))'} />
                    ))}
                  </Bar>
                </BarChart>
              </ResponsiveContainer>
            </div>
          </CardContent>
        </Card>

        {/* Flotta */}
        <Card className="shadow-sm">
          <CardHeader className="pb-2">
            <div className="flex items-baseline gap-2">
              <CardTitle className="font-display text-base">Flotta</CardTitle>
              <span className="ml-auto text-xs text-muted-foreground">{stats?.total_motrici || 0} motrici · {stats?.total_drivers || 0} autisti</span>
            </div>
          </CardHeader>
          <CardContent>
            <div className="flex flex-col">
              {fleetAlerts.map((a) => (
                <div key={a.label} className="flex items-center gap-2.5 py-2 border-b last:border-b-0 text-sm">
                  <span
                    className={`h-2 w-2 rounded-full shrink-0 ${a.blink ? 'animate-pulse' : ''}`}
                    style={{ backgroundColor: a.color }}
                  />
                  <span className="flex-1 text-muted-foreground">{a.label}</span>
                  <span className="font-semibold tabular-nums">{a.value}</span>
                </div>
              ))}
            </div>
            <button type="button" onClick={() => navigate('/mappa')} className="mt-2.5 text-xs font-semibold text-primary hover:underline flex items-center gap-1">
              Vedi sulla mappa <ArrowRight className="h-3 w-3" />
            </button>
          </CardContent>
        </Card>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        {/* Da approvare */}
        <Card className="shadow-sm">
          <CardHeader className="pb-2 flex flex-row items-center gap-2">
            <CardTitle className="font-display text-base">Da approvare</CardTitle>
            <CountPill n={daApprovareAll.length} tone="red" />
            <HeaderLink label="In arrivo" onClick={() => navigate('/ordini-in-ingresso')} />
          </CardHeader>
          <CardContent className="space-y-1" data-testid="dashboard-da-approvare">
            {daApprovare.length === 0 && <p className="text-sm text-muted-foreground py-2">Nessun ordine in attesa di approvazione ✓</p>}
            {daApprovare.map((o) => (
              <button
                type="button"
                key={o.id}
                onClick={() => navigate('/ordini-in-ingresso')}
                className="w-full flex items-center justify-between gap-2 rounded-md px-2 py-2 -mx-2 text-left hover:bg-muted/60"
              >
                <div className="min-w-0">
                  <p className="text-sm font-medium truncate">{o.client || '—'}</p>
                  <p className="text-xs text-muted-foreground truncate">{o.load_place || '?'} → {o.delivery_place || '?'}</p>
                </div>
                <span className="text-xs text-muted-foreground whitespace-nowrap tabular-nums">{formatDayMonth(o.load_date)}</span>
              </button>
            ))}
          </CardContent>
        </Card>

        {/* Da pianificare */}
        <Card className="shadow-sm">
          <CardHeader className="pb-2 flex flex-row items-center gap-2">
            <CardTitle className="font-display text-base">Da pianificare</CardTitle>
            <CountPill n={stats?.pianificabili || 0} tone="amber" />
            <HeaderLink label="Registro" onClick={goPlanQueue} />
          </CardHeader>
          <CardContent className="space-y-1" data-testid="dashboard-da-pianificare">
            {daPianificareTop.length === 0 && <p className="text-sm text-muted-foreground py-2">Niente da pianificare ✓</p>}
            {daPianificareTop.map((o) => {
              const r = o.data_ritiro || '9999';
              const late = r < todayIso;
              const soon = !late && r <= soonIso;
              return (
                <button
                  type="button"
                  key={o.id}
                  onClick={() => openOrder(o.id)}
                  className="w-full flex items-center justify-between gap-2 rounded-md px-2 py-2 -mx-2 text-left hover:bg-muted/60"
                >
                  <div className="min-w-0">
                    <p className="text-sm font-medium truncate">{o.cliente?.ragione_sociale || '—'}</p>
                    <p className="text-xs text-muted-foreground truncate">{o.destinazione_carico?.nome || '?'} → {o.destinazione_scarico?.nome || '?'}</p>
                  </div>
                  <div className="text-right shrink-0">
                    <p className={`text-xs font-semibold whitespace-nowrap ${late ? 'text-red-600 dark:text-red-400' : soon ? 'text-amber-600 dark:text-amber-400' : ''}`}>ritiro <span className="tabular-nums">{formatDayMonth(o.data_ritiro)}</span></p>
                    {(late || soon) && <p className={`text-[10px] whitespace-nowrap ${late ? 'text-red-600 dark:text-red-400' : 'text-amber-600 dark:text-amber-400'}`}>{late ? 'ritiro in ritardo' : 'entro 48h'}</p>}
                  </div>
                </button>
              );
            })}
          </CardContent>
        </Card>
      </div>

      {/* Viaggi in corso e pianificati */}
      <Card className="shadow-sm">
        <CardHeader className="pb-2 flex flex-row items-center gap-2">
          <CardTitle className="font-display text-base">Viaggi in corso e pianificati</CardTitle>
          <HeaderLink label="Apri planner" onClick={() => navigate('/planner')} />
        </CardHeader>
        <CardContent className="p-0" data-testid="dashboard-viaggi">
          {viaggiTop.length === 0 && <p className="text-sm text-muted-foreground px-4 py-3">Nessun viaggio in corso o pianificato</p>}
          {viaggiTop.length > 0 && (
            <div className="divide-y">
              {viaggiTop.map((t) => {
                const live = t.stato === 'IN_CORSO';
                const progress = live ? tripProgressPct(t) : null;
                return (
                  <button
                    type="button"
                    key={t.id}
                    onClick={() => openTrip(t)}
                    className="w-full flex items-center gap-4 px-4 py-2.5 text-left hover:bg-muted/60"
                  >
                    <StatusBadge stato={t.stato} pulse={live} />
                    <span className="text-sm w-36 shrink-0 truncate">{tripAutista(t)}</span>
                    <span className="text-xs text-muted-foreground w-24 shrink-0 truncate">{tripMezzo(t)}</span>
                    <span className="text-sm flex-1 truncate">{tripTratta(t, orderById)}</span>
                    {live ? (
                      <span className="text-xs text-muted-foreground whitespace-nowrap w-20 shrink-0">{t.data_arrivo ? `ETA ${formatDayMonth(t.data_arrivo)}` : ''}</span>
                    ) : (
                      <span className="text-xs text-muted-foreground whitespace-nowrap w-20 shrink-0">pianificato</span>
                    )}
                    {live && progress != null ? (
                      <div className="w-20 h-1 rounded-full bg-muted overflow-hidden shrink-0">
                        <div className="h-full rounded-full bg-emerald-600" style={{ width: `${progress}%` }} />
                      </div>
                    ) : (
                      <div className="w-20 shrink-0" />
                    )}
                  </button>
                );
              })}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
