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
	"github.com/gridify/backend/internal/domain"
	transporthttp "github.com/gridify/backend/internal/transport/http"
	"github.com/gridify/backend/internal/transport/http/httputil"
	"github.com/gridify/backend/internal/transport/http/middleware"
)

func createTestUser(t *testing.T, router http.Handler, usernamePrefix string) (string, *http.Cookie) {
	uniqueID := uuid.New().String()[:8]
	email := fmt.Sprintf("%s_%s@gridify.test", usernamePrefix, uniqueID)
	username := fmt.Sprintf("%s_%s", usernamePrefix, uniqueID)
	password := "Password123!"

	regBody, _ := json.Marshal(map[string]interface{}{
		"email":      email,
		"username":   username,
		"password":   password,
		"first_name": "Test",
		"last_name":  "User",
	})
	regReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(regBody))
	regReq.Header.Set("Content-Type", "application/json")
	regRec := httptest.NewRecorder()
	router.ServeHTTP(regRec, regReq)

	loginBody, _ := json.Marshal(map[string]string{
		"identifier": email,
		"password":   password,
	})
	loginReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(loginBody))
	loginReq.Header.Set("Content-Type", "application/json")
	loginRec := httptest.NewRecorder()
	router.ServeHTTP(loginRec, loginReq)

	for _, c := range loginRec.Result().Cookies() {
		if c.Name == middleware.SessionCookieName {
			return email, c
		}
	}
	t.Fatal("failed to get session cookie")
	return "", nil
}

func TestTournamentCRUDAndOwnershipGuard(t *testing.T) {
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

	// Get a valid game ID
	var gameID string
	err = pool.QueryRow(ctx, "SELECT id FROM games LIMIT 1").Scan(&gameID)
	if err != nil {
		t.Fatalf("failed to query a game: %v", err)
	}

	// 1. Create User A (Owner) and User B (Attacker)
	_, userACookie := createTestUser(t, router, "ownerA")
	_, userBCookie := createTestUser(t, router, "attackerB")

	var createdTourney domain.Tournament

	// 2. User A creates a HYBRID_UCL tournament (should create 2 stages atomically)
	t.Run("User A creates HYBRID_UCL tournament with 2 stages", func(t *testing.T) {
		body, _ := json.Marshal(map[string]interface{}{
			"game_id":        gameID,
			"title":          "Champions League 2026",
			"description":    "Hybrid UCL format tournament",
			"format_type":    "HYBRID_UCL",
			"best_of":        3,
			"home_away":      true,
			"points_per_win": 3,
		})

		req := httptest.NewRequest(http.MethodPost, "/api/v1/tournaments", bytes.NewReader(body))
		req.AddCookie(userACookie)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("expected status 201, got %d: %s", rec.Code, rec.Body.String())
		}

		var resp struct {
			Success bool              `json:"success"`
			Data    domain.Tournament `json:"data"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to unmarshal JSON: %v", err)
		}

		createdTourney = resp.Data
		if len(createdTourney.Stages) != 2 {
			t.Errorf("expected 2 stages for HYBRID_UCL, got %d", len(createdTourney.Stages))
		}
	})

	// 3. User B attempts to UPDATE User A's tournament -> 403 Forbidden
	t.Run("User B (non-owner) update tournament returns 403 Forbidden", func(t *testing.T) {
		body, _ := json.Marshal(map[string]interface{}{
			"title": "Hacked Title",
		})

		path := fmt.Sprintf("/api/v1/tournaments/%s", createdTourney.ID)
		req := httptest.NewRequest(http.MethodPatch, path, bytes.NewReader(body))
		req.AddCookie(userBCookie)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Fatalf("expected status 403 Forbidden, got %d: %s", rec.Code, rec.Body.String())
		}
	})

	// 4. User B attempts to DELETE User A's tournament -> 403 Forbidden
	t.Run("User B (non-owner) delete tournament returns 403 Forbidden", func(t *testing.T) {
		path := fmt.Sprintf("/api/v1/tournaments/%s", createdTourney.ID)
		req := httptest.NewRequest(http.MethodDelete, path, nil)
		req.AddCookie(userBCookie)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Fatalf("expected status 403 Forbidden, got %d: %s", rec.Code, rec.Body.String())
		}
	})

	// 5. User A updates tournament -> 200 OK
	t.Run("User A (owner) updates tournament successfully", func(t *testing.T) {
		body, _ := json.Marshal(map[string]interface{}{
			"title": "Champions League 2026 - Official",
		})

		path := fmt.Sprintf("/api/v1/tournaments/%s", createdTourney.ID)
		req := httptest.NewRequest(http.MethodPatch, path, bytes.NewReader(body))
		req.AddCookie(userACookie)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200 OK, got %d: %s", rec.Code, rec.Body.String())
		}
	})

	// 6. User A lists tournaments -> 200 OK with total count >= 1
	t.Run("User A lists their tournaments", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/tournaments", nil)
		req.AddCookie(userACookie)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}

		var resp httputil.Envelope
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to unmarshal JSON: %v", err)
		}

		list, ok := resp.Data.([]interface{})
		if !ok || len(list) < 1 {
			t.Errorf("expected at least 1 tournament, got %d", len(list))
		}
	})

	// 7. User A deletes tournament -> 200 OK
	t.Run("User A (owner) deletes tournament successfully", func(t *testing.T) {
		path := fmt.Sprintf("/api/v1/tournaments/%s", createdTourney.ID)
		req := httptest.NewRequest(http.MethodDelete, path, nil)
		req.AddCookie(userACookie)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200 OK, got %d: %s", rec.Code, rec.Body.String())
		}

		// Verify deletion
		var count int
		_ = pool.QueryRow(ctx, "SELECT COUNT(*) FROM tournaments WHERE id = $1", createdTourney.ID).Scan(&count)
		if count != 0 {
			t.Errorf("expected tournament to be deleted from database, found %d", count)
		}
	})
}
