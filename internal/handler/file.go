package handler

import (
	"context"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/ajejoseph22/siteproof/internal/domain"
	"github.com/ajejoseph22/siteproof/internal/middleware"
)

// FileDownloadService defines operations needed for secure file downloads.
type FileDownloadService interface {
	// GetPhotoByID returns a photo if it belongs to the company.
	GetPhotoByID(ctx context.Context, companyID, photoID uuid.UUID) (*domain.Photo, error)
	// GetSignatureByID returns a signature if it belongs to the company.
	GetSignatureByID(ctx context.Context, companyID, signatureID uuid.UUID) (*domain.Signature, error)
	// GetTicketByID returns a ticket if it belongs to the company.
	GetTicketByID(ctx context.Context, companyID, ticketID uuid.UUID) (*domain.Ticket, error)
}

// FileHandler handles secure file download endpoints.
type FileHandler struct {
	fileService FileDownloadService
	storageBase string
}

// NewFileHandler creates a new file handler.
func NewFileHandler(fileService FileDownloadService, storageBase string) *FileHandler {
	return &FileHandler{
		fileService: fileService,
		storageBase: storageBase,
	}
}

// DownloadPhoto handles GET /api/v1/files/photos/{photo_id}
func (h *FileHandler) DownloadPhoto(w http.ResponseWriter, r *http.Request) {
	photoID, ok := parseFileIDParam(w, r, "photo_id")
	if !ok {
		return
	}

	companyID, ok := middleware.CompanyIDFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "not authenticated")
		return
	}

	photo, err := h.fileService.GetPhotoByID(r.Context(), companyID, photoID)
	if err != nil {
		if err == domain.ErrNotFound {
			WriteError(w, http.StatusNotFound, "NOT_FOUND", "photo not found")
			return
		}
		WriteErrorFrom(w, err)
		return
	}

	h.serveFile(w, r, photo.FilePath, photo.MimeType, photo.FileHash, true)
}

// DownloadSignature handles GET /api/v1/files/signatures/{signature_id}
func (h *FileHandler) DownloadSignature(w http.ResponseWriter, r *http.Request) {
	signatureID, ok := parseFileIDParam(w, r, "signature_id")
	if !ok {
		return
	}

	companyID, ok := middleware.CompanyIDFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "not authenticated")
		return
	}

	signature, err := h.fileService.GetSignatureByID(r.Context(), companyID, signatureID)
	if err != nil {
		if err == domain.ErrNotFound {
			WriteError(w, http.StatusNotFound, "NOT_FOUND", "signature not found")
			return
		}
		WriteErrorFrom(w, err)
		return
	}

	// Signature images are always PNG. Not marked as immutable since signatures
	// are tied to ticket workflow and could theoretically be regenerated.
	h.serveFile(w, r, signature.SignatureImage, "image/png", signature.SignatureHash, false)
}

// DownloadPDF handles GET /api/v1/files/pdfs/{ticket_id}
func (h *FileHandler) DownloadPDF(w http.ResponseWriter, r *http.Request) {
	ticketID, ok := parseFileIDParam(w, r, "ticket_id")
	if !ok {
		return
	}

	companyID, ok := middleware.CompanyIDFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "not authenticated")
		return
	}

	ticket, err := h.fileService.GetTicketByID(r.Context(), companyID, ticketID)
	if err != nil {
		if err == domain.ErrNotFound {
			WriteError(w, http.StatusNotFound, "NOT_FOUND", "ticket not found")
			return
		}
		WriteErrorFrom(w, err)
		return
	}

	if ticket.PDFPath == nil || *ticket.PDFPath == "" {
		WriteError(w, http.StatusNotFound, "NOT_FOUND", "pdf not generated")
		return
	}

	pdfHash := ""
	if ticket.PDFHash != nil {
		pdfHash = *ticket.PDFHash
	}

	h.serveFile(w, r, *ticket.PDFPath, "application/pdf", pdfHash, false)
}

// serveFile serves a file from the storage directory with proper headers.
// If isImmutablePhoto is true, applies long-lived caching for immutable photo files.
func (h *FileHandler) serveFile(w http.ResponseWriter, r *http.Request, relativePath, contentType, hash string, isImmutablePhoto bool) {
	if strings.TrimSpace(h.storageBase) == "" {
		WriteError(w, http.StatusInternalServerError, "STORAGE_NOT_CONFIGURED", "storage base path not configured")
		return
	}

	if strings.TrimSpace(relativePath) == "" {
		WriteError(w, http.StatusNotFound, "NOT_FOUND", "file not found")
		return
	}

	// Convert relative path to absolute path safely
	absolutePath := filepath.Join(h.storageBase, filepath.FromSlash(relativePath))

	// Security: ensure the path is within storageBase
	absStorageBase, err := filepath.Abs(h.storageBase)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "storage configuration error")
		return
	}

	absFilePath, err := filepath.Abs(absolutePath)
	if err != nil {
		WriteError(w, http.StatusNotFound, "NOT_FOUND", "file not found")
		return
	}

	if !strings.HasPrefix(absFilePath, absStorageBase+string(filepath.Separator)) && absFilePath != absStorageBase {
		WriteError(w, http.StatusNotFound, "NOT_FOUND", "file not found")
		return
	}

	// Open the file
	file, err := os.Open(absFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			WriteError(w, http.StatusNotFound, "NOT_FOUND", "file not found")
			return
		}
		WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to read file")
		return
	}
	defer file.Close()

	// Get file info for size and mod time
	stat, err := file.Stat()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to read file info")
		return
	}

	if stat.IsDir() {
		WriteError(w, http.StatusNotFound, "NOT_FOUND", "file not found")
		return
	}

	// Determine content type
	if contentType == "" {
		contentType = mime.TypeByExtension(filepath.Ext(absFilePath))
		if contentType == "" {
			contentType = "application/octet-stream"
		}
	}

	// Set headers
	w.Header().Set("Content-Type", contentType)

	// Set ETag if we have a hash (for caching)
	if hash != "" {
		w.Header().Set("ETag", `"`+hash+`"`)
	}

	// Cache control: private (requires auth).
	// Photos are immutable (keyed by ID with content hash) and can be cached for 1 year.
	// Other files (PDFs, signatures) use shorter cache duration as they may be regenerated.
	if isImmutablePhoto {
		w.Header().Set("Cache-Control", "private, max-age=31536000, immutable")
	} else {
		w.Header().Set("Cache-Control", "private, max-age=3600")
	}

	// Content-Disposition for PDFs (triggers download)
	if contentType == "application/pdf" {
		filename := filepath.Base(absFilePath)
		w.Header().Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
	}

	// Handle conditional requests (If-None-Match)
	if hash != "" {
		ifNoneMatch := r.Header.Get("If-None-Match")
		if ifNoneMatch == `"`+hash+`"` || ifNoneMatch == hash {
			w.WriteHeader(http.StatusNotModified)
			return
		}
	}

	// Serve the file with range support
	http.ServeContent(w, r, filepath.Base(absFilePath), stat.ModTime(), file)
}

// parseFileIDParam parses a UUID parameter from the URL.
func parseFileIDParam(w http.ResponseWriter, r *http.Request, paramName string) (uuid.UUID, bool) {
	idStr := chi.URLParam(r, paramName)
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid "+strings.ReplaceAll(paramName, "_", " "))
		return uuid.UUID{}, false
	}
	return id, true
}
