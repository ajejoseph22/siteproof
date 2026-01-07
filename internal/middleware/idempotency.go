package middleware

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/ajejoseph22/siteproof/internal/repository/db"
)

const (
	idempotencyHeader          = "Idempotency-Key"
	idempotentReplayedHeader   = "X-Idempotent-Replayed"
	idempotencyResponseDefault = "null"
	idempotencyTTL             = 24 * time.Hour
)

type idempotencyResponseWriter struct {
	http.ResponseWriter
	status int
	body   bytes.Buffer
}

func (w *idempotencyResponseWriter) WriteHeader(code int) {
	if w.status == 0 {
		w.status = code
	}
	w.ResponseWriter.WriteHeader(code)
}

func (w *idempotencyResponseWriter) Write(data []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	w.body.Write(data)
	return w.ResponseWriter.Write(data)
}

// Idempotency caches successful POST responses for safe request retries.
func Idempotency(queries *db.Queries) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				next.ServeHTTP(w, r)
				return
			}

			rawKey := strings.TrimSpace(r.Header.Get(idempotencyHeader))
			if rawKey == "" {
				next.ServeHTTP(w, r)
				return
			}

			keyHash := hashIdempotencyKey(rawKey, r.Method, r.URL.Path)
			if existing, err := queries.GetIdempotencyKey(r.Context(), keyHash); err == nil {
				writeIdempotentReplay(w, existing)
				return
			} else if err != nil && err != pgx.ErrNoRows {
				slog.Default().Warn("idempotency lookup failed", slog.String("error", err.Error()))
			}

			capture := &idempotencyResponseWriter{ResponseWriter: w}
			next.ServeHTTP(capture, r)

			status := capture.status
			if status == 0 {
				status = http.StatusOK
			}

			if status < http.StatusOK || status >= http.StatusMultipleChoices {
				return
			}

			if err := storeIdempotencyResponse(r, queries, keyHash, status, capture.body.Bytes()); err != nil {
				slog.Default().Warn("idempotency store failed", slog.String("error", err.Error()))
			}
		})
	}
}

func hashIdempotencyKey(key, method, path string) string {
	payload := key + ":" + method + ":" + path
	sum := sha256.Sum256([]byte(payload))
	return hex.EncodeToString(sum[:])
}

func writeIdempotentReplay(w http.ResponseWriter, entry *db.IdempotencyKey) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set(idempotentReplayedHeader, "true")
	w.WriteHeader(int(entry.ResponseStatus))
	if len(entry.ResponseBody) > 0 {
		_, _ = w.Write(entry.ResponseBody)
	}
}

func storeIdempotencyResponse(r *http.Request, queries *db.Queries, keyHash string, status int, body []byte) error {
	userID, ok := UserIDFromContext(r.Context())
	var userUUID pgtype.UUID
	if ok {
		userUUID = pgtype.UUID{Bytes: userID, Valid: true}
	}

	responseBody := json.RawMessage(body)
	if len(responseBody) == 0 {
		responseBody = json.RawMessage(idempotencyResponseDefault)
	}

	_, err := queries.CreateIdempotencyKey(r.Context(), db.CreateIdempotencyKeyParams{
		KeyHash:        keyHash,
		UserID:         userUUID,
		Endpoint:       r.URL.Path,
		Method:         r.Method,
		ResponseStatus: int32(status),
		ResponseBody:   responseBody,
		ExpiresAt:      time.Now().Add(idempotencyTTL),
	})
	if err == nil {
		return nil
	}

	var pgErr *pgconn.PgError
	if ok := errors.As(err, &pgErr); ok && pgErr.Code == "23505" {
		return nil
	}
	return err
}
