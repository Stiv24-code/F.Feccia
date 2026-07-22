import { useMemo, useState } from 'react';
import { addDays, addWeeks, subWeeks, format, isSameDay, isToday, isValid, parseISO, startOfWeek } from 'date-fns';
import { it } from 'date-fns/locale';
import { Button } from '@/components/ui/button';
import { Card } from '@/components/ui/card';
import { ChevronLeft, ChevronRight, Truck } from 'lucide-react';

// Colore della card: diretto dallo stato reale dell'ordine (PIANIFICABILE/
// PIANIFICATO/VIAGGIO/CHIUSO/SCARTATO) — nessuna euristica sulla data,
// stessa mappatura di StatusBadge (rosso/giallo/blu/verde/grigio).
const COLOR_BY_STATO = {
  PIANIFICABILE: 'red',
  PIANIFICATO: 'yellow',
  VIAGGIO: 'blue',
  CHIUSO: 'green',
  SCARTATO: 'gray',
};

const COLOR_STYLES = {
  red: { card: 'bg-red-50 border-red-300 text-red-900', dot: 'bg-red-500' },
  yellow: { card: 'bg-amber-50 border-amber-300 text-amber-900', dot: 'bg-amber-500' },
  blue: { card: 'bg-blue-50 border-blue-300 text-blue-900', dot: 'bg-blue-500' },
  green: { card: 'bg-emerald-50 border-emerald-300 text-emerald-900', dot: 'bg-emerald-500' },
  gray: { card: 'bg-muted/40 border-muted-foreground/20 text-muted-foreground opacity-75', dot: 'bg-muted-foreground/50' },
};

const LEGEND = [
  { color: 'red', label: 'Da pianificare' },
  { color: 'yellow', label: 'Pianificato' },
  { color: 'blue', label: 'In viaggio' },
  { color: 'green', label: 'Consegnato' },
  { color: 'gray', label: 'Scartato' },
];

export default function PlannerCalendar({ orders, onAssign, onStart, onClose }) {
  const [weekStart, setWeekStart] = useState(() => startOfWeek(new Date(), { weekStartsOn: 1 }));

  const days = useMemo(() => Array.from({ length: 7 }, (_, i) => addDays(weekStart, i)), [weekStart]);

  const ordersByDay = useMemo(() => {
    const buckets = days.map(() => []);
    (orders || []).forEach(o => {
      if (!o.data_ritiro) return;
      const d = parseISO(o.data_ritiro);
      if (!isValid(d)) return;
      const idx = days.findIndex(day => isSameDay(day, d));
      if (idx >= 0) buckets[idx].push(o);
    });
    buckets.forEach(list => list.sort((a, b) => (a.ora_ritiro_da || '').localeCompare(b.ora_ritiro_da || '')));
    return buckets;
  }, [orders, days]);

  const weekLabel = useMemo(() => {
    const end = days[6];
    const sameMonth = weekStart.getMonth() === end.getMonth();
    const from = format(weekStart, sameMonth ? 'd' : 'd MMM', { locale: it });
    const to = format(end, 'd MMM yyyy', { locale: it });
    return `${from} – ${to}`;
  }, [days, weekStart]);

  const handleCardClick = (order) => {
    if (order.stato === 'PIANIFICABILE') onAssign(order);
    else if (order.stato === 'PIANIFICATO' && !order.viaggio_id) onStart(order);
    else if (order.stato === 'VIAGGIO') onClose(order);
  };

  return (
    <Card className="rounded-xl border shadow-sm">
      <div className="flex flex-wrap items-center justify-between gap-2 px-4 py-2.5 border-b bg-muted/30">
        <div className="flex items-center gap-2">
          <Button variant="outline" size="sm" className="h-7 text-xs" onClick={() => setWeekStart(startOfWeek(new Date(), { weekStartsOn: 1 }))} data-testid="calendar-today-button">
            Oggi
          </Button>
          <Button variant="ghost" size="icon" className="h-7 w-7" onClick={() => setWeekStart(w => subWeeks(w, 1))} data-testid="calendar-prev-week">
            <ChevronLeft className="h-4 w-4" />
          </Button>
          <Button variant="ghost" size="icon" className="h-7 w-7" onClick={() => setWeekStart(w => addWeeks(w, 1))} data-testid="calendar-next-week">
            <ChevronRight className="h-4 w-4" />
          </Button>
          <span className="text-sm font-semibold capitalize">{weekLabel}</span>
        </div>
        <div className="flex items-center gap-3 flex-wrap">
          {LEGEND.map(l => (
            <span key={l.color} className="flex items-center gap-1.5 text-[11px] text-muted-foreground">
              <span className={`h-2 w-2 rounded-full ${COLOR_STYLES[l.color].dot}`} />
              {l.label}
            </span>
          ))}
        </div>
      </div>

      <div className="grid grid-cols-7 divide-x">
        {days.map(day => (
          <div key={day.toISOString()} className={`px-2 py-1.5 text-center text-xs font-semibold uppercase tracking-wide border-b ${isToday(day) ? 'bg-primary/5 text-primary' : 'text-muted-foreground'}`}>
            {format(day, 'EEE d MMM', { locale: it })}
          </div>
        ))}
      </div>

      <div className="grid grid-cols-7 divide-x min-h-[320px]">
        {ordersByDay.map((list, i) => (
          <div key={days[i].toISOString()} className={`flex flex-col gap-1.5 p-1.5 ${isToday(days[i]) ? 'bg-primary/5' : ''}`}>
            {list.length === 0 && (
              <span className="text-[11px] text-muted-foreground/50 text-center pt-2">—</span>
            )}
            {list.map(o => {
              const color = COLOR_BY_STATO[o.stato] || 'gray';
              const style = COLOR_STYLES[color];
              return (
                <button
                  key={o.id}
                  type="button"
                  onClick={() => handleCardClick(o)}
                  title={`${o.progressivo || ''} · ${o.cliente_nome || ''} · ${o.autista_nome || 'nessun autista'}`}
                  className={`text-left rounded-md border px-2 py-1.5 text-[11px] leading-tight shadow-sm hover:brightness-95 transition ${style.card}`}
                  data-testid="calendar-order-card"
                >
                  <div className="flex items-center justify-between gap-1">
                    <span className="font-semibold truncate">{o.cliente_nome || o.progressivo}</span>
                    <span className={`h-1.5 w-1.5 rounded-full shrink-0 ${style.dot}`} />
                  </div>
                  <div className="truncate opacity-80">
                    {o.destinazione_carico_nome} → {o.destinazione_scarico_nome}
                  </div>
                  <div className="flex items-center justify-between mt-0.5 opacity-70">
                    <span>{o.ora_ritiro_da || '—'}</span>
                    {o.autista_nome && (
                      <span className="flex items-center gap-0.5 truncate max-w-[70%]">
                        <Truck className="h-2.5 w-2.5 shrink-0" />{o.targa_motrice || o.autista_nome}
                      </span>
                    )}
                  </div>
                </button>
              );
            })}
          </div>
        ))}
      </div>
    </Card>
  );
}
