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
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/ajejoseph22/siteproof/internal/domain"
	"github.com/ajejoseph22/siteproof/internal/middleware"
	"github.com/ajejoseph22/siteproof/internal/service"
)

const (
	signatureMetadataMaxBytes = int64(512 * 1024) // 512KB for metadata
	signatureFieldMaxBytes    = int64(32 * 1024)  // 32KB per field
	signatureMaxFileSize      = int64(5 * 1024 * 1024) // 5MB max signature image
)

// SignatureService defines signature operations needed by the handler.
type SignatureService interface {
	CaptureSignature(ctx context.Context, token string, input *service.CaptureSignatureInput) (*domain.Signature, error)
	CaptureInternalSignature(ctx context.Context, companyID, ticketID, signerID uuid.UUID, input *service.CaptureSignatureInput) (*domain.Signature, error)
	ListByTicket(ctx context.Context, companyID, ticketID uuid.UUID) ([]*domain.Signature, error)
	GetByID(ctx context.Context, companyID, signatureID uuid.UUID) (*domain.Signature, error)
}

// SignatureHandler handles signature endpoints.
type SignatureHandler struct {
	signatureService SignatureService
	storageBase      string
}

// NewSignatureHandler creates a new signature handler.
func NewSignatureHandler(signatureService SignatureService, storageBase string) *SignatureHandler {
	return &SignatureHandler{
		signatureService: signatureService,
		storageBase:      storageBase,
	}
}

// SignatureResponse represents a signature in API responses.
type SignatureResponse struct {
	ID            uuid.UUID  `json:"id"`
	TicketID      uuid.UUID  `json:"ticket_id"`
	SignerType    string     `json:"signer_type"`
	SignerName    string     `json:"signer_name"`
	SignerEmail   *string    `json:"signer_email,omitempty"`
	SignatureURL  string     `json:"signature_url"`
	SigningMethod string     `json:"signing_method"`
	SignedAt      time.Time  `json:"signed_at"`
}

func newSignatureResponse(r *http.Request, sig *domain.Signature) SignatureResponse {
	return SignatureResponse{
		ID:            sig.ID,
		TicketID:      sig.TicketID,
		SignerType:    string(sig.SignerType),
		SignerName:    sig.SignerName,
		SignerEmail:   sig.SignerEmail,
		SignatureURL:  buildSignatureFileURL(r, sig.ID),
		SigningMethod: string(sig.SigningMethod),
		SignedAt:      sig.SignedAt,
	}
}

func buildSignatureFileURL(r *http.Request, signatureID uuid.UUID) string {
	base := requestBaseURL(r)
	urlPath := fmt.Sprintf("/api/v1/files/signatures/%s", signatureID.String())
	if base == "" {
		return urlPath
	}
	return base + urlPath
}

// CaptureViaToken handles POST /api/v1/sign/{token}/signature
// This is a public endpoint for GC to submit their signature via signing link.
func (h *SignatureHandler) CaptureViaToken(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	if token == "" {
		WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "token is required")
		return
	}

	if strings.TrimSpace(h.storageBase) == "" {
		WriteError(w, http.StatusInternalServerError, "STORAGE_NOT_CONFIGURED", "storage base path not configured")
		return
	}

	// Parse the multipart form
	input, filePath, err := h.parseSignatureUpload(r)
	if err != nil {
		WriteErrorFrom(w, err)
		return
	}

	// Extract client info
	input.IPAddress = parseClientIP(r)
	ua := r.Header.Get("User-Agent")
	if ua != "" {
		input.UserAgent = &ua
	}

	sig, err := h.signatureService.CaptureSignature(r.Context(), token, input)
	if err != nil {
		h.cleanupFile(filePath)
		WriteErrorFrom(w, err)
		return
	}

	WriteJSON(w, http.StatusCreated, newSignatureResponse(r, sig))
}

// CaptureForeman handles POST /api/v1/tickets/{id}/signatures
// This is a protected endpoint for foremen to sign their own tickets.
func (h *SignatureHandler) CaptureForeman(w http.ResponseWriter, r *http.Request) {
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

	// Parse the multipart form
	input, filePath, err := h.parseSignatureUploadForTicket(r, companyID, ticketID)
	if err != nil {
		WriteErrorFrom(w, err)
		return
	}

	// Extract client info
	input.IPAddress = parseClientIP(r)
	ua := r.Header.Get("User-Agent")
	if ua != "" {
		input.UserAgent = &ua
	}

	sig, err := h.signatureService.CaptureInternalSignature(r.Context(), companyID, ticketID, userID, input)
	if err != nil {
		h.cleanupFile(filePath)
		WriteErrorFrom(w, err)
		return
	}

	WriteJSON(w, http.StatusCreated, newSignatureResponse(r, sig))
}

// List handles GET /api/v1/tickets/{id}/signatures
func (h *SignatureHandler) List(w http.ResponseWriter, r *http.Request) {
	ticketID, ok := parseTicketIDParam(w, r)
	if !ok {
		return
	}

	companyID, ok := middleware.CompanyIDFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "not authenticated")
		return
	}

	signatures, err := h.signatureService.ListByTicket(r.Context(), companyID, ticketID)
	if err != nil {
		WriteErrorFrom(w, err)
		return
	}

	response := make([]SignatureResponse, 0, len(signatures))
	for _, sig := range signatures {
		response = append(response, newSignatureResponse(r, sig))
	}

	WriteJSON(w, http.StatusOK, response)
}

// parseSignatureUpload parses the multipart signature upload for public token-based capture.
// It stores the file in a temporary location since we don't know the ticket ID yet.
func (h *SignatureHandler) parseSignatureUpload(r *http.Request) (*service.CaptureSignatureInput, string, error) {
	maxBytes := signatureMaxFileSize + signatureMetadataMaxBytes
	r.Body = http.MaxBytesReader(nil, r.Body, maxBytes)

	reader, err := r.MultipartReader()
	if err != nil {
		return nil, "", domain.NewDomainError(domain.ErrValidation, "INVALID_MULTIPART", "invalid multipart form data")
	}

	var (
		signerName      string
		signerEmail     string
		signingMethod   string
		deviceID        string
		clientHashRaw   string

		filePath      string
		fileHash      string
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

		if fieldName != "signature" {
			value, err := readSignatureField(part)
			_ = part.Close()
			if err != nil {
				parseErr = domain.NewDomainError(domain.ErrValidation, "INVALID_FIELD", err.Error())
				break
			}

			switch fieldName {
			case "signer_name":
				signerName = value
			case "signer_email":
				signerEmail = value
			case "signing_method":
				signingMethod = value
			case "device_id":
				deviceID = value
			case "client_hash":
				clientHashRaw = value
			}
			continue
		}

		if filePath != "" {
			_ = part.Close()
			parseErr = domain.NewDomainError(domain.ErrValidation, "INVALID_FILE", "multiple signature files provided")
			break
		}

		// Store to temporary location - will be moved after we know the ticket
		filePath, fileHash, err = h.storeSignatureFile(part, "pending", uuid.New())
		_ = part.Close()
		if err != nil {
			parseErr = err
			break
		}
	}

	if parseErr != nil {
		h.cleanupFile(filePath)
		return nil, "", parseErr
	}

	// Validate required fields
	validationErrs := domain.ValidationErrors{}
	if filePath == "" {
		validationErrs.Add("signature", "is required")
	}
	if signerName == "" {
		validationErrs.Add("signer_name", "is required")
	}
	if signingMethod == "" {
		validationErrs.Add("signing_method", "is required")
	}

	clientHash := strings.ToLower(strings.TrimSpace(clientHashRaw))
	if clientHash == "" {
		validationErrs.Add("client_hash", "is required")
	} else if len(clientHash) != 64 {
		validationErrs.Add("client_hash", "must be a valid SHA-256 hash")
	} else if _, err := hex.DecodeString(clientHash); err != nil {
		validationErrs.Add("client_hash", "must be a valid SHA-256 hash")
	}

	if validationErrs.HasErrors() {
		h.cleanupFile(filePath)
		return nil, "", validationErrs
	}

	// Verify hash
	if clientHash != fileHash {
		h.cleanupFile(filePath)
		return nil, "", domain.NewDomainError(domain.ErrValidation, "HASH_MISMATCH", "client hash does not match uploaded signature")
	}

	input := &service.CaptureSignatureInput{
		SignerName:     signerName,
		SignatureImage: filePath,
		SignatureHash:  fileHash,
		SigningMethod:  domain.SigningMethod(signingMethod),
	}

	if signerEmail != "" {
		input.SignerEmail = &signerEmail
	}
	if deviceID != "" {
		input.DeviceID = &deviceID
	}

	return input, filePath, nil
}

// parseSignatureUploadForTicket parses the multipart signature upload for authenticated users.
func (h *SignatureHandler) parseSignatureUploadForTicket(r *http.Request, companyID, ticketID uuid.UUID) (*service.CaptureSignatureInput, string, error) {
	maxBytes := signatureMaxFileSize + signatureMetadataMaxBytes
	r.Body = http.MaxBytesReader(nil, r.Body, maxBytes)

	reader, err := r.MultipartReader()
	if err != nil {
		return nil, "", domain.NewDomainError(domain.ErrValidation, "INVALID_MULTIPART", "invalid multipart form data")
	}

	var (
		signerName      string
		signerEmail     string
		signingMethod   string
		deviceID        string
		clientHashRaw   string

		filePath      string
		fileHash      string
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

		if fieldName != "signature" {
			value, err := readSignatureField(part)
			_ = part.Close()
			if err != nil {
				parseErr = domain.NewDomainError(domain.ErrValidation, "INVALID_FIELD", err.Error())
				break
			}

			switch fieldName {
			case "signer_name":
				signerName = value
			case "signer_email":
				signerEmail = value
			case "signing_method":
				signingMethod = value
			case "device_id":
				deviceID = value
			case "client_hash":
				clientHashRaw = value
			}
			continue
		}

		if filePath != "" {
			_ = part.Close()
			parseErr = domain.NewDomainError(domain.ErrValidation, "INVALID_FILE", "multiple signature files provided")
			break
		}

		filePath, fileHash, err = h.storeSignatureFile(part, companyID.String(), ticketID)
		_ = part.Close()
		if err != nil {
			parseErr = err
			break
		}
	}

	if parseErr != nil {
		h.cleanupFile(filePath)
		return nil, "", parseErr
	}

	// Validate required fields
	validationErrs := domain.ValidationErrors{}
	if filePath == "" {
		validationErrs.Add("signature", "is required")
	}
	if signerName == "" {
		validationErrs.Add("signer_name", "is required")
	}
	if signingMethod == "" {
		validationErrs.Add("signing_method", "is required")
	}

	clientHash := strings.ToLower(strings.TrimSpace(clientHashRaw))
	if clientHash == "" {
		validationErrs.Add("client_hash", "is required")
	} else if len(clientHash) != 64 {
		validationErrs.Add("client_hash", "must be a valid SHA-256 hash")
	} else if _, err := hex.DecodeString(clientHash); err != nil {
		validationErrs.Add("client_hash", "must be a valid SHA-256 hash")
	}

	if validationErrs.HasErrors() {
		h.cleanupFile(filePath)
		return nil, "", validationErrs
	}

	// Verify hash
	if clientHash != fileHash {
		h.cleanupFile(filePath)
		return nil, "", domain.NewDomainError(domain.ErrValidation, "HASH_MISMATCH", "client hash does not match uploaded signature")
	}

	input := &service.CaptureSignatureInput{
		SignerName:     signerName,
		SignatureImage: filePath,
		SignatureHash:  fileHash,
		SigningMethod:  domain.SigningMethod(signingMethod),
	}

	if signerEmail != "" {
		input.SignerEmail = &signerEmail
	}
	if deviceID != "" {
		input.DeviceID = &deviceID
	}

	return input, filePath, nil
}

// storeSignatureFile stores the signature image file and returns the relative path and hash.
func (h *SignatureHandler) storeSignatureFile(part *multipart.Part, companyOrPending string, ticketID uuid.UUID) (string, string, error) {
	reader := bufio.NewReader(part)
	sniff, _ := reader.Peek(512)
	if len(sniff) == 0 {
		return "", "", domain.NewDomainError(domain.ErrValidation, "INVALID_FILE", "signature file is empty")
	}

	// Signature images should be PNG
	mimeType := http.DetectContentType(sniff)
	if mimeType != "image/png" {
		return "", "", domain.NewDomainError(domain.ErrValidation, "INVALID_FILE_TYPE", "signature must be a PNG image")
	}

	filename := uuid.New().String() + ".png"
	relativePath := filepath.Join(companyOrPending, "signatures", ticketID.String(), filename)
	absolutePath := filepath.Join(h.storageBase, relativePath)

	if err := os.MkdirAll(filepath.Dir(absolutePath), 0755); err != nil {
		return "", "", fmt.Errorf("create signature directory: %w", err)
	}

	dest, err := os.Create(absolutePath)
	if err != nil {
		return "", "", fmt.Errorf("create signature file: %w", err)
	}
	defer dest.Close()

	hash := sha256.New()
	writer := io.MultiWriter(dest, hash)
	written, err := io.Copy(writer, reader)
	if err != nil {
		_ = os.Remove(absolutePath)
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			return "", "", domain.NewDomainError(domain.ErrValidation, "FILE_TOO_LARGE", "signature file exceeds max size")
		}
		return "", "", fmt.Errorf("write signature file: %w", err)
	}

	if written <= 0 {
		_ = os.Remove(absolutePath)
		return "", "", domain.NewDomainError(domain.ErrValidation, "INVALID_FILE", "signature file is empty")
	}

	if written > signatureMaxFileSize {
		_ = os.Remove(absolutePath)
		return "", "", domain.NewDomainError(domain.ErrValidation, "FILE_TOO_LARGE", "signature file exceeds max size")
	}

	fileHash := hex.EncodeToString(hash.Sum(nil))
	return filepath.ToSlash(relativePath), fileHash, nil
}

func (h *SignatureHandler) cleanupFile(relativePath string) error {
	if strings.TrimSpace(relativePath) == "" || strings.TrimSpace(h.storageBase) == "" {
		return nil
	}
	absolutePath := filepath.Join(h.storageBase, filepath.FromSlash(relativePath))
	if err := os.Remove(absolutePath); err != nil && !errors.Is(err, os.ErrNotExist) {
		slog.Default().Warn("failed to cleanup signature file", slog.String("error", err.Error()))
		return err
	}
	return nil
}

func readSignatureField(part *multipart.Part) (string, error) {
	value, err := io.ReadAll(io.LimitReader(part, signatureFieldMaxBytes+1))
	if err != nil {
		return "", err
	}
	if int64(len(value)) > signatureFieldMaxBytes {
		return "", errors.New("field value too large")
	}
	return strings.TrimSpace(string(value)), nil
}

func parseClientIP(r *http.Request) *net.IP {
	// Try X-Forwarded-For first
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		// Take the first IP in the chain
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			ipStr := strings.TrimSpace(parts[0])
			ip := net.ParseIP(ipStr)
			if ip != nil {
				return &ip
			}
		}
	}

	// Try X-Real-IP
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		ip := net.ParseIP(strings.TrimSpace(xri))
		if ip != nil {
			return &ip
		}
	}

	// Fall back to RemoteAddr
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		ip := net.ParseIP(host)
		if ip != nil {
			return &ip
		}
	}

	return nil
}
