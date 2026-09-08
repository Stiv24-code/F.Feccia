/**
 * Formatta un importo in euro con separatore migliaia (punto) per locale IT.
 * Es: 2150 → "2.150", 29750 → "29.750", 950 → "950"
 */
export const formatEuro = (value: number | string | null | undefined): string => {
  const num = typeof value === 'number' ? value : Number(value) || 0;
  return num.toLocaleString('it-IT', {
    minimumFractionDigits: 0,
    maximumFractionDigits: 2,
    useGrouping: true,
  });
};

/**
 * Data breve per tabelle e liste: `2026-09-13` → `13/09`.
 *
 * Le date arrivano dall'API in ISO e finora venivano stampate così com'erano:
 * l'ISO è un formato di trasporto, non di lettura, e in una colonna stretta
 * costa il doppio dei caratteri utili. L'anno si ricava dal contesto (filtro
 * di periodo, settimana del planner); dove serve esplicito c'è
 * `formatDayMonthYear`.
 *
 * Parsing sulla stringa e non via `new Date(iso)`: quest'ultimo interpreta
 * `YYYY-MM-DD` come UTC e a fuso positivo restituisce il giorno prima.
 */
const ISO_DATE = /^(\d{4})-(\d{2})-(\d{2})/;

export const formatDayMonth = (iso: string | null | undefined): string => {
  if (!iso) return '—';
  const m = ISO_DATE.exec(iso);
  return m ? `${m[3]}/${m[2]}` : iso;
};

export const formatDayMonthYear = (iso: string | null | undefined): string => {
  if (!iso) return '—';
  const m = ISO_DATE.exec(iso);
  return m ? `${m[3]}/${m[2]}/${m[1]}` : iso;
};

/** `06:00:00` → `06:00`. Tollera già-HH:MM e valori vuoti. */
export const formatTime = (value: string | null | undefined): string => {
  if (!value) return '';
  const m = /^(\d{1,2}):(\d{2})/.exec(value);
  return m ? `${m[1].padStart(2, '0')}:${m[2]}` : value;
};

/**
 * Finestra orario accanto alla data, come nel mockup: `06:00–08:00`.
 * Stringa vuota se non c'è nessuno dei due estremi, così il chiamante può
 * ometterla senza controlli.
 */
export const formatTimeWindow = (
  from: string | null | undefined,
  to: string | null | undefined,
): string => {
  const a = formatTime(from);
  const b = formatTime(to);
  if (a && b) return a === b ? a : `${a}–${b}`;
  return a || b;
};
