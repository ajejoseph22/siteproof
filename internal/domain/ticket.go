package domain

import (
	"time"

	"github.com/google/uuid"
)

// TicketStatus represents the workflow status of a ticket.
type TicketStatus string

const (
	TicketStatusDraft         TicketStatus = "draft"
	TicketStatusPendingReview TicketStatus = "pending_review"
	TicketStatusApproved      TicketStatus = "approved"
	TicketStatusRejected      TicketStatus = "rejected"
	TicketStatusLocked        TicketStatus = "locked"
)

// IsValid returns true if the status is valid.
func (s TicketStatus) IsValid() bool {
	switch s {
	case TicketStatusDraft, TicketStatusPendingReview, TicketStatusApproved, TicketStatusRejected, TicketStatusLocked:
		return true
	}
	return false
}

// CanTransitionTo returns true if this status can transition to the target status.
func (s TicketStatus) CanTransitionTo(target TicketStatus) bool {
	switch s {
	case TicketStatusDraft:
		return target == TicketStatusPendingReview
	case TicketStatusPendingReview:
		return target == TicketStatusApproved || target == TicketStatusRejected || target == TicketStatusDraft
	case TicketStatusApproved:
		return target == TicketStatusLocked
	case TicketStatusRejected:
		return target == TicketStatusDraft
	case TicketStatusLocked:
		return false // No transitions from locked
	}
	return false
}

// TicketType represents the type of ticket.
type TicketType string

const (
	TicketTypeTimeMaterials TicketType = "time_materials"
	TicketTypeVariation     TicketType = "variation"
)

// IsValid returns true if the ticket type is valid.
func (t TicketType) IsValid() bool {
	return t == TicketTypeTimeMaterials || t == TicketTypeVariation
}

// Ticket represents a T&M or variation ticket.
type Ticket struct {
	ID                 uuid.UUID    `json:"id"`
	CompanyID          uuid.UUID    `json:"company_id"`
	ProjectID          uuid.UUID    `json:"project_id"`
	CreatedByID        uuid.UUID    `json:"created_by_id"`
	TicketNumber       string       `json:"ticket_number"`
	TicketType         TicketType   `json:"ticket_type"`
	Title              string       `json:"title"`
	DescriptionRaw     *string      `json:"description_raw,omitempty"`
	DescriptionCleaned *string      `json:"description_cleaned,omitempty"`
	DescriptionFinal   *string      `json:"description_final,omitempty"`
	WorkDate           time.Time    `json:"work_date"`
	Status             TicketStatus `json:"status"`
	GCSignedAt         *time.Time   `json:"gc_signed_at,omitempty"`
	GCSignedByID       *uuid.UUID   `json:"gc_signed_by_id,omitempty"`
	RejectedAt         *time.Time   `json:"rejected_at,omitempty"`
	RejectedByID       *uuid.UUID   `json:"rejected_by_id,omitempty"`
	RejectionReason    *string      `json:"rejection_reason,omitempty"`
	PDFGeneratedAt     *time.Time   `json:"pdf_generated_at,omitempty"`
	PDFPath            *string      `json:"pdf_path,omitempty"`
	PDFHash            *string      `json:"pdf_hash,omitempty"`
	LockedAt           *time.Time   `json:"locked_at,omitempty"`
	ContentHash        *string      `json:"content_hash,omitempty"`
	CreatedAt          time.Time    `json:"created_at"`
	UpdatedAt          time.Time    `json:"updated_at"`
}

// TicketCursor represents a cursor for ticket pagination.
type TicketCursor struct {
	Time time.Time
	ID   uuid.UUID
}

// IsLocked returns true if the ticket is locked.
func (t *Ticket) IsLocked() bool {
	return t.Status == TicketStatusLocked
}

// CanEdit returns true if the ticket can be edited.
func (t *Ticket) CanEdit() bool {
	return t.Status == TicketStatusDraft
}

// CreateTicketInput represents input for creating a new ticket.
type CreateTicketInput struct {
	CompanyID   uuid.UUID
	ProjectID   uuid.UUID
	CreatedByID uuid.UUID
	TicketType  TicketType
	Title       string
	WorkDate    time.Time
}

// Validate validates the create ticket input.
func (i *CreateTicketInput) Validate() error {
	var errs ValidationErrors

	if i.ProjectID == uuid.Nil {
		errs.Add("project_id", "is required")
	}
	if !i.TicketType.IsValid() {
		errs.Add("ticket_type", "must be time_materials or variation")
	}
	if i.Title == "" {
		errs.Add("title", "is required")
	}
	if i.WorkDate.IsZero() {
		errs.Add("work_date", "is required")
	}

	if errs.HasErrors() {
		return errs
	}
	return nil
}

// UpdateTicketInput represents input for updating a ticket.
type UpdateTicketInput struct {
	Title            *string
	DescriptionFinal *string
	WorkDate         *time.Time
}

// LineItemType represents the type of line item.
type LineItemType string

const (
	LineItemTypeLabor     LineItemType = "labor"
	LineItemTypeMaterial  LineItemType = "material"
	LineItemTypeEquipment LineItemType = "equipment"
)

// IsValid returns true if the line item type is valid.
func (t LineItemType) IsValid() bool {
	switch t {
	case LineItemTypeLabor, LineItemTypeMaterial, LineItemTypeEquipment:
		return true
	}
	return false
}

// TicketLineItem represents a line item on a ticket.
type TicketLineItem struct {
	ID          uuid.UUID    `json:"id"`
	TicketID    uuid.UUID    `json:"ticket_id"`
	ItemType    LineItemType `json:"item_type"`
	Description string       `json:"description"`
	Quantity    float64      `json:"quantity"`
	Unit        string       `json:"unit"`
	UnitRate    *float64     `json:"unit_rate,omitempty"`
	Total       *float64     `json:"total,omitempty"`
	SortOrder   int          `json:"sort_order"`
	CreatedAt   time.Time    `json:"created_at"`
}

// CreateLineItemInput represents input for creating a line item.
type CreateLineItemInput struct {
	TicketID    uuid.UUID
	ItemType    LineItemType
	Description string
	Quantity    float64
	Unit        string
	UnitRate    *float64
	SortOrder   int
}

// Validate validates the create line item input.
func (i *CreateLineItemInput) Validate() error {
	var errs ValidationErrors

	if !i.ItemType.IsValid() {
		errs.Add("item_type", "must be labor, material, or equipment")
	}
	if i.Description == "" {
		errs.Add("description", "is required")
	}
	if i.Quantity <= 0 {
		errs.Add("quantity", "must be positive")
	}
	if i.Unit == "" {
		errs.Add("unit", "is required")
	}

	if errs.HasErrors() {
		return errs
	}
	return nil
}
