package application

import (
	"context"
	"testing"

	"github.com/coopa11y/DraftMeld/backend/internal/domain/draft"
	draftsqlite "github.com/coopa11y/DraftMeld/backend/internal/persistence/sqlite"
)

func TestDraftTakenAndUndo(t *testing.T) {
	store, err := draftsqlite.Open(t.TempDir() + "/draftmeld.db")
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	defer store.Close()
	service := newTestDraftService(t, store)
	ctx := context.Background()

	afterDraft, err := service.Record(ctx, "demo", "p001", draft.ActionDraft)
	if err != nil {
		t.Fatalf("draft player: %v", err)
	}
	if len(afterDraft.MyTeam) != 1 || afterDraft.MyTeam[0].ID != "p001" {
		t.Fatalf("player was not added to team: %#v", afterDraft.MyTeam)
	}
	if len(afterDraft.Available) != len(draft.DemoCatalog())-1 {
		t.Fatalf("drafted player remained available")
	}

	afterTaken, err := service.Record(ctx, "demo", "p002", draft.ActionTaken)
	if err != nil {
		t.Fatalf("mark player taken: %v", err)
	}
	if len(afterTaken.MyTeam) != 1 || len(afterTaken.History) != 2 {
		t.Fatalf("unexpected state after taken action: %#v", afterTaken)
	}

	afterUndo, err := service.Undo(ctx, "demo")
	if err != nil {
		t.Fatalf("undo action: %v", err)
	}
	if len(afterUndo.History) != 1 || !containsPlayer(afterUndo.Available, "p002") {
		t.Fatalf("undo did not restore player: %#v", afterUndo)
	}
}

func TestRecommendationsReactToRosterNeed(t *testing.T) {
	store, err := draftsqlite.Open(t.TempDir() + "/draftmeld.db")
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	defer store.Close()
	service := newTestDraftService(t, store)
	before, err := service.Snapshot(context.Background(), "demo")
	if err != nil {
		t.Fatalf("load snapshot: %v", err)
	}
	if len(before.Recommendations) != 5 {
		t.Fatalf("expected five recommendations, got %d", len(before.Recommendations))
	}
	after, err := service.Record(context.Background(), "demo", before.Recommendations[0].Player.ID, draft.ActionTaken)
	if err != nil {
		t.Fatalf("mark recommendation taken: %v", err)
	}
	if after.Recommendations[0].Player.ID == before.Recommendations[0].Player.ID {
		t.Fatalf("taken player remained recommended")
	}
}

func TestLeagueConfigurationControlsSnapshotAndRecommendationLimit(t *testing.T) {
	store, err := draftsqlite.Open(t.TempDir() + "/draftmeld.db")
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	defer store.Close()
	configuration := DemoLeagueConfiguration()
	configuration.ID = "custom"
	configuration.Rules.Name = "Custom League"
	configuration.Recommendation.RecommendationLimit = 2
	service, err := NewDraftService(store, draft.DemoCatalog(), configuration)
	if err != nil {
		t.Fatalf("create service: %v", err)
	}
	snapshot, err := service.Snapshot(context.Background(), "custom")
	if err != nil {
		t.Fatalf("load snapshot: %v", err)
	}
	if snapshot.LeagueName != "Custom League" || len(snapshot.Recommendations) != 2 {
		t.Fatalf("configuration was not applied: %#v", snapshot)
	}
}

func newTestDraftService(t *testing.T, repository DraftEventRepository) *DraftService {
	t.Helper()
	service, err := NewDraftService(repository, draft.DemoCatalog(), DemoLeagueConfiguration())
	if err != nil {
		t.Fatalf("create draft service: %v", err)
	}
	return service
}

func containsPlayer(players []draft.Player, playerID string) bool {
	for _, player := range players {
		if player.ID == playerID {
			return true
		}
	}
	return false
}
