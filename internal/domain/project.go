package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// ProjectStatus represents the status of a project.
type ProjectStatus string

const (
	ProjectStatusActive    ProjectStatus = "active"
	ProjectStatusCompleted ProjectStatus = "completed"
	ProjectStatusArchived  ProjectStatus = "archived"
)

// IsValid returns true if the status is valid.
func (s ProjectStatus) IsValid() bool {
	switch s {
	case ProjectStatusActive, ProjectStatusCompleted, ProjectStatusArchived:
		return true
	}
	return false
}

// Project represents a construction project/job site.
type Project struct {
	ID             uuid.UUID       `json:"id"`
	CompanyID      uuid.UUID       `json:"company_id"`
	Name           string          `json:"name"`
	Address        *string         `json:"address,omitempty"`
	GCCompanyName  *string         `json:"gc_company_name,omitempty"`
	GCContactEmail *string         `json:"gc_contact_email,omitempty"`
	Status         ProjectStatus   `json:"status"`
	Timezone       string          `json:"timezone"`
	Settings       json.RawMessage `json:"settings"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

// CreateProjectInput represents input for creating a new project.
type CreateProjectInput struct {
	CompanyID      uuid.UUID
	Name           string
	Address        *string
	GCCompanyName  *string
	GCContactEmail *string
	Timezone       string
}

// Validate validates the create project input.
func (i *CreateProjectInput) Validate() error {
	var errs ValidationErrors

	if i.Name == "" {
		errs.Add("name", "is required")
	}
	if i.Timezone == "" {
		i.Timezone = "America/New_York" // Default
	}

	if errs.HasErrors() {
		return errs
	}
	return nil
}

// UpdateProjectInput represents input for updating a project.
type UpdateProjectInput struct {
	Name           *string
	Address        *string
	GCCompanyName  *string
	GCContactEmail *string
	Status         *ProjectStatus
	Timezone       *string
}

