package handler

import (
	"net/http"

	"github.com/gridify/backend/internal/repository"
	"github.com/gridify/backend/internal/transport/http/httputil"
	"github.com/gridify/backend/internal/transport/http/middleware"
)

type GameHandler struct {
	gameRepo *repository.GameRepository
}

func NewGameHandler(gameRepo *repository.GameRepository) *GameHandler {
	return &GameHandler{gameRepo: gameRepo}
}

func (h *GameHandler) ListGames(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())
	games, err := h.gameRepo.GetAllGames(r.Context())
	if err != nil {
		httputil.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Gagal mengambil daftar game", nil, reqID)
		return
	}
	httputil.JSON(w, http.StatusOK, games)
}
