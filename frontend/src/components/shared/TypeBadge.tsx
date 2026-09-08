import { Badge } from '@/components/ui/badge';

// Tag "Tipo": pill PIENA con testo bianco — è il contrario dello Stato, che
// nel design è tenue (vedi StatusBadge.tsx), e sono i due pieni/tenui a far
// distinguere le due colonne a colpo d'occhio.
//
// Il colore cambia col tipo, come in ui/tms-unificato.html:
//   tipoFill = bg => 'color:#fff;background:' + bg + ';border:1px solid ' + bg
//   { 'Export': tipoFill('var(--acc)'), 'Nazionale': tipoFill('#5b6b86'),
//     'Import': tipoFill('#22375a') }
//
// Export usa l'accento, quindi qui `hsl(var(--primary))`: così segue da sé la
// palette (#2a6fdb nel tema base, #3565af in Glass) invece di fissare un hex.
// `solo_estero` e `internazionale` non sono nel design: prendono il navy scuro
// di Import, con cui condividono il senso (fuori confine).
const TIPO_CLASS: Record<string, string> = {
  nazionale: 'tipo-nazionale',
  import: 'tipo-estero',
  solo_estero: 'tipo-estero',
  internazionale: 'tipo-estero',
  export: 'tipo-export',
};

const LABELS: Record<string, string> = {
  nazionale: 'Nazionale',
  internazionale: 'Internazionale',
  import: 'Import',
  export: 'Export',
  solo_estero: 'Solo estero',
};

export interface TypeBadgeProps {
  tipologia?: string;
}

export const TypeBadge = ({ tipologia }: TypeBadgeProps) => {
  if (!tipologia) return <span className="text-muted-foreground">—</span>;
  return (
    <Badge
      variant="outline"
      className={`${TIPO_CLASS[tipologia] ?? 'tipo-nazionale'} border rounded-full text-[10px] font-semibold px-[9px] py-[3px] whitespace-nowrap`}
      data-testid="order-type-badge"
    >
      {LABELS[tipologia] ?? tipologia}
    </Badge>
  );
};
