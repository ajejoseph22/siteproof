package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/ajejoseph22/siteproof/internal/domain"
	"github.com/ajejoseph22/siteproof/internal/middleware"
	"github.com/ajejoseph22/siteproof/internal/service"
)

// TicketHandler handles ticket endpoints.
type TicketHandler struct {
	ticketService *service.TicketService
}

// NewTicketHandler creates a new ticket handler.
func NewTicketHandler(ticketService *service.TicketService) *TicketHandler {
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

	companyID, _ := middleware.CompanyIDFromContext(r.Context())

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
		if _, ok := err.(domain.ValidationErrors); ok {
			WriteErrorWithDetails(w, http.StatusBadRequest, "VALIDATION_ERROR", "validation failed", err)
			return
		}
		WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create ticket")
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

	companyID, _ := middleware.CompanyIDFromContext(r.Context())

	ticket, err := h.ticketService.GetByID(r.Context(), id)
	if err != nil {
		if err == domain.ErrNotFound {
			WriteError(w, http.StatusNotFound, "NOT_FOUND", "ticket not found")
			return
		}
		WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get ticket")
		return
	}

	// Verify company access
	if ticket.CompanyID != companyID {
		WriteError(w, http.StatusNotFound, "NOT_FOUND", "ticket not found")
		return
	}

	WriteJSON(w, http.StatusOK, ticket)
}

// ListByProject handles GET /api/v1/projects/{project_id}/tickets
func (h *TicketHandler) ListByProject(w http.ResponseWriter, r *http.Request) {
	projectIDStr := chi.URLParam(r, "project_id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid project id")
		return
	}

	limit, offset := parsePagination(r)

	tickets, err := h.ticketService.ListByProject(r.Context(), projectID, limit, offset)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list tickets")
		return
	}

	WriteJSON(w, http.StatusOK, tickets)
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

	ticket, err := h.ticketService.Update(r.Context(), id, input)
	if err != nil {
		switch err {
		case domain.ErrNotFound:
			WriteError(w, http.StatusNotFound, "NOT_FOUND", "ticket not found")
		case domain.ErrTicketLocked:
			WriteError(w, http.StatusUnprocessableEntity, "TICKET_LOCKED", "ticket is locked and cannot be modified")
		default:
			WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to update ticket")
		}
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

	ticket, err := h.ticketService.SubmitForReview(r.Context(), id)
	if err != nil {
		switch err {
		case domain.ErrNotFound:
			WriteError(w, http.StatusNotFound, "NOT_FOUND", "ticket not found")
		case domain.ErrInvalidTransition:
			WriteError(w, http.StatusUnprocessableEntity, "INVALID_TRANSITION", "cannot submit ticket in current status")
		default:
			WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to submit ticket")
		}
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

	userID, _ := middleware.UserIDFromContext(r.Context())

	var req RejectRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_JSON", "invalid request body")
		return
	}

	ticket, err := h.ticketService.Reject(r.Context(), id, userID, req.Reason)
	if err != nil {
		switch err {
		case domain.ErrNotFound:
			WriteError(w, http.StatusNotFound, "NOT_FOUND", "ticket not found")
		case domain.ErrInvalidTransition:
			WriteError(w, http.StatusUnprocessableEntity, "INVALID_TRANSITION", "cannot reject ticket in current status")
		default:
			if ve, ok := err.(*domain.ValidationError); ok {
				WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", ve.Message)
			} else {
				WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to reject ticket")
			}
		}
		return
	}

	WriteJSON(w, http.StatusOK, ticket)
}

func parsePagination(r *http.Request) (limit, offset int) {
	limit = 20
	offset = 0

	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	return limit, offset
}
