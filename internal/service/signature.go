package service

import (
	"context"
	"fmt"
	"log/slog"
	"net"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ajejoseph22/siteproof/internal/domain"
	"github.com/ajejoseph22/siteproof/internal/repository/db"
	"github.com/ajejoseph22/siteproof/pkg/slogctx"
)

// SignatureService handles signature operations.
type SignatureService struct {
	pool    *pgxpool.Pool
	queries *db.Queries
}

// NewSignatureService creates a new signature service.
func NewSignatureService(pool *pgxpool.Pool, queries *db.Queries) *SignatureService {
	return &SignatureService{
		pool:    pool,
		queries: queries,
	}
}

// CaptureSignatureInput represents input for capturing a signature via signing link.
type CaptureSignatureInput struct {
	SignerName     string
	SignerEmail    *string
	SignatureImage string // relative path where file was stored
	SignatureHash  string // SHA-256 hash of the file
	SigningMethod  domain.SigningMethod
	IPAddress      *net.IP
	UserAgent      *string
	DeviceID       *string
}

// CaptureSignature captures a signature using a signing link token.
// This is the main entry point for GC signatures via the public signing flow.
func (s *SignatureService) CaptureSignature(
	ctx context.Context,
	token string,
	input *CaptureSignatureInput,
) (*domain.Signature, error) {
	// Validate the signing link token
	dbLink, err := s.queries.GetSigningLinkByToken(ctx, token)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.NewDomainError(domain.ErrInvalidToken, "INVALID_TOKEN", "signing link not found")
		}
		return nil, fmt.Errorf("get signing link: %w", err)
	}

	link := toDomainSigningLink(dbLink)
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

	// Build CreateSignatureInput
	createInput := &domain.CreateSignatureInput{
		TicketID:       link.TicketID,
		SigningLinkID:  &link.ID,
		SignerID:       nil, // GC super is not a user in our system
		SignerType:     link.SignerType,
		SignerName:     input.SignerName,
		SignerEmail:    input.SignerEmail,
		SignatureImage: input.SignatureImage,
		SignatureHash:  input.SignatureHash,
		SigningMethod:  input.SigningMethod,
		IPAddress:      input.IPAddress,
		UserAgent:      input.UserAgent,
		DeviceID:       input.DeviceID,
	}

	if err := createInput.Validate(); err != nil {
		return nil, err
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("begin signature capture: %w", err)
	}
	defer tx.Rollback(ctx)

	txQueries := s.queries.WithTx(tx)

	// Verify ticket status - must be pending_review for GC signature
	var ticketStatus string
	var ticketCompanyID uuid.UUID
	err = tx.QueryRow(ctx, `
		SELECT status, company_id FROM tickets WHERE id = $1 FOR UPDATE
	`, link.TicketID).Scan(&ticketStatus, &ticketCompanyID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get ticket: %w", err)
	}

	if ticketStatus != string(domain.TicketStatusPendingReview) {
		return nil, domain.NewDomainError(domain.ErrInvalidTransition, "INVALID_STATUS", "ticket must be in pending_review status to accept signature")
	}

	// Check for duplicate signature type
	_, err = txQueries.GetSignatureByTicketAndType(ctx, db.GetSignatureByTicketAndTypeParams{
		TicketID:   link.TicketID,
		SignerType: string(link.SignerType),
	})
	if err == nil {
		// Signature already exists
		return nil, domain.NewDomainError(domain.ErrConflict, "DUPLICATE_SIGNATURE", fmt.Sprintf("signature of type %s already exists for this ticket", link.SignerType))
	} else if err != pgx.ErrNoRows {
		return nil, fmt.Errorf("check existing signature: %w", err)
	}

	// Create the signature
	signerEmail := pgtype.Text{}
	if createInput.SignerEmail != nil {
		signerEmail = pgtype.Text{String: *createInput.SignerEmail, Valid: true}
	}
	signingLinkID := pgtype.UUID{}
	if createInput.SigningLinkID != nil {
		signingLinkID = pgtype.UUID{Bytes: *createInput.SigningLinkID, Valid: true}
	}
	signerID := pgtype.UUID{}
	if createInput.SignerID != nil {
		signerID = pgtype.UUID{Bytes: *createInput.SignerID, Valid: true}
	}
	var ipAddress net.IP
	if createInput.IPAddress != nil {
		ipAddress = *createInput.IPAddress
	}
	userAgent := pgtype.Text{}
	if createInput.UserAgent != nil {
		userAgent = pgtype.Text{String: *createInput.UserAgent, Valid: true}
	}
	deviceID := pgtype.Text{}
	if createInput.DeviceID != nil {
		deviceID = pgtype.Text{String: *createInput.DeviceID, Valid: true}
	}

	dbSig, err := txQueries.CreateSignature(ctx, db.CreateSignatureParams{
		TicketID:       createInput.TicketID,
		SigningLinkID:  signingLinkID,
		SignerID:       signerID,
		SignerType:     string(createInput.SignerType),
		SignerName:     createInput.SignerName,
		SignerEmail:    signerEmail,
		SignatureImage: createInput.SignatureImage,
		SignatureHash:  createInput.SignatureHash,
		SigningMethod:  string(createInput.SigningMethod),
		IpAddress:      ipAddress,
		UserAgent:      userAgent,
		DeviceID:       deviceID,
	})
	if err != nil {
		return nil, fmt.Errorf("create signature: %w", err)
	}

	// Mark signing link as used
	if _, err := txQueries.MarkSigningLinkUsed(ctx, link.ID); err != nil {
		return nil, fmt.Errorf("mark signing link used: %w", err)
	}

	// Update ticket status to approved when GC signs
	if link.SignerType == domain.SignerTypeGCSuper {
		_, err = tx.Exec(ctx, `
			UPDATE tickets
			SET status = 'approved',
			    gc_signed_at = now(),
			    updated_at = now()
			WHERE id = $1
		`, link.TicketID)
		if err != nil {
			return nil, fmt.Errorf("update ticket status: %w", err)
		}
	}

	// Record audit log
	if err := s.insertAuditLog(ctx, tx, domain.CreateAuditLogInput{
		EntityType: "ticket",
		EntityID:   link.TicketID,
		Action:     domain.AuditActionSigned,
		ActorID:    nil, // External signer, no user ID
		ActorIP:    createInput.IPAddress,
		Changes: map[string]domain.ChangeValue{
			"signature_id":    {Old: nil, New: dbSig.ID},
			"signer_type":     {Old: nil, New: dbSig.SignerType},
			"signer_name":     {Old: nil, New: dbSig.SignerName},
			"signature_hash":  {Old: nil, New: dbSig.SignatureHash},
			"signing_method":  {Old: nil, New: dbSig.SigningMethod},
			"signing_link_id": {Old: nil, New: link.ID},
			"status_change":   {Old: ticketStatus, New: "approved"},
		},
	}); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit signature capture: %w", err)
	}

	signature := toDomainSignature(dbSig)

	slogctx.Info(ctx, "signature_captured",
		slog.String("signature_id", signature.ID.String()),
		slog.String("ticket_id", link.TicketID.String()),
		slog.String("signer_type", string(link.SignerType)),
		slog.String("signer_name", input.SignerName),
		slog.String("signing_method", string(input.SigningMethod)),
		slog.String("company_id", ticketCompanyID.String()),
	)

	return signature, nil
}

// CaptureInternalSignature captures a signature from an authenticated user (foreman).
// This is used when a foreman signs from within the app.
func (s *SignatureService) CaptureInternalSignature(
	ctx context.Context,
	companyID, ticketID, signerID uuid.UUID,
	input *CaptureSignatureInput,
) (*domain.Signature, error) {
	createInput := &domain.CreateSignatureInput{
		TicketID:       ticketID,
		SigningLinkID:  nil,
		SignerID:       &signerID,
		SignerType:     domain.SignerTypeForeman,
		SignerName:     input.SignerName,
		SignerEmail:    input.SignerEmail,
		SignatureImage: input.SignatureImage,
		SignatureHash:  input.SignatureHash,
		SigningMethod:  input.SigningMethod,
		IPAddress:      input.IPAddress,
		UserAgent:      input.UserAgent,
		DeviceID:       input.DeviceID,
	}

	if err := createInput.Validate(); err != nil {
		return nil, err
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("begin internal signature: %w", err)
	}
	defer tx.Rollback(ctx)

	txQueries := s.queries.WithTx(tx)

	// Verify ticket exists and belongs to company, and is in draft status
	var ticketStatus string
	err = tx.QueryRow(ctx, `
		SELECT status FROM tickets WHERE id = $1 AND company_id = $2 FOR UPDATE
	`, ticketID, companyID).Scan(&ticketStatus)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get ticket: %w", err)
	}

	// Foreman can sign tickets in draft status (before submission)
	if ticketStatus != string(domain.TicketStatusDraft) {
		return nil, domain.NewDomainError(domain.ErrInvalidTransition, "INVALID_STATUS", "ticket must be in draft status for foreman signature")
	}

	// Check for duplicate foreman signature
	_, err = txQueries.GetSignatureByTicketAndType(ctx, db.GetSignatureByTicketAndTypeParams{
		TicketID:   ticketID,
		SignerType: string(domain.SignerTypeForeman),
	})
	if err == nil {
		return nil, domain.NewDomainError(domain.ErrConflict, "DUPLICATE_SIGNATURE", "foreman signature already exists for this ticket")
	} else if err != pgx.ErrNoRows {
		return nil, fmt.Errorf("check existing signature: %w", err)
	}

	// Create the signature
	signerEmail := pgtype.Text{}
	if createInput.SignerEmail != nil {
		signerEmail = pgtype.Text{String: *createInput.SignerEmail, Valid: true}
	}
	signerIDPG := pgtype.UUID{Bytes: signerID, Valid: true}
	var ipAddress net.IP
	if createInput.IPAddress != nil {
		ipAddress = *createInput.IPAddress
	}
	userAgent := pgtype.Text{}
	if createInput.UserAgent != nil {
		userAgent = pgtype.Text{String: *createInput.UserAgent, Valid: true}
	}
	deviceID := pgtype.Text{}
	if createInput.DeviceID != nil {
		deviceID = pgtype.Text{String: *createInput.DeviceID, Valid: true}
	}

	dbSig, err := txQueries.CreateSignature(ctx, db.CreateSignatureParams{
		TicketID:       ticketID,
		SigningLinkID:  pgtype.UUID{},
		SignerID:       signerIDPG,
		SignerType:     string(domain.SignerTypeForeman),
		SignerName:     createInput.SignerName,
		SignerEmail:    signerEmail,
		SignatureImage: createInput.SignatureImage,
		SignatureHash:  createInput.SignatureHash,
		SigningMethod:  string(createInput.SigningMethod),
		IpAddress:      ipAddress,
		UserAgent:      userAgent,
		DeviceID:       deviceID,
	})
	if err != nil {
		return nil, fmt.Errorf("create signature: %w", err)
	}

	// Record audit log
	if err := s.insertAuditLog(ctx, tx, domain.CreateAuditLogInput{
		EntityType: "ticket",
		EntityID:   ticketID,
		Action:     domain.AuditActionSigned,
		ActorID:    &signerID,
		ActorIP:    createInput.IPAddress,
		Changes: map[string]domain.ChangeValue{
			"signature_id":   {Old: nil, New: dbSig.ID},
			"signer_type":    {Old: nil, New: dbSig.SignerType},
			"signer_name":    {Old: nil, New: dbSig.SignerName},
			"signature_hash": {Old: nil, New: dbSig.SignatureHash},
			"signing_method": {Old: nil, New: dbSig.SigningMethod},
		},
	}); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit internal signature: %w", err)
	}

	signature := toDomainSignature(dbSig)

	slogctx.Info(ctx, "signature_captured",
		slog.String("signature_id", signature.ID.String()),
		slog.String("ticket_id", ticketID.String()),
		slog.String("signer_type", string(domain.SignerTypeForeman)),
		slog.String("signer_id", signerID.String()),
		slog.String("signer_name", input.SignerName),
		slog.String("signing_method", string(input.SigningMethod)),
		slog.String("company_id", companyID.String()),
	)

	return signature, nil
}

// ListByTicket retrieves all signatures for a ticket.
func (s *SignatureService) ListByTicket(ctx context.Context, companyID, ticketID uuid.UUID) ([]*domain.Signature, error) {
	// Verify ticket belongs to company
	var exists bool
	err := s.pool.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM tickets WHERE id = $1 AND company_id = $2)
	`, ticketID, companyID).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("check ticket exists: %w", err)
	}
	if !exists {
		return nil, domain.ErrNotFound
	}

	dbSigs, err := s.queries.GetSignaturesByTicket(ctx, ticketID)
	if err != nil {
		return nil, fmt.Errorf("list signatures: %w", err)
	}

	signatures := make([]*domain.Signature, 0, len(dbSigs))
	for _, dbSig := range dbSigs {
		signatures = append(signatures, toDomainSignature(dbSig))
	}

	return signatures, nil
}

// GetByID retrieves a signature by ID.
func (s *SignatureService) GetByID(ctx context.Context, companyID, signatureID uuid.UUID) (*domain.Signature, error) {
	dbSig, err := s.queries.GetSignatureByID(ctx, signatureID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get signature: %w", err)
	}

	// Verify company ownership via ticket
	var ticketCompanyID uuid.UUID
	err = s.pool.QueryRow(ctx, `
		SELECT company_id FROM tickets WHERE id = $1
	`, dbSig.TicketID).Scan(&ticketCompanyID)
	if err != nil {
		return nil, fmt.Errorf("get ticket company: %w", err)
	}
	if ticketCompanyID != companyID {
		return nil, domain.ErrNotFound
	}

	return toDomainSignature(dbSig), nil
}

func (s *SignatureService) insertAuditLog(ctx context.Context, tx pgx.Tx, input domain.CreateAuditLogInput) error {
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

func toDomainSignature(dbSig *db.Signature) *domain.Signature {
	sig := &domain.Signature{
		ID:             dbSig.ID,
		TicketID:       dbSig.TicketID,
		SignerType:     domain.SignerType(dbSig.SignerType),
		SignerName:     dbSig.SignerName,
		SignatureImage: dbSig.SignatureImage,
		SignatureHash:  dbSig.SignatureHash,
		SigningMethod:  domain.SigningMethod(dbSig.SigningMethod),
		SignedAt:       dbSig.SignedAt,
	}

	if dbSig.SigningLinkID.Valid {
		id := uuid.UUID(dbSig.SigningLinkID.Bytes)
		sig.SigningLinkID = &id
	}
	if dbSig.SignerID.Valid {
		id := uuid.UUID(dbSig.SignerID.Bytes)
		sig.SignerID = &id
	}
	if dbSig.SignerEmail.Valid {
		sig.SignerEmail = &dbSig.SignerEmail.String
	}
	if dbSig.IpAddress != nil {
		sig.IPAddress = &dbSig.IpAddress
	}
	if dbSig.UserAgent.Valid {
		sig.UserAgent = &dbSig.UserAgent.String
	}
	if dbSig.DeviceID.Valid {
		sig.DeviceID = &dbSig.DeviceID.String
	}

	return sig
}
