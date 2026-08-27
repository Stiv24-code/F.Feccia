import { configureStore } from '@reduxjs/toolkit';
import { setupListeners } from '@reduxjs/toolkit/query';
import { appApi } from './api/appApi';
import themeReducer from './themeSlice';

export const store = configureStore({
  reducer: {
    [appApi.reducerPath]: appApi.reducer,
    theme: themeReducer,
  },
  middleware: (getDefaultMiddleware) => getDefaultMiddleware().concat(appApi.middleware),
});

// Abilita refetchOnFocus/refetchOnReconnect per gli endpoint che li richiedono
// esplicitamente (es. getNavCounts) — nessun effetto sugli altri endpoint,
// che non li richiedono.
setupListeners(store.dispatch);

export type RootState = ReturnType<typeof store.getState>;
export type AppDispatch = typeof store.dispatch;
