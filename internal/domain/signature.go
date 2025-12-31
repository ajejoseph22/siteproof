package domain

import (
	"net"
	"time"

	"github.com/google/uuid"
)

// SignerType represents who is signing.
type SignerType string

const (
	SignerTypeForeman SignerType = "foreman"
	SignerTypeGCSuper SignerType = "gc_super"
)

// IsValid returns true if the signer type is valid.
func (s SignerType) IsValid() bool {
	return s == SignerTypeForeman || s == SignerTypeGCSuper
}

// SigningMethod represents how the signature was captured.
type SigningMethod string

const (
	SigningMethodLinkOwnDevice      SigningMethod = "link_own_device"
	SigningMethodLinkSameDevice     SigningMethod = "link_same_device"
	SigningMethodSameDeviceVerified SigningMethod = "same_device_verified"
)

// IsValid returns true if the signing method is valid.
func (m SigningMethod) IsValid() bool {
	switch m {
	case SigningMethodLinkOwnDevice, SigningMethodLinkSameDevice, SigningMethodSameDeviceVerified:
		return true
	}
	return false
}

// Signature represents a digital signature on a ticket.
type Signature struct {
	ID             uuid.UUID     `json:"id"`
	TicketID       uuid.UUID     `json:"ticket_id"`
	SigningLinkID  *uuid.UUID    `json:"signing_link_id,omitempty"`
	SignerID       *uuid.UUID    `json:"signer_id,omitempty"`
	SignerType     SignerType    `json:"signer_type"`
	SignerName     string        `json:"signer_name"`
	SignerEmail    *string       `json:"signer_email,omitempty"`
	SignatureImage string        `json:"signature_image"`
	SignatureHash  string        `json:"signature_hash"`
	SigningMethod  SigningMethod `json:"signing_method"`
	IPAddress      *net.IP       `json:"ip_address,omitempty"`
	UserAgent      *string       `json:"user_agent,omitempty"`
	DeviceID       *string       `json:"device_id,omitempty"`
	SignedAt       time.Time     `json:"signed_at"`
}

// SigningLink represents a secure link for GC signature.
type SigningLink struct {
	ID                   uuid.UUID  `json:"id"`
	TicketID             uuid.UUID  `json:"ticket_id"`
	Token                string     `json:"-"` // Never expose
	SignerType           SignerType `json:"signer_type"`
	SignerEmail          string     `json:"signer_email"`
	SignerPhone          *string    `json:"signer_phone,omitempty"`
	SentVia              []string   `json:"sent_via"`
	SentAt               time.Time  `json:"sent_at"`
	ExpiresAt            time.Time  `json:"expires_at"`
	UsedAt               *time.Time `json:"used_at,omitempty"`
	RevokedAt            *time.Time `json:"revoked_at,omitempty"`
	VerificationCode     *string    `json:"-"` // Never expose
	VerificationAttempts int        `json:"verification_attempts"`
	CreatedAt            time.Time  `json:"created_at"`
}

// IsExpired returns true if the signing link has expired.
func (l *SigningLink) IsExpired() bool {
	return time.Now().After(l.ExpiresAt)
}

// IsUsed returns true if the signing link has been used.
func (l *SigningLink) IsUsed() bool {
	return l.UsedAt != nil
}

// IsRevoked returns true if the signing link has been revoked.
func (l *SigningLink) IsRevoked() bool {
	return l.RevokedAt != nil
}

// IsValid returns true if the signing link can still be used.
func (l *SigningLink) IsValid() bool {
	return !l.IsExpired() && !l.IsUsed() && !l.IsRevoked()
}

// CanAttemptVerification returns true if more verification attempts are allowed.
func (l *SigningLink) CanAttemptVerification() bool {
	return l.VerificationAttempts < 3
}

// CreateSigningLinkInput represents input for creating a signing link.
type CreateSigningLinkInput struct {
	TicketID    uuid.UUID
	SignerType  SignerType
	SignerEmail string
	SignerPhone *string
}

// Validate validates the create signing link input.
func (i *CreateSigningLinkInput) Validate() error {
	var errs ValidationErrors

	if i.TicketID == uuid.Nil {
		errs.Add("ticket_id", "is required")
	}
	if !i.SignerType.IsValid() {
		errs.Add("signer_type", "is invalid")
	}
	if i.SignerEmail == "" {
		errs.Add("signer_email", "is required")
	}

	if errs.HasErrors() {
		return errs
	}
	return nil
}

// CreateSignatureInput represents input for creating a signature.
type CreateSignatureInput struct {
	TicketID       uuid.UUID
	SigningLinkID  *uuid.UUID
	SignerID       *uuid.UUID
	SignerType     SignerType
	SignerName     string
	SignerEmail    *string
	SignatureImage string
	SignatureHash  string
	SigningMethod  SigningMethod
	IPAddress      *net.IP
	UserAgent      *string
	DeviceID       *string
}

// Validate validates the create signature input.
func (i *CreateSignatureInput) Validate() error {
	var errs ValidationErrors

	if i.TicketID == uuid.Nil {
		errs.Add("ticket_id", "is required")
	}
	if !i.SignerType.IsValid() {
		errs.Add("signer_type", "is invalid")
	}
	if i.SignerName == "" {
		errs.Add("signer_name", "is required")
	}
	if i.SignatureImage == "" {
		errs.Add("signature_image", "is required")
	}
	if len(i.SignatureHash) != 64 {
		errs.Add("signature_hash", "must be a valid SHA-256 hash")
	}
	if !i.SigningMethod.IsValid() {
		errs.Add("signing_method", "is invalid")
	}

	if errs.HasErrors() {
		return errs
	}
	return nil
}

