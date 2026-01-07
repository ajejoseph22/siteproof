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
	getByID func(ctx context.Context, companyID, id uuid.UUID) (*domain.Ticket, error)
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

func (s ticketServiceStub) ListByProject(ctx context.Context, companyID, projectID uuid.UUID, limit, offset int) ([]*domain.Ticket, error) {
	return nil, errors.New("not implemented")
}

func (s ticketServiceStub) Update(ctx context.Context, companyID, id uuid.UUID, input *domain.UpdateTicketInput) (*domain.Ticket, error) {
	return nil, errors.New("not implemented")
}

func (s ticketServiceStub) SubmitForReview(ctx context.Context, companyID, id uuid.UUID) (*domain.Ticket, error) {
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
