package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/gridify/backend/internal/config"
	"github.com/gridify/backend/internal/domain"
	"github.com/gridify/backend/internal/service"
	"github.com/gridify/backend/internal/transport/http/httputil"
	"github.com/gridify/backend/internal/transport/http/middleware"
)

type AuthHandler struct {
	authSvc *service.AuthService
	cfg     *config.Config
}

func NewAuthHandler(authSvc *service.AuthService, cfg *config.Config) *AuthHandler {
	return &AuthHandler{
		authSvc: authSvc,
		cfg:     cfg,
	}
}

type RegisterPayload struct {
	Email     string  `json:"email"`
	Username  string  `json:"username"`
	Password  string  `json:"password"`
	FirstName string  `json:"first_name"`
	LastName  string  `json:"last_name"`
	Phone     *string `json:"phone,omitempty"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())

	var p RegisterPayload
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&p); err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "INVALID_PAYLOAD", "Payload JSON tidak valid", nil, reqID)
		return
	}

	user, err := h.authSvc.Register(r.Context(), service.RegisterInput{
		Email:     p.Email,
		Username:  p.Username,
		Password:  p.Password,
		FirstName: p.FirstName,
		LastName:  p.LastName,
		Phone:     p.Phone,
	})
	if err != nil {
		if errors.Is(err, domain.ErrConflict) {
			httputil.RespondError(w, http.StatusConflict, "USER_EXISTS", "Email atau username sudah terdaftar", nil, reqID)
			return
		}
		if errors.Is(err, domain.ErrBadRequest) {
			httputil.RespondError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil, reqID)
			return
		}
		httputil.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Gagal mendaftarkan akun", nil, reqID)
		return
	}

	httputil.JSON(w, http.StatusCreated, user)
}

type LoginPayload struct {
	Identifier string `json:"identifier"`
	Password   string `json:"password"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())

	var p LoginPayload
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&p); err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "INVALID_PAYLOAD", "Payload JSON tidak valid", nil, reqID)
		return
	}

	userAgent := r.UserAgent()
	ipAddress := r.RemoteAddr
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		ipAddress = strings.Split(forwarded, ",")[0]
	}

	user, rawToken, err := h.authSvc.Login(r.Context(), p.Identifier, p.Password, userAgent, ipAddress)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			httputil.RespondError(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Email/username atau password salah", nil, reqID)
			return
		}
		httputil.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Terjadi kesalahan saat memproses login", nil, reqID)
		return
	}

	// Set HttpOnly session cookie
	isProd := h.cfg.Env == "production"
	http.SetCookie(w, &http.Cookie{
		Name:     middleware.SessionCookieName,
		Value:    rawToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   isProd,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   30 * 24 * 3600, // 30 days
	})

	httputil.JSON(w, http.StatusOK, user)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var token string
	if cookie, err := r.Cookie(middleware.SessionCookieName); err == nil {
		token = cookie.Value
	}

	_ = h.authSvc.Logout(r.Context(), token)

	isProd := h.cfg.Env == "production"
	http.SetCookie(w, &http.Cookie{
		Name:     middleware.SessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   isProd,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})

	httputil.JSON(w, http.StatusOK, map[string]string{
		"message": "Berhasil logout",
	})
}

type ForgotPasswordPayload struct {
	Email string `json:"email"`
}

func (h *AuthHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())

	var p ForgotPasswordPayload
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&p); err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "INVALID_PAYLOAD", "Payload JSON tidak valid", nil, reqID)
		return
	}

	_, _ = h.authSvc.ForgotPassword(r.Context(), p.Email)

	httputil.JSON(w, http.StatusOK, map[string]string{
		"message": "Jika email terdaftar, instruksi reset password telah dikirim",
	})
}

type ResetPasswordPayload struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())

	var p ResetPasswordPayload
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&p); err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "INVALID_PAYLOAD", "Payload JSON tidak valid", nil, reqID)
		return
	}

	if err := h.authSvc.ResetPassword(r.Context(), p.Token, p.NewPassword); err != nil {
		if errors.Is(err, domain.ErrBadRequest) {
			httputil.RespondError(w, http.StatusBadRequest, "INVALID_TOKEN", err.Error(), nil, reqID)
			return
		}
		httputil.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Gagal mereset password", nil, reqID)
		return
	}

	httputil.JSON(w, http.StatusOK, map[string]string{
		"message": "Password berhasil diubah. Silakan login kembali.",
	})
}
