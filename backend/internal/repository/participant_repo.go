package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/gridify/backend/internal/domain"
)

type ParticipantRepository struct {
	pool *pgxpool.Pool
}

func NewParticipantRepository(pool *pgxpool.Pool) *ParticipantRepository {
	return &ParticipantRepository{pool: pool}
}

func (r *ParticipantRepository) CreateParticipant(ctx context.Context, p *domain.Participant) error {
	query := `
		INSERT INTO participants (id, tournament_id, name, contact_info, seed_number, group_name, is_verified, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	now := time.Now().UTC()
	p.CreatedAt = now
	p.UpdatedAt = now

	_, err := r.pool.Exec(ctx, query,
		p.ID, p.TournamentID, p.Name, p.ContactInfo, p.SeedNumber, p.GroupName, p.IsVerified, p.CreatedAt, p.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert participant: %w", err)
	}
	return nil
}

func (r *ParticipantRepository) CreateParticipantWithTx(ctx context.Context, tx pgx.Tx, p *domain.Participant) error {
	query := `
		INSERT INTO participants (id, tournament_id, name, contact_info, seed_number, group_name, is_verified, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	now := time.Now().UTC()
	p.CreatedAt = now
	p.UpdatedAt = now

	_, err := tx.Exec(ctx, query,
		p.ID, p.TournamentID, p.Name, p.ContactInfo, p.SeedNumber, p.GroupName, p.IsVerified, p.CreatedAt, p.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert participant: %w", err)
	}
	return nil
}

func (r *ParticipantRepository) GetParticipantByID(ctx context.Context, id uuid.UUID) (*domain.Participant, error) {
	query := `
		SELECT id, tournament_id, name, contact_info, seed_number, group_name, is_verified, created_at, updated_at
		FROM participants
		WHERE id = $1
	`
	row := r.pool.QueryRow(ctx, query, id)

	var p domain.Participant
	err := row.Scan(
		&p.ID, &p.TournamentID, &p.Name, &p.ContactInfo, &p.SeedNumber, &p.GroupName, &p.IsVerified, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to query participant: %w", err)
	}
	return &p, nil
}

func (r *ParticipantRepository) ListParticipantsByTournamentID(ctx context.Context, tournamentID uuid.UUID) ([]domain.Participant, error) {
	query := `
		SELECT id, tournament_id, name, contact_info, seed_number, group_name, is_verified, created_at, updated_at
		FROM participants
		WHERE tournament_id = $1
		ORDER BY seed_number ASC, created_at ASC
	`
	rows, err := r.pool.Query(ctx, query, tournamentID)
	if err != nil {
		return nil, fmt.Errorf("failed to query participants: %w", err)
	}
	defer rows.Close()

	var list []domain.Participant
	for rows.Next() {
		var p domain.Participant
		err := rows.Scan(
			&p.ID, &p.TournamentID, &p.Name, &p.ContactInfo, &p.SeedNumber, &p.GroupName, &p.IsVerified, &p.CreatedAt, &p.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan participant: %w", err)
		}
		list = append(list, p)
	}
	return list, nil
}

func (r *ParticipantRepository) UpdateParticipant(ctx context.Context, p *domain.Participant) error {
	query := `
		UPDATE participants
		SET name = $1, contact_info = $2, seed_number = $3, group_name = $4, is_verified = $5, updated_at = NOW()
		WHERE id = $6 AND tournament_id = $7
		RETURNING updated_at
	`
	err := r.pool.QueryRow(ctx, query,
		p.Name, p.ContactInfo, p.SeedNumber, p.GroupName, p.IsVerified, p.ID, p.TournamentID,
	).Scan(&p.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrNotFound
		}
		return fmt.Errorf("failed to update participant: %w", err)
	}
	return nil
}

func (r *ParticipantRepository) DeleteParticipant(ctx context.Context, id, tournamentID uuid.UUID) error {
	query := `DELETE FROM participants WHERE id = $1 AND tournament_id = $2`
	cmd, err := r.pool.Exec(ctx, query, id, tournamentID)
	if err != nil {
		return fmt.Errorf("failed to delete participant: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *ParticipantRepository) GetNextSeedNumber(ctx context.Context, tournamentID uuid.UUID) (int, error) {
	var maxSeed *int
	query := `SELECT MAX(seed_number) FROM participants WHERE tournament_id = $1`
	err := r.pool.QueryRow(ctx, query, tournamentID).Scan(&maxSeed)
	if err != nil {
		return 1, fmt.Errorf("failed to query max seed: %w", err)
	}
	if maxSeed == nil {
		return 1, nil
	}
	return *maxSeed + 1, nil
}

func (r *ParticipantRepository) CountParticipants(ctx context.Context, tournamentID uuid.UUID) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM participants WHERE tournament_id = $1", tournamentID).Scan(&count)
	return count, err
}
