package handler

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

// HealthHandler handles health check endpoints.
type HealthHandler struct {
	db              dbPinger
	storageBasePath string
	startTime       time.Time
	version         string
}

type dbPinger interface {
	Ping(ctx context.Context) error
}

const dependencyTimeout = 2 * time.Second

// NewHealthHandler creates a new health handler.
func NewHealthHandler(db dbPinger, storageBasePath, version string) *HealthHandler {
	return &HealthHandler{
		db:              db,
		storageBasePath: storageBasePath,
		startTime:       time.Now(),
		version:         version,
	}
}

// HealthResponse represents a health check response.
type HealthResponse struct {
	Status   string `json:"status"`
	Database string `json:"database"`
	Storage  string `json:"storage"`
	Version  string `json:"version"`
	Uptime   string `json:"uptime"`
}

// Health handles GET /health
func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	dbStatus := h.checkDatabaseStatus(r.Context())
	storageStatus := h.checkStorageStatus(r.Context())

	status := "ok"
	if dbStatus != "ok" || storageStatus != "ok" {
		status = "degraded"
	}

	uptime := time.Since(h.startTime).Round(time.Second).String()

	statusCode := http.StatusOK
	if status != "ok" {
		statusCode = http.StatusServiceUnavailable
	}

	WriteJSON(w, statusCode, HealthResponse{
		Status:   status,
		Database: dbStatus,
		Storage:  storageStatus,
		Version:  h.version,
		Uptime:   uptime,
	})
}

// Ready handles GET /ready
func (h *HealthHandler) Ready(w http.ResponseWriter, r *http.Request) {
	if err := h.checkDatabase(r.Context()); err != nil {
		WriteError(w, http.StatusServiceUnavailable, "NOT_READY", "database not available")
		return
	}

	if err := h.checkStorage(r.Context()); err != nil {
		WriteError(w, http.StatusServiceUnavailable, "NOT_READY", "storage not available")
		return
	}

	WriteJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

func (h *HealthHandler) checkDatabaseStatus(ctx context.Context) string {
	if err := h.checkDatabase(ctx); err != nil {
		return "error"
	}
	return "ok"
}

func (h *HealthHandler) checkStorageStatus(ctx context.Context) string {
	if err := h.checkStorage(ctx); err != nil {
		return "error"
	}
	return "ok"
}

func (h *HealthHandler) checkDatabase(ctx context.Context) error {
	if h.db == nil {
		return errors.New("database not configured")
	}

	dbCtx, cancel := context.WithTimeout(ctx, dependencyTimeout)
	defer cancel()

	return h.db.Ping(dbCtx)
}

func (h *HealthHandler) checkStorage(ctx context.Context) error {
	storageCtx, cancel := context.WithTimeout(ctx, dependencyTimeout)
	defer cancel()

	if err := storageCtx.Err(); err != nil {
		return err
	}

	if strings.TrimSpace(h.storageBasePath) == "" {
		return errors.New("storage base path not configured")
	}

	info, err := os.Stat(h.storageBasePath)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("storage base path is not a directory")
	}

	tmpFile, err := os.CreateTemp(h.storageBasePath, ".health-check-*")
	if err != nil {
		return err
	}
	if err := tmpFile.Close(); err != nil {
		return err
	}
	if err := os.Remove(tmpFile.Name()); err != nil {
		return err
	}

	return storageCtx.Err()
}
