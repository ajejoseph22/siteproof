package domain

import (
	"time"

	"github.com/google/uuid"
)

// Photo represents a photo attached to a ticket.
type Photo struct {
	ID           uuid.UUID  `json:"id"`
	TicketID     uuid.UUID  `json:"ticket_id"`
	UploadedByID uuid.UUID  `json:"uploaded_by_id"`
	FilePath     string     `json:"file_path"`
	FileSize     int        `json:"file_size"`
	FileHash     string     `json:"file_hash"`
	MimeType     string     `json:"mime_type"`
	Caption      *string    `json:"caption,omitempty"`
	TakenAt      *time.Time `json:"taken_at,omitempty"`
	Latitude     *float64   `json:"latitude,omitempty"`
	Longitude    *float64   `json:"longitude,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}

// CreatePhotoInput represents input for creating a photo record.
type CreatePhotoInput struct {
	TicketID     uuid.UUID
	UploadedByID uuid.UUID
	FilePath     string
	FileSize     int
	FileHash     string
	MimeType     string
	Caption      *string
	TakenAt      *time.Time
	Latitude     *float64
	Longitude    *float64
}

// Validate validates the create photo input.
func (i *CreatePhotoInput) Validate() error {
	var errs ValidationErrors

	if i.TicketID == uuid.Nil {
		errs.Add("ticket_id", "is required")
	}
	if i.FilePath == "" {
		errs.Add("file_path", "is required")
	}
	if i.FileSize <= 0 {
		errs.Add("file_size", "must be positive")
	}
	if i.FileHash == "" || len(i.FileHash) != 64 {
		errs.Add("file_hash", "must be a valid SHA-256 hash")
	}
	if i.MimeType == "" {
		errs.Add("mime_type", "is required")
	}

	if errs.HasErrors() {
		return errs
	}
	return nil
}

// IsValidMimeType returns true if the mime type is allowed.
func IsValidMimeType(mimeType string) bool {
	allowed := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/webp": true,
		"image/heic": true,
		"image/heif": true,
	}
	return allowed[mimeType]
}
