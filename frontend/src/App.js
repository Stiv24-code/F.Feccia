import { lazy, Suspense } from 'react';
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { Provider } from 'react-redux';
import { store } from '@/store/store';
import { AuthProvider, useAuth } from '@/lib/auth-context';
import { Toaster } from '@/components/ui/sonner';
import LoginPage from '@/pages/LoginPage';
import ClientRegisterPage from '@/pages/portale/ClientRegisterPage';
import VerifyEmailPage from '@/pages/portale/VerifyEmailPage';
import ForgotPasswordPage from '@/pages/ForgotPasswordPage';
import ResetPasswordPage from '@/pages/ResetPasswordPage';
import AppShell from '@/components/layout/AppShell';
import OrdiniTabsLayout from '@/components/layout/OrdiniTabsLayout';
import AnagraficheTabsLayout from '@/components/layout/AnagraficheTabsLayout';
import ClientPortalShell from '@/components/layout/ClientPortalShell';
import DashboardPage from '@/pages/DashboardPage';
import '@/App.css';

// Lazy routes — solo Login + Dashboard restano nel bundle iniziale. Tutto il
// resto viene caricato al primo accesso alla rotta (~50% del bundle erano
// pagine secondarie). Il Suspense fallback usa lo stesso spinner del loading
// di auth context.
const CustomersPage = lazy(() => import('@/pages/anagrafiche/CustomersPage'));
const DestinationsPage = lazy(() => import('@/pages/anagrafiche/DestinationsPage'));
const VehiclesPage = lazy(() => import('@/pages/anagrafiche/VehiclesPage'));
const DriversPage = lazy(() => import('@/pages/anagrafiche/DriversPage'));
const CarriersPage = lazy(() => import('@/pages/anagrafiche/CarriersPage'));
const ProductsPage = lazy(() => import('@/pages/anagrafiche/ProductsPage'));
const GaragesPage = lazy(() => import('@/pages/anagrafiche/GaragesPage'));
const WashStationsPage = lazy(() => import('@/pages/anagrafiche/WashStationsPage'));
const CountriesPage = lazy(() => import('@/pages/anagrafiche/CountriesPage'));
const BanksPage = lazy(() => import('@/pages/anagrafiche/BanksPage'));
const AccountingEntriesPage = lazy(() => import('@/pages/anagrafiche/AccountingEntriesPage'));
const OrdersPage = lazy(() => import('@/pages/OrdersPage'));
const InboundOrdersPage = lazy(() => import('@/pages/InboundOrdersPage'));
const PdfTemplatesPage = lazy(() => import('@/pages/PdfTemplatesPage'));
const PlannerPage = lazy(() => import('@/pages/PlannerPage'));
const OrderDetailPage = lazy(() => import('@/pages/OrderDetailPage'));
const InvoicesPage = lazy(() => import('@/pages/InvoicesPage'));
const PriceListsPage = lazy(() => import('@/pages/PriceListsPage'));
const MapPage = lazy(() => import('@/pages/MapPage'));
const CustomerDashboardPage = lazy(() => import('@/pages/CustomerDashboardPage'));
const UsersPage = lazy(() => import('@/pages/admin/UsersPage'));
const ClientOrdersPage = lazy(() => import('@/pages/portale/ClientOrdersPage'));
const ClientAnagraficaPage = lazy(() => import('@/pages/portale/ClientAnagraficaPage'));

const PageFallback = () => (
  <div className="flex h-64 items-center justify-center">
    <div className="animate-spin h-6 w-6 border-2 border-primary border-t-transparent rounded-full" />
  </div>
);

const ProtectedRoute = ({ children }) => {
  const { user, loading } = useAuth();
  if (loading) return <div className="flex h-screen items-center justify-center"><div className="animate-spin h-8 w-8 border-4 border-primary border-t-transparent rounded-full" /></div>;
  if (!user) return <Navigate to="/login" replace />;
  return children;
};

// Route tree per lo staff (admin/amministrazione/planner/operatore) — invariato.
function StaffRoutes() {
  return (
    <AppShell>
      <Suspense fallback={<PageFallback />}>
        <Routes>
          <Route index element={<DashboardPage />} />
          <Route path="dashboard" element={<DashboardPage />} />
          <Route path="anagrafiche/clienti" element={<AnagraficheTabsLayout><CustomersPage /></AnagraficheTabsLayout>} />
          <Route path="anagrafiche/clienti/:id/cruscotto" element={<CustomerDashboardPage />} />
          <Route path="anagrafiche/destinazioni" element={<AnagraficheTabsLayout><DestinationsPage /></AnagraficheTabsLayout>} />
          <Route path="anagrafiche/mezzi" element={<AnagraficheTabsLayout><VehiclesPage /></AnagraficheTabsLayout>} />
          <Route path="anagrafiche/autisti" element={<AnagraficheTabsLayout><DriversPage /></AnagraficheTabsLayout>} />
          <Route path="anagrafiche/vettori" element={<AnagraficheTabsLayout><CarriersPage /></AnagraficheTabsLayout>} />
          <Route path="anagrafiche/prodotti" element={<AnagraficheTabsLayout><ProductsPage /></AnagraficheTabsLayout>} />
          <Route path="anagrafiche/garage" element={<AnagraficheTabsLayout><GaragesPage /></AnagraficheTabsLayout>} />
          <Route path="anagrafiche/lavaggi" element={<AnagraficheTabsLayout><WashStationsPage /></AnagraficheTabsLayout>} />
          <Route path="anagrafiche/nazioni" element={<AnagraficheTabsLayout><CountriesPage /></AnagraficheTabsLayout>} />
          <Route path="anagrafiche/banche" element={<AnagraficheTabsLayout><BanksPage /></AnagraficheTabsLayout>} />
          <Route path="anagrafiche/voci-contabili" element={<AnagraficheTabsLayout><AccountingEntriesPage /></AnagraficheTabsLayout>} />
          <Route path="listini" element={<PriceListsPage />} />
          <Route path="ordini" element={<OrdiniTabsLayout><OrdersPage /></OrdiniTabsLayout>} />
          <Route path="ordini-in-ingresso" element={<OrdiniTabsLayout><InboundOrdersPage /></OrdiniTabsLayout>} />
          <Route path="ordini-in-ingresso/template" element={<OrdiniTabsLayout><PdfTemplatesPage /></OrdiniTabsLayout>} />
          <Route path="planner" element={<PlannerPage />} />
          <Route path="planner/ordini/:id" element={<OrderDetailPage />} />
          <Route path="mappa" element={<MapPage />} />
          <Route path="fatturazione" element={<InvoicesPage />} />
          <Route path="admin/utenti" element={<UsersPage />} />
        </Routes>
      </Suspense>
    </AppShell>
  );
}

// Route tree del portale cliente (ruolo "cliente") — deliberatamente separato
// da StaffRoutes: un cliente non monta mai il codice delle pagine staff, non
// è solo "nascosto" come nel menu di AppShell (che filtra solo le voci
// visibili, non blocca la navigazione diretta via URL). Qualsiasi path
// diverso da /portale o /portale/anagrafica riporta a /portale — non esiste
// un modo per un cliente di "arrivare" su una pagina staff.
function ClientPortalRoutes() {
  return (
    <ClientPortalShell>
      <Suspense fallback={<PageFallback />}>
        <Routes>
          <Route path="portale" element={<ClientOrdersPage />} />
          <Route path="portale/anagrafica" element={<ClientAnagraficaPage />} />
          <Route path="*" element={<Navigate to="/portale" replace />} />
        </Routes>
      </Suspense>
    </ClientPortalShell>
  );
}

function AppRoutes() {
  const { user } = useAuth();
  return (
    <Routes>
      <Route path="/login" element={user ? <Navigate to="/" replace /> : <LoginPage />} />
      <Route path="/registrati" element={user ? <Navigate to="/" replace /> : <ClientRegisterPage />} />
      <Route path="/verifica-email" element={user ? <Navigate to="/" replace /> : <VerifyEmailPage />} />
      <Route path="/password-dimenticata" element={user ? <Navigate to="/" replace /> : <ForgotPasswordPage />} />
      <Route path="/reimposta-password" element={user ? <Navigate to="/" replace /> : <ResetPasswordPage />} />
      <Route path="/*" element={
        <ProtectedRoute>
          {user?.role === 'cliente' ? <ClientPortalRoutes /> : <StaffRoutes />}
        </ProtectedRoute>
      } />
    </Routes>
  );
}

function App() {
  return (
    <Provider store={store}>
      <BrowserRouter>
        <AuthProvider>
          <AppRoutes />
          <Toaster position="top-right" richColors closeButton />
        </AuthProvider>
      </BrowserRouter>
    </Provider>
  );
}

export default App;
