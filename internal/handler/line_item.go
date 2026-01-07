package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/ajejoseph22/siteproof/internal/domain"
	"github.com/ajejoseph22/siteproof/internal/middleware"
)

// CreateLineItemRequest represents a line item creation request.
type CreateLineItemRequest struct {
	ItemType    string   `json:"item_type"`
	Description string   `json:"description"`
	Quantity    float64  `json:"quantity"`
	Unit        string   `json:"unit"`
	UnitRate    *float64 `json:"unit_rate,omitempty"`
	SortOrder   *int     `json:"sort_order,omitempty"`
}

// UpdateLineItemRequest represents a line item update request.
type UpdateLineItemRequest struct {
	ItemType    *string  `json:"item_type,omitempty"`
	Description *string  `json:"description,omitempty"`
	Quantity    *float64 `json:"quantity,omitempty"`
	Unit        *string  `json:"unit,omitempty"`
	UnitRate    *float64 `json:"unit_rate,omitempty"`
	SortOrder   *int     `json:"sort_order,omitempty"`
}

// ListLineItems handles GET /api/v1/tickets/{id}/line-items
func (h *TicketHandler) ListLineItems(w http.ResponseWriter, r *http.Request) {
	ticketID, ok := parseTicketIDParam(w, r)
	if !ok {
		return
	}

	companyID, ok := middleware.CompanyIDFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "not authenticated")
		return
	}

	items, err := h.ticketService.ListLineItems(r.Context(), companyID, ticketID)
	if err != nil {
		WriteErrorFrom(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, items)
}

// CreateLineItem handles POST /api/v1/tickets/{id}/line-items
func (h *TicketHandler) CreateLineItem(w http.ResponseWriter, r *http.Request) {
	ticketID, ok := parseTicketIDParam(w, r)
	if !ok {
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

	var req CreateLineItemRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_JSON", "invalid request body")
		return
	}

	sortOrder := 0
	if req.SortOrder != nil {
		sortOrder = *req.SortOrder
	}

	input := &domain.CreateLineItemInput{
		TicketID:    ticketID,
		ItemType:    domain.LineItemType(req.ItemType),
		Description: req.Description,
		Quantity:    req.Quantity,
		Unit:        req.Unit,
		UnitRate:    req.UnitRate,
		SortOrder:   sortOrder,
	}

	item, err := h.ticketService.CreateLineItem(r.Context(), companyID, ticketID, userID, input)
	if err != nil {
		WriteErrorFrom(w, err)
		return
	}

	WriteJSON(w, http.StatusCreated, item)
}

// UpdateLineItem handles PATCH /api/v1/tickets/{id}/line-items/{line_item_id}
func (h *TicketHandler) UpdateLineItem(w http.ResponseWriter, r *http.Request) {
	ticketID, ok := parseTicketIDParam(w, r)
	if !ok {
		return
	}

	lineItemID, ok := parseLineItemIDParam(w, r)
	if !ok {
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

	var req UpdateLineItemRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_JSON", "invalid request body")
		return
	}

	input := &domain.UpdateLineItemInput{
		Description: req.Description,
		Quantity:    req.Quantity,
		Unit:        req.Unit,
		UnitRate:    req.UnitRate,
		SortOrder:   req.SortOrder,
	}

	if req.ItemType != nil {
		itemType := domain.LineItemType(*req.ItemType)
		input.ItemType = &itemType
	}

	item, err := h.ticketService.UpdateLineItem(r.Context(), companyID, ticketID, lineItemID, userID, input)
	if err != nil {
		WriteErrorFrom(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, item)
}

// DeleteLineItem handles DELETE /api/v1/tickets/{id}/line-items/{line_item_id}
func (h *TicketHandler) DeleteLineItem(w http.ResponseWriter, r *http.Request) {
	ticketID, ok := parseTicketIDParam(w, r)
	if !ok {
		return
	}

	lineItemID, ok := parseLineItemIDParam(w, r)
	if !ok {
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

	if err := h.ticketService.DeleteLineItem(r.Context(), companyID, ticketID, lineItemID, userID); err != nil {
		WriteErrorFrom(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func parseTicketIDParam(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid ticket id")
		return uuid.UUID{}, false
	}
	return id, true
}

func parseLineItemIDParam(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	idStr := chi.URLParam(r, "line_item_id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid line item id")
		return uuid.UUID{}, false
	}
	return id, true
}
