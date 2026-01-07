package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ajejoseph22/siteproof/internal/domain"
)

// ProjectService handles project operations.
type ProjectService struct {
	pool *pgxpool.Pool
}

// NewProjectService creates a new project service.
func NewProjectService(pool *pgxpool.Pool) *ProjectService {
	return &ProjectService{pool: pool}
}

// Create creates a new project.
func (s *ProjectService) Create(ctx context.Context, input *domain.CreateProjectInput) (*domain.Project, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	var project domain.Project
	err := s.pool.QueryRow(ctx, `
		INSERT INTO projects (company_id, name, address, gc_company_name, gc_contact_email, status, timezone, settings)
		VALUES ($1, $2, $3, $4, $5, 'active', $6, '{}')
		RETURNING id, company_id, name, address, gc_company_name, gc_contact_email, status, timezone, settings, created_at, updated_at
	`, input.CompanyID, input.Name, input.Address, input.GCCompanyName, input.GCContactEmail, input.Timezone).Scan(
		&project.ID, &project.CompanyID, &project.Name, &project.Address,
		&project.GCCompanyName, &project.GCContactEmail, &project.Status,
		&project.Timezone, &project.Settings, &project.CreatedAt, &project.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create project: %w", err)
	}

	return &project, nil
}

// GetByID retrieves a project by ID.
func (s *ProjectService) GetByID(ctx context.Context, companyID, id uuid.UUID) (*domain.Project, error) {
	var project domain.Project
	err := s.pool.QueryRow(ctx, `
		SELECT id, company_id, name, address, gc_company_name, gc_contact_email,
		       status, timezone, settings, created_at, updated_at
		FROM projects WHERE id = $1 AND company_id = $2
	`, id, companyID).Scan(
		&project.ID, &project.CompanyID, &project.Name, &project.Address,
		&project.GCCompanyName, &project.GCContactEmail, &project.Status,
		&project.Timezone, &project.Settings, &project.CreatedAt, &project.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get project: %w", err)
	}

	return &project, nil
}

// ListByCompany retrieves projects for a company.
func (s *ProjectService) ListByCompany(ctx context.Context, companyID uuid.UUID, limit, offset int) ([]*domain.Project, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, company_id, name, address, gc_company_name, gc_contact_email,
		       status, timezone, settings, created_at, updated_at
		FROM projects
		WHERE company_id = $1
		ORDER BY 
		    CASE status WHEN 'active' THEN 0 WHEN 'completed' THEN 1 ELSE 2 END,
		    name
		LIMIT $2 OFFSET $3
	`, companyID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list projects: %w", err)
	}
	defer rows.Close()

	var projects []*domain.Project
	for rows.Next() {
		var p domain.Project
		if err := rows.Scan(
			&p.ID, &p.CompanyID, &p.Name, &p.Address,
			&p.GCCompanyName, &p.GCContactEmail, &p.Status,
			&p.Timezone, &p.Settings, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan project: %w", err)
		}
		projects = append(projects, &p)
	}

	return projects, nil
}

// ListActiveByCompany retrieves active projects for a company.
func (s *ProjectService) ListActiveByCompany(ctx context.Context, companyID uuid.UUID) ([]*domain.Project, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, company_id, name, address, gc_company_name, gc_contact_email,
		       status, timezone, settings, created_at, updated_at
		FROM projects
		WHERE company_id = $1 AND status = 'active'
		ORDER BY name
	`, companyID)
	if err != nil {
		return nil, fmt.Errorf("list active projects: %w", err)
	}
	defer rows.Close()

	var projects []*domain.Project
	for rows.Next() {
		var p domain.Project
		if err := rows.Scan(
			&p.ID, &p.CompanyID, &p.Name, &p.Address,
			&p.GCCompanyName, &p.GCContactEmail, &p.Status,
			&p.Timezone, &p.Settings, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan project: %w", err)
		}
		projects = append(projects, &p)
	}

	return projects, nil
}

// Update updates a project.
func (s *ProjectService) Update(ctx context.Context, companyID, id uuid.UUID, input *domain.UpdateProjectInput) (*domain.Project, error) {
	var project domain.Project
	err := s.pool.QueryRow(ctx, `
		UPDATE projects
		SET name = COALESCE($2, name),
		    address = COALESCE($3, address),
		    gc_company_name = COALESCE($4, gc_company_name),
		    gc_contact_email = COALESCE($5, gc_contact_email),
		    status = COALESCE($6, status),
		    timezone = COALESCE($7, timezone),
		    updated_at = now()
		WHERE id = $1 AND company_id = $8
		RETURNING id, company_id, name, address, gc_company_name, gc_contact_email,
		          status, timezone, settings, created_at, updated_at
	`, id, input.Name, input.Address, input.GCCompanyName, input.GCContactEmail,
		input.Status, input.Timezone, companyID).Scan(
		&project.ID, &project.CompanyID, &project.Name, &project.Address,
		&project.GCCompanyName, &project.GCContactEmail, &project.Status,
		&project.Timezone, &project.Settings, &project.CreatedAt, &project.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("update project: %w", err)
	}

	return &project, nil
}
