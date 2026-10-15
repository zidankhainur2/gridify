package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"net/mail"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/argon2"

	"github.com/gridify/backend/internal/domain"
	"github.com/gridify/backend/internal/repository"
)

var usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_.-]{3,50}$`)

type AuthService struct {
	userRepo *repository.UserRepository
}

func NewAuthService(userRepo *repository.UserRepository) *AuthService {
	return &AuthService{userRepo: userRepo}
}

// ----------------------------------------------------------------------------
// Argon2id Password Hashing & Verification
// ----------------------------------------------------------------------------

const (
	argonMemory      = 64 * 1024 // 64 MB
	argonIterations  = 3
	argonParallelism = 2
	argonSaltLength  = 16
	argonKeyLength   = 32
)

func HashPassword(password string) (string, error) {
	salt := make([]byte, argonSaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate random salt: %w", err)
	}

	hash := argon2.IDKey([]byte(password), salt, argonIterations, argonMemory, argonParallelism, argonKeyLength)

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	encoded := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, argonMemory, argonIterations, argonParallelism, b64Salt, b64Hash)

	return encoded, nil
}

func VerifyPassword(password, encodedHash string) (bool, error) {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 || parts[1] != "argon2id" {
		return false, errors.New("invalid argon2id hash format")
	}

	var memory uint32
	var iterations uint32
	var parallelism uint8
	_, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &iterations, &parallelism)
	if err != nil {
		return false, fmt.Errorf("failed to parse argon2id params: %w", err)
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, fmt.Errorf("failed to decode salt: %w", err)
	}

	expectedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, fmt.Errorf("failed to decode hash: %w", err)
	}

	computedHash := argon2.IDKey([]byte(password), salt, iterations, memory, parallelism, uint32(len(expectedHash)))

	if subtle.ConstantTimeCompare(computedHash, expectedHash) == 1 {
		return true, nil
	}
	return false, nil
}

// ----------------------------------------------------------------------------
// Token & Session Helpers
// ----------------------------------------------------------------------------

func GenerateSecureToken() (string, string, error) {
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", "", fmt.Errorf("failed to generate random token: %w", err)
	}
	rawToken := hex.EncodeToString(tokenBytes)
	hash := HashToken(rawToken)
	return rawToken, hash, nil
}

func HashToken(rawToken string) string {
	sum := sha256.Sum256([]byte(rawToken))
	return hex.EncodeToString(sum[:])
}

// ----------------------------------------------------------------------------
// Authentication Business Logic
// ----------------------------------------------------------------------------

type RegisterInput struct {
	Email     string
	Username  string
	Password  string
	FirstName string
	LastName  string
	Phone     *string
}

func (s *AuthService) Register(ctx context.Context, input RegisterInput) (*domain.User, error) {
	cleanEmail := strings.TrimSpace(strings.ToLower(input.Email))
	if _, err := mail.ParseAddress(cleanEmail); err != nil {
		return nil, fmt.Errorf("%w: invalid email address", domain.ErrBadRequest)
	}

	cleanUsername := strings.TrimSpace(input.Username)
	if !usernameRegex.MatchString(cleanUsername) {
		return nil, fmt.Errorf("%w: username must be 3-50 alphanumeric characters", domain.ErrBadRequest)
	}

	if len(input.Password) < 8 {
		return nil, fmt.Errorf("%w: password must be at least 8 characters", domain.ErrBadRequest)
	}

	cleanFirstName := strings.TrimSpace(input.FirstName)
	if len(cleanFirstName) < 1 {
		return nil, fmt.Errorf("%w: first name cannot be empty", domain.ErrBadRequest)
	}

	passwordHash, err := HashPassword(input.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &domain.User{
		ID:           uuid.New(),
		Email:        cleanEmail,
		Username:     cleanUsername,
		PasswordHash: passwordHash,
		FirstName:    cleanFirstName,
		LastName:     strings.TrimSpace(input.LastName),
		Phone:        input.Phone,
	}

	if err := s.userRepo.CreateUser(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *AuthService) Login(ctx context.Context, identifier, password, userAgent, ipAddress string) (*domain.User, string, error) {
	user, err := s.userRepo.GetUserByEmailOrUsername(ctx, identifier)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, "", domain.ErrInvalidCredentials
		}
		return nil, "", err
	}

	match, err := VerifyPassword(password, user.PasswordHash)
	if err != nil || !match {
		return nil, "", domain.ErrInvalidCredentials
	}

	rawToken, tokenHash, err := GenerateSecureToken()
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate session token: %w", err)
	}

	session := &domain.UserSession{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: tokenHash,
		UserAgent: &userAgent,
		IPAddress: &ipAddress,
		ExpiresAt: time.Now().UTC().Add(30 * 24 * time.Hour), // 30-day session
	}

	if err := s.userRepo.CreateSession(ctx, session); err != nil {
		return nil, "", fmt.Errorf("failed to create session: %w", err)
	}

	return user, rawToken, nil
}

func (s *AuthService) Logout(ctx context.Context, rawToken string) error {
	if rawToken == "" {
		return nil
	}
	tokenHash := HashToken(rawToken)
	return s.userRepo.DeleteSession(ctx, tokenHash)
}

func (s *AuthService) ValidateSession(ctx context.Context, rawToken string) (*domain.User, error) {
	if rawToken == "" {
		return nil, domain.ErrUnauthorized
	}
	tokenHash := HashToken(rawToken)
	_, user, err := s.userRepo.GetSessionAndUserByTokenHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.ErrUnauthorized
		}
		return nil, err
	}
	return user, nil
}

func (s *AuthService) ForgotPassword(ctx context.Context, email string) (string, error) {
	cleanEmail := strings.TrimSpace(strings.ToLower(email))
	user, err := s.userRepo.GetUserByEmailOrUsername(ctx, cleanEmail)
	if err != nil {
		// Silent success if user not found to prevent email enumeration
		return "", nil
	}

	rawToken, tokenHash, err := GenerateSecureToken()
	if err != nil {
		return "", err
	}

	resetToken := &domain.PasswordResetToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().UTC().Add(30 * time.Minute), // 30 mins
	}

	if err := s.userRepo.CreatePasswordResetToken(ctx, resetToken); err != nil {
		return "", err
	}

	return rawToken, nil
}

func (s *AuthService) ResetPassword(ctx context.Context, rawToken, newPassword string) error {
	if len(newPassword) < 8 {
		return fmt.Errorf("%w: password must be at least 8 characters", domain.ErrBadRequest)
	}

	tokenHash := HashToken(rawToken)
	tokenRecord, err := s.userRepo.GetValidPasswordResetToken(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return fmt.Errorf("%w: token reset tidak valid atau sudah kadaluarsa", domain.ErrBadRequest)
		}
		return err
	}

	newHash, err := HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	if err := s.userRepo.UpdatePassword(ctx, tokenRecord.UserID, newHash); err != nil {
		return err
	}

	_ = s.userRepo.MarkPasswordResetTokenUsed(ctx, tokenRecord.ID)
	// Revoke all existing sessions for security
	_ = s.userRepo.DeleteAllUserSessions(ctx, tokenRecord.UserID)

	return nil
}

func (s *AuthService) ChangePassword(ctx context.Context, userID uuid.UUID, currentPassword, newPassword string) error {
	if len(newPassword) < 8 {
		return fmt.Errorf("%w: password baru minimal 8 karakter", domain.ErrBadRequest)
	}

	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	match, err := VerifyPassword(currentPassword, user.PasswordHash)
	if err != nil || !match {
		return fmt.Errorf("%w: password saat ini tidak sesuai", domain.ErrBadRequest)
	}

	newHash, err := HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	if err := s.userRepo.UpdatePassword(ctx, userID, newHash); err != nil {
		return err
	}

	// Revoke old sessions
	_ = s.userRepo.DeleteAllUserSessions(ctx, userID)

	return nil
}

func (s *AuthService) UpdateProfile(ctx context.Context, userID uuid.UUID, firstName, lastName, phone, username string) (*domain.User, error) {
	cleanFirstName := strings.TrimSpace(firstName)
	if len(cleanFirstName) < 1 {
		return nil, fmt.Errorf("%w: nama depan tidak boleh kosong", domain.ErrBadRequest)
	}

	cleanUsername := strings.TrimSpace(username)
	if !usernameRegex.MatchString(cleanUsername) {
		return nil, fmt.Errorf("%w: username harus 3-50 karakter alfanumerik", domain.ErrBadRequest)
	}

	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	user.FirstName = cleanFirstName
	user.LastName = strings.TrimSpace(lastName)
	if phone != "" {
		p := strings.TrimSpace(phone)
		user.Phone = &p
	} else {
		user.Phone = nil
	}
	user.Username = cleanUsername

	if err := s.userRepo.UpdateUser(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *AuthService) GetUserByID(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	return s.userRepo.GetUserByID(ctx, userID)
}
