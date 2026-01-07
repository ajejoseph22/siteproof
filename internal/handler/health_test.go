package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

type dbPingerStub struct {
	err error
}

func (s dbPingerStub) Ping(ctx context.Context) error {
	return s.err
}

func TestHealthHandlerHealthOK(t *testing.T) {
	storageDir := createTempDir(t)
	handler := NewHealthHandler(dbPingerStub{}, storageDir, "test-version")

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	recorder := httptest.NewRecorder()

	handler.Health(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}

	var response struct {
		Data HealthResponse `json:"data"`
	}
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Data.Status != "ok" {
		t.Fatalf("expected status ok, got %s", response.Data.Status)
	}
	if response.Data.Database != "ok" {
		t.Fatalf("expected database ok, got %s", response.Data.Database)
	}
	if response.Data.Storage != "ok" {
		t.Fatalf("expected storage ok, got %s", response.Data.Storage)
	}
	if response.Data.Version != "test-version" {
		t.Fatalf("expected version test-version, got %s", response.Data.Version)
	}
	if response.Data.Uptime == "" {
		t.Fatalf("expected uptime to be set")
	}
}

func TestHealthHandlerHealthDegradedOnStorageFailure(t *testing.T) {
	storageDir := createTempDir(t)
	storageFile := createTempFile(t, storageDir)
	handler := NewHealthHandler(dbPingerStub{}, storageFile, "test-version")

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	recorder := httptest.NewRecorder()

	handler.Health(recorder, req)

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status %d, got %d", http.StatusServiceUnavailable, recorder.Code)
	}

	var response struct {
		Data HealthResponse `json:"data"`
	}
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Data.Status != "degraded" {
		t.Fatalf("expected status degraded, got %s", response.Data.Status)
	}
	if response.Data.Storage != "error" {
		t.Fatalf("expected storage error, got %s", response.Data.Storage)
	}
}

func TestHealthHandlerReadyReturnsErrors(t *testing.T) {
	storageDir := createTempDir(t)
	handler := NewHealthHandler(dbPingerStub{err: errors.New("db down")}, storageDir, "test-version")

	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	recorder := httptest.NewRecorder()

	handler.Ready(recorder, req)

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status %d, got %d", http.StatusServiceUnavailable, recorder.Code)
	}
}

func TestHealthHandlerReadyOK(t *testing.T) {
	storageDir := createTempDir(t)
	handler := NewHealthHandler(dbPingerStub{}, storageDir, "test-version")

	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	recorder := httptest.NewRecorder()

	handler.Ready(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}

	var response struct {
		Data map[string]string `json:"data"`
	}
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Data["status"] != "ready" {
		t.Fatalf("expected status ready, got %s", response.Data["status"])
	}
}

func createTempDir(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp(".", "health-storage-")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		_ = os.RemoveAll(dir)
	})
	return dir
}

func createTempFile(t *testing.T, dir string) string {
	t.Helper()
	file, err := os.CreateTemp(dir, "health-file-")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Remove(file.Name())
	})
	if err := file.Close(); err != nil {
		t.Fatalf("failed to close temp file: %v", err)
	}
	return file.Name()
}
