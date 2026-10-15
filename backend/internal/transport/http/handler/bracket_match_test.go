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

	"github.com/gridify/backend/internal/config"
	"github.com/gridify/backend/internal/database"
	"github.com/gridify/backend/internal/domain"
	"github.com/gridify/backend/internal/service"
	transporthttp "github.com/gridify/backend/internal/transport/http"
)

func TestBracketAndProgressionEngineFlow(t *testing.T) {
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

	// Setup User A (Organizer) and User B (Attacker)
	_, cookieA := createTestUser(t, router, "org_prog_a")
	_, cookieB := createTestUser(t, router, "org_prog_b")

	// Get a master game
	reqGame := httptest.NewRequest(http.MethodGet, "/api/v1/games", nil)
	recGame := httptest.NewRecorder()
	router.ServeHTTP(recGame, reqGame)
	if recGame.Code != http.StatusOK {
		t.Fatalf("failed to get games: %s", recGame.Body.String())
	}
	var gameResp struct {
		Data []domain.Game `json:"data"`
	}
	_ = json.Unmarshal(recGame.Body.Bytes(), &gameResp)
	gameID := gameResp.Data[0].ID.String()

	// 1. User A creates a SINGLE_ELIMINATION tournament with HasThirdPlace = true, BestOf = 1
	createTourPayload := map[string]interface{}{
		"game_id":         gameID,
		"title":           "Progression Cup 2026",
		"format_type":     "SINGLE_ELIMINATION",
		"best_of":         1,
		"has_third_place": true,
	}
	body, _ := json.Marshal(createTourPayload)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/tournaments", bytes.NewReader(body))
	req.AddCookie(cookieA)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("failed to create tournament: %s", rec.Body.String())
	}
	var tourResp struct {
		Data domain.Tournament `json:"data"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &tourResp)
	tourID := tourResp.Data.ID.String()
	stageID := tourResp.Data.Stages[0].ID.String()

	// 2. Add 4 participants
	teamIDs := make([]string, 4)
	for i := 0; i < 4; i++ {
		pPayload := map[string]interface{}{
			"name": fmt.Sprintf("Team Alpha %d", i+1),
		}
		pBody, _ := json.Marshal(pPayload)
		pReq := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tournaments/%s/participants", tourID), bytes.NewReader(pBody))
		pReq.AddCookie(cookieA)
		pRec := httptest.NewRecorder()
		router.ServeHTTP(pRec, pReq)
		if pRec.Code != http.StatusCreated {
			t.Fatalf("failed to add participant %d: %s", i, pRec.Body.String())
		}
		var pResp struct {
			Data domain.Participant `json:"data"`
		}
		_ = json.Unmarshal(pRec.Body.Bytes(), &pResp)
		teamIDs[i] = pResp.Data.ID.String()
	}

	t.Run("User B (non-owner) generate bracket returns 403 Forbidden", func(t *testing.T) {
		genReq := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tournaments/%s/stages/%s/generate", tourID, stageID), nil)
		genReq.AddCookie(cookieB)
		genRec := httptest.NewRecorder()
		router.ServeHTTP(genRec, genReq)
		if genRec.Code != http.StatusForbidden {
			t.Errorf("expected 403 Forbidden, got %d", genRec.Code)
		}
	})

	var matches []domain.Match
	t.Run("User A generates bracket successfully", func(t *testing.T) {
		genReq := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tournaments/%s/stages/%s/generate", tourID, stageID), nil)
		genReq.AddCookie(cookieA)
		genRec := httptest.NewRecorder()
		router.ServeHTTP(genRec, genReq)
		if genRec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d: %s", genRec.Code, genRec.Body.String())
		}

		var viewResp struct {
			Data struct {
				Matches []domain.Match `json:"matches"`
			} `json:"data"`
		}
		_ = json.Unmarshal(genRec.Body.Bytes(), &viewResp)
		matches = viewResp.Data.Matches
		// 4 teams + bronze match -> 4 matches total (2 Semis, 1 Final, 1 Third Place)
		if len(matches) != 4 {
			t.Fatalf("expected 4 matches, got %d", len(matches))
		}
	})

	// Find R1 Match 1 and R1 Match 2 (the Semifinals)
	var semi1, semi2, finalMatch, thirdPlace *domain.Match
	for i := range matches {
		m := &matches[i]
		if m.BracketType == "UPPER" && m.RoundNumber == 1 && m.MatchNumber == 1 {
			semi1 = m
		} else if m.BracketType == "UPPER" && m.RoundNumber == 1 && m.MatchNumber == 2 {
			semi2 = m
		} else if m.BracketType == "UPPER" && m.RoundNumber == 2 {
			finalMatch = m
		} else if m.BracketType == "THIRD_PLACE" {
			thirdPlace = m
		}
	}

	if semi1 == nil || semi2 == nil || finalMatch == nil || thirdPlace == nil {
		t.Fatalf("could not find all matches")
	}

	t.Run("User B (non-owner) update score returns 403 Forbidden", func(t *testing.T) {
		scorePayload := map[string]interface{}{
			"score_a": 1,
			"score_b": 0,
			"status":  "COMPLETED",
		}
		sBody, _ := json.Marshal(scorePayload)
		sReq := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/v1/tournaments/%s/matches/%s/score", tourID, semi1.ID.String()), bytes.NewReader(sBody))
		sReq.AddCookie(cookieB)
		sRec := httptest.NewRecorder()
		router.ServeHTTP(sRec, sReq)
		if sRec.Code != http.StatusForbidden {
			t.Errorf("expected 403 Forbidden, got %d", sRec.Code)
		}
	})

	t.Run("User A completes Semifinal 1 and winner advances to Final, loser to Third Place", func(t *testing.T) {
		scorePayload := map[string]interface{}{
			"score_a": 1,
			"score_b": 0,
			"status":  "COMPLETED",
		}
		sBody, _ := json.Marshal(scorePayload)
		sReq := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/v1/tournaments/%s/matches/%s/score", tourID, semi1.ID.String()), bytes.NewReader(sBody))
		sReq.AddCookie(cookieA)
		sRec := httptest.NewRecorder()
		router.ServeHTTP(sRec, sReq)
		if sRec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d: %s", sRec.Code, sRec.Body.String())
		}

		// Check Final match slot A received the winner
		vReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/tournaments/%s/stages/%s/view", tourID, stageID), nil)
		vRec := httptest.NewRecorder()
		router.ServeHTTP(vRec, vReq)
		var viewResp struct {
			Data struct {
				Matches []domain.Match `json:"matches"`
			} `json:"data"`
		}
		_ = json.Unmarshal(vRec.Body.Bytes(), &viewResp)
		for _, m := range viewResp.Data.Matches {
			if m.ID == finalMatch.ID {
				if m.ParticipantAID == nil || *m.ParticipantAID != *semi1.ParticipantAID {
					t.Errorf("expected final match slot A to have semi1 winner")
				}
			}
			if m.ID == thirdPlace.ID {
				if m.ParticipantAID == nil || *m.ParticipantAID != *semi1.ParticipantBID {
					t.Errorf("expected third place match slot A to have semi1 loser")
				}
			}
		}
	})

	t.Run("User A completes Semifinal 2 and then Final match", func(t *testing.T) {
		// Semifinal 2: Team A wins 1-0
		scorePayload2 := map[string]interface{}{
			"score_a": 1,
			"score_b": 0,
			"status":  "COMPLETED",
		}
		sBody2, _ := json.Marshal(scorePayload2)
		sReq2 := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/v1/tournaments/%s/matches/%s/score", tourID, semi2.ID.String()), bytes.NewReader(sBody2))
		sReq2.AddCookie(cookieA)
		sRec2 := httptest.NewRecorder()
		router.ServeHTTP(sRec2, sReq2)
		if sRec2.Code != http.StatusOK {
			t.Fatalf("failed to complete semi2: %s", sRec2.Body.String())
		}

		// Final match: play and complete
		scorePayloadFinal := map[string]interface{}{
			"score_a": 1,
			"score_b": 0,
			"status":  "COMPLETED",
		}
		sBodyF, _ := json.Marshal(scorePayloadFinal)
		sReqF := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/v1/tournaments/%s/matches/%s/score", tourID, finalMatch.ID.String()), bytes.NewReader(sBodyF))
		sReqF.AddCookie(cookieA)
		sRecF := httptest.NewRecorder()
		router.ServeHTTP(sRecF, sReqF)
		if sRecF.Code != http.StatusOK {
			t.Fatalf("failed to complete final match: %s", sRecF.Body.String())
		}
	})

	t.Run("Changing winner of Semifinal 1 now returns 409 Conflict (downstream played)", func(t *testing.T) {
		// Final match is already COMPLETED! Changing semi1 winner should be rejected with 409
		scorePayloadReversal := map[string]interface{}{
			"score_a": 0,
			"score_b": 1,
			"status":  "COMPLETED",
		}
		sBodyR, _ := json.Marshal(scorePayloadReversal)
		sReqR := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/v1/tournaments/%s/matches/%s/score", tourID, semi1.ID.String()), bytes.NewReader(sBodyR))
		sReqR.AddCookie(cookieA)
		sRecR := httptest.NewRecorder()
		router.ServeHTTP(sRecR, sReqR)
		if sRecR.Code != http.StatusConflict {
			t.Errorf("expected 409 Conflict, got %d: %s", sRecR.Code, sRecR.Body.String())
		}
	})

	t.Run("Spectator view returns stage matches and champions", func(t *testing.T) {
		vReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/tournaments/%s/stages/%s/view", tourID, stageID), nil)
		vRec := httptest.NewRecorder()
		router.ServeHTTP(vRec, vReq)
		if vRec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d", vRec.Code)
		}
		var viewResp struct {
			Data struct {
				Champions *service.Champions `json:"champions"`
			} `json:"data"`
		}
		_ = json.Unmarshal(vRec.Body.Bytes(), &viewResp)
		if viewResp.Data.Champions == nil || viewResp.Data.Champions.First == nil {
			t.Errorf("expected 1st place champion in view response")
		}
	})

	t.Run("Hybrid UCL flow: generate group stage then advance to knockout stage", func(t *testing.T) {
		// 1. Create Hybrid UCL tournament
		createPayload := map[string]interface{}{
			"game_id":     gameID,
			"title":       "Hybrid UCL Champions League",
			"format_type": "HYBRID_UCL",
			"best_of":     1,
		}
		cBody, _ := json.Marshal(createPayload)
		cReq := httptest.NewRequest(http.MethodPost, "/api/v1/tournaments", bytes.NewReader(cBody))
		cReq.AddCookie(cookieA)
		cRec := httptest.NewRecorder()
		router.ServeHTTP(cRec, cReq)
		if cRec.Code != http.StatusCreated {
			t.Fatalf("failed to create hybrid ucl tournament: %s", cRec.Body.String())
		}
		var tResp struct {
			Data domain.Tournament `json:"data"`
		}
		_ = json.Unmarshal(cRec.Body.Bytes(), &tResp)
		hTourID := tResp.Data.ID.String()
		hStageID := tResp.Data.Stages[0].ID.String()

		// 2. Add 8 participants (will form 2 groups of 4)
		for i := 0; i < 8; i++ {
			pPayload := map[string]interface{}{
				"name": fmt.Sprintf("UCL Club %d", i+1),
			}
			pBody, _ := json.Marshal(pPayload)
			pReq := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tournaments/%s/participants", hTourID), bytes.NewReader(pBody))
			pReq.AddCookie(cookieA)
			pRec := httptest.NewRecorder()
			router.ServeHTTP(pRec, pReq)
			if pRec.Code != http.StatusCreated {
				t.Fatalf("failed to add participant %d: %s", i, pRec.Body.String())
			}
		}

		// 3. Generate group stage schedule
		genReq := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tournaments/%s/stages/%s/generate", hTourID, hStageID), nil)
		genReq.AddCookie(cookieA)
		genRec := httptest.NewRecorder()
		router.ServeHTTP(genRec, genReq)
		if genRec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK for UCL group generate, got %d: %s", genRec.Code, genRec.Body.String())
		}

		// 4. Advance to Knockout stage
		advReq := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tournaments/%s/stages/%s/advance", hTourID, hStageID), nil)
		advReq.AddCookie(cookieA)
		advRec := httptest.NewRecorder()
		router.ServeHTTP(advRec, advReq)
		if advRec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK for UCL advance to knockout, got %d: %s", advRec.Code, advRec.Body.String())
		}

		var koView struct {
			Data struct {
				Stage   domain.Stage   `json:"stage"`
				Matches []domain.Match `json:"matches"`
			} `json:"data"`
		}
		_ = json.Unmarshal(advRec.Body.Bytes(), &koView)
		if koView.Data.Stage.Type != "SINGLE_ELIMINATION" {
			t.Errorf("expected SINGLE_ELIMINATION stage type, got %s", koView.Data.Stage.Type)
		}
		// 4 qualified teams (top 2 from each of 2 groups) -> 3 knockout matches (2 Semis, 1 Final)
		if len(koView.Data.Matches) != 3 {
			t.Errorf("expected 3 knockout matches for 4 teams, got %d", len(koView.Data.Matches))
		}
	})

	t.Run("Battle Royale flow: generate lobby and submit match results", func(t *testing.T) {
		createPayload := map[string]interface{}{
			"game_id":                gameID,
			"title":                  "Apex Legends Showdown",
			"format_type":            "BATTLE_ROYALE",
			"br_total_matches":       2,
			"br_max_teams_per_match": 16,
		}
		cBody, _ := json.Marshal(createPayload)
		cReq := httptest.NewRequest(http.MethodPost, "/api/v1/tournaments", bytes.NewReader(cBody))
		cReq.AddCookie(cookieA)
		cRec := httptest.NewRecorder()
		router.ServeHTTP(cRec, cReq)
		if cRec.Code != http.StatusCreated {
			t.Fatalf("failed to create BR tournament: %s", cRec.Body.String())
		}
		var tResp struct {
			Data domain.Tournament `json:"data"`
		}
		_ = json.Unmarshal(cRec.Body.Bytes(), &tResp)
		brTourID := tResp.Data.ID.String()
		brStageID := tResp.Data.Stages[0].ID.String()

		// Add 3 teams
		var brTeams []string
		for i := 0; i < 3; i++ {
			pPayload := map[string]interface{}{
				"name": fmt.Sprintf("Squad %c", 'A'+i),
			}
			pBody, _ := json.Marshal(pPayload)
			pReq := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tournaments/%s/participants", brTourID), bytes.NewReader(pBody))
			pReq.AddCookie(cookieA)
			pRec := httptest.NewRecorder()
			router.ServeHTTP(pRec, pReq)
			var pResp struct {
				Data domain.Participant `json:"data"`
			}
			_ = json.Unmarshal(pRec.Body.Bytes(), &pResp)
			brTeams = append(brTeams, pResp.Data.ID.String())
		}

		// Generate BR matches
		genReq := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tournaments/%s/stages/%s/generate", brTourID, brStageID), nil)
		genReq.AddCookie(cookieA)
		genRec := httptest.NewRecorder()
		router.ServeHTTP(genRec, genReq)
		if genRec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK for BR generate, got %d: %s", genRec.Code, genRec.Body.String())
		}
		var brView struct {
			Data struct {
				Matches []domain.Match `json:"matches"`
			} `json:"data"`
		}
		_ = json.Unmarshal(genRec.Body.Bytes(), &brView)
		if len(brView.Data.Matches) != 2 {
			t.Fatalf("expected 2 BR matches, got %d", len(brView.Data.Matches))
		}

		// Submit BR results for Match 1
		match1ID := brView.Data.Matches[0].ID.String()
		resultsPayload := map[string]interface{}{
			"results": []map[string]interface{}{
				{
					"participant_id":   brTeams[0],
					"placement_rank":   1,
					"kills":            5,
					"placement_points": 10,
					"kill_points":      5,
				},
				{
					"participant_id":   brTeams[1],
					"placement_rank":   2,
					"kills":            2,
					"placement_points": 6,
					"kill_points":      2,
				},
				{
					"participant_id":   brTeams[2],
					"placement_rank":   3,
					"kills":            0,
					"placement_points": 4,
					"kill_points":      0,
				},
			},
		}
		rBody, _ := json.Marshal(resultsPayload)
		rReq := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/tournaments/%s/matches/%s/br-results", brTourID, match1ID), bytes.NewReader(rBody))
		rReq.AddCookie(cookieA)
		rRec := httptest.NewRecorder()
		router.ServeHTTP(rRec, rReq)
		if rRec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK for submit BR results, got %d: %s", rRec.Code, rRec.Body.String())
		}

		// Verify stage view has BR leaderboard
		vReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/tournaments/%s/stages/%s/view", brTourID, brStageID), nil)
		vRec := httptest.NewRecorder()
		router.ServeHTTP(vRec, vReq)
		var leadView struct {
			Data struct {
				BRLeaderboard []service.BRLeaderboardRow `json:"br_leaderboard"`
			} `json:"data"`
		}
		_ = json.Unmarshal(vRec.Body.Bytes(), &leadView)
		if len(leadView.Data.BRLeaderboard) != 3 {
			t.Fatalf("expected 3 teams in BR leaderboard, got %d", len(leadView.Data.BRLeaderboard))
		}
		if leadView.Data.BRLeaderboard[0].TotalPoints != 15 {
			t.Errorf("Rank 1 expected 15 pts, got %d", leadView.Data.BRLeaderboard[0].TotalPoints)
		}
	})
}
