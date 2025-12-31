-- name: GetTranscriptionVersionsByTicket :many
SELECT * FROM transcription_versions
WHERE ticket_id = $1
ORDER BY version;

-- name: GetLatestTranscriptionVersion :one
SELECT * FROM transcription_versions
WHERE ticket_id = $1
ORDER BY version DESC
LIMIT 1;

-- name: GetNextTranscriptionVersion :one
SELECT COALESCE(MAX(version), 0) + 1 AS next_version
FROM transcription_versions
WHERE ticket_id = $1;

-- name: CreateTranscriptionVersion :one
INSERT INTO transcription_versions (
    ticket_id, version, content, source, llm_model, llm_prompt_hash, created_by_id
)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

