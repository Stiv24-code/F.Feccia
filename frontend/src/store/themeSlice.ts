import { createSlice } from '@reduxjs/toolkit';
import { getInitialTheme, getInitialGlass } from '@/lib/theme';

export type Theme = 'light' | 'dark';

export interface ThemeState {
  theme: Theme;
  // "Glass" (tema SBG del mockup "TMS Unificato"): indipendente da
  // light/dark, si compone con entrambi — non un terzo valore esclusivo di
  // Theme, vedi lib/theme.js applyTheme().
  glass: boolean;
}

const initialState: ThemeState = { theme: getInitialTheme(), glass: getInitialGlass() };

// Stato del tema centralizzato nello store — prima viveva come useState
// locale in AppShell (unico consumer) e ogni altro componente che aveva
// bisogno di sapere se il tema è scuro (es. OrderRouteMap, per scegliere le
// tile della mappa) doveva rileggerlo dal DOM (classList su <html>). Qui
// diventa leggibile ovunque con useAppSelector(s => s.theme.theme).
// L'effetto collaterale (classe "dark" su <html> + persistenza
// localStorage, in lib/theme.js) resta fuori dal reducer — applicato da un
// useEffect in AppShell che osserva questo stato, non dentro l'azione.
const themeSlice = createSlice({
  name: 'theme',
  initialState,
  reducers: {
    setTheme: (state, action: { payload: Theme }) => {
      state.theme = action.payload;
    },
    toggleTheme: (state) => {
      state.theme = state.theme === 'dark' ? 'light' : 'dark';
    },
    setGlass: (state, action: { payload: boolean }) => {
      state.glass = action.payload;
    },
    toggleGlass: (state) => {
      state.glass = !state.glass;
    },
  },
});

export const { setTheme, toggleTheme, setGlass, toggleGlass } = themeSlice.actions;
export default themeSlice.reducer;
