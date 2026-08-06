import type { AxiosError } from 'axios';

// Normalizza un errore axios nel contratto { status, data } che RTK Query
// si aspetta dal campo `error` di una queryFn. Estratta a parte così le
// queryFn che compongono più chiamate apiClient (es. updateAdminUser in
// appApi.ts) possono riusarla senza duplicare il mapping.
export function toQueryError(err: unknown) {
  const axiosErr = err as AxiosError;
  return {
    status: axiosErr?.response?.status ?? 'FETCH_ERROR',
    data: axiosErr?.response?.data ?? axiosErr?.message ?? 'Errore di rete',
  };
}

// Adatta le Promise axios del client generato (src/lib/apiClient.ts) al
// contratto { data } | { error } che RTK Query si aspetta da una queryFn.
export async function toQueryResult<T>(promise: Promise<{ data: T }>) {
  try {
    const res = await promise;
    return { data: res.data };
  } catch (err) {
    return { error: toQueryError(err) };
  }
}

// Estrae il messaggio leggibile dal payload di errore che una mutation RTK
// Query rifiuta dopo `.unwrap()` (shape { status, data } prodotta da
// toQueryError sopra, con `data` che è il body JSON del backend Go —
// tipicamente { detail: string }). Centralizzato qui così i componenti non
// ripetono lo stesso cast/optional-chaining in ogni catch.
export function getMutationErrorMessage(err: unknown): string | undefined {
  const data = (err as { data?: unknown } | undefined)?.data;
  if (data && typeof data === 'object' && 'detail' in data) {
    const detail = (data as { detail?: unknown }).detail;
    if (typeof detail === 'string') return detail;
  }
  return undefined;
}
