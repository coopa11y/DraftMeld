package league

import "testing"

func TestRulesValidate(t *testing.T) {
	rules := Rules{
		Name:      "Home League",
		TeamCount: 12,
		DraftType: DraftTypeSnake,
		RosterSlots: []RosterSlot{
			{Name: "QB", Count: 1, Positions: []string{"QB"}, IsStarting: true},
		},
	}
	if err := rules.Validate(); err != nil {
		t.Fatalf("expected valid rules, got %v", err)
	}
}
