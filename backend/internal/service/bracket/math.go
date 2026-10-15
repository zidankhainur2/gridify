package bracket

import (
	"math"
	"math/bits"
	"strings"
)

// NextPowerOfTwo returns the smallest power of two greater than or equal to n.
// For n <= 1, it returns 1. For n <= 0, it returns 0.
func NextPowerOfTwo(n int) int {
	if n <= 0 {
		return 0
	}
	if n == 1 {
		return 1
	}
	// Using bits.LeadingZeros to compute next power of 2 efficiently
	return 1 << (bits.Len(uint(n - 1)))
}

// StandardSeedingOrder generates standard tournament seed distribution (e.g. 1 vs 8, 4 vs 5, 2 vs 7, 3 vs 6).
// Size must be a power of two (e.g. 2, 4, 8, 16, 32, 64).
func StandardSeedingOrder(size int) []int {
	if size <= 0 {
		return []int{}
	}
	if size == 1 {
		return []int{1}
	}
	if size == 2 {
		return []int{1, 2}
	}

	rounds := int(math.Round(math.Log2(float64(size))))
	order := []int{1, 2}

	for i := 0; i < rounds-1; i++ {
		next := make([]int, 0, len(order)*2)
		sum := len(order)*2 + 1
		for _, val := range order {
			next = append(next, val, sum-val)
		}
		order = next
	}
	return order
}

// DistributeMultiSlotParticipants distributes participants to prevent same-clan/same-name
// collisions in the first round (anti-team kill).
func DistributeMultiSlotParticipants[T any](participants []T, getName func(T) string) []T {
	if len(participants) == 0 {
		return []T{}
	}

	groups := make(map[string][]T)
	var keys []string

	for _, p := range participants {
		nameKey := strings.ToLower(strings.TrimSpace(getName(p)))
		if len(groups[nameKey]) == 0 {
			keys = append(keys, nameKey)
		}
		groups[nameKey] = append(groups[nameKey], p)
	}

	maxCount := 0
	for _, k := range keys {
		if len(groups[k]) > maxCount {
			maxCount = len(groups[k])
		}
	}

	result := make([]T, 0, len(participants))
	for i := 0; i < maxCount; i++ {
		for _, k := range keys {
			if i < len(groups[k]) {
				result = append(result, groups[k][i])
			}
		}
	}

	return result
}

// DistributeTeamsToGroups distributes teams into N groups evenly (e.g. for Hybrid UCL group stage).
func DistributeTeamsToGroups[T any](teams []T, groupCount int) [][]T {
	if groupCount <= 0 {
		return [][]T{}
	}
	groups := make([][]T, groupCount)
	for i := range groups {
		groups[i] = make([]T, 0)
	}

	for i, team := range teams {
		idx := i % groupCount
		groups[idx] = append(groups[idx], team)
	}
	return groups
}

// GenerateGroupPairs returns all unique 2-group combinations (e.g. for multi-lobby Battle Royale).
func GenerateGroupPairs(groups []string) [][2]string {
	var pairs [][2]string
	for i := 0; i < len(groups); i++ {
		for j := i + 1; j < len(groups); j++ {
			pairs = append(pairs, [2]string{groups[i], groups[j]})
		}
	}
	return pairs
}
