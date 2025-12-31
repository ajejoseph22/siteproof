-- Migration: 000012_transcription_versions
-- Description: Create transcription version history for voice-to-text audit trail
-- Author: SiteProof
-- Date: 2024-12-31

CREATE TABLE transcription_versions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ticket_id       UUID NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
    
    -- Version info
    version         INTEGER NOT NULL,
    content         TEXT NOT NULL,
    source          TEXT NOT NULL,
    
    -- LLM metadata (if applicable)
    llm_model       TEXT,
    llm_prompt_hash TEXT,
    
    -- Who created this version
    created_by_id   UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    
    -- Constraints
    CONSTRAINT transcription_versions_unique UNIQUE (ticket_id, version),
    CONSTRAINT transcription_versions_version_positive CHECK (version > 0),
    CONSTRAINT transcription_versions_source_valid CHECK (source IN (
        'voice_raw',    -- Original STT output
        'llm_cleaned',  -- After GPT cleanup
        'user_edit'     -- Manual edit by user
    )),
    CONSTRAINT transcription_versions_llm_metadata CHECK (
        (source != 'llm_cleaned') OR (llm_model IS NOT NULL)
    )
);

-- Indexes
CREATE INDEX idx_transcription_ticket ON transcription_versions(ticket_id);
CREATE INDEX idx_transcription_version ON transcription_versions(ticket_id, version DESC);

-- Comments
COMMENT ON TABLE transcription_versions IS 'Version history for ticket descriptions from voice transcription';
COMMENT ON COLUMN transcription_versions.version IS 'Version number: 1=raw, 2=cleaned, 3+=user edits';
COMMENT ON COLUMN transcription_versions.source IS 'How this version was created: voice_raw, llm_cleaned, user_edit';
COMMENT ON COLUMN transcription_versions.llm_model IS 'Model used for cleanup (e.g., gpt-4o-mini)';
COMMENT ON COLUMN transcription_versions.llm_prompt_hash IS 'SHA-256 hash of system prompt for reproducibility';

