// Debugging: Global Error Handler to display errors on screen
window.onerror = function (message, source, lineno, colno, error) {
  const errorBox = document.createElement('div');
  errorBox.style.position = 'fixed';
  errorBox.style.top = '0';
  errorBox.style.left = '0';
  errorBox.style.width = '100%';
  errorBox.style.backgroundColor = 'red';
  errorBox.style.color = 'white';
  errorBox.style.zIndex = '99999';
  errorBox.style.padding = '20px';
  errorBox.style.whiteSpace = 'pre-wrap';
  errorBox.innerHTML = `<h1>Runtime Error</h1><p>${message}</p><p>${source}:${lineno}:${colno}</p><pre>${error?.stack || ''}</pre>`;
  document.body.appendChild(errorBox);
};

// Also catch unhandled promise rejections
window.addEventListener('unhandledrejection', function (event) {
  const errorBox = document.createElement('div');
  errorBox.style.position = 'fixed';
  errorBox.style.bottom = '0';
  errorBox.style.left = '0';
  errorBox.style.width = '100%';
  errorBox.style.backgroundColor = 'darkred';
  errorBox.style.color = 'white';
  errorBox.style.zIndex = '99999';
  errorBox.style.padding = '20px';
  errorBox.style.whiteSpace = 'pre-wrap';
  errorBox.innerHTML = `<h1>Unhandled Promise Rejection</h1><p>${event.reason}</p>`;
  document.body.appendChild(errorBox);
});

import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import './index.css'
import './styles/print.css'
import App from './App.tsx'

import './i18n'; // Import i18n configuration
import { Suspense } from 'react';
import { ThemeProvider } from './contexts/ThemeContext';

console.log('Mounting App...');

try {
  const rootElement = document.getElementById('root');
  if (!rootElement) {
    throw new Error("Root element 'root' not found in document");
  }
  createRoot(rootElement).render(
    <StrictMode>
      <ThemeProvider>
        <Suspense fallback={<div>Loading...</div>}>
          <App />
        </Suspense>
      </ThemeProvider>
    </StrictMode>,
  )
} catch (e) {
  console.error("Mounting error:", e);
  throw e;
}
