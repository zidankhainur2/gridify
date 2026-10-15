package bracket

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/gridify/backend/internal/domain"
	"github.com/gridify/backend/internal/repository"
	"github.com/gridify/backend/internal/service"
)

type ProgressionEngine struct {
	matchRepo       *repository.MatchRepository
	stageRepo       *repository.StageRepository
	tournamentRepo  *repository.TournamentRepository
	participantRepo *repository.ParticipantRepository
}

func NewProgressionEngine(
	matchRepo *repository.MatchRepository,
	stageRepo *repository.StageRepository,
	tournamentRepo *repository.TournamentRepository,
	participantRepo *repository.ParticipantRepository,
) *ProgressionEngine {
	return &ProgressionEngine{
		matchRepo:       matchRepo,
		stageRepo:       stageRepo,
		tournamentRepo:  tournamentRepo,
		participantRepo: participantRepo,
	}
}

// UpdateMatchScore updates score, enforces best_of rules, checks downstream match safety on edits,
// advances winners and losers, and updates stage/tournament status.
func (p *ProgressionEngine) UpdateMatchScore(
	ctx context.Context,
	tx pgx.Tx,
	tournamentID, matchID uuid.UUID,
	scoreA, scoreB int,
	reqStatus string,
) (*domain.Match, error) {
	if scoreA < 0 || scoreB < 0 {
		return nil, fmt.Errorf("%w: skor tidak boleh negatif", domain.ErrInvalidInput)
	}

	tournament, err := p.tournamentRepo.GetTournamentByID(ctx, tournamentID)
	if err != nil {
		return nil, fmt.Errorf("turnamen tidak ditemukan: %w", err)
	}

	currMatch, err := p.matchRepo.GetMatchForUpdate(ctx, tx, matchID)
	if err != nil {
		return nil, fmt.Errorf("pertandingan tidak ditemukan: %w", err)
	}

	if currMatch.TournamentID != tournamentID {
		return nil, fmt.Errorf("%w: pertandingan bukan bagian dari turnamen ini", domain.ErrBadRequest)
	}

	if reqStatus == "" {
		reqStatus = "COMPLETED"
	}

	// Validate status
	if reqStatus != "ONGOING" && reqStatus != "COMPLETED" {
		return nil, fmt.Errorf("%w: status harus ONGOING atau COMPLETED", domain.ErrBadRequest)
	}

	// Best-Of and Draw validation
	isElimination := currMatch.BracketType == "UPPER" ||
		currMatch.BracketType == "LOWER" ||
		currMatch.BracketType == "GRAND_FINAL" ||
		currMatch.BracketType == "THIRD_PLACE"

	if reqStatus == "COMPLETED" && isElimination && scoreA == scoreB {
		return nil, fmt.Errorf("%w: pertandingan eliminasi tidak boleh berakhir seri", domain.ErrBadRequest)
	}

	bestOf := tournament.BestOf
	if bestOf <= 0 {
		bestOf = 1
	}
	reqWins := (bestOf / 2) + 1

	if reqStatus == "COMPLETED" && isElimination && bestOf > 1 {
		if scoreA < reqWins && scoreB < reqWins {
			return nil, fmt.Errorf("%w: format Best of %d membutuhkan minimal %d kemenangan untuk menyelesaikan pertandingan",
				domain.ErrBadRequest, bestOf, reqWins)
		}
		if scoreA > reqWins || scoreB > reqWins {
			return nil, fmt.Errorf("%w: skor kemenangan tidak boleh melebihi target %d pada Best of %d",
				domain.ErrBadRequest, reqWins, bestOf)
		}
	}

	var newWinnerID *uuid.UUID
	var newLoserID *uuid.UUID

	if reqStatus == "COMPLETED" {
		if scoreA > scoreB {
			newWinnerID = currMatch.ParticipantAID
			newLoserID = currMatch.ParticipantBID
		} else if scoreB > scoreA {
			newWinnerID = currMatch.ParticipantBID
			newLoserID = currMatch.ParticipantAID
		}
	}

	// Check for Score Correction Safety (RISK-02)
	// If the match was already completed, and the winner is changing
	isWinnerChanging := false
	if currMatch.Status == "COMPLETED" {
		if currMatch.WinnerID == nil && newWinnerID != nil {
			isWinnerChanging = true
		} else if currMatch.WinnerID != nil && newWinnerID == nil {
			isWinnerChanging = true
		} else if currMatch.WinnerID != nil && newWinnerID != nil && *currMatch.WinnerID != *newWinnerID {
			isWinnerChanging = true
		}
	}

	if isWinnerChanging {
		// Verify downstream matches have not started yet
		if currMatch.NextMatchID != nil {
			nextM, err := p.matchRepo.GetMatchForUpdate(ctx, tx, *currMatch.NextMatchID)
			if err != nil && !errors.Is(err, domain.ErrNotFound) {
				return nil, fmt.Errorf("gagal memeriksa pertandingan lanjutan: %w", err)
			}
			if nextM != nil && nextM.Status != "SCHEDULED" {
				return nil, domain.ErrDownstreamMatchPlayed
			}
		}

		if currMatch.LoserNextMatchID != nil {
			loserM, err := p.matchRepo.GetMatchForUpdate(ctx, tx, *currMatch.LoserNextMatchID)
			if err != nil && !errors.Is(err, domain.ErrNotFound) {
				return nil, fmt.Errorf("gagal memeriksa pertandingan loser bracket: %w", err)
			}
			if loserM != nil && loserM.Status != "SCHEDULED" {
				return nil, domain.ErrDownstreamMatchPlayed
			}
		}
	}

	// Advance Winner to next_match_id
	if (reqStatus == "COMPLETED" && (currMatch.Status != "COMPLETED" || isWinnerChanging)) && newWinnerID != nil {
		if currMatch.NextMatchID != nil && currMatch.NextMatchSlot != nil {
			nextM, err := p.matchRepo.GetMatchForUpdate(ctx, tx, *currMatch.NextMatchID)
			if err != nil {
				return nil, fmt.Errorf("gagal mengunci pertandingan berikutnya: %w", err)
			}
			if *currMatch.NextMatchSlot == "A" {
				nextM.ParticipantAID = newWinnerID
			} else {
				nextM.ParticipantBID = newWinnerID
			}
			if err := p.matchRepo.UpdateMatch(ctx, tx, nextM); err != nil {
				return nil, fmt.Errorf("gagal memajukan pemenang: %w", err)
			}
		}

		// Advance Loser to loser_next_match_id
		if currMatch.LoserNextMatchID != nil && currMatch.LoserNextMatchSlot != nil && newLoserID != nil {
			loserM, err := p.matchRepo.GetMatchForUpdate(ctx, tx, *currMatch.LoserNextMatchID)
			if err != nil {
				return nil, fmt.Errorf("gagal mengunci pertandingan lower bracket: %w", err)
			}
			if *currMatch.LoserNextMatchSlot == "A" {
				loserM.ParticipantAID = newLoserID
			} else {
				loserM.ParticipantBID = newLoserID
			}
			if err := p.matchRepo.UpdateMatch(ctx, tx, loserM); err != nil {
				return nil, fmt.Errorf("gagal memajukan tim yang kalah: %w", err)
			}
		}
	}

	// Update current match
	currMatch.ScoreA = scoreA
	currMatch.ScoreB = scoreB
	currMatch.Status = reqStatus
	currMatch.WinnerID = newWinnerID

	if err := p.matchRepo.UpdateMatch(ctx, tx, currMatch); err != nil {
		return nil, fmt.Errorf("gagal memperbarui skor pertandingan: %w", err)
	}

	// If tournament is DRAFT, change to ONGOING
	if tournament.Status == "DRAFT" {
		_ = p.tournamentRepo.UpdateTournamentStatus(ctx, tx, tournament.ID, "ONGOING")
	}

	// Check if stage is completed
	stageMatches, err := p.matchRepo.GetMatchesByStageID(ctx, currMatch.StageID)
	if err == nil {
		allCompleted := true
		for _, sm := range stageMatches {
			if sm.ID == currMatch.ID {
				if reqStatus != "COMPLETED" {
					allCompleted = false
					break
				}
			} else if sm.Status != "COMPLETED" {
				allCompleted = false
				break
			}
		}

		if allCompleted && len(stageMatches) > 0 {
			_ = p.stageRepo.UpdateStageStatus(ctx, tx, currMatch.StageID, "COMPLETED")

			// Check all stages in tournament
			stages, err := p.stageRepo.GetStagesByTournamentID(ctx, tournamentID)
			if err == nil {
				allStagesCompleted := true
				for _, stg := range stages {
					if stg.ID == currMatch.StageID {
						continue
					}
					if stg.Status != "COMPLETED" {
						allStagesCompleted = false
						break
					}
				}
				if allStagesCompleted {
					_ = p.tournamentRepo.UpdateTournamentStatus(ctx, tx, tournament.ID, "COMPLETED")
				}
			}
		}
	}

	return currMatch, nil
}

// UpdateMatchSlot allows manual admin placement of a participant into a scheduled match slot ('A' or 'B').
func (p *ProgressionEngine) UpdateMatchSlot(
	ctx context.Context,
	tx pgx.Tx,
	tournamentID, matchID uuid.UUID,
	slot string,
	participantID *uuid.UUID,
) (*domain.Match, error) {
	if slot != "A" && slot != "B" {
		return nil, fmt.Errorf("%w: slot harus bernilai 'A' atau 'B'", domain.ErrBadRequest)
	}

	currMatch, err := p.matchRepo.GetMatchForUpdate(ctx, tx, matchID)
	if err != nil {
		return nil, fmt.Errorf("pertandingan tidak ditemukan: %w", err)
	}

	if currMatch.TournamentID != tournamentID {
		return nil, fmt.Errorf("%w: pertandingan bukan bagian dari turnamen ini", domain.ErrBadRequest)
	}

	if currMatch.Status != "SCHEDULED" {
		return nil, fmt.Errorf("%w: hanya dapat mengubah slot pada pertandingan SCHEDULED", domain.ErrConflict)
	}

	if slot == "A" {
		currMatch.ParticipantAID = participantID
	} else {
		currMatch.ParticipantBID = participantID
	}

	if err := p.matchRepo.UpdateMatch(ctx, tx, currMatch); err != nil {
		return nil, fmt.Errorf("gagal memperbarui slot pertandingan: %w", err)
	}

	return currMatch, nil
}

// AdvanceGroupToKnockout advances top 2 teams per group from Group Stage to a new Knockout Stage in Hybrid UCL.
func (p *ProgressionEngine) AdvanceGroupToKnockout(
	ctx context.Context,
	tx pgx.Tx,
	tournamentID, stageID uuid.UUID,
) (*domain.Stage, []domain.Match, error) {
	tournament, err := p.tournamentRepo.GetTournamentByID(ctx, tournamentID)
	if err != nil {
		return nil, nil, fmt.Errorf("turnamen tidak ditemukan: %w", err)
	}

	stage, err := p.stageRepo.GetStageByID(ctx, stageID)
	if err != nil {
		return nil, nil, fmt.Errorf("babak tidak ditemukan: %w", err)
	}

	if stage.SequenceOrder != 1 && stage.Type != "ROUND_ROBIN" {
		return nil, nil, fmt.Errorf("%w: hanya babak grup yang dapat dimajukan ke babak knockout", domain.ErrBadRequest)
	}

	participants, err := p.participantRepo.ListParticipantsByTournamentID(ctx, tournamentID)
	if err != nil {
		return nil, nil, fmt.Errorf("gagal memuat peserta: %w", err)
	}

	matches, err := p.matchRepo.GetMatchesByStageID(ctx, stageID)
	if err != nil {
		return nil, nil, fmt.Errorf("gagal memuat pertandingan grup: %w", err)
	}

	standings := service.CalculateRoundRobinStandings(tournament, participants, matches)
	if len(standings) < 2 {
		return nil, nil, fmt.Errorf("%w: minimal 2 grup untuk melakukan transisi knockout", domain.ErrBadRequest)
	}

	// Pick Top 2 from each group: 1st and 2nd
	var qualified []domain.Participant
	for _, grp := range standings {
		if len(grp.Rows) < 2 {
			return nil, nil, fmt.Errorf("%w: grup %s memiliki peserta kurang dari 2", domain.ErrBadRequest, grp.GroupLabel)
		}
		qualified = append(qualified, grp.Rows[0].Participant, grp.Rows[1].Participant)
	}

	// Cross seeding:
	// Group A 1st vs Group B 2nd, Group B 1st vs Group A 2nd, etc.
	// For 2 groups:
	// A1 (seed 1), B2 (seed 4), B1 (seed 2), A2 (seed 3)
	var crossSeeded []domain.Participant
	for i := 0; i < len(standings); i++ {
		nextIdx := (i + 1) % len(standings)
		firstTeam := standings[i].Rows[0].Participant
		secondTeam := standings[nextIdx].Rows[1].Participant
		crossSeeded = append(crossSeeded, firstTeam, secondTeam)
	}

	// Re-assign seed numbers for bracket generator
	for i := range crossSeeded {
		crossSeeded[i].SeedNumber = i + 1
	}

	// Check if Knockout stage already exists or create new one
	stages, err := p.stageRepo.GetStagesByTournamentID(ctx, tournamentID)
	if err != nil {
		return nil, nil, fmt.Errorf("gagal membaca daftar babak: %w", err)
	}

	var knockoutStage *domain.Stage
	for i := range stages {
		if stages[i].SequenceOrder == 2 || stages[i].Type == "SINGLE_ELIMINATION" || stages[i].Type == "DOUBLE_ELIMINATION" {
			knockoutStage = &stages[i]
			break
		}
	}

	if knockoutStage == nil {
		knockoutStage = &domain.Stage{
			ID:            uuid.New(),
			TournamentID:  tournamentID,
			Name:          "Babak Knockout",
			Type:          "SINGLE_ELIMINATION",
			SequenceOrder: len(stages) + 1,
			Status:        "PENDING",
		}
		if err := p.stageRepo.CreateStageWithTx(ctx, tx, knockoutStage); err != nil {
			return nil, nil, fmt.Errorf("gagal membuat babak knockout: %w", err)
		}
	}

	// Mark Group Stage as COMPLETED
	_ = p.stageRepo.UpdateStageStatus(ctx, tx, stageID, "COMPLETED")

	// Generate full Single Elimination bracket for Knockout stage
	knockoutMatches, err := GenerateSingleElimination(ctx, tx, p.matchRepo, tournament, knockoutStage, crossSeeded)
	if err != nil {
		return nil, nil, fmt.Errorf("gagal membangun pohon bracket knockout: %w", err)
	}

	return knockoutStage, knockoutMatches, nil
}
