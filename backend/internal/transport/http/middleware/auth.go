package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/gridify/backend/internal/domain"
	"github.com/gridify/backend/internal/transport/http/httputil"
)

type userContextKey string

const UserContextKey userContextKey = "authenticated_user"
const SessionCookieName = "gridify_session"

type SessionValidator interface {
	ValidateSession(ctx context.Context, token string) (*domain.User, error)
}

func RequireAuth(validator SessionValidator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := extractToken(r)
			if token == "" {
				httputil.RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Sesi autentikasi diperlukan", nil, GetRequestID(r.Context()))
				return
			}

			user, err := validator.ValidateSession(r.Context(), token)
			if err != nil {
				httputil.RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Sesi tidak valid atau telah berakhir", nil, GetRequestID(r.Context()))
				return
			}

			ctx := context.WithValue(r.Context(), UserContextKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func OptionalAuth(validator SessionValidator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := extractToken(r)
			if token != "" {
				if user, err := validator.ValidateSession(r.Context(), token); err == nil {
					r = r.WithContext(context.WithValue(r.Context(), UserContextKey, user))
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

func extractToken(r *http.Request) string {
	// 1. Try Cookie
	if cookie, err := r.Cookie(SessionCookieName); err == nil && cookie.Value != "" {
		return cookie.Value
	}

	// 2. Try Authorization Bearer
	authHeader := r.Header.Get("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}

	return ""
}

func GetUser(ctx context.Context) *domain.User {
	if val, ok := ctx.Value(UserContextKey).(*domain.User); ok {
		return val
	}
	return nil
}

func GetUserID(ctx context.Context) uuid.UUID {
	if u := GetUser(ctx); u != nil {
		return u.ID
	}
	return uuid.Nil
}
