package handler

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/ajejoseph22/siteproof/internal/domain"
	"github.com/ajejoseph22/siteproof/internal/middleware"
)

// TicketService defines ticket operations needed by the handler.
type TicketService interface {
	Create(ctx context.Context, input *domain.CreateTicketInput) (*domain.Ticket, error)
	GetByID(ctx context.Context, companyID, id uuid.UUID) (*domain.Ticket, error)
	ListByCompany(ctx context.Context, companyID uuid.UUID, status *domain.TicketStatus, limit int, cursor *domain.TicketCursor) ([]*domain.Ticket, error)
	ListByProject(ctx context.Context, companyID, projectID uuid.UUID, limit int, cursor *domain.TicketCursor) ([]*domain.Ticket, error)
	Update(ctx context.Context, companyID, id, actorID uuid.UUID, input *domain.UpdateTicketInput) (*domain.Ticket, error)
	SubmitForReview(ctx context.Context, companyID, id, actorID uuid.UUID) (*domain.Ticket, error)
	Reject(ctx context.Context, companyID, id, rejectorID uuid.UUID, reason string) (*domain.Ticket, error)
}

// TicketHandler handles ticket endpoints.
type TicketHandler struct {
	ticketService TicketService
}

// NewTicketHandler creates a new ticket handler.
func NewTicketHandler(ticketService TicketService) *TicketHandler {
	return &TicketHandler{ticketService: ticketService}
}

// CreateTicketRequest represents a ticket creation request.
type CreateTicketRequest struct {
	ProjectID  string `json:"project_id"`
	TicketType string `json:"ticket_type"`
	Title      string `json:"title"`
	WorkDate   string `json:"work_date"` // YYYY-MM-DD
}

// Create handles POST /api/v1/tickets
func (h *TicketHandler) Create(w http.ResponseWriter, r *http.Request) {
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

	var req CreateTicketRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_JSON", "invalid request body")
		return
	}

	projectID, err := uuid.Parse(req.ProjectID)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid project_id")
		return
	}

	workDate, err := time.Parse("2006-01-02", req.WorkDate)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid work_date format, use YYYY-MM-DD")
		return
	}

	input := &domain.CreateTicketInput{
		CompanyID:   companyID,
		ProjectID:   projectID,
		CreatedByID: userID,
		TicketType:  domain.TicketType(req.TicketType),
		Title:       req.Title,
		WorkDate:    workDate,
	}

	ticket, err := h.ticketService.Create(r.Context(), input)
	if err != nil {
		WriteErrorFrom(w, err)
		return
	}

	WriteJSON(w, http.StatusCreated, ticket)
}

// Get handles GET /api/v1/tickets/{id}
func (h *TicketHandler) Get(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid ticket id")
		return
	}

	companyID, ok := middleware.CompanyIDFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "not authenticated")
		return
	}

	ticket, err := h.ticketService.GetByID(r.Context(), companyID, id)
	if err != nil {
		WriteErrorFrom(w, err)
		return
	}

	// Verify company access
	if ticket.CompanyID != companyID {
		WriteError(w, http.StatusNotFound, "NOT_FOUND", "ticket not found")
		return
	}

	WriteJSON(w, http.StatusOK, ticket)
}

// List handles GET /api/v1/tickets
func (h *TicketHandler) List(w http.ResponseWriter, r *http.Request) {
	companyID, ok := middleware.CompanyIDFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "not authenticated")
		return
	}

	limit := parseLimit(r)

	cursor, err := parseTicketCursor(r.URL.Query().Get("cursor"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid cursor")
		return
	}

	var status *domain.TicketStatus
	if statusRaw := strings.TrimSpace(r.URL.Query().Get("status")); statusRaw != "" {
		statusValue := domain.TicketStatus(statusRaw)
		if !statusValue.IsValid() {
			WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid status")
			return
		}
		status = &statusValue
	}

	tickets, err := h.ticketService.ListByCompany(r.Context(), companyID, status, limit, cursor)
	if err != nil {
		WriteErrorFrom(w, err)
		return
	}

	nextCursor, hasMore := buildTicketCursor(tickets, limit)
	WriteCursorPaginated(w, tickets, limit, nextCursor, hasMore)
}

// ListByProject handles GET /api/v1/projects/{project_id}/tickets
func (h *TicketHandler) ListByProject(w http.ResponseWriter, r *http.Request) {
	projectIDStr := chi.URLParam(r, "project_id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid project id")
		return
	}

	companyID, ok := middleware.CompanyIDFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "not authenticated")
		return
	}

	limit := parseLimit(r)
	cursor, err := parseTicketCursor(r.URL.Query().Get("cursor"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid cursor")
		return
	}

	tickets, err := h.ticketService.ListByProject(r.Context(), companyID, projectID, limit, cursor)
	if err != nil {
		WriteErrorFrom(w, err)
		return
	}

	nextCursor, hasMore := buildTicketCursor(tickets, limit)
	WriteCursorPaginated(w, tickets, limit, nextCursor, hasMore)
}

// UpdateTicketRequest represents a ticket update request.
type UpdateTicketRequest struct {
	Title            *string `json:"title,omitempty"`
	DescriptionFinal *string `json:"description_final,omitempty"`
	WorkDate         *string `json:"work_date,omitempty"`
}

// Update handles PATCH /api/v1/tickets/{id}
func (h *TicketHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid ticket id")
		return
	}

	companyID, ok := middleware.CompanyIDFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "not authenticated")
		return
	}
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "not authenticated")
		return
	}

	var req UpdateTicketRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_JSON", "invalid request body")
		return
	}

	input := &domain.UpdateTicketInput{
		Title:            req.Title,
		DescriptionFinal: req.DescriptionFinal,
	}

	if req.WorkDate != nil {
		workDate, err := time.Parse("2006-01-02", *req.WorkDate)
		if err != nil {
			WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid work_date format")
			return
		}
		input.WorkDate = &workDate
	}

	ticket, err := h.ticketService.Update(r.Context(), companyID, id, userID, input)
	if err != nil {
		WriteErrorFrom(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, ticket)
}

// Submit handles POST /api/v1/tickets/{id}/submit
func (h *TicketHandler) Submit(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid ticket id")
		return
	}

	companyID, ok := middleware.CompanyIDFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "not authenticated")
		return
	}

	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "not authenticated")
		return
	}

	ticket, err := h.ticketService.SubmitForReview(r.Context(), companyID, id, userID)
	if err != nil {
		WriteErrorFrom(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, ticket)
}

// RejectRequest represents a ticket rejection request.
type RejectRequest struct {
	Reason string `json:"reason"`
}

// Reject handles POST /api/v1/tickets/{id}/reject
func (h *TicketHandler) Reject(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
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

	var req RejectRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_JSON", "invalid request body")
		return
	}

	ticket, err := h.ticketService.Reject(r.Context(), companyID, id, userID, req.Reason)
	if err != nil {
		WriteErrorFrom(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, ticket)
}

func parseLimit(r *http.Request) int {
	limit := 20

	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	return limit
}

func parseTicketCursor(raw string) (*domain.TicketCursor, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}

	parts := strings.Split(raw, "|")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid cursor format")
	}

	cursorTime, err := time.Parse(time.RFC3339Nano, strings.TrimSpace(parts[0]))
	if err != nil {
		return nil, err
	}

	cursorID, err := uuid.Parse(strings.TrimSpace(parts[1]))
	if err != nil {
		return nil, err
	}

	return &domain.TicketCursor{Time: cursorTime, ID: cursorID}, nil
}

func buildTicketCursor(tickets []*domain.Ticket, limit int) (string, bool) {
	if len(tickets) == 0 {
		return "", false
	}

	hasMore := len(tickets) == limit
	if !hasMore {
		return "", false
	}

	last := tickets[len(tickets)-1]
	return fmt.Sprintf("%s|%s", last.CreatedAt.UTC().Format(time.RFC3339Nano), last.ID.String()), true
}
