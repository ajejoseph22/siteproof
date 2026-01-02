/**
 * Authentication API
 */

import { api, setTokens, clearTokens } from './client';
import { isDemoMode } from './config';
import { mockUser, mockCompany } from '@/data/mock';
import type { User, Company } from '@/types';

interface LoginRequest {
  email: string;
  password: string;
}

interface LoginResponse {
  data: {
    user: {
      id: string;
      company_id: string;
      email: string;
      name: string;
      phone: string | null;
      role: string;
      active: boolean;
      created_at: string;
      updated_at: string;
    };
    tokens: {
      access_token: string;
      refresh_token: string;
      expires_at: string;
    };
  };
}

interface CurrentUserResponse {
  data: {
    user_id: string;
    company_id: string;
    role: string;
  };
}

// Transform API user to app User type
function transformUser(apiUser: LoginResponse['data']['user']): User {
  const [firstName, ...lastNameParts] = apiUser.name.split(' ');
  return {
    id: apiUser.id,
    email: apiUser.email,
    firstName: firstName || '',
    lastName: lastNameParts.join(' ') || '',
    role: apiUser.role as User['role'],
    companyId: apiUser.company_id,
    createdAt: apiUser.created_at,
  };
}

export async function login(email: string, password: string): Promise<{
  user: User;
  company: Company;
} | null> {
  // Demo mode - return mock data
  if (isDemoMode()) {
    await new Promise(resolve => setTimeout(resolve, 800));
    if (email) {
      return {
        user: mockUser,
        company: mockCompany,
      };
    }
    return null;
  }

  // Real API call
  try {
    const response = await api.post<LoginResponse>('/api/v1/auth/login', {
      email,
      password,
    } as LoginRequest, { skipAuth: true });

    setTokens(
      response.data.tokens.access_token,
      response.data.tokens.refresh_token,
      response.data.tokens.expires_at
    );

    // TODO: Fetch company details from API
    // For now, use a placeholder
    const company: Company = {
      id: response.data.user.company_id,
      name: 'Your Company', // This should come from a company endpoint
    };

    return {
      user: transformUser(response.data.user),
      company,
    };
  } catch (error) {
    console.error('Login failed:', error);
    return null;
  }
}

export async function logout(): Promise<void> {
  if (isDemoMode()) {
    clearTokens();
    return;
  }

  try {
    await api.post('/api/v1/auth/logout');
  } catch {
    // Ignore logout errors
  } finally {
    clearTokens();
  }
}

export async function getCurrentUser(): Promise<CurrentUserResponse['data'] | null> {
  if (isDemoMode()) {
    return {
      user_id: mockUser.id,
      company_id: mockUser.companyId,
      role: mockUser.role,
    };
  }

  try {
    const response = await api.get<CurrentUserResponse>('/api/v1/auth/me');
    return response.data;
  } catch {
    return null;
  }
}
