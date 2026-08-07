package application

import (
	"context"
	"net/http"
	"net/http/httptest"
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

func TestMockDraftAdvancesToNextUserTurn(t *testing.T) {
	store, err := draftsqlite.Open(t.TempDir() + "/draftmeld.db")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	service := newTestDraftService(t, store)
	if _, err = service.Record(t.Context(), "demo", "p001", draft.ActionDraft); err != nil {
		t.Fatal(err)
	}
	snapshot, err := service.MockToNextTurn(t.Context(), "demo")
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.PickNumber != 24 || len(snapshot.History) != 23 {
		t.Fatalf("mock stopped at pick %d with %d events", snapshot.PickNumber, len(snapshot.History))
	}
}

func TestSleeperSyncImportsKnownPlayersWithoutSubmittingPicks(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet {
			t.Fatalf("unexpected method %s", request.Method)
		}
		_, _ = response.Write([]byte(`[{"pick_no":1,"roster_id":7,"metadata":{"first_name":"Alex","last_name":"Rivers","position":"RB","team":"ATL"}}]`))
	}))
	defer server.Close()
	store, err := draftsqlite.Open(t.TempDir() + "/draftmeld.db")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	service := newTestDraftService(t, store)
	service.sleeperBaseURL, service.sleeperClient = server.URL, server.Client()
	snapshot, err := service.SyncSleeper(t.Context(), "demo", "draft-1", 7)
	if err != nil {
		t.Fatal(err)
	}
	if len(snapshot.MyTeam) != 1 || snapshot.MyTeam[0].ID != "p001" {
		t.Fatalf("Sleeper pick was not assigned to my team: %#v", snapshot.MyTeam)
	}
}

func TestAuctionActionsTrackBudget(t *testing.T) {
	store, err := draftsqlite.Open(t.TempDir() + "/draftmeld.db")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	configuration := DemoLeagueConfiguration()
	configuration.Rules.DraftType = "auction"
	service, err := NewDraftService(store, draft.DemoCatalog(), configuration)
	if err != nil {
		t.Fatal(err)
	}
	snapshot, err := service.Record(t.Context(), "demo", "p001", draft.ActionDraft, 37)
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.BudgetRemaining != 163 || snapshot.History[0].Cost != 37 {
		t.Fatalf("auction budget not tracked: %#v", snapshot)
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
