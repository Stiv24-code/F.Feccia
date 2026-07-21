import type { AxiosError } from 'axios';

// Adatta le Promise axios del client generato (src/lib/apiClient.ts) al
// contratto { data } | { error } che RTK Query si aspetta da una queryFn.
export async function toQueryResult<T>(promise: Promise<{ data: T }>) {
  try {
    const res = await promise;
    return { data: res.data };
  } catch (err) {
    const axiosErr = err as AxiosError;
    return {
      error: {
        status: axiosErr?.response?.status ?? 'FETCH_ERROR',
        data: axiosErr?.response?.data ?? axiosErr?.message ?? 'Errore di rete',
      },
    };
  }
}
