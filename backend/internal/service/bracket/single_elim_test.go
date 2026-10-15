package bracket

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/gridify/backend/internal/domain"
)

type mockMatchStore struct {
	deletedStageID uuid.UUID
	inserted       []domain.Match
}

func (m *mockMatchStore) DeleteMatchesByStageID(ctx context.Context, tx pgx.Tx, stageID uuid.UUID) error {
	m.deletedStageID = stageID
	return nil
}

func (m *mockMatchStore) BatchInsertMatches(ctx context.Context, tx pgx.Tx, matches []domain.Match) error {
	m.inserted = matches
	return nil
}

func makeParticipants(n int) []domain.Participant {
	var list []domain.Participant
	for i := 1; i <= n; i++ {
		list = append(list, domain.Participant{
			ID:         uuid.New(),
			Name:       string(rune('A' + i - 1)),
			SeedNumber: i,
		})
	}
	return list
}

func TestGenerateSingleElimination_4Teams(t *testing.T) {
	ctx := context.Background()
	store := &mockMatchStore{}
	tournament := &domain.Tournament{
		ID:            uuid.New(),
		HasThirdPlace: false,
	}
	stage := &domain.Stage{
		ID:           uuid.New(),
		TournamentID: tournament.ID,
		Name:         "Playoffs",
	}
	participants := makeParticipants(4)

	matches, err := GenerateSingleElimination(ctx, nil, store, tournament, stage, participants)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 4 teams -> bracketSize 4 -> 2 rounds -> 2 + 1 = 3 matches
	if len(matches) != 3 {
		t.Fatalf("expected 3 matches, got %d", len(matches))
	}

	// Check R1 matches
	var r1Count, r2Count int
	for _, m := range matches {
		if m.RoundNumber == 1 {
			r1Count++
			if m.Status != "SCHEDULED" {
				t.Errorf("expected R1 status SCHEDULED, got %s", m.Status)
			}
			if m.NextMatchID == nil || m.NextMatchSlot == nil {
				t.Errorf("expected next match link for R1 match #%d", m.MatchNumber)
			}
		} else if m.RoundNumber == 2 {
			r2Count++
			if m.NextMatchID != nil {
				t.Errorf("Final should not have next_match_id")
			}
		}
	}
	if r1Count != 2 || r2Count != 1 {
		t.Errorf("expected 2 R1 and 1 R2 matches, got %d and %d", r1Count, r2Count)
	}
}

func TestGenerateSingleElimination_4Teams_WithThirdPlace(t *testing.T) {
	ctx := context.Background()
	store := &mockMatchStore{}
	tournament := &domain.Tournament{
		ID:            uuid.New(),
		HasThirdPlace: true,
	}
	stage := &domain.Stage{
		ID:           uuid.New(),
		TournamentID: tournament.ID,
	}
	participants := makeParticipants(4)

	matches, err := GenerateSingleElimination(ctx, nil, store, tournament, stage, participants)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 4 teams + third place -> 2 + 1 + 1 (third place) = 4 matches
	if len(matches) != 4 {
		t.Fatalf("expected 4 matches, got %d", len(matches))
	}

	var thirdPlace *domain.Match
	var r1Matches []*domain.Match
	for i := range matches {
		m := &matches[i]
		if m.BracketType == "THIRD_PLACE" {
			thirdPlace = m
		}
		if m.RoundNumber == 1 && m.BracketType == "UPPER" {
			r1Matches = append(r1Matches, m)
		}
	}

	if thirdPlace == nil {
		t.Fatalf("expected third place match")
	}

	// In 4-team bracket, Semifinals are Round 1
	for _, m := range r1Matches {
		if m.LoserNextMatchID == nil || *m.LoserNextMatchID != thirdPlace.ID {
			t.Errorf("R1 semifinal match #%d should link loser to third place match", m.MatchNumber)
		}
	}
}

func TestGenerateSingleElimination_5Teams_BYEAutoAdvance(t *testing.T) {
	ctx := context.Background()
	store := &mockMatchStore{}
	tournament := &domain.Tournament{
		ID:            uuid.New(),
		HasThirdPlace: false,
	}
	stage := &domain.Stage{
		ID:           uuid.New(),
		TournamentID: tournament.ID,
	}
	participants := makeParticipants(5) // Seeds 1..5

	matches, err := GenerateSingleElimination(ctx, nil, store, tournament, stage, participants)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 5 teams -> bracketSize 8 -> 3 rounds -> 4 + 2 + 1 = 7 matches
	if len(matches) != 7 {
		t.Fatalf("expected 7 matches, got %d", len(matches))
	}

	// Seeds in order(8): [1, 8, 4, 5, 2, 7, 3, 6]
	// Match 1: 1 vs 8(BYE) -> 1 auto-advances
	// Match 2: 4 vs 5 (Both present)
	// Match 3: 2 vs 7(BYE) -> 2 auto-advances
	// Match 4: 3 vs 6(BYE) -> 3 auto-advances
	byeCount := 0
	for _, m := range matches {
		if m.RoundNumber == 1 && m.IsBye {
			byeCount++
			if m.Status != "COMPLETED" {
				t.Errorf("BYE match should be COMPLETED, got %s", m.Status)
			}
			if m.WinnerID == nil {
				t.Errorf("BYE match should have WinnerID set")
			}
		}
	}
	if byeCount != 3 {
		t.Errorf("expected 3 BYE matches in Round 1, got %d", byeCount)
	}

	// Verify Round 2 has auto-advanced participants
	var r2Match1, r2Match2 *domain.Match
	for i := range matches {
		m := &matches[i]
		if m.RoundNumber == 2 && m.MatchNumber == 1 {
			r2Match1 = m
		}
		if m.RoundNumber == 2 && m.MatchNumber == 2 {
			r2Match2 = m
		}
	}

	if r2Match1 == nil || r2Match2 == nil {
		t.Fatalf("expected R2 match 1 and 2 to exist")
	}

	// Seed 1 should be in R2 Match 1 Slot A
	if r2Match1.ParticipantAID == nil || *r2Match1.ParticipantAID != participants[0].ID {
		t.Errorf("expected Seed 1 to auto-advance to R2 Match 1 Slot A")
	}
	// Seed 2 and Seed 3 should be in R2 Match 2 Slot A and Slot B
	if r2Match2.ParticipantAID == nil || *r2Match2.ParticipantAID != participants[1].ID {
		t.Errorf("expected Seed 2 in R2 Match 2 Slot A")
	}
	if r2Match2.ParticipantBID == nil || *r2Match2.ParticipantBID != participants[2].ID {
		t.Errorf("expected Seed 3 in R2 Match 2 Slot B")
	}
}

func TestGenerateSingleElimination_8Teams(t *testing.T) {
	ctx := context.Background()
	store := &mockMatchStore{}
	tournament := &domain.Tournament{
		ID:            uuid.New(),
		HasThirdPlace: true,
	}
	stage := &domain.Stage{
		ID:           uuid.New(),
		TournamentID: tournament.ID,
		Name:         "Playoffs 8 Teams",
	}
	participants := makeParticipants(8)

	matches, err := GenerateSingleElimination(ctx, nil, store, tournament, stage, participants)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 8 teams -> 4 QF + 2 SF + 1 Final + 1 Third Place = 8 matches
	if len(matches) != 8 {
		t.Fatalf("expected 8 matches for 8 teams with 3rd place, got %d", len(matches))
	}

	for _, m := range matches {
		if m.RoundNumber == 1 && m.Status != "SCHEDULED" {
			t.Errorf("expected R1 status SCHEDULED, got %s", m.Status)
		}
	}
}

func TestGenerateSingleElimination_16Teams(t *testing.T) {
	ctx := context.Background()
	store := &mockMatchStore{}
	tournament := &domain.Tournament{
		ID:            uuid.New(),
		HasThirdPlace: false,
	}
	stage := &domain.Stage{
		ID:           uuid.New(),
		TournamentID: tournament.ID,
		Name:         "Playoffs 16 Teams",
	}
	participants := makeParticipants(16)

	matches, err := GenerateSingleElimination(ctx, nil, store, tournament, stage, participants)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 16 teams -> 8 R1 + 4 QF + 2 SF + 1 Final = 15 matches
	if len(matches) != 15 {
		t.Fatalf("expected 15 matches for 16 teams, got %d", len(matches))
	}
}

func TestGenerateSingleElimination_InvalidParticipants(t *testing.T) {
	ctx := context.Background()
	store := &mockMatchStore{}
	tournament := &domain.Tournament{ID: uuid.New()}
	stage := &domain.Stage{ID: uuid.New()}

	_, err := GenerateSingleElimination(ctx, nil, store, tournament, stage, makeParticipants(1))
	if err == nil {
		t.Errorf("expected error for 1 participant, got nil")
	}
}

