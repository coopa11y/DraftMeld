package league

import "testing"

func TestRulesCloneOwnsMutableCollections(t *testing.T) {
	original := Rules{
		TeamNames:         []string{"One", "Two"},
		DraftOrder:        []int{1, 2},
		RosterSlots:       []RosterSlot{{Name: "FLEX", Positions: []string{"RB", "WR"}}},
		ScoringRules:      map[string]float64{"reception": 1},
		SourcePreferences: map[string]RankingSourcePreference{"source": {Weight: 1, Enabled: true}},
		PlayerPreferences: map[string]string{"player": "target"},
	}
	cloned := original.Clone()
	cloned.TeamNames[0] = "Changed"
	cloned.DraftOrder[0] = 2
	cloned.RosterSlots[0].Positions[0] = "QB"
	cloned.ScoringRules["reception"] = 0
	cloned.SourcePreferences["source"] = RankingSourcePreference{Weight: 2}
	cloned.PlayerPreferences["player"] = "avoid"

	if original.TeamNames[0] != "One" || original.DraftOrder[0] != 1 || original.RosterSlots[0].Positions[0] != "RB" ||
		original.ScoringRules["reception"] != 1 || original.SourcePreferences["source"].Weight != 1 || original.PlayerPreferences["player"] != "target" {
		t.Fatalf("clone mutated its source: %#v", original)
	}
}

func TestRulesValidate(t *testing.T) {
	rules := Rules{
		Name:          "Home League",
		TeamCount:     12,
		DraftPosition: 4,
		DraftType:     DraftTypeSnake,
		RosterSlots: []RosterSlot{
			{Name: "QB", Count: 1, Positions: []string{"QB"}, IsStarting: true},
		},
	}
	if err := rules.Validate(); err != nil {
		t.Fatalf("expected valid rules, got %v", err)
	}
}

func TestRulesAllowDraftPositionToRemainUnassigned(t *testing.T) {
	rules := Rules{
		Name: "Home League", TeamCount: 12, DraftPosition: 0, DraftType: DraftTypeSnake,
		RosterSlots: []RosterSlot{{Name: "QB", Count: 1, Positions: []string{"QB"}, IsStarting: true}},
	}
	if err := rules.Validate(); err != nil {
		t.Fatalf("expected an unassigned draft position to be valid during setup, got %v", err)
	}
}

func TestRulesRejectInvalidSourcePreferences(t *testing.T) {
	rules := Rules{
		Name: "Home League", TeamCount: 12, DraftPosition: 4, DraftType: DraftTypeSnake,
		RosterSlots:       []RosterSlot{{Name: "QB", Count: 1, Positions: []string{"QB"}, IsStarting: true}},
		SourcePreferences: map[string]RankingSourcePreference{"cbs-ppr": {Weight: 0, Enabled: false}},
	}
	if err := rules.Validate(); err == nil {
		t.Fatal("expected a zero source weight to be rejected")
	}
}

func TestRulesRequireAnEnabledRankingSource(t *testing.T) {
	rules := Rules{
		Name: "Home League", TeamCount: 12, DraftPosition: 4, DraftType: DraftTypeSnake,
		RosterSlots:       []RosterSlot{{Name: "QB", Count: 1, Positions: []string{"QB"}, IsStarting: true}},
		SourcePreferences: map[string]RankingSourcePreference{"cbs-ppr": {Weight: 1, Enabled: false}},
	}
	if err := rules.Validate(); err == nil {
		t.Fatal("expected all ranking sources disabled to be rejected")
	}
}
