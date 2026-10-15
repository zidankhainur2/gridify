package http_test

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gridify/backend/internal/config"
	"github.com/gridify/backend/internal/database"
	transporthttp "github.com/gridify/backend/internal/transport/http"
	"github.com/gridify/backend/internal/transport/http/httputil"
)

func TestHealthAndReadinessEndpoints(t *testing.T) {
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := database.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	router := transporthttp.NewRouter(cfg, pool, logger)

	// Test 1: GET /healthz
	t.Run("GET /healthz returns 200 OK", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}

		var resp httputil.Envelope
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to unmarshal JSON response: %v", err)
		}

		if !resp.Success {
			t.Errorf("expected success: true, got false")
		}

		dataMap, ok := resp.Data.(map[string]interface{})
		if !ok || dataMap["status"] != "ok" {
			t.Errorf("expected status 'ok', got %v", resp.Data)
		}
	})

	// Test 2: GET /readyz
	t.Run("GET /readyz returns 200 OK with database connected", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}

		var resp httputil.Envelope
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to unmarshal JSON response: %v", err)
		}

		if !resp.Success {
			t.Errorf("expected success: true, got false")
		}

		dataMap, ok := resp.Data.(map[string]interface{})
		if !ok || dataMap["status"] != "ok" || dataMap["database"] != "connected" {
			t.Errorf("expected status 'ok' and database 'connected', got %v", resp.Data)
		}
	})
}
