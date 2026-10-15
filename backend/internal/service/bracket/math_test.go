package bracket_test

import (
	"reflect"
	"testing"

	"github.com/gridify/backend/internal/service/bracket"
)

func TestNextPowerOfTwo(t *testing.T) {
	cases := []struct {
		input    int
		expected int
	}{
		{0, 0},
		{1, 1},
		{2, 2},
		{3, 4},
		{4, 4},
		{5, 8},
		{6, 8},
		{7, 8},
		{8, 8},
		{9, 16},
		{15, 16},
		{16, 16},
		{17, 32},
		{31, 32},
		{32, 32},
		{33, 64},
		{64, 64},
	}

	for _, c := range cases {
		got := bracket.NextPowerOfTwo(c.input)
		if got != c.expected {
			t.Errorf("NextPowerOfTwo(%d) = %d, expected %d", c.input, got, c.expected)
		}
	}
}

func TestStandardSeedingOrder(t *testing.T) {
	// Size 2: [1, 2]
	s2 := bracket.StandardSeedingOrder(2)
	if !reflect.DeepEqual(s2, []int{1, 2}) {
		t.Errorf("expected [1, 2], got %v", s2)
	}

	// Size 4: [1, 4, 2, 3]
	s4 := bracket.StandardSeedingOrder(4)
	if !reflect.DeepEqual(s4, []int{1, 4, 2, 3}) {
		t.Errorf("expected [1, 4, 2, 3], got %v", s4)
	}

	// Size 8: [1, 8, 4, 5, 2, 7, 3, 6]
	s8 := bracket.StandardSeedingOrder(8)
	expected8 := []int{1, 8, 4, 5, 2, 7, 3, 6}
	if !reflect.DeepEqual(s8, expected8) {
		t.Errorf("expected %v, got %v", expected8, s8)
	}

	// Size 16: Check pairing sum = 17 for each match (pair)
	s16 := bracket.StandardSeedingOrder(16)
	if len(s16) != 16 {
		t.Fatalf("expected 16 seeds, got %d", len(s16))
	}
	for i := 0; i < len(s16); i += 2 {
		sum := s16[i] + s16[i+1]
		if sum != 17 {
			t.Errorf("seed pair (%d, %d) sum = %d, expected 17", s16[i], s16[i+1], sum)
		}
	}
}

func TestDistributeMultiSlotParticipants(t *testing.T) {
	type Team struct {
		Name string
	}
	input := []Team{
		{"Team A"},
		{"Team A"},
		{"Team B"},
		{"Team B"},
		{"Team C"},
	}

	output := bracket.DistributeMultiSlotParticipants(input, func(t Team) string {
		return t.Name
	})

	if len(output) != len(input) {
		t.Fatalf("expected %d teams, got %d", len(input), len(output))
	}

	// First two items should not be the same clan
	if output[0].Name == output[1].Name {
		t.Errorf("expected interleaved clan names, got consecutive %s", output[0].Name)
	}
}

func TestDistributeTeamsToGroups(t *testing.T) {
	teams := []string{"T1", "T2", "T3", "T4", "T5", "T6", "T7", "T8"}
	groups := bracket.DistributeTeamsToGroups(teams, 2)

	if len(groups) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(groups))
	}
	if len(groups[0]) != 4 || len(groups[1]) != 4 {
		t.Errorf("expected 4 teams per group, got %d and %d", len(groups[0]), len(groups[1]))
	}
}

func TestGenerateGroupPairs(t *testing.T) {
	groups := []string{"A", "B", "C"}
	pairs := bracket.GenerateGroupPairs(groups)

	// Comb(3, 2) = 3: [A, B], [A, C], [B, C]
	if len(pairs) != 3 {
		t.Fatalf("expected 3 pairs, got %d", len(pairs))
	}
	if pairs[0] != [2]string{"A", "B"} || pairs[1] != [2]string{"A", "C"} || pairs[2] != [2]string{"B", "C"} {
		t.Errorf("unexpected pairs: %v", pairs)
	}
}
