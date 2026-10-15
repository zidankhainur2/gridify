package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/gridify/backend/internal/domain"
	"github.com/gridify/backend/internal/realtime"
	"github.com/gridify/backend/internal/repository"
	"github.com/gridify/backend/internal/service"
	"github.com/gridify/backend/internal/transport/http/httputil"
	"github.com/gridify/backend/internal/transport/http/middleware"
)

type PublicHandler struct {
	tournamentRepo *repository.TournamentRepository
	stageRepo      *repository.StageRepository
	standingsSvc   *service.StandingsService
	hub            *realtime.Hub
	logger         *slog.Logger
}

func NewPublicHandler(
	tournamentRepo *repository.TournamentRepository,
	stageRepo *repository.StageRepository,
	standingsSvc *service.StandingsService,
	hub *realtime.Hub,
	logger *slog.Logger,
) *PublicHandler {
	if logger == nil {
		logger = slog.Default()
	}
	return &PublicHandler{
		tournamentRepo: tournamentRepo,
		stageRepo:      stageRepo,
		standingsSvc:   standingsSvc,
		hub:            hub,
		logger:         logger,
	}
}

// GetTournament handles GET /api/v1/public/tournaments/{id}
func (h *PublicHandler) GetTournament(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())

	tournamentID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "INVALID_ID", "ID turnamen tidak valid", nil, reqID)
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

	stages, err := h.stageRepo.ListStagesByTournamentID(r.Context(), tournamentID)
	if err != nil {
		httputil.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Gagal memuat babak turnamen", nil, reqID)
		return
	}

	if len(stages) == 0 {
		httputil.JSON(w, http.StatusOK, map[string]any{
			"tournament": tournament,
			"stages":     []domain.Stage{},
		})
		return
	}

	// Determine which stage to show
	var targetStageID uuid.UUID
	stageIDQuery := r.URL.Query().Get("stage_id")
	if stageIDQuery != "" {
		parsed, err := uuid.Parse(stageIDQuery)
		if err == nil {
			targetStageID = parsed
		}
	}

	if targetStageID == uuid.Nil {
		// Prefer ONGOING stage, else first stage
		for _, s := range stages {
			if s.Status == "ONGOING" {
				targetStageID = s.ID
				break
			}
		}
		if targetStageID == uuid.Nil {
			targetStageID = stages[0].ID
		}
	}

	view, err := h.standingsSvc.GetStageView(r.Context(), tournamentID, targetStageID)
	if err != nil {
		httputil.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Gagal memuat tampilan turnamen", nil, reqID)
		return
	}

	httputil.JSON(w, http.StatusOK, view)
}

// StreamTournament handles GET /api/v1/public/tournaments/{id}/stream
func (h *PublicHandler) StreamTournament(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())

	tournamentID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "INVALID_ID", "ID turnamen tidak valid", nil, reqID)
		return
	}

	// Verify tournament exists
	_, err = h.tournamentRepo.GetTournamentByID(r.Context(), tournamentID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			httputil.RespondError(w, http.StatusNotFound, "NOT_FOUND", "Turnamen tidak ditemukan", nil, reqID)
			return
		}
		httputil.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Gagal memuat turnamen", nil, reqID)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		httputil.RespondError(w, http.StatusInternalServerError, "STREAMING_UNSUPPORTED", "Browser atau server tidak mendukung HTTP streaming", nil, reqID)
		return
	}

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	// Send initial connection comment
	fmt.Fprintf(w, ": connected\n\n")
	flusher.Flush()

	clientChan := h.hub.Subscribe(tournamentID)
	defer h.hub.Unsubscribe(tournamentID, clientChan)

	ticker := time.NewTicker(25 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-ticker.C:
			fmt.Fprintf(w, ": ping\n\n")
			flusher.Flush()
		case evt, ok := <-clientChan:
			if !ok {
				return
			}
			dataBytes, err := json.Marshal(evt.Data)
			if err != nil {
				continue
			}
			fmt.Fprintf(w, "event: %s\ndata: %s\n\n", evt.Type, string(dataBytes))
			flusher.Flush()
		}
	}
}
