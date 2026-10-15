package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/gridify/backend/internal/database"
	"github.com/gridify/backend/internal/domain"
	"github.com/gridify/backend/internal/realtime"
	"github.com/gridify/backend/internal/repository"
	"github.com/gridify/backend/internal/service/bracket"
	"github.com/gridify/backend/internal/transport/http/httputil"
	"github.com/gridify/backend/internal/transport/http/middleware"
)

type MatchHandler struct {
	pool           *pgxpool.Pool
	tournamentRepo *repository.TournamentRepository
	matchRepo      *repository.MatchRepository
	brRepo         *repository.BattleRoyaleRepository
	progression    *bracket.ProgressionEngine
	hub            *realtime.Hub
}

func NewMatchHandler(
	pool *pgxpool.Pool,
	tournamentRepo *repository.TournamentRepository,
	matchRepo *repository.MatchRepository,
	brRepo *repository.BattleRoyaleRepository,
	progression *bracket.ProgressionEngine,
	hub *realtime.Hub,
) *MatchHandler {
	return &MatchHandler{
		pool:           pool,
		tournamentRepo: tournamentRepo,
		matchRepo:      matchRepo,
		brRepo:         brRepo,
		progression:    progression,
		hub:            hub,
	}
}

type UpdateMatchScorePayload struct {
	ScoreA int    `json:"score_a"`
	ScoreB int    `json:"score_b"`
	Status string `json:"status,omitempty"`
}

func (h *MatchHandler) UpdateScore(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())
	userID := middleware.GetUserID(r.Context())

	tournamentID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "INVALID_ID", "ID turnamen tidak valid", nil, reqID)
		return
	}

	matchID, err := uuid.Parse(chi.URLParam(r, "matchId"))
	if err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "INVALID_MATCH_ID", "ID pertandingan tidak valid", nil, reqID)
		return
	}

	// Verify tournament ownership
	tournament, err := h.tournamentRepo.GetTournamentByID(r.Context(), tournamentID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			httputil.RespondError(w, http.StatusNotFound, "NOT_FOUND", "Turnamen tidak ditemukan", nil, reqID)
			return
		}
		httputil.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Gagal memuat turnamen", nil, reqID)
		return
	}

	if tournament.OwnerID != userID {
		httputil.RespondError(w, http.StatusForbidden, "FORBIDDEN", "Anda bukan pemilik turnamen ini", nil, reqID)
		return
	}

	var p UpdateMatchScorePayload
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&p); err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "INVALID_PAYLOAD", "Payload JSON tidak valid", nil, reqID)
		return
	}

	var updatedMatch *domain.Match
	err = database.WithTx(r.Context(), h.pool, func(tx pgx.Tx) error {
		m, err := h.progression.UpdateMatchScore(r.Context(), tx, tournamentID, matchID, p.ScoreA, p.ScoreB, p.Status)
		if err != nil {
			return err
		}
		updatedMatch = m
		return nil
	})

	if err != nil {
		if errors.Is(err, domain.ErrDownstreamMatchPlayed) {
			httputil.RespondError(w, http.StatusConflict, "DOWNSTREAM_MATCH_ALREADY_PLAYED",
				"Pertandingan lanjutan sudah dimainkan. Reset hasil pertandingan lanjutan terlebih dahulu untuk mengubah pemenang.", nil, reqID)
			return
		}
		if errors.Is(err, domain.ErrBadRequest) || errors.Is(err, domain.ErrInvalidInput) {
			httputil.RespondError(w, http.StatusBadRequest, "INVALID_SCORE", err.Error(), nil, reqID)
			return
		}
		if errors.Is(err, domain.ErrNotFound) {
			httputil.RespondError(w, http.StatusNotFound, "NOT_FOUND", "Pertandingan tidak ditemukan", nil, reqID)
			return
		}
		httputil.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Gagal memperbarui skor pertandingan", nil, reqID)
		return
	}

	if h.hub != nil {
		h.hub.Broadcast(tournamentID, "match_updated", updatedMatch)
	}

	httputil.JSON(w, http.StatusOK, updatedMatch)
}

type BRResultEntryPayload struct {
	ParticipantID   string `json:"participant_id"`
	PlacementRank   int    `json:"placement_rank"`
	Kills           int    `json:"kills"`
	PlacementPoints int    `json:"placement_points"`
	KillPoints      int    `json:"kill_points"`
}

type UpdateBRResultsPayload struct {
	Results []BRResultEntryPayload `json:"results"`
}

func (h *MatchHandler) UpdateBRResults(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())
	userID := middleware.GetUserID(r.Context())

	tournamentID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "INVALID_ID", "ID turnamen tidak valid", nil, reqID)
		return
	}

	matchID, err := uuid.Parse(chi.URLParam(r, "matchId"))
	if err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "INVALID_MATCH_ID", "ID pertandingan tidak valid", nil, reqID)
		return
	}

	tournament, err := h.tournamentRepo.GetTournamentByID(r.Context(), tournamentID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			httputil.RespondError(w, http.StatusNotFound, "NOT_FOUND", "Turnamen tidak ditemukan", nil, reqID)
			return
		}
		httputil.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Gagal memuat turnamen", nil, reqID)
		return
	}

	if tournament.OwnerID != userID {
		httputil.RespondError(w, http.StatusForbidden, "FORBIDDEN", "Anda bukan pemilik turnamen ini", nil, reqID)
		return
	}

	var p UpdateBRResultsPayload
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&p); err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "INVALID_PAYLOAD", "Payload JSON tidak valid", nil, reqID)
		return
	}

	var results []domain.BattleRoyaleMatchResult
	for _, entry := range p.Results {
		pUUID, err := uuid.Parse(entry.ParticipantID)
		if err != nil {
			httputil.RespondError(w, http.StatusBadRequest, "INVALID_PARTICIPANT_ID", "ID peserta tidak valid", nil, reqID)
			return
		}
		results = append(results, domain.BattleRoyaleMatchResult{
			MatchID:         matchID,
			ParticipantID:   pUUID,
			PlacementRank:   entry.PlacementRank,
			Kills:           entry.Kills,
			PlacementPoints: entry.PlacementPoints,
			KillPoints:      entry.KillPoints,
			TotalPoints:     entry.PlacementPoints + entry.KillPoints,
		})
	}

	err = database.WithTx(r.Context(), h.pool, func(tx pgx.Tx) error {
		if err := h.brRepo.SaveMatchResults(r.Context(), tx, matchID, results); err != nil {
			return err
		}

		// Mark match as COMPLETED
		currMatch, err := h.matchRepo.GetMatchForUpdate(r.Context(), tx, matchID)
		if err != nil {
			return err
		}
		currMatch.Status = "COMPLETED"
		return h.matchRepo.UpdateMatch(r.Context(), tx, currMatch)
	})

	if err != nil {
		httputil.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Gagal menyimpan hasil Battle Royale", nil, reqID)
		return
	}

	if h.hub != nil {
		h.hub.Broadcast(tournamentID, "match_updated", map[string]any{"match_id": matchID})
	}

	httputil.JSON(w, http.StatusOK, map[string]string{
		"message": "Hasil pertandingan Battle Royale berhasil disimpan",
	})
}

type UpdateMatchSlotPayload struct {
	Slot          string  `json:"slot"`
	ParticipantID *string `json:"participant_id,omitempty"`
}

func (h *MatchHandler) UpdateSlot(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())
	userID := middleware.GetUserID(r.Context())

	tournamentID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "INVALID_ID", "ID turnamen tidak valid", nil, reqID)
		return
	}

	matchID, err := uuid.Parse(chi.URLParam(r, "matchId"))
	if err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "INVALID_MATCH_ID", "ID pertandingan tidak valid", nil, reqID)
		return
	}

	tournament, err := h.tournamentRepo.GetTournamentByID(r.Context(), tournamentID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			httputil.RespondError(w, http.StatusNotFound, "NOT_FOUND", "Turnamen tidak ditemukan", nil, reqID)
			return
		}
		httputil.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Gagal memuat turnamen", nil, reqID)
		return
	}

	if tournament.OwnerID != userID {
		httputil.RespondError(w, http.StatusForbidden, "FORBIDDEN", "Anda bukan pemilik turnamen ini", nil, reqID)
		return
	}

	var p UpdateMatchSlotPayload
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&p); err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "INVALID_PAYLOAD", "Payload JSON tidak valid", nil, reqID)
		return
	}

	var partUUID *uuid.UUID
	if p.ParticipantID != nil && *p.ParticipantID != "" {
		parsed, err := uuid.Parse(*p.ParticipantID)
		if err != nil {
			httputil.RespondError(w, http.StatusBadRequest, "INVALID_PARTICIPANT_ID", "ID peserta tidak valid", nil, reqID)
			return
		}
		partUUID = &parsed
	}

	var updatedMatch *domain.Match
	err = database.WithTx(r.Context(), h.pool, func(tx pgx.Tx) error {
		m, err := h.progression.UpdateMatchSlot(r.Context(), tx, tournamentID, matchID, p.Slot, partUUID)
		if err != nil {
			return err
		}
		updatedMatch = m
		return nil
	})

	if err != nil {
		if errors.Is(err, domain.ErrConflict) {
			httputil.RespondError(w, http.StatusConflict, "MATCH_NOT_SCHEDULED", err.Error(), nil, reqID)
			return
		}
		if errors.Is(err, domain.ErrBadRequest) || errors.Is(err, domain.ErrInvalidInput) {
			httputil.RespondError(w, http.StatusBadRequest, "INVALID_INPUT", err.Error(), nil, reqID)
			return
		}
		httputil.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Gagal memperbarui slot pertandingan", nil, reqID)
		return
	}

	if h.hub != nil {
		h.hub.Broadcast(tournamentID, "match_updated", updatedMatch)
	}

	httputil.JSON(w, http.StatusOK, updatedMatch)
}
