import { Fragment } from 'react';
import { Warehouse, Droplets } from 'lucide-react';
import { haversineKm } from '@/lib/geo';
import { cn } from '@/lib/utils';

export interface ItineraryStop {
  variant: 'garage' | 'carico' | 'scarico' | 'wash';
  nome?: string;
  sub?: string;
  chip?: string;
  lat?: number | null;
  lng?: number | null;
}

const WASH_COLOR = '#0d9488';

const StopMarker = ({ variant }: { variant: ItineraryStop['variant'] }) => {
  if (variant === 'garage') {
    return (
      <div className="flex h-[15px] w-[15px] shrink-0 items-center justify-center rounded-full bg-foreground text-background">
        <Warehouse className="h-2.5 w-2.5" />
      </div>
    );
  }
  if (variant === 'wash') {
    return (
      <div className="flex h-[15px] w-[15px] shrink-0 items-center justify-center rounded-full border-2 bg-background" style={{ borderColor: WASH_COLOR, color: WASH_COLOR }}>
        <Droplets className="h-2.5 w-2.5" />
      </div>
    );
  }
  if (variant === 'carico') {
    return <div className="h-[13px] w-[13px] shrink-0 rounded-full bg-primary" />;
  }
  return <div className="h-[13px] w-[13px] shrink-0 rounded-full border-[3px] border-primary bg-background" />;
};

// Ogni tratta ha un'enfasi visiva diversa (replica il mockup "TMS Unificato"):
// garage→carico è un avvicinamento non fatturato (grigio, tratta secondaria),
// carico→scarico è il trasporto vero e proprio (blu primary, tratta
// principale), scarico→lavaggio è un giro accessorio dopo la consegna (teal
// tratteggiato). Nessuna delle due tratte laterali è sempre presente.
const segmentFor = (a: ItineraryStop['variant'], b: ItineraryStop['variant']) => {
  if (a === 'carico' && b === 'scarico') {
    return { flexClass: 'flex-1', lineClassName: 'bg-primary', lineStyle: undefined, labelClassName: 'text-primary' };
  }
  if (a === 'scarico' && b === 'wash') {
    return { flexClass: 'flex-[0.6]', lineClassName: 'border-t-2 border-dashed', lineStyle: { borderColor: WASH_COLOR }, labelClassName: '', labelStyle: { color: WASH_COLOR } };
  }
  return { flexClass: 'flex-[0.6]', lineClassName: 'bg-border', lineStyle: undefined, labelClassName: 'text-muted-foreground' };
};

const LABELS: Record<ItineraryStop['variant'], string> = {
  garage: 'PARTENZA',
  carico: '↑ CARICO',
  scarico: '↓ SCARICO',
  wash: 'LAVAGGIO',
};

const hasCoords = (s: ItineraryStop) => s.lat != null && s.lng != null;

export interface RouteItineraryProps {
  stops: ItineraryStop[];
}

// Timeline orizzontale partenza→carico→scarico→lavaggio (le tappe garage/
// lavaggio sono opzionali — il chiamante passa solo quelle note). Il km di
// ogni tratta è una distanza in linea d'aria fra le coordinate delle due
// tappe, non un percorso stradale — coerente con l'uso di haversineKm già
// fatto altrove in questa pagina/form quando non c'è un routing calcolato.
export default function RouteItinerary({ stops }: RouteItineraryProps) {
  if (stops.length < 2) return null;

  return (
    <div>
      <p className="text-[10px] uppercase tracking-wide text-muted-foreground font-semibold mb-4">Itinerario</p>
      <div className="flex items-start gap-2">
        {stops.map((stop, i) => {
          const isFirst = i === 0;
          const isLast = i === stops.length - 1;
          const next = stops[i + 1];
          const km = next && hasCoords(stop) && hasCoords(next) ? haversineKm(stop, next) : null;
          const seg = next ? segmentFor(stop.variant, next.variant) : null;
          return (
            <Fragment key={`${stop.variant}-${i}`}>
              <div className={cn(
                'flex flex-col gap-1 w-32 shrink-0',
                isFirst && 'items-start text-left',
                isLast && 'items-end text-right',
                !isFirst && !isLast && 'items-center text-center',
              )}>
                {(stop.variant === 'garage' || stop.variant === 'wash') && <StopMarker variant={stop.variant} />}
                <span
                  className={cn('text-[10px] font-bold', stop.variant === 'carico' || stop.variant === 'scarico' ? 'text-primary' : 'text-muted-foreground')}
                  style={stop.variant === 'wash' ? { color: WASH_COLOR } : undefined}
                >
                  {LABELS[stop.variant]}
                </span>
                <span className="text-sm font-bold leading-tight">{stop.nome}</span>
                {stop.sub && <span className="text-xs text-muted-foreground">{stop.sub}</span>}
                {stop.chip && (
                  <span className="inline-flex w-fit items-center text-[11px] font-medium border rounded px-1.5 py-0.5 bg-muted/50">
                    {stop.chip}
                  </span>
                )}
              </div>
              {seg && (
                <div className={cn('relative mt-[7px] h-0.5', seg.flexClass, seg.lineClassName)} style={seg.lineStyle}>
                  {km != null && (
                    <span
                      className={cn('absolute -top-5 left-1/2 -translate-x-1/2 whitespace-nowrap text-xs font-bold', seg.labelClassName)}
                      style={seg.labelStyle}
                    >
                      {km} km
                    </span>
                  )}
                </div>
              )}
            </Fragment>
          );
        })}
      </div>
    </div>
  );
}
