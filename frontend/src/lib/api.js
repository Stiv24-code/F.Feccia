import axios from 'axios';
import { toast } from 'sonner';

// VITE_BACKEND_URL viene iniettata al build (production) o letta da .env (dev).
// Se vuota → URL relativi (`/api/v1/...`), così in dev Vite proxia /api al
// backend locale e in prod nginx serve sia SPA sia /api dallo stesso host.
// nginx (dev.conf e default.conf) fa passthrough 1:1 su /api/*, quindi il
// prefisso /v1 lo mettiamo qui — le route Go vivono sotto /api/v1 (vedi
// anche src/api/, il client generato da swagger che usa gli stessi path pieni).
const BACKEND_URL = import.meta.env.VITE_BACKEND_URL || '';
const API = `${BACKEND_URL}/api/v1`;

// Crea un'istanza axios con credentials abilitate (per il cookie refresh httpOnly).
const api = axios.create({
  baseURL: API,
  withCredentials: true,
});

// ─── Access token in memoria (non in localStorage: riduce esposizione XSS) ───
let accessToken = null;
let onAuthFailure = null;

export const setAccessToken = (token) => { accessToken = token || null; };
export const getAccessToken = () => accessToken;
export const setOnAuthFailure = (fn) => { onAuthFailure = fn; };

// ─── Interceptor auth condivisi ───
// Estratti in una funzione applicabile a più istanze axios (questa `api` +
// l'HttpClient del client generato da swagger in src/lib/apiClient.ts), così
// il token in memoria, il refresh dedupato e il feedback 429 restano un'unica
// fonte di verità invece di duplicare la logica per ogni client.
//
//   1. request: aggiunge Authorization: Bearer se presente
//   2. response: su 401 (token scaduto) tenta un refresh dedupato e ripropone
//      la richiesta; su 429 (rate limit) mostra un toast con il tempo
//      d'attesa estratto dall'header Retry-After del backend
let refreshInFlight = null;

export const attachAuthInterceptors = (instance) => {
  instance.interceptors.request.use((config) => {
    if (accessToken) {
      config.headers = config.headers || {};
      config.headers.Authorization = `Bearer ${accessToken}`;
    }
    return config;
  });

  instance.interceptors.response.use(
    (response) => response,
    async (error) => {
      const original = error.config;
      const status = error.response?.status;

      // Rate limit: lasciamo propagare l'errore ma diamo feedback immediato
      // all'utente. Il messaggio backend in `detail` è gia' specifico se presente.
      if (status === 429) {
        const retryAfter = Number(error.response?.headers?.['retry-after']) || 0;
        const waitMsg = retryAfter > 0
          ? `Riprova tra ${retryAfter} secondi.`
          : 'Riprova tra qualche minuto.';
        toast.error('Troppi tentativi', { description: waitMsg });
        return Promise.reject(error);
      }

      const isAuthEndpoint =
        original?.url?.includes('/auth/login') ||
        original?.url?.includes('/auth/refresh');

      if (status !== 401 || original?._retry || isAuthEndpoint) {
        return Promise.reject(error);
      }

      original._retry = true;
      try {
        // Dedup: se un refresh è già in corso, riusa la stessa Promise
        if (!refreshInFlight) {
          refreshInFlight = api.post('/auth/refresh').finally(() => {
            refreshInFlight = null;
          });
        }
        const res = await refreshInFlight;
        if (res?.data?.access_token) {
          accessToken = res.data.access_token;
          original.headers = original.headers || {};
          original.headers.Authorization = `Bearer ${accessToken}`;
          return instance(original);
        }
        throw new Error('refresh senza access_token');
      } catch (refreshErr) {
        accessToken = null;
        if (typeof onAuthFailure === 'function') onAuthFailure();
        return Promise.reject(refreshErr);
      }
    },
  );
};

attachAuthInterceptors(api);

// Auth
export const login = (data) => api.post('/auth/login', data);
export const getMe = () => api.get('/auth/me');
export const refreshSession = () => api.post('/auth/refresh');
export const logout = () => api.post('/auth/logout');
// Autoregistrazione cliente (pubblica): stessa risposta di login (access
// token in body, refresh token nel cookie httpOnly) — auto-login immediato,
// nessun approval. Vedi backend/internal/services/auth.RegisterClient.
// Le route scoped del portale cliente (/me/anagrafica, /me/orders) sono
// invece su RTK Query in store/api/appApi.ts, come le altre pagine anagrafiche.
export const registerClient = (data) => api.post('/auth/register-cliente', data);

// Admin: gestione utenti migrata su RTK Query (src/store/api/appApi.ts,
// client generato da swagger). I profili RBAC custom non sono mai stati
// implementati lato backend e sono stati rimossi dal frontend.

// Dashboard e CRUD clienti: migrati su RTK Query (src/store/api/appApi.ts,
// client generato da swagger).

// Destinations, Motrici, Semirimorchi, Drivers, Carriers, Products, Garages:
// CRUD migrato su RTK Query (src/store/api/appApi.ts). Le liste restano qui
// perché usate anche da pagine non ancora migrate (Orders, PriceLists,
// Planner, Trips).
export const getDestinations = (search = '') => api.get(`/destinations?search=${search}`);

export const getMotrici = (search = '') => api.get(`/motrici?search=${search}`);
export const getSemirimorchi = (search = '') => api.get(`/semirimorchi?search=${search}`);

export const getDrivers = (search = '') => api.get(`/drivers?search=${search}`);

export const getCarriers = (search = '') => api.get(`/carriers?search=${search}`);

export const getProducts = (search = '') => api.get(`/products?search=${search}`);

export const getGarages = () => api.get('/garages');

// Vehicle Types
export const getVehicleTypes = () => api.get('/vehicle-types');
export const createVehicleType = (data) => api.post('/vehicle-types', data);

// Accessory Costs
export const getAccessoryCosts = () => api.get('/accessory-costs');
export const createAccessoryCost = (data) => api.post('/accessory-costs', data);

// Transport Categories
export const getTransportCategories = () => api.get('/transport-categories');
export const createTransportCategory = (data) => api.post('/transport-categories', data);

// Price Lists
export const getPriceLists = (clienteId = '') => api.get(`/pricelists?cliente_id=${clienteId}`);
export const getPriceList = (id) => api.get(`/pricelists/${id}`);
export const createPriceList = (data) => api.post('/pricelists', data);
export const updatePriceList = (id, data) => api.put(`/pricelists/${id}`, data);
export const deletePriceList = (id) => api.delete(`/pricelists/${id}`);
export const addPriceListItem = (id, item) => api.post(`/pricelists/${id}/items`, item);
export const updatePriceListItem = (id, itemId, item) => api.put(`/pricelists/${id}/items/${itemId}`, item);
export const deletePriceListItem = (id, itemId) => api.delete(`/pricelists/${id}/items/${itemId}`);
export const lookupTariff = (params) => {
  const qs = new URLSearchParams();
  Object.entries(params).forEach(([k, v]) => { if (v) qs.set(k, v); });
  return api.get(`/pricelists/lookup-tariff?${qs.toString()}`);
};

// Orders
export const getOrders = (params = {}) => {
  const qs = new URLSearchParams();
  Object.entries(params).forEach(([k, v]) => { if (v) qs.set(k, v); });
  return api.get(`/orders?${qs.toString()}`);
};
export const createOrder = (data) => api.post('/orders', data);
export const getOrder = (id) => api.get(`/orders/${id}`);
export const updateOrder = (id, data) => api.put(`/orders/${id}`, data);
export const assignOrder = (id, data) => api.patch(`/orders/${id}/assign`, data);
export const unassignOrder = (id) => api.patch(`/orders/${id}/unassign`);
// Percorso stradale truck-aware (OpenRouteService) — fino a 3 alternative
// effimere (nessuna scrittura su DB) da mostrare al manager in fase di
// assegnazione, poi ricalcolo/persistenza su modifica manuale dei waypoint.
export const getOrderRouteAlternatives = (id, { garageId, washStationId } = {}) =>
  api.post(`/orders/${id}/route-alternatives`, { garage_id: garageId || '', wash_station_id: washStationId || '' });
export const updateOrderRoute = (id, waypoints) => api.patch(`/orders/${id}/route`, { waypoints });
// Ricerca indirizzo per il MapPicker (Destinazioni/Garage/Punti di Lavaggio):
// text -> fino a 5 candidati con coordinate, l'utente sceglie quello giusto.
export const geocodeSearch = (query) => api.get(`/geocode/search?q=${encodeURIComponent(query)}`);
export const startOrder = (id) => api.patch(`/orders/${id}/start`);
export const closeOrder = (id) => api.patch(`/orders/${id}/close`);
export const discardOrder = (id) => api.patch(`/orders/${id}/discard`);
export const deleteOrder = (id) => api.delete(`/orders/${id}`);
// Suggerimenti ordini di ritorno per riempire viaggi vuoti (issue #32)
export const getReturnSuggestions = (id, params = {}) => {
  const qs = new URLSearchParams();
  Object.entries(params).forEach(([k, v]) => { if (v !== undefined && v !== '') qs.set(k, v); });
  return api.get(`/orders/${id}/return-suggestions?${qs.toString()}`);
};

// Trips
export const getTrips = (stato = '') => api.get(`/trips?stato=${stato}`);
export const createTrip = (data) => api.post('/trips', data);
export const getTrip = (id) => api.get(`/trips/${id}`);
export const startTrip = (id) => api.patch(`/trips/${id}/start`);
export const completeTrip = (id) => api.patch(`/trips/${id}/complete`);
export const addOrderToTrip = (tripId, orderId) => api.patch(`/trips/${tripId}/add-order?order_id=${orderId}`);
export const recomputeTripSegments = (tripId) => api.post(`/trips/${tripId}/recompute-segments`);
// PDF Istruzioni Operative per autista (issue #33)
export const downloadTripInstructionsPdf = (tripId) => api.get(`/trips/${tripId}/instructions/pdf`, { responseType: 'blob' });
// PDF CMR per ordine internazionale (issue #34)
export const downloadOrderCmrPdf = (orderId) => api.get(`/orders/${orderId}/cmr/pdf`, { responseType: 'blob' });

// Invoices
export const getInvoices = (params = {}) => {
  const qs = new URLSearchParams();
  Object.entries(params).forEach(([k, v]) => { if (v) qs.set(k, v); });
  return api.get(`/invoices?${qs.toString()}`);
};
export const createInvoice = (data) => api.post('/invoices', data);
export const getInvoice = (id) => api.get(`/invoices/${id}`);
export const finalizeInvoice = (id) => api.patch(`/invoices/${id}/finalize`);
export const deleteInvoice = (id) => api.delete(`/invoices/${id}`);
// PDF: ritorna axios Response con responseType blob — il chiamante salva
// con saveAs / window.URL.createObjectURL.
export const downloadInvoicePdf = (id) => api.get(`/invoices/${id}/pdf`, { responseType: 'blob' });
// Presigned URL S3 (issue #35). 404 se la fattura non è archiviata,
// il chiamante deve fare fallback a downloadInvoicePdf.
export const getInvoicePdfUrl = (id) => api.get(`/invoices/${id}/pdf-url`);

// Anagrafiche extra (#29): Nazioni, Banche, Voci Contabili — CRUD completo
// migrato su RTK Query (src/store/api/appApi.ts).

// Export
export const exportOrdersExcel = (params = {}) => {
  const qs = new URLSearchParams();
  Object.entries(params).forEach(([k, v]) => { if (v) qs.set(k, v); });
  return api.get(`/export/orders?${qs.toString()}`, { responseType: 'blob' });
};

// Mappa
export const getMapTrips = () => api.get('/map/trips');

// Availability
export const getMotriceAvailability = (dataDa, dataA) => api.get(`/availability/motrici?data_da=${dataDa}&data_a=${dataA}`);
export const getSemirimorchioAvailability = (dataDa, dataA) => api.get(`/availability/semirimorchi?data_da=${dataDa}&data_a=${dataA}`);
export const getDriverAvailability = (dataDa, dataA) => api.get(`/availability/drivers?data_da=${dataDa}&data_a=${dataA}`);

// Driver Unavailability
export const getDriverUnavailability = (params = {}) => {
  const qs = new URLSearchParams();
  Object.entries(params).forEach(([k, v]) => { if (v) qs.set(k, v); });
  return api.get(`/driver-unavailability?${qs.toString()}`);
};
export const createDriverUnavailability = (data) => api.post('/driver-unavailability', data);
export const deleteDriverUnavailability = (id) => api.delete(`/driver-unavailability/${id}`);

export default api;
