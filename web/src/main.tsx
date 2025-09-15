import React from 'react';
import { createRoot } from 'react-dom/client';
import { Provider } from 'react-redux';
import { App } from './routes/App';
import { store } from './store';

const container = document.getElementById('root');
if (!container) {
  throw new Error('Root container #root not found');
}
const root = createRoot(container);
root.render(
  <React.StrictMode>
    <Provider store={store}>
      <App />
    </Provider>
  </React.StrictMode>,
);
