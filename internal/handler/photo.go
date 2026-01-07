package handler

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/ajejoseph22/siteproof/internal/domain"
	"github.com/ajejoseph22/siteproof/internal/middleware"
)

const (
	photoMetadataMaxBytes = int64(1024 * 1024)
	photoFieldMaxBytes    = int64(32 * 1024)
)

// PhotoService defines photo operations needed by the handler.
type PhotoService interface {
	ListByTicket(ctx context.Context, companyID, ticketID uuid.UUID) ([]*domain.Photo, error)
	Create(ctx context.Context, companyID, ticketID, actorID uuid.UUID, input *domain.CreatePhotoInput) (*domain.Photo, error)
	Delete(ctx context.Context, companyID, ticketID, photoID, actorID uuid.UUID) (*domain.Photo, error)
}

// PhotoHandler handles photo endpoints.
type PhotoHandler struct {
	photoService   PhotoService
	storageBase    string
	maxFileSize    int64
	allowedMimeMap map[string]struct{}
}

// NewPhotoHandler creates a new photo handler.
func NewPhotoHandler(photoService PhotoService, storageBase string, maxFileSize int64, allowedTypes []string) *PhotoHandler {
	allowed := make(map[string]struct{}, len(allowedTypes))
	for _, mimeType := range allowedTypes {
		allowed[mimeType] = struct{}{}
	}

	return &PhotoHandler{
		photoService:   photoService,
		storageBase:    storageBase,
		maxFileSize:    maxFileSize,
		allowedMimeMap: allowed,
	}
}

// List handles GET /api/v1/tickets/{id}/photos
func (h *PhotoHandler) List(w http.ResponseWriter, r *http.Request) {
	ticketID, ok := parseTicketIDParam(w, r)
	if !ok {
		return
	}

	companyID, ok := middleware.CompanyIDFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "not authenticated")
		return
	}

	photos, err := h.photoService.ListByTicket(r.Context(), companyID, ticketID)
	if err != nil {
		WriteErrorFrom(w, err)
		return
	}

	response := make([]PhotoResponse, 0, len(photos))
	for _, photo := range photos {
		response = append(response, newPhotoResponse(r, photo))
	}

	WriteJSON(w, http.StatusOK, response)
}

// Upload handles POST /api/v1/tickets/{id}/photos
func (h *PhotoHandler) Upload(w http.ResponseWriter, r *http.Request) {
	ticketID, ok := parseTicketIDParam(w, r)
	if !ok {
		return
	}

	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "not authenticated")
		return
	}

	companyID, ok := middleware.CompanyIDFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "not authenticated")
		return
	}

	if strings.TrimSpace(h.storageBase) == "" {
		WriteError(w, http.StatusInternalServerError, "STORAGE_NOT_CONFIGURED", "storage base path not configured")
		return
	}

	maxBytes := h.maxFileSize + photoMetadataMaxBytes
	r.Body = http.MaxBytesReader(w, r.Body, maxBytes)

	reader, err := r.MultipartReader()
	if err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_MULTIPART", "invalid multipart form data")
		return
	}

	var (
		captionRaw    string
		takenAtRaw    string
		latitudeRaw   string
		longitudeRaw  string
		clientHashRaw string

		filePath string
		fileSize int64
		fileHash string
		mimeType string
	)

	var parseErr error
	for {
		part, err := reader.NextPart()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			parseErr = domain.NewDomainError(domain.ErrValidation, "INVALID_MULTIPART", "invalid multipart form data")
			break
		}
		if part == nil {
			break
		}

		fieldName := part.FormName()
		if fieldName == "" {
			_ = part.Close()
			continue
		}

		if fieldName != "file" {
			value, err := readPhotoField(part)
			_ = part.Close()
			if err != nil {
				parseErr = domain.NewDomainError(domain.ErrValidation, "INVALID_FIELD", err.Error())
				break
			}

			switch fieldName {
			case "caption":
				captionRaw = value
			case "taken_at":
				takenAtRaw = value
			case "latitude":
				latitudeRaw = value
			case "longitude":
				longitudeRaw = value
			case "client_hash":
				clientHashRaw = value
			}
			continue
		}

		if filePath != "" {
			_ = part.Close()
			parseErr = domain.NewDomainError(domain.ErrValidation, "INVALID_FILE", "multiple files provided")
			break
		}

		filePath, fileSize, fileHash, mimeType, err = h.storePhotoFile(part, companyID, ticketID)
		_ = part.Close()
		if err != nil {
			parseErr = err
			break
		}
	}

	if parseErr != nil {
		h.cleanupFile(filePath)
		WriteErrorFrom(w, parseErr)
		return
	}

	validationErrs := domain.ValidationErrors{}
	if filePath == "" {
		validationErrs.Add("file", "is required")
	}

	clientHash := strings.ToLower(strings.TrimSpace(clientHashRaw))
	if clientHash == "" {
		validationErrs.Add("client_hash", "is required")
	} else if len(clientHash) != 64 {
		validationErrs.Add("client_hash", "must be a valid SHA-256 hash")
	} else if _, err := hex.DecodeString(clientHash); err != nil {
		validationErrs.Add("client_hash", "must be a valid SHA-256 hash")
	}

	if caption := strings.TrimSpace(captionRaw); caption != "" {
		if len(caption) > 500 {
			validationErrs.Add("caption", "must be 500 characters or fewer")
		}
	}

	takenAt, takenAtErr := parseOptionalTime(takenAtRaw)
	if takenAtErr != nil {
		validationErrs.Add("taken_at", "must be an RFC3339 timestamp")
	}

	latitude, latErr := parseOptionalFloat(latitudeRaw, -90, 90)
	if latErr != nil {
		validationErrs.Add("latitude", latErr.Error())
	}

	longitude, lonErr := parseOptionalFloat(longitudeRaw, -180, 180)
	if lonErr != nil {
		validationErrs.Add("longitude", lonErr.Error())
	}

	if validationErrs.HasErrors() {
		h.cleanupFile(filePath)
		WriteErrorFrom(w, validationErrs)
		return
	}

	if clientHash != "" && fileHash != "" && clientHash != fileHash {
		h.cleanupFile(filePath)
		WriteError(w, http.StatusBadRequest, "HASH_MISMATCH", "client hash does not match uploaded file")
		return
	}

	fileSizeInt, err := safeFileSize(fileSize)
	if err != nil {
		h.cleanupFile(filePath)
		WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "file size exceeds limits")
		return
	}

	var caption *string
	if trimmed := strings.TrimSpace(captionRaw); trimmed != "" {
		caption = &trimmed
	}

	input := &domain.CreatePhotoInput{
		TicketID:     ticketID,
		UploadedByID: userID,
		FilePath:     filePath,
		FileSize:     fileSizeInt,
		FileHash:     fileHash,
		MimeType:     mimeType,
		Caption:      caption,
		TakenAt:      takenAt,
		Latitude:     latitude,
		Longitude:    longitude,
	}

	photo, err := h.photoService.Create(r.Context(), companyID, ticketID, userID, input)
	if err != nil {
		h.cleanupFile(filePath)
		WriteErrorFrom(w, err)
		return
	}

	WriteJSON(w, http.StatusCreated, newPhotoResponse(r, photo))
}

// Delete handles DELETE /api/v1/tickets/{id}/photos/{photo_id}
func (h *PhotoHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ticketID, ok := parseTicketIDParam(w, r)
	if !ok {
		return
	}

	photoID, ok := parsePhotoIDParam(w, r)
	if !ok {
		return
	}

	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "not authenticated")
		return
	}

	companyID, ok := middleware.CompanyIDFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "not authenticated")
		return
	}

	photo, err := h.photoService.Delete(r.Context(), companyID, ticketID, photoID, userID)
	if err != nil {
		WriteErrorFrom(w, err)
		return
	}

	if err := h.cleanupFile(photo.FilePath); err != nil {
		slog.Default().Warn("failed to delete photo file", slog.String("error", err.Error()))
	}

	w.WriteHeader(http.StatusNoContent)
}

type PhotoResponse struct {
	ID           uuid.UUID  `json:"id"`
	TicketID     uuid.UUID  `json:"ticket_id"`
	UploadedByID uuid.UUID  `json:"uploaded_by_id"`
	FilePath     string     `json:"file_path"`
	FileURL      string     `json:"file_url"`
	ThumbnailURL string     `json:"thumbnail_url"`
	FileSize     int        `json:"file_size"`
	FileHash     string     `json:"file_hash"`
	MimeType     string     `json:"mime_type"`
	Caption      *string    `json:"caption,omitempty"`
	TakenAt      *time.Time `json:"taken_at,omitempty"`
	Latitude     *float64   `json:"latitude,omitempty"`
	Longitude    *float64   `json:"longitude,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}

func newPhotoResponse(r *http.Request, photo *domain.Photo) PhotoResponse {
	fileURL := buildFileURL(r, photo.FilePath)

	return PhotoResponse{
		ID:           photo.ID,
		TicketID:     photo.TicketID,
		UploadedByID: photo.UploadedByID,
		FilePath:     photo.FilePath,
		FileURL:      fileURL,
		ThumbnailURL: fileURL,
		FileSize:     photo.FileSize,
		FileHash:     photo.FileHash,
		MimeType:     photo.MimeType,
		Caption:      photo.Caption,
		TakenAt:      photo.TakenAt,
		Latitude:     photo.Latitude,
		Longitude:    photo.Longitude,
		CreatedAt:    photo.CreatedAt,
	}
}

func (h *PhotoHandler) storePhotoFile(part *multipart.Part, companyID, ticketID uuid.UUID) (string, int64, string, string, error) {
	reader := bufio.NewReader(part)
	sniff, _ := reader.Peek(512)
	if len(sniff) == 0 {
		return "", 0, "", "", domain.NewDomainError(domain.ErrValidation, "INVALID_FILE", "photo file is empty")
	}

	mimeType := http.DetectContentType(sniff)
	if !h.isAllowedMimeType(mimeType) {
		return "", 0, "", "", domain.NewDomainError(domain.ErrValidation, "INVALID_FILE_TYPE", "unsupported file type")
	}

	extension := extensionForMime(mimeType)
	if extension == "" {
		return "", 0, "", "", domain.NewDomainError(domain.ErrValidation, "INVALID_FILE_TYPE", "unsupported file type")
	}

	filename := uuid.New().String() + extension
	relativePath := filepath.Join(companyID.String(), "photos", ticketID.String(), filename)
	absolutePath := filepath.Join(h.storageBase, relativePath)

	if err := os.MkdirAll(filepath.Dir(absolutePath), 0755); err != nil {
		return "", 0, "", "", fmt.Errorf("create photo directory: %w", err)
	}

	dest, err := os.Create(absolutePath)
	if err != nil {
		return "", 0, "", "", fmt.Errorf("create photo file: %w", err)
	}
	defer dest.Close()

	hash := sha256.New()
	writer := io.MultiWriter(dest, hash)
	written, err := io.Copy(writer, reader)
	if err != nil {
		_ = os.Remove(absolutePath)
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			return "", 0, "", "", domain.NewDomainError(domain.ErrValidation, "FILE_TOO_LARGE", "file exceeds max size")
		}
		return "", 0, "", "", fmt.Errorf("write photo file: %w", err)
	}

	if written <= 0 {
		_ = os.Remove(absolutePath)
		return "", 0, "", "", domain.NewDomainError(domain.ErrValidation, "INVALID_FILE", "photo file is empty")
	}

	if written > h.maxFileSize {
		_ = os.Remove(absolutePath)
		return "", 0, "", "", domain.NewDomainError(domain.ErrValidation, "FILE_TOO_LARGE", "file exceeds max size")
	}

	fileHash := hex.EncodeToString(hash.Sum(nil))
	return filepath.ToSlash(relativePath), written, fileHash, mimeType, nil
}

func (h *PhotoHandler) cleanupFile(relativePath string) error {
	if strings.TrimSpace(relativePath) == "" || strings.TrimSpace(h.storageBase) == "" {
		return nil
	}
	absolutePath := filepath.Join(h.storageBase, filepath.FromSlash(relativePath))
	if err := os.Remove(absolutePath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

func (h *PhotoHandler) isAllowedMimeType(mimeType string) bool {
	if len(h.allowedMimeMap) == 0 {
		return domain.IsValidMimeType(mimeType)
	}
	_, ok := h.allowedMimeMap[mimeType]
	return ok
}

func readPhotoField(part *multipart.Part) (string, error) {
	value, err := io.ReadAll(io.LimitReader(part, photoFieldMaxBytes+1))
	if err != nil {
		return "", err
	}
	if int64(len(value)) > photoFieldMaxBytes {
		return "", errors.New("field value too large")
	}
	return strings.TrimSpace(string(value)), nil
}

func parseOptionalTime(value string) (*time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func parseOptionalFloat(value string, min, max float64) (*float64, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return nil, errors.New("must be a number")
	}
	if parsed < min || parsed > max {
		return nil, fmt.Errorf("must be between %g and %g", min, max)
	}
	return &parsed, nil
}

func parsePhotoIDParam(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	idStr := chi.URLParam(r, "photo_id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid photo id")
		return uuid.UUID{}, false
	}
	return id, true
}

func safeFileSize(size int64) (int, error) {
	if size <= 0 {
		return 0, errors.New("invalid size")
	}
	if size > int64(^uint(0)>>1) {
		return 0, errors.New("size overflow")
	}
	return int(size), nil
}

func extensionForMime(mimeType string) string {
	switch mimeType {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/webp":
		return ".webp"
	case "image/heic":
		return ".heic"
	case "image/heif":
		return ".heif"
	default:
		return ""
	}
}

func buildFileURL(r *http.Request, relativePath string) string {
	relativePath = strings.TrimSpace(relativePath)
	if relativePath == "" {
		return ""
	}

	base := requestBaseURL(r)
	urlPath := path.Join("/uploads", filepath.ToSlash(relativePath))
	if base == "" {
		return urlPath
	}
	return base + urlPath
}

func requestBaseURL(r *http.Request) string {
	host := strings.TrimSpace(r.Host)
	if host == "" {
		return ""
	}

	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	if forwarded := strings.TrimSpace(r.Header.Get("X-Forwarded-Proto")); forwarded != "" {
		scheme = forwarded
	}

	return scheme + "://" + host
}
