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

type StageRepository struct {
	pool *pgxpool.Pool
}

func NewStageRepository(pool *pgxpool.Pool) *StageRepository {
	return &StageRepository{pool: pool}
}

func (r *StageRepository) CreateStageWithTx(ctx context.Context, tx pgx.Tx, s *domain.Stage) error {
	query := `
		INSERT INTO stages (id, tournament_id, name, type, sequence_order, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	if s.Status == "" {
		s.Status = "PENDING"
	}
	s.CreatedAt = time.Now().UTC()

	_, err := tx.Exec(ctx, query,
		s.ID, s.TournamentID, s.Name, s.Type, s.SequenceOrder, s.Status, s.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert stage: %w", err)
	}
	return nil
}

func (r *StageRepository) GetStagesByTournamentID(ctx context.Context, tournamentID uuid.UUID) ([]domain.Stage, error) {
	query := `
		SELECT id, tournament_id, name, type, sequence_order, status, created_at
		FROM stages
		WHERE tournament_id = $1
		ORDER BY sequence_order ASC
	`
	rows, err := r.pool.Query(ctx, query, tournamentID)
	if err != nil {
		return nil, fmt.Errorf("failed to query stages: %w", err)
	}
	defer rows.Close()

	var stages []domain.Stage
	for rows.Next() {
		var s domain.Stage
		if err := rows.Scan(&s.ID, &s.TournamentID, &s.Name, &s.Type, &s.SequenceOrder, &s.Status, &s.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan stage: %w", err)
		}
		stages = append(stages, s)
	}
	return stages, nil
}

func (r *StageRepository) GetStageByID(ctx context.Context, id uuid.UUID) (*domain.Stage, error) {
	query := `
		SELECT id, tournament_id, name, type, sequence_order, status, created_at
		FROM stages
		WHERE id = $1
	`
	row := r.pool.QueryRow(ctx, query, id)

	var s domain.Stage
	if err := row.Scan(&s.ID, &s.TournamentID, &s.Name, &s.Type, &s.SequenceOrder, &s.Status, &s.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to query stage by id: %w", err)
	}
	return &s, nil
}

func (r *StageRepository) UpdateStageStatus(ctx context.Context, tx pgx.Tx, id uuid.UUID, status string) error {
	query := `UPDATE stages SET status = $1 WHERE id = $2`
	var err error
	if tx != nil {
		_, err = tx.Exec(ctx, query, status, id)
	} else {
		_, err = r.pool.Exec(ctx, query, status, id)
	}
	return err
}

func (r *StageRepository) ListStagesByTournamentID(ctx context.Context, tournamentID uuid.UUID) ([]domain.Stage, error) {
	return r.GetStagesByTournamentID(ctx, tournamentID)
}
