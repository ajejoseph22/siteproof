package service

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ajejoseph22/siteproof/internal/domain"
)

// FileService handles file access operations with company-scoped authorization.
type FileService struct {
	pool *pgxpool.Pool
}

// NewFileService creates a new file service.
func NewFileService(pool *pgxpool.Pool) *FileService {
	return &FileService{pool: pool}
}

// GetPhotoByID returns a photo if it belongs to the company.
func (s *FileService) GetPhotoByID(ctx context.Context, companyID, photoID uuid.UUID) (*domain.Photo, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT p.id, p.ticket_id, p.uploaded_by_id, p.file_path, p.file_size, p.file_hash,
		       p.mime_type, p.caption, p.taken_at, p.latitude::float8, p.longitude::float8, p.created_at
		FROM photos p
		INNER JOIN tickets t ON p.ticket_id = t.id
		WHERE p.id = $1 AND t.company_id = $2
	`, photoID, companyID)

	photo, err := scanPhotoRow(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get photo: %w", err)
	}

	return photo, nil
}

// GetSignatureByID returns a signature if it belongs to the company.
func (s *FileService) GetSignatureByID(ctx context.Context, companyID, signatureID uuid.UUID) (*domain.Signature, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT s.id, s.ticket_id, s.signing_link_id, s.signer_id, s.signer_type, s.signer_name,
		       s.signer_email, s.signature_image, s.signature_hash, s.signing_method,
		       s.ip_address, s.user_agent, s.device_id, s.signed_at
		FROM signatures s
		INNER JOIN tickets t ON s.ticket_id = t.id
		WHERE s.id = $1 AND t.company_id = $2
	`, signatureID, companyID)

	signature, err := scanSignatureRow(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get signature: %w", err)
	}

	return signature, nil
}

// GetTicketByID returns a ticket if it belongs to the company.
func (s *FileService) GetTicketByID(ctx context.Context, companyID, ticketID uuid.UUID) (*domain.Ticket, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id, company_id, project_id, created_by_id, ticket_number, ticket_type,
		       title, description_raw, description_cleaned, description_final, work_date,
		       status, gc_signed_at, gc_signed_by_id, rejected_at, rejected_by_id,
		       rejection_reason, pdf_generated_at, pdf_path, pdf_hash, locked_at,
		       content_hash, created_at, updated_at
		FROM tickets
		WHERE id = $1 AND company_id = $2
	`, ticketID, companyID)

	ticket, err := scanTicketRow(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get ticket: %w", err)
	}

	return ticket, nil
}

// scanPhotoRow scans a photo row from the database.
func scanPhotoRow(row pgx.Row) (*domain.Photo, error) {
	var photo domain.Photo
	var latitude, longitude *float64

	err := row.Scan(
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
	)
	if err != nil {
		return nil, err
	}

	photo.Latitude = latitude
	photo.Longitude = longitude

	return &photo, nil
}

// scanSignatureRow scans a signature row from the database.
func scanSignatureRow(row pgx.Row) (*domain.Signature, error) {
	var sig domain.Signature
	var ipStr *string

	err := row.Scan(
		&sig.ID,
		&sig.TicketID,
		&sig.SigningLinkID,
		&sig.SignerID,
		&sig.SignerType,
		&sig.SignerName,
		&sig.SignerEmail,
		&sig.SignatureImage,
		&sig.SignatureHash,
		&sig.SigningMethod,
		&ipStr,
		&sig.UserAgent,
		&sig.DeviceID,
		&sig.SignedAt,
	)
	if err != nil {
		return nil, err
	}

	// Parse IP address if present
	if ipStr != nil {
		parsed := net.ParseIP(strings.TrimSpace(*ipStr))
		if parsed != nil {
			sig.IPAddress = &parsed
		}
	}

	return &sig, nil
}

// scanTicketRow scans a ticket row from the database.
func scanTicketRow(row pgx.Row) (*domain.Ticket, error) {
	var ticket domain.Ticket

	err := row.Scan(
		&ticket.ID,
		&ticket.CompanyID,
		&ticket.ProjectID,
		&ticket.CreatedByID,
		&ticket.TicketNumber,
		&ticket.TicketType,
		&ticket.Title,
		&ticket.DescriptionRaw,
		&ticket.DescriptionCleaned,
		&ticket.DescriptionFinal,
		&ticket.WorkDate,
		&ticket.Status,
		&ticket.GCSignedAt,
		&ticket.GCSignedByID,
		&ticket.RejectedAt,
		&ticket.RejectedByID,
		&ticket.RejectionReason,
		&ticket.PDFGeneratedAt,
		&ticket.PDFPath,
		&ticket.PDFHash,
		&ticket.LockedAt,
		&ticket.ContentHash,
		&ticket.CreatedAt,
		&ticket.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &ticket, nil
}
