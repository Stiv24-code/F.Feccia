import type { AxiosError } from 'axios';

// Estrae il messaggio d'errore dal backend Go ({ detail: string }) da un
// errore axios grezzo — per le chiamate non ancora migrate su RTK Query
// (src/lib/api.js). Centralizzato qui così ogni pagina non ripete lo stesso
// optional-chaining; per gli errori di mutation RTK Query vedi invece
// getMutationErrorMessage in src/store/api/rtkQueryHelpers.ts.
export function getApiErrorMessage(err: unknown): string | undefined {
  const axiosErr = err as AxiosError<{ detail?: string }> | undefined;
  return axiosErr?.response?.data?.detail;
}
