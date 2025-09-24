import { configureStore, createSlice } from '@reduxjs/toolkit';
import { api } from './api';
import bookSpreadReducer from './store/bookSpreadSlice';

const uiSlice = createSlice({
  name: 'ui',
  initialState: { toasts: [] as { id: string; text: string; type?: 'info' | 'error' }[] },
  reducers: {
    addToast: (s, a) => {
      s.toasts.push(a.payload);
    },
    removeToast: (s, a) => {
      s.toasts = s.toasts.filter((t) => t.id !== a.payload);
    },
  },
});

export const store = configureStore({
  reducer: {
    ui: uiSlice.reducer,
    bookSpread: bookSpreadReducer,
    [api.reducerPath]: api.reducer,
  },
  middleware: (gDM) => gDM().concat(api.middleware),
});

export type RootState = ReturnType<typeof store.getState>;
export type AppDispatch = typeof store.dispatch;
