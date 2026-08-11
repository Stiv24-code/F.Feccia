import type { ReactNode } from 'react';
import { NavLink } from 'react-router-dom';
import { useAuth } from '@/lib/auth-context';
import { Button } from '@/components/ui/button';
import { Truck, LogOut, Package, Building2 } from 'lucide-react';

// Shell dedicata al portale cliente — deliberatamente NON AppShell: quella
// ha branding/titolo e gruppi di nav hardcoded per lo staff (vedi
// AppShell.js's getTitle()/navItems) senza un hook per uno "shell mode"
// diverso, e le sole 3 pagine del portale cliente non hanno bisogno di una
// sidebar. Retrofittarla avrebbe rischiato regressioni su un componente
// condiviso da tutte le pagine staff.
const navItems = [
  { label: 'I miei ordini', path: '/portale', icon: Package },
  { label: 'La mia anagrafica', path: '/portale/anagrafica', icon: Building2 },
];

export interface ClientPortalShellProps {
  children: ReactNode;
}

export default function ClientPortalShell({ children }: ClientPortalShellProps) {
  const { user, logout } = useAuth();

  return (
    <div className="min-h-screen flex flex-col">
      <header className="border-b bg-card">
        <div className="flex items-center gap-6 px-4 md:px-6 h-14">
          <div className="flex items-center gap-2 shrink-0">
            <div className="w-8 h-8 rounded-lg flex items-center justify-center" style={{ background: '#2A6FDB' }}>
              <Truck className="h-4 w-4 text-white" />
            </div>
            <span className="font-bold text-sm hidden sm:inline" style={{ fontFamily: "'Space Grotesk', sans-serif" }}>
              TMS · F.lli Feccia — Portale Cliente
            </span>
          </div>

          <nav className="flex items-center gap-1 flex-1">
            {navItems.map(({ label, path, icon: Icon }) => (
              <NavLink
                key={path}
                to={path}
                end
                className={({ isActive }) =>
                  `flex items-center gap-1.5 text-sm px-3 py-1.5 rounded-md transition-colors ${
                    isActive ? 'bg-primary/10 text-primary font-medium' : 'text-muted-foreground hover:text-foreground hover:bg-muted/60'
                  }`
                }
              >
                <Icon className="h-3.5 w-3.5" /> {label}
              </NavLink>
            ))}
          </nav>

          <div className="flex items-center gap-3 shrink-0">
            {user?.name && <span className="text-sm text-muted-foreground hidden md:inline">{user.name}</span>}
            <Button variant="outline" size="sm" onClick={logout} data-testid="client-portal-logout-button" className="gap-1.5">
              <LogOut className="h-3.5 w-3.5" /> Esci
            </Button>
          </div>
        </div>
      </header>

      <main id="main-content" className="flex-1 p-4 md:p-6 max-w-6xl w-full mx-auto">
        {children}
      </main>
    </div>
  );
}
