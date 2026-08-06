package ranking

import "testing"

func TestWeightedAverage(t *testing.T) {
	entries, err := WeightedAverage([]Source{
		{ID: "expert-a", Weight: 2, Ranks: map[string]int{"player-a": 1, "player-b": 2}},
		{ID: "expert-b", Weight: 1, Ranks: map[string]int{"player-a": 3, "player-b": 1}},
	})
	if err != nil {
		t.Fatalf("weighted average failed: %v", err)
	}
	if len(entries) != 2 || entries[0].PlayerID != "player-a" {
		t.Fatalf("unexpected consensus: %#v", entries)
	}
}
