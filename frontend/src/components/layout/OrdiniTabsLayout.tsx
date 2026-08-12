import { useEffect, useState, type ReactNode } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';
import { apiClient } from '@/lib/apiClient';
import { getOrders } from '@/lib/api';

const TABS = [
  { path: '/ordini-in-ingresso', label: 'In arrivo' },
  { path: '/ordini', label: 'Registro ordini' },
  { path: '/ordini-in-ingresso/template', label: 'Template PDF' },
] as const;

// Striscia di tab condivisa da Raccolta Ordini / Ordini in Ingresso / Template
// PDF: le tre sezioni vivono su route distinte (invariate, per non rompere i
// link diretti) ma nel design sono un unico blocco "Ordini" — questo layout
// le presenta come tale senza spostarne la logica. I conteggi delle tab non
// attive richiedono una loro fetch indipendente (leggera, solo la length):
// il componente della tab attiva fa comunque la propria fetch completa.
export default function OrdiniTabsLayout({ children }: { children: ReactNode }) {
  const location = useLocation();
  const navigate = useNavigate();
  const [nInbound, setNInbound] = useState<number | null>(null);
  const [nOrders, setNOrders] = useState<number | null>(null);

  useEffect(() => {
    apiClient.v1InboundOrdersList()
      .then((res) => setNInbound((res.data ?? []).filter((o) => o.status === 'pending').length))
      .catch(() => setNInbound(null));
    getOrders({})
      .then((res: { data: unknown[] }) => setNOrders(res.data.length))
      .catch(() => setNOrders(null));
  }, []);

  const countFor = (path: (typeof TABS)[number]['path']) => {
    if (path === '/ordini-in-ingresso') return nInbound;
    if (path === '/ordini') return nOrders;
    return null;
  };

  return (
    <div className="space-y-4">
      <div className="flex items-center gap-1 border-b" role="tablist" aria-label="Sezioni Ordini">
        {TABS.map((tab) => {
          const active = location.pathname === tab.path;
          const count = countFor(tab.path);
          return (
            <button
              key={tab.path}
              role="tab"
              aria-selected={active}
              onClick={() => navigate(tab.path)}
              className={`flex items-center gap-1.5 px-3 py-2 text-sm font-medium border-b-2 -mb-px transition-colors ${
                active
                  ? 'border-primary text-foreground'
                  : 'border-transparent text-muted-foreground hover:text-foreground'
              }`}
            >
              {tab.label}
              {count != null && (
                <span className="rounded-full bg-muted px-1.5 py-0.5 text-[10px] font-semibold tabular-nums text-muted-foreground">
                  {count}
                </span>
              )}
            </button>
          );
        })}
      </div>
      {children}
    </div>
  );
}
