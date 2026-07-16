import { Badge } from '@/components/ui/badge';

const statusConfig = {
  PIANIFICABILE: { label: 'Da pianificare', className: 'status-pianificabile', dot: '#F0B429' },
  VIAGGIO: { label: 'In viaggio', className: 'status-viaggio', dot: '#E24A4A' },
  CHIUSO: { label: 'Chiuso', className: 'status-chiuso', dot: '#F28B2C' },
  FATTURATO: { label: 'Fatturato', className: 'status-fatturato', dot: '#2AA36B' },
  PROFORMA: { label: 'Proforma', className: 'status-pianificabile', dot: '#F0B429' },
  DEFINITIVA: { label: 'Definitiva', className: 'status-fatturato', dot: '#2AA36B' },
  IN_CORSO: { label: 'In corso', className: 'status-viaggio', dot: '#E24A4A' },
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
