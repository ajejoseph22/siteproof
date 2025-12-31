-- Migration: 000007_photos
-- Description: Create photos table for ticket evidence
-- Author: SiteProof
-- Date: 2024-12-31

CREATE TABLE photos (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ticket_id       UUID NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
    uploaded_by_id  UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    
    -- File details
    file_path       TEXT NOT NULL,
    file_size       INTEGER NOT NULL,
    file_hash       TEXT NOT NULL,
    mime_type       TEXT NOT NULL,
    
    -- Metadata
    caption         TEXT,
    taken_at        TIMESTAMPTZ,
    latitude        DECIMAL(9, 6),
    longitude       DECIMAL(9, 6),
    
    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    
    -- Constraints
    CONSTRAINT photos_file_path_not_empty CHECK (length(trim(file_path)) > 0),
    CONSTRAINT photos_file_size_positive CHECK (file_size > 0),
    CONSTRAINT photos_file_hash_not_empty CHECK (length(file_hash) = 64),  -- SHA-256 hex
    CONSTRAINT photos_mime_type_valid CHECK (mime_type IN (
        'image/jpeg',
        'image/png',
        'image/webp',
        'image/heic',
        'image/heif'
    )),
    CONSTRAINT photos_latitude_valid CHECK (latitude IS NULL OR (latitude >= -90 AND latitude <= 90)),
    CONSTRAINT photos_longitude_valid CHECK (longitude IS NULL OR (longitude >= -180 AND longitude <= 180))
);

-- Indexes
CREATE INDEX idx_photos_ticket ON photos(ticket_id);
CREATE INDEX idx_photos_uploaded_by ON photos(uploaded_by_id);
CREATE INDEX idx_photos_hash ON photos(file_hash);  -- For deduplication

-- Comments
COMMENT ON TABLE photos IS 'Photo evidence attached to tickets';
COMMENT ON COLUMN photos.file_hash IS 'SHA-256 hash of file content for integrity verification';
COMMENT ON COLUMN photos.taken_at IS 'EXIF timestamp if available from photo metadata';
COMMENT ON COLUMN photos.latitude IS 'GPS latitude from EXIF metadata';
COMMENT ON COLUMN photos.longitude IS 'GPS longitude from EXIF metadata';

