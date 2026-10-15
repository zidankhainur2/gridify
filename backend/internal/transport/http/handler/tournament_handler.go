package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/gridify/backend/internal/domain"
	"github.com/gridify/backend/internal/realtime"
	"github.com/gridify/backend/internal/service"
	"github.com/gridify/backend/internal/transport/http/httputil"
	"github.com/gridify/backend/internal/transport/http/middleware"
)

type TournamentHandler struct {
	tournamentSvc *service.TournamentService
	hub           *realtime.Hub
}

func NewTournamentHandler(tournamentSvc *service.TournamentService, hub *realtime.Hub) *TournamentHandler {
	return &TournamentHandler{
		tournamentSvc: tournamentSvc,
		hub:           hub,
	}
}

type CreateTournamentPayload struct {
	GameID             string          `json:"game_id"`
	Title              string          `json:"title"`
	Description        string          `json:"description"`
	FormatType         string          `json:"format_type"`
	BestOf             int             `json:"best_of"`
	HomeAway           bool            `json:"home_away"`
	PointsPerWin       int             `json:"points_per_win"`
	PointsPerDraw      int             `json:"points_per_draw"`
	HasThirdPlace      bool            `json:"has_third_place"`
	BRTotalMatches     int             `json:"br_total_matches"`
	BRMaxTeamsPerMatch int             `json:"br_max_teams_per_match"`
	BRPointSystem      json.RawMessage `json:"br_point_system,omitempty"`
}

func (h *TournamentHandler) Create(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())
	userID := middleware.GetUserID(r.Context())
	if userID == uuid.Nil {
		httputil.RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Sesi tidak valid", nil, reqID)
		return
	}

	var p CreateTournamentPayload
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&p); err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "INVALID_PAYLOAD", "Payload JSON tidak valid", nil, reqID)
		return
	}

	gameUUID, err := uuid.Parse(p.GameID)
	if err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "INVALID_GAME_ID", "ID Game tidak valid", nil, reqID)
		return
	}

	tourney, err := h.tournamentSvc.CreateTournament(r.Context(), userID, service.CreateTournamentInput{
		GameID:             gameUUID,
		Title:              p.Title,
		Description:        p.Description,
		FormatType:         p.FormatType,
		BestOf:             p.BestOf,
		HomeAway:           p.HomeAway,
		PointsPerWin:       p.PointsPerWin,
		PointsPerDraw:      p.PointsPerDraw,
		HasThirdPlace:      p.HasThirdPlace,
		BRTotalMatches:     p.BRTotalMatches,
		BRMaxTeamsPerMatch: p.BRMaxTeamsPerMatch,
		BRPointSystem:      p.BRPointSystem,
	})
	if err != nil {
		if errors.Is(err, domain.ErrBadRequest) {
			httputil.RespondError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil, reqID)
			return
		}
		httputil.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Gagal membuat turnamen", nil, reqID)
		return
	}

	httputil.JSON(w, http.StatusCreated, tourney)
}

func (h *TournamentHandler) List(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())
	userID := middleware.GetUserID(r.Context())
	if userID == uuid.Nil {
		httputil.RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Sesi tidak valid", nil, reqID)
		return
	}

	q := r.URL.Query().Get("q")
	status := r.URL.Query().Get("status")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
	if perPage < 1 {
		perPage = 20
	}

	list, total, err := h.tournamentSvc.ListTournaments(r.Context(), userID, q, status, page, perPage)
	if err != nil {
		httputil.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Gagal mengambil daftar turnamen", nil, reqID)
		return
	}

	meta := map[string]interface{}{
		"page":     page,
		"per_page": perPage,
		"total":    total,
	}
	httputil.JSONWithMeta(w, http.StatusOK, list, meta)
}

func (h *TournamentHandler) Get(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())
	tourneyIDStr := chi.URLParam(r, "id")
	tourneyID, err := uuid.Parse(tourneyIDStr)
	if err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "INVALID_ID", "ID Turnamen tidak valid", nil, reqID)
		return
	}

	tourney, err := h.tournamentSvc.GetTournament(r.Context(), tourneyID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			httputil.RespondError(w, http.StatusNotFound, "NOT_FOUND", "Turnamen tidak ditemukan", nil, reqID)
			return
		}
		httputil.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Gagal mengambil detail turnamen", nil, reqID)
		return
	}

	httputil.JSON(w, http.StatusOK, tourney)
}

type UpdateTournamentPayload struct {
	Title         *string `json:"title,omitempty"`
	Description   *string `json:"description,omitempty"`
	Status        *string `json:"status,omitempty"`
	BestOf        *int    `json:"best_of,omitempty"`
	HomeAway      *bool   `json:"home_away,omitempty"`
	PointsPerWin  *int    `json:"points_per_win,omitempty"`
	PointsPerDraw *int    `json:"points_per_draw,omitempty"`
	HasThirdPlace *bool   `json:"has_third_place,omitempty"`
}

func (h *TournamentHandler) Update(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())
	userID := middleware.GetUserID(r.Context())
	if userID == uuid.Nil {
		httputil.RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Sesi tidak valid", nil, reqID)
		return
	}

	tourneyIDStr := chi.URLParam(r, "id")
	tourneyID, err := uuid.Parse(tourneyIDStr)
	if err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "INVALID_ID", "ID Turnamen tidak valid", nil, reqID)
		return
	}

	var p UpdateTournamentPayload
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&p); err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "INVALID_PAYLOAD", "Payload JSON tidak valid", nil, reqID)
		return
	}

	tourney, err := h.tournamentSvc.UpdateTournament(r.Context(), tourneyID, userID, service.UpdateTournamentInput{
		Title:         p.Title,
		Description:   p.Description,
		Status:        p.Status,
		BestOf:        p.BestOf,
		HomeAway:      p.HomeAway,
		PointsPerWin:  p.PointsPerWin,
		PointsPerDraw: p.PointsPerDraw,
		HasThirdPlace: p.HasThirdPlace,
	})
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			httputil.RespondError(w, http.StatusNotFound, "NOT_FOUND", "Turnamen tidak ditemukan", nil, reqID)
			return
		}
		if errors.Is(err, domain.ErrForbidden) {
			httputil.RespondError(w, http.StatusForbidden, "FORBIDDEN", "Anda tidak memiliki akses ke turnamen ini", nil, reqID)
			return
		}
		if errors.Is(err, domain.ErrBadRequest) {
			httputil.RespondError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil, reqID)
			return
		}
		httputil.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Gagal memperbarui turnamen", nil, reqID)
		return
	}

	if h.hub != nil {
		h.hub.Broadcast(tourneyID, "tournament_updated", tourney)
	}

	httputil.JSON(w, http.StatusOK, tourney)
}

func (h *TournamentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())
	userID := middleware.GetUserID(r.Context())
	if userID == uuid.Nil {
		httputil.RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Sesi tidak valid", nil, reqID)
		return
	}

	tourneyIDStr := chi.URLParam(r, "id")
	tourneyID, err := uuid.Parse(tourneyIDStr)
	if err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "INVALID_ID", "ID Turnamen tidak valid", nil, reqID)
		return
	}

	if err := h.tournamentSvc.DeleteTournament(r.Context(), tourneyID, userID); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			httputil.RespondError(w, http.StatusNotFound, "NOT_FOUND", "Turnamen tidak ditemukan", nil, reqID)
			return
		}
		if errors.Is(err, domain.ErrForbidden) {
			httputil.RespondError(w, http.StatusForbidden, "FORBIDDEN", "Anda tidak memiliki akses untuk menghapus turnamen ini", nil, reqID)
			return
		}
		httputil.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Gagal menghapus turnamen", nil, reqID)
		return
	}

	httputil.JSON(w, http.StatusOK, map[string]string{
		"message": "Turnamen berhasil dihapus",
	})
}
