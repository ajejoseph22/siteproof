/**
 * Tickets API
 */

import { api } from './client';
import { isDemoMode } from './config';
import { mockTickets, mockProjects } from '@/data/mock';
import type { Ticket, TicketStatus, TicketType, LineItem, Photo } from '@/types';

interface ApiTicket {
  id: string;
  company_id: string;
  project_id: string;
  created_by_id: string;
  ticket_number: string;
  ticket_type: string;
  title: string;
  description_raw: string | null;
  description_cleaned: string | null;
  description_final: string | null;
  work_date: string;
  status: string;
  gc_signed_at: string | null;
  gc_signed_by_id: string | null;
  rejected_at: string | null;
  rejection_reason: string | null;
  pdf_path: string | null;
  locked_at: string | null;
  created_at: string;
  updated_at: string;
}

interface ApiLineItem {
  id: string;
  ticket_id: string;
  item_type: string;
  description: string;
  quantity: number;
  unit: string;
  unit_rate: number | null;
  total: number | null;
  sort_order: number;
  created_at: string;
}

interface ApiPhoto {
  id: string;
  ticket_id: string;
  file_url: string;
  thumbnail_url: string;
  caption: string | null;
  taken_at: string | null;
  latitude: number | null;
  longitude: number | null;
}

interface TicketListResponse {
  data: ApiTicket[];
}

interface TicketDetailResponse {
  data: ApiTicket & {
    line_items: ApiLineItem[];
    photos: ApiPhoto[];
  };
}

interface TicketResponse {
  data: ApiTicket;
}

function transformTicket(apiTicket: ApiTicket, lineItems?: LineItem[], photos?: Photo[]): Ticket {
  const project = mockProjects.find(p => p.id === apiTicket.project_id);

  return {
    id: apiTicket.id,
    companyId: apiTicket.company_id,
    projectId: apiTicket.project_id,
    project: project,
    createdById: apiTicket.created_by_id,
    ticketNumber: apiTicket.ticket_number,
    ticketType: apiTicket.ticket_type as TicketType,
    title: apiTicket.title,
    descriptionRaw: apiTicket.description_raw || undefined,
    descriptionCleaned: apiTicket.description_cleaned || undefined,
    workDate: apiTicket.work_date,
    status: apiTicket.status as TicketStatus,
    lineItems: lineItems || [],
    photos: photos || [],
    signatures: [],
    submittedAt: undefined,
    approvedAt: apiTicket.gc_signed_at || undefined,
    rejectedAt: apiTicket.rejected_at || undefined,
    lockedAt: apiTicket.locked_at || undefined,
    rejectionReason: apiTicket.rejection_reason || undefined,
    total: lineItems?.reduce((sum, item) => sum + item.total, 0) || 0,
    createdAt: apiTicket.created_at,
    updatedAt: apiTicket.updated_at,
  };
}

function transformLineItem(apiItem: ApiLineItem): LineItem {
  return {
    id: apiItem.id,
    ticketId: apiItem.ticket_id,
    itemType: apiItem.item_type as LineItem['itemType'],
    description: apiItem.description,
    quantity: apiItem.quantity,
    unit: apiItem.unit,
    unitPrice: apiItem.unit_rate || 0,
    total: apiItem.total || 0,
    sortOrder: apiItem.sort_order,
  };
}

function transformPhoto(apiPhoto: ApiPhoto): Photo {
  return {
    id: apiPhoto.id,
    ticketId: apiPhoto.ticket_id,
    fileUrl: apiPhoto.file_url,
    thumbnailUrl: apiPhoto.thumbnail_url,
    caption: apiPhoto.caption || undefined,
    takenAt: apiPhoto.taken_at || new Date().toISOString(),
    latitude: apiPhoto.latitude || undefined,
    longitude: apiPhoto.longitude || undefined,
  };
}

export async function listTickets(filters?: {
  projectId?: string;
  status?: TicketStatus | 'all';
}): Promise<Ticket[]> {
  if (isDemoMode()) {
    await new Promise(resolve => setTimeout(resolve, 300));
    let tickets = [...mockTickets];

    if (filters?.projectId) {
      tickets = tickets.filter(t => t.projectId === filters.projectId);
    }
    if (filters?.status && filters.status !== 'all') {
      tickets = tickets.filter(t => t.status === filters.status);
    }

    return tickets;
  }

  try {
    const params = new URLSearchParams();
    if (filters?.status && filters.status !== 'all') {
      params.set('status', filters.status);
    }

    let endpoint: string;
    if (filters?.projectId) {
      endpoint = `/api/v1/projects/${filters.projectId}/tickets`;
    } else {
      // Note: API may not have a global tickets list - adjust as needed
      endpoint = '/api/v1/tickets';
    }

    if (params.toString()) {
      endpoint += `?${params}`;
    }

    const response = await api.get<TicketListResponse>(endpoint);
    return response.data.map(t => transformTicket(t));
  } catch (error) {
    console.error('Failed to fetch tickets:', error);
    return [];
  }
}

export async function getTicket(id: string): Promise<Ticket | null> {
  if (isDemoMode()) {
    await new Promise(resolve => setTimeout(resolve, 200));
    return mockTickets.find(t => t.id === id) || null;
  }

  try {
    const response = await api.get<TicketDetailResponse>(`/api/v1/tickets/${id}`);
    const lineItems = response.data.line_items.map(transformLineItem);
    const photos = response.data.photos.map(transformPhoto);
    return transformTicket(response.data, lineItems, photos);
  } catch {
    return null;
  }
}

interface CreateTicketRequest {
  project_id: string;
  ticket_type: TicketType;
  title: string;
  work_date: string;
}

export async function createTicket(data: CreateTicketRequest): Promise<Ticket | null> {
  if (isDemoMode()) {
    await new Promise(resolve => setTimeout(resolve, 500));
    const project = mockProjects.find(p => p.id === data.project_id);
    const ticketNumber = `TM-2024-${String(mockTickets.length + 1).padStart(4, '0')}`;

    const newTicket: Ticket = {
      id: crypto.randomUUID(),
      companyId: 'demo-company',
      projectId: data.project_id,
      project: project,
      createdById: 'demo-user',
      ticketNumber,
      ticketType: data.ticket_type,
      title: data.title,
      workDate: data.work_date,
      status: 'draft',
      lineItems: [],
      photos: [],
      signatures: [],
      total: 0,
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
    };

    return newTicket;
  }

  try {
    const response = await api.post<TicketResponse>('/api/v1/tickets', data, {
      idempotencyKey: crypto.randomUUID(),
    });
    return transformTicket(response.data);
  } catch {
    return null;
  }
}

interface UpdateTicketRequest {
  title?: string;
  description_final?: string;
  work_date?: string;
}

export async function updateTicket(id: string, data: UpdateTicketRequest): Promise<Ticket | null> {
  if (isDemoMode()) {
    await new Promise(resolve => setTimeout(resolve, 300));
    const ticket = mockTickets.find(t => t.id === id);
    if (!ticket) return null;

    return {
      ...ticket,
      title: data.title || ticket.title,
      workDate: data.work_date || ticket.workDate,
      updatedAt: new Date().toISOString(),
    };
  }

  try {
    const response = await api.patch<TicketResponse>(`/api/v1/tickets/${id}`, data);
    return transformTicket(response.data);
  } catch {
    return null;
  }
}

export async function submitTicket(id: string): Promise<Ticket | null> {
  if (isDemoMode()) {
    await new Promise(resolve => setTimeout(resolve, 500));
    const ticket = mockTickets.find(t => t.id === id);
    if (!ticket) return null;

    return {
      ...ticket,
      status: 'pending_review',
      submittedAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
    };
  }

  try {
    const response = await api.post<TicketResponse>(`/api/v1/tickets/${id}/submit`);
    return transformTicket(response.data);
  } catch {
    return null;
  }
}

// Line Items

interface CreateLineItemRequest {
  item_type: LineItem['itemType'];
  description: string;
  quantity: number;
  unit: string;
  unit_rate?: number;
}

export async function addLineItem(ticketId: string, data: CreateLineItemRequest): Promise<LineItem | null> {
  if (isDemoMode()) {
    await new Promise(resolve => setTimeout(resolve, 200));
    return {
      id: crypto.randomUUID(),
      ticketId,
      itemType: data.item_type,
      description: data.description,
      quantity: data.quantity,
      unit: data.unit,
      unitPrice: data.unit_rate || 0,
      total: data.quantity * (data.unit_rate || 0),
      sortOrder: 0,
    };
  }

  try {
    const response = await api.post<{ data: ApiLineItem }>(
      `/api/v1/tickets/${ticketId}/line-items`,
      data,
      { idempotencyKey: crypto.randomUUID() }
    );
    return transformLineItem(response.data);
  } catch {
    return null;
  }
}

export async function deleteLineItem(ticketId: string, lineItemId: string): Promise<boolean> {
  if (isDemoMode()) {
    await new Promise(resolve => setTimeout(resolve, 200));
    return true;
  }

  try {
    await api.delete(`/api/v1/tickets/${ticketId}/line-items/${lineItemId}`);
    return true;
  } catch {
    return false;
  }
}

// Signature

interface SignatureRequestResponse {
  data: {
    signing_link_id: string;
    expires_at: string;
    sent_via: string[];
  };
}

export async function requestSignature(ticketId: string, signerEmail: string, signerPhone: string): Promise<SignatureRequestResponse['data'] | null> {
  if (isDemoMode()) {
    await new Promise(resolve => setTimeout(resolve, 1000));
    return {
      signing_link_id: crypto.randomUUID(),
      expires_at: new Date(Date.now() + 24 * 60 * 60 * 1000).toISOString(),
      sent_via: ['sms', 'email'],
    };
  }

  try {
    const response = await api.post<SignatureRequestResponse>(
      `/api/v1/tickets/${ticketId}/signature-request`,
      { signer_email: signerEmail, signer_phone: signerPhone }
    );
    return response.data;
  } catch {
    return null;
  }
}

// Transcription cleanup

interface CleanupResponse {
  data: {
    cleaned_text: string;
    raw_preserved: boolean;
  };
}

export async function cleanupTranscription(ticketId: string, rawText: string): Promise<string | null> {
  if (isDemoMode()) {
    await new Promise(resolve => setTimeout(resolve, 1500));
    // Simple demo cleanup
    return rawText
      .replace(/\b(uh|um|like|you know|basically)\b/gi, '')
      .replace(/\s+/g, ' ')
      .trim();
  }

  try {
    const response = await api.post<CleanupResponse>(
      `/api/v1/tickets/${ticketId}/description/cleanup`,
      { raw_text: rawText }
    );
    return response.data.cleaned_text;
  } catch {
    return null;
  }
}
