import { createContext, useContext, useState, type ReactNode } from 'react';
import { createPortal } from 'react-dom';

// Nel prototipo tutta la zona di testa di una pagina-registro — tab, ricerca,
// filtri, chip di stato — vive dentro una sola lastra continua, e la tabella
// comincia staccata sotto. Prima le tab galleggiavano sul fondale e i filtri
// stavano per conto loro: tre piani diversi al posto di uno (rilievo 02
// dell'audit, valido per Ordini, Planner e Template PDF).
//
// Le tab però stanno in un layout di route (OrdiniTabsLayout) e i filtri nella
// pagina figlia: non possono finire nello stesso elemento senza spostare
// logica da una parte all'altra. Quindi la lastra espone uno slot e la pagina
// ci porta la sua toolbar via portale — `<SlabToolbar>` si dichiara dove sta
// la logica dei filtri e rende dove sta il disegno.
//
// La lastra è `data-glass="container"`: nel tema Glass è il livello più
// trasparente dei tre, e i pannelli che ci stanno sopra sono più pieni (vedi
// index.css).
// `undefined` = nessuna lastra nell'albero; `null` = lastra presente ma nodo
// non ancora montato. I due casi vanno distinti: nel secondo bisogna non
// rendere nulla e aspettare, altrimenti la toolbar comparirebbe in linea per
// un frame e poi salterebbe dentro la lastra.
const SlabContext = createContext<HTMLElement | null | undefined>(undefined);

export interface PageSlabProps {
  /** Striscia di tab, quando la pagina vive dentro un layout che le ha. */
  tabs?: ReactNode;
  children: ReactNode;
}

export const PageSlab = ({ tabs, children }: PageSlabProps) => {
  // Stato e non ref: il portale deve ri-renderizzare quando il nodo compare.
  const [toolbarNode, setToolbarNode] = useState<HTMLElement | null>(null);

  return (
    <SlabContext.Provider value={toolbarNode}>
      <div className="space-y-3">
        {/* Una fascia sola, non una card: prosegue la testa di pagina di
            AppShell senza stacchi — niente angoli, niente bordo, niente
            ombra, e i margini negativi annullano il padding del <main> per
            arrivare a filo dei due lati e della testa. Titolo + CTA, tab e
            filtri stanno così sullo stesso piano, e la tabella comincia
            staccata sotto. */}
        <div
          data-glass="header"
          data-testid="page-slab"
          className="-mt-4 -mx-3 sm:-mx-4 lg:-mx-6 px-3 sm:px-4 lg:px-6 bg-card"
        >
          {tabs && <div className="pt-2 pb-1.5">{tabs}</div>}
          {/* `empty:hidden`: senza toolbar la fascia non deve lasciare una
              banda vuota sotto le tab. Nessun divisore fra le due righe — nel
              design sono lo stesso piano. */}
          <div
            ref={setToolbarNode}
            data-glass="toolbar"
            data-testid="page-slab-toolbar"
            className="py-2.5 empty:hidden"
          />
        </div>
        {children}
      </div>
    </SlabContext.Provider>
  );
};

// Da usare dentro un <PageSlab>: il contenuto atterra nella lastra. Fuori da
// una lastra rende in linea, così una pagina che non ne ha una non perde la
// propria toolbar.
export const SlabToolbar = ({ children }: { children: ReactNode }) => {
  const node = useContext(SlabContext);
  if (node === undefined) return <>{children}</>;
  if (node === null) return null;
  return createPortal(children, node);
};
