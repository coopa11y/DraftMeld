package application

import (
	"testing"

	"github.com/coopa11y/DraftMeld/backend/internal/domain/draft"
	"github.com/coopa11y/DraftMeld/backend/internal/domain/league"
)

func TestReplacementValuesAllocateSuperflexToBestRemainingPosition(t *testing.T) {
	players := []draft.Player{
		{ID: "qb1", Position: "QB", ProjectedPoints: 400}, {ID: "qb2", Position: "QB", ProjectedPoints: 350}, {ID: "qb3", Position: "QB", ProjectedPoints: 300}, {ID: "qb4", Position: "QB", ProjectedPoints: 250},
		{ID: "rb1", Position: "RB", ProjectedPoints: 250}, {ID: "rb2", Position: "RB", ProjectedPoints: 200}, {ID: "rb3", Position: "RB", ProjectedPoints: 150},
	}
	rules := league.Rules{TeamCount: 2, AuctionBudget: 200, RosterSlots: []league.RosterSlot{
		{Name: "QB", Count: 1, Positions: []string{"QB"}, IsStarting: true},
		{Name: "RB", Count: 1, Positions: []string{"RB"}, IsStarting: true},
		{Name: "SUPERFLEX", Count: 1, Positions: []string{"QB", "RB"}, IsStarting: true},
	}}
	applyReplacementValues(players, rules)
	if players[0].ValueOverReplacement <= players[3].ValueOverReplacement {
		t.Fatalf("superflex demand should preserve quarterback value: %#v", players)
	}
	if players[0].AuctionValue <= 1 {
		t.Fatalf("positive VOR should produce an auction value: %#v", players[0])
	}
}

func TestNextUserPickHandlesSnakeTurns(t *testing.T) {
	rules := league.Rules{TeamCount: 12, DraftPosition: 1, DraftType: league.DraftTypeSnake}
	if got := nextUserPick(2, rules); got != 24 {
		t.Fatalf("next turn = %d, want 24", got)
	}
}
