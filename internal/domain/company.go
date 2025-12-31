package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Company represents a tenant in the multi-tenant system.
type Company struct {
	ID        uuid.UUID       `json:"id"`
	Name      string          `json:"name"`
	Slug      string          `json:"slug"`
	Settings  json.RawMessage `json:"settings"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

// CreateCompanyInput represents input for creating a new company.
type CreateCompanyInput struct {
	Name string
	Slug string
}

// Validate validates the create company input.
func (i *CreateCompanyInput) Validate() error {
	var errs ValidationErrors

	if i.Name == "" {
		errs.Add("name", "is required")
	}
	if i.Slug == "" {
		errs.Add("slug", "is required")
	}

	if errs.HasErrors() {
		return errs
	}
	return nil
}

