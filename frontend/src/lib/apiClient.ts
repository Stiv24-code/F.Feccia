// Client tipizzato generato da swagger (`yarn generate:api`, src/api/) —
// wiring hand-written, non toccato dalla rigenerazione (che riscrive solo
// src/api/*). Riusa gli stessi interceptor auth di lib/api.js (token in
// memoria, refresh dedupato, feedback 429) così le due istanze axios
// condividono un'unica fonte di verità sullo stato di autenticazione.
//
// A differenza di `api` in lib/api.js (baseURL già .../api/v1, path brevi
// come '/auth/login'), il client generato usa i path pieni dichiarati da
// swagger (es. '/api/v1/auth/login'): baseURL qui è la sola origin/vuoto.
import { Api } from '@/api/Api';
import { HttpClient } from '@/api/http-client';
import { attachAuthInterceptors } from './api';

const BACKEND_URL = import.meta.env.VITE_BACKEND_URL || '';

// HttpClient generato fa `axiosConfig.baseURL || "https://localhost:8080"`:
// una stringa vuota (dev, path relativi) sarebbe falsy e attiverebbe quel
// fallback. window.location.origin è equivalente a "nessun baseURL" per
// richieste same-origin, ma resta truthy ed evita il default hardcoded.
const httpClient = new HttpClient({
  baseURL: BACKEND_URL || window.location.origin,
  withCredentials: true,
});

attachAuthInterceptors(httpClient.instance);

export const apiClient = new Api(httpClient);
export * from '@/api/data-contracts';
