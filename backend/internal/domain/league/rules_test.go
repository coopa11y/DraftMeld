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

func TestRulesRejectNonPositiveSourceWeight(t *testing.T) {
	rules := Rules{
		Name: "Home League", TeamCount: 12, DraftPosition: 4, DraftType: DraftTypeSnake,
		RosterSlots:   []RosterSlot{{Name: "QB", Count: 1, Positions: []string{"QB"}, IsStarting: true}},
		SourceWeights: map[string]float64{"cbs-ppr": 0},
	}
	if err := rules.Validate(); err == nil {
		t.Fatal("expected a zero source weight to be rejected")
	}
}
