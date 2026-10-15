package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/gridify/backend/internal/database"
	"github.com/gridify/backend/internal/domain"
	"github.com/gridify/backend/internal/realtime"
	"github.com/gridify/backend/internal/repository"
	"github.com/gridify/backend/internal/service"
	"github.com/gridify/backend/internal/service/bracket"
	"github.com/gridify/backend/internal/transport/http/httputil"
	"github.com/gridify/backend/internal/transport/http/middleware"
)

type BracketHandler struct {
	pool            *pgxpool.Pool
	tournamentRepo  *repository.TournamentRepository
	stageRepo       *repository.StageRepository
	participantRepo *repository.ParticipantRepository
	matchRepo       *repository.MatchRepository
	standingsSvc    *service.StandingsService
	progression     *bracket.ProgressionEngine
	hub             *realtime.Hub
}

func NewBracketHandler(
	pool *pgxpool.Pool,
	tournamentRepo *repository.TournamentRepository,
	stageRepo *repository.StageRepository,
	participantRepo *repository.ParticipantRepository,
	matchRepo *repository.MatchRepository,
	standingsSvc *service.StandingsService,
	progression *bracket.ProgressionEngine,
	hub *realtime.Hub,
) *BracketHandler {
	return &BracketHandler{
		pool:            pool,
		tournamentRepo:  tournamentRepo,
		stageRepo:       stageRepo,
		participantRepo: participantRepo,
		matchRepo:       matchRepo,
		standingsSvc:    standingsSvc,
		progression:     progression,
		hub:             hub,
	}
}

// Generate creates or resets bracket/matches for a stage based on tournament format.
func (h *BracketHandler) Generate(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())
	userID := middleware.GetUserID(r.Context())

	tournamentID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "INVALID_ID", "ID turnamen tidak valid", nil, reqID)
		return
	}

	stageID, err := uuid.Parse(chi.URLParam(r, "stageId"))
	if err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "INVALID_STAGE_ID", "ID babak tidak valid", nil, reqID)
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

	// Verify ownership
	if tournament.OwnerID != userID {
		httputil.RespondError(w, http.StatusForbidden, "FORBIDDEN", "Anda bukan pemilik turnamen ini", nil, reqID)
		return
	}

	stage, err := h.stageRepo.GetStageByID(r.Context(), stageID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			httputil.RespondError(w, http.StatusNotFound, "NOT_FOUND", "Babak tidak ditemukan", nil, reqID)
			return
		}
		httputil.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Gagal memuat babak", nil, reqID)
		return
	}

	participants, err := h.participantRepo.ListParticipantsByTournamentID(r.Context(), tournamentID)
	if err != nil {
		httputil.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Gagal memuat peserta", nil, reqID)
		return
	}

	if len(participants) < 2 {
		httputil.RespondError(w, http.StatusBadRequest, "INSUFFICIENT_PARTICIPANTS", "Minimal 2 peserta diperlukan untuk membuat bagan pertandingan", nil, reqID)
		return
	}

	err = database.WithTx(r.Context(), h.pool, func(tx pgx.Tx) error {
		switch tournament.FormatType {
		case "SINGLE_ELIMINATION":
			_, err = bracket.GenerateSingleElimination(r.Context(), tx, h.matchRepo, tournament, stage, participants)
		case "DOUBLE_ELIMINATION":
			_, err = bracket.GenerateDoubleElimination(r.Context(), tx, h.matchRepo, tournament, stage, participants)
		case "ROUND_ROBIN":
			_, err = bracket.GenerateRoundRobin(r.Context(), tx, h.matchRepo, tournament, stage, participants)
		case "HYBRID_UCL":
			if stage.Type == "KNOCKOUT" {
				_, err = bracket.GenerateSingleElimination(r.Context(), tx, h.matchRepo, tournament, stage, participants)
			} else {
				_, err = bracket.GenerateRoundRobin(r.Context(), tx, h.matchRepo, tournament, stage, participants)
			}
		case "BATTLE_ROYALE":
			_, err = bracket.GenerateBattleRoyale(r.Context(), tx, h.matchRepo, tournament, stage, participants)
		default:
			return fmt.Errorf("%w: format turnamen tidak didukung", domain.ErrBadRequest)
		}
		return err
	})

	if err != nil {
		if errors.Is(err, domain.ErrInvalidInput) || errors.Is(err, domain.ErrBadRequest) {
			httputil.RespondError(w, http.StatusBadRequest, "INVALID_INPUT", err.Error(), nil, reqID)
			return
		}
		httputil.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Gagal membuat jadwal bagan: "+err.Error(), nil, reqID)
		return
	}

	// Fetch updated stage view
	view, err := h.standingsSvc.GetStageView(r.Context(), tournamentID, stageID)
	if err != nil {
		httputil.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Gagal memuat tampilan bagan", nil, reqID)
		return
	}

	if h.hub != nil {
		h.hub.Broadcast(tournamentID, "stage_updated", view)
	}

	httputil.JSON(w, http.StatusOK, view)
}

// AdvanceKnockout advances top teams from Group Stage to Knockout stage (Hybrid UCL).
func (h *BracketHandler) AdvanceKnockout(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())
	userID := middleware.GetUserID(r.Context())

	tournamentID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "INVALID_ID", "ID turnamen tidak valid", nil, reqID)
		return
	}

	stageID, err := uuid.Parse(chi.URLParam(r, "stageId"))
	if err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "INVALID_STAGE_ID", "ID babak tidak valid", nil, reqID)
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

	var knockoutStage *domain.Stage
	err = database.WithTx(r.Context(), h.pool, func(tx pgx.Tx) error {
		stg, _, err := h.progression.AdvanceGroupToKnockout(r.Context(), tx, tournamentID, stageID)
		if err != nil {
			return err
		}
		knockoutStage = stg
		return nil
	})

	if err != nil {
		if errors.Is(err, domain.ErrBadRequest) || errors.Is(err, domain.ErrInvalidInput) {
			httputil.RespondError(w, http.StatusBadRequest, "INVALID_INPUT", err.Error(), nil, reqID)
			return
		}
		httputil.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Gagal memajukan ke babak knockout", nil, reqID)
		return
	}

	view, err := h.standingsSvc.GetStageView(r.Context(), tournamentID, knockoutStage.ID)
	if err != nil {
		httputil.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Gagal memuat tampilan knockout", nil, reqID)
		return
	}

	if h.hub != nil {
		h.hub.Broadcast(tournamentID, "stage_updated", view)
	}

	httputil.JSON(w, http.StatusOK, view)
}

// GetView returns the full stage view (matches, standings/leaderboard, and champions).
func (h *BracketHandler) GetView(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())

	tournamentID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "INVALID_ID", "ID turnamen tidak valid", nil, reqID)
		return
	}

	stageID, err := uuid.Parse(chi.URLParam(r, "stageId"))
	if err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "INVALID_STAGE_ID", "ID babak tidak valid", nil, reqID)
		return
	}

	view, err := h.standingsSvc.GetStageView(r.Context(), tournamentID, stageID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			httputil.RespondError(w, http.StatusNotFound, "NOT_FOUND", "Data babak atau turnamen tidak ditemukan", nil, reqID)
			return
		}
		httputil.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Gagal memuat tampilan babak", nil, reqID)
		return
	}

	httputil.JSON(w, http.StatusOK, view)
}
