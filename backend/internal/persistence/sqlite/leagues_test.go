package sqlite

import (
	"context"
	"testing"

	"github.com/coopa11y/DraftMeld/backend/internal/domain/draft"
	"github.com/coopa11y/DraftMeld/backend/internal/domain/league"
)

func TestLeagueConfigurationPersistsAndDeletesItsDraft(t *testing.T) {
	path := t.TempDir() + "/draftmeld.db"
	store, err := Open(path)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	configuration := testLeagueConfiguration()
	if err = store.SaveLeague(t.Context(), configuration); err != nil {
		t.Fatalf("save league: %v", err)
	}
	if _, err = store.Append(t.Context(), draft.Event{LeagueID: configuration.ID, PlayerID: "p001", Action: draft.ActionDraft}); err != nil {
		t.Fatalf("append draft event: %v", err)
	}
	if err = store.Close(); err != nil {
		t.Fatalf("close database: %v", err)
	}

	store, err = Open(path)
	if err != nil {
		t.Fatalf("reopen database: %v", err)
	}
	defer store.Close()
	loaded, found, err := store.GetLeague(context.Background(), configuration.ID)
	if err != nil {
		t.Fatalf("get league: %v", err)
	}
	if !found || loaded.Rules.DraftPosition != 7 || len(loaded.Rules.RosterSlots) != 2 || loaded.Rules.ScoringRules["reception"] != 0.5 {
		t.Fatalf("unexpected persisted league: %#v", loaded)
	}

	deleted, err := store.DeleteLeague(t.Context(), configuration.ID)
	if err != nil || !deleted {
		t.Fatalf("delete league: deleted=%v err=%v", deleted, err)
	}
	events, err := store.List(t.Context(), configuration.ID)
	if err != nil {
		t.Fatalf("list deleted league events: %v", err)
	}
	if len(events) != 0 {
		t.Fatalf("expected draft history to be deleted, got %#v", events)
	}
}

func testLeagueConfiguration() league.Configuration {
	return league.Configuration{
		ID: "league-a",
		Rules: league.Rules{
			Name: "League A", TeamCount: 12, DraftPosition: 7, DraftType: league.DraftTypeSnake,
			RosterSlots: []league.RosterSlot{
				{Name: "QB", Count: 1, Positions: []string{"QB"}, IsStarting: true},
				{Name: "Bench", Count: 5, Positions: []string{"QB", "RB", "WR", "TE"}, IsStarting: false},
			},
			ScoringRules: map[string]float64{"reception": 0.5},
		},
		Recommendation: league.RecommendationPolicy{
			BaseScore: 200, StartingNeedBonus: 24, ADPValueThreshold: 5,
			ScarcityBonus: 8, ScarcityDropOff: 5, RecommendationLimit: 5,
		},
	}
}
