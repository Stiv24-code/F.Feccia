import { configureStore } from '@reduxjs/toolkit';
import { appApi } from './api/appApi';
import themeReducer from './themeSlice';

export const store = configureStore({
  reducer: {
    [appApi.reducerPath]: appApi.reducer,
    theme: themeReducer,
  },
  middleware: (getDefaultMiddleware) => getDefaultMiddleware().concat(appApi.middleware),
});

export type RootState = ReturnType<typeof store.getState>;
export type AppDispatch = typeof store.dispatch;
