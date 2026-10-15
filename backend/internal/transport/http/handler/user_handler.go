package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gridify/backend/internal/domain"
	"github.com/gridify/backend/internal/service"
	"github.com/gridify/backend/internal/transport/http/httputil"
	"github.com/gridify/backend/internal/transport/http/middleware"
)

type UserHandler struct {
	authSvc *service.AuthService
}

func NewUserHandler(authSvc *service.AuthService) *UserHandler {
	return &UserHandler{authSvc: authSvc}
}

func (h *UserHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())
	userID := middleware.GetUserID(r.Context())
	if userID == [16]byte{} {
		httputil.RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Sesi tidak valid", nil, reqID)
		return
	}

	user, err := h.authSvc.GetUserByID(r.Context(), userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			httputil.RespondError(w, http.StatusNotFound, "USER_NOT_FOUND", "Pengguna tidak ditemukan", nil, reqID)
			return
		}
		httputil.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Gagal mengambil data profil", nil, reqID)
		return
	}

	httputil.JSON(w, http.StatusOK, user)
}

type UpdateProfilePayload struct {
	FirstName string  `json:"first_name"`
	LastName  string  `json:"last_name"`
	Phone     *string `json:"phone,omitempty"`
	Username  string  `json:"username"`
}

func (h *UserHandler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())
	userID := middleware.GetUserID(r.Context())
	if userID == [16]byte{} {
		httputil.RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Sesi tidak valid", nil, reqID)
		return
	}

	var p UpdateProfilePayload
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&p); err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "INVALID_PAYLOAD", "Payload JSON tidak valid", nil, reqID)
		return
	}

	var phoneStr string
	if p.Phone != nil {
		phoneStr = *p.Phone
	}

	user, err := h.authSvc.UpdateProfile(r.Context(), userID, p.FirstName, p.LastName, phoneStr, p.Username)
	if err != nil {
		if errors.Is(err, domain.ErrConflict) {
			httputil.RespondError(w, http.StatusConflict, "USERNAME_TAKEN", "Username sudah digunakan oleh akun lain", nil, reqID)
			return
		}
		if errors.Is(err, domain.ErrBadRequest) {
			httputil.RespondError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil, reqID)
			return
		}
		httputil.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Gagal memperbarui profil", nil, reqID)
		return
	}

	httputil.JSON(w, http.StatusOK, user)
}

type ChangePasswordPayload struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

func (h *UserHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())
	userID := middleware.GetUserID(r.Context())
	if userID == [16]byte{} {
		httputil.RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Sesi tidak valid", nil, reqID)
		return
	}

	var p ChangePasswordPayload
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&p); err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "INVALID_PAYLOAD", "Payload JSON tidak valid", nil, reqID)
		return
	}

	if err := h.authSvc.ChangePassword(r.Context(), userID, p.CurrentPassword, p.NewPassword); err != nil {
		if errors.Is(err, domain.ErrBadRequest) {
			httputil.RespondError(w, http.StatusBadRequest, "INVALID_PASSWORD", err.Error(), nil, reqID)
			return
		}
		httputil.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Gagal mengubah password", nil, reqID)
		return
	}

	httputil.JSON(w, http.StatusOK, map[string]string{
		"message": "Password berhasil diperbarui",
	})
}
