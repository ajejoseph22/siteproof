package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/ajejoseph22/siteproof/pkg/slogctx"
)

// responseWriter wraps http.ResponseWriter to capture the status code.
type responseWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func wrapResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w, status: http.StatusOK}
}

func (rw *responseWriter) WriteHeader(code int) {
	if !rw.wroteHeader {
		rw.status = code
		rw.wroteHeader = true
		rw.ResponseWriter.WriteHeader(code)
	}
}

// Logging creates a request logging middleware that attaches a context-aware
// logger with request_id and logs HTTP requests with user/company context.
func Logging(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Get request ID from context (set by RequestID middleware)
			requestID := RequestIDFromContext(r.Context())

			// Create a request-scoped logger with request_id
			reqLogger := logger.With(slog.String("request_id", requestID))

			// Attach the logger to the context for downstream use
			ctx := slogctx.WithLogger(r.Context(), reqLogger)
			r = r.WithContext(ctx)

			wrapped := wrapResponseWriter(w)
			next.ServeHTTP(wrapped, r)

			duration := time.Since(start)

			// Build log attributes
			attrs := []any{
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", wrapped.status),
				slog.Duration("duration", duration),
				slog.String("remote_addr", r.RemoteAddr),
			}

			// Add user_id and company_id if present (set by Auth middleware)
			if userID, ok := UserIDFromContext(r.Context()); ok {
				attrs = append(attrs, slog.String("user_id", userID.String()))
			}
			if companyID, ok := CompanyIDFromContext(r.Context()); ok {
				attrs = append(attrs, slog.String("company_id", companyID.String()))
			}

			// Log request completion
			reqLogger.Info("http_request", attrs...)
		})
	}
}
