package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/gridify/backend/internal/config"
	"github.com/gridify/backend/internal/database"
	"github.com/gridify/backend/internal/domain"
	transporthttp "github.com/gridify/backend/internal/transport/http"
)

// TestIntegration_SecurityAndAuthorizationMatrix memverifikasi pemenuhan TEST-002:
// - Akses anonim ke resource berotentikasi menghasilkan 401 Unauthorized
// - Percobaan manipulasi resource milik user lain menghasilkan 403 Forbidden
// - Percobaan manipulasi peserta turnamen saat ONGOING menghasilkan 409 Conflict
func TestIntegration_SecurityAndAuthorizationMatrix(t *testing.T) {
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := database.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	router := transporthttp.NewRouter(cfg, pool, logger)

	var gameID string
	err = pool.QueryRow(ctx, "SELECT id FROM games LIMIT 1").Scan(&gameID)
	if err != nil {
		t.Fatalf("failed to query a game: %v", err)
	}

	_, ownerCookie := createTestUser(t, router, "secOwner")
	_, attackerCookie := createTestUser(t, router, "secAttacker")

	// 1. Anonim accessing /api/v1/tournaments -> 401 Unauthorized
	t.Run("Anonymous request to protected endpoint returns 401", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/tournaments", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("expected status 401, got %d", rec.Code)
		}
	})

	// 2. Owner creates a tournament
	var tourney domain.Tournament
	var stageID string
	t.Run("Owner creates tournament", func(t *testing.T) {
		body, _ := json.Marshal(map[string]interface{}{
			"game_id":     gameID,
			"title":       "Sec Matrix Tournament",
			"format_type": "SINGLE_ELIMINATION",
		})
		req := httptest.NewRequest(http.MethodPost, "/api/v1/tournaments", bytes.NewReader(body))
		req.AddCookie(ownerCookie)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d", rec.Code)
		}

		var resp struct {
			Success bool              `json:"success"`
			Data    domain.Tournament `json:"data"`
		}
		_ = json.Unmarshal(rec.Body.Bytes(), &resp)
		tourney = resp.Data
		if len(tourney.Stages) > 0 {
			stageID = tourney.Stages[0].ID.String()
		}
	})

	// 3. Attacker attempts to update or delete tournament -> 403 Forbidden
	t.Run("Attacker updating owner's tournament returns 403 Forbidden", func(t *testing.T) {
		updateBody, _ := json.Marshal(map[string]interface{}{
			"title": "Hacked Tournament Title",
		})
		req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/v1/tournaments/%s", tourney.ID), bytes.NewReader(updateBody))
		req.AddCookie(attackerCookie)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Errorf("expected 403, got %d", rec.Code)
		}
	})

	t.Run("Attacker generating bracket on owner's stage returns 403 Forbidden", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tournaments/%s/stages/%s/generate", tourney.ID, stageID), nil)
		req.AddCookie(attackerCookie)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Errorf("expected 403, got %d", rec.Code)
		}
	})

	// 4. Owner adds participants
	var partA, partB domain.Participant
	t.Run("Owner adds 2 participants in DRAFT status", func(t *testing.T) {
		for _, name := range []string{"Tim Alpha", "Tim Beta"} {
			b, _ := json.Marshal(map[string]interface{}{"name": name})
			req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tournaments/%s/participants", tourney.ID), bytes.NewReader(b))
			req.AddCookie(ownerCookie)
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			if rec.Code != http.StatusCreated {
				t.Fatalf("expected 201, got %d", rec.Code)
			}
			var resp struct {
				Data domain.Participant `json:"data"`
			}
			_ = json.Unmarshal(rec.Body.Bytes(), &resp)
			if name == "Tim Alpha" {
				partA = resp.Data
			} else {
				partB = resp.Data
			}
		}
	})

	// 5. Attacker modifying participant -> 403
	t.Run("Attacker modifying participant returns 403 Forbidden", func(t *testing.T) {
		b, _ := json.Marshal(map[string]interface{}{"name": "Tampered Tim"})
		req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/v1/tournaments/%s/participants/%s", tourney.ID, partA.ID), bytes.NewReader(b))
		req.AddCookie(attackerCookie)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Errorf("expected 403, got %d", rec.Code)
		}
	})

	// 6. Generate bracket and set tournament status to ONGOING
	t.Run("Owner generates bracket and sets tournament status to ONGOING", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tournaments/%s/stages/%s/generate", tourney.ID, stageID), nil)
		req.AddCookie(ownerCookie)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
		}

		// Explicitly set tournament status to ONGOING via PATCH
		statusOngoing := "ONGOING"
		b, _ := json.Marshal(map[string]interface{}{"status": statusOngoing})
		patchReq := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/v1/tournaments/%s", tourney.ID), bytes.NewReader(b))
		patchReq.AddCookie(ownerCookie)
		patchReq.Header.Set("Content-Type", "application/json")
		patchRec := httptest.NewRecorder()
		router.ServeHTTP(patchRec, patchReq)

		if patchRec.Code != http.StatusOK {
			t.Fatalf("expected 200 for status update, got %d: %s", patchRec.Code, patchRec.Body.String())
		}
	})

	// 7. Adding / Deleting participant while ONGOING -> 409 Conflict
	t.Run("Adding participant while tournament is ONGOING returns 409 Conflict", func(t *testing.T) {
		b, _ := json.Marshal(map[string]interface{}{"name": "Tim Terlambat"})
		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tournaments/%s/participants", tourney.ID), bytes.NewReader(b))
		req.AddCookie(ownerCookie)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusConflict {
			t.Errorf("expected 409 Conflict when adding participant during ONGOING, got %d", rec.Code)
		}
	})

	t.Run("Deleting participant while tournament is ONGOING returns 409 Conflict", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/tournaments/%s/participants/%s", tourney.ID, partB.ID), nil)
		req.AddCookie(ownerCookie)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusConflict {
			t.Errorf("expected 409 Conflict when deleting participant during ONGOING, got %d", rec.Code)
		}
	})
}

// TestIntegration_DatabaseConstraintsAndTransactionRollback memverifikasi:
// - CHECK constraint chk_matches_distinct_participants (menolak match dengan participant_a == participant_b)
// - CHECK constraint chk_matches_valid_winner (menolak match dengan winner_id bukan peserta A maupun B)
// - Rollback transaksi database saat eksekusi WithTx gagal di tengah proses
func TestIntegration_DatabaseConstraintsAndTransactionRollback(t *testing.T) {
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := database.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	// 1. Create a dummy organizer and tournament directly
	var userID, tourneyID, stageID, part1ID, part2ID, strangerID uuid.UUID
	err = pool.QueryRow(ctx, `
		INSERT INTO users (email, username, password_hash, first_name, last_name)
		VALUES ($1, $2, 'dummyhash', 'Test', 'Constraint')
		RETURNING id
	`, fmt.Sprintf("const_%s@test.com", uuid.New().String()[:8]), fmt.Sprintf("u_%s", uuid.New().String()[:8])).Scan(&userID)
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	var gameID string
	_ = pool.QueryRow(ctx, "SELECT id FROM games LIMIT 1").Scan(&gameID)

	err = pool.QueryRow(ctx, `
		INSERT INTO tournaments (owner_id, game_id, title, format_type)
		VALUES ($1, $2, 'Constraint Test', 'SINGLE_ELIMINATION')
		RETURNING id
	`, userID, gameID).Scan(&tourneyID)
	if err != nil {
		t.Fatalf("failed to create tournament: %v", err)
	}

	err = pool.QueryRow(ctx, `
		INSERT INTO stages (tournament_id, name, type, sequence_order)
		VALUES ($1, 'Main Stage', 'SINGLE_ELIMINATION', 1)
		RETURNING id
	`, tourneyID).Scan(&stageID)
	if err != nil {
		t.Fatalf("failed to create stage: %v", err)
	}

	// Create participants
	_ = pool.QueryRow(ctx, "INSERT INTO participants (tournament_id, name) VALUES ($1, 'P1') RETURNING id", tourneyID).Scan(&part1ID)
	_ = pool.QueryRow(ctx, "INSERT INTO participants (tournament_id, name) VALUES ($1, 'P2') RETURNING id", tourneyID).Scan(&part2ID)
	_ = pool.QueryRow(ctx, "INSERT INTO participants (tournament_id, name) VALUES ($1, 'Stranger') RETURNING id", tourneyID).Scan(&strangerID)

	// 2. Test constraint: chk_matches_distinct_participants
	t.Run("Database rejects match where participant_a_id == participant_b_id", func(t *testing.T) {
		matchID := uuid.New()
		err := pool.QueryRow(ctx, `
			INSERT INTO matches (id, tournament_id, stage_id, bracket_type, round_number, match_number, participant_a_id, participant_b_id)
			VALUES ($1, $2, $3, 'UPPER', 1, 101, $4, $4)
			RETURNING id
		`, matchID, tourneyID, stageID, part1ID).Scan(&matchID)

		if err == nil {
			t.Errorf("expected CHECK constraint violation for identical participants, got nil")
		}
	})

	// 3. Test constraint: chk_matches_valid_winner
	t.Run("Database rejects winner_id that is neither participant A nor B", func(t *testing.T) {
		matchID := uuid.New()
		err := pool.QueryRow(ctx, `
			INSERT INTO matches (id, tournament_id, stage_id, bracket_type, round_number, match_number, participant_a_id, participant_b_id, winner_id)
			VALUES ($1, $2, $3, 'UPPER', 1, 102, $4, $5, $6)
			RETURNING id
		`, matchID, tourneyID, stageID, part1ID, part2ID, strangerID).Scan(&matchID)

		if err == nil {
			t.Errorf("expected CHECK constraint violation for invalid winner_id, got nil")
		}
	})

	// 4. Test Transaction Rollback with database.WithTx
	t.Run("Transaction Rollback restores database state on error", func(t *testing.T) {
		matchIDToRollback := uuid.New()
		expectedErr := errors.New("simulated error to trigger rollback")

		txErr := database.WithTx(ctx, pool, func(tx pgx.Tx) error {
			// Step 1: Insert a valid match
			_, err := tx.Exec(ctx, `
				INSERT INTO matches (id, tournament_id, stage_id, bracket_type, round_number, match_number, participant_a_id, participant_b_id)
				VALUES ($1, $2, $3, 'UPPER', 1, 999, $4, $5)
			`, matchIDToRollback, tourneyID, stageID, part1ID, part2ID)
			if err != nil {
				return err
			}

			// Step 2: Trigger error
			return expectedErr
		})

		if !errors.Is(txErr, expectedErr) {
			t.Fatalf("expected txErr to be simulated error, got %v", txErr)
		}

		// Verify that matchIDToRollback was NOT persisted
		var exists bool
		err := pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM matches WHERE id = $1)", matchIDToRollback).Scan(&exists)
		if err != nil {
			t.Fatalf("failed to check match existence: %v", err)
		}
		if exists {
			t.Errorf("expected match to be rolled back, but it exists in database!")
		}
	})
}
