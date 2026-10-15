package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/gridify/backend/internal/domain"
	"github.com/gridify/backend/internal/realtime"
	"github.com/gridify/backend/internal/service"
	"github.com/gridify/backend/internal/transport/http/httputil"
	"github.com/gridify/backend/internal/transport/http/middleware"
)

type ParticipantHandler struct {
	participantSvc *service.ParticipantService
	hub            *realtime.Hub
}

func NewParticipantHandler(participantSvc *service.ParticipantService, hub *realtime.Hub) *ParticipantHandler {
	return &ParticipantHandler{
		participantSvc: participantSvc,
		hub:            hub,
	}
}

type AddParticipantPayload struct {
	Name        string  `json:"name"`
	ContactInfo *string `json:"contact_info,omitempty"`
	SeedNumber  int     `json:"seed_number,omitempty"`
}

func (h *ParticipantHandler) Add(w http.ResponseWriter, r *http.Request) {
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

	var p AddParticipantPayload
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&p); err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "INVALID_PAYLOAD", "Payload JSON tidak valid", nil, reqID)
		return
	}

	participant, err := h.participantSvc.AddParticipant(r.Context(), service.AddParticipantInput{
		TournamentID: tourneyID,
		UserID:       userID,
		Name:         p.Name,
		ContactInfo:  p.ContactInfo,
		SeedNumber:   p.SeedNumber,
	})
	if err != nil {
		if errors.Is(err, domain.ErrForbidden) {
			httputil.RespondError(w, http.StatusForbidden, "FORBIDDEN", "Anda bukan pemilik turnamen ini", nil, reqID)
			return
		}
		if errors.Is(err, domain.ErrTournamentLocked) {
			httputil.RespondError(w, http.StatusConflict, "TOURNAMENT_LOCKED", "Peserta tidak dapat ditambahkan karena turnamen sedang berlangsung atau selesai", nil, reqID)
			return
		}
		if errors.Is(err, domain.ErrBadRequest) {
			httputil.RespondError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil, reqID)
			return
		}
		httputil.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Gagal menambahkan peserta", nil, reqID)
		return
	}

	if h.hub != nil {
		h.hub.Broadcast(tourneyID, "participant_updated", participant)
	}

	httputil.JSON(w, http.StatusCreated, participant)
}

func (h *ParticipantHandler) List(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())
	tourneyIDStr := chi.URLParam(r, "id")
	tourneyID, err := uuid.Parse(tourneyIDStr)
	if err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "INVALID_ID", "ID Turnamen tidak valid", nil, reqID)
		return
	}

	list, err := h.participantSvc.ListParticipants(r.Context(), tourneyID)
	if err != nil {
		httputil.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Gagal mengambil daftar peserta", nil, reqID)
		return
	}

	httputil.JSON(w, http.StatusOK, list)
}

type UpdateParticipantPayload struct {
	Name        *string `json:"name,omitempty"`
	ContactInfo *string `json:"contact_info,omitempty"`
	SeedNumber  *int    `json:"seed_number,omitempty"`
	GroupName   *string `json:"group_name,omitempty"`
	IsVerified  *bool   `json:"is_verified,omitempty"`
}

func (h *ParticipantHandler) Update(w http.ResponseWriter, r *http.Request) {
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

	pidStr := chi.URLParam(r, "pid")
	pid, err := uuid.Parse(pidStr)
	if err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "INVALID_ID", "ID Peserta tidak valid", nil, reqID)
		return
	}

	var p UpdateParticipantPayload
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&p); err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "INVALID_PAYLOAD", "Payload JSON tidak valid", nil, reqID)
		return
	}

	updated, err := h.participantSvc.UpdateParticipant(r.Context(), service.UpdateParticipantInput{
		TournamentID:  tourneyID,
		ParticipantID: pid,
		UserID:        userID,
		Name:          p.Name,
		ContactInfo:   p.ContactInfo,
		SeedNumber:    p.SeedNumber,
		GroupName:     p.GroupName,
		IsVerified:    p.IsVerified,
	})
	if err != nil {
		if errors.Is(err, domain.ErrForbidden) {
			httputil.RespondError(w, http.StatusForbidden, "FORBIDDEN", "Anda bukan pemilik turnamen ini", nil, reqID)
			return
		}
		if errors.Is(err, domain.ErrNotFound) {
			httputil.RespondError(w, http.StatusNotFound, "NOT_FOUND", "Peserta tidak ditemukan", nil, reqID)
			return
		}
		if errors.Is(err, domain.ErrBadRequest) {
			httputil.RespondError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil, reqID)
			return
		}
		httputil.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Gagal memperbarui peserta", nil, reqID)
		return
	}

	if h.hub != nil {
		h.hub.Broadcast(tourneyID, "participant_updated", updated)
	}

	httputil.JSON(w, http.StatusOK, updated)
}

func (h *ParticipantHandler) Delete(w http.ResponseWriter, r *http.Request) {
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

	pidStr := chi.URLParam(r, "pid")
	pid, err := uuid.Parse(pidStr)
	if err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "INVALID_ID", "ID Peserta tidak valid", nil, reqID)
		return
	}

	if err := h.participantSvc.DeleteParticipant(r.Context(), tourneyID, pid, userID); err != nil {
		if errors.Is(err, domain.ErrForbidden) {
			httputil.RespondError(w, http.StatusForbidden, "FORBIDDEN", "Anda bukan pemilik turnamen ini", nil, reqID)
			return
		}
		if errors.Is(err, domain.ErrTournamentLocked) {
			httputil.RespondError(w, http.StatusConflict, "TOURNAMENT_LOCKED", "Peserta tidak dapat dihapus karena turnamen sedang berlangsung atau selesai", nil, reqID)
			return
		}
		if errors.Is(err, domain.ErrNotFound) {
			httputil.RespondError(w, http.StatusNotFound, "NOT_FOUND", "Peserta tidak ditemukan", nil, reqID)
			return
		}
		httputil.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Gagal menghapus peserta", nil, reqID)
		return
	}

	if h.hub != nil {
		h.hub.Broadcast(tourneyID, "participant_updated", map[string]any{"deleted_id": pid})
	}

	httputil.JSON(w, http.StatusOK, map[string]string{
		"message": "Peserta berhasil dihapus",
	})
}
