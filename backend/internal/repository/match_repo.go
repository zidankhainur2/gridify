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

type MatchRepository struct {
	pool *pgxpool.Pool
}

func NewMatchRepository(pool *pgxpool.Pool) *MatchRepository {
	return &MatchRepository{pool: pool}
}

func (r *MatchRepository) DeleteMatchesByStageID(ctx context.Context, tx pgx.Tx, stageID uuid.UUID) error {
	query := `DELETE FROM matches WHERE stage_id = $1`
	var err error
	if tx != nil {
		_, err = tx.Exec(ctx, query, stageID)
	} else {
		_, err = r.pool.Exec(ctx, query, stageID)
	}
	if err != nil {
		return fmt.Errorf("failed to delete matches by stage_id: %w", err)
	}
	return nil
}

func (r *MatchRepository) BatchInsertMatches(ctx context.Context, tx pgx.Tx, matches []domain.Match) error {
	if len(matches) == 0 {
		return nil
	}

	insertQuery := `
		INSERT INTO matches (
			id, tournament_id, stage_id, bracket_type, round_number, match_number,
			group_label, participant_a_id, participant_b_id, score_a, score_b,
			winner_id, is_bye, status, next_match_id, next_match_slot,
			loser_next_match_id, loser_next_match_slot, playing_groups, map_name,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22
		)
	`
	// Pass 1: Insert all match records with next_match_id and loser_next_match_id set to NULL
	for _, m := range matches {
		groups := m.PlayingGroups
		if groups == nil {
			groups = []string{}
		}

		var err error
		if tx != nil {
			_, err = tx.Exec(ctx, insertQuery,
				m.ID, m.TournamentID, m.StageID, m.BracketType, m.RoundNumber, m.MatchNumber,
				m.GroupLabel, m.ParticipantAID, m.ParticipantBID, m.ScoreA, m.ScoreB,
				m.WinnerID, m.IsBye, m.Status, nil, m.NextMatchSlot,
				nil, m.LoserNextMatchSlot, groups, m.MapName,
				m.CreatedAt, m.UpdatedAt,
			)
		} else {
			_, err = r.pool.Exec(ctx, insertQuery,
				m.ID, m.TournamentID, m.StageID, m.BracketType, m.RoundNumber, m.MatchNumber,
				m.GroupLabel, m.ParticipantAID, m.ParticipantBID, m.ScoreA, m.ScoreB,
				m.WinnerID, m.IsBye, m.Status, nil, m.NextMatchSlot,
				nil, m.LoserNextMatchSlot, groups, m.MapName,
				m.CreatedAt, m.UpdatedAt,
			)
		}
		if err != nil {
			return fmt.Errorf("failed to insert match #%d R%d (%s): %w", m.MatchNumber, m.RoundNumber, m.BracketType, err)
		}
	}

	// Pass 2: Link self-referencing next_match_id and loser_next_match_id
	linkQuery := `
		UPDATE matches
		SET next_match_id = $1, next_match_slot = $2,
		    loser_next_match_id = $3, loser_next_match_slot = $4
		WHERE id = $5
	`
	for _, m := range matches {
		if m.NextMatchID == nil && m.LoserNextMatchID == nil {
			continue
		}

		var err error
		if tx != nil {
			_, err = tx.Exec(ctx, linkQuery,
				m.NextMatchID, m.NextMatchSlot,
				m.LoserNextMatchID, m.LoserNextMatchSlot,
				m.ID,
			)
		} else {
			_, err = r.pool.Exec(ctx, linkQuery,
				m.NextMatchID, m.NextMatchSlot,
				m.LoserNextMatchID, m.LoserNextMatchSlot,
				m.ID,
			)
		}
		if err != nil {
			return fmt.Errorf("failed to link match #%d R%d (%s): %w", m.MatchNumber, m.RoundNumber, m.BracketType, err)
		}
	}

	return nil
}

func (r *MatchRepository) GetMatchesByStageID(ctx context.Context, stageID uuid.UUID) ([]domain.Match, error) {
	query := `
		SELECT
			m.id, m.tournament_id, m.stage_id, m.bracket_type, m.round_number, m.match_number,
			m.group_label, m.participant_a_id, m.participant_b_id, m.score_a, m.score_b,
			m.winner_id, m.is_bye, m.status, m.next_match_id, m.next_match_slot,
			m.loser_next_match_id, m.loser_next_match_slot, m.playing_groups, m.map_name,
			m.created_at, m.updated_at,
			pa.id, pa.name, pa.seed_number, pa.group_name,
			pb.id, pb.name, pb.seed_number, pb.group_name
		FROM matches m
		LEFT JOIN participants pa ON pa.id = m.participant_a_id
		LEFT JOIN participants pb ON pb.id = m.participant_b_id
		WHERE m.stage_id = $1
		ORDER BY m.round_number ASC, m.match_number ASC
	`
	rows, err := r.pool.Query(ctx, query, stageID)
	if err != nil {
		return nil, fmt.Errorf("failed to query stage matches: %w", err)
	}
	defer rows.Close()

	var matches []domain.Match
	for rows.Next() {
		var m domain.Match
		var paID, pbID *uuid.UUID
		var paName, pbName *string
		var paSeed, pbSeed *int
		var paGroup, pbGroup *string

		err := rows.Scan(
			&m.ID, &m.TournamentID, &m.StageID, &m.BracketType, &m.RoundNumber, &m.MatchNumber,
			&m.GroupLabel, &m.ParticipantAID, &m.ParticipantBID, &m.ScoreA, &m.ScoreB,
			&m.WinnerID, &m.IsBye, &m.Status, &m.NextMatchID, &m.NextMatchSlot,
			&m.LoserNextMatchID, &m.LoserNextMatchSlot, &m.PlayingGroups, &m.MapName,
			&m.CreatedAt, &m.UpdatedAt,
			&paID, &paName, &paSeed, &paGroup,
			&pbID, &pbName, &pbSeed, &pbGroup,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan match: %w", err)
		}

		if paID != nil {
			m.ParticipantA = &domain.Participant{
				ID:           *paID,
				TournamentID: m.TournamentID,
				Name:         *paName,
				SeedNumber:   *paSeed,
				GroupName:    paGroup,
			}
		}
		if pbID != nil {
			m.ParticipantB = &domain.Participant{
				ID:           *pbID,
				TournamentID: m.TournamentID,
				Name:         *pbName,
				SeedNumber:   *pbSeed,
				GroupName:    pbGroup,
			}
		}

		matches = append(matches, m)
	}
	return matches, nil
}

func (r *MatchRepository) GetMatchByID(ctx context.Context, matchID uuid.UUID) (*domain.Match, error) {
	query := `
		SELECT
			id, tournament_id, stage_id, bracket_type, round_number, match_number,
			group_label, participant_a_id, participant_b_id, score_a, score_b,
			winner_id, is_bye, status, next_match_id, next_match_slot,
			loser_next_match_id, loser_next_match_slot, playing_groups, map_name,
			created_at, updated_at
		FROM matches
		WHERE id = $1
	`
	row := r.pool.QueryRow(ctx, query, matchID)

	var m domain.Match
	err := row.Scan(
		&m.ID, &m.TournamentID, &m.StageID, &m.BracketType, &m.RoundNumber, &m.MatchNumber,
		&m.GroupLabel, &m.ParticipantAID, &m.ParticipantBID, &m.ScoreA, &m.ScoreB,
		&m.WinnerID, &m.IsBye, &m.Status, &m.NextMatchID, &m.NextMatchSlot,
		&m.LoserNextMatchID, &m.LoserNextMatchSlot, &m.PlayingGroups, &m.MapName,
		&m.CreatedAt, &m.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to query match: %w", err)
	}
	return &m, nil
}

func (r *MatchRepository) GetMatchForUpdate(ctx context.Context, tx pgx.Tx, matchID uuid.UUID) (*domain.Match, error) {
	query := `
		SELECT
			id, tournament_id, stage_id, bracket_type, round_number, match_number,
			group_label, participant_a_id, participant_b_id, score_a, score_b,
			winner_id, is_bye, status, next_match_id, next_match_slot,
			loser_next_match_id, loser_next_match_slot, playing_groups, map_name,
			created_at, updated_at
		FROM matches
		WHERE id = $1
		FOR UPDATE
	`
	row := tx.QueryRow(ctx, query, matchID)

	var m domain.Match
	err := row.Scan(
		&m.ID, &m.TournamentID, &m.StageID, &m.BracketType, &m.RoundNumber, &m.MatchNumber,
		&m.GroupLabel, &m.ParticipantAID, &m.ParticipantBID, &m.ScoreA, &m.ScoreB,
		&m.WinnerID, &m.IsBye, &m.Status, &m.NextMatchID, &m.NextMatchSlot,
		&m.LoserNextMatchID, &m.LoserNextMatchSlot, &m.PlayingGroups, &m.MapName,
		&m.CreatedAt, &m.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to lock match: %w", err)
	}
	return &m, nil
}

func (r *MatchRepository) UpdateMatch(ctx context.Context, tx pgx.Tx, m *domain.Match) error {
	groups := m.PlayingGroups
	if groups == nil {
		groups = []string{}
	}

	query := `
		UPDATE matches
		SET participant_a_id = $1, participant_b_id = $2, score_a = $3, score_b = $4,
		    winner_id = $5, is_bye = $6, status = $7, playing_groups = $8, map_name = $9,
		    updated_at = NOW()
		WHERE id = $10
	`
	var err error
	if tx != nil {
		_, err = tx.Exec(ctx, query,
			m.ParticipantAID, m.ParticipantBID, m.ScoreA, m.ScoreB,
			m.WinnerID, m.IsBye, m.Status, groups, m.MapName, m.ID,
		)
	} else {
		_, err = r.pool.Exec(ctx, query,
			m.ParticipantAID, m.ParticipantBID, m.ScoreA, m.ScoreB,
			m.WinnerID, m.IsBye, m.Status, groups, m.MapName, m.ID,
		)
	}
	if err != nil {
		return fmt.Errorf("failed to update match: %w", err)
	}
	return nil
}
