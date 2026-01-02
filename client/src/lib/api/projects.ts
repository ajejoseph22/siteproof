/**
 * Projects API
 */

import { api } from './client';
import { isDemoMode } from './config';
import { mockProjects } from '@/data/mock';
import type { Project } from '@/types';

interface ApiProject {
  id: string;
  company_id: string;
  name: string;
  address: string | null;
  gc_company_name: string | null;
  gc_contact_email: string | null;
  status: string;
  timezone: string;
  created_at: string;
  updated_at: string;
}

interface ProjectListResponse {
  data: ApiProject[];
}

interface ProjectResponse {
  data: ApiProject;
}

// Transform API project to app Project type
function transformProject(apiProject: ApiProject): Project {
  return {
    id: apiProject.id,
    companyId: apiProject.company_id,
    name: apiProject.name,
    address: apiProject.address || '',
    gcCompanyName: apiProject.gc_company_name || '',
    gcContactName: '', // Not in API, would need separate endpoint or extended schema
    gcContactEmail: apiProject.gc_contact_email || '',
    gcContactPhone: undefined,
    status: apiProject.status as Project['status'],
    startDate: apiProject.created_at,
    endDate: undefined,
    ticketCount: 0, // Would need to be calculated or added to API
    totalValue: 0,
  };
}

export async function listProjects(status?: string): Promise<Project[]> {
  if (isDemoMode()) {
    await new Promise(resolve => setTimeout(resolve, 300));
    if (status && status !== 'all') {
      return mockProjects.filter(p => p.status === status);
    }
    return mockProjects;
  }

  try {
    const params = new URLSearchParams();
    if (status && status !== 'all') {
      params.set('status', status);
    }

    const endpoint = `/api/v1/projects${params.toString() ? `?${params}` : ''}`;
    const response = await api.get<ProjectListResponse>(endpoint);
    return response.data.map(transformProject);
  } catch (error) {
    console.error('Failed to fetch projects:', error);
    return [];
  }
}

export async function getProject(id: string): Promise<Project | null> {
  if (isDemoMode()) {
    await new Promise(resolve => setTimeout(resolve, 200));
    return mockProjects.find(p => p.id === id) || null;
  }

  try {
    const response = await api.get<ProjectResponse>(`/api/v1/projects/${id}`);
    return transformProject(response.data);
  } catch {
    return null;
  }
}

interface CreateProjectRequest {
  name: string;
  address?: string;
  gc_company_name?: string;
  gc_contact_email?: string;
  timezone: string;
}

export async function createProject(data: CreateProjectRequest): Promise<Project | null> {
  if (isDemoMode()) {
    await new Promise(resolve => setTimeout(resolve, 500));
    const newProject: Project = {
      id: crypto.randomUUID(),
      companyId: 'demo-company',
      name: data.name,
      address: data.address || '',
      gcCompanyName: data.gc_company_name || '',
      gcContactName: '',
      gcContactEmail: data.gc_contact_email || '',
      status: 'active',
      startDate: new Date().toISOString(),
      ticketCount: 0,
      totalValue: 0,
    };
    return newProject;
  }

  try {
    const response = await api.post<ProjectResponse>('/api/v1/projects', data, {
      idempotencyKey: crypto.randomUUID(),
    });
    return transformProject(response.data);
  } catch {
    return null;
  }
}
