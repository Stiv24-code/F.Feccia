import { Badge } from '@/components/ui/badge';

// Pill di stato come nel design (ui/tms-unificato.html, mappa `badge`):
// fondo tenue + bordo + testo nel colore semantico, non fondo pieno.
//
//   plan    → --neg  #c0392b su --neg-bg  #fdecec, bordo --neg-bd  #f0c5c0
//   planned → --warn #8a6508 su --warn-bg #fdf1c7, bordo --warn-bd #eedda0
//   trip    → --acc-2 #1d55ad su --acc-bg #e4eefc, bordo --acc-bd #bfd3f2
//   closed  → --pos  #1f7a4d su --pos-bg  #e9f5ee, bordo --pos-bd  #bfe0cd
//
// Sono esattamente i valori delle classi `.status-order-*` in index.css, che
// erano già allineati: il passaggio al pieno era una lettura sbagliata del
// rilievo 05 dell'audit ("Stato con il colore semantico pieno"), smentita dal
// design. Il pieno resta solo sul tag Tipo (vedi TypeBadge.tsx), ed è quello
// che distingue le due colonne a colpo d'occhio.
const statusConfig: Record<string, { label: string; className: string; dot: string }> = {
  PIANIFICABILE: { label: 'Da pianificare', className: 'status-order-red', dot: '#C0392B' },
  PIANIFICATO: { label: 'Pianificato', className: 'status-order-yellow', dot: '#8A6508' },
  VIAGGIO: { label: 'In viaggio', className: 'status-order-blue', dot: '#1D55AD' },
  CHIUSO: { label: 'Consegnato', className: 'status-order-green', dot: '#1F7A4D' },
  SCARTATO: { label: 'Scartato', className: 'status-order-gray', dot: '#7C879A' },
  PROFORMA: { label: 'Proforma', className: 'status-pianificabile', dot: '#8A6508' },
  DEFINITIVA: { label: 'Definitiva', className: 'status-fatturato', dot: '#1F7A4D' },
  IN_CORSO: { label: 'In corso', className: 'status-order-blue', dot: '#1D55AD' },
  COMPLETATO: { label: 'Completato', className: 'status-fatturato', dot: '#1F7A4D' },
};

export interface StatusBadgeProps {
  stato?: string;
  // Punto lampeggiante (replica lo stile "live" del mockup, es. viaggi
  // IN_CORSO in dashboard) — opt-in, default spento: nel design le pill in
  // tabella non hanno pallino, ce l'ha solo l'indicatore "in diretta".
  pulse?: boolean;
}

export const StatusBadge = ({ stato, pulse }: StatusBadgeProps) => {
  const config = (stato && statusConfig[stato]) || { label: stato, className: 'status-order-gray', dot: '#7C879A' };
  return (
    <Badge
      variant="outline"
      className={`${config.className} border rounded-full text-[10px] font-semibold px-[9px] py-[3px] gap-1.5 whitespace-nowrap`}
      data-testid="order-status-badge"
    >
      {pulse && <span className="h-1.5 w-1.5 rounded-full animate-pulse" style={{ backgroundColor: config.dot }} />}
      {config.label}
    </Badge>
  );
};
