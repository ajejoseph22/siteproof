package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ajejoseph22/siteproof/internal/domain"
	"github.com/ajejoseph22/siteproof/pkg/slogctx"
)

// PhotoService handles photo operations.
type PhotoService struct {
	pool *pgxpool.Pool
}

// NewPhotoService creates a new photo service.
func NewPhotoService(pool *pgxpool.Pool) *PhotoService {
	return &PhotoService{pool: pool}
}

// ListByTicket retrieves photos for a ticket.
func (s *PhotoService) ListByTicket(ctx context.Context, companyID, ticketID uuid.UUID) ([]*domain.Photo, error) {
	if err := s.ensureTicketExists(ctx, companyID, ticketID); err != nil {
		return nil, err
	}

	rows, err := s.pool.Query(ctx, `
		SELECT id, ticket_id, uploaded_by_id, file_path, file_size, file_hash, mime_type,
		       caption, taken_at, latitude::float8, longitude::float8, created_at
		FROM photos
		WHERE ticket_id = $1
		ORDER BY created_at DESC
	`, ticketID)
	if err != nil {
		return nil, fmt.Errorf("list photos: %w", err)
	}
	defer rows.Close()

	photos := []*domain.Photo{}
	for rows.Next() {
		photo, err := scanPhoto(rows)
		if err != nil {
			return nil, fmt.Errorf("scan photo: %w", err)
		}
		photos = append(photos, photo)
	}

	return photos, nil
}

// Create creates a new photo record.
func (s *PhotoService) Create(ctx context.Context, companyID, ticketID, actorID uuid.UUID, input *domain.CreatePhotoInput) (*domain.Photo, error) {
	input.TicketID = ticketID
	input.UploadedByID = actorID
	if err := input.Validate(); err != nil {
		return nil, err
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("begin photo create: %w", err)
	}
	defer tx.Rollback(ctx)

	status, err := s.lockTicket(ctx, tx, companyID, ticketID)
	if err != nil {
		return nil, err
	}
	if status != domain.TicketStatusDraft {
		return nil, domain.ErrTicketLocked
	}

	row := tx.QueryRow(ctx, `
		INSERT INTO photos (
			ticket_id, uploaded_by_id, file_path, file_size, file_hash, mime_type,
			caption, taken_at, latitude, longitude
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, ticket_id, uploaded_by_id, file_path, file_size, file_hash, mime_type,
		          caption, taken_at, latitude::float8, longitude::float8, created_at
	`, input.TicketID, input.UploadedByID, input.FilePath, input.FileSize, input.FileHash,
		input.MimeType, input.Caption, input.TakenAt, input.Latitude, input.Longitude)

	photo, err := scanPhoto(row)
	if err != nil {
		return nil, fmt.Errorf("create photo: %w", err)
	}

	if err := s.insertAuditLog(ctx, tx, domain.CreateAuditLogInput{
		EntityType: "ticket",
		EntityID:   ticketID,
		Action:     domain.AuditActionPhotoAdded,
		ActorID:    &actorID,
		Changes:    photoChanges(nil, photo),
	}); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit photo create: %w", err)
	}

	slogctx.Info(ctx, "photo_uploaded",
		slog.String("photo_id", photo.ID.String()),
		slog.String("ticket_id", ticketID.String()),
		slog.String("uploaded_by_id", actorID.String()),
		slog.String("file_hash", photo.FileHash),
		slog.Int("file_size", photo.FileSize),
	)

	return photo, nil
}

// Delete removes a photo record.
func (s *PhotoService) Delete(ctx context.Context, companyID, ticketID, photoID, actorID uuid.UUID) (*domain.Photo, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("begin photo delete: %w", err)
	}
	defer tx.Rollback(ctx)

	status, err := s.lockTicket(ctx, tx, companyID, ticketID)
	if err != nil {
		return nil, err
	}
	if status != domain.TicketStatusDraft {
		return nil, domain.ErrTicketLocked
	}

	row := tx.QueryRow(ctx, `
		SELECT id, ticket_id, uploaded_by_id, file_path, file_size, file_hash, mime_type,
		       caption, taken_at, latitude::float8, longitude::float8, created_at
		FROM photos
		WHERE id = $1 AND ticket_id = $2
	`, photoID, ticketID)

	photo, err := scanPhoto(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get photo: %w", err)
	}

	if _, err := tx.Exec(ctx, `
		DELETE FROM photos WHERE id = $1 AND ticket_id = $2
	`, photoID, ticketID); err != nil {
		return nil, fmt.Errorf("delete photo: %w", err)
	}

	if err := s.insertAuditLog(ctx, tx, domain.CreateAuditLogInput{
		EntityType: "ticket",
		EntityID:   ticketID,
		Action:     domain.AuditActionPhotoDeleted,
		ActorID:    &actorID,
		Changes:    photoChanges(photo, nil),
	}); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit photo delete: %w", err)
	}

	return photo, nil
}

func (s *PhotoService) ensureTicketExists(ctx context.Context, companyID, ticketID uuid.UUID) error {
	var exists bool
	err := s.pool.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM tickets WHERE id = $1 AND company_id = $2
		)
	`, ticketID, companyID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("check ticket exists: %w", err)
	}
	if !exists {
		return domain.ErrNotFound
	}
	return nil
}

func (s *PhotoService) lockTicket(ctx context.Context, tx pgx.Tx, companyID, ticketID uuid.UUID) (domain.TicketStatus, error) {
	var status domain.TicketStatus
	if err := tx.QueryRow(ctx, `
		SELECT status
		FROM tickets
		WHERE id = $1 AND company_id = $2
		FOR UPDATE
	`, ticketID, companyID).Scan(&status); err != nil {
		if err == pgx.ErrNoRows {
			return "", domain.ErrNotFound
		}
		return "", fmt.Errorf("lock ticket: %w", err)
	}

	return status, nil
}

func (s *PhotoService) insertAuditLog(ctx context.Context, tx pgx.Tx, input domain.CreateAuditLogInput) error {
	changesJSON, err := input.ToJSON()
	if err != nil {
		return fmt.Errorf("marshal audit changes: %w", err)
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO audit_logs (entity_type, entity_id, action, actor_id, actor_ip, changes)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, input.EntityType, input.EntityID, input.Action, input.ActorID, input.ActorIP, changesJSON)
	if err != nil {
		return fmt.Errorf("insert audit log: %w", err)
	}

	return nil
}

func scanPhoto(scanner lineItemScanner) (*domain.Photo, error) {
	var photo domain.Photo
	var latitude pgtype.Float8
	var longitude pgtype.Float8

	if err := scanner.Scan(
		&photo.ID,
		&photo.TicketID,
		&photo.UploadedByID,
		&photo.FilePath,
		&photo.FileSize,
		&photo.FileHash,
		&photo.MimeType,
		&photo.Caption,
		&photo.TakenAt,
		&latitude,
		&longitude,
		&photo.CreatedAt,
	); err != nil {
		return nil, err
	}

	photo.Latitude = floatPointer(latitude)
	photo.Longitude = floatPointer(longitude)

	return &photo, nil
}

func photoChanges(oldPhoto, newPhoto *domain.Photo) map[string]domain.ChangeValue {
	changes := map[string]domain.ChangeValue{
		"photo_id":       {Old: photoIDOrNil(oldPhoto), New: photoIDOrNil(newPhoto)},
		"file_path":      {Old: photoPathOrNil(oldPhoto), New: photoPathOrNil(newPhoto)},
		"file_hash":      {Old: photoHashOrNil(oldPhoto), New: photoHashOrNil(newPhoto)},
		"file_size":      {Old: photoFileSizeOrNil(oldPhoto), New: photoFileSizeOrNil(newPhoto)},
		"mime_type":      {Old: photoMimeTypeOrNil(oldPhoto), New: photoMimeTypeOrNil(newPhoto)},
		"caption":        {Old: photoCaptionOrNil(oldPhoto), New: photoCaptionOrNil(newPhoto)},
		"taken_at":       {Old: photoTakenAtOrNil(oldPhoto), New: photoTakenAtOrNil(newPhoto)},
		"latitude":       {Old: photoLatitudeOrNil(oldPhoto), New: photoLatitudeOrNil(newPhoto)},
		"longitude":      {Old: photoLongitudeOrNil(oldPhoto), New: photoLongitudeOrNil(newPhoto)},
		"uploaded_by_id": {Old: photoUploadedByOrNil(oldPhoto), New: photoUploadedByOrNil(newPhoto)},
	}

	return changes
}

func photoIDOrNil(photo *domain.Photo) interface{} {
	if photo == nil {
		return nil
	}
	return photo.ID
}

func photoPathOrNil(photo *domain.Photo) interface{} {
	if photo == nil {
		return nil
	}
	return photo.FilePath
}

func photoHashOrNil(photo *domain.Photo) interface{} {
	if photo == nil {
		return nil
	}
	return photo.FileHash
}

func photoMimeTypeOrNil(photo *domain.Photo) interface{} {
	if photo == nil {
		return nil
	}
	return photo.MimeType
}

func photoCaptionOrNil(photo *domain.Photo) interface{} {
	if photo == nil {
		return nil
	}
	return photo.Caption
}

func photoTakenAtOrNil(photo *domain.Photo) interface{} {
	if photo == nil {
		return nil
	}
	return photo.TakenAt
}

func photoLatitudeOrNil(photo *domain.Photo) interface{} {
	if photo == nil {
		return nil
	}
	return photo.Latitude
}

func photoLongitudeOrNil(photo *domain.Photo) interface{} {
	if photo == nil {
		return nil
	}
	return photo.Longitude
}

func photoFileSizeOrNil(photo *domain.Photo) interface{} {
	if photo == nil {
		return nil
	}
	return photo.FileSize
}

func photoUploadedByOrNil(photo *domain.Photo) interface{} {
	if photo == nil {
		return nil
	}
	return photo.UploadedByID
}
