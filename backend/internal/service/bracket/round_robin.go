package bracket

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/gridify/backend/internal/domain"
)

// GenerateRoundRobin generates a single- or double-leg round robin schedule for a tournament stage.
// For HYBRID_UCL or multi-group tournaments, it splits teams into groups and generates schedules per group.
func GenerateRoundRobin(
	ctx context.Context,
	tx pgx.Tx,
	matchRepo MatchStore,
	tournament *domain.Tournament,
	stage *domain.Stage,
	participants []domain.Participant,
) ([]domain.Match, error) {
	if len(participants) < 2 {
		return nil, fmt.Errorf("%w: minimal 2 peserta untuk membuat jadwal round robin", domain.ErrInvalidInput)
	}

	now := time.Now()
	var allMatches []domain.Match

	// Check if this is a multi-group stage (e.g. HYBRID_UCL group stage)
	// If participants already have GroupName assigned or tournament is HYBRID_UCL
	isMultiGroup := tournament.FormatType == "HYBRID_UCL"

	if isMultiGroup {
		// Divide teams into groups of 4 (or at least 2 groups)
		groupCount := len(participants) / 4
		if groupCount < 2 {
			groupCount = 2
		}
		groups := DistributeTeamsToGroups(participants, groupCount)

		// Find max rounds across groups to coordinate round numbers
		for gIdx, groupTeams := range groups {
			groupLabel := fmt.Sprintf("Grup %c", 'A'+gIdx)
			groupMatches := generateGroupRoundRobin(tournament, stage, groupTeams, &groupLabel, now)
			allMatches = append(allMatches, groupMatches...)
		}
	} else {
		// Single group round robin
		allMatches = generateGroupRoundRobin(tournament, stage, participants, nil, now)
	}

	// Re-assign match_number sequentially per (bracket_type, round_number)
	// to strictly satisfy uq_matches_stage_bracket_round_num constraint
	roundMatchCounter := make(map[int]int)
	for i := range allMatches {
		r := allMatches[i].RoundNumber
		roundMatchCounter[r]++
		allMatches[i].MatchNumber = roundMatchCounter[r]
	}

	// Persist to database in transactional batch
	if err := matchRepo.DeleteMatchesByStageID(ctx, tx, stage.ID); err != nil {
		return nil, fmt.Errorf("gagal menghapus match lama: %w", err)
	}

	if err := matchRepo.BatchInsertMatches(ctx, tx, allMatches); err != nil {
		return nil, fmt.Errorf("gagal menyimpan batch match round robin: %w", err)
	}

	return allMatches, nil
}

// generateGroupRoundRobin generates the matches for a single group using the Circle Method.
func generateGroupRoundRobin(
	tournament *domain.Tournament,
	stage *domain.Stage,
	teams []domain.Participant,
	groupLabel *string,
	now time.Time,
) []domain.Match {
	if len(teams) < 2 {
		return nil
	}

	n := len(teams)
	// If odd number of teams, we use a dummy (nil) to represent a BYE
	teamList := make([]*domain.Participant, n)
	for i := range teams {
		teamList[i] = &teams[i]
	}
	if n%2 != 0 {
		teamList = append(teamList, nil)
		n++
	}

	totalRounds := n - 1
	matchesPerRound := n / 2
	var matches []domain.Match

	// Circle Method:
	// Fix teamList[0], rotate the rest (1 .. n-1)
	rotation := make([]*domain.Participant, n-1)
	copy(rotation, teamList[1:])

	for r := 1; r <= totalRounds; r++ {
		// Construct round pairings
		roundTeams := make([]*domain.Participant, n)
		roundTeams[0] = teamList[0]
		for i := 0; i < n-1; i++ {
			roundTeams[i+1] = rotation[i]
		}

		for m := 0; m < matchesPerRound; m++ {
			teamA := roundTeams[m]
			teamB := roundTeams[n-1-m]

			// Skip matches involving dummy BYE
			if teamA == nil || teamB == nil {
				continue
			}

			// Alternate home/away for the fixed team to maintain balance
			if m == 0 && r%2 == 0 {
				teamA, teamB = teamB, teamA
			}

			matches = append(matches, domain.Match{
				ID:             uuid.New(),
				TournamentID:   tournament.ID,
				StageID:        stage.ID,
				BracketType:    "ROUND_ROBIN",
				RoundNumber:    r,
				GroupLabel:     groupLabel,
				ParticipantAID: &teamA.ID,
				ParticipantBID: &teamB.ID,
				ScoreA:         0,
				ScoreB:         0,
				Status:         "SCHEDULED",
				CreatedAt:      now,
				UpdatedAt:      now,
			})
		}

		// Rotate: last element moves to first of rotation
		last := rotation[len(rotation)-1]
		copy(rotation[1:], rotation[:len(rotation)-1])
		rotation[0] = last
	}

	// If 2-Leg Home & Away is enabled, generate return leg
	if tournament.HomeAway {
		leg1Count := len(matches)
		for i := 0; i < leg1Count; i++ {
			leg1Match := matches[i]
			matches = append(matches, domain.Match{
				ID:             uuid.New(),
				TournamentID:   tournament.ID,
				StageID:        stage.ID,
				BracketType:    "ROUND_ROBIN",
				RoundNumber:    leg1Match.RoundNumber + totalRounds,
				GroupLabel:     groupLabel,
				ParticipantAID: leg1Match.ParticipantBID, // Swapped home/away
				ParticipantBID: leg1Match.ParticipantAID, // Swapped home/away
				ScoreA:         0,
				ScoreB:         0,
				Status:         "SCHEDULED",
				CreatedAt:      now,
				UpdatedAt:      now,
			})
		}
	}

	return matches
}
