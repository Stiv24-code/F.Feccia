import type { AxiosError } from 'axios';

// Estrae il messaggio d'errore da un errore axios grezzo — per le chiamate
// non ancora migrate su RTK Query (src/lib/api.js). Il backend Go risponde
// { error: msg } (utils.ErrorResponse); { detail } è solo per eventuali
// endpoint legacy non-Go. Centralizzato qui così ogni pagina non ripete lo
// stesso optional-chaining; per gli errori di mutation RTK Query vedi invece
// getMutationErrorMessage in src/store/api/rtkQueryHelpers.ts.
export function getApiErrorMessage(err: unknown): string | undefined {
  const axiosErr = err as AxiosError<{ error?: string; detail?: string }> | undefined;
  return axiosErr?.response?.data?.error || axiosErr?.response?.data?.detail;
}

// Status HTTP della risposta d'errore, quando serve distinguere due errori
// con lo stesso testo generico ma significato diverso (qui non capita, ma
// per sicurezza usare status+testo insieme è più robusto del solo testo).
export function getApiErrorStatus(err: unknown): number | undefined {
  const axiosErr = err as AxiosError | undefined;
  return axiosErr?.response?.status;
}
