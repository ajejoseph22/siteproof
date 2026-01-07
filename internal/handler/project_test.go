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

type projectServiceStub struct {
	getByID func(ctx context.Context, companyID, id uuid.UUID) (*domain.Project, error)
}

func (s projectServiceStub) Create(ctx context.Context, input *domain.CreateProjectInput) (*domain.Project, error) {
	return nil, errors.New("not implemented")
}

func (s projectServiceStub) GetByID(ctx context.Context, companyID, id uuid.UUID) (*domain.Project, error) {
	if s.getByID == nil {
		return nil, errors.New("not implemented")
	}
	return s.getByID(ctx, companyID, id)
}

func (s projectServiceStub) ListByCompany(ctx context.Context, companyID uuid.UUID, limit, offset int) ([]*domain.Project, error) {
	return nil, errors.New("not implemented")
}

func (s projectServiceStub) ListActiveByCompany(ctx context.Context, companyID uuid.UUID) ([]*domain.Project, error) {
	return nil, errors.New("not implemented")
}

func (s projectServiceStub) Update(ctx context.Context, companyID, id uuid.UUID, input *domain.UpdateProjectInput) (*domain.Project, error) {
	return nil, errors.New("not implemented")
}

func TestProjectHandlerGetDeniesCrossCompany(t *testing.T) {
	companyID := uuid.New()
	otherCompanyID := uuid.New()
	projectID := uuid.New()

	handler := NewProjectHandler(projectServiceStub{
		getByID: func(ctx context.Context, companyIDArg, id uuid.UUID) (*domain.Project, error) {
			return &domain.Project{
				ID:        id,
				CompanyID: otherCompanyID,
			}, nil
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/"+projectID.String(), nil)
	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add("id", projectID.String())
	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx)
	ctx = middleware.WithUserContext(ctx, uuid.New(), companyID, domain.RoleAdmin)
	req = req.WithContext(ctx)

	recorder := httptest.NewRecorder()
	handler.Get(recorder, req)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, recorder.Code)
	}
}
