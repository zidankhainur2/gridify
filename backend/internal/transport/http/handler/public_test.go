package handler_test

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/gridify/backend/internal/config"
	"github.com/gridify/backend/internal/database"
	"github.com/gridify/backend/internal/domain"
	transporthttp "github.com/gridify/backend/internal/transport/http"
)

func TestPublicHandler_GetTournamentAndSSEStream(t *testing.T) {
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	ctx := context.Background()
	pool, err := database.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		t.Fatalf("failed to connect to postgres: %v", err)
	}
	defer pool.Close()

	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	router := transporthttp.NewRouter(cfg, pool, logger)

	// Create an organizer
	_, cookie := createTestUser(t, router, "public_test_org")

	// Get a master game
	reqGame := httptest.NewRequest(http.MethodGet, "/api/v1/games", nil)
	recGame := httptest.NewRecorder()
	router.ServeHTTP(recGame, reqGame)
	var gameResp struct {
		Data []domain.Game `json:"data"`
	}
	_ = json.Unmarshal(recGame.Body.Bytes(), &gameResp)
	if len(gameResp.Data) == 0 {
		t.Fatal("no games found")
	}
	gameID := gameResp.Data[0].ID.String()

	// Create tournament
	createTourPayload := map[string]interface{}{
		"game_id":     gameID,
		"title":       "Public Spectator Cup 2026",
		"format_type": "SINGLE_ELIMINATION",
		"best_of":     1,
	}
	body, _ := json.Marshal(createTourPayload)
	reqTour := httptest.NewRequest(http.MethodPost, "/api/v1/tournaments", bytes.NewReader(body))
	reqTour.Header.Set("Content-Type", "application/json")
	reqTour.AddCookie(cookie)
	recTour := httptest.NewRecorder()
	router.ServeHTTP(recTour, reqTour)
	if recTour.Code != http.StatusCreated {
		t.Fatalf("failed to create tournament: %s", recTour.Body.String())
	}

	var tourResp struct {
		Data struct {
			ID     string         `json:"id"`
			Stages []domain.Stage `json:"stages"`
		} `json:"data"`
	}
	_ = json.Unmarshal(recTour.Body.Bytes(), &tourResp)
	tournamentID := tourResp.Data.ID
	stageID := tourResp.Data.Stages[0].ID.String()

	// Add 2 participants
	for i := 1; i <= 2; i++ {
		pBody, _ := json.Marshal(map[string]interface{}{
			"name":        fmt.Sprintf("Team Public %d", i),
			"seed_number": i,
		})
		pReq := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tournaments/%s/participants", tournamentID), bytes.NewReader(pBody))
		pReq.Header.Set("Content-Type", "application/json")
		pReq.AddCookie(cookie)
		pRec := httptest.NewRecorder()
		router.ServeHTTP(pRec, pReq)
		if pRec.Code != http.StatusCreated {
			t.Fatalf("failed to add participant: %s", pRec.Body.String())
		}
	}

	// Generate bracket
	genReq := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tournaments/%s/stages/%s/generate", tournamentID, stageID), nil)
	genReq.AddCookie(cookie)
	genRec := httptest.NewRecorder()
	router.ServeHTTP(genRec, genReq)
	if genRec.Code != http.StatusOK {
		t.Fatalf("failed to generate bracket: %s", genRec.Body.String())
	}

	// 1. Test Public Spectator View: GET /api/v1/public/tournaments/{id}
	t.Run("Public Tournament View Success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/public/tournaments/%s", tournamentID), nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
		}

		var viewResp struct {
			Data struct {
				Tournament domain.Tournament `json:"tournament"`
				Stage      domain.Stage      `json:"stage"`
				Matches    []domain.Match    `json:"matches"`
			} `json:"data"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &viewResp); err != nil {
			t.Fatalf("failed to unmarshal public view response: %v", err)
		}

		if viewResp.Data.Tournament.ID.String() != tournamentID {
			t.Errorf("expected tournament id %s, got %s", tournamentID, viewResp.Data.Tournament.ID.String())
		}
		if len(viewResp.Data.Matches) != 1 {
			t.Errorf("expected 1 match, got %d", len(viewResp.Data.Matches))
		}
	})

	t.Run("Public Tournament View 404 for unknown ID", func(t *testing.T) {
		unknownID := uuid.New().String()
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/public/tournaments/%s", unknownID), nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", rec.Code)
		}
	})

	// 2. Test Realtime SSE Stream: GET /api/v1/public/tournaments/{id}/stream
	t.Run("Realtime SSE Stream Broadcast", func(t *testing.T) {
		reqCtx, cancelReq := context.WithCancel(context.Background())
		defer cancelReq()

		reqStream := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/public/tournaments/%s/stream", tournamentID), nil).WithContext(reqCtx)
		recStream := newMockFlusherRecorder()

		// Run SSE stream in background goroutine
		doneChan := make(chan struct{})
		go func() {
			router.ServeHTTP(recStream, reqStream)
			close(doneChan)
		}()

		// Wait briefly for connection and initial handshake comment
		time.Sleep(100 * time.Millisecond)

		initialOutput := recStream.String()
		if !strings.Contains(initialOutput, ": connected") {
			t.Fatalf("expected ': connected' in SSE stream, got: %s", initialOutput)
		}

		// Now trigger a match update from organizer to test SSE broadcast
		// First get the match ID
		pubReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/public/tournaments/%s", tournamentID), nil)
		pubRec := httptest.NewRecorder()
		router.ServeHTTP(pubRec, pubReq)
		var pubData struct {
			Data struct {
				Matches []domain.Match `json:"matches"`
			} `json:"data"`
		}
		_ = json.Unmarshal(pubRec.Body.Bytes(), &pubData)
		matchID := pubData.Data.Matches[0].ID.String()

		// Update match score (2 - 0, COMPLETED)
		scorePayload := map[string]interface{}{
			"score_a": 2,
			"score_b": 0,
			"status":  "COMPLETED",
		}
		sBody, _ := json.Marshal(scorePayload)
		scoreReq := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/v1/tournaments/%s/matches/%s/score", tournamentID, matchID), bytes.NewReader(sBody))
		scoreReq.Header.Set("Content-Type", "application/json")
		scoreReq.AddCookie(cookie)
		scoreRec := httptest.NewRecorder()
		router.ServeHTTP(scoreRec, scoreReq)

		if scoreRec.Code != http.StatusOK {
			t.Fatalf("failed to update score: %s", scoreRec.Body.String())
		}

		// Wait for event to arrive in the SSE stream
		time.Sleep(150 * time.Millisecond)

		streamOutput := recStream.String()
		if !strings.Contains(streamOutput, "event: match_updated") {
			t.Fatalf("expected 'event: match_updated' in SSE stream, got:\n%s", streamOutput)
		}

		// Cancel request context to close stream
		cancelReq()

		select {
		case <-doneChan:
			// Clean shutdown
		case <-time.After(1 * time.Second):
			t.Fatal("SSE handler did not terminate upon context cancellation")
		}
	})
}

// mockFlusherRecorder implements http.ResponseWriter and http.Flusher with concurrent-safe buffer.
type mockFlusherRecorder struct {
	*httptest.ResponseRecorder
	flushed bool
}

func newMockFlusherRecorder() *mockFlusherRecorder {
	return &mockFlusherRecorder{
		ResponseRecorder: httptest.NewRecorder(),
	}
}

func (m *mockFlusherRecorder) Flush() {
	m.flushed = true
}

func (m *mockFlusherRecorder) String() string {
	return m.Body.String()
}

func (m *mockFlusherRecorder) Reader() *bufio.Reader {
	return bufio.NewReader(m.Body)
}
