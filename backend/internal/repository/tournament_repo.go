package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/gridify/backend/internal/domain"
)

type TournamentRepository struct {
	pool *pgxpool.Pool
}

func NewTournamentRepository(pool *pgxpool.Pool) *TournamentRepository {
	return &TournamentRepository{pool: pool}
}

func (r *TournamentRepository) GetDashboardSummary(ctx context.Context, userID uuid.UUID) (*domain.DashboardSummary, error) {
	summary := &domain.DashboardSummary{
		RecentTournaments: []domain.Tournament{},
	}

	// 1. Total Tournaments
	err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM tournaments WHERE owner_id = $1
	`, userID).Scan(&summary.TotalTournaments)
	if err != nil {
		return nil, fmt.Errorf("failed to count total tournaments: %w", err)
	}

	// 2. Active Tournaments
	err = r.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM tournaments WHERE owner_id = $1 AND status = 'ONGOING'
	`, userID).Scan(&summary.ActiveTournaments)
	if err != nil {
		return nil, fmt.Errorf("failed to count active tournaments: %w", err)
	}

	// 3. Upcoming Matches Count (matches in user's tournaments that are SCHEDULED or ONGOING)
	err = r.pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM matches m
		JOIN tournaments t ON t.id = m.tournament_id
		WHERE t.owner_id = $1 AND m.status IN ('SCHEDULED', 'ONGOING')
	`, userID).Scan(&summary.UpcomingMatchesCount)
	if err != nil {
		return nil, fmt.Errorf("failed to count upcoming matches: %w", err)
	}

	// 4. 4 Recent Tournaments with Game
	rows, err := r.pool.Query(ctx, `
		SELECT
			t.id, t.owner_id, t.game_id, t.title, t.description, t.format_type,
			t.status, t.best_of, t.home_away, t.points_per_win, t.points_per_draw,
			t.has_third_place, t.br_total_matches, t.br_max_teams_per_match,
			t.br_point_system, t.created_at, t.updated_at,
			g.id, g.slug, g.name, g.genre, g.platform,
			(SELECT COUNT(*) FROM participants WHERE tournament_id = t.id) AS participants_count
		FROM tournaments t
		JOIN games g ON g.id = t.game_id
		WHERE t.owner_id = $1
		ORDER BY t.created_at DESC
		LIMIT 4
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query recent tournaments: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var t domain.Tournament
		var g domain.Game
		err := rows.Scan(
			&t.ID, &t.OwnerID, &t.GameID, &t.Title, &t.Description, &t.FormatType,
			&t.Status, &t.BestOf, &t.HomeAway, &t.PointsPerWin, &t.PointsPerDraw,
			&t.HasThirdPlace, &t.BRTotalMatches, &t.BRMaxTeamsPerMatch,
			&t.BRPointSystem, &t.CreatedAt, &t.UpdatedAt,
			&g.ID, &g.Slug, &g.Name, &g.Genre, &g.Platform,
			&t.ParticipantsCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan recent tournament: %w", err)
		}
		t.Game = &g
		summary.RecentTournaments = append(summary.RecentTournaments, t)
	}

	return summary, nil
}

func (r *TournamentRepository) VerifyTournamentOwnership(ctx context.Context, tournamentID, userID uuid.UUID) error {
	var ownerID uuid.UUID
	err := r.pool.QueryRow(ctx, `
		SELECT owner_id FROM tournaments WHERE id = $1
	`, tournamentID).Scan(&ownerID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrNotFound
		}
		return fmt.Errorf("failed to query tournament owner: %w", err)
	}

	if ownerID != userID {
		return domain.ErrForbidden
	}
	return nil
}

func (r *TournamentRepository) CreateTournamentWithTx(ctx context.Context, tx pgx.Tx, t *domain.Tournament) error {
	query := `
		INSERT INTO tournaments (
			id, owner_id, game_id, title, description, format_type, status,
			best_of, home_away, points_per_win, points_per_draw, has_third_place,
			br_total_matches, br_max_teams_per_match, br_point_system, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17
		)
	`
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	now := time.Now().UTC()
	t.CreatedAt = now
	t.UpdatedAt = now
	if t.Status == "" {
		t.Status = "DRAFT"
	}
	if t.BestOf == 0 {
		t.BestOf = 1
	}
	if t.PointsPerWin == 0 {
		t.PointsPerWin = 3
	}
	if t.PointsPerDraw == 0 {
		t.PointsPerDraw = 1
	}
	if t.BRTotalMatches == 0 {
		t.BRTotalMatches = 6
	}
	if t.BRMaxTeamsPerMatch == 0 {
		t.BRMaxTeamsPerMatch = 16
	}
	if t.BRPointSystem == nil {
		t.BRPointSystem = []byte(`{"killPoint": 1, "placementPoints": {"1": 10, "2": 6, "3": 5, "4": 4, "5": 3, "6": 2, "7": 1, "8": 1}}`)
	}

	_, err := tx.Exec(ctx, query,
		t.ID, t.OwnerID, t.GameID, t.Title, t.Description, t.FormatType, t.Status,
		t.BestOf, t.HomeAway, t.PointsPerWin, t.PointsPerDraw, t.HasThirdPlace,
		t.BRTotalMatches, t.BRMaxTeamsPerMatch, t.BRPointSystem, t.CreatedAt, t.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert tournament: %w", err)
	}
	return nil
}

func (r *TournamentRepository) GetTournamentByID(ctx context.Context, id uuid.UUID) (*domain.Tournament, error) {
	query := `
		SELECT
			t.id, t.owner_id, t.game_id, t.title, t.description, t.format_type,
			t.status, t.best_of, t.home_away, t.points_per_win, t.points_per_draw,
			t.has_third_place, t.br_total_matches, t.br_max_teams_per_match,
			t.br_point_system, t.created_at, t.updated_at,
			g.id, g.slug, g.name, g.genre, g.platform,
			(SELECT COUNT(*) FROM participants WHERE tournament_id = t.id) AS participants_count
		FROM tournaments t
		JOIN games g ON g.id = t.game_id
		WHERE t.id = $1
	`
	row := r.pool.QueryRow(ctx, query, id)

	var t domain.Tournament
	var g domain.Game
	err := row.Scan(
		&t.ID, &t.OwnerID, &t.GameID, &t.Title, &t.Description, &t.FormatType,
		&t.Status, &t.BestOf, &t.HomeAway, &t.PointsPerWin, &t.PointsPerDraw,
		&t.HasThirdPlace, &t.BRTotalMatches, &t.BRMaxTeamsPerMatch,
		&t.BRPointSystem, &t.CreatedAt, &t.UpdatedAt,
		&g.ID, &g.Slug, &g.Name, &g.Genre, &g.Platform,
		&t.ParticipantsCount,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to query tournament: %w", err)
	}
	t.Game = &g
	return &t, nil
}

func (r *TournamentRepository) ListTournaments(ctx context.Context, ownerID uuid.UUID, q, status string, page, perPage int) ([]domain.Tournament, int, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	var conditions []string
	var args []interface{}
	argIdx := 1

	conditions = append(conditions, fmt.Sprintf("t.owner_id = $%d", argIdx))
	args = append(args, ownerID)
	argIdx++

	if q != "" {
		conditions = append(conditions, fmt.Sprintf("t.title ILIKE $%d", argIdx))
		args = append(args, "%"+strings.TrimSpace(q)+"%")
		argIdx++
	}

	if status != "" {
		conditions = append(conditions, fmt.Sprintf("t.status = $%d", argIdx))
		args = append(args, status)
		argIdx++
	}

	whereClause := "WHERE " + strings.Join(conditions, " AND ")

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM tournaments t %s", whereClause)
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count tournaments: %w", err)
	}

	// Query page items
	itemsQuery := fmt.Sprintf(`
		SELECT
			t.id, t.owner_id, t.game_id, t.title, t.description, t.format_type,
			t.status, t.best_of, t.home_away, t.points_per_win, t.points_per_draw,
			t.has_third_place, t.br_total_matches, t.br_max_teams_per_match,
			t.br_point_system, t.created_at, t.updated_at,
			g.id, g.slug, g.name, g.genre, g.platform,
			(SELECT COUNT(*) FROM participants WHERE tournament_id = t.id) AS participants_count
		FROM tournaments t
		JOIN games g ON g.id = t.game_id
		%s
		ORDER BY t.created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIdx, argIdx+1)

	args = append(args, perPage, offset)

	rows, err := r.pool.Query(ctx, itemsQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query tournaments: %w", err)
	}
	defer rows.Close()

	var list []domain.Tournament
	for rows.Next() {
		var t domain.Tournament
		var g domain.Game
		err := rows.Scan(
			&t.ID, &t.OwnerID, &t.GameID, &t.Title, &t.Description, &t.FormatType,
			&t.Status, &t.BestOf, &t.HomeAway, &t.PointsPerWin, &t.PointsPerDraw,
			&t.HasThirdPlace, &t.BRTotalMatches, &t.BRMaxTeamsPerMatch,
			&t.BRPointSystem, &t.CreatedAt, &t.UpdatedAt,
			&g.ID, &g.Slug, &g.Name, &g.Genre, &g.Platform,
			&t.ParticipantsCount,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan tournament: %w", err)
		}
		t.Game = &g
		list = append(list, t)
	}

	return list, total, nil
}

func (r *TournamentRepository) UpdateTournament(ctx context.Context, t *domain.Tournament) error {
	query := `
		UPDATE tournaments
		SET title = $1, description = $2, status = $3, best_of = $4,
		    home_away = $5, points_per_win = $6, points_per_draw = $7,
		    has_third_place = $8, updated_at = NOW()
		WHERE id = $9 AND owner_id = $10
		RETURNING updated_at
	`
	err := r.pool.QueryRow(ctx, query,
		t.Title, t.Description, t.Status, t.BestOf,
		t.HomeAway, t.PointsPerWin, t.PointsPerDraw,
		t.HasThirdPlace, t.ID, t.OwnerID,
	).Scan(&t.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrNotFound
		}
		return fmt.Errorf("failed to update tournament: %w", err)
	}
	return nil
}

func (r *TournamentRepository) UpdateTournamentStatus(ctx context.Context, tx pgx.Tx, id uuid.UUID, status string) error {
	query := `UPDATE tournaments SET status = $1, updated_at = NOW() WHERE id = $2`
	var err error
	if tx != nil {
		_, err = tx.Exec(ctx, query, status, id)
	} else {
		_, err = r.pool.Exec(ctx, query, status, id)
	}
	return err
}

func (r *TournamentRepository) DeleteTournament(ctx context.Context, id, ownerID uuid.UUID) error {
	// First verify ownership
	if err := r.VerifyTournamentOwnership(ctx, id, ownerID); err != nil {
		return err
	}

	// Atomic cascading delete via PostgreSQL FK ON DELETE CASCADE
	query := `DELETE FROM tournaments WHERE id = $1 AND owner_id = $2`
	cmd, err := r.pool.Exec(ctx, query, id, ownerID)
	if err != nil {
		return fmt.Errorf("failed to delete tournament: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
