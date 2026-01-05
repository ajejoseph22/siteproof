package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ajejoseph22/siteproof/internal/domain"
)

// TicketService handles ticket operations.
type TicketService struct {
	pool *pgxpool.Pool
}

// NewTicketService creates a new ticket service.
func NewTicketService(pool *pgxpool.Pool) *TicketService {
	return &TicketService{pool: pool}
}

// Create creates a new ticket.
func (s *TicketService) Create(ctx context.Context, input *domain.CreateTicketInput) (*domain.Ticket, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	// Get next ticket number
	var nextNumber int
	err := s.pool.QueryRow(ctx, `
		SELECT COALESCE(MAX(CAST(SUBSTRING(ticket_number FROM '[0-9]+$') AS INTEGER)), 0) + 1
		FROM tickets WHERE company_id = $1
	`, input.CompanyID).Scan(&nextNumber)
	if err != nil {
		return nil, fmt.Errorf("get next ticket number: %w", err)
	}

	// Generate ticket number based on type
	prefix := "TM"
	if input.TicketType == domain.TicketTypeVariation {
		prefix = "VAR"
	}
	ticketNumber := fmt.Sprintf("%s-%d-%04d", prefix, time.Now().Year(), nextNumber)

	// Create ticket
	var ticket domain.Ticket
	err = s.pool.QueryRow(ctx, `
		INSERT INTO tickets (
			company_id, project_id, created_by_id, ticket_number, ticket_type,
			title, work_date, status
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, 'draft')
		RETURNING id, company_id, project_id, created_by_id, ticket_number, ticket_type,
		          title, description_raw, description_cleaned, description_final,
		          work_date, status, created_at, updated_at
	`, input.CompanyID, input.ProjectID, input.CreatedByID, ticketNumber,
		input.TicketType, input.Title, input.WorkDate).Scan(
		&ticket.ID, &ticket.CompanyID, &ticket.ProjectID, &ticket.CreatedByID,
		&ticket.TicketNumber, &ticket.TicketType, &ticket.Title,
		&ticket.DescriptionRaw, &ticket.DescriptionCleaned, &ticket.DescriptionFinal,
		&ticket.WorkDate, &ticket.Status, &ticket.CreatedAt, &ticket.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create ticket: %w", err)
	}

	return &ticket, nil
}

// GetByID retrieves a ticket by ID.
func (s *TicketService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Ticket, error) {
	var ticket domain.Ticket
	err := s.pool.QueryRow(ctx, `
		SELECT id, company_id, project_id, created_by_id, ticket_number, ticket_type,
		       title, description_raw, description_cleaned, description_final,
		       work_date, status, gc_signed_at, gc_signed_by_id,
		       rejected_at, rejected_by_id, rejection_reason,
		       pdf_generated_at, pdf_path, pdf_hash, locked_at, content_hash,
		       created_at, updated_at
		FROM tickets WHERE id = $1
	`, id).Scan(
		&ticket.ID, &ticket.CompanyID, &ticket.ProjectID, &ticket.CreatedByID,
		&ticket.TicketNumber, &ticket.TicketType, &ticket.Title,
		&ticket.DescriptionRaw, &ticket.DescriptionCleaned, &ticket.DescriptionFinal,
		&ticket.WorkDate, &ticket.Status, &ticket.GCSignedAt, &ticket.GCSignedByID,
		&ticket.RejectedAt, &ticket.RejectedByID, &ticket.RejectionReason,
		&ticket.PDFGeneratedAt, &ticket.PDFPath, &ticket.PDFHash,
		&ticket.LockedAt, &ticket.ContentHash, &ticket.CreatedAt, &ticket.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get ticket: %w", err)
	}

	return &ticket, nil
}

// ListByProject retrieves tickets for a project.
func (s *TicketService) ListByProject(ctx context.Context, projectID uuid.UUID, limit, offset int) ([]*domain.Ticket, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, company_id, project_id, created_by_id, ticket_number, ticket_type,
		       title, work_date, status, created_at, updated_at
		FROM tickets
		WHERE project_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, projectID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list tickets: %w", err)
	}
	defer rows.Close()

	var tickets []*domain.Ticket
	for rows.Next() {
		var t domain.Ticket
		if err := rows.Scan(
			&t.ID, &t.CompanyID, &t.ProjectID, &t.CreatedByID,
			&t.TicketNumber, &t.TicketType, &t.Title, &t.WorkDate,
			&t.Status, &t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan ticket: %w", err)
		}
		tickets = append(tickets, &t)
	}

	return tickets, nil
}

// Update updates a ticket (only drafts can be updated).
func (s *TicketService) Update(ctx context.Context, id uuid.UUID, input *domain.UpdateTicketInput) (*domain.Ticket, error) {
	// First check if ticket is editable
	ticket, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if !ticket.CanEdit() {
		return nil, domain.ErrTicketLocked
	}

	var updated domain.Ticket
	err = s.pool.QueryRow(ctx, `
		UPDATE tickets
		SET title = COALESCE($2, title),
		    description_final = COALESCE($3, description_final),
		    work_date = COALESCE($4, work_date),
		    updated_at = now()
		WHERE id = $1 AND status = 'draft'
		RETURNING id, company_id, project_id, created_by_id, ticket_number, ticket_type,
		          title, description_raw, description_cleaned, description_final,
		          work_date, status, created_at, updated_at
	`, id, input.Title, input.DescriptionFinal, input.WorkDate).Scan(
		&updated.ID, &updated.CompanyID, &updated.ProjectID, &updated.CreatedByID,
		&updated.TicketNumber, &updated.TicketType, &updated.Title,
		&updated.DescriptionRaw, &updated.DescriptionCleaned, &updated.DescriptionFinal,
		&updated.WorkDate, &updated.Status, &updated.CreatedAt, &updated.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("update ticket: %w", err)
	}

	return &updated, nil
}

// SubmitForReview changes ticket status to pending_review.
func (s *TicketService) SubmitForReview(ctx context.Context, id uuid.UUID) (*domain.Ticket, error) {
	ticket, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if !ticket.Status.CanTransitionTo(domain.TicketStatusPendingReview) {
		return nil, domain.ErrInvalidTransition
	}

	var updated domain.Ticket
	err = s.pool.QueryRow(ctx, `
		UPDATE tickets
		SET status = 'pending_review', updated_at = now()
		WHERE id = $1
		RETURNING id, company_id, project_id, created_by_id, ticket_number, ticket_type,
		          title, work_date, status, created_at, updated_at
	`, id).Scan(
		&updated.ID, &updated.CompanyID, &updated.ProjectID, &updated.CreatedByID,
		&updated.TicketNumber, &updated.TicketType, &updated.Title,
		&updated.WorkDate, &updated.Status, &updated.CreatedAt, &updated.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("submit ticket: %w", err)
	}

	return &updated, nil
}

// Approve approves a ticket.
func (s *TicketService) Approve(ctx context.Context, id, signerID uuid.UUID) (*domain.Ticket, error) {
	ticket, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if !ticket.Status.CanTransitionTo(domain.TicketStatusApproved) {
		return nil, domain.ErrInvalidTransition
	}

	var updated domain.Ticket
	err = s.pool.QueryRow(ctx, `
		UPDATE tickets
		SET status = 'approved',
		    gc_signed_at = now(),
		    gc_signed_by_id = $2,
		    updated_at = now()
		WHERE id = $1
		RETURNING id, company_id, project_id, created_by_id, ticket_number, ticket_type,
		          title, work_date, status, gc_signed_at, gc_signed_by_id,
		          created_at, updated_at
	`, id, signerID).Scan(
		&updated.ID, &updated.CompanyID, &updated.ProjectID, &updated.CreatedByID,
		&updated.TicketNumber, &updated.TicketType, &updated.Title,
		&updated.WorkDate, &updated.Status, &updated.GCSignedAt, &updated.GCSignedByID,
		&updated.CreatedAt, &updated.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("approve ticket: %w", err)
	}

	return &updated, nil
}

// Reject rejects a ticket.
func (s *TicketService) Reject(ctx context.Context, id, rejectorID uuid.UUID, reason string) (*domain.Ticket, error) {
	ticket, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if !ticket.Status.CanTransitionTo(domain.TicketStatusRejected) {
		return nil, domain.ErrInvalidTransition
	}

	if reason == "" {
		return nil, &domain.ValidationError{Field: "reason", Message: "rejection reason is required"}
	}

	var updated domain.Ticket
	err = s.pool.QueryRow(ctx, `
		UPDATE tickets
		SET status = 'rejected',
		    rejected_at = now(),
		    rejected_by_id = $2,
		    rejection_reason = $3,
		    updated_at = now()
		WHERE id = $1
		RETURNING id, company_id, project_id, created_by_id, ticket_number, ticket_type,
		          title, work_date, status, rejected_at, rejected_by_id, rejection_reason,
		          created_at, updated_at
	`, id, rejectorID, reason).Scan(
		&updated.ID, &updated.CompanyID, &updated.ProjectID, &updated.CreatedByID,
		&updated.TicketNumber, &updated.TicketType, &updated.Title,
		&updated.WorkDate, &updated.Status, &updated.RejectedAt, &updated.RejectedByID,
		&updated.RejectionReason, &updated.CreatedAt, &updated.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("reject ticket: %w", err)
	}

	return &updated, nil
}

// Lock locks a ticket with PDF information.
func (s *TicketService) Lock(ctx context.Context, id uuid.UUID, pdfPath, pdfHash, contentHash string) (*domain.Ticket, error) {
	ticket, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if !ticket.Status.CanTransitionTo(domain.TicketStatusLocked) {
		return nil, domain.ErrInvalidTransition
	}

	var updated domain.Ticket
	err = s.pool.QueryRow(ctx, `
		UPDATE tickets
		SET status = 'locked',
		    locked_at = now(),
		    pdf_generated_at = now(),
		    pdf_path = $2,
		    pdf_hash = $3,
		    content_hash = $4,
		    updated_at = now()
		WHERE id = $1
		RETURNING id, company_id, project_id, created_by_id, ticket_number, ticket_type,
		          title, work_date, status, locked_at, pdf_path, pdf_hash, content_hash,
		          created_at, updated_at
	`, id, pdfPath, pdfHash, contentHash).Scan(
		&updated.ID, &updated.CompanyID, &updated.ProjectID, &updated.CreatedByID,
		&updated.TicketNumber, &updated.TicketType, &updated.Title,
		&updated.WorkDate, &updated.Status, &updated.LockedAt, &updated.PDFPath,
		&updated.PDFHash, &updated.ContentHash, &updated.CreatedAt, &updated.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("lock ticket: %w", err)
	}

	return &updated, nil
}
