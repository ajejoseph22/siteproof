/**
 * HTTP Client for SiteProof API
 * Handles authentication, error handling, and token refresh
 */

import { getApiBaseUrl } from './config';

interface TokenStore {
  accessToken: string | null;
  refreshToken: string | null;
  expiresAt: number | null;
}

let tokens: TokenStore = {
  accessToken: null,
  refreshToken: null,
  expiresAt: null,
};

// Load tokens from storage on init
if (typeof window !== 'undefined') {
  const stored = localStorage.getItem('siteproof-tokens');
  if (stored) {
    try {
      tokens = JSON.parse(stored);
    } catch {
      // Invalid stored data
    }
  }
}

export function setTokens(accessToken: string, refreshToken: string, expiresAt: string) {
  tokens = {
    accessToken,
    refreshToken,
    expiresAt: new Date(expiresAt).getTime(),
  };
  localStorage.setItem('siteproof-tokens', JSON.stringify(tokens));
}

export function clearTokens() {
  tokens = { accessToken: null, refreshToken: null, expiresAt: null };
  localStorage.removeItem('siteproof-tokens');
}

export function getAccessToken() {
  return tokens.accessToken;
}

export function isTokenExpired(): boolean {
  if (!tokens.expiresAt) return true;
  // Consider expired if less than 1 minute remaining
  return Date.now() > tokens.expiresAt - 60000;
}

async function refreshAccessToken(): Promise<boolean> {
  if (!tokens.refreshToken) return false;

  try {
    const response = await fetch(`${getApiBaseUrl()}/api/v1/auth/refresh`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ refresh_token: tokens.refreshToken }),
    });

    if (!response.ok) {
      clearTokens();
      return false;
    }

    const data = await response.json();
    setTokens(
      data.data.access_token,
      data.data.refresh_token,
      data.data.expires_at
    );
    return true;
  } catch {
    clearTokens();
    return false;
  }
}

export interface ApiError {
  code: string;
  message: string;
  details?: Record<string, unknown>;
}

export class ApiException extends Error {
  status: number;
  error: ApiError;

  constructor(status: number, error: ApiError) {
    super(error.message);
    this.name = 'ApiException';
    this.status = status;
    this.error = error;
  }
}

interface RequestOptions extends RequestInit {
  skipAuth?: boolean;
  idempotencyKey?: string;
}

export async function apiRequest<T>(
  endpoint: string,
  options: RequestOptions = {}
): Promise<T> {
  const { skipAuth = false, idempotencyKey, ...fetchOptions } = options;

  // Check if we need to refresh token
  if (!skipAuth && isTokenExpired() && tokens.refreshToken) {
    const refreshed = await refreshAccessToken();
    if (!refreshed) {
      throw new ApiException(401, {
        code: 'UNAUTHORIZED',
        message: 'Session expired. Please log in again.',
      });
    }
  }

  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...(fetchOptions.headers as Record<string, string>),
  };

  if (!skipAuth && tokens.accessToken) {
    headers['Authorization'] = `Bearer ${tokens.accessToken}`;
  }

  if (idempotencyKey) {
    headers['Idempotency-Key'] = idempotencyKey;
  }

  const response = await fetch(`${getApiBaseUrl()}${endpoint}`, {
    ...fetchOptions,
    headers,
  });

  // Handle empty responses (204 No Content)
  if (response.status === 204) {
    return undefined as T;
  }

  const data = await response.json();

  if (!response.ok) {
    throw new ApiException(response.status, data.error);
  }

  return data;
}

// Convenience methods
export const api = {
  get: <T>(endpoint: string, options?: RequestOptions) =>
    apiRequest<T>(endpoint, { ...options, method: 'GET' }),

  post: <T>(endpoint: string, body?: unknown, options?: RequestOptions) =>
    apiRequest<T>(endpoint, {
      ...options,
      method: 'POST',
      body: body ? JSON.stringify(body) : undefined,
    }),

  patch: <T>(endpoint: string, body: unknown, options?: RequestOptions) =>
    apiRequest<T>(endpoint, {
      ...options,
      method: 'PATCH',
      body: JSON.stringify(body),
    }),

  delete: <T>(endpoint: string, options?: RequestOptions) =>
    apiRequest<T>(endpoint, { ...options, method: 'DELETE' }),
};
