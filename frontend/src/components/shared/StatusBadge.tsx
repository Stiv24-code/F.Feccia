import { Badge } from '@/components/ui/badge';

// Ordine (PIANIFICABILE/PIANIFICATO/VIAGGIO/CHIUSO/SCARTATO) e Viaggio
// (IN_CORSO) usano classi dedicate (status-order-*, rosso/giallo/blu/verde/
// grigio) per non toccare i colori di Fatture/Proforma/Definitiva, che
// restano sulle classi condivise originali.
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
}

export const StatusBadge = ({ stato }: StatusBadgeProps) => {
  const config = (stato && statusConfig[stato]) || { label: stato, className: '', dot: '#999' };
  return (
    <Badge variant="outline" className={`${config.className} border text-[10px] px-2 py-0.5 font-medium gap-1.5`} data-testid="order-status-badge">
      <span className="h-1.5 w-1.5 rounded-full" style={{ backgroundColor: config.dot }} />
      {config.label}
    </Badge>
  );
};
