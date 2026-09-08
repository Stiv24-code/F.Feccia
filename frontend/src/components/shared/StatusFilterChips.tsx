export type OrderStatus = 'PIANIFICABILE' | 'PIANIFICATO' | 'VIAGGIO' | 'CHIUSO' | 'SCARTATO';

export const ORDER_STATUSES: OrderStatus[] = [
  'PIANIFICABILE',
  'PIANIFICATO',
  'VIAGGIO',
  'CHIUSO',
  'SCARTATO',
];

// Chip contatore per stato. Mostrano quanti ordini ci sono in ciascuno stato
// senza aprire un menù: in Ordini erano stati sostituiti da una select, che
// nasconde proprio l'informazione utile (rilievo "Filtro stati" dell'audit).
// Il Planner li aveva già — questo componente è la versione condivisa, così
// le due pagine non divergono più.
//
// Il fondo resta la variante pastello `.status-order-*` e non la pill piena di
// StatusBadge: qui sono filtri, non lo stato di una riga, e sei di fila a
// colore pieno coprirebbero la testa della pagina.
const CHIPS: { key: OrderStatus | null; label: string; className: string }[] = [
  { key: null, label: 'Tutti', className: 'border-muted-foreground/30 bg-background text-foreground' },
  { key: 'PIANIFICABILE', label: 'Da pianificare', className: 'status-order-red' },
  { key: 'PIANIFICATO', label: 'Pianificati', className: 'status-order-yellow' },
  { key: 'VIAGGIO', label: 'In viaggio', className: 'status-order-blue' },
  { key: 'CHIUSO', label: 'Consegnati', className: 'status-order-green' },
  { key: 'SCARTATO', label: 'Scartati', className: 'status-order-gray' },
];

export interface StatusFilterChipsProps {
  value: OrderStatus | null;
  onChange: (value: OrderStatus | null) => void;
  /** Conteggi per stato più `all` per il chip "Tutti". */
  counts: Record<string, number>;
  /** Prefisso dei data-testid, per non collidere fra Ordini e Planner. */
  testIdPrefix: string;
}

export const StatusFilterChips = ({ value, onChange, counts, testIdPrefix }: StatusFilterChipsProps) => (
  <div className="flex items-center gap-1.5 flex-wrap">
    {CHIPS.map((c) => {
      const active = value === c.key;
      const count = c.key ? counts[c.key] : counts.all;
      return (
        <button
          key={c.key ?? 'all'}
          type="button"
          onClick={() => onChange(c.key)}
          aria-pressed={active}
          data-testid={`${testIdPrefix}-chip-${c.key ?? 'all'}`}
          className={`text-xs font-semibold rounded-full border px-3 py-1 transition ${c.className} ${
            active ? 'ring-2 ring-offset-1 ring-primary' : 'opacity-70 hover:opacity-100'
          }`}
        >
          {c.label} <span className="tabular-nums">{count ?? 0}</span>
        </button>
      );
    })}
  </div>
);
