import { Badge } from '@/components/ui/badge';

// Tag "Tipo" (tipologia ordine): pill piena su fondo navy con testo bianco.
// In outline, e con il valore in minuscolo così come arriva dall'API, a colpo
// d'occhio non si distingueva dalla colonna Stato accanto — sono due
// informazioni diverse e devono leggersi come tali.
//
// Il colore sta in `.tag-navy` (index.css, token --tag-navy) perché cambia in
// tema scuro: il navy del tema chiaro non si staccherebbe dalla card.
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
      className="tag-navy border text-[10px] px-2 py-0.5 font-medium"
      data-testid="order-type-badge"
    >
      {LABELS[tipologia] ?? tipologia}
    </Badge>
  );
};
