export type TicketStatus = 'draft' | 'pending_review' | 'approved' | 'rejected' | 'locked';
export type TicketType = 'time_materials' | 'variation';
export type UserRole = 'admin' | 'foreman' | 'gc_super' | 'pm' | 'accounting';
export type LineItemType = 'labor' | 'material' | 'equipment';

export interface User {
  id: string;
  email: string;
  firstName: string;
  lastName: string;
  role: UserRole;
  companyId: string;
  avatarUrl?: string;
  createdAt: string;
}

export interface Company {
  id: string;
  name: string;
  logoUrl?: string;
  address?: string;
  phone?: string;
}

export interface Project {
  id: string;
  companyId: string;
  name: string;
  address: string;
  gcCompanyName: string;
  gcContactName: string;
  gcContactEmail: string;
  gcContactPhone?: string;
  status: 'active' | 'completed' | 'on_hold';
  startDate: string;
  endDate?: string;
  ticketCount: number;
  totalValue: number;
}

export interface LineItem {
  id: string;
  ticketId: string;
  itemType: LineItemType;
  description: string;
  quantity: number;
  unit: string;
  unitPrice: number;
  total: number;
  sortOrder: number;
}

export interface Photo {
  id: string;
  ticketId: string;
  fileUrl: string;
  thumbnailUrl: string;
  caption?: string;
  takenAt: string;
  latitude?: number;
  longitude?: number;
}

export interface Signature {
  id: string;
  ticketId: string;
  signerType: 'foreman' | 'gc_super';
  signerName: string;
  signerEmail: string;
  signatureImageUrl: string;
  signingMethod: 'link_own_device' | 'link_same_device' | 'same_device_verified';
  signedAt: string;
}

export interface Ticket {
  id: string;
  companyId: string;
  projectId: string;
  project?: Project;
  createdById: string;
  createdBy?: User;
  ticketNumber: string;
  ticketType: TicketType;
  title: string;
  descriptionRaw?: string;
  descriptionCleaned?: string;
  workDate: string;
  status: TicketStatus;
  lineItems: LineItem[];
  photos: Photo[];
  signatures: Signature[];
  foremanSignature?: Signature;
  gcSignature?: Signature;
  submittedAt?: string;
  approvedAt?: string;
  rejectedAt?: string;
  lockedAt?: string;
  rejectionReason?: string;
  total: number;
  createdAt: string;
  updatedAt: string;
}

export interface DashboardStats {
  totalTickets: number;
  draftTickets: number;
  pendingTickets: number;
  approvedTickets: number;
  totalValue: number;
  thisMonthValue: number;
  recentActivity: ActivityItem[];
}

export interface ActivityItem {
  id: string;
  type: 'ticket_created' | 'ticket_submitted' | 'ticket_approved' | 'ticket_rejected' | 'photo_added' | 'signature_added';
  ticketId: string;
  ticketNumber: string;
  message: string;
  timestamp: string;
  userId: string;
  userName: string;
}
