import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  CommandDialog,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
  CommandSeparator,
} from '@/components/ui/command';
import {
  ClipboardList,
  Truck,
  Users,
  FileText,
  MapPin,
  Calendar,
  Package,
  Building2,
  UserCircle,
  PiggyBank,
  Plus,
  LayoutDashboard,
} from 'lucide-react';

// Command palette globale (Ctrl+K / ⌘K).
//
// Pattern:
// - bind keyboard listener al mount, unbind al unmount
// - apre `CommandDialog` di shadcn, chiude su Esc / click esterno / scelta
// - elenco navigazione + azioni rapide. La quick-search ordini per
//   progressivo arriverà quando i listing avranno l'endpoint /search server-side.

const NAV_ITEMS = [
  { label: 'Dashboard', to: '/', icon: LayoutDashboard, keywords: ['home', 'kpi'] },
  { label: 'Ordini', to: '/ordini', icon: ClipboardList, keywords: ['orders'] },
  { label: 'Planner', to: '/planner', icon: Calendar, keywords: ['pianificazione', 'plan'] },
  { label: 'Mappa', to: '/mappa', icon: MapPin, keywords: ['map', 'gps'] },
  { label: 'Listini', to: '/listini', icon: PiggyBank, keywords: ['pricelist', 'tariffe'] },
  { label: 'Fatturazione', to: '/fatturazione', icon: FileText, keywords: ['invoices', 'fatture'] },
];

const ANAGRAFICHE_ITEMS = [
  { label: 'Clienti', to: '/anagrafiche/clienti', icon: Users },
  { label: 'Mezzi', to: '/anagrafiche/mezzi', icon: Truck },
  { label: 'Autisti', to: '/anagrafiche/autisti', icon: UserCircle },
  { label: 'Destinazioni', to: '/anagrafiche/destinazioni', icon: MapPin },
  { label: 'Vettori', to: '/anagrafiche/vettori', icon: Building2 },
  { label: 'Prodotti', to: '/anagrafiche/prodotti', icon: Package },
  { label: 'Garage', to: '/anagrafiche/garage', icon: Building2 },
];

export function CommandPalette() {
  const [open, setOpen] = useState(false);
  const navigate = useNavigate();

  useEffect(() => {
    const handler = (e) => {
      if ((e.key === 'k' || e.key === 'K') && (e.metaKey || e.ctrlKey)) {
        e.preventDefault();
        setOpen((prev) => !prev);
      }
    };
    window.addEventListener('keydown', handler);
    return () => window.removeEventListener('keydown', handler);
  }, []);

  const go = (to) => {
    setOpen(false);
    navigate(to);
  };

  return (
    <CommandDialog open={open} onOpenChange={setOpen}>
      <CommandInput placeholder="Cerca pagina o digita un comando..." />
      <CommandList>
        <CommandEmpty>Nessun risultato.</CommandEmpty>

        <CommandGroup heading="Azioni rapide">
          <CommandItem
            value="nuovo ordine"
            onSelect={() => go('/ordini?new=1')}
          >
            <Plus className="mr-2 h-4 w-4" />
            Nuovo ordine
          </CommandItem>
          <CommandItem
            value="nuova fattura"
            onSelect={() => go('/fatturazione?new=1')}
          >
            <Plus className="mr-2 h-4 w-4" />
            Nuova fattura
          </CommandItem>
        </CommandGroup>

        <CommandSeparator />

        <CommandGroup heading="Vai a">
          {NAV_ITEMS.map((item) => {
            const Icon = item.icon;
            return (
              <CommandItem
                key={item.to}
                value={`${item.label} ${(item.keywords || []).join(' ')}`}
                onSelect={() => go(item.to)}
              >
                <Icon className="mr-2 h-4 w-4" />
                {item.label}
              </CommandItem>
            );
          })}
        </CommandGroup>

        <CommandSeparator />

        <CommandGroup heading="Anagrafiche">
          {ANAGRAFICHE_ITEMS.map((item) => {
            const Icon = item.icon;
            return (
              <CommandItem
                key={item.to}
                value={item.label}
                onSelect={() => go(item.to)}
              >
                <Icon className="mr-2 h-4 w-4" />
                {item.label}
              </CommandItem>
            );
          })}
        </CommandGroup>
      </CommandList>
    </CommandDialog>
  );
}
