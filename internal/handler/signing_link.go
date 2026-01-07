package handler

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/ajejoseph22/siteproof/internal/domain"
	"github.com/ajejoseph22/siteproof/internal/middleware"
	"github.com/ajejoseph22/siteproof/internal/service"
)

// SigningLinkService defines signing link operations needed by the handler.
type SigningLinkService interface {
	RequestSignature(ctx context.Context, companyID, actorID uuid.UUID, input *domain.CreateSigningLinkInput) (*service.SigningLinkResult, error)
	GetActiveByTicket(ctx context.Context, companyID, ticketID uuid.UUID) (*domain.SigningLink, error)
	RevokeByTicket(ctx context.Context, companyID, ticketID, actorID uuid.UUID) error
	ValidateToken(ctx context.Context, token string) (*domain.SigningLink, error)
}

// SigningLinkHandler handles signing link endpoints.
type SigningLinkHandler struct {
	signingLinkService SigningLinkService
}

// NewSigningLinkHandler creates a new signing link handler.
func NewSigningLinkHandler(signingLinkService SigningLinkService) *SigningLinkHandler {
	return &SigningLinkHandler{signingLinkService: signingLinkService}
}

// RequestSignatureRequest represents a signature request payload.
type RequestSignatureRequest struct {
	SignerType  string  `json:"signer_type"`
	SignerEmail string  `json:"signer_email"`
	SignerPhone *string `json:"signer_phone,omitempty"`
}

// RequestSignatureResponse represents the response for a signature request.
type RequestSignatureResponse struct {
	ID          uuid.UUID `json:"id"`
	TicketID    uuid.UUID `json:"ticket_id"`
	SignerEmail string    `json:"signer_email"`
	SignerPhone *string   `json:"signer_phone,omitempty"`
	SigningURL  string    `json:"signing_url"`
	SentVia     []string  `json:"sent_via"`
	ExpiresAt   string    `json:"expires_at"`
}

// RequestSignature handles POST /api/v1/tickets/{id}/signature-request
func (h *SigningLinkHandler) RequestSignature(w http.ResponseWriter, r *http.Request) {
	ticketIDStr := chi.URLParam(r, "id")
	ticketID, err := uuid.Parse(ticketIDStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid ticket id")
		return
	}

	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "not authenticated")
		return
	}

	companyID, ok := middleware.CompanyIDFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "not authenticated")
		return
	}

	var req RequestSignatureRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_JSON", "invalid request body")
		return
	}

	input := &domain.CreateSigningLinkInput{
		TicketID:    ticketID,
		SignerType:  domain.SignerType(req.SignerType),
		SignerEmail: req.SignerEmail,
		SignerPhone: req.SignerPhone,
	}

	result, err := h.signingLinkService.RequestSignature(r.Context(), companyID, userID, input)
	if err != nil {
		WriteErrorFrom(w, err)
		return
	}

	response := RequestSignatureResponse{
		ID:          result.SigningLink.ID,
		TicketID:    result.SigningLink.TicketID,
		SignerEmail: result.SigningLink.SignerEmail,
		SignerPhone: result.SigningLink.SignerPhone,
		SigningURL:  result.SigningURL,
		SentVia:     result.SentVia,
		ExpiresAt:   result.SigningLink.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	WriteJSON(w, http.StatusCreated, response)
}

// GetActiveSigningLink handles GET /api/v1/tickets/{id}/signature-request
func (h *SigningLinkHandler) GetActiveSigningLink(w http.ResponseWriter, r *http.Request) {
	ticketIDStr := chi.URLParam(r, "id")
	ticketID, err := uuid.Parse(ticketIDStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid ticket id")
		return
	}

	companyID, ok := middleware.CompanyIDFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "not authenticated")
		return
	}

	link, err := h.signingLinkService.GetActiveByTicket(r.Context(), companyID, ticketID)
	if err != nil {
		WriteErrorFrom(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, link)
}

// RevokeSigningLinks handles DELETE /api/v1/tickets/{id}/signature-request
func (h *SigningLinkHandler) RevokeSigningLinks(w http.ResponseWriter, r *http.Request) {
	ticketIDStr := chi.URLParam(r, "id")
	ticketID, err := uuid.Parse(ticketIDStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid ticket id")
		return
	}

	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "not authenticated")
		return
	}

	companyID, ok := middleware.CompanyIDFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "not authenticated")
		return
	}

	if err := h.signingLinkService.RevokeByTicket(r.Context(), companyID, ticketID, userID); err != nil {
		WriteErrorFrom(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ValidateToken handles GET /api/v1/sign/{token} (public endpoint for GC)
func (h *SigningLinkHandler) ValidateToken(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	if token == "" {
		WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "token is required")
		return
	}

	link, err := h.signingLinkService.ValidateToken(r.Context(), token)
	if err != nil {
		WriteErrorFrom(w, err)
		return
	}

	// Return limited info for the signing page (don't expose internal details)
	type ValidateResponse struct {
		ID         uuid.UUID `json:"id"`
		TicketID   uuid.UUID `json:"ticket_id"`
		SignerType string    `json:"signer_type"`
		ExpiresAt  string    `json:"expires_at"`
	}

	WriteJSON(w, http.StatusOK, ValidateResponse{
		ID:         link.ID,
		TicketID:   link.TicketID,
		SignerType: string(link.SignerType),
		ExpiresAt:  link.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}
