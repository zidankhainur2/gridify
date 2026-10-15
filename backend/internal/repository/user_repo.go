package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/gridify/backend/internal/domain"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) CreateUser(ctx context.Context, u *domain.User) error {
	query := `
		INSERT INTO users (id, email, username, password_hash, first_name, last_name, phone, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	now := time.Now().UTC()
	u.CreatedAt = now
	u.UpdatedAt = now

	_, err := r.pool.Exec(ctx, query,
		u.ID,
		u.Email,
		u.Username,
		u.PasswordHash,
		u.FirstName,
		u.LastName,
		u.Phone,
		u.CreatedAt,
		u.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			// Unique violation
			return domain.ErrConflict
		}
		return fmt.Errorf("failed to insert user: %w", err)
	}
	return nil
}

func (r *UserRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	query := `
		SELECT id, email, username, password_hash, first_name, last_name, phone, created_at, updated_at
		FROM users
		WHERE id = $1
	`
	row := r.pool.QueryRow(ctx, query, id)

	var u domain.User
	err := row.Scan(
		&u.ID,
		&u.Email,
		&u.Username,
		&u.PasswordHash,
		&u.FirstName,
		&u.LastName,
		&u.Phone,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to query user by id: %w", err)
	}
	return &u, nil
}

func (r *UserRepository) GetUserByEmailOrUsername(ctx context.Context, identifier string) (*domain.User, error) {
	query := `
		SELECT id, email, username, password_hash, first_name, last_name, phone, created_at, updated_at
		FROM users
		WHERE LOWER(email) = LOWER($1) OR LOWER(username) = LOWER($1)
		LIMIT 1
	`
	row := r.pool.QueryRow(ctx, query, strings.TrimSpace(identifier))

	var u domain.User
	err := row.Scan(
		&u.ID,
		&u.Email,
		&u.Username,
		&u.PasswordHash,
		&u.FirstName,
		&u.LastName,
		&u.Phone,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to query user by email/username: %w", err)
	}
	return &u, nil
}

func (r *UserRepository) UpdateUser(ctx context.Context, u *domain.User) error {
	query := `
		UPDATE users
		SET first_name = $1, last_name = $2, phone = $3, username = $4, updated_at = NOW()
		WHERE id = $5
		RETURNING updated_at
	`
	err := r.pool.QueryRow(ctx, query,
		u.FirstName,
		u.LastName,
		u.Phone,
		u.Username,
		u.ID,
	).Scan(&u.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrConflict
		}
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrNotFound
		}
		return fmt.Errorf("failed to update user: %w", err)
	}
	return nil
}

func (r *UserRepository) UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error {
	query := `
		UPDATE users
		SET password_hash = $1, updated_at = NOW()
		WHERE id = $2
	`
	cmd, err := r.pool.Exec(ctx, query, passwordHash, userID)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *UserRepository) CreateSession(ctx context.Context, s *domain.UserSession) error {
	query := `
		INSERT INTO user_sessions (id, user_id, token_hash, user_agent, ip_address, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	now := time.Now().UTC()
	s.CreatedAt = now

	_, err := r.pool.Exec(ctx, query,
		s.ID,
		s.UserID,
		s.TokenHash,
		s.UserAgent,
		s.IPAddress,
		s.ExpiresAt,
		s.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert session: %w", err)
	}
	return nil
}

func (r *UserRepository) GetSessionAndUserByTokenHash(ctx context.Context, tokenHash string) (*domain.UserSession, *domain.User, error) {
	query := `
		SELECT
			s.id, s.user_id, s.token_hash, s.user_agent, s.ip_address, s.expires_at, s.created_at,
			u.id, u.email, u.username, u.password_hash, u.first_name, u.last_name, u.phone, u.created_at, u.updated_at
		FROM user_sessions s
		JOIN users u ON u.id = s.user_id
		WHERE s.token_hash = $1 AND s.expires_at > NOW()
		LIMIT 1
	`
	row := r.pool.QueryRow(ctx, query, tokenHash)

	var s domain.UserSession
	var u domain.User

	err := row.Scan(
		&s.ID, &s.UserID, &s.TokenHash, &s.UserAgent, &s.IPAddress, &s.ExpiresAt, &s.CreatedAt,
		&u.ID, &u.Email, &u.Username, &u.PasswordHash, &u.FirstName, &u.LastName, &u.Phone, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil, domain.ErrNotFound
		}
		return nil, nil, fmt.Errorf("failed to query session and user: %w", err)
	}
	return &s, &u, nil
}

func (r *UserRepository) DeleteSession(ctx context.Context, tokenHash string) error {
	query := `DELETE FROM user_sessions WHERE token_hash = $1`
	_, err := r.pool.Exec(ctx, query, tokenHash)
	return err
}

func (r *UserRepository) DeleteAllUserSessions(ctx context.Context, userID uuid.UUID) error {
	query := `DELETE FROM user_sessions WHERE user_id = $1`
	_, err := r.pool.Exec(ctx, query, userID)
	return err
}

func (r *UserRepository) CreatePasswordResetToken(ctx context.Context, t *domain.PasswordResetToken) error {
	query := `
		INSERT INTO password_reset_tokens (id, user_id, token_hash, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	now := time.Now().UTC()
	t.CreatedAt = now

	_, err := r.pool.Exec(ctx, query,
		t.ID,
		t.UserID,
		t.TokenHash,
		t.ExpiresAt,
		t.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert password reset token: %w", err)
	}
	return nil
}

func (r *UserRepository) GetValidPasswordResetToken(ctx context.Context, tokenHash string) (*domain.PasswordResetToken, error) {
	query := `
		SELECT id, user_id, token_hash, expires_at, used_at, created_at
		FROM password_reset_tokens
		WHERE token_hash = $1 AND used_at IS NULL AND expires_at > NOW()
		LIMIT 1
	`
	row := r.pool.QueryRow(ctx, query, tokenHash)

	var t domain.PasswordResetToken
	err := row.Scan(
		&t.ID,
		&t.UserID,
		&t.TokenHash,
		&t.ExpiresAt,
		&t.UsedAt,
		&t.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to query password reset token: %w", err)
	}
	return &t, nil
}

func (r *UserRepository) MarkPasswordResetTokenUsed(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE password_reset_tokens SET used_at = NOW() WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}
