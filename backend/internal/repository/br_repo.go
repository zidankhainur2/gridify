package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/gridify/backend/internal/domain"
)

type BattleRoyaleRepository struct {
	pool *pgxpool.Pool
}

func NewBattleRoyaleRepository(pool *pgxpool.Pool) *BattleRoyaleRepository {
	return &BattleRoyaleRepository{pool: pool}
}

func (r *BattleRoyaleRepository) SaveMatchResults(
	ctx context.Context,
	tx pgx.Tx,
	matchID uuid.UUID,
	results []domain.BattleRoyaleMatchResult,
) error {
	deleteQuery := `DELETE FROM battle_royale_match_results WHERE match_id = $1`
	var err error
	if tx != nil {
		_, err = tx.Exec(ctx, deleteQuery, matchID)
	} else {
		_, err = r.pool.Exec(ctx, deleteQuery, matchID)
	}
	if err != nil {
		return fmt.Errorf("failed to clear previous BR results: %w", err)
	}

	if len(results) == 0 {
		return nil
	}

	insertQuery := `
		INSERT INTO battle_royale_match_results (
			id, match_id, participant_id, placement_rank, kills,
			placement_points, kill_points, total_points, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
		)
	`
	now := time.Now()
	for _, res := range results {
		id := res.ID
		if id == uuid.Nil {
			id = uuid.New()
		}
		totalPts := res.PlacementPoints + res.KillPoints

		if tx != nil {
			_, err = tx.Exec(ctx, insertQuery,
				id, matchID, res.ParticipantID, res.PlacementRank, res.Kills,
				res.PlacementPoints, res.KillPoints, totalPts, now, now,
			)
		} else {
			_, err = r.pool.Exec(ctx, insertQuery,
				id, matchID, res.ParticipantID, res.PlacementRank, res.Kills,
				res.PlacementPoints, res.KillPoints, totalPts, now, now,
			)
		}
		if err != nil {
			return fmt.Errorf("failed to insert BR match result for participant %s: %w", res.ParticipantID, err)
		}
	}

	return nil
}

func (r *BattleRoyaleRepository) GetResultsByMatchID(ctx context.Context, matchID uuid.UUID) ([]domain.BattleRoyaleMatchResult, error) {
	query := `
		SELECT
			r.id, r.match_id, r.participant_id, r.placement_rank, r.kills,
			r.placement_points, r.kill_points, r.total_points, r.created_at, r.updated_at,
			p.id, p.tournament_id, p.name, p.seed_number, p.group_name
		FROM battle_royale_match_results r
		JOIN participants p ON p.id = r.participant_id
		WHERE r.match_id = $1
		ORDER BY r.placement_rank ASC
	`
	rows, err := r.pool.Query(ctx, query, matchID)
	if err != nil {
		return nil, fmt.Errorf("failed to query BR results: %w", err)
	}
	defer rows.Close()

	var list []domain.BattleRoyaleMatchResult
	for rows.Next() {
		var res domain.BattleRoyaleMatchResult
		var p domain.Participant
		err := rows.Scan(
			&res.ID, &res.MatchID, &res.ParticipantID, &res.PlacementRank, &res.Kills,
			&res.PlacementPoints, &res.KillPoints, &res.TotalPoints, &res.CreatedAt, &res.UpdatedAt,
			&p.ID, &p.TournamentID, &p.Name, &p.SeedNumber, &p.GroupName,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan BR result: %w", err)
		}
		res.Participant = &p
		list = append(list, res)
	}
	return list, nil
}

func (r *BattleRoyaleRepository) GetResultsByStageID(ctx context.Context, stageID uuid.UUID) ([]domain.BattleRoyaleMatchResult, error) {
	query := `
		SELECT
			r.id, r.match_id, r.participant_id, r.placement_rank, r.kills,
			r.placement_points, r.kill_points, r.total_points, r.created_at, r.updated_at,
			p.id, p.tournament_id, p.name, p.seed_number, p.group_name
		FROM battle_royale_match_results r
		JOIN matches m ON m.id = r.match_id
		JOIN participants p ON p.id = r.participant_id
		WHERE m.stage_id = $1
		ORDER BY m.round_number ASC, r.placement_rank ASC
	`
	rows, err := r.pool.Query(ctx, query, stageID)
	if err != nil {
		return nil, fmt.Errorf("failed to query stage BR results: %w", err)
	}
	defer rows.Close()

	var list []domain.BattleRoyaleMatchResult
	for rows.Next() {
		var res domain.BattleRoyaleMatchResult
		var p domain.Participant
		err := rows.Scan(
			&res.ID, &res.MatchID, &res.ParticipantID, &res.PlacementRank, &res.Kills,
			&res.PlacementPoints, &res.KillPoints, &res.TotalPoints, &res.CreatedAt, &res.UpdatedAt,
			&p.ID, &p.TournamentID, &p.Name, &p.SeedNumber, &p.GroupName,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan BR result: %w", err)
		}
		res.Participant = &p
		list = append(list, res)
	}
	return list, nil
}
