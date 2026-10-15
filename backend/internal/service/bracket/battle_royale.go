package bracket

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/gridify/backend/internal/domain"
)

// Default maps for Battle Royale match rotation
var defaultBRMaps = []string{"Erangel", "Miramar", "Sanhok", "Erangel", "Miramar", "Vikendi"}

// GenerateBattleRoyale generates the match schedule for a Battle Royale tournament stage.
// If total participants exceed br_max_teams_per_match, it creates multi-group lobbies with rotating pairs.
func GenerateBattleRoyale(
	ctx context.Context,
	tx pgx.Tx,
	matchRepo MatchStore,
	tournament *domain.Tournament,
	stage *domain.Stage,
	participants []domain.Participant,
) ([]domain.Match, error) {
	if len(participants) < 2 {
		return nil, fmt.Errorf("%w: minimal 2 peserta untuk membuat jadwal battle royale", domain.ErrInvalidInput)
	}

	totalMatches := tournament.BRTotalMatches
	if totalMatches <= 0 {
		totalMatches = 6
	}

	maxTeams := tournament.BRMaxTeamsPerMatch
	if maxTeams <= 0 {
		maxTeams = 16
	}

	now := time.Now()
	var matches []domain.Match

	if len(participants) <= maxTeams {
		// Single lobby: all teams play in every match
		for r := 1; r <= totalMatches; r++ {
			mapName := defaultBRMaps[(r-1)%len(defaultBRMaps)]
			matches = append(matches, domain.Match{
				ID:            uuid.New(),
				TournamentID:  tournament.ID,
				StageID:       stage.ID,
				BracketType:   "BATTLE_ROYALE",
				RoundNumber:   r,
				MatchNumber:   1,
				Status:        "SCHEDULED",
				PlayingGroups: []string{},
				MapName:       &mapName,
				CreatedAt:     now,
				UpdatedAt:     now,
			})
		}
	} else {
		// Multi-group lobby: split participants into groups
		// Target roughly half of maxTeams per group so 2 groups form 1 lobby
		targetPerGroup := maxTeams / 2
		if targetPerGroup < 2 {
			targetPerGroup = 2
		}
		groupCount := (len(participants) + targetPerGroup - 1) / targetPerGroup
		if groupCount < 3 {
			groupCount = 3
		}

		var groupNames []string
		for i := 0; i < groupCount; i++ {
			groupNames = append(groupNames, fmt.Sprintf("Grup %c", 'A'+i))
		}

		pairs := GenerateGroupPairs(groupNames)
		if len(pairs) == 0 {
			pairs = [][2]string{{"Grup A", "Grup B"}}
		}

		for r := 1; r <= totalMatches; r++ {
			pair := pairs[(r-1)%len(pairs)]
			mapName := defaultBRMaps[(r-1)%len(defaultBRMaps)]
			matches = append(matches, domain.Match{
				ID:            uuid.New(),
				TournamentID:  tournament.ID,
				StageID:       stage.ID,
				BracketType:   "BATTLE_ROYALE",
				RoundNumber:   r,
				MatchNumber:   1,
				Status:        "SCHEDULED",
				PlayingGroups: []string{pair[0], pair[1]},
				MapName:       &mapName,
				CreatedAt:     now,
				UpdatedAt:     now,
			})
		}
	}

	// Persist to database in transactional batch
	if err := matchRepo.DeleteMatchesByStageID(ctx, tx, stage.ID); err != nil {
		return nil, fmt.Errorf("gagal menghapus match lama: %w", err)
	}

	if err := matchRepo.BatchInsertMatches(ctx, tx, matches); err != nil {
		return nil, fmt.Errorf("gagal menyimpan batch match battle royale: %w", err)
	}

	return matches, nil
}
