package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// HealthHandler handles health check endpoints.
type HealthHandler struct {
	pool      *pgxpool.Pool
	startTime time.Time
	version   string
}

// NewHealthHandler creates a new health handler.
func NewHealthHandler(pool *pgxpool.Pool, version string) *HealthHandler {
	return &HealthHandler{
		pool:      pool,
		startTime: time.Now(),
		version:   version,
	}
}

// HealthResponse represents a health check response.
type HealthResponse struct {
	Status   string `json:"status"`
	Database string `json:"database"`
	Version  string `json:"version"`
	Uptime   string `json:"uptime"`
}

// Health handles GET /health
func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	dbStatus := "ok"
	if err := h.pool.Ping(ctx); err != nil {
		dbStatus = "error"
	}

	status := "ok"
	if dbStatus != "ok" {
		status = "degraded"
	}

	uptime := time.Since(h.startTime).Round(time.Second).String()

	w.Header().Set("Content-Type", "application/json")
	if status != "ok" {
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	WriteJSON(w, http.StatusOK, HealthResponse{
		Status:   status,
		Database: dbStatus,
		Version:  h.version,
		Uptime:   uptime,
	})
}

// Ready handles GET /ready
func (h *HealthHandler) Ready(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	if err := h.pool.Ping(ctx); err != nil {
		WriteError(w, http.StatusServiceUnavailable, "NOT_READY", "database not available")
		return
	}

	WriteJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

