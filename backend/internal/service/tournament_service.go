package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/gridify/backend/internal/database"
	"github.com/gridify/backend/internal/domain"
	"github.com/gridify/backend/internal/repository"
)

type TournamentService struct {
	pool           *pgxpool.Pool
	tournamentRepo *repository.TournamentRepository
	stageRepo      *repository.StageRepository
	gameRepo       *repository.GameRepository
}

func NewTournamentService(
	pool *pgxpool.Pool,
	tournamentRepo *repository.TournamentRepository,
	stageRepo *repository.StageRepository,
	gameRepo *repository.GameRepository,
) *TournamentService {
	return &TournamentService{
		pool:           pool,
		tournamentRepo: tournamentRepo,
		stageRepo:      stageRepo,
		gameRepo:       gameRepo,
	}
}

type CreateTournamentInput struct {
	GameID             uuid.UUID
	Title              string
	Description        string
	FormatType         string
	BestOf             int
	HomeAway           bool
	PointsPerWin       int
	PointsPerDraw      int
	HasThirdPlace      bool
	BRTotalMatches     int
	BRMaxTeamsPerMatch int
	BRPointSystem      json.RawMessage
}

func (s *TournamentService) CreateTournament(ctx context.Context, userID uuid.UUID, in CreateTournamentInput) (*domain.Tournament, error) {
	// 1. Validate Game
	game, err := s.gameRepo.GetGameByID(ctx, in.GameID)
	if err != nil {
		return nil, fmt.Errorf("%w: game id tidak ditemukan", domain.ErrBadRequest)
	}

	cleanTitle := strings.TrimSpace(in.Title)
	if len(cleanTitle) < 3 {
		return nil, fmt.Errorf("%w: judul turnamen minimal 3 karakter", domain.ErrBadRequest)
	}

	validFormats := map[string]bool{
		"SINGLE_ELIMINATION": true,
		"DOUBLE_ELIMINATION": true,
		"ROUND_ROBIN":        true,
		"HYBRID_UCL":         true,
		"BATTLE_ROYALE":      true,
	}
	if !validFormats[in.FormatType] {
		return nil, fmt.Errorf("%w: format turnamen tidak valid", domain.ErrBadRequest)
	}

	bestOf := in.BestOf
	if bestOf != 1 && bestOf != 3 && bestOf != 5 && bestOf != 7 {
		bestOf = 1
	}

	pointsPerWin := in.PointsPerWin
	if pointsPerWin < 1 {
		pointsPerWin = 3
	}

	pointsPerDraw := in.PointsPerDraw
	if pointsPerDraw < 0 {
		pointsPerDraw = 1
	}

	brTotalMatches := in.BRTotalMatches
	if brTotalMatches < 1 || brTotalMatches > 24 {
		brTotalMatches = 6
	}

	brMaxTeams := in.BRMaxTeamsPerMatch
	if brMaxTeams < 4 || brMaxTeams > 64 {
		brMaxTeams = 16
	}

	tournamentID := uuid.New()
	tourney := &domain.Tournament{
		ID:                 tournamentID,
		OwnerID:            userID,
		GameID:             in.GameID,
		Game:               game,
		Title:              cleanTitle,
		Description:        strings.TrimSpace(in.Description),
		FormatType:         in.FormatType,
		Status:             "DRAFT",
		BestOf:             bestOf,
		HomeAway:           in.HomeAway,
		PointsPerWin:       pointsPerWin,
		PointsPerDraw:      pointsPerDraw,
		HasThirdPlace:      in.HasThirdPlace,
		BRTotalMatches:     brTotalMatches,
		BRMaxTeamsPerMatch: brMaxTeams,
		BRPointSystem:      in.BRPointSystem,
	}

	// 2. Transactional execution of tournament + initial stages
	err = database.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		if err := s.tournamentRepo.CreateTournamentWithTx(ctx, tx, tourney); err != nil {
			return err
		}

		// Create stages based on format
		switch in.FormatType {
		case "SINGLE_ELIMINATION":
			s1 := &domain.Stage{
				ID:            uuid.New(),
				TournamentID:  tournamentID,
				Name:          "Babak Gugur",
				Type:          "SINGLE_ELIMINATION",
				SequenceOrder: 1,
				Status:        "PENDING",
			}
			if err := s.stageRepo.CreateStageWithTx(ctx, tx, s1); err != nil {
				return err
			}
			tourney.Stages = append(tourney.Stages, *s1)

		case "DOUBLE_ELIMINATION":
			s1 := &domain.Stage{
				ID:            uuid.New(),
				TournamentID:  tournamentID,
				Name:          "Double Elimination Bracket",
				Type:          "DOUBLE_ELIMINATION",
				SequenceOrder: 1,
				Status:        "PENDING",
			}
			if err := s.stageRepo.CreateStageWithTx(ctx, tx, s1); err != nil {
				return err
			}
			tourney.Stages = append(tourney.Stages, *s1)

		case "ROUND_ROBIN":
			s1 := &domain.Stage{
				ID:            uuid.New(),
				TournamentID:  tournamentID,
				Name:          "Babak Liga",
				Type:          "ROUND_ROBIN",
				SequenceOrder: 1,
				Status:        "PENDING",
			}
			if err := s.stageRepo.CreateStageWithTx(ctx, tx, s1); err != nil {
				return err
			}
			tourney.Stages = append(tourney.Stages, *s1)

		case "BATTLE_ROYALE":
			s1 := &domain.Stage{
				ID:            uuid.New(),
				TournamentID:  tournamentID,
				Name:          "Leaderboard Battle Royale",
				Type:          "LEADERBOARD",
				SequenceOrder: 1,
				Status:        "PENDING",
			}
			if err := s.stageRepo.CreateStageWithTx(ctx, tx, s1); err != nil {
				return err
			}
			tourney.Stages = append(tourney.Stages, *s1)

		case "HYBRID_UCL":
			s1 := &domain.Stage{
				ID:            uuid.New(),
				TournamentID:  tournamentID,
				Name:          "Fase Grup",
				Type:          "ROUND_ROBIN",
				SequenceOrder: 1,
				Status:        "PENDING",
			}
			if err := s.stageRepo.CreateStageWithTx(ctx, tx, s1); err != nil {
				return err
			}
			s2 := &domain.Stage{
				ID:            uuid.New(),
				TournamentID:  tournamentID,
				Name:          "Babak Knockout",
				Type:          "SINGLE_ELIMINATION",
				SequenceOrder: 2,
				Status:        "PENDING",
			}
			if err := s.stageRepo.CreateStageWithTx(ctx, tx, s2); err != nil {
				return err
			}
			tourney.Stages = append(tourney.Stages, *s1, *s2)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("gagal membuat turnamen: %w", err)
	}

	return tourney, nil
}

func (s *TournamentService) GetTournament(ctx context.Context, tournamentID uuid.UUID) (*domain.Tournament, error) {
	tourney, err := s.tournamentRepo.GetTournamentByID(ctx, tournamentID)
	if err != nil {
		return nil, err
	}

	stages, err := s.stageRepo.GetStagesByTournamentID(ctx, tournamentID)
	if err == nil {
		tourney.Stages = stages
	}

	return tourney, nil
}

func (s *TournamentService) ListTournaments(ctx context.Context, ownerID uuid.UUID, q, status string, page, perPage int) ([]domain.Tournament, int, error) {
	return s.tournamentRepo.ListTournaments(ctx, ownerID, q, status, page, perPage)
}

type UpdateTournamentInput struct {
	Title         *string
	Description   *string
	Status        *string
	BestOf        *int
	HomeAway      *bool
	PointsPerWin  *int
	PointsPerDraw *int
	HasThirdPlace *bool
}

func (s *TournamentService) UpdateTournament(ctx context.Context, tournamentID, userID uuid.UUID, in UpdateTournamentInput) (*domain.Tournament, error) {
	if err := s.tournamentRepo.VerifyTournamentOwnership(ctx, tournamentID, userID); err != nil {
		return nil, err
	}

	tourney, err := s.tournamentRepo.GetTournamentByID(ctx, tournamentID)
	if err != nil {
		return nil, err
	}

	if in.Title != nil {
		clean := strings.TrimSpace(*in.Title)
		if len(clean) < 3 {
			return nil, fmt.Errorf("%w: judul turnamen minimal 3 karakter", domain.ErrBadRequest)
		}
		tourney.Title = clean
	}
	if in.Description != nil {
		tourney.Description = strings.TrimSpace(*in.Description)
	}
	if in.Status != nil {
		st := *in.Status
		if st == "DRAFT" || st == "ONGOING" || st == "COMPLETED" {
			tourney.Status = st
		}
	}
	if in.BestOf != nil {
		bo := *in.BestOf
		if bo == 1 || bo == 3 || bo == 5 || bo == 7 {
			tourney.BestOf = bo
		}
	}
	if in.HomeAway != nil {
		tourney.HomeAway = *in.HomeAway
	}
	if in.PointsPerWin != nil && *in.PointsPerWin >= 1 {
		tourney.PointsPerWin = *in.PointsPerWin
	}
	if in.PointsPerDraw != nil && *in.PointsPerDraw >= 0 {
		tourney.PointsPerDraw = *in.PointsPerDraw
	}
	if in.HasThirdPlace != nil {
		tourney.HasThirdPlace = *in.HasThirdPlace
	}

	if err := s.tournamentRepo.UpdateTournament(ctx, tourney); err != nil {
		return nil, err
	}

	stages, _ := s.stageRepo.GetStagesByTournamentID(ctx, tournamentID)
	tourney.Stages = stages
	return tourney, nil
}

func (s *TournamentService) DeleteTournament(ctx context.Context, tournamentID, userID uuid.UUID) error {
	return s.tournamentRepo.DeleteTournament(ctx, tournamentID, userID)
}

func (s *TournamentService) VerifyOwnership(ctx context.Context, tournamentID, userID uuid.UUID) error {
	return s.tournamentRepo.VerifyTournamentOwnership(ctx, tournamentID, userID)
}
