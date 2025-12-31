package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/yourusername/siteproof/internal/domain"
	"github.com/yourusername/siteproof/internal/middleware"
	"github.com/yourusername/siteproof/internal/service"
)

// ProjectHandler handles project endpoints.
type ProjectHandler struct {
	projectService *service.ProjectService
}

// NewProjectHandler creates a new project handler.
func NewProjectHandler(projectService *service.ProjectService) *ProjectHandler {
	return &ProjectHandler{projectService: projectService}
}

// CreateProjectRequest represents a project creation request.
type CreateProjectRequest struct {
	Name           string  `json:"name"`
	Address        *string `json:"address,omitempty"`
	GCCompanyName  *string `json:"gc_company_name,omitempty"`
	GCContactEmail *string `json:"gc_contact_email,omitempty"`
	Timezone       string  `json:"timezone"`
}

// Create handles POST /api/v1/projects
func (h *ProjectHandler) Create(w http.ResponseWriter, r *http.Request) {
	companyID, ok := middleware.CompanyIDFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "not authenticated")
		return
	}

	var req CreateProjectRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_JSON", "invalid request body")
		return
	}

	input := &domain.CreateProjectInput{
		CompanyID:      companyID,
		Name:           req.Name,
		Address:        req.Address,
		GCCompanyName:  req.GCCompanyName,
		GCContactEmail: req.GCContactEmail,
		Timezone:       req.Timezone,
	}

	project, err := h.projectService.Create(r.Context(), input)
	if err != nil {
		if _, ok := err.(domain.ValidationErrors); ok {
			WriteErrorWithDetails(w, http.StatusBadRequest, "VALIDATION_ERROR", "validation failed", err)
			return
		}
		WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create project")
		return
	}

	WriteJSON(w, http.StatusCreated, project)
}

// Get handles GET /api/v1/projects/{id}
func (h *ProjectHandler) Get(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid project id")
		return
	}

	companyID, _ := middleware.CompanyIDFromContext(r.Context())

	project, err := h.projectService.GetByID(r.Context(), id)
	if err != nil {
		if err == domain.ErrNotFound {
			WriteError(w, http.StatusNotFound, "NOT_FOUND", "project not found")
			return
		}
		WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get project")
		return
	}

	// Verify company access
	if project.CompanyID != companyID {
		WriteError(w, http.StatusNotFound, "NOT_FOUND", "project not found")
		return
	}

	WriteJSON(w, http.StatusOK, project)
}

// List handles GET /api/v1/projects
func (h *ProjectHandler) List(w http.ResponseWriter, r *http.Request) {
	companyID, ok := middleware.CompanyIDFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "not authenticated")
		return
	}

	limit, offset := parsePagination(r)

	// Check if only active projects requested
	if r.URL.Query().Get("status") == "active" {
		projects, err := h.projectService.ListActiveByCompany(r.Context(), companyID)
		if err != nil {
			WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list projects")
			return
		}
		WriteJSON(w, http.StatusOK, projects)
		return
	}

	projects, err := h.projectService.ListByCompany(r.Context(), companyID, limit, offset)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list projects")
		return
	}

	WriteJSON(w, http.StatusOK, projects)
}

// UpdateProjectRequest represents a project update request.
type UpdateProjectRequest struct {
	Name           *string `json:"name,omitempty"`
	Address        *string `json:"address,omitempty"`
	GCCompanyName  *string `json:"gc_company_name,omitempty"`
	GCContactEmail *string `json:"gc_contact_email,omitempty"`
	Status         *string `json:"status,omitempty"`
	Timezone       *string `json:"timezone,omitempty"`
}

// Update handles PATCH /api/v1/projects/{id}
func (h *ProjectHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid project id")
		return
	}

	var req UpdateProjectRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_JSON", "invalid request body")
		return
	}

	input := &domain.UpdateProjectInput{
		Name:           req.Name,
		Address:        req.Address,
		GCCompanyName:  req.GCCompanyName,
		GCContactEmail: req.GCContactEmail,
		Timezone:       req.Timezone,
	}

	if req.Status != nil {
		status := domain.ProjectStatus(*req.Status)
		if !status.IsValid() {
			WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid status")
			return
		}
		input.Status = &status
	}

	project, err := h.projectService.Update(r.Context(), id, input)
	if err != nil {
		if err == domain.ErrNotFound {
			WriteError(w, http.StatusNotFound, "NOT_FOUND", "project not found")
			return
		}
		WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to update project")
		return
	}

	WriteJSON(w, http.StatusOK, project)
}

