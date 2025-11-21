import React from 'react';
import ReactDOM from 'react-dom/client';
import App from './App';
import { loadConfig } from './utils/config';
import { apiClient } from './utils/api';

const root = ReactDOM.createRoot(
  document.getElementById('root') as HTMLElement
);

loadConfig().then(() => {
  apiClient.init();
  root.render(
    <React.StrictMode>
      <App />
    </React.StrictMode>
  );
});
