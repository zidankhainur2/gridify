package bracket

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/gridify/backend/internal/domain"
)

func TestGenerateBattleRoyale_SingleLobby(t *testing.T) {
	ctx := context.Background()
	store := &mockMatchStore{}
	tournament := &domain.Tournament{
		ID:                 uuid.New(),
		FormatType:         "BATTLE_ROYALE",
		BRTotalMatches:     6,
		BRMaxTeamsPerMatch: 16,
	}
	stage := &domain.Stage{ID: uuid.New(), TournamentID: tournament.ID}
	participants := makeParticipants(12)

	matches, err := GenerateBattleRoyale(ctx, nil, store, tournament, stage, participants)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(matches) != 6 {
		t.Fatalf("expected 6 matches, got %d", len(matches))
	}

	for i, m := range matches {
		if m.RoundNumber != i+1 {
			t.Errorf("match %d expected round %d, got %d", i, i+1, m.RoundNumber)
		}
		if m.MatchNumber != 1 {
			t.Errorf("expected match number 1, got %d", m.MatchNumber)
		}
		if m.BracketType != "BATTLE_ROYALE" {
			t.Errorf("expected BATTLE_ROYALE bracket type, got %s", m.BracketType)
		}
		if len(m.PlayingGroups) != 0 {
			t.Errorf("expected empty playing groups for single lobby")
		}
		if m.MapName == nil || *m.MapName == "" {
			t.Errorf("expected map name to be set")
		}
	}
}

func TestGenerateBattleRoyale_MultiGroup(t *testing.T) {
	ctx := context.Background()
	store := &mockMatchStore{}
	tournament := &domain.Tournament{
		ID:                 uuid.New(),
		FormatType:         "BATTLE_ROYALE",
		BRTotalMatches:     6,
		BRMaxTeamsPerMatch: 8, // small limit to trigger multi-group with 16 teams
	}
	stage := &domain.Stage{ID: uuid.New(), TournamentID: tournament.ID}
	participants := makeParticipants(16)

	matches, err := GenerateBattleRoyale(ctx, nil, store, tournament, stage, participants)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(matches) != 6 {
		t.Fatalf("expected 6 matches, got %d", len(matches))
	}

	for i, m := range matches {
		if len(m.PlayingGroups) != 2 {
			t.Errorf("match %d expected 2 playing groups, got %d", i, len(m.PlayingGroups))
		}
	}
}
