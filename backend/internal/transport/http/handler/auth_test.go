package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/gridify/backend/internal/config"
	"github.com/gridify/backend/internal/database"
	transporthttp "github.com/gridify/backend/internal/transport/http"
	"github.com/gridify/backend/internal/transport/http/httputil"
	"github.com/gridify/backend/internal/transport/http/middleware"
)

func setupTestRouter(t *testing.T) (*config.Config, http.Handler) {
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := database.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}
	t.Cleanup(func() {
		pool.Close()
	})

	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	router := transporthttp.NewRouter(cfg, pool, logger)
	return cfg, router
}

func TestAuthAndUserProfileFlow(t *testing.T) {
	_, router := setupTestRouter(t)

	uniqueID := uuid.New().String()[:8]
	testEmail := fmt.Sprintf("organizer_%s@gridify.test", uniqueID)
	testUsername := fmt.Sprintf("org_%s", uniqueID)
	testPassword := "SecurePassword123!"

	// 1. Register User
	t.Run("Register new user", func(t *testing.T) {
		body, _ := json.Marshal(map[string]interface{}{
			"email":      testEmail,
			"username":   testUsername,
			"password":   testPassword,
			"first_name": "Gridify",
			"last_name":  "Organizer",
			"phone":      "08123456789",
		})

		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("expected status 201, got %d: %s", rec.Code, rec.Body.String())
		}
	})

	// 2. Reject Duplicate Email / Username
	t.Run("Reject duplicate register with 409 Conflict", func(t *testing.T) {
		body, _ := json.Marshal(map[string]interface{}{
			"email":      testEmail,
			"username":   testUsername,
			"password":   testPassword,
			"first_name": "Duplicate",
		})

		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusConflict {
			t.Fatalf("expected status 409, got %d: %s", rec.Code, rec.Body.String())
		}
	})

	// 3. Login with Invalid Credentials
	t.Run("Login with invalid password returns 401 generic message", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{
			"identifier": testEmail,
			"password":   "WrongPassword!",
		})

		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected status 401, got %d", rec.Code)
		}
	})

	// 4. Login with Username (Valid)
	var sessionCookie *http.Cookie
	t.Run("Login with username succeeds and sets HttpOnly cookie", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{
			"identifier": testUsername,
			"password":   testPassword,
		})

		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
		}

		cookies := rec.Result().Cookies()
		for _, c := range cookies {
			if c.Name == middleware.SessionCookieName {
				sessionCookie = c
				break
			}
		}

		if sessionCookie == nil {
			t.Fatal("expected gridify_session cookie in response")
		}

		if !sessionCookie.HttpOnly {
			t.Errorf("expected session cookie to be HttpOnly")
		}
	})

	// 5. Access /api/v1/users/me Without Session
	t.Run("Access /api/v1/users/me without session returns 401", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/users/me", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected status 401, got %d", rec.Code)
		}
	})

	// 6. Access /api/v1/users/me With Session Cookie
	t.Run("Access /api/v1/users/me with session cookie returns profile", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/users/me", nil)
		req.AddCookie(sessionCookie)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
		}

		var resp httputil.Envelope
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to unmarshal JSON: %v", err)
		}

		dataMap := resp.Data.(map[string]interface{})
		if !strings.EqualFold(dataMap["username"].(string), testUsername) {
			t.Errorf("expected username %s, got %v", testUsername, dataMap["username"])
		}
	})

	// 7. Update Profile
	newUsername := fmt.Sprintf("renamed_%s", uniqueID)
	t.Run("Update profile succeeds", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{
			"first_name": "UpdatedName",
			"last_name":  "UpdatedLast",
			"phone":      "08999999999",
			"username":   newUsername,
		})

		req := httptest.NewRequest(http.MethodPatch, "/api/v1/users/me", bytes.NewReader(body))
		req.AddCookie(sessionCookie)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
		}
	})

	// 8. Change Password (Success)
	newPassword := "BrandNewPassword123!"
	t.Run("Change password with valid current password succeeds", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{
			"current_password": testPassword,
			"new_password":     newPassword,
		})

		req := httptest.NewRequest(http.MethodPost, "/api/v1/users/me/password", bytes.NewReader(body))
		req.AddCookie(sessionCookie)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
		}
	})

	// 9. Old Session Revoked After Password Change
	t.Run("Old session revoked after password change", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/users/me", nil)
		req.AddCookie(sessionCookie)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected status 401 after password change, got %d", rec.Code)
		}
	})

	// 10. Login With New Password
	t.Run("Login with new password succeeds", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{
			"identifier": newUsername,
			"password":   newPassword,
		})

		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
		}
	})
}
