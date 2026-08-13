import { createSlice } from '@reduxjs/toolkit';
import { getInitialTheme } from '@/lib/theme';

export type Theme = 'light' | 'dark';

export interface ThemeState {
  theme: Theme;
}

const initialState: ThemeState = { theme: getInitialTheme() };

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
  },
});

export const { setTheme, toggleTheme } = themeSlice.actions;
export default themeSlice.reducer;
