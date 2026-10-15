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

// lbMatchCount returns the number of matches in Lower Bracket round r for a given bracket size.
func lbMatchCount(bracketSize int, r int) int {
	shift := (r+1)/2 + 1
	count := bracketSize >> shift
	if count < 1 {
		return 1
	}
	return count
}

// GenerateDoubleElimination generates a double elimination bracket with Upper Bracket,
// Lower Bracket, and Grand Final, and persists it in a single database transaction.
func GenerateDoubleElimination(
	ctx context.Context,
	tx pgx.Tx,
	matchRepo MatchStore,
	tournament *domain.Tournament,
	stage *domain.Stage,
	participants []domain.Participant,
) ([]domain.Match, error) {
	if len(participants) < 3 {
		return nil, fmt.Errorf("%w: minimal 3 peserta untuk membuat bracket double elimination", domain.ErrInvalidInput)
	}

	// Sort participants by SeedNumber
	sorted := make([]domain.Participant, len(participants))
	copy(sorted, participants)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].SeedNumber < sorted[j].SeedNumber
	})

	bracketSize := NextPowerOfTwo(len(sorted))
	if bracketSize < 4 {
		bracketSize = 4
	}

	totalWBRounds := int(math.Round(math.Log2(float64(bracketSize))))
	totalLBRounds := 2 * (totalWBRounds - 1)
	seedingOrder := StandardSeedingOrder(bracketSize)

	now := time.Now()

	matchMap := make(map[string]*domain.Match)
	var allMatches []*domain.Match

	// 1. Allocate Upper Bracket matches
	for r := 1; r <= totalWBRounds; r++ {
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

	// 2. Allocate Lower Bracket matches
	for r := 1; r <= totalLBRounds; r++ {
		matchCount := lbMatchCount(bracketSize, r)
		for m := 1; m <= matchCount; m++ {
			match := &domain.Match{
				ID:           uuid.New(),
				TournamentID: tournament.ID,
				StageID:      stage.ID,
				BracketType:  "LOWER",
				RoundNumber:  r,
				MatchNumber:  m,
				ScoreA:       0,
				ScoreB:       0,
				Status:       "SCHEDULED",
				CreatedAt:    now,
				UpdatedAt:    now,
			}
			key := fmt.Sprintf("LOWER:%d:%d", r, m)
			matchMap[key] = match
			allMatches = append(allMatches, match)
		}
	}

	// 3. Allocate Grand Final match
	grandFinal := &domain.Match{
		ID:           uuid.New(),
		TournamentID: tournament.ID,
		StageID:      stage.ID,
		BracketType:  "GRAND_FINAL",
		RoundNumber:  1,
		MatchNumber:  1,
		ScoreA:       0,
		ScoreB:       0,
		Status:       "SCHEDULED",
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	matchMap["GRAND_FINAL:1:1"] = grandFinal
	allMatches = append(allMatches, grandFinal)

	// 4. Link Upper Bracket progression (winners)
	for r := 1; r < totalWBRounds; r++ {
		matchCount := bracketSize / (1 << r)
		for m := 1; m <= matchCount; m++ {
			curr := matchMap[fmt.Sprintf("UPPER:%d:%d", r, m)]
			next := matchMap[fmt.Sprintf("UPPER:%d:%d", r+1, (m+1)/2)]
			curr.NextMatchID = &next.ID
			if m%2 != 0 {
				curr.NextMatchSlot = strPtr("A")
			} else {
				curr.NextMatchSlot = strPtr("B")
			}
		}
	}
	// Winner of WB Final goes to Grand Final Slot A
	wbFinal := matchMap[fmt.Sprintf("UPPER:%d:1", totalWBRounds)]
	wbFinal.NextMatchID = &grandFinal.ID
	wbFinal.NextMatchSlot = strPtr("A")

	// 5. Link Lower Bracket progression (winners)
	for r := 1; r < totalLBRounds; r++ {
		matchCount := lbMatchCount(bracketSize, r)

		for m := 1; m <= matchCount; m++ {
			curr := matchMap[fmt.Sprintf("LOWER:%d:%d", r, m)]
			if r%2 != 0 {
				// Odd LB round to Even LB round: same match count, feeds slot A
				next := matchMap[fmt.Sprintf("LOWER:%d:%d", r+1, m)]
				curr.NextMatchID = &next.ID
				curr.NextMatchSlot = strPtr("A")
			} else {
				// Even LB round to Odd LB round: halves match count
				next := matchMap[fmt.Sprintf("LOWER:%d:%d", r+1, (m+1)/2)]
				curr.NextMatchID = &next.ID
				if m%2 != 0 {
					curr.NextMatchSlot = strPtr("A")
				} else {
					curr.NextMatchSlot = strPtr("B")
				}
			}
		}
	}
	// Winner of LB Final goes to Grand Final Slot B
	lbFinal := matchMap[fmt.Sprintf("LOWER:%d:1", totalLBRounds)]
	lbFinal.NextMatchID = &grandFinal.ID
	lbFinal.NextMatchSlot = strPtr("B")

	// 6. Link Upper Bracket losers dropping down to Lower Bracket
	// 6a. WB Round 1 losers -> LB Round 1
	wbR1Count := bracketSize / 2
	for m := 1; m <= wbR1Count; m++ {
		curr := matchMap[fmt.Sprintf("UPPER:1:%d", m)]
		targetLBMatchNum := (m + 1) / 2
		targetLB := matchMap[fmt.Sprintf("LOWER:1:%d", targetLBMatchNum)]
		curr.LoserNextMatchID = &targetLB.ID
		if m%2 != 0 {
			curr.LoserNextMatchSlot = strPtr("A")
		} else {
			curr.LoserNextMatchSlot = strPtr("B")
		}
	}

	// 6b. WB Round r (r >= 2) losers -> LB Round 2*(r - 1) Slot B
	for r := 2; r <= totalWBRounds; r++ {
		matchCount := bracketSize / (1 << r)
		targetLBRound := 2 * (r - 1)
		for m := 1; m <= matchCount; m++ {
			curr := matchMap[fmt.Sprintf("UPPER:%d:%d", r, m)]
			targetLB := matchMap[fmt.Sprintf("LOWER:%d:%d", targetLBRound, m)]
			curr.LoserNextMatchID = &targetLB.ID
			curr.LoserNextMatchSlot = strPtr("B")
		}
	}

	// 7. Seed Round 1 Upper Bracket participants & handle BYEs
	wbR1Matches := bracketSize / 2
	for m := 1; m <= wbR1Matches; m++ {
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

		if partA != nil && partB == nil {
			// A gets BYE win
			match.IsBye = true
			match.Status = "COMPLETED"
			match.WinnerID = &partA.ID

			// Advance A to WB Round 2
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

			// Advance B to WB Round 2
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

	// 8. Check LB Round 1 matches where both WB feeds were BYEs
	// In such cases, no team will ever drop into either slot, so mark LB match as completed BYE
	lbR1Count := bracketSize / 4
	for m := 1; m <= lbR1Count; m++ {
		wbMatch1 := matchMap[fmt.Sprintf("UPPER:1:%d", 2*m-1)]
		wbMatch2 := matchMap[fmt.Sprintf("UPPER:1:%d", 2*m)]
		lbMatch := matchMap[fmt.Sprintf("LOWER:1:%d", m)]

		if wbMatch1 != nil && wbMatch2 != nil && wbMatch1.IsBye && wbMatch2.IsBye {
			lbMatch.IsBye = true
			lbMatch.Status = "COMPLETED"
		}
	}

	// 9. Persist to database in transactional batch
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
