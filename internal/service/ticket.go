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

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("begin ticket create: %w", err)
	}
	defer tx.Rollback(ctx)

	var projectExists bool
	err = tx.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM projects WHERE id = $1 AND company_id = $2)
	`, input.ProjectID, input.CompanyID).Scan(&projectExists)
	if err != nil {
		return nil, fmt.Errorf("check project: %w", err)
	}
	if !projectExists {
		return nil, domain.ErrNotFound
	}

	// Get next ticket number
	var nextNumber int
	err = tx.QueryRow(ctx, `
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
	err = tx.QueryRow(ctx, `
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

	changes := map[string]domain.ChangeValue{
		"status":        {Old: nil, New: ticket.Status},
		"title":         {Old: nil, New: ticket.Title},
		"ticket_type":   {Old: nil, New: ticket.TicketType},
		"ticket_number": {Old: nil, New: ticket.TicketNumber},
		"work_date":     {Old: nil, New: ticket.WorkDate},
	}
	if err := s.insertAuditLog(ctx, tx, domain.CreateAuditLogInput{
		EntityType: "ticket",
		EntityID:   ticket.ID,
		Action:     domain.AuditActionCreated,
		ActorID:    &input.CreatedByID,
		Changes:    changes,
	}); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit ticket create: %w", err)
	}

	return &ticket, nil
}

// GetByID retrieves a ticket by ID.
func (s *TicketService) GetByID(ctx context.Context, companyID, id uuid.UUID) (*domain.Ticket, error) {
	var ticket domain.Ticket
	err := s.pool.QueryRow(ctx, `
		SELECT id, company_id, project_id, created_by_id, ticket_number, ticket_type,
		       title, description_raw, description_cleaned, description_final,
		       work_date, status, gc_signed_at, gc_signed_by_id,
		       rejected_at, rejected_by_id, rejection_reason,
		       pdf_generated_at, pdf_path, pdf_hash, locked_at, content_hash,
		       created_at, updated_at
		FROM tickets WHERE id = $1 AND company_id = $2
	`, id, companyID).Scan(
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

// ListByCompany retrieves tickets for a company with optional status filtering.
func (s *TicketService) ListByCompany(ctx context.Context, companyID uuid.UUID, status *domain.TicketStatus, limit int, cursor *domain.TicketCursor) ([]*domain.Ticket, error) {
	var statusFilter *string
	if status != nil {
		statusValue := string(*status)
		statusFilter = &statusValue
	}

	var cursorTime *time.Time
	var cursorID *uuid.UUID
	if cursor != nil {
		cursorTime = &cursor.Time
		cursorID = &cursor.ID
	}

	rows, err := s.pool.Query(ctx, `
		SELECT id, company_id, project_id, created_by_id, ticket_number, ticket_type,
		       title, work_date, status, created_at, updated_at
		FROM tickets
		WHERE company_id = $1
		  AND ($2::text IS NULL OR status = $2)
		  AND ($3::timestamptz IS NULL OR (created_at, id) < ($3::timestamptz, $4::uuid))
		ORDER BY created_at DESC, id DESC
		LIMIT $5
	`, companyID, statusFilter, cursorTime, cursorID, limit)
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

// ListByProject retrieves tickets for a project.
func (s *TicketService) ListByProject(ctx context.Context, companyID, projectID uuid.UUID, limit int, cursor *domain.TicketCursor) ([]*domain.Ticket, error) {
	var cursorTime *time.Time
	var cursorID *uuid.UUID
	if cursor != nil {
		cursorTime = &cursor.Time
		cursorID = &cursor.ID
	}

	rows, err := s.pool.Query(ctx, `
		SELECT id, company_id, project_id, created_by_id, ticket_number, ticket_type,
		       title, work_date, status, created_at, updated_at
		FROM tickets
		WHERE project_id = $1 AND company_id = $2
		  AND ($3::timestamptz IS NULL OR (created_at, id) < ($3::timestamptz, $4::uuid))
		ORDER BY created_at DESC, id DESC
		LIMIT $5
	`, projectID, companyID, cursorTime, cursorID, limit)
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
func (s *TicketService) Update(ctx context.Context, companyID, id, actorID uuid.UUID, input *domain.UpdateTicketInput) (*domain.Ticket, error) {
	// First check if ticket is editable
	ticket, err := s.GetByID(ctx, companyID, id)
	if err != nil {
		return nil, err
	}

	if !ticket.CanEdit() {
		return nil, domain.ErrTicketLocked
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("begin ticket update: %w", err)
	}
	defer tx.Rollback(ctx)

	var updated domain.Ticket
	err = tx.QueryRow(ctx, `
		UPDATE tickets
		SET title = COALESCE($3, title),
		    description_final = COALESCE($4, description_final),
		    work_date = COALESCE($5, work_date),
		    updated_at = now()
		WHERE id = $1 AND company_id = $2 AND status = 'draft'
		RETURNING id, company_id, project_id, created_by_id, ticket_number, ticket_type,
		          title, description_raw, description_cleaned, description_final,
		          work_date, status, created_at, updated_at
	`, id, companyID, input.Title, input.DescriptionFinal, input.WorkDate).Scan(
		&updated.ID, &updated.CompanyID, &updated.ProjectID, &updated.CreatedByID,
		&updated.TicketNumber, &updated.TicketType, &updated.Title,
		&updated.DescriptionRaw, &updated.DescriptionCleaned, &updated.DescriptionFinal,
		&updated.WorkDate, &updated.Status, &updated.CreatedAt, &updated.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("update ticket: %w", err)
	}

	changes := map[string]domain.ChangeValue{}
	if input.Title != nil {
		changes["title"] = domain.ChangeValue{Old: ticket.Title, New: updated.Title}
	}
	if input.DescriptionFinal != nil {
		changes["description_final"] = domain.ChangeValue{Old: ticket.DescriptionFinal, New: updated.DescriptionFinal}
	}
	if input.WorkDate != nil {
		changes["work_date"] = domain.ChangeValue{Old: ticket.WorkDate, New: updated.WorkDate}
	}

	if err := s.insertAuditLog(ctx, tx, domain.CreateAuditLogInput{
		EntityType: "ticket",
		EntityID:   updated.ID,
		Action:     domain.AuditActionUpdated,
		ActorID:    &actorID,
		Changes:    changes,
	}); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit ticket update: %w", err)
	}

	return &updated, nil
}

// SubmitForReview changes ticket status to pending_review.
func (s *TicketService) SubmitForReview(ctx context.Context, companyID, id, actorID uuid.UUID) (*domain.Ticket, error) {
	ticket, err := s.GetByID(ctx, companyID, id)
	if err != nil {
		return nil, err
	}

	if !ticket.Status.CanTransitionTo(domain.TicketStatusPendingReview) {
		return nil, domain.ErrInvalidTransition
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("begin ticket submit: %w", err)
	}
	defer tx.Rollback(ctx)

	var updated domain.Ticket
	err = tx.QueryRow(ctx, `
		UPDATE tickets
		SET status = 'pending_review', updated_at = now()
		WHERE id = $1 AND company_id = $2
		RETURNING id, company_id, project_id, created_by_id, ticket_number, ticket_type,
		          title, work_date, status, created_at, updated_at
	`, id, companyID).Scan(
		&updated.ID, &updated.CompanyID, &updated.ProjectID, &updated.CreatedByID,
		&updated.TicketNumber, &updated.TicketType, &updated.Title,
		&updated.WorkDate, &updated.Status, &updated.CreatedAt, &updated.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("submit ticket: %w", err)
	}

	changes := map[string]domain.ChangeValue{
		"status": {Old: ticket.Status, New: updated.Status},
	}
	if err := s.insertAuditLog(ctx, tx, domain.CreateAuditLogInput{
		EntityType: "ticket",
		EntityID:   updated.ID,
		Action:     domain.AuditActionSubmitted,
		ActorID:    &actorID,
		Changes:    changes,
	}); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit ticket submit: %w", err)
	}

	return &updated, nil
}

// Approve approves a ticket.
func (s *TicketService) Approve(ctx context.Context, companyID, id, signerID uuid.UUID) (*domain.Ticket, error) {
	ticket, err := s.GetByID(ctx, companyID, id)
	if err != nil {
		return nil, err
	}

	if !ticket.Status.CanTransitionTo(domain.TicketStatusApproved) {
		return nil, domain.ErrInvalidTransition
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("begin ticket approve: %w", err)
	}
	defer tx.Rollback(ctx)

	var updated domain.Ticket
	err = tx.QueryRow(ctx, `
		UPDATE tickets
		SET status = 'approved',
		    gc_signed_at = now(),
		    gc_signed_by_id = $2,
		    updated_at = now()
		WHERE id = $1 AND company_id = $3
		RETURNING id, company_id, project_id, created_by_id, ticket_number, ticket_type,
		          title, work_date, status, gc_signed_at, gc_signed_by_id,
		          created_at, updated_at
	`, id, signerID, companyID).Scan(
		&updated.ID, &updated.CompanyID, &updated.ProjectID, &updated.CreatedByID,
		&updated.TicketNumber, &updated.TicketType, &updated.Title,
		&updated.WorkDate, &updated.Status, &updated.GCSignedAt, &updated.GCSignedByID,
		&updated.CreatedAt, &updated.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("approve ticket: %w", err)
	}

	changes := map[string]domain.ChangeValue{
		"status":          {Old: ticket.Status, New: updated.Status},
		"gc_signed_at":    {Old: ticket.GCSignedAt, New: updated.GCSignedAt},
		"gc_signed_by_id": {Old: ticket.GCSignedByID, New: updated.GCSignedByID},
	}
	if err := s.insertAuditLog(ctx, tx, domain.CreateAuditLogInput{
		EntityType: "ticket",
		EntityID:   updated.ID,
		Action:     domain.AuditActionApproved,
		ActorID:    &signerID,
		Changes:    changes,
	}); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit ticket approve: %w", err)
	}

	return &updated, nil
}

// Reject rejects a ticket.
func (s *TicketService) Reject(ctx context.Context, companyID, id, rejectorID uuid.UUID, reason string) (*domain.Ticket, error) {
	ticket, err := s.GetByID(ctx, companyID, id)
	if err != nil {
		return nil, err
	}

	if !ticket.Status.CanTransitionTo(domain.TicketStatusRejected) {
		return nil, domain.ErrInvalidTransition
	}

	if reason == "" {
		return nil, &domain.ValidationError{Field: "reason", Message: "rejection reason is required"}
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("begin ticket reject: %w", err)
	}
	defer tx.Rollback(ctx)

	var updated domain.Ticket
	err = tx.QueryRow(ctx, `
		UPDATE tickets
		SET status = 'rejected',
		    rejected_at = now(),
		    rejected_by_id = $2,
		    rejection_reason = $3,
		    updated_at = now()
		WHERE id = $1 AND company_id = $4
		RETURNING id, company_id, project_id, created_by_id, ticket_number, ticket_type,
		          title, work_date, status, rejected_at, rejected_by_id, rejection_reason,
		          created_at, updated_at
	`, id, rejectorID, reason, companyID).Scan(
		&updated.ID, &updated.CompanyID, &updated.ProjectID, &updated.CreatedByID,
		&updated.TicketNumber, &updated.TicketType, &updated.Title,
		&updated.WorkDate, &updated.Status, &updated.RejectedAt, &updated.RejectedByID,
		&updated.RejectionReason, &updated.CreatedAt, &updated.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("reject ticket: %w", err)
	}

	changes := map[string]domain.ChangeValue{
		"status":           {Old: ticket.Status, New: updated.Status},
		"rejected_at":      {Old: ticket.RejectedAt, New: updated.RejectedAt},
		"rejected_by_id":   {Old: ticket.RejectedByID, New: updated.RejectedByID},
		"rejection_reason": {Old: ticket.RejectionReason, New: updated.RejectionReason},
	}
	if err := s.insertAuditLog(ctx, tx, domain.CreateAuditLogInput{
		EntityType: "ticket",
		EntityID:   updated.ID,
		Action:     domain.AuditActionRejected,
		ActorID:    &rejectorID,
		Changes:    changes,
	}); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit ticket reject: %w", err)
	}

	return &updated, nil
}

// Lock locks a ticket with PDF information.
func (s *TicketService) Lock(ctx context.Context, companyID, id uuid.UUID, pdfPath, pdfHash, contentHash string) (*domain.Ticket, error) {
	ticket, err := s.GetByID(ctx, companyID, id)
	if err != nil {
		return nil, err
	}

	if !ticket.Status.CanTransitionTo(domain.TicketStatusLocked) {
		return nil, domain.ErrInvalidTransition
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("begin ticket lock: %w", err)
	}
	defer tx.Rollback(ctx)

	var updated domain.Ticket
	err = tx.QueryRow(ctx, `
		UPDATE tickets
		SET status = 'locked',
		    locked_at = now(),
		    pdf_generated_at = now(),
		    pdf_path = $2,
		    pdf_hash = $3,
		    content_hash = $4,
		    updated_at = now()
		WHERE id = $1 AND company_id = $5
		RETURNING id, company_id, project_id, created_by_id, ticket_number, ticket_type,
		          title, work_date, status, locked_at, pdf_path, pdf_hash, content_hash,
		          created_at, updated_at
	`, id, pdfPath, pdfHash, contentHash, companyID).Scan(
		&updated.ID, &updated.CompanyID, &updated.ProjectID, &updated.CreatedByID,
		&updated.TicketNumber, &updated.TicketType, &updated.Title,
		&updated.WorkDate, &updated.Status, &updated.LockedAt, &updated.PDFPath,
		&updated.PDFHash, &updated.ContentHash, &updated.CreatedAt, &updated.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("lock ticket: %w", err)
	}

	changes := map[string]domain.ChangeValue{
		"status":       {Old: ticket.Status, New: updated.Status},
		"locked_at":    {Old: ticket.LockedAt, New: updated.LockedAt},
		"pdf_path":     {Old: ticket.PDFPath, New: updated.PDFPath},
		"pdf_hash":     {Old: ticket.PDFHash, New: updated.PDFHash},
		"content_hash": {Old: ticket.ContentHash, New: updated.ContentHash},
	}
	if err := s.insertAuditLog(ctx, tx, domain.CreateAuditLogInput{
		EntityType: "ticket",
		EntityID:   updated.ID,
		Action:     domain.AuditActionLocked,
		ActorID:    nil,
		Changes:    changes,
	}); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit ticket lock: %w", err)
	}

	return &updated, nil
}

func (s *TicketService) insertAuditLog(ctx context.Context, tx pgx.Tx, input domain.CreateAuditLogInput) error {
	changesJSON, err := input.ToJSON()
	if err != nil {
		return fmt.Errorf("marshal audit changes: %w", err)
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO audit_logs (entity_type, entity_id, action, actor_id, actor_ip, changes)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, input.EntityType, input.EntityID, input.Action, input.ActorID, input.ActorIP, changesJSON)
	if err != nil {
		return fmt.Errorf("insert audit log: %w", err)
	}

	return nil
}
