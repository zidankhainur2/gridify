package handler

import (
	"net/http"

	"github.com/gridify/backend/internal/repository"
	"github.com/gridify/backend/internal/transport/http/httputil"
	"github.com/gridify/backend/internal/transport/http/middleware"
)

type DashboardHandler struct {
	tournamentRepo *repository.TournamentRepository
}

func NewDashboardHandler(tournamentRepo *repository.TournamentRepository) *DashboardHandler {
	return &DashboardHandler{tournamentRepo: tournamentRepo}
}

func (h *DashboardHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())
	userID := middleware.GetUserID(r.Context())
	if userID == [16]byte{} {
		httputil.RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Sesi tidak valid", nil, reqID)
		return
	}

	summary, err := h.tournamentRepo.GetDashboardSummary(r.Context(), userID)
	if err != nil {
		httputil.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Gagal mengambil ringkasan dashboard", nil, reqID)
		return
	}

	httputil.JSON(w, http.StatusOK, summary)
}
