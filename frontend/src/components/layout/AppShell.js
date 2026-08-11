import { useEffect, useState } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';
import { useAuth } from '@/lib/auth-context';
import { getInitialTheme, applyTheme } from '@/lib/theme';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Button } from '@/components/ui/button';
import { Sheet, SheetContent } from '@/components/ui/sheet';
import { CommandPalette } from '@/components/shared/CommandPalette';
import {
  LayoutDashboard, Users, MapPin, Truck, UserCircle, Building2,
  Package, Warehouse, ClipboardList, CalendarRange, Route, FileText,
  ListOrdered, Menu, LogOut, ChevronDown, ChevronRight, Map,
  Globe, Landmark, BookOpen, UserCog, Droplets, Sun, Moon, Inbox,
  PanelLeftClose, PanelLeftOpen
} from 'lucide-react';

const SIDEBAR_COLLAPSED_KEY = 'tms-sidebar-collapsed';

// Se `roles` è presente, la voce è visibile solo agli utenti con quel ruolo.
// Se `roles` è omesso, la voce è visibile a tutti gli autenticati.
// Vedi matrice in backend/dependencies.py + rule .claude/rules/backend-python.md.
const navItems = [
  { label: 'Dashboard', path: '/dashboard', icon: LayoutDashboard },
  {
    label: 'Anagrafiche', icon: ClipboardList, children: [
      { label: 'Clienti', path: '/anagrafiche/clienti', icon: Users },
      { label: 'Mezzi', path: '/anagrafiche/mezzi', icon: Truck },
      { label: 'Autisti', path: '/anagrafiche/autisti', icon: UserCircle },
      { label: 'Destinazioni', path: '/anagrafiche/destinazioni', icon: MapPin },
      { label: 'Vettori', path: '/anagrafiche/vettori', icon: Building2 },
      { label: 'Prodotti', path: '/anagrafiche/prodotti', icon: Package },
      { label: 'Garage', path: '/anagrafiche/garage', icon: Warehouse },
      { label: 'Punti di Lavaggio', path: '/anagrafiche/lavaggi', icon: Droplets },
      { label: 'Nazioni', path: '/anagrafiche/nazioni', icon: Globe },
      { label: 'Banche', path: '/anagrafiche/banche', icon: Landmark, roles: ['admin', 'amministrazione'] },
      { label: 'Voci Contabili', path: '/anagrafiche/voci-contabili', icon: BookOpen, roles: ['admin', 'amministrazione'] },
    ]
  },
  { label: 'Listini', path: '/listini', icon: ListOrdered },
  { label: 'Raccolta Ordini', path: '/ordini', icon: ClipboardList },
  {
    label: 'Ordini in Ingresso', icon: Inbox, children: [
      { label: 'Accettazione', path: '/ordini-in-ingresso', icon: Inbox },
      { label: 'Template PDF', path: '/ordini-in-ingresso/template', icon: FileText },
    ]
  },
  { label: 'Planner', path: '/planner', icon: CalendarRange, roles: ['admin', 'planner'] },
  { label: 'Planner Drag&Drop', path: '/planner/dnd', icon: CalendarRange, roles: ['admin', 'planner'] },
  { label: 'Gestione Viaggi', path: '/viaggi', icon: Route, roles: ['admin', 'planner'] },
  { label: 'Mappa Viaggi', path: '/mappa', icon: Map },
  { label: 'Fatturazione', path: '/fatturazione', icon: FileText, roles: ['admin', 'amministrazione'] },
  { label: 'Utenti', path: '/admin/utenti', icon: UserCog, roles: ['admin'] },
];

const filterByRole = (items, role) => {
  const isVisible = (item) => !item.roles || item.roles.includes(role);
  return items
    .filter(isVisible)
    .map((item) =>
      item.children
        ? { ...item, children: item.children.filter(isVisible) }
        : item,
    );
};

const SidebarContent = ({ collapsed, onNavigate, theme, toggleTheme, onToggleCollapsed }) => {
  const location = useLocation();
  const navigate = useNavigate();
  const { user, logout } = useAuth();
  const [openGroups, setOpenGroups] = useState({ 'Anagrafiche': true });
  const visibleNavItems = filterByRole(navItems, user?.role);

  const toggleGroup = (label) => {
    setOpenGroups(prev => ({ ...prev, [label]: !prev[label] }));
  };

  const isActive = (path) => {
    if (path === '/dashboard' && (location.pathname === '/' || location.pathname === '/dashboard')) return true;
    return location.pathname === path;
  };

  const handleNav = (path) => {
    navigate(path);
    if (onNavigate) onNavigate();
  };

  return (
    <div className="flex flex-col h-full" style={{ background: 'var(--sidebar-bg)' }}>
      {/* Logo */}
      <div className="flex items-center gap-3 px-4 h-14 shrink-0" style={{ borderBottom: '1px solid var(--sidebar-border)' }}>
        <div className="w-8 h-8 rounded-lg flex items-center justify-center text-sm font-bold shrink-0" style={{ background: 'var(--sidebar-accent)', color: '#fff' }}>
          FF
        </div>
        {!collapsed && (
          <span className="flex-1 min-w-0 text-base font-semibold tracking-tight truncate" style={{ color: 'var(--sidebar-text)', fontFamily: "'Space Grotesk', sans-serif" }}>
            TMS <span className="font-normal" style={{ color: 'var(--sidebar-muted)' }}>· F.lli Feccia</span>
          </span>
        )}
        {!collapsed && onToggleCollapsed && (
          <button
            onClick={onToggleCollapsed}
            data-testid="sidebar-collapse-toggle"
            aria-label="Comprimi sidebar"
            title="Comprimi sidebar"
            className="flex items-center justify-center h-7 w-7 rounded-md shrink-0 transition-colors duration-150 hover:bg-white/10 dark:hover:bg-black/10"
            style={{ color: 'var(--sidebar-muted)' }}
          >
            <PanelLeftClose className="h-4 w-4" />
          </button>
        )}
      </div>
      {collapsed && onToggleCollapsed && (
        <div className="px-2 pt-2 shrink-0">
          <button
            onClick={onToggleCollapsed}
            data-testid="sidebar-collapse-toggle"
            aria-label="Espandi sidebar"
            title="Espandi sidebar"
            className="w-full flex items-center justify-center h-8 rounded-lg transition-colors duration-150 hover:bg-white/10 dark:hover:bg-black/10"
            style={{ color: 'var(--sidebar-muted)' }}
          >
            <PanelLeftOpen className="h-4 w-4" />
          </button>
        </div>
      )}

      {/* Nav */}
      <ScrollArea className="flex-1 px-2 py-3">
        <nav className="flex flex-col gap-0.5">
          {visibleNavItems.map((item) => {
            if (item.children) {
              const isOpen = openGroups[item.label];
              const hasActiveChild = item.children.some(c => isActive(c.path));
              return (
                <div key={item.label}>
                  <button
                    data-testid={`sidebar-nav-group-${item.label.toLowerCase()}`}
                    onClick={() => toggleGroup(item.label)}
                    className="w-full flex items-center gap-3 rounded-lg px-3 py-2 text-sm transition-colors duration-150 hover:bg-white/5 dark:hover:bg-black/5"
                    style={{ color: hasActiveChild ? 'var(--sidebar-active-text)' : 'var(--sidebar-muted)' }}
                  >
                    <item.icon className="h-4 w-4 shrink-0" />
                    {!collapsed && (
                      <>
                        <span className="flex-1 text-left">{item.label}</span>
                        {isOpen ? <ChevronDown className="h-3.5 w-3.5" /> : <ChevronRight className="h-3.5 w-3.5" />}
                      </>
                    )}
                  </button>
                  {isOpen && !collapsed && (
                    <div className="ml-4 pl-3 border-l" style={{ borderColor: 'var(--sidebar-border)' }}>
                      {item.children.map((child) => (
                        <button
                          key={child.path}
                          data-testid={`sidebar-nav-item-${child.path.replace(/\//g, '-').slice(1)}`}
                          onClick={() => handleNav(child.path)}
                          className={`w-full flex items-center gap-3 rounded-lg px-3 py-1.5 text-sm transition-colors duration-150 border-l-2 ${isActive(child.path) ? '' : 'hover:bg-white/5 dark:hover:bg-black/5'}`}
                          style={isActive(child.path)
                            ? { background: 'var(--sidebar-active-bg)', color: 'var(--sidebar-active-text)', borderColor: 'var(--sidebar-accent)' }
                            : { color: 'var(--sidebar-muted)', borderColor: 'transparent' }}
                        >
                          <child.icon className="h-3.5 w-3.5 shrink-0" />
                          <span>{child.label}</span>
                        </button>
                      ))}
                    </div>
                  )}
                </div>
              );
            }
            return (
              <button
                key={item.path}
                data-testid={`sidebar-nav-item-${item.path.replace(/\//g, '-').slice(1)}`}
                onClick={() => handleNav(item.path)}
                className={`w-full flex items-center gap-3 rounded-lg px-3 py-2 text-sm transition-colors duration-150 ${isActive(item.path) ? '' : 'hover:bg-white/5 dark:hover:bg-black/5'}`}
                style={isActive(item.path) ? { background: 'var(--sidebar-active-bg)', color: 'var(--sidebar-active-text)' } : { color: 'var(--sidebar-muted)' }}
              >
                <item.icon className="h-4 w-4 shrink-0" />
                {!collapsed && <span>{item.label}</span>}
              </button>
            );
          })}
        </nav>
      </ScrollArea>

      {/* Theme toggle */}
      <div className="px-2 pt-1 pb-2 shrink-0">
        <button
          onClick={toggleTheme}
          data-testid="theme-toggle-button"
          aria-label={theme === 'dark' ? 'Passa al tema chiaro' : 'Passa al tema scuro'}
          title={theme === 'dark' ? 'Tema chiaro' : 'Tema scuro'}
          className="w-full flex items-center gap-3 rounded-lg px-3 py-2 text-sm transition-colors duration-150 hover:bg-white/5 dark:hover:bg-black/5"
          style={{ color: 'var(--sidebar-muted)' }}
        >
          {theme === 'dark' ? <Sun className="h-4 w-4 shrink-0" /> : <Moon className="h-4 w-4 shrink-0" />}
          {!collapsed && <span>{theme === 'dark' ? 'Tema chiaro' : 'Tema scuro'}</span>}
        </button>
      </div>

      {/* User section */}
      <div className="px-3 py-3 shrink-0" style={{ borderTop: '1px solid var(--sidebar-border)' }}>
        {!collapsed && user && (
          <div className="flex items-center gap-2 mb-2 px-1">
            <div className="w-7 h-7 rounded-full flex items-center justify-center text-xs font-medium" style={{ background: '#3A4A63', color: '#fff' }}>
              {user.name?.charAt(0)?.toUpperCase()}
            </div>
            <div className="flex-1 min-w-0">
              <p className="text-xs font-medium truncate" style={{ color: 'var(--sidebar-active-text)' }}>{user.name}</p>
              <p className="text-[10px] truncate" style={{ color: 'var(--sidebar-muted)' }}>{user.role}</p>
            </div>
          </div>
        )}
        <Button
          data-testid="sidebar-logout-button"
          variant="ghost"
          size="sm"
          onClick={logout}
          className="w-full justify-start gap-2 text-xs hover:bg-white/[0.08] dark:hover:bg-black/[0.08]"
          style={{ color: 'var(--sidebar-muted)' }}
        >
          <LogOut className="h-3.5 w-3.5" />
          {!collapsed && 'Esci'}
        </Button>
      </div>
    </div>
  );
};

const AppShell = ({ children }) => {
  const [mobileOpen, setMobileOpen] = useState(false);
  const [theme, setTheme] = useState(getInitialTheme);
  const [collapsed, setCollapsed] = useState(() => {
    try { return window.localStorage.getItem(SIDEBAR_COLLAPSED_KEY) === '1'; } catch { return false; }
  });
  const location = useLocation();

  useEffect(() => { applyTheme(theme); }, [theme]);

  const toggleTheme = () => setTheme(t => (t === 'dark' ? 'light' : 'dark'));
  const toggleCollapsed = () => setCollapsed(c => {
    const next = !c;
    try { window.localStorage.setItem(SIDEBAR_COLLAPSED_KEY, next ? '1' : '0'); } catch { /* ignore */ }
    return next;
  });

  // Get page title from path
  const getTitle = () => {
    const path = location.pathname;
    if (path === '/' || path === '/dashboard') return 'Dashboard';
    if (path.includes('clienti')) return 'Clienti';
    if (path.includes('destinazioni')) return 'Destinazioni';
    if (path.includes('mezzi')) return 'Mezzi';
    if (path.includes('autisti')) return 'Autisti';
    if (path.includes('vettori')) return 'Vettori';
    if (path.includes('prodotti')) return 'Prodotti';
    if (path.includes('lavaggi')) return 'Punti di Lavaggio';
    if (path.includes('garage')) return 'Garage / Parcheggio';
    if (path.includes('listini')) return 'Listini';
    // prima del generico 'ordini': 'ordini-in-ingresso' lo conterrebbe
    if (path.includes('ordini-in-ingresso')) return path.includes('template') ? 'Template PDF per cliente' : 'Ordini in Ingresso';
    if (path.includes('ordini')) return 'Raccolta Ordini';
    if (path.includes('planner')) return 'Planner';
    if (path.includes('viaggi')) return 'Gestione Viaggi';
    if (path.includes('mappa')) return 'Mappa Viaggi';
    if (path.includes('fatturazione')) return 'Fatturazione';
    return 'F.lli Feccia';
  };

  return (
    <div className="flex h-screen overflow-hidden">
      {/* Skip-link per screen reader / keyboard nav (WCAG 2.4.1, issue #28) */}
      <a href="#main-content" className="skip-to-main">Salta al contenuto principale</a>
      <CommandPalette />

      {/* Desktop Sidebar */}
      <aside
        className={`hidden lg:block shrink-0 h-full transition-[width] duration-200 ${collapsed ? 'w-[76px]' : 'w-[260px]'}`}
        style={{ boxShadow: '4px 0 24px rgba(0,0,0,0.15)' }}
      >
        <SidebarContent collapsed={collapsed} theme={theme} toggleTheme={toggleTheme} onToggleCollapsed={toggleCollapsed} />
      </aside>

      {/* Mobile Sidebar */}
      <Sheet open={mobileOpen} onOpenChange={setMobileOpen}>
        <SheetContent side="left" className="p-0 w-[280px]" style={{ background: 'var(--sidebar-bg)' }}>
          <SidebarContent collapsed={false} onNavigate={() => setMobileOpen(false)} theme={theme} toggleTheme={toggleTheme} />
        </SheetContent>
      </Sheet>

      {/* Main */}
      <div className="flex-1 flex flex-col min-w-0 h-full">
        {/* Topbar */}
        <header className="h-14 shrink-0 flex items-center gap-3 px-4 lg:px-6 border-b bg-card">
          <Button
            variant="ghost"
            size="icon"
            className="lg:hidden h-8 w-8"
            onClick={() => setMobileOpen(true)}
            data-testid="mobile-menu-button"
            aria-label="Apri menu di navigazione"
          >
            <Menu className="h-4 w-4" />
          </Button>
          <h1 className="text-lg font-semibold tracking-tight" style={{ fontFamily: "'Space Grotesk', sans-serif" }} data-testid="page-title">
            {getTitle()}
          </h1>
          <div className="ml-auto flex items-center gap-3">
            <div className="hidden md:flex items-center text-xs text-muted-foreground gap-2">
              <kbd className="px-1.5 py-0.5 rounded border bg-muted font-mono text-[10px]">Ctrl</kbd>
              <span>+</span>
              <kbd className="px-1.5 py-0.5 rounded border bg-muted font-mono text-[10px]">K</kbd>
              <span>per cercare</span>
            </div>
          </div>
        </header>

        {/* Content */}
        <main id="main-content" className="flex-1 overflow-y-auto" tabIndex={-1}>
          <div className="px-3 sm:px-4 lg:px-6 py-4">
            {children}
          </div>
        </main>
      </div>
    </div>
  );
};

export default AppShell;
