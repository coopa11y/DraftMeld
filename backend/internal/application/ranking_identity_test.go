package application

import "testing"

func TestCanonicalRankingKeyMergesDefenseAliases(t *testing.T) {
	tests := []struct {
		name, position, team, expected string
	}{
		{name: "Denver Defense", position: "DST", expected: "dstden"},
		{name: "Broncos D/ST", position: "D/ST", expected: "dstden"},
		{name: "Unknown Defense", position: "DEF", team: "DEN", expected: "dstden"},
		{name: "Jacksonville Jaguars", position: "DST", team: "JAC", expected: "dstjax"},
		{name: "Denver Runner", position: "RB", team: "DEN", expected: "denverrunner"},
	}
	for _, test := range tests {
		if actual := canonicalRankingKey(test.name, test.position, test.team); actual != test.expected {
			t.Errorf("canonicalRankingKey(%q, %q, %q) = %q, want %q", test.name, test.position, test.team, actual, test.expected)
		}
	}
}
