// Costanti condivise del flusso "ordini in ingresso" (porting OrderMesh):
// target dei campi template (snake_case, allineati a models.InboundOrderFieldTargets
// nel backend) e stile dei badge di stato.
import type { AxiosError } from 'axios';

export const TARGETS: Array<[string, string]> = [
  ['client', 'Cliente'],
  ['sender_email', 'E-mail mittente'],
  ['ref', 'Riferimento'],
  ['product', 'Prodotto'],
  ['kg', 'Kg'],
  ['load_date', 'Data carico'],
  ['load_place', 'Luogo carico'],
  ['delivery_date', 'Data consegna'],
  ['delivery_place', 'Luogo consegna'],
  ['rate', 'Nolo'],
  ['notes', 'Note'],
];

export const TARGET_LABEL: Record<string, string> = Object.fromEntries(TARGETS);

export const STATUS_BADGE: Record<string, { className: string; label: string }> = {
  pending: { className: 'bg-amber-100 text-amber-800 dark:bg-amber-500/15 dark:text-amber-300', label: 'Da confermare' },
  accepted: { className: 'bg-emerald-100 text-emerald-800 dark:bg-emerald-500/15 dark:text-emerald-300', label: 'Accettato' },
  modify: { className: 'bg-red-100 text-red-800 dark:bg-red-500/15 dark:text-red-300', label: 'In modifica' },
};

// Il backend Go risponde { error: msg } (utils.ErrorResponse) — diverso dal
// { detail } letto da lib/apiError.ts per gli endpoint legacy.
export function getInboundApiError(err: unknown): string {
  const axiosErr = err as AxiosError<{ error?: string; detail?: string }> | undefined;
  return (
    axiosErr?.response?.data?.error ||
    axiosErr?.response?.data?.detail ||
    (err instanceof Error ? err.message : 'Errore imprevisto')
  );
}

export const fmtKg = (kg?: number) => (!kg ? '—' : kg.toLocaleString('it-IT'));
