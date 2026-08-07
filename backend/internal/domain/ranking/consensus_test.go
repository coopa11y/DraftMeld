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

func TestCombineNormalizesDifferentListDepths(t *testing.T) {
	eligible := []string{"a", "b", "c", "d"}
	entries, err := Combine([]Source{
		{ID: "short", Role: "ranking", Weight: 1, Ranks: map[string]int{"a": 1, "b": 2}},
		{ID: "deep", Role: "ranking", Weight: 1, Ranks: map[string]int{"a": 1, "b": 50, "c": 75, "d": 100}},
	}, eligible, MethodWeightedAverage)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 4 || entries[0].PlayerID != "a" {
		t.Fatalf("unexpected normalized consensus: %#v", entries)
	}
	if entries[3].PlayerID != "d" {
		t.Fatalf("missing primary ranks should be conservative: %#v", entries)
	}
}

func TestWeightedMedianResistsOneExtremeContextSignal(t *testing.T) {
	entries, err := Combine([]Source{
		{ID: "rank-a", Role: "ranking", Weight: 1, Ranks: map[string]int{"a": 10, "b": 11}},
		{ID: "rank-b", Role: "ranking", Weight: 1, Ranks: map[string]int{"a": 11, "b": 10}},
		{ID: "market", Role: "market", Weight: .2, Ranks: map[string]int{"b": 1}},
	}, []string{"a", "b"}, MethodWeightedMedian)
	if err != nil {
		t.Fatal(err)
	}
	if entries[0].SourceCount != 2 || entries[1].SourceCount != 3 {
		t.Fatalf("context coverage should only count present values: %#v", entries)
	}
}

func TestGoldenConsensusMethodsKeepStableTopFour(t *testing.T) {
	sources := []Source{
		{ID: "a", Role: "ranking", Weight: 1, Ranks: map[string]int{"p1": 1, "p2": 2, "p3": 3, "p4": 4}},
		{ID: "b", Role: "ranking", Weight: 1, Ranks: map[string]int{"p1": 1, "p2": 3, "p3": 2, "p4": 4}},
		{ID: "c", Role: "ranking", Weight: 1, Ranks: map[string]int{"p1": 4, "p2": 1, "p3": 2, "p4": 3}},
		{ID: "d", Role: "ranking", Weight: 1, Ranks: map[string]int{"p1": 1, "p2": 2, "p3": 4, "p4": 3}},
	}
	for _, method := range []string{MethodWeightedAverage, MethodWeightedMedian, MethodTrimmedMean} {
		entries, err := Combine(sources, []string{"p1", "p2", "p3", "p4"}, method)
		if err != nil {
			t.Fatalf("%s: %v", method, err)
		}
		got := []string{entries[0].PlayerID, entries[1].PlayerID, entries[2].PlayerID, entries[3].PlayerID}
		want := []string{"p1", "p2", "p3", "p4"}
		for index := range want {
			if got[index] != want[index] {
				t.Fatalf("%s order changed: got %v want %v", method, got, want)
			}
		}
	}
}
