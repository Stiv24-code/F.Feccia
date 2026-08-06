import { useEffect, useState, useCallback } from 'react';

// Hook per persistere filtri di pagina su localStorage (issue #28).
//
// Uso tipico in OrdersPage / PlannerPage / InvoicesPage:
//
//   const [filters, setFilters, resetFilters] = usePersistentFilters(
//     'orders',
//     { stato: '', cliente_id: '', search: '', tipologia: '', data_da: '', data_a: '' },
//   );
//
// La key `orders` viene namespacizzata con prefisso `tms.filters.` per non
// collidere con altri usi di localStorage (auth, ecc.).

const PREFIX = 'tms.filters.';

const safeGet = <T,>(key: string, fallback: T): T => {
  try {
    const raw = localStorage.getItem(PREFIX + key);
    if (!raw) return fallback;
    const parsed = JSON.parse(raw);
    return { ...fallback, ...parsed };
  } catch {
    return fallback;
  }
};

const safeSet = <T,>(key: string, value: T) => {
  try {
    localStorage.setItem(PREFIX + key, JSON.stringify(value));
  } catch {
    // localStorage piena o disabilitato — fail silenzioso, i filtri restano
    // solo in memoria fino al refresh.
  }
};

const safeClear = (key: string) => {
  try {
    localStorage.removeItem(PREFIX + key);
  } catch {
    /* idem */
  }
};

export type FiltersUpdater<T> = Partial<T> | ((prev: T) => T);

export function usePersistentFilters<T extends object>(key: string, defaults: T): [T, (updater: FiltersUpdater<T>) => void, () => void] {
  const [filters, setFiltersState] = useState<T>(() => safeGet(key, defaults));

  useEffect(() => {
    safeSet(key, filters);
  }, [key, filters]);

  const setFilters = useCallback((updater: FiltersUpdater<T>) => {
    setFiltersState((prev) =>
      typeof updater === 'function' ? (updater as (prev: T) => T)(prev) : { ...prev, ...updater },
    );
  }, []);

  const resetFilters = useCallback(() => {
    safeClear(key);
    setFiltersState(defaults);
  }, [key, defaults]);

  return [filters, setFilters, resetFilters];
}
