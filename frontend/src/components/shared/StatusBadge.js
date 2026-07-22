import { Badge } from '@/components/ui/badge';

// Ordine (PIANIFICABILE/PIANIFICATO/VIAGGIO/CHIUSO/SCARTATO) e Viaggio
// (IN_CORSO) usano classi dedicate (status-order-*, rosso/giallo/blu/verde/
// grigio) per non toccare i colori di Fatture/Proforma/Definitiva, che
// restano sulle classi condivise originali.
const statusConfig = {
  PIANIFICABILE: { label: 'Da pianificare', className: 'status-order-red', dot: '#C0392B' },
  PIANIFICATO: { label: 'Pianificato', className: 'status-order-yellow', dot: '#B58105' },
  VIAGGIO: { label: 'In viaggio', className: 'status-order-blue', dot: '#1D55AD' },
  CHIUSO: { label: 'Consegnato', className: 'status-order-green', dot: '#1F7A4D' },
  SCARTATO: { label: 'Scartato', className: 'status-order-gray', dot: '#7C879A' },
  PROFORMA: { label: 'Proforma', className: 'status-pianificabile', dot: '#F0B429' },
  DEFINITIVA: { label: 'Definitiva', className: 'status-fatturato', dot: '#2AA36B' },
  IN_CORSO: { label: 'In corso', className: 'status-order-blue', dot: '#1D55AD' },
  COMPLETATO: { label: 'Completato', className: 'status-fatturato', dot: '#2AA36B' },
};

export const StatusBadge = ({ stato }) => {
  const config = statusConfig[stato] || { label: stato, className: '', dot: '#999' };
  return (
    <Badge variant="outline" className={`${config.className} border text-[10px] px-2 py-0.5 font-medium gap-1.5`} data-testid="order-status-badge">
      <span className="h-1.5 w-1.5 rounded-full" style={{ backgroundColor: config.dot }} />
      {config.label}
    </Badge>
  );
};
