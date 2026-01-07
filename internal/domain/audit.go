package domain

import (
	"encoding/json"
	"net"
	"time"

	"github.com/google/uuid"
)

// AuditAction represents the type of action being audited.
type AuditAction string

const (
	AuditActionCreated            AuditAction = "created"
	AuditActionUpdated            AuditAction = "updated"
	AuditActionSubmitted          AuditAction = "submitted"
	AuditActionApproved           AuditAction = "approved"
	AuditActionDeleted            AuditAction = "deleted"
	AuditActionSigned             AuditAction = "signed"
	AuditActionLocked             AuditAction = "locked"
	AuditActionRejected           AuditAction = "rejected"
	AuditActionPhotoAdded         AuditAction = "photo_added"
	AuditActionDescriptionCleaned AuditAction = "description_cleaned"
	AuditActionEmailSent          AuditAction = "email_sent"
	AuditActionLineItemAdded      AuditAction = "line_item_added"
	AuditActionLineItemUpdated    AuditAction = "line_item_updated"
	AuditActionLineItemDeleted    AuditAction = "line_item_deleted"
	AuditActionLogin              AuditAction = "login"
	AuditActionLogout             AuditAction = "logout"
)

// AuditLog represents an immutable audit log entry.
type AuditLog struct {
	ID         int64           `json:"id"`
	EntityType string          `json:"entity_type"`
	EntityID   uuid.UUID       `json:"entity_id"`
	Action     AuditAction     `json:"action"`
	ActorID    *uuid.UUID      `json:"actor_id,omitempty"`
	ActorIP    *net.IP         `json:"actor_ip,omitempty"`
	OccurredAt time.Time       `json:"occurred_at"`
	Changes    json.RawMessage `json:"changes,omitempty"`
	PrevHash   *string         `json:"prev_hash,omitempty"`
	EntryHash  string          `json:"entry_hash"`
}

// CreateAuditLogInput represents input for creating an audit log entry.
type CreateAuditLogInput struct {
	EntityType string
	EntityID   uuid.UUID
	Action     AuditAction
	ActorID    *uuid.UUID
	ActorIP    *net.IP
	Changes    map[string]ChangeValue
}

// ChangeValue represents a changed field with old and new values.
type ChangeValue struct {
	Old interface{} `json:"old"`
	New interface{} `json:"new"`
}

// ToJSON converts changes to JSON for storage.
func (i *CreateAuditLogInput) ToJSON() (json.RawMessage, error) {
	if len(i.Changes) == 0 {
		return nil, nil
	}
	return json.Marshal(i.Changes)
}
