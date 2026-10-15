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

	"github.com/gridify/backend/internal/config"
	"github.com/gridify/backend/internal/database"
	"github.com/gridify/backend/internal/domain"
	transporthttp "github.com/gridify/backend/internal/transport/http"
	"github.com/gridify/backend/internal/transport/http/httputil"
)

func TestParticipantCRUDAndLockingRules(t *testing.T) {
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

	// Get a game ID
	var gameID string
	err = pool.QueryRow(ctx, "SELECT id FROM games LIMIT 1").Scan(&gameID)
	if err != nil {
		t.Fatalf("failed to query a game: %v", err)
	}

	// Create User A (owner) & User B (attacker)
	_, userACookie := createTestUser(t, router, "ownerPart")
	_, userBCookie := createTestUser(t, router, "attPart")

	// Create a Tournament for User A
	tourneyBody, _ := json.Marshal(map[string]interface{}{
		"game_id":     gameID,
		"title":       "Participant Test Tournament",
		"format_type": "SINGLE_ELIMINATION",
	})
	tourneyReq := httptest.NewRequest(http.MethodPost, "/api/v1/tournaments", bytes.NewReader(tourneyBody))
	tourneyReq.AddCookie(userACookie)
	tourneyReq.Header.Set("Content-Type", "application/json")
	tourneyRec := httptest.NewRecorder()
	router.ServeHTTP(tourneyRec, tourneyReq)

	var tourneyResp struct {
		Data domain.Tournament `json:"data"`
	}
	_ = json.Unmarshal(tourneyRec.Body.Bytes(), &tourneyResp)
	tourneyID := tourneyResp.Data.ID

	var createdParticipant domain.Participant

	// 1. User A adds Participant 1
	t.Run("User A adds participant in DRAFT status", func(t *testing.T) {
		body, _ := json.Marshal(map[string]interface{}{
			"name":         "Team Secret",
			"contact_info": "secret@team.com",
		})
		path := fmt.Sprintf("/api/v1/tournaments/%s/participants", tourneyID)
		req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(body))
		req.AddCookie(userACookie)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("expected status 201, got %d: %s", rec.Code, rec.Body.String())
		}

		var resp struct {
			Data domain.Participant `json:"data"`
		}
		_ = json.Unmarshal(rec.Body.Bytes(), &resp)
		createdParticipant = resp.Data

		if createdParticipant.Name != "Team Secret" || createdParticipant.SeedNumber != 1 {
			t.Errorf("unexpected participant data: %+v", createdParticipant)
		}
	})

	// 2. User B tries to add participant to User A's tournament -> 403 Forbidden
	t.Run("User B (non-owner) add participant returns 403 Forbidden", func(t *testing.T) {
		body, _ := json.Marshal(map[string]interface{}{
			"name": "Intruder Team",
		})
		path := fmt.Sprintf("/api/v1/tournaments/%s/participants", tourneyID)
		req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(body))
		req.AddCookie(userBCookie)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Fatalf("expected status 403, got %d: %s", rec.Code, rec.Body.String())
		}
	})

	// 3. User B tries to update participant -> 403 Forbidden
	t.Run("User B (non-owner) update participant returns 403 Forbidden", func(t *testing.T) {
		body, _ := json.Marshal(map[string]interface{}{
			"name": "Hacked Team",
		})
		path := fmt.Sprintf("/api/v1/tournaments/%s/participants/%s", tourneyID, createdParticipant.ID)
		req := httptest.NewRequest(http.MethodPatch, path, bytes.NewReader(body))
		req.AddCookie(userBCookie)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Fatalf("expected status 403, got %d: %s", rec.Code, rec.Body.String())
		}
	})

	// 4. User A updates participant info
	t.Run("User A updates participant name and contact", func(t *testing.T) {
		body, _ := json.Marshal(map[string]interface{}{
			"name":         "Team Secret Official",
			"contact_info": "contact@secret.gg",
		})
		path := fmt.Sprintf("/api/v1/tournaments/%s/participants/%s", tourneyID, createdParticipant.ID)
		req := httptest.NewRequest(http.MethodPatch, path, bytes.NewReader(body))
		req.AddCookie(userACookie)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
		}
	})

	// 5. Change Tournament Status to ONGOING
	_, err = pool.Exec(ctx, "UPDATE tournaments SET status = 'ONGOING' WHERE id = $1", tourneyID)
	if err != nil {
		t.Fatalf("failed to update tournament status: %v", err)
	}

	// 6. User A tries to add participant while ONGOING -> 409 Conflict (TOURNAMENT_LOCKED)
	t.Run("Adding participant while ONGOING returns 409 Conflict", func(t *testing.T) {
		body, _ := json.Marshal(map[string]interface{}{
			"name": "Late Team",
		})
		path := fmt.Sprintf("/api/v1/tournaments/%s/participants", tourneyID)
		req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(body))
		req.AddCookie(userACookie)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusConflict {
			t.Fatalf("expected status 409 Conflict, got %d: %s", rec.Code, rec.Body.String())
		}
	})

	// 7. User A tries to delete participant while ONGOING -> 409 Conflict (TOURNAMENT_LOCKED)
	t.Run("Deleting participant while ONGOING returns 409 Conflict", func(t *testing.T) {
		path := fmt.Sprintf("/api/v1/tournaments/%s/participants/%s", tourneyID, createdParticipant.ID)
		req := httptest.NewRequest(http.MethodDelete, path, nil)
		req.AddCookie(userACookie)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusConflict {
			t.Fatalf("expected status 409 Conflict, got %d: %s", rec.Code, rec.Body.String())
		}
	})

	// 8. User A CAN still edit participant name while ONGOING (for display correction)
	t.Run("Editing participant name while ONGOING is permitted", func(t *testing.T) {
		body, _ := json.Marshal(map[string]interface{}{
			"name": "Team Secret [Corrected]",
		})
		path := fmt.Sprintf("/api/v1/tournaments/%s/participants/%s", tourneyID, createdParticipant.ID)
		req := httptest.NewRequest(http.MethodPatch, path, bytes.NewReader(body))
		req.AddCookie(userACookie)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200 OK, got %d: %s", rec.Code, rec.Body.String())
		}
	})

	// 9. Revert status to DRAFT and delete participant -> 200 OK
	_, err = pool.Exec(ctx, "UPDATE tournaments SET status = 'DRAFT' WHERE id = $1", tourneyID)
	if err != nil {
		t.Fatalf("failed to update tournament status: %v", err)
	}

	t.Run("Deleting participant in DRAFT status succeeds", func(t *testing.T) {
		path := fmt.Sprintf("/api/v1/tournaments/%s/participants/%s", tourneyID, createdParticipant.ID)
		req := httptest.NewRequest(http.MethodDelete, path, nil)
		req.AddCookie(userACookie)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200 OK, got %d: %s", rec.Code, rec.Body.String())
		}
	})

	// 10. Public spectator can list participants (no auth required)
	t.Run("Spectator lists participants without authentication", func(t *testing.T) {
		path := fmt.Sprintf("/api/v1/tournaments/%s/participants", tourneyID)
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200 OK, got %d: %s", rec.Code, rec.Body.String())
		}

		var resp httputil.Envelope
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to unmarshal JSON: %v", err)
		}
	})
}
