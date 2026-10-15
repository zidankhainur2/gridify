package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/gridify/backend/internal/domain"
)

type GameRepository struct {
	pool *pgxpool.Pool
}

func NewGameRepository(pool *pgxpool.Pool) *GameRepository {
	return &GameRepository{pool: pool}
}

func (r *GameRepository) GetAllGames(ctx context.Context) ([]domain.Game, error) {
	query := `
		SELECT id, slug, name, genre, platform, created_at
		FROM games
		ORDER BY genre, name
	`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query games: %w", err)
	}
	defer rows.Close()

	var games []domain.Game
	for rows.Next() {
		var g domain.Game
		if err := rows.Scan(&g.ID, &g.Slug, &g.Name, &g.Genre, &g.Platform, &g.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan game: %w", err)
		}
		games = append(games, g)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error reading games rows: %w", err)
	}
	return games, nil
}

func (r *GameRepository) GetGameByID(ctx context.Context, id uuid.UUID) (*domain.Game, error) {
	query := `
		SELECT id, slug, name, genre, platform, created_at
		FROM games
		WHERE id = $1
	`
	row := r.pool.QueryRow(ctx, query, id)

	var g domain.Game
	if err := row.Scan(&g.ID, &g.Slug, &g.Name, &g.Genre, &g.Platform, &g.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to query game by id: %w", err)
	}
	return &g, nil
}
