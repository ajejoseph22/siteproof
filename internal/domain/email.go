package domain

import (
	"time"

	"github.com/google/uuid"
)

// EmailStatus represents the delivery status of an email.
type EmailStatus string

const (
	EmailStatusSent          EmailStatus = "sent"
	EmailStatusDelivered     EmailStatus = "delivered"
	EmailStatusOpened        EmailStatus = "opened"
	EmailStatusBounced       EmailStatus = "bounced"
	EmailStatusSpamComplaint EmailStatus = "spam_complaint"
)

// EmailEvent represents an email delivery event.
type EmailEvent struct {
	ID           uuid.UUID   `json:"id"`
	TicketID     *uuid.UUID  `json:"ticket_id,omitempty"`
	MessageID    string      `json:"message_id"`
	ToEmail      string      `json:"to_email"`
	ToName       *string     `json:"to_name,omitempty"`
	Subject      string      `json:"subject"`
	Template     *string     `json:"template,omitempty"`
	Status       EmailStatus `json:"status"`
	SentAt       time.Time   `json:"sent_at"`
	DeliveredAt  *time.Time  `json:"delivered_at,omitempty"`
	OpenedAt     *time.Time  `json:"opened_at,omitempty"`
	BouncedAt    *time.Time  `json:"bounced_at,omitempty"`
	ErrorCode    *string     `json:"error_code,omitempty"`
	ErrorMessage *string     `json:"error_message,omitempty"`
}

// CreateEmailEventInput represents input for creating an email event.
type CreateEmailEventInput struct {
	TicketID  *uuid.UUID
	MessageID string
	ToEmail   string
	ToName    *string
	Subject   string
	Template  *string
}
