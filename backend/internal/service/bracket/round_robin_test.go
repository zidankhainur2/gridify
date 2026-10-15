package bracket

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/gridify/backend/internal/domain"
)

func TestGenerateRoundRobin_4Teams_SingleLeg(t *testing.T) {
	ctx := context.Background()
	store := &mockMatchStore{}
	tournament := &domain.Tournament{
		ID:         uuid.New(),
		FormatType: "ROUND_ROBIN",
		HomeAway:   false,
	}
	stage := &domain.Stage{ID: uuid.New(), TournamentID: tournament.ID}
	participants := makeParticipants(4)

	matches, err := GenerateRoundRobin(ctx, nil, store, tournament, stage, participants)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 4 teams, 1 leg -> N*(N-1)/2 = 6 matches
	if len(matches) != 6 {
		t.Fatalf("expected 6 matches, got %d", len(matches))
	}

	// 3 rounds, 2 matches per round
	roundCounts := make(map[int]int)
	for _, m := range matches {
		roundCounts[m.RoundNumber]++
	}
	if len(roundCounts) != 3 {
		t.Errorf("expected 3 rounds, got %d", len(roundCounts))
	}
	for r, count := range roundCounts {
		if count != 2 {
			t.Errorf("expected 2 matches in round %d, got %d", r, count)
		}
	}
}

func TestGenerateRoundRobin_4Teams_DoubleLeg_HomeAway(t *testing.T) {
	ctx := context.Background()
	store := &mockMatchStore{}
	tournament := &domain.Tournament{
		ID:         uuid.New(),
		FormatType: "ROUND_ROBIN",
		HomeAway:   true,
	}
	stage := &domain.Stage{ID: uuid.New(), TournamentID: tournament.ID}
	participants := makeParticipants(4)

	matches, err := GenerateRoundRobin(ctx, nil, store, tournament, stage, participants)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 4 teams, 2 legs -> N*(N-1) = 12 matches
	if len(matches) != 12 {
		t.Fatalf("expected 12 matches, got %d", len(matches))
	}

	// 6 rounds total
	roundCounts := make(map[int]int)
	for _, m := range matches {
		roundCounts[m.RoundNumber]++
	}
	if len(roundCounts) != 6 {
		t.Errorf("expected 6 rounds, got %d", len(roundCounts))
	}
}

func TestGenerateRoundRobin_5Teams_Odd(t *testing.T) {
	ctx := context.Background()
	store := &mockMatchStore{}
	tournament := &domain.Tournament{
		ID:         uuid.New(),
		FormatType: "ROUND_ROBIN",
		HomeAway:   false,
	}
	stage := &domain.Stage{ID: uuid.New(), TournamentID: tournament.ID}
	participants := makeParticipants(5)

	matches, err := GenerateRoundRobin(ctx, nil, store, tournament, stage, participants)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 5 teams -> 5 * 4 / 2 = 10 matches
	if len(matches) != 10 {
		t.Fatalf("expected 10 matches, got %d", len(matches))
	}

	// 5 rounds, 2 matches per round (1 team bye each round)
	roundCounts := make(map[int]int)
	for _, m := range matches {
		roundCounts[m.RoundNumber]++
	}
	if len(roundCounts) != 5 {
		t.Errorf("expected 5 rounds, got %d", len(roundCounts))
	}
	for r, count := range roundCounts {
		if count != 2 {
			t.Errorf("expected 2 matches in round %d, got %d", r, count)
		}
	}
}

func TestGenerateRoundRobin_HybridUCL_MultiGroup(t *testing.T) {
	ctx := context.Background()
	store := &mockMatchStore{}
	tournament := &domain.Tournament{
		ID:         uuid.New(),
		FormatType: "HYBRID_UCL",
		HomeAway:   false,
	}
	stage := &domain.Stage{ID: uuid.New(), TournamentID: tournament.ID}
	participants := makeParticipants(8) // 8 teams -> 2 groups of 4

	matches, err := GenerateRoundRobin(ctx, nil, store, tournament, stage, participants)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 2 groups * 6 matches each = 12 matches
	if len(matches) != 12 {
		t.Fatalf("expected 12 matches, got %d", len(matches))
	}

	// Verify all matches have unique (round_number, match_number)
	keys := make(map[string]bool)
	for _, m := range matches {
		k := string(rune('0'+m.RoundNumber)) + ":" + string(rune('0'+m.MatchNumber))
		if keys[k] {
			t.Errorf("duplicate round/match number: %s", k)
		}
		keys[k] = true
	}
}
