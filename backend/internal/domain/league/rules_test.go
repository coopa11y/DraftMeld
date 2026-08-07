package league

import "testing"

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
