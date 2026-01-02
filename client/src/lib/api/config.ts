/**
 * API Configuration
 * Demo mode is controlled by VITE_DEMO_MODE environment variable
 * Set to 'false' to use the real API
 */

export const API_CONFIG = {
  baseUrl: import.meta.env.VITE_API_URL || 'http://localhost:8080',
  demoMode: import.meta.env.VITE_DEMO_MODE !== 'false', // Default to true
};

export const isDemoMode = () => API_CONFIG.demoMode;
export const getApiBaseUrl = () => API_CONFIG.baseUrl;
