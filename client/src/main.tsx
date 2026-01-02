import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';
import './index.css';
import App from './App';

// Initialize dark mode from localStorage
const stored = localStorage.getItem('siteproof-storage');
if (stored) {
  try {
    const parsed = JSON.parse(stored);
    if (parsed?.state?.isDarkMode) {
      document.documentElement.classList.add('dark');
    }
  } catch {
    // Ignore parse errors
  }
}

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <App />
  </StrictMode>
);
