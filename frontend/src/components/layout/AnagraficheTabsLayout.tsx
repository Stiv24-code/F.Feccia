import { type ReactNode } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';
import { useAuth } from '@/lib/auth-context';

type Tab = { path: string; label: string; roles?: readonly string[] };

// Stessa struttura di OrdiniTabsLayout — le pagine restano su route distinte
// (invariate, per non rompere i link diretti), questo layout le presenta
// come un unico blocco "Anagrafiche" con la striscia di tab del design.
// Banche/Voci Contabili restano filtrate per ruolo come nella sidebar
// (vedi AppShell.js — stessa matrice ruoli).
const TABS: Tab[] = [
  { path: '/anagrafiche/clienti', label: 'Clienti' },
  { path: '/anagrafiche/mezzi', label: 'Mezzi' },
  { path: '/anagrafiche/autisti', label: 'Autisti' },
  { path: '/anagrafiche/destinazioni', label: 'Destinazioni' },
  { path: '/anagrafiche/vettori', label: 'Vettori' },
  { path: '/anagrafiche/prodotti', label: 'Prodotti' },
  { path: '/anagrafiche/garage', label: 'Garage' },
  { path: '/anagrafiche/lavaggi', label: 'Punti di Lavaggio' },
  { path: '/anagrafiche/nazioni', label: 'Nazioni' },
  { path: '/anagrafiche/banche', label: 'Banche', roles: ['admin', 'amministrazione'] },
  { path: '/anagrafiche/voci-contabili', label: 'Voci Contabili', roles: ['admin', 'amministrazione'] },
];

export default function AnagraficheTabsLayout({ children }: { children: ReactNode }) {
  const location = useLocation();
  const navigate = useNavigate();
  const { user } = useAuth();
  const tabs = TABS.filter((t) => !t.roles || t.roles.includes(user?.role ?? ''));

  return (
    <div className="space-y-4">
      <div className="flex items-center gap-1 border-b overflow-x-auto" role="tablist" aria-label="Sezioni Anagrafiche">
        {tabs.map((tab) => {
          const active = location.pathname === tab.path;
          return (
            <button
              key={tab.path}
              role="tab"
              aria-selected={active}
              onClick={() => navigate(tab.path)}
              className={`shrink-0 px-3 py-2 text-sm font-medium border-b-2 -mb-px transition-colors ${
                active
                  ? 'border-primary text-foreground'
                  : 'border-transparent text-muted-foreground hover:text-foreground'
              }`}
            >
              {tab.label}
            </button>
          );
        })}
      </div>
      {children}
    </div>
  );
}
