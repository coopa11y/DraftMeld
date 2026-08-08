package application

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/coopa11y/DraftMeld/backend/internal/domain/draft"
	"github.com/coopa11y/DraftMeld/backend/internal/domain/league"
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
	if afterDraft.History[0].TeamNumber != 1 || afterDraft.History[0].TeamName != "My Team" || len(afterDraft.Teams[0].Roster) != 1 {
		t.Fatalf("first pick was not assigned to the named user team: %#v", afterDraft)
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
	if afterTaken.History[1].TeamNumber != 2 || afterTaken.History[1].TeamName != "Team 2" || len(afterTaken.Teams[1].Roster) != 1 {
		t.Fatalf("opponent pick was not assigned to Team 2: %#v", afterTaken)
	}

	afterUndo, err := service.Undo(ctx, "demo")
	if err != nil {
		t.Fatalf("undo action: %v", err)
	}
	if len(afterUndo.History) != 1 || !containsPlayer(afterUndo.Available, "p002") {
		t.Fatalf("undo did not restore player: %#v", afterUndo)
	}
}

func TestDraftEnforcesPickOwnershipAndCompletion(t *testing.T) {
	store, err := draftsqlite.Open(t.TempDir() + "/draftmeld.db")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	configuration := DemoLeagueConfiguration()
	configuration.Rules.TeamCount = 2
	configuration.Rules.DraftPosition = 1
	configuration.Rules.TeamNames = []string{"Marcus", "Opponent"}
	configuration.Rules.RosterSlots = []league.RosterSlot{{Name: "FLEX", Count: 1, Positions: []string{"RB", "WR"}, IsStarting: true}}
	service, err := NewDraftService(store, draft.DemoCatalog(), configuration)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = service.RecordForTeam(t.Context(), "demo", "p001", draft.ActionTaken, 0, 2); err == nil {
		t.Fatal("expected an opponent action to be rejected on the user's pick")
	}
	if _, err = service.RecordForTeam(t.Context(), "demo", "p001", draft.ActionDraft, 0, 1); err != nil {
		t.Fatal(err)
	}
	complete, err := service.RecordForTeam(t.Context(), "demo", "p002", draft.ActionTaken, 0, 2)
	if err != nil {
		t.Fatal(err)
	}
	if !complete.IsComplete || complete.TotalPicks != 2 || len(complete.Teams[0].Roster) != 1 || len(complete.Teams[1].Roster) != 1 {
		t.Fatalf("draft did not complete with owned rosters: %#v", complete)
	}
	if _, err = service.RecordForTeam(t.Context(), "demo", "p003", draft.ActionDraft, 0, 1); !errors.Is(err, ErrDraftComplete) {
		t.Fatalf("expected completed draft error, got %v", err)
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
	after, err := service.Record(context.Background(), "demo", before.Recommendations[0].Player.ID, draft.ActionDraft)
	if err != nil {
		t.Fatalf("draft recommendation: %v", err)
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
	result, err := service.SyncSleeper(t.Context(), "demo", "draft-1", 7)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Snapshot.MyTeam) != 1 || result.Snapshot.MyTeam[0].ID != "p001" {
		t.Fatalf("Sleeper pick was not assigned to my team: %#v", result.Snapshot.MyTeam)
	}
}

func TestSleeperSyncReconcilesChangedAndDeletedPicks(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		requestCount++
		if requestCount == 1 {
			_, _ = response.Write([]byte(`[{"pick_no":1,"roster_id":7,"metadata":{"first_name":"Alex","last_name":"Rivers","position":"RB","team":"ATL"}}]`))
			return
		}
		_, _ = response.Write([]byte(`[{"pick_no":1,"roster_id":2,"metadata":{"first_name":"Jordan","last_name":"Hale","position":"WR","team":"MIN"}},{"pick_no":2,"roster_id":2,"metadata":{"first_name":"Unknown","last_name":"Player","position":"WR","team":"MIN"}}]`))
	}))
	defer server.Close()
	store, err := draftsqlite.Open(t.TempDir() + "/draftmeld.db")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	service := newTestDraftService(t, store)
	service.ConfigureSleeperClient(server.Client(), server.URL)
	if _, err = service.SyncSleeper(t.Context(), "demo", "draft-1", 7); err != nil {
		t.Fatal(err)
	}
	result, err := service.SyncSleeper(t.Context(), "demo", "draft-1", 7)
	if err != nil {
		t.Fatal(err)
	}
	if result.Added != 1 || result.Removed != 1 || result.Unmatched != 1 || len(result.Snapshot.History) != 1 || result.Snapshot.History[0].Player.ID != "p002" {
		t.Fatalf("unexpected reconciliation result: %#v", result)
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
