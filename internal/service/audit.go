package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ajejoseph22/siteproof/internal/domain"
	"github.com/ajejoseph22/siteproof/internal/repository/db"
)

// AuditService handles audit log operations.
type AuditService struct {
	pool    *pgxpool.Pool
	queries *db.Queries
}

// NewAuditService creates a new audit service.
func NewAuditService(pool *pgxpool.Pool, queries *db.Queries) *AuditService {
	return &AuditService{
		pool:    pool,
		queries: queries,
	}
}

// ChainVerificationResult contains the result of a hash chain verification.
type ChainVerificationResult struct {
	Valid         bool      `json:"valid"`
	EntityType    string    `json:"entity_type"`
	EntityID      uuid.UUID `json:"entity_id"`
	TotalEntries  int64     `json:"total_entries"`
	BrokenEntries int64     `json:"broken_entries"`
	VerifiedAt    time.Time `json:"verified_at"`
}

// VerifyChainForEntity verifies the audit log hash chain for a specific entity.
// The chain is valid if all entries from the entity's first to last audit log
// have correct prev_hash references.
func (s *AuditService) VerifyChainForEntity(ctx context.Context, companyID uuid.UUID, entityType string, entityID uuid.UUID) (*ChainVerificationResult, error) {
	// First verify the entity belongs to this company (for tickets)
	if entityType == "ticket" {
		var exists bool
		err := s.pool.QueryRow(ctx, `
			SELECT EXISTS(SELECT 1 FROM tickets WHERE id = $1 AND company_id = $2)
		`, entityID, companyID).Scan(&exists)
		if err != nil {
			return nil, fmt.Errorf("check ticket ownership: %w", err)
		}
		if !exists {
			return nil, domain.ErrNotFound
		}
	}

	// Count total entries for this entity
	totalCount, err := s.queries.CountAuditLogsByEntity(ctx, db.CountAuditLogsByEntityParams{
		EntityType: entityType,
		EntityID:   entityID,
	})
	if err != nil {
		return nil, fmt.Errorf("count audit logs: %w", err)
	}

	if totalCount == 0 {
		return &ChainVerificationResult{
			Valid:         true,
			EntityType:    entityType,
			EntityID:      entityID,
			TotalEntries:  0,
			BrokenEntries: 0,
			VerifiedAt:    time.Now().UTC(),
		}, nil
	}

	// Verify the chain for this entity's audit log range
	brokenCount, err := s.queries.VerifyAuditChainForEntity(ctx, db.VerifyAuditChainForEntityParams{
		EntityType: entityType,
		EntityID:   entityID,
	})
	if err != nil {
		return nil, fmt.Errorf("verify audit chain: %w", err)
	}

	return &ChainVerificationResult{
		Valid:         brokenCount == 0,
		EntityType:    entityType,
		EntityID:      entityID,
		TotalEntries:  totalCount,
		BrokenEntries: brokenCount,
		VerifiedAt:    time.Now().UTC(),
	}, nil
}

// VerifyGlobalChain verifies the entire audit log hash chain.
func (s *AuditService) VerifyGlobalChain(ctx context.Context) (*ChainVerificationResult, error) {
	brokenCount, err := s.queries.VerifyAuditChain(ctx)
	if err != nil {
		return nil, fmt.Errorf("verify global audit chain: %w", err)
	}

	// Get total count
	var totalCount int64
	err = s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM audit_logs`).Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("count all audit logs: %w", err)
	}

	return &ChainVerificationResult{
		Valid:         brokenCount == 0,
		EntityType:    "*",
		EntityID:      uuid.Nil,
		TotalEntries:  totalCount,
		BrokenEntries: brokenCount,
		VerifiedAt:    time.Now().UTC(),
	}, nil
}

// GetByEntity retrieves audit logs for an entity with company authorization.
func (s *AuditService) GetByEntity(ctx context.Context, companyID uuid.UUID, entityType string, entityID uuid.UUID, limit, offset int32) ([]*domain.AuditLog, error) {
	// Verify entity belongs to company (for tickets)
	if entityType == "ticket" {
		var exists bool
		err := s.pool.QueryRow(ctx, `
			SELECT EXISTS(SELECT 1 FROM tickets WHERE id = $1 AND company_id = $2)
		`, entityID, companyID).Scan(&exists)
		if err != nil {
			return nil, fmt.Errorf("check ticket ownership: %w", err)
		}
		if !exists {
			return nil, domain.ErrNotFound
		}
	}

	dbLogs, err := s.queries.GetAuditLogsByEntity(ctx, db.GetAuditLogsByEntityParams{
		EntityType: entityType,
		EntityID:   entityID,
		Limit:      limit,
		Offset:     offset,
	})
	if err != nil {
		return nil, fmt.Errorf("get audit logs: %w", err)
	}

	return convertAuditLogs(dbLogs), nil
}

// GetByEntityWithHashes retrieves audit logs for an entity including hash info (chronological order).
func (s *AuditService) GetByEntityWithHashes(ctx context.Context, companyID uuid.UUID, entityType string, entityID uuid.UUID) ([]*domain.AuditLog, error) {
	// Verify entity belongs to company (for tickets)
	if entityType == "ticket" {
		var exists bool
		err := s.pool.QueryRow(ctx, `
			SELECT EXISTS(SELECT 1 FROM tickets WHERE id = $1 AND company_id = $2)
		`, entityID, companyID).Scan(&exists)
		if err != nil {
			return nil, fmt.Errorf("check ticket ownership: %w", err)
		}
		if !exists {
			return nil, domain.ErrNotFound
		}
	}

	dbLogs, err := s.queries.GetAuditLogsByEntityWithHashes(ctx, db.GetAuditLogsByEntityWithHashesParams{
		EntityType: entityType,
		EntityID:   entityID,
	})
	if err != nil {
		return nil, fmt.Errorf("get audit logs with hashes: %w", err)
	}

	return convertAuditLogs(dbLogs), nil
}

// convertAuditLogs converts database audit logs to domain audit logs.
func convertAuditLogs(dbLogs []*db.AuditLog) []*domain.AuditLog {
	logs := make([]*domain.AuditLog, len(dbLogs))
	for i, dbLog := range dbLogs {
		logs[i] = convertAuditLog(dbLog)
	}
	return logs
}

// convertAuditLog converts a database audit log to a domain audit log.
func convertAuditLog(dbLog *db.AuditLog) *domain.AuditLog {
	log := &domain.AuditLog{
		ID:         dbLog.ID,
		EntityType: dbLog.EntityType,
		EntityID:   dbLog.EntityID,
		Action:     domain.AuditAction(dbLog.Action),
		OccurredAt: dbLog.OccurredAt,
		EntryHash:  dbLog.EntryHash,
	}

	// Handle nullable actor_id
	if dbLog.ActorID.Valid {
		actorID := dbLog.ActorID.Bytes
		log.ActorID = (*uuid.UUID)(&actorID)
	}

	// Handle nullable actor_ip
	if dbLog.ActorIp != nil {
		ip := net.IP(dbLog.ActorIp)
		log.ActorIP = &ip
	}

	// Handle nullable prev_hash
	if dbLog.PrevHash.Valid {
		log.PrevHash = &dbLog.PrevHash.String
	}

	// Handle changes JSON
	if len(dbLog.Changes) > 0 {
		log.Changes = json.RawMessage(dbLog.Changes)
	}

	return log
}

// Create creates a new audit log entry.
func (s *AuditService) Create(ctx context.Context, input domain.CreateAuditLogInput) (*domain.AuditLog, error) {
	changesJSON, err := input.ToJSON()
	if err != nil {
		return nil, fmt.Errorf("marshal audit changes: %w", err)
	}

	params := db.CreateAuditLogParams{
		EntityType: input.EntityType,
		EntityID:   input.EntityID,
		Action:     string(input.Action),
		Changes:    changesJSON,
	}

	if input.ActorID != nil {
		params.ActorID = pgtype.UUID{Bytes: *input.ActorID, Valid: true}
	}

	if input.ActorIP != nil {
		params.ActorIp = *input.ActorIP
	}

	dbLog, err := s.queries.CreateAuditLog(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("create audit log: %w", err)
	}

	return convertAuditLog(dbLog), nil
}

// CreateWithTx creates a new audit log entry within a transaction.
func (s *AuditService) CreateWithTx(ctx context.Context, tx pgx.Tx, input domain.CreateAuditLogInput) (*domain.AuditLog, error) {
	changesJSON, err := input.ToJSON()
	if err != nil {
		return nil, fmt.Errorf("marshal audit changes: %w", err)
	}

	var dbLog db.AuditLog
	err = tx.QueryRow(ctx, `
		INSERT INTO audit_logs (entity_type, entity_id, action, actor_id, actor_ip, changes)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, entity_type, entity_id, action, actor_id, actor_ip, occurred_at, changes, prev_hash, entry_hash
	`, input.EntityType, input.EntityID, string(input.Action), input.ActorID, input.ActorIP, changesJSON).Scan(
		&dbLog.ID,
		&dbLog.EntityType,
		&dbLog.EntityID,
		&dbLog.Action,
		&dbLog.ActorID,
		&dbLog.ActorIp,
		&dbLog.OccurredAt,
		&dbLog.Changes,
		&dbLog.PrevHash,
		&dbLog.EntryHash,
	)
	if err != nil {
		return nil, fmt.Errorf("insert audit log: %w", err)
	}

	return convertAuditLog(&dbLog), nil
}
