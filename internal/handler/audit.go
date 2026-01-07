package handler

import (
	"context"
	"net/http"
	"strconv"

	"github.com/google/uuid"

	"github.com/ajejoseph22/siteproof/internal/domain"
	"github.com/ajejoseph22/siteproof/internal/middleware"
	"github.com/ajejoseph22/siteproof/internal/service"
)

// AuditServiceInterface defines audit operations needed by the handler.
type AuditServiceInterface interface {
	GetByEntity(ctx context.Context, companyID uuid.UUID, entityType string, entityID uuid.UUID, limit, offset int32) ([]*domain.AuditLog, error)
	GetByEntityWithHashes(ctx context.Context, companyID uuid.UUID, entityType string, entityID uuid.UUID) ([]*domain.AuditLog, error)
	VerifyChainForEntity(ctx context.Context, companyID uuid.UUID, entityType string, entityID uuid.UUID) (*service.ChainVerificationResult, error)
}

// AuditHandler handles audit log endpoints.
type AuditHandler struct {
	auditService AuditServiceInterface
}

// NewAuditHandler creates a new audit handler.
func NewAuditHandler(auditService AuditServiceInterface) *AuditHandler {
	return &AuditHandler{
		auditService: auditService,
	}
}

// ListByTicket handles GET /api/v1/tickets/{id}/audit-logs
func (h *AuditHandler) ListByTicket(w http.ResponseWriter, r *http.Request) {
	ticketID, ok := parseTicketIDParam(w, r)
	if !ok {
		return
	}

	companyID, ok := middleware.CompanyIDFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "not authenticated")
		return
	}

	// Parse pagination params
	limit, offset := parsePaginationParams(r)

	logs, err := h.auditService.GetByEntity(r.Context(), companyID, "ticket", ticketID, limit, offset)
	if err != nil {
		WriteErrorFrom(w, err)
		return
	}

	response := make([]AuditLogResponse, 0, len(logs))
	for _, log := range logs {
		response = append(response, newAuditLogResponse(log))
	}

	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"data":   response,
		"limit":  limit,
		"offset": offset,
	})
}

// VerifyTicketChain handles GET /api/v1/tickets/{id}/audit-logs/verify
func (h *AuditHandler) VerifyTicketChain(w http.ResponseWriter, r *http.Request) {
	ticketID, ok := parseTicketIDParam(w, r)
	if !ok {
		return
	}

	companyID, ok := middleware.CompanyIDFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "not authenticated")
		return
	}

	result, err := h.auditService.VerifyChainForEntity(r.Context(), companyID, "ticket", ticketID)
	if err != nil {
		WriteErrorFrom(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, ChainVerificationResponse{
		Valid:         result.Valid,
		EntityType:    result.EntityType,
		EntityID:      result.EntityID,
		TotalEntries:  result.TotalEntries,
		BrokenEntries: result.BrokenEntries,
		VerifiedAt:    result.VerifiedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

// GetTicketAuditLogWithHashes handles GET /api/v1/tickets/{id}/audit-logs/full
// Returns audit logs in chronological order with hash information for verification display.
func (h *AuditHandler) GetTicketAuditLogWithHashes(w http.ResponseWriter, r *http.Request) {
	ticketID, ok := parseTicketIDParam(w, r)
	if !ok {
		return
	}

	companyID, ok := middleware.CompanyIDFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "not authenticated")
		return
	}

	logs, err := h.auditService.GetByEntityWithHashes(r.Context(), companyID, "ticket", ticketID)
	if err != nil {
		WriteErrorFrom(w, err)
		return
	}

	response := make([]AuditLogWithHashResponse, 0, len(logs))
	for _, log := range logs {
		response = append(response, newAuditLogWithHashResponse(log))
	}

	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"data":  response,
		"count": len(response),
	})
}

// AuditLogResponse represents an audit log entry for API responses.
type AuditLogResponse struct {
	ID         int64       `json:"id"`
	EntityType string      `json:"entity_type"`
	EntityID   uuid.UUID   `json:"entity_id"`
	Action     string      `json:"action"`
	ActorID    *uuid.UUID  `json:"actor_id,omitempty"`
	OccurredAt string      `json:"occurred_at"`
	Changes    interface{} `json:"changes,omitempty"`
}

// AuditLogWithHashResponse includes hash information for chain verification display.
type AuditLogWithHashResponse struct {
	AuditLogResponse
	PrevHash  *string `json:"prev_hash,omitempty"`
	EntryHash string  `json:"entry_hash"`
}

// ChainVerificationResponse represents the result of hash chain verification.
type ChainVerificationResponse struct {
	Valid         bool      `json:"valid"`
	EntityType    string    `json:"entity_type"`
	EntityID      uuid.UUID `json:"entity_id"`
	TotalEntries  int64     `json:"total_entries"`
	BrokenEntries int64     `json:"broken_entries"`
	VerifiedAt    string    `json:"verified_at"`
}

func newAuditLogResponse(log *domain.AuditLog) AuditLogResponse {
	resp := AuditLogResponse{
		ID:         log.ID,
		EntityType: log.EntityType,
		EntityID:   log.EntityID,
		Action:     string(log.Action),
		ActorID:    log.ActorID,
		OccurredAt: log.OccurredAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	if len(log.Changes) > 0 {
		resp.Changes = log.Changes
	}

	return resp
}

func newAuditLogWithHashResponse(log *domain.AuditLog) AuditLogWithHashResponse {
	return AuditLogWithHashResponse{
		AuditLogResponse: newAuditLogResponse(log),
		PrevHash:         log.PrevHash,
		EntryHash:        log.EntryHash,
	}
}

func parsePaginationParams(r *http.Request) (limit, offset int32) {
	limit = 50 // default
	offset = 0

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsed, err := strconv.ParseInt(limitStr, 10, 32); err == nil && parsed > 0 && parsed <= 100 {
			limit = int32(parsed)
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if parsed, err := strconv.ParseInt(offsetStr, 10, 32); err == nil && parsed >= 0 {
			offset = int32(parsed)
		}
	}

	return limit, offset
}
