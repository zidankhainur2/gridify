package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/gridify/backend/internal/config"
	"github.com/gridify/backend/internal/database"
	transporthttp "github.com/gridify/backend/internal/transport/http"
	"github.com/gridify/backend/internal/transport/http/httputil"
	"github.com/gridify/backend/internal/transport/http/middleware"
)

func TestMasterGamesAndDashboardSummary(t *testing.T) {
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

	// 1. GET /api/v1/games (Public)
	t.Run("GET /api/v1/games returns master games list", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/games", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
		}

		var resp httputil.Envelope
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to parse JSON: %v", err)
		}

		gamesList, ok := resp.Data.([]interface{})
		if !ok || len(gamesList) < 10 {
			t.Fatalf("expected at least 10 games, got %d", len(gamesList))
		}
	})

	// 2. Register & Login a user to test dashboard
	uniqueID := uuid.New().String()[:8]
	testEmail := fmt.Sprintf("dash_%s@gridify.test", uniqueID)
	testUsername := fmt.Sprintf("dash_%s", uniqueID)
	testPassword := "Password123!"

	regBody, _ := json.Marshal(map[string]interface{}{
		"email":      testEmail,
		"username":   testUsername,
		"password":   testPassword,
		"first_name": "Dashboard",
		"last_name":  "Tester",
	})
	regReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(regBody))
	regReq.Header.Set("Content-Type", "application/json")
	regRec := httptest.NewRecorder()
	router.ServeHTTP(regRec, regReq)

	loginBody, _ := json.Marshal(map[string]string{
		"identifier": testEmail,
		"password":   testPassword,
	})
	loginReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(loginBody))
	loginReq.Header.Set("Content-Type", "application/json")
	loginRec := httptest.NewRecorder()
	router.ServeHTTP(loginRec, loginReq)

	var sessionCookie *http.Cookie
	for _, c := range loginRec.Result().Cookies() {
		if c.Name == middleware.SessionCookieName {
			sessionCookie = c
			break
		}
	}
	if sessionCookie == nil {
		t.Fatal("failed to get session cookie")
	}

	// 3. GET /api/v1/dashboard/summary without session -> 401
	t.Run("GET /api/v1/dashboard/summary without session returns 401", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/dashboard/summary", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected status 401, got %d", rec.Code)
		}
	})

	// 4. GET /api/v1/dashboard/summary with session -> 200 OK
	t.Run("GET /api/v1/dashboard/summary with session returns valid metrics", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/dashboard/summary", nil)
		req.AddCookie(sessionCookie)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
		}

		var resp httputil.Envelope
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to parse JSON: %v", err)
		}

		dataMap := resp.Data.(map[string]interface{})
		if _, ok := dataMap["total_tournaments"]; !ok {
			t.Errorf("missing total_tournaments metric")
		}
		if _, ok := dataMap["active_tournaments"]; !ok {
			t.Errorf("missing active_tournaments metric")
		}
		if _, ok := dataMap["upcoming_matches_count"]; !ok {
			t.Errorf("missing upcoming_matches_count metric")
		}
	})
}
