package domain

import (
	"time"

	"github.com/google/uuid"
)

// TranscriptionSource represents how a transcription version was created.
type TranscriptionSource string

const (
	TranscriptionSourceVoiceRaw   TranscriptionSource = "voice_raw"
	TranscriptionSourceLLMCleaned TranscriptionSource = "llm_cleaned"
	TranscriptionSourceUserEdit   TranscriptionSource = "user_edit"
)

// IsValid returns true if the source is valid.
func (s TranscriptionSource) IsValid() bool {
	switch s {
	case TranscriptionSourceVoiceRaw, TranscriptionSourceLLMCleaned, TranscriptionSourceUserEdit:
		return true
	}
	return false
}

// TranscriptionVersion represents a version of a ticket description.
type TranscriptionVersion struct {
	ID            uuid.UUID           `json:"id"`
	TicketID      uuid.UUID           `json:"ticket_id"`
	Version       int                 `json:"version"`
	Content       string              `json:"content"`
	Source        TranscriptionSource `json:"source"`
	LLMModel      *string             `json:"llm_model,omitempty"`
	LLMPromptHash *string             `json:"llm_prompt_hash,omitempty"`
	CreatedByID   *uuid.UUID          `json:"created_by_id,omitempty"`
	CreatedAt     time.Time           `json:"created_at"`
}

// CreateTranscriptionVersionInput represents input for creating a transcription version.
type CreateTranscriptionVersionInput struct {
	TicketID      uuid.UUID
	Version       int
	Content       string
	Source        TranscriptionSource
	LLMModel      *string
	LLMPromptHash *string
	CreatedByID   *uuid.UUID
}

// Validate validates the create transcription version input.
func (i *CreateTranscriptionVersionInput) Validate() error {
	var errs ValidationErrors

	if i.TicketID == uuid.Nil {
		errs.Add("ticket_id", "is required")
	}
	if i.Version < 1 {
		errs.Add("version", "must be at least 1")
	}
	if i.Content == "" {
		errs.Add("content", "is required")
	}
	if !i.Source.IsValid() {
		errs.Add("source", "is invalid")
	}

	if errs.HasErrors() {
		return errs
	}
	return nil
}

