package mcp

import (
	"sort"
	"strings"
)

// fuzzyScore rates how well needle matches haystack, 0 (no match) to 1 (exact).
// Tiers: exact > prefix > substring > all-words-contained > typo distance >
// subsequence. Case-insensitive.
func fuzzyScore(needle, haystack string) float64 {
	needle = strings.ToLower(strings.TrimSpace(needle))
	haystack = strings.ToLower(haystack)
	if needle == "" {
		return 0
	}
	if haystack == needle {
		return 1.0
	}
	if strings.HasPrefix(haystack, needle) {
		return 0.9
	}
	if strings.Contains(haystack, needle) {
		return 0.75
	}

	// All query words appear somewhere (order-independent multi-word match)
	words := strings.Fields(needle)
	if len(words) > 1 {
		all := true
		for _, word := range words {
			if !strings.Contains(haystack, word) {
				all = false
				break
			}
		}
		if all {
			return 0.7
		}
	}

	// Typo tolerance against the full string and against each word
	best := 0.0
	if score := typoScore(needle, haystack); score > best {
		best = score
	}
	for _, hayWord := range strings.Fields(haystack) {
		if score := typoScore(needle, hayWord); score > best {
			best = score
		}
	}
	if best > 0 {
		return best
	}

	// Subsequence: characters of needle appear in order
	if isSubsequence(needle, haystack) {
		return 0.3 * float64(len(needle)) / float64(len(haystack))
	}

	return 0
}

// typoScore returns a score based on edit distance when the strings are of
// comparable length and at most 2 edits apart
func typoScore(needle, candidate string) float64 {
	if len(needle) < 3 {
		return 0
	}
	distance := levenshtein(needle, candidate)
	maxEdits := 1
	if len(needle) >= 5 {
		maxEdits = 2
	}
	if distance > maxEdits {
		return 0
	}
	return 0.65 - 0.1*float64(distance)
}

func isSubsequence(needle, haystack string) bool {
	i := 0
	for j := 0; j < len(haystack) && i < len(needle); j++ {
		if needle[i] == haystack[j] {
			i++
		}
	}
	return i == len(needle)
}

func levenshtein(a, b string) int {
	if len(a) == 0 {
		return len(b)
	}
	if len(b) == 0 {
		return len(a)
	}
	previous := make([]int, len(b)+1)
	current := make([]int, len(b)+1)
	for j := range previous {
		previous[j] = j
	}
	for i := 1; i <= len(a); i++ {
		current[0] = i
		for j := 1; j <= len(b); j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			current[j] = min(previous[j]+1, min(current[j-1]+1, previous[j-1]+cost))
		}
		previous, current = current, previous
	}
	return previous[len(b)]
}

type scored struct {
	item  map[string]any
	score float64
}

// fuzzyRank scores items by the given text extractor, drops non-matches, sorts
// by confidence descending and returns items as maps with a matchScore field
func fuzzyRank[T any](items []T, search string, textOf func(T) string) ([]map[string]any, error) {
	ranked := make([]scored, 0, len(items))
	for _, item := range items {
		score := fuzzyScore(search, textOf(item))
		if score <= 0 {
			continue
		}
		asMap, err := toMap(item)
		if err != nil {
			return nil, err
		}
		asMap["matchScore"] = score
		ranked = append(ranked, scored{item: asMap, score: score})
	}
	sort.SliceStable(ranked, func(i, j int) bool { return ranked[i].score > ranked[j].score })
	result := make([]map[string]any, len(ranked))
	for i, entry := range ranked {
		result[i] = entry.item
	}
	return result, nil
}
