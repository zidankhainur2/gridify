package bracket

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/gridify/backend/internal/domain"
)

func TestGenerateDoubleElimination_4Teams(t *testing.T) {
	ctx := context.Background()
	store := &mockMatchStore{}
	tournament := &domain.Tournament{ID: uuid.New()}
	stage := &domain.Stage{ID: uuid.New(), TournamentID: tournament.ID}
	participants := makeParticipants(4)

	matches, err := GenerateDoubleElimination(ctx, nil, store, tournament, stage, participants)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 4 teams -> bracketSize 4
	// WB: R1 (2), R2 (1) = 3
	// LB: R1 (1), R2 (1) = 2
	// GF: 1
	// Total = 6 matches
	if len(matches) != 6 {
		t.Fatalf("expected 6 matches, got %d", len(matches))
	}

	var gf *domain.Match
	var wbFinal, lbFinal *domain.Match
	var wbR1_1, wbR1_2 *domain.Match
	var lbR1 *domain.Match

	for i := range matches {
		m := &matches[i]
		if m.BracketType == "GRAND_FINAL" {
			gf = m
		} else if m.BracketType == "UPPER" {
			if m.RoundNumber == 1 && m.MatchNumber == 1 {
				wbR1_1 = m
			} else if m.RoundNumber == 1 && m.MatchNumber == 2 {
				wbR1_2 = m
			} else if m.RoundNumber == 2 {
				wbFinal = m
			}
		} else if m.BracketType == "LOWER" {
			if m.RoundNumber == 1 {
				lbR1 = m
			} else if m.RoundNumber == 2 {
				lbFinal = m
			}
		}
	}

	if gf == nil || wbFinal == nil || lbFinal == nil || wbR1_1 == nil || wbR1_2 == nil || lbR1 == nil {
		t.Fatalf("missing required matches in 4-team double elim")
	}

	// Verify links
	// WB Final -> GF Slot A
	if wbFinal.NextMatchID == nil || *wbFinal.NextMatchID != gf.ID || *wbFinal.NextMatchSlot != "A" {
		t.Errorf("WB Final should link to GF Slot A")
	}
	// LB Final -> GF Slot B
	if lbFinal.NextMatchID == nil || *lbFinal.NextMatchID != gf.ID || *lbFinal.NextMatchSlot != "B" {
		t.Errorf("LB Final should link to GF Slot B")
	}
	// WB R1 losers -> LB R1
	if wbR1_1.LoserNextMatchID == nil || *wbR1_1.LoserNextMatchID != lbR1.ID || *wbR1_1.LoserNextMatchSlot != "A" {
		t.Errorf("WB R1 Match 1 loser should go to LB R1 Slot A")
	}
	if wbR1_2.LoserNextMatchID == nil || *wbR1_2.LoserNextMatchID != lbR1.ID || *wbR1_2.LoserNextMatchSlot != "B" {
		t.Errorf("WB R1 Match 2 loser should go to LB R1 Slot B")
	}
	// WB Final loser -> LB Final Slot B
	if wbFinal.LoserNextMatchID == nil || *wbFinal.LoserNextMatchID != lbFinal.ID || *wbFinal.LoserNextMatchSlot != "B" {
		t.Errorf("WB Final loser should go to LB Final Slot B")
	}
	// LB R1 winner -> LB Final Slot A
	if lbR1.NextMatchID == nil || *lbR1.NextMatchID != lbFinal.ID || *lbR1.NextMatchSlot != "A" {
		t.Errorf("LB R1 winner should go to LB Final Slot A")
	}
}

func TestGenerateDoubleElimination_8Teams(t *testing.T) {
	ctx := context.Background()
	store := &mockMatchStore{}
	tournament := &domain.Tournament{ID: uuid.New()}
	stage := &domain.Stage{ID: uuid.New(), TournamentID: tournament.ID}
	participants := makeParticipants(8)

	matches, err := GenerateDoubleElimination(ctx, nil, store, tournament, stage, participants)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 8 teams -> bracketSize 8
	// WB: R1 (4), R2 (2), R3 (1) = 7
	// LB: R1 (2), R2 (2), R3 (1), R4 (1) = 6
	// GF: 1
	// Total = 14 matches
	if len(matches) != 14 {
		t.Fatalf("expected 14 matches, got %d", len(matches))
	}

	// Check drop downs from WB R2 (semifinals) into LB R2
	var wbR2_1, wbR2_2 *domain.Match
	var lbR2_1, lbR2_2 *domain.Match
	for i := range matches {
		m := &matches[i]
		if m.BracketType == "UPPER" && m.RoundNumber == 2 {
			if m.MatchNumber == 1 {
				wbR2_1 = m
			} else if m.MatchNumber == 2 {
				wbR2_2 = m
			}
		}
		if m.BracketType == "LOWER" && m.RoundNumber == 2 {
			if m.MatchNumber == 1 {
				lbR2_1 = m
			} else if m.MatchNumber == 2 {
				lbR2_2 = m
			}
		}
	}

	if wbR2_1 == nil || wbR2_2 == nil || lbR2_1 == nil || lbR2_2 == nil {
		t.Fatalf("missing WB R2 or LB R2 matches")
	}

	if wbR2_1.LoserNextMatchID == nil || *wbR2_1.LoserNextMatchID != lbR2_1.ID || *wbR2_1.LoserNextMatchSlot != "B" {
		t.Errorf("WB R2 M1 loser should go to LB R2 M1 Slot B")
	}
	if wbR2_2.LoserNextMatchID == nil || *wbR2_2.LoserNextMatchID != lbR2_2.ID || *wbR2_2.LoserNextMatchSlot != "B" {
		t.Errorf("WB R2 M2 loser should go to LB R2 M2 Slot B")
	}
}

func TestGenerateDoubleElimination_5Teams_BYE(t *testing.T) {
	ctx := context.Background()
	store := &mockMatchStore{}
	tournament := &domain.Tournament{ID: uuid.New()}
	stage := &domain.Stage{ID: uuid.New(), TournamentID: tournament.ID}
	participants := makeParticipants(5)

	matches, err := GenerateDoubleElimination(ctx, nil, store, tournament, stage, participants)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// bracketSize 8 -> 14 matches
	if len(matches) != 14 {
		t.Fatalf("expected 14 matches, got %d", len(matches))
	}

	byeCount := 0
	for _, m := range matches {
		if m.BracketType == "UPPER" && m.RoundNumber == 1 && m.IsBye {
			byeCount++
		}
	}
	if byeCount != 3 {
		t.Errorf("expected 3 BYE matches in WB R1, got %d", byeCount)
	}
}

func TestGenerateDoubleElimination_6Teams(t *testing.T) {
	ctx := context.Background()
	store := &mockMatchStore{}
	tournament := &domain.Tournament{ID: uuid.New()}
	stage := &domain.Stage{ID: uuid.New(), TournamentID: tournament.ID}
	participants := makeParticipants(6)

	matches, err := GenerateDoubleElimination(ctx, nil, store, tournament, stage, participants)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 6 teams -> bracketSize 8 -> 14 matches
	if len(matches) != 14 {
		t.Fatalf("expected 14 matches, got %d", len(matches))
	}
}

func TestGenerateDoubleElimination_7Teams(t *testing.T) {
	ctx := context.Background()
	store := &mockMatchStore{}
	tournament := &domain.Tournament{ID: uuid.New()}
	stage := &domain.Stage{ID: uuid.New(), TournamentID: tournament.ID}
	participants := makeParticipants(7)

	matches, err := GenerateDoubleElimination(ctx, nil, store, tournament, stage, participants)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(matches) != 14 {
		t.Fatalf("expected 14 matches, got %d", len(matches))
	}
}

func TestGenerateDoubleElimination_12Teams(t *testing.T) {
	ctx := context.Background()
	store := &mockMatchStore{}
	tournament := &domain.Tournament{ID: uuid.New()}
	stage := &domain.Stage{ID: uuid.New(), TournamentID: tournament.ID}
	participants := makeParticipants(12)

	matches, err := GenerateDoubleElimination(ctx, nil, store, tournament, stage, participants)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 12 teams -> bracketSize 16 -> 30 matches
	if len(matches) != 30 {
		t.Fatalf("expected 30 matches, got %d", len(matches))
	}
}

func TestGenerateDoubleElimination_16Teams(t *testing.T) {
	ctx := context.Background()
	store := &mockMatchStore{}
	tournament := &domain.Tournament{ID: uuid.New()}
	stage := &domain.Stage{ID: uuid.New(), TournamentID: tournament.ID}
	participants := makeParticipants(16)

	matches, err := GenerateDoubleElimination(ctx, nil, store, tournament, stage, participants)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 16 teams -> bracketSize 16 -> 30 matches
	if len(matches) != 30 {
		t.Fatalf("expected 30 matches, got %d", len(matches))
	}

	// Verify all matches have unique (bracket_type, round_number, match_number)
	keys := make(map[string]bool)
	for _, m := range matches {
		k := m.BracketType + ":" + string(rune('0'+m.RoundNumber)) + ":" + string(rune('0'+m.MatchNumber))
		if keys[k] {
			t.Errorf("duplicate match key: %s", k)
		}
		keys[k] = true
	}
}
