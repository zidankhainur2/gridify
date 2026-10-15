package service

import (
	"context"
	"fmt"
	"sort"

	"github.com/google/uuid"

	"github.com/gridify/backend/internal/domain"
	"github.com/gridify/backend/internal/repository"
)

type StandingRow struct {
	ParticipantID  uuid.UUID          `json:"participant_id"`
	ParticipantName string            `json:"participant_name"`
	Participant    domain.Participant `json:"participant"`
	Played         int                `json:"played"`
	Won            int                `json:"won"`
	Drawn          int                `json:"drawn"`
	Lost           int                `json:"lost"`
	GoalsFor       int                `json:"goals_for"`
	GoalsAgainst   int                `json:"goals_against"`
	GoalDifference int                `json:"goal_difference"`
	Points         int                `json:"points"`
	Rank           int                `json:"rank"`
}

type GroupStanding struct {
	GroupLabel string        `json:"group_label"`
	Rows       []StandingRow `json:"rows"`
}

type BRLeaderboardRow struct {
	ParticipantID   uuid.UUID          `json:"participant_id"`
	ParticipantName string             `json:"participant_name"`
	Participant     domain.Participant `json:"participant"`
	MatchesPlayed   int                `json:"matches_played"`
	Wins            int                `json:"wins"` // WWCD
	TotalKills      int                `json:"total_kills"`
	PlacementPoints int                `json:"placement_points"`
	KillPoints      int                `json:"kill_points"`
	TotalPoints     int                `json:"total_points"`
	Rank            int                `json:"rank"`
}

type Champions struct {
	First  *domain.Participant `json:"first"`
	Second *domain.Participant `json:"second"`
	Third  *domain.Participant `json:"third"`
}

type StageView struct {
	Stage         domain.Stage        `json:"stage"`
	Tournament    *domain.Tournament  `json:"tournament,omitempty"`
	Stages        []domain.Stage      `json:"stages,omitempty"`
	Matches       []domain.Match      `json:"matches"`
	Standings     []GroupStanding     `json:"standings,omitempty"`
	BRLeaderboard []BRLeaderboardRow  `json:"br_leaderboard,omitempty"`
	Champions     *Champions          `json:"champions,omitempty"`
}

type StandingsService struct {
	matchRepo       *repository.MatchRepository
	brRepo          *repository.BattleRoyaleRepository
	participantRepo *repository.ParticipantRepository
	stageRepo       *repository.StageRepository
	tournamentRepo  *repository.TournamentRepository
}

func NewStandingsService(
	matchRepo *repository.MatchRepository,
	brRepo *repository.BattleRoyaleRepository,
	participantRepo *repository.ParticipantRepository,
	stageRepo *repository.StageRepository,
	tournamentRepo *repository.TournamentRepository,
) *StandingsService {
	return &StandingsService{
		matchRepo:       matchRepo,
		brRepo:          brRepo,
		participantRepo: participantRepo,
		stageRepo:       stageRepo,
		tournamentRepo:  tournamentRepo,
	}
}

// GetStageView compiles matches, calculated standings / BR leaderboard, and champions for a stage.
func (s *StandingsService) GetStageView(ctx context.Context, tournamentID, stageID uuid.UUID) (*StageView, error) {
	tournament, err := s.tournamentRepo.GetTournamentByID(ctx, tournamentID)
	if err != nil {
		return nil, fmt.Errorf("gagal memuat turnamen: %w", err)
	}

	stage, err := s.stageRepo.GetStageByID(ctx, stageID)
	if err != nil {
		return nil, fmt.Errorf("gagal memuat babak turnamen: %w", err)
	}

	participants, err := s.participantRepo.ListParticipantsByTournamentID(ctx, tournamentID)
	if err != nil {
		return nil, fmt.Errorf("gagal memuat peserta: %w", err)
	}

	matches, err := s.matchRepo.GetMatchesByStageID(ctx, stageID)
	if err != nil {
		return nil, fmt.Errorf("gagal memuat jadwal pertandingan: %w", err)
	}

	stages, _ := s.stageRepo.ListStagesByTournamentID(ctx, tournamentID)

	view := &StageView{
		Stage:      *stage,
		Tournament: tournament,
		Stages:     stages,
		Matches:    matches,
	}

	if tournament.FormatType == "ROUND_ROBIN" || (tournament.FormatType == "HYBRID_UCL" && (stage.Type == "GROUP_STAGE" || stage.SequenceOrder == 1)) {
		view.Standings = CalculateRoundRobinStandings(tournament, participants, matches)
	} else if tournament.FormatType == "BATTLE_ROYALE" {
		brResults, err := s.brRepo.GetResultsByStageID(ctx, stageID)
		if err != nil {
			return nil, fmt.Errorf("gagal memuat hasil battle royale: %w", err)
		}
		view.BRLeaderboard = CalculateBRLeaderboard(participants, brResults)

		// Group results by match ID for pre-population in edit modal
		brByMatch := make(map[uuid.UUID][]domain.BattleRoyaleMatchResult)
		for _, res := range brResults {
			brByMatch[res.MatchID] = append(brByMatch[res.MatchID], res)
		}
		for i := range view.Matches {
			if results, ok := brByMatch[view.Matches[i].ID]; ok {
				view.Matches[i].BRResults = results
			}
		}
	}

	view.Champions = DetermineChampions(tournament, stage, participants, matches, view.Standings, view.BRLeaderboard)

	return view, nil
}

// CalculateRoundRobinStandings calculates standings for Round Robin and Hybrid UCL group stages.
func CalculateRoundRobinStandings(
	tournament *domain.Tournament,
	participants []domain.Participant,
	matches []domain.Match,
) []GroupStanding {
	ptsWin := tournament.PointsPerWin
	if ptsWin <= 0 {
		ptsWin = 3
	}
	ptsDraw := tournament.PointsPerDraw
	if ptsDraw <= 0 {
		ptsDraw = 1
	}

	partMap := make(map[uuid.UUID]domain.Participant)
	for _, p := range participants {
		partMap[p.ID] = p
	}

	// Group matches by group label
	groupMatches := make(map[string][]domain.Match)
	for _, m := range matches {
		label := ""
		if m.GroupLabel != nil {
			label = *m.GroupLabel
		}
		groupMatches[label] = append(groupMatches[label], m)
	}

	// Also find which participants belong to which group label
	groupParts := make(map[string]map[uuid.UUID]bool)
	for _, m := range matches {
		label := ""
		if m.GroupLabel != nil {
			label = *m.GroupLabel
		}
		if groupParts[label] == nil {
			groupParts[label] = make(map[uuid.UUID]bool)
		}
		if m.ParticipantAID != nil {
			groupParts[label][*m.ParticipantAID] = true
		}
		if m.ParticipantBID != nil {
			groupParts[label][*m.ParticipantBID] = true
		}
	}

	var result []GroupStanding

	// Deterministic group labels order
	var labels []string
	for l := range groupMatches {
		labels = append(labels, l)
	}
	sort.Strings(labels)

	for _, label := range labels {
		mList := groupMatches[label]
		rowsMap := make(map[uuid.UUID]*StandingRow)

		// Initialize rows for all participants in this group
		for pID := range groupParts[label] {
			if p, ok := partMap[pID]; ok {
				rowsMap[pID] = &StandingRow{
					ParticipantID:   p.ID,
					ParticipantName: p.Name,
					Participant:     p,
				}
			}
		}
		// If single group and no group parts populated, initialize with all participants
		if len(groupParts[label]) == 0 {
			for _, p := range participants {
				rowsMap[p.ID] = &StandingRow{
					ParticipantID:   p.ID,
					ParticipantName: p.Name,
					Participant:     p,
				}
			}
		}

		// Process matches
		for _, m := range mList {
			if m.Status != "COMPLETED" || m.ParticipantAID == nil || m.ParticipantBID == nil {
				continue
			}

			rowA := rowsMap[*m.ParticipantAID]
			rowB := rowsMap[*m.ParticipantBID]
			if rowA == nil || rowB == nil {
				continue
			}

			rowA.Played++
			rowB.Played++
			rowA.GoalsFor += m.ScoreA
			rowA.GoalsAgainst += m.ScoreB
			rowB.GoalsFor += m.ScoreB
			rowB.GoalsAgainst += m.ScoreA

			if m.ScoreA > m.ScoreB {
				rowA.Won++
				rowA.Points += ptsWin
				rowB.Lost++
			} else if m.ScoreB > m.ScoreA {
				rowB.Won++
				rowB.Points += ptsWin
				rowA.Lost++
			} else {
				rowA.Drawn++
				rowA.Points += ptsDraw
				rowB.Drawn++
				rowB.Points += ptsDraw
			}
		}

		var rows []StandingRow
		for _, r := range rowsMap {
			r.GoalDifference = r.GoalsFor - r.GoalsAgainst
			rows = append(rows, *r)
		}

		// Sort rows: Points DESC -> GD DESC -> GF DESC -> HeadToHead -> Seed ASC
		sort.Slice(rows, func(i, j int) bool {
			a, b := rows[i], rows[j]
			if a.Points != b.Points {
				return a.Points > b.Points
			}
			if a.GoalDifference != b.GoalDifference {
				return a.GoalDifference > b.GoalDifference
			}
			if a.GoalsFor != b.GoalsFor {
				return a.GoalsFor > b.GoalsFor
			}
			// Head-to-Head check
			h2hA, h2hB := 0, 0
			for _, m := range mList {
				if m.Status != "COMPLETED" || m.ParticipantAID == nil || m.ParticipantBID == nil {
					continue
				}
				if *m.ParticipantAID == a.Participant.ID && *m.ParticipantBID == b.Participant.ID {
					if m.ScoreA > m.ScoreB {
						h2hA += ptsWin
					} else if m.ScoreB > m.ScoreA {
						h2hB += ptsWin
					}
				} else if *m.ParticipantAID == b.Participant.ID && *m.ParticipantBID == a.Participant.ID {
					if m.ScoreA > m.ScoreB {
						h2hB += ptsWin
					} else if m.ScoreB > m.ScoreA {
						h2hA += ptsWin
					}
				}
			}
			if h2hA != h2hB {
				return h2hA > h2hB
			}
			return a.Participant.SeedNumber < b.Participant.SeedNumber
		})

		for idx := range rows {
			rows[idx].Rank = idx + 1
		}

		result = append(result, GroupStanding{
			GroupLabel: label,
			Rows:       rows,
		})
	}

	return result
}

// CalculateBRLeaderboard calculates Battle Royale standings from individual match results.
func CalculateBRLeaderboard(
	participants []domain.Participant,
	results []domain.BattleRoyaleMatchResult,
) []BRLeaderboardRow {
	partMap := make(map[uuid.UUID]domain.Participant)
	for _, p := range participants {
		partMap[p.ID] = p
	}

	rowMap := make(map[uuid.UUID]*BRLeaderboardRow)
	for _, p := range participants {
		rowMap[p.ID] = &BRLeaderboardRow{
			ParticipantID:   p.ID,
			ParticipantName: p.Name,
			Participant:     p,
		}
	}

	for _, res := range results {
		row, ok := rowMap[res.ParticipantID]
		if !ok {
			p, hasP := partMap[res.ParticipantID]
			if !hasP {
				p = domain.Participant{ID: res.ParticipantID, Name: "Unknown"}
			}
			row = &BRLeaderboardRow{
				ParticipantID:   p.ID,
				ParticipantName: p.Name,
				Participant:     p,
			}
			rowMap[res.ParticipantID] = row
		}

		row.MatchesPlayed++
		if res.PlacementRank == 1 {
			row.Wins++
		}
		row.TotalKills += res.Kills
		row.PlacementPoints += res.PlacementPoints
		row.KillPoints += res.KillPoints
		row.TotalPoints += res.TotalPoints
	}

	var rows []BRLeaderboardRow
	for _, r := range rowMap {
		rows = append(rows, *r)
	}

	// Tie-breakers: TotalPoints DESC -> Wins DESC -> PlacementPoints DESC -> TotalKills DESC -> Seed ASC
	sort.Slice(rows, func(i, j int) bool {
		a, b := rows[i], rows[j]
		if a.TotalPoints != b.TotalPoints {
			return a.TotalPoints > b.TotalPoints
		}
		if a.Wins != b.Wins {
			return a.Wins > b.Wins
		}
		if a.PlacementPoints != b.PlacementPoints {
			return a.PlacementPoints > b.PlacementPoints
		}
		if a.TotalKills != b.TotalKills {
			return a.TotalKills > b.TotalKills
		}
		return a.Participant.SeedNumber < b.Participant.SeedNumber
	})

	for idx := range rows {
		rows[idx].Rank = idx + 1
	}

	return rows
}

// DetermineChampions calculates 1st, 2nd, and 3rd place champions uniformly for all 5 tournament formats.
func DetermineChampions(
	tournament *domain.Tournament,
	stage *domain.Stage,
	participants []domain.Participant,
	matches []domain.Match,
	standings []GroupStanding,
	brLeaderboard []BRLeaderboardRow,
) *Champions {
	partMap := make(map[uuid.UUID]domain.Participant)
	for _, p := range participants {
		partMap[p.ID] = p
	}

	switch tournament.FormatType {
	case "SINGLE_ELIMINATION":
		// Find highest round UPPER match (The Final)
		var finalMatch *domain.Match
		var maxRound int
		for i := range matches {
			m := &matches[i]
			if m.BracketType == "UPPER" && m.RoundNumber > maxRound {
				maxRound = m.RoundNumber
				finalMatch = m
			}
		}

		if finalMatch == nil || finalMatch.Status != "COMPLETED" || finalMatch.WinnerID == nil {
			return nil
		}

		champ := &Champions{}
		if p, ok := partMap[*finalMatch.WinnerID]; ok {
			champ.First = &p
		}

		// Runner up is the other participant in Final
		if finalMatch.ParticipantAID != nil && *finalMatch.ParticipantAID != *finalMatch.WinnerID {
			if p, ok := partMap[*finalMatch.ParticipantAID]; ok {
				champ.Second = &p
			}
		} else if finalMatch.ParticipantBID != nil && *finalMatch.ParticipantBID != *finalMatch.WinnerID {
			if p, ok := partMap[*finalMatch.ParticipantBID]; ok {
				champ.Second = &p
			}
		}

		// Bronze match
		if tournament.HasThirdPlace {
			for i := range matches {
				m := &matches[i]
				if m.BracketType == "THIRD_PLACE" && m.Status == "COMPLETED" && m.WinnerID != nil {
					if p, ok := partMap[*m.WinnerID]; ok {
						champ.Third = &p
					}
					break
				}
			}
		}
		return champ

	case "DOUBLE_ELIMINATION":
		var grandFinal *domain.Match
		var lbFinal *domain.Match
		var maxLBRound int

		for i := range matches {
			m := &matches[i]
			if m.BracketType == "GRAND_FINAL" {
				grandFinal = m
			} else if m.BracketType == "LOWER" && m.RoundNumber > maxLBRound {
				maxLBRound = m.RoundNumber
				lbFinal = m
			}
		}

		if grandFinal == nil || grandFinal.Status != "COMPLETED" || grandFinal.WinnerID == nil {
			return nil
		}

		champ := &Champions{}
		if p, ok := partMap[*grandFinal.WinnerID]; ok {
			champ.First = &p
		}

		// Runner up is the loser of Grand Final
		if grandFinal.ParticipantAID != nil && *grandFinal.ParticipantAID != *grandFinal.WinnerID {
			if p, ok := partMap[*grandFinal.ParticipantAID]; ok {
				champ.Second = &p
			}
		} else if grandFinal.ParticipantBID != nil && *grandFinal.ParticipantBID != *grandFinal.WinnerID {
			if p, ok := partMap[*grandFinal.ParticipantBID]; ok {
				champ.Second = &p
			}
		}

		// Third place is the team that lost LB Final (didn't reach Grand Final)
		if lbFinal != nil && lbFinal.Status == "COMPLETED" && lbFinal.WinnerID != nil {
			if lbFinal.ParticipantAID != nil && *lbFinal.ParticipantAID != *lbFinal.WinnerID {
				if p, ok := partMap[*lbFinal.ParticipantAID]; ok {
					champ.Third = &p
				}
			} else if lbFinal.ParticipantBID != nil && *lbFinal.ParticipantBID != *lbFinal.WinnerID {
				if p, ok := partMap[*lbFinal.ParticipantBID]; ok {
					champ.Third = &p
				}
			}
		}
		return champ

	case "ROUND_ROBIN":
		if len(standings) == 0 || len(standings[0].Rows) == 0 {
			return nil
		}
		rows := standings[0].Rows
		champ := &Champions{}
		if len(rows) >= 1 {
			p := rows[0].Participant
			champ.First = &p
		}
		if len(rows) >= 2 {
			p := rows[1].Participant
			champ.Second = &p
		}
		if len(rows) >= 3 {
			p := rows[2].Participant
			champ.Third = &p
		}
		return champ

	case "BATTLE_ROYALE":
		if len(brLeaderboard) == 0 {
			return nil
		}
		champ := &Champions{}
		if len(brLeaderboard) >= 1 {
			p := brLeaderboard[0].Participant
			champ.First = &p
		}
		if len(brLeaderboard) >= 2 {
			p := brLeaderboard[1].Participant
			champ.Second = &p
		}
		if len(brLeaderboard) >= 3 {
			p := brLeaderboard[2].Participant
			champ.Third = &p
		}
		return champ

	case "HYBRID_UCL":
		if stage.Type == "KNOCKOUT" || stage.Type == "SINGLE_ELIMINATION" || stage.SequenceOrder >= 2 {
			// Knockout stage champions follows Single/Double Elim logic
			for i := range matches {
				m := &matches[i]
				if (m.BracketType == "UPPER" || m.BracketType == "GRAND_FINAL") && m.Status == "COMPLETED" && m.WinnerID != nil {
					// Check if this is the final
					if m.NextMatchID == nil {
						champ := &Champions{}
						if p, ok := partMap[*m.WinnerID]; ok {
							champ.First = &p
						}
						if m.ParticipantAID != nil && *m.ParticipantAID != *m.WinnerID {
							if p, ok := partMap[*m.ParticipantAID]; ok {
								champ.Second = &p
							}
						} else if m.ParticipantBID != nil && *m.ParticipantBID != *m.WinnerID {
							if p, ok := partMap[*m.ParticipantBID]; ok {
								champ.Second = &p
							}
						}
						return champ
					}
				}
			}
		}
		return nil

	default:
		return nil
	}
}
