package bracket

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/gridify/backend/internal/domain"
)

// Helper to convert string to pointer
func strPtr(s string) *string {
	return &s
}

// MatchStore defines the database operations required for match persistence.
type MatchStore interface {
	DeleteMatchesByStageID(ctx context.Context, tx pgx.Tx, stageID uuid.UUID) error
	BatchInsertMatches(ctx context.Context, tx pgx.Tx, matches []domain.Match) error
}

// GenerateSingleElimination generates a single elimination bracket (and optional bronze match)
// and persists it in a single database transaction.
func GenerateSingleElimination(
	ctx context.Context,
	tx pgx.Tx,
	matchRepo MatchStore,
	tournament *domain.Tournament,
	stage *domain.Stage,
	participants []domain.Participant,
) ([]domain.Match, error) {
	if len(participants) < 2 {
		return nil, fmt.Errorf("%w: minimal 2 peserta untuk membuat bracket single elimination", domain.ErrInvalidInput)
	}

	// Sort participants by SeedNumber
	sorted := make([]domain.Participant, len(participants))
	copy(sorted, participants)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].SeedNumber < sorted[j].SeedNumber
	})

	bracketSize := NextPowerOfTwo(len(sorted))
	totalRounds := int(math.Round(math.Log2(float64(bracketSize))))
	seedingOrder := StandardSeedingOrder(bracketSize)

	now := time.Now()

	// Map to hold matches by round and match number for linking
	// key: fmt.Sprintf("%s:%d:%d", bracketType, round, matchNum)
	matchMap := make(map[string]*domain.Match)
	var allMatches []*domain.Match

	// 1. Allocate all Upper Bracket matches
	for r := 1; r <= totalRounds; r++ {
		matchCount := bracketSize / (1 << r)
		for m := 1; m <= matchCount; m++ {
			match := &domain.Match{
				ID:           uuid.New(),
				TournamentID: tournament.ID,
				StageID:      stage.ID,
				BracketType:  "UPPER",
				RoundNumber:  r,
				MatchNumber:  m,
				ScoreA:       0,
				ScoreB:       0,
				Status:       "SCHEDULED",
				CreatedAt:    now,
				UpdatedAt:    now,
			}
			key := fmt.Sprintf("UPPER:%d:%d", r, m)
			matchMap[key] = match
			allMatches = append(allMatches, match)
		}
	}

	// 2. Allocate Third Place (Bronze) match if requested and there are at least 2 rounds
	var thirdPlaceMatch *domain.Match
	if tournament.HasThirdPlace && totalRounds >= 2 {
		thirdPlaceMatch = &domain.Match{
			ID:           uuid.New(),
			TournamentID: tournament.ID,
			StageID:      stage.ID,
			BracketType:  "THIRD_PLACE",
			RoundNumber:  totalRounds, // runs alongside Final
			MatchNumber:  1,
			ScoreA:       0,
			ScoreB:       0,
			Status:       "SCHEDULED",
			CreatedAt:    now,
			UpdatedAt:    now,
		}
		key := fmt.Sprintf("THIRD_PLACE:%d:1", totalRounds)
		matchMap[key] = thirdPlaceMatch
		allMatches = append(allMatches, thirdPlaceMatch)
	}

	// 3. Link progression: Upper Bracket next_match_id & next_match_slot
	for r := 1; r < totalRounds; r++ {
		matchCount := bracketSize / (1 << r)
		for m := 1; m <= matchCount; m++ {
			currKey := fmt.Sprintf("UPPER:%d:%d", r, m)
			currMatch := matchMap[currKey]

			nextRound := r + 1
			nextMatchNum := (m + 1) / 2
			nextKey := fmt.Sprintf("UPPER:%d:%d", nextRound, nextMatchNum)
			nextMatch := matchMap[nextKey]

			currMatch.NextMatchID = &nextMatch.ID
			if m%2 != 0 {
				currMatch.NextMatchSlot = strPtr("A")
			} else {
				currMatch.NextMatchSlot = strPtr("B")
			}
		}
	}

	// 4. Link Semifinals losers to Third Place match
	if thirdPlaceMatch != nil {
		semiRound := totalRounds - 1
		semiMatch1 := matchMap[fmt.Sprintf("UPPER:%d:1", semiRound)]
		semiMatch2 := matchMap[fmt.Sprintf("UPPER:%d:2", semiRound)]

		if semiMatch1 != nil {
			semiMatch1.LoserNextMatchID = &thirdPlaceMatch.ID
			semiMatch1.LoserNextMatchSlot = strPtr("A")
		}
		if semiMatch2 != nil {
			semiMatch2.LoserNextMatchID = &thirdPlaceMatch.ID
			semiMatch2.LoserNextMatchSlot = strPtr("B")
		}
	}

	// 5. Populate Round 1 participants and handle BYEs
	r1Count := bracketSize / 2
	for m := 1; m <= r1Count; m++ {
		match := matchMap[fmt.Sprintf("UPPER:1:%d", m)]
		seedA := seedingOrder[2*(m-1)]
		seedB := seedingOrder[2*(m-1)+1]

		var partA, partB *domain.Participant
		if seedA <= len(sorted) {
			partA = &sorted[seedA-1]
			match.ParticipantAID = &partA.ID
		}
		if seedB <= len(sorted) {
			partB = &sorted[seedB-1]
			match.ParticipantBID = &partB.ID
		}

		// Handle BYE auto-advance
		if partA != nil && partB == nil {
			// A gets BYE win
			match.IsBye = true
			match.Status = "COMPLETED"
			match.WinnerID = &partA.ID

			// Advance A to Round 2 immediately
			if match.NextMatchID != nil && match.NextMatchSlot != nil {
				for _, nextM := range allMatches {
					if nextM.ID == *match.NextMatchID {
						if *match.NextMatchSlot == "A" {
							nextM.ParticipantAID = &partA.ID
						} else {
							nextM.ParticipantBID = &partA.ID
						}
						break
					}
				}
			}
		} else if partA == nil && partB != nil {
			// B gets BYE win
			match.IsBye = true
			match.Status = "COMPLETED"
			match.WinnerID = &partB.ID

			// Advance B to Round 2 immediately
			if match.NextMatchID != nil && match.NextMatchSlot != nil {
				for _, nextM := range allMatches {
					if nextM.ID == *match.NextMatchID {
						if *match.NextMatchSlot == "A" {
							nextM.ParticipantAID = &partB.ID
						} else {
							nextM.ParticipantBID = &partB.ID
						}
						break
					}
				}
			}
		} else if partA != nil && partB != nil {
			match.Status = "SCHEDULED"
			match.IsBye = false
		}
	}

	// 6. Persist to database in transactional batch
	if err := matchRepo.DeleteMatchesByStageID(ctx, tx, stage.ID); err != nil {
		return nil, fmt.Errorf("gagal menghapus match lama: %w", err)
	}

	resultMatches := make([]domain.Match, len(allMatches))
	for i, m := range allMatches {
		resultMatches[i] = *m
	}

	if err := matchRepo.BatchInsertMatches(ctx, tx, resultMatches); err != nil {
		return nil, fmt.Errorf("gagal menyimpan batch match: %w", err)
	}

	return resultMatches, nil
}
