import { useEffect, useState } from 'react';

// Stato di paginazione condiviso dalle pagine Anagrafiche: pagina corrente
// (1-based, come il backend — vedi pkg/utils.PageParams) che si azzera a 1
// ogni volta che cambia `resetKey` (tipicamente il testo di ricerca), così
// un nuovo filtro non lascia l'utente bloccato su una pagina vuota.
export function usePagination(resetKey: unknown) {
  const [page, setPage] = useState(1);
  useEffect(() => { setPage(1); }, [resetKey]);
  return [page, setPage] as const;
}
