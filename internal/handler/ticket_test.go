package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/ajejoseph22/siteproof/internal/domain"
	"github.com/ajejoseph22/siteproof/internal/middleware"
)

type ticketServiceStub struct {
	getByID       func(ctx context.Context, companyID, id uuid.UUID) (*domain.Ticket, error)
	listByCompany func(ctx context.Context, companyID uuid.UUID, status *domain.TicketStatus, limit int, cursor *domain.TicketCursor) ([]*domain.Ticket, error)
	listByProject func(ctx context.Context, companyID, projectID uuid.UUID, limit int, cursor *domain.TicketCursor) ([]*domain.Ticket, error)
}

func (s ticketServiceStub) Create(ctx context.Context, input *domain.CreateTicketInput) (*domain.Ticket, error) {
	return nil, errors.New("not implemented")
}

func (s ticketServiceStub) GetByID(ctx context.Context, companyID, id uuid.UUID) (*domain.Ticket, error) {
	if s.getByID == nil {
		return nil, errors.New("not implemented")
	}
	return s.getByID(ctx, companyID, id)
}

func (s ticketServiceStub) ListByCompany(ctx context.Context, companyID uuid.UUID, status *domain.TicketStatus, limit int, cursor *domain.TicketCursor) ([]*domain.Ticket, error) {
	if s.listByCompany == nil {
		return nil, errors.New("not implemented")
	}
	return s.listByCompany(ctx, companyID, status, limit, cursor)
}

func (s ticketServiceStub) ListByProject(ctx context.Context, companyID, projectID uuid.UUID, limit int, cursor *domain.TicketCursor) ([]*domain.Ticket, error) {
	if s.listByProject == nil {
		return nil, errors.New("not implemented")
	}
	return s.listByProject(ctx, companyID, projectID, limit, cursor)
}

func (s ticketServiceStub) ListLineItems(ctx context.Context, companyID, ticketID uuid.UUID) ([]*domain.TicketLineItem, error) {
	return nil, errors.New("not implemented")
}

func (s ticketServiceStub) Update(ctx context.Context, companyID, id, actorID uuid.UUID, input *domain.UpdateTicketInput) (*domain.Ticket, error) {
	return nil, errors.New("not implemented")
}

func (s ticketServiceStub) CreateLineItem(ctx context.Context, companyID, ticketID, actorID uuid.UUID, input *domain.CreateLineItemInput) (*domain.TicketLineItem, error) {
	return nil, errors.New("not implemented")
}

func (s ticketServiceStub) UpdateLineItem(ctx context.Context, companyID, ticketID, lineItemID, actorID uuid.UUID, input *domain.UpdateLineItemInput) (*domain.TicketLineItem, error) {
	return nil, errors.New("not implemented")
}

func (s ticketServiceStub) DeleteLineItem(ctx context.Context, companyID, ticketID, lineItemID, actorID uuid.UUID) error {
	return errors.New("not implemented")
}

func (s ticketServiceStub) SubmitForReview(ctx context.Context, companyID, id, actorID uuid.UUID) (*domain.Ticket, error) {
	return nil, errors.New("not implemented")
}

func (s ticketServiceStub) Reject(ctx context.Context, companyID, id, rejectorID uuid.UUID, reason string) (*domain.Ticket, error) {
	return nil, errors.New("not implemented")
}

func TestTicketHandlerGetDeniesCrossCompany(t *testing.T) {
	companyID := uuid.New()
	otherCompanyID := uuid.New()
	ticketID := uuid.New()

	handler := NewTicketHandler(ticketServiceStub{
		getByID: func(ctx context.Context, companyIDArg, id uuid.UUID) (*domain.Ticket, error) {
			return &domain.Ticket{
				ID:        id,
				CompanyID: otherCompanyID,
			}, nil
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/tickets/"+ticketID.String(), nil)
	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add("id", ticketID.String())
	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx)
	ctx = middleware.WithUserContext(ctx, uuid.New(), companyID, domain.RoleAdmin)
	req = req.WithContext(ctx)

	recorder := httptest.NewRecorder()
	handler.Get(recorder, req)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, recorder.Code)
	}
}

func TestTicketHandlerListInvalidStatus(t *testing.T) {
	handler := NewTicketHandler(ticketServiceStub{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/tickets?status=bogus", nil)
	ctx := middleware.WithUserContext(req.Context(), uuid.New(), uuid.New(), domain.RoleAdmin)
	req = req.WithContext(ctx)

	recorder := httptest.NewRecorder()
	handler.List(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
	}
}

func TestTicketHandlerListInvalidCursor(t *testing.T) {
	handler := NewTicketHandler(ticketServiceStub{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/tickets?cursor=bad", nil)
	ctx := middleware.WithUserContext(req.Context(), uuid.New(), uuid.New(), domain.RoleAdmin)
	req = req.WithContext(ctx)

	recorder := httptest.NewRecorder()
	handler.List(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
	}
}
