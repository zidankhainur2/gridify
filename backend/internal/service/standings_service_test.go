package service

import (
	"testing"

	"github.com/google/uuid"

	"github.com/gridify/backend/internal/domain"
)

func TestCalculateRoundRobinStandings(t *testing.T) {
	tournament := &domain.Tournament{
		PointsPerWin:  3,
		PointsPerDraw: 1,
	}

	pA := domain.Participant{ID: uuid.New(), Name: "Team A", SeedNumber: 1}
	pB := domain.Participant{ID: uuid.New(), Name: "Team B", SeedNumber: 2}
	pC := domain.Participant{ID: uuid.New(), Name: "Team C", SeedNumber: 3}
	participants := []domain.Participant{pA, pB, pC}

	matches := []domain.Match{
		// Match 1: A vs B -> 2-0 (A wins)
		{
			ParticipantAID: &pA.ID,
			ParticipantBID: &pB.ID,
			ScoreA:         2,
			ScoreB:         0,
			Status:         "COMPLETED",
		},
		// Match 2: B vs C -> 3-1 (B wins)
		{
			ParticipantAID: &pB.ID,
			ParticipantBID: &pC.ID,
			ScoreA:         3,
			ScoreB:         1,
			Status:         "COMPLETED",
		},
		// Match 3: C vs A -> 1-1 (Draw)
		{
			ParticipantAID: &pC.ID,
			ParticipantBID: &pA.ID,
			ScoreA:         1,
			ScoreB:         1,
			Status:         "COMPLETED",
		},
	}

	standings := CalculateRoundRobinStandings(tournament, participants, matches)
	if len(standings) != 1 {
		t.Fatalf("expected 1 group standing, got %d", len(standings))
	}

	rows := standings[0].Rows
	if len(rows) != 3 {
		t.Fatalf("expected 3 rows, got %d", len(rows))
	}

	// Rank 1: Team A (1W, 1D, 0L -> 4 pts)
	if rows[0].Participant.ID != pA.ID || rows[0].Points != 4 || rows[0].Rank != 1 {
		t.Errorf("Rank 1 should be Team A with 4 pts, got %s with %d pts", rows[0].Participant.Name, rows[0].Points)
	}

	// Rank 2: Team B (1W, 0D, 1L -> 3 pts)
	if rows[1].Participant.ID != pB.ID || rows[1].Points != 3 || rows[1].Rank != 2 {
		t.Errorf("Rank 2 should be Team B with 3 pts, got %s with %d pts", rows[1].Participant.Name, rows[1].Points)
	}

	// Rank 3: Team C (0W, 1D, 1L -> 1 pt)
	if rows[2].Participant.ID != pC.ID || rows[2].Points != 1 || rows[2].Rank != 3 {
		t.Errorf("Rank 3 should be Team C with 1 pt, got %s with %d pts", rows[2].Participant.Name, rows[2].Points)
	}
}

func TestCalculateBRLeaderboard(t *testing.T) {
	p1 := domain.Participant{ID: uuid.New(), Name: "Team 1", SeedNumber: 1}
	p2 := domain.Participant{ID: uuid.New(), Name: "Team 2", SeedNumber: 2}
	participants := []domain.Participant{p1, p2}

	match1ID := uuid.New()
	results := []domain.BattleRoyaleMatchResult{
		// Match 1: Team 1 wins (Rank 1, 10 placement pts, 2 kills = 12 total)
		{
			MatchID:         match1ID,
			ParticipantID:   p1.ID,
			PlacementRank:   1,
			Kills:           2,
			PlacementPoints: 10,
			KillPoints:      2,
			TotalPoints:     12,
		},
		// Match 1: Team 2 is 2nd (Rank 2, 6 placement pts, 10 kills = 16 total)
		{
			MatchID:         match1ID,
			ParticipantID:   p2.ID,
			PlacementRank:   2,
			Kills:           10,
			PlacementPoints: 6,
			KillPoints:      10,
			TotalPoints:     16,
		},
	}

	leaderboard := CalculateBRLeaderboard(participants, results)
	if len(leaderboard) != 2 {
		t.Fatalf("expected 2 teams in leaderboard, got %d", len(leaderboard))
	}

	// Team 2 should be Rank 1 (16 pts > 12 pts)
	if leaderboard[0].Participant.ID != p2.ID || leaderboard[0].TotalPoints != 16 {
		t.Errorf("expected Team 2 Rank 1 with 16 pts, got %s with %d pts", leaderboard[0].Participant.Name, leaderboard[0].TotalPoints)
	}

	// Team 1 should be Rank 2 (12 pts, 1 win)
	if leaderboard[1].Participant.ID != p1.ID || leaderboard[1].Wins != 1 {
		t.Errorf("expected Team 1 Rank 2 with 1 win, got %s with %d wins", leaderboard[1].Participant.Name, leaderboard[1].Wins)
	}
}

func TestDetermineChampions_SingleElimination(t *testing.T) {
	tournament := &domain.Tournament{
		FormatType:    "SINGLE_ELIMINATION",
		HasThirdPlace: true,
	}
	stage := &domain.Stage{Type: "PLAYOFFS"}

	p1 := domain.Participant{ID: uuid.New(), Name: "Winner"}
	p2 := domain.Participant{ID: uuid.New(), Name: "Runner Up"}
	p3 := domain.Participant{ID: uuid.New(), Name: "Bronze Winner"}
	p4 := domain.Participant{ID: uuid.New(), Name: "Bronze Loser"}
	participants := []domain.Participant{p1, p2, p3, p4}

	matches := []domain.Match{
		// Final match (Round 2 UPPER)
		{
			BracketType:    "UPPER",
			RoundNumber:    2,
			MatchNumber:    1,
			ParticipantAID: &p1.ID,
			ParticipantBID: &p2.ID,
			WinnerID:       &p1.ID,
			Status:         "COMPLETED",
		},
		// Bronze match (Round 2 THIRD_PLACE)
		{
			BracketType:    "THIRD_PLACE",
			RoundNumber:    2,
			MatchNumber:    1,
			ParticipantAID: &p3.ID,
			ParticipantBID: &p4.ID,
			WinnerID:       &p3.ID,
			Status:         "COMPLETED",
		},
	}

	champs := DetermineChampions(tournament, stage, participants, matches, nil, nil)
	if champs == nil {
		t.Fatalf("expected champions, got nil")
	}

	if champs.First == nil || champs.First.ID != p1.ID {
		t.Errorf("expected 1st place to be p1")
	}
	if champs.Second == nil || champs.Second.ID != p2.ID {
		t.Errorf("expected 2nd place to be p2")
	}
	if champs.Third == nil || champs.Third.ID != p3.ID {
		t.Errorf("expected 3rd place to be p3")
	}
}
