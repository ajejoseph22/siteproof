package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/ajejoseph22/siteproof/internal/domain"
	"github.com/ajejoseph22/siteproof/internal/middleware"
)

type fileServiceStub struct {
	getPhotoByID     func(ctx context.Context, companyID, photoID uuid.UUID) (*domain.Photo, error)
	getSignatureByID func(ctx context.Context, companyID, signatureID uuid.UUID) (*domain.Signature, error)
	getTicketByID    func(ctx context.Context, companyID, ticketID uuid.UUID) (*domain.Ticket, error)
}

func (s fileServiceStub) GetPhotoByID(ctx context.Context, companyID, photoID uuid.UUID) (*domain.Photo, error) {
	if s.getPhotoByID == nil {
		return nil, domain.ErrNotFound
	}
	return s.getPhotoByID(ctx, companyID, photoID)
}

func (s fileServiceStub) GetSignatureByID(ctx context.Context, companyID, signatureID uuid.UUID) (*domain.Signature, error) {
	if s.getSignatureByID == nil {
		return nil, domain.ErrNotFound
	}
	return s.getSignatureByID(ctx, companyID, signatureID)
}

func (s fileServiceStub) GetTicketByID(ctx context.Context, companyID, ticketID uuid.UUID) (*domain.Ticket, error) {
	if s.getTicketByID == nil {
		return nil, domain.ErrNotFound
	}
	return s.getTicketByID(ctx, companyID, ticketID)
}

func TestFileHandlerDownloadPhotoNotFound(t *testing.T) {
	handler := NewFileHandler(fileServiceStub{}, "/tmp/storage")

	photoID := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/files/photos/"+photoID.String(), nil)
	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add("photo_id", photoID.String())
	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx)
	ctx = middleware.WithUserContext(ctx, uuid.New(), uuid.New(), domain.RoleAdmin)
	req = req.WithContext(ctx)

	recorder := httptest.NewRecorder()
	handler.DownloadPhoto(recorder, req)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, recorder.Code)
	}
}

func TestFileHandlerDownloadPhotoCrossCompanyDenied(t *testing.T) {
	companyID := uuid.New()
	otherCompanyID := uuid.New()
	photoID := uuid.New()
	ticketID := uuid.New()

	handler := NewFileHandler(fileServiceStub{
		getPhotoByID: func(ctx context.Context, companyIDArg, photoIDArg uuid.UUID) (*domain.Photo, error) {
			// Service filters by company_id, so if a different company tries to access,
			// the query returns ErrNotFound (not the photo from another company)
			if companyIDArg != companyID {
				return nil, domain.ErrNotFound
			}
			return &domain.Photo{
				ID:       photoIDArg,
				TicketID: ticketID,
				FilePath: "test/photo.jpg",
			}, nil
		},
	}, "/tmp/storage")

	// Request as a user from otherCompanyID
	req := httptest.NewRequest(http.MethodGet, "/api/v1/files/photos/"+photoID.String(), nil)
	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add("photo_id", photoID.String())
	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx)
	ctx = middleware.WithUserContext(ctx, uuid.New(), otherCompanyID, domain.RoleAdmin)
	req = req.WithContext(ctx)

	recorder := httptest.NewRecorder()
	handler.DownloadPhoto(recorder, req)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected status %d for cross-company access, got %d", http.StatusNotFound, recorder.Code)
	}
}

func TestFileHandlerDownloadPhotoInvalidID(t *testing.T) {
	handler := NewFileHandler(fileServiceStub{}, "/tmp/storage")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/files/photos/invalid-uuid", nil)
	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add("photo_id", "invalid-uuid")
	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx)
	ctx = middleware.WithUserContext(ctx, uuid.New(), uuid.New(), domain.RoleAdmin)
	req = req.WithContext(ctx)

	recorder := httptest.NewRecorder()
	handler.DownloadPhoto(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
	}
}

func TestFileHandlerDownloadPhotoUnauthenticated(t *testing.T) {
	handler := NewFileHandler(fileServiceStub{}, "/tmp/storage")

	photoID := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/files/photos/"+photoID.String(), nil)
	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add("photo_id", photoID.String())
	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx)
	// No user context - simulates unauthenticated request
	req = req.WithContext(ctx)

	recorder := httptest.NewRecorder()
	handler.DownloadPhoto(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, recorder.Code)
	}
}

func TestFileHandlerDownloadPhotoSuccess(t *testing.T) {
	// Create a temporary directory and file for this test
	tmpDir, err := os.MkdirTemp("", "file_handler_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a test photo file
	testPhotoPath := filepath.Join(tmpDir, "test", "photo.jpg")
	if err := os.MkdirAll(filepath.Dir(testPhotoPath), 0755); err != nil {
		t.Fatalf("failed to create test dir: %v", err)
	}
	testContent := []byte("fake jpeg content")
	if err := os.WriteFile(testPhotoPath, testContent, 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	companyID := uuid.New()
	photoID := uuid.New()
	ticketID := uuid.New()

	handler := NewFileHandler(fileServiceStub{
		getPhotoByID: func(ctx context.Context, companyIDArg, photoIDArg uuid.UUID) (*domain.Photo, error) {
			return &domain.Photo{
				ID:        photoIDArg,
				TicketID:  ticketID,
				FilePath:  "test/photo.jpg",
				MimeType:  "image/jpeg",
				FileHash:  "abc123",
				CreatedAt: time.Now(),
			}, nil
		},
	}, tmpDir)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/files/photos/"+photoID.String(), nil)
	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add("photo_id", photoID.String())
	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx)
	ctx = middleware.WithUserContext(ctx, uuid.New(), companyID, domain.RoleAdmin)
	req = req.WithContext(ctx)

	recorder := httptest.NewRecorder()
	handler.DownloadPhoto(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}

	if recorder.Header().Get("Content-Type") != "image/jpeg" {
		t.Fatalf("expected Content-Type image/jpeg, got %s", recorder.Header().Get("Content-Type"))
	}

	if recorder.Header().Get("ETag") != `"abc123"` {
		t.Fatalf("expected ETag \"abc123\", got %s", recorder.Header().Get("ETag"))
	}

	// Photos use long-lived immutable caching (1 year)
	expectedCacheControl := "private, max-age=31536000, immutable"
	if recorder.Header().Get("Cache-Control") != expectedCacheControl {
		t.Fatalf("expected Cache-Control %q, got %s", expectedCacheControl, recorder.Header().Get("Cache-Control"))
	}
}

func TestFileHandlerDownloadPhotoConditionalRequest(t *testing.T) {
	// Create a temporary directory and file for this test
	tmpDir, err := os.MkdirTemp("", "file_handler_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a test photo file
	testPhotoPath := filepath.Join(tmpDir, "test", "photo.jpg")
	if err := os.MkdirAll(filepath.Dir(testPhotoPath), 0755); err != nil {
		t.Fatalf("failed to create test dir: %v", err)
	}
	if err := os.WriteFile(testPhotoPath, []byte("fake jpeg content"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	companyID := uuid.New()
	photoID := uuid.New()
	fileHash := "abc123def456"

	handler := NewFileHandler(fileServiceStub{
		getPhotoByID: func(ctx context.Context, companyIDArg, photoIDArg uuid.UUID) (*domain.Photo, error) {
			return &domain.Photo{
				ID:        photoIDArg,
				TicketID:  uuid.New(),
				FilePath:  "test/photo.jpg",
				MimeType:  "image/jpeg",
				FileHash:  fileHash,
				CreatedAt: time.Now(),
			}, nil
		},
	}, tmpDir)

	// Request with matching If-None-Match header
	req := httptest.NewRequest(http.MethodGet, "/api/v1/files/photos/"+photoID.String(), nil)
	req.Header.Set("If-None-Match", `"`+fileHash+`"`)
	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add("photo_id", photoID.String())
	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx)
	ctx = middleware.WithUserContext(ctx, uuid.New(), companyID, domain.RoleAdmin)
	req = req.WithContext(ctx)

	recorder := httptest.NewRecorder()
	handler.DownloadPhoto(recorder, req)

	if recorder.Code != http.StatusNotModified {
		t.Fatalf("expected status %d for conditional request, got %d", http.StatusNotModified, recorder.Code)
	}
}

func TestFileHandlerDownloadSignatureNotFound(t *testing.T) {
	handler := NewFileHandler(fileServiceStub{}, "/tmp/storage")

	signatureID := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/files/signatures/"+signatureID.String(), nil)
	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add("signature_id", signatureID.String())
	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx)
	ctx = middleware.WithUserContext(ctx, uuid.New(), uuid.New(), domain.RoleAdmin)
	req = req.WithContext(ctx)

	recorder := httptest.NewRecorder()
	handler.DownloadSignature(recorder, req)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, recorder.Code)
	}
}

func TestFileHandlerDownloadPDFNotGenerated(t *testing.T) {
	companyID := uuid.New()
	ticketID := uuid.New()

	handler := NewFileHandler(fileServiceStub{
		getTicketByID: func(ctx context.Context, companyIDArg, ticketIDArg uuid.UUID) (*domain.Ticket, error) {
			return &domain.Ticket{
				ID:        ticketIDArg,
				CompanyID: companyIDArg,
				PDFPath:   nil, // No PDF generated
			}, nil
		},
	}, "/tmp/storage")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/files/pdfs/"+ticketID.String(), nil)
	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add("ticket_id", ticketID.String())
	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx)
	ctx = middleware.WithUserContext(ctx, uuid.New(), companyID, domain.RoleAdmin)
	req = req.WithContext(ctx)

	recorder := httptest.NewRecorder()
	handler.DownloadPDF(recorder, req)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected status %d when PDF not generated, got %d", http.StatusNotFound, recorder.Code)
	}
}

func TestFileHandlerDownloadPDFCrossCompanyDenied(t *testing.T) {
	companyID := uuid.New()
	otherCompanyID := uuid.New()
	ticketID := uuid.New()
	pdfPath := "test/ticket.pdf"

	handler := NewFileHandler(fileServiceStub{
		getTicketByID: func(ctx context.Context, companyIDArg, ticketIDArg uuid.UUID) (*domain.Ticket, error) {
			// Service filters by company_id, so if a different company tries to access,
			// the query returns ErrNotFound
			if companyIDArg != companyID {
				return nil, domain.ErrNotFound
			}
			return &domain.Ticket{
				ID:        ticketIDArg,
				CompanyID: companyIDArg,
				PDFPath:   &pdfPath,
			}, nil
		},
	}, "/tmp/storage")

	// Request as a user from otherCompanyID
	req := httptest.NewRequest(http.MethodGet, "/api/v1/files/pdfs/"+ticketID.String(), nil)
	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add("ticket_id", ticketID.String())
	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx)
	ctx = middleware.WithUserContext(ctx, uuid.New(), otherCompanyID, domain.RoleAdmin)
	req = req.WithContext(ctx)

	recorder := httptest.NewRecorder()
	handler.DownloadPDF(recorder, req)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected status %d for cross-company PDF access, got %d", http.StatusNotFound, recorder.Code)
	}
}

func TestFileHandlerPathTraversalPrevented(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "file_handler_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	companyID := uuid.New()
	photoID := uuid.New()

	handler := NewFileHandler(fileServiceStub{
		getPhotoByID: func(ctx context.Context, companyIDArg, photoIDArg uuid.UUID) (*domain.Photo, error) {
			return &domain.Photo{
				ID:       photoIDArg,
				TicketID: uuid.New(),
				// Attempt path traversal
				FilePath: "../../../etc/passwd",
				MimeType: "text/plain",
			}, nil
		},
	}, tmpDir)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/files/photos/"+photoID.String(), nil)
	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add("photo_id", photoID.String())
	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx)
	ctx = middleware.WithUserContext(ctx, uuid.New(), companyID, domain.RoleAdmin)
	req = req.WithContext(ctx)

	recorder := httptest.NewRecorder()
	handler.DownloadPhoto(recorder, req)

	// Should return 404 due to path traversal protection
	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected status %d for path traversal attempt, got %d", http.StatusNotFound, recorder.Code)
	}
}

func TestFileHandlerDownloadSignatureSuccess(t *testing.T) {
	// Create a temporary directory and file for this test
	tmpDir, err := os.MkdirTemp("", "file_handler_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a test signature file
	testSignaturePath := filepath.Join(tmpDir, "test", "signature.png")
	if err := os.MkdirAll(filepath.Dir(testSignaturePath), 0755); err != nil {
		t.Fatalf("failed to create test dir: %v", err)
	}
	testContent := []byte("fake png content")
	if err := os.WriteFile(testSignaturePath, testContent, 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	companyID := uuid.New()
	signatureID := uuid.New()
	ticketID := uuid.New()

	handler := NewFileHandler(fileServiceStub{
		getSignatureByID: func(ctx context.Context, companyIDArg, signatureIDArg uuid.UUID) (*domain.Signature, error) {
			return &domain.Signature{
				ID:             signatureIDArg,
				TicketID:       ticketID,
				SignatureImage: "test/signature.png",
				SignatureHash:  "sig123",
				SignerType:     domain.SignerTypeForeman,
				SignedAt:       time.Now(),
			}, nil
		},
	}, tmpDir)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/files/signatures/"+signatureID.String(), nil)
	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add("signature_id", signatureID.String())
	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx)
	ctx = middleware.WithUserContext(ctx, uuid.New(), companyID, domain.RoleAdmin)
	req = req.WithContext(ctx)

	recorder := httptest.NewRecorder()
	handler.DownloadSignature(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}

	if recorder.Header().Get("Content-Type") != "image/png" {
		t.Fatalf("expected Content-Type image/png, got %s", recorder.Header().Get("Content-Type"))
	}

	if recorder.Header().Get("ETag") != `"sig123"` {
		t.Fatalf("expected ETag \"sig123\", got %s", recorder.Header().Get("ETag"))
	}

	// Signatures use standard caching (not immutable) since they're tied to ticket workflow
	expectedCacheControl := "private, max-age=3600"
	if recorder.Header().Get("Cache-Control") != expectedCacheControl {
		t.Fatalf("expected Cache-Control %q, got %s", expectedCacheControl, recorder.Header().Get("Cache-Control"))
	}
}

func TestFileHandlerDownloadPDFSuccess(t *testing.T) {
	// Create a temporary directory and file for this test
	tmpDir, err := os.MkdirTemp("", "file_handler_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a test PDF file
	testPDFPath := filepath.Join(tmpDir, "test", "ticket.pdf")
	if err := os.MkdirAll(filepath.Dir(testPDFPath), 0755); err != nil {
		t.Fatalf("failed to create test dir: %v", err)
	}
	testContent := []byte("%PDF-1.4 fake pdf content")
	if err := os.WriteFile(testPDFPath, testContent, 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	companyID := uuid.New()
	ticketID := uuid.New()
	pdfPath := "test/ticket.pdf"
	pdfHash := "pdf456"

	handler := NewFileHandler(fileServiceStub{
		getTicketByID: func(ctx context.Context, companyIDArg, ticketIDArg uuid.UUID) (*domain.Ticket, error) {
			return &domain.Ticket{
				ID:        ticketIDArg,
				CompanyID: companyIDArg,
				PDFPath:   &pdfPath,
				PDFHash:   &pdfHash,
			}, nil
		},
	}, tmpDir)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/files/pdfs/"+ticketID.String(), nil)
	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add("ticket_id", ticketID.String())
	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx)
	ctx = middleware.WithUserContext(ctx, uuid.New(), companyID, domain.RoleAdmin)
	req = req.WithContext(ctx)

	recorder := httptest.NewRecorder()
	handler.DownloadPDF(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}

	if recorder.Header().Get("Content-Type") != "application/pdf" {
		t.Fatalf("expected Content-Type application/pdf, got %s", recorder.Header().Get("Content-Type"))
	}

	if recorder.Header().Get("ETag") != `"pdf456"` {
		t.Fatalf("expected ETag \"pdf456\", got %s", recorder.Header().Get("ETag"))
	}

	// PDFs use standard caching (not immutable) since they can be regenerated
	expectedCacheControl := "private, max-age=3600"
	if recorder.Header().Get("Cache-Control") != expectedCacheControl {
		t.Fatalf("expected Cache-Control %q, got %s", expectedCacheControl, recorder.Header().Get("Cache-Control"))
	}

	// PDFs should have Content-Disposition attachment header
	contentDisposition := recorder.Header().Get("Content-Disposition")
	if contentDisposition == "" {
		t.Fatal("expected Content-Disposition header for PDF")
	}
	if !strings.Contains(contentDisposition, "attachment") {
		t.Fatalf("expected Content-Disposition to contain 'attachment', got %s", contentDisposition)
	}
}

func TestFileHandlerDownloadPhotoStorageNotConfigured(t *testing.T) {
	companyID := uuid.New()
	photoID := uuid.New()

	// Handler with empty storage base
	handler := NewFileHandler(fileServiceStub{
		getPhotoByID: func(ctx context.Context, companyIDArg, photoIDArg uuid.UUID) (*domain.Photo, error) {
			return &domain.Photo{
				ID:       photoIDArg,
				TicketID: uuid.New(),
				FilePath: "test/photo.jpg",
				MimeType: "image/jpeg",
			}, nil
		},
	}, "")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/files/photos/"+photoID.String(), nil)
	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add("photo_id", photoID.String())
	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx)
	ctx = middleware.WithUserContext(ctx, uuid.New(), companyID, domain.RoleAdmin)
	req = req.WithContext(ctx)

	recorder := httptest.NewRecorder()
	handler.DownloadPhoto(recorder, req)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d when storage not configured, got %d", http.StatusInternalServerError, recorder.Code)
	}
}
