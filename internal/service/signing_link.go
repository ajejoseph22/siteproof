package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ajejoseph22/siteproof/internal/domain"
	"github.com/ajejoseph22/siteproof/internal/repository/db"
	"github.com/ajejoseph22/siteproof/pkg/crypto"
	"github.com/ajejoseph22/siteproof/pkg/slogctx"
)

// SigningLinkService handles signing link operations.
type SigningLinkService struct {
	pool    *pgxpool.Pool
	queries *db.Queries
	baseURL string
}

// NewSigningLinkService creates a new signing link service.
func NewSigningLinkService(pool *pgxpool.Pool, queries *db.Queries, baseURL string) *SigningLinkService {
	return &SigningLinkService{
		pool:    pool,
		queries: queries,
		baseURL: baseURL,
	}
}

// SigningLinkResult represents the result of creating a signing link.
type SigningLinkResult struct {
	SigningLink *domain.SigningLink `json:"signing_link"`
	SigningURL  string              `json:"signing_url"`
	SentVia     []string            `json:"sent_via"`
}

// RequestSignature creates a signing link and sends it to the signer.
func (s *SigningLinkService) RequestSignature(
	ctx context.Context,
	companyID uuid.UUID,
	actorID uuid.UUID,
	input *domain.CreateSigningLinkInput,
) (*SigningLinkResult, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	// Verify ticket exists and belongs to company
	var ticketCompanyID uuid.UUID
	var ticketStatus string
	err := s.pool.QueryRow(ctx, `
		SELECT company_id, status FROM tickets WHERE id = $1
	`, input.TicketID).Scan(&ticketCompanyID, &ticketStatus)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get ticket: %w", err)
	}

	if ticketCompanyID != companyID {
		return nil, domain.ErrNotFound
	}

	// Only allow signature requests for pending_review tickets
	if ticketStatus != string(domain.TicketStatusPendingReview) {
		return nil, domain.NewDomainError(domain.ErrInvalidTransition, "INVALID_STATUS", "ticket must be in pending_review status to request signature")
	}

	// Generate secure token
	token, err := crypto.GenerateToken()
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

	// Generate verification code for same-device fallback
	verificationCode, err := crypto.GenerateVerificationCode()
	if err != nil {
		return nil, fmt.Errorf("generate verification code: %w", err)
	}

	// Determine expiration (24 hours from now)
	expiresAt := time.Now().Add(24 * time.Hour)

	// Determine delivery methods
	sentVia := []string{"email"}
	if input.SignerPhone != nil && *input.SignerPhone != "" {
		sentVia = append(sentVia, "sms")
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	txQueries := s.queries.WithTx(tx)

	// Revoke any existing active signing links for this ticket
	if err := txQueries.RevokeSigningLinksByTicket(ctx, input.TicketID); err != nil {
		return nil, fmt.Errorf("revoke existing links: %w", err)
	}

	// Create the signing link
	signerPhone := pgtype.Text{}
	if input.SignerPhone != nil {
		signerPhone = pgtype.Text{String: *input.SignerPhone, Valid: true}
	}

	dbLink, err := txQueries.CreateSigningLink(ctx, db.CreateSigningLinkParams{
		TicketID:         input.TicketID,
		Token:            token,
		SignerType:       string(input.SignerType),
		SignerEmail:      input.SignerEmail,
		SignerPhone:      signerPhone,
		SentVia:          sentVia,
		ExpiresAt:        expiresAt,
		VerificationCode: pgtype.Text{String: verificationCode, Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("create signing link: %w", err)
	}

	// Record audit log entry
	if err := s.insertAuditLog(ctx, tx, domain.CreateAuditLogInput{
		EntityType: "ticket",
		EntityID:   input.TicketID,
		Action:     domain.AuditActionSignatureRequested,
		ActorID:    &actorID,
		Changes: map[string]domain.ChangeValue{
			"signing_link_id": {Old: nil, New: dbLink.ID},
			"signer_type":     {Old: nil, New: dbLink.SignerType},
			"signer_email":    {Old: nil, New: dbLink.SignerEmail},
			"sent_via":        {Old: nil, New: sentVia},
			"expires_at":      {Old: nil, New: dbLink.ExpiresAt},
		},
	}); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	// Convert to domain model
	signingLink := toDomainSigningLink(dbLink)
	signingURL := fmt.Sprintf("%s/sign/%s", s.baseURL, token)

	slogctx.Info(ctx, "signature_requested",
		slog.String("ticket_id", input.TicketID.String()),
		slog.String("signing_link_id", dbLink.ID.String()),
		slog.String("signer_email", input.SignerEmail),
		slog.String("signer_type", string(input.SignerType)),
		slog.Any("sent_via", sentVia),
		slog.String("company_id", companyID.String()),
		slog.String("actor_id", actorID.String()),
	)

	return &SigningLinkResult{
		SigningLink: signingLink,
		SigningURL:  signingURL,
		SentVia:     sentVia,
	}, nil
}

// GetByToken retrieves a signing link by its token.
func (s *SigningLinkService) GetByToken(ctx context.Context, token string) (*domain.SigningLink, error) {
	dbLink, err := s.queries.GetSigningLinkByToken(ctx, token)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get signing link: %w", err)
	}

	return toDomainSigningLink(dbLink), nil
}

// GetActiveByTicket retrieves the active signing link for a ticket.
func (s *SigningLinkService) GetActiveByTicket(ctx context.Context, companyID, ticketID uuid.UUID) (*domain.SigningLink, error) {
	// Verify ticket belongs to company
	var ticketCompanyID uuid.UUID
	err := s.pool.QueryRow(ctx, `
		SELECT company_id FROM tickets WHERE id = $1
	`, ticketID).Scan(&ticketCompanyID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get ticket: %w", err)
	}

	if ticketCompanyID != companyID {
		return nil, domain.ErrNotFound
	}

	dbLink, err := s.queries.GetActiveSigningLinkByTicket(ctx, ticketID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get active signing link: %w", err)
	}

	return toDomainSigningLink(dbLink), nil
}

// RevokeByTicket revokes all active signing links for a ticket.
func (s *SigningLinkService) RevokeByTicket(ctx context.Context, companyID, ticketID, actorID uuid.UUID) error {
	// Verify ticket belongs to company
	var ticketCompanyID uuid.UUID
	err := s.pool.QueryRow(ctx, `
		SELECT company_id FROM tickets WHERE id = $1
	`, ticketID).Scan(&ticketCompanyID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return domain.ErrNotFound
		}
		return fmt.Errorf("get ticket: %w", err)
	}

	if ticketCompanyID != companyID {
		return domain.ErrNotFound
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	txQueries := s.queries.WithTx(tx)

	if err := txQueries.RevokeSigningLinksByTicket(ctx, ticketID); err != nil {
		return fmt.Errorf("revoke signing links: %w", err)
	}

	if err := s.insertAuditLog(ctx, tx, domain.CreateAuditLogInput{
		EntityType: "ticket",
		EntityID:   ticketID,
		Action:     domain.AuditActionSigningLinkRevoked,
		ActorID:    &actorID,
		Changes:    nil,
	}); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	slogctx.Info(ctx, "signing_links_revoked",
		slog.String("ticket_id", ticketID.String()),
		slog.String("company_id", companyID.String()),
		slog.String("actor_id", actorID.String()),
	)

	return nil
}

// ValidateToken validates a signing link token and returns the link if valid.
func (s *SigningLinkService) ValidateToken(ctx context.Context, token string) (*domain.SigningLink, error) {
	link, err := s.GetByToken(ctx, token)
	if err != nil {
		return nil, err
	}

	if !link.IsValid() {
		if link.IsExpired() {
			return nil, domain.NewDomainError(domain.ErrExpiredToken, "SIGNING_LINK_EXPIRED", "signing link has expired")
		}
		if link.IsUsed() {
			return nil, domain.NewDomainError(domain.ErrInvalidToken, "SIGNING_LINK_USED", "signing link has already been used")
		}
		if link.IsRevoked() {
			return nil, domain.NewDomainError(domain.ErrInvalidToken, "SIGNING_LINK_REVOKED", "signing link has been revoked")
		}
	}

	return link, nil
}

func (s *SigningLinkService) insertAuditLog(ctx context.Context, tx pgx.Tx, input domain.CreateAuditLogInput) error {
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

func toDomainSigningLink(dbLink *db.SigningLink) *domain.SigningLink {
	link := &domain.SigningLink{
		ID:                   dbLink.ID,
		TicketID:             dbLink.TicketID,
		Token:                dbLink.Token,
		SignerType:           domain.SignerType(dbLink.SignerType),
		SignerEmail:          dbLink.SignerEmail,
		SentVia:              dbLink.SentVia,
		SentAt:               dbLink.SentAt,
		ExpiresAt:            dbLink.ExpiresAt,
		VerificationAttempts: int(dbLink.VerificationAttempts),
		CreatedAt:            dbLink.CreatedAt,
	}

	if dbLink.SignerPhone.Valid {
		link.SignerPhone = &dbLink.SignerPhone.String
	}
	if dbLink.UsedAt.Valid {
		link.UsedAt = &dbLink.UsedAt.Time
	}
	if dbLink.RevokedAt.Valid {
		link.RevokedAt = &dbLink.RevokedAt.Time
	}
	if dbLink.VerificationCode.Valid {
		link.VerificationCode = &dbLink.VerificationCode.String
	}

	return link
}
