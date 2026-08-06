package application

import (
	"testing"

	"github.com/coopa11y/DraftMeld/backend/internal/domain/draft"
	"github.com/coopa11y/DraftMeld/backend/internal/domain/league"
)

func TestOpenStartingPositionsHonorsFlexibleSlots(t *testing.T) {
	slots := []league.RosterSlot{
		{Name: "RB", Count: 1, Positions: []string{"RB"}, IsStarting: true},
		{Name: "FLEX", Count: 1, Positions: []string{"RB", "WR"}, IsStarting: true},
		{Name: "Bench", Count: 4, Positions: []string{"RB", "WR"}, IsStarting: false},
	}

	needs := openStartingPositions(slots, []draft.Player{{Position: "WR"}})
	if !needs["RB"] || needs["WR"] {
		t.Fatalf("expected only RB to remain needed, got %#v", needs)
	}

	needs = openStartingPositions(slots, []draft.Player{{Position: "WR"}, {Position: "RB"}})
	if len(needs) != 0 {
		t.Fatalf("expected all starting slots to be filled, got %#v", needs)
	}
}
