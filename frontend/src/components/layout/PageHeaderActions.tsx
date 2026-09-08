import { createContext, useContext, type ReactNode } from 'react';
import { createPortal } from 'react-dom';

// La testa di pagina è una fascia unica in cima (vedi AppShell.js e
// ".glass [data-glass='header']" in index.css) e ha due punti d'innesto, come
// nel prototipo "TMS Unificato":
//
//   <titolo>  <meta>                              <azioni>
//
// - `meta`: accanto al titolo. Nel prototipo la usa solo la Dashboard, per la
//   data corrente — le altre pagine hanno il titolo nudo, quindi la data NON
//   va messa qui in AppShell per tutti.
// - `actions`: a destra, l'azione primaria della pagina ("+ Nuovo ordine" in
//   Ordini, "Carica PDF" in Dashboard, "Salva template" nell'editor PDF).
//
// Le pagine le dichiarano dove sta la loro logica e il contenuto atterra nella
// fascia via portale: nessuna prop da far scendere attraverso il router, e
// nessuna mappa path → azione in AppShell che si scollerebbe dalle pagine.
interface HeaderSlots {
  meta: HTMLElement | null;
  actions: HTMLElement | null;
}

// `undefined` = nessuna shell con la fascia (non deve capitare, ma se capita
// il contenuto va mostrato in linea invece di sparire); un nodo `null` =
// shell presente, nodo non ancora montato. Stessa distinzione di PageSlab.tsx.
const HeaderSlotContext = createContext<HeaderSlots | undefined>(undefined);

export const HeaderSlotProvider = ({
  meta,
  actions,
  children,
}: {
  meta: HTMLElement | null;
  actions: HTMLElement | null;
  children: ReactNode;
}) => (
  <HeaderSlotContext.Provider value={{ meta, actions }}>{children}</HeaderSlotContext.Provider>
);

const Slot = ({ which, children }: { which: keyof HeaderSlots; children: ReactNode }) => {
  const slots = useContext(HeaderSlotContext);
  if (slots === undefined) return <>{children}</>;
  const node = slots[which];
  if (node === null) return null;
  return createPortal(children, node);
};

/** Accanto al titolo. Nel design la usa solo la Dashboard (data corrente). */
export const PageHeaderMeta = ({ children }: { children: ReactNode }) => (
  <Slot which="meta">{children}</Slot>
);

/** A destra nella fascia: l'azione primaria della pagina. */
export const PageHeaderActions = ({ children }: { children: ReactNode }) => (
  <Slot which="actions">{children}</Slot>
);
