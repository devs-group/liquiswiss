package mcp

import "testing"

func TestFuzzyScoreTiers(t *testing.T) {
	cases := []struct {
		needle, haystack string
		min, max         float64
	}{
		{"serverkosten", "Serverkosten", 1.0, 1.0},
		{"server", "Serverkosten", 0.9, 0.9},
		{"kosten", "Serverkosten", 0.75, 0.75},
		{"claude florian", "Claude Florian", 1.0, 1.0},
		{"claude flor", "Claude Florian", 0.9, 0.9},
		{"florian claude", "Claude Florian", 0.7, 0.7},
		{"adrain", "Adrian F. (100%)", 0.4, 0.5},
		{"bexo", "Bexio", 0.5, 0.6},
		{"xyz", "Serverkosten", 0, 0},
		{"", "Serverkosten", 0, 0},
	}
	for _, testCase := range cases {
		score := fuzzyScore(testCase.needle, testCase.haystack)
		if score < testCase.min || score > testCase.max {
			t.Errorf("fuzzyScore(%q, %q) = %v, want in [%v, %v]", testCase.needle, testCase.haystack, score, testCase.min, testCase.max)
		}
	}
}

func TestFuzzyScoreOrdering(t *testing.T) {
	exact := fuzzyScore("bexio", "Bexio")
	typo := fuzzyScore("bexo", "Bexio")
	if exact <= typo {
		t.Errorf("exact match must outrank typo match: %v <= %v", exact, typo)
	}
}
