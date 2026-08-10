package application

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
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
	configuration.Rules.UserTeamNumber = 1
	configuration.Rules.TeamNames = []string{"Marcus", "Opponent"}
	configuration.Rules.RosterSlots = []league.RosterSlot{{Name: "FLEX", Count: 1, Positions: []string{"RB", "WR"}, IsStarting: true}}
	service, err := NewDraftService(store, draft.DemoCatalog(), configuration)
	if err != nil {
		t.Fatal(err)
	}
	startTestDraft(t, service, "demo")
	if _, err = service.RecordForTeam(t.Context(), "demo", "p001", draft.ActionDraft, 0, 2); err == nil {
		t.Fatal("expected a user draft action to be rejected when the pick is assigned to an opponent")
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

func TestDraftSupportsTradedPickOwnership(t *testing.T) {
	store, err := draftsqlite.Open(t.TempDir() + "/draftmeld.db")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	configuration := DemoLeagueConfiguration()
	configuration.Rules.TeamCount = 2
	configuration.Rules.DraftPosition = 2
	configuration.Rules.UserTeamNumber = 2
	configuration.Rules.TeamNames = []string{"Team 1", "Marcus"}
	configuration.Rules.RosterSlots = []league.RosterSlot{{Name: "FLEX", Count: 1, Positions: []string{"RB", "WR"}, IsStarting: true}}
	service, err := NewDraftService(store, draft.DemoCatalog(), configuration)
	if err != nil {
		t.Fatal(err)
	}
	startTestDraft(t, service, "demo")

	afterTrade, err := service.RecordForTeam(t.Context(), "demo", "p001", draft.ActionDraft, 0, 2)
	if err != nil {
		t.Fatalf("record pick traded to user: %v", err)
	}
	if afterTrade.History[0].TeamNumber != 2 || afterTrade.History[0].TeamName != "Marcus" || len(afterTrade.MyTeam) != 1 {
		t.Fatalf("traded pick was not assigned to its new owner: %#v", afterTrade)
	}
}

func TestDraftPickTradeUpdatesAnyUnusedRoundAndPersists(t *testing.T) {
	store, err := draftsqlite.Open(t.TempDir() + "/draftmeld.db")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	configuration := DemoLeagueConfiguration()
	configuration.Rules.DraftPosition = 12
	configuration.Rules.UserTeamNumber = 12
	configuration.Rules.RosterSlots = []league.RosterSlot{{Name: "Bench", Count: 2, Positions: []string{"RB", "WR"}, IsStarting: false}}
	service, err := NewDraftService(store, draft.DemoCatalog(), configuration)
	if err != nil {
		t.Fatal(err)
	}
	startTestDraft(t, service, "demo")

	afterTrade, err := service.CreatePickTrade(t.Context(), "demo", 12, 1, []int{1}, []int{12, 13}, nil, nil, 0, 0)
	if err != nil {
		t.Fatalf("create multi-pick trade: %v", err)
	}
	if len(afterTrade.PickTrades) != 1 || afterTrade.OnClockTeamNumber != 12 || !afterTrade.IsUserTurn {
		t.Fatalf("trade did not update the current pick: %#v", afterTrade)
	}
	for pick, owner := range map[int]int{1: 12, 12: 1, 13: 1} {
		if afterTrade.PickSlots[pick-1].OwnerTeamNumber != owner {
			t.Fatalf("pick %d owner = %d, want %d", pick, afterTrade.PickSlots[pick-1].OwnerTeamNumber, owner)
		}
	}

	afterPick, err := service.Record(t.Context(), "demo", "p001", draft.ActionDraft)
	if err != nil {
		t.Fatalf("use traded current pick: %v", err)
	}
	if afterPick.History[0].TeamNumber != 12 || afterPick.OnClockTeamNumber != 2 {
		t.Fatalf("unexpected ownership after traded pick: %#v", afterPick)
	}
	if _, err = service.CreatePickTrade(t.Context(), "demo", 1, 12, []int{1}, []int{24}, nil, nil, 0, 0); err == nil {
		t.Fatal("expected an already-used pick to be rejected")
	}

	reloaded, err := NewDraftService(store, draft.DemoCatalog(), configuration)
	if err != nil {
		t.Fatal(err)
	}
	persisted, err := reloaded.Snapshot(t.Context(), "demo")
	if err != nil || len(persisted.PickTrades) != 1 || persisted.PickSlots[12].OwnerTeamNumber != 1 {
		t.Fatalf("trade did not persist: snapshot=%#v error=%v", persisted, err)
	}
}

func TestDynastyFuturePickTradeSurvivesSeasonChange(t *testing.T) {
	store, err := draftsqlite.Open(t.TempDir() + "/draftmeld.db")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	configuration := DemoLeagueConfiguration()
	configuration.Rules.LeagueFormat = league.LeagueFormatDynasty
	configuration.Rules.Season = 2026
	configuration.Rules.FuturePickSeasons = 2
	configuration.Rules.RookieDraftRounds = 4
	if err = store.SaveLeague(t.Context(), configuration); err != nil {
		t.Fatal(err)
	}
	service, err := NewDraftServiceWithLeagues(store, store, draft.DemoCatalog())
	if err != nil {
		t.Fatal(err)
	}
	startTestDraft(t, service, "demo")
	future := draft.FuturePick{Season: 2027, Round: 1, OriginalTeamNumber: 2}
	afterTrade, err := service.CreatePickTrade(t.Context(), "demo", 1, 2, nil, []int{1}, []draft.FuturePick{future}, nil, 0, 0)
	if err != nil {
		t.Fatalf("trade future dynasty pick: %v", err)
	}
	if owner := futureSlotOwner(afterTrade.PickSlots, 2027, 1, 2); owner != 1 {
		t.Fatalf("future pick owner = %d, want 1", owner)
	}
	if _, err = service.RecordForTeam(t.Context(), "demo", "p001", draft.ActionTaken, 0, 2); err != nil {
		t.Fatalf("record 2026 pick: %v", err)
	}

	configuration.Rules.Season = 2027
	if err = store.SaveLeague(t.Context(), configuration); err != nil {
		t.Fatal(err)
	}
	nextSeason, err := service.Snapshot(t.Context(), "demo")
	if err != nil {
		t.Fatal(err)
	}
	if nextSeason.PickSlots[1].OriginalTeamNumber != 2 || nextSeason.PickSlots[1].OwnerTeamNumber != 1 {
		t.Fatalf("future trade was not applied to the new season: %#v", nextSeason.PickSlots[1])
	}
	if len(nextSeason.History) != 0 || nextSeason.PickNumber != 1 {
		t.Fatalf("prior-season draft events leaked into 2027: history=%d pick=%d", len(nextSeason.History), nextSeason.PickNumber)
	}
}

func TestAuctionLeagueTradesConfiguredBudgetAndFuturePicks(t *testing.T) {
	store, err := draftsqlite.Open(t.TempDir() + "/draftmeld.db")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	configuration := DemoLeagueConfiguration()
	configuration.Rules.DraftType = league.DraftTypeAuction
	configuration.Rules.LeagueFormat = league.LeagueFormatDynasty
	configuration.Rules.FuturePickSeasons = 2
	configuration.Rules.RookieDraftRounds = 4
	configuration.Rules.AuctionBudgetTrades = true
	if err = store.SaveLeague(t.Context(), configuration); err != nil {
		t.Fatal(err)
	}
	service, err := NewDraftServiceWithLeagues(store, store, draft.DemoCatalog())
	if err != nil {
		t.Fatal(err)
	}
	future := draft.FuturePick{Season: 2027, Round: 1, OriginalTeamNumber: 2}
	snapshot, err := service.CreatePickTrade(t.Context(), "demo", 1, 2, nil, nil, []draft.FuturePick{future}, nil, 0, 25)
	if err != nil {
		t.Fatalf("trade auction assets: %v", err)
	}
	if snapshot.Teams[0].AuctionBudgetRemaining != 175 || snapshot.Teams[1].AuctionBudgetRemaining != 225 {
		t.Fatalf("unexpected auction budgets: %#v", snapshot.Teams[:2])
	}
	if owner := futureSlotOwner(snapshot.PickSlots, 2027, 1, 2); owner != 1 {
		t.Fatalf("future auction-league pick owner = %d, want 1", owner)
	}
}

func TestTradeAssetsRespectLeagueRules(t *testing.T) {
	store, err := draftsqlite.Open(t.TempDir() + "/draftmeld.db")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	configuration := DemoLeagueConfiguration()
	if err = store.SaveLeague(t.Context(), configuration); err != nil {
		t.Fatal(err)
	}
	service, err := NewDraftServiceWithLeagues(store, store, draft.DemoCatalog())
	if err != nil {
		t.Fatal(err)
	}

	future := []draft.FuturePick{{Season: 2027, Round: 1, OriginalTeamNumber: 2}}
	if _, err = service.CreatePickTrade(t.Context(), "demo", 1, 2, nil, nil, future, nil, 0, 0); err == nil || !strings.Contains(err.Error(), "redraft") {
		t.Fatalf("redraft future pick error = %v", err)
	}

	configuration.Rules.DraftType = league.DraftTypeAuction
	if err = store.SaveLeague(t.Context(), configuration); err != nil {
		t.Fatal(err)
	}
	if _, err = service.CreatePickTrade(t.Context(), "demo", 1, 2, nil, nil, nil, nil, 10, 0); err == nil || !strings.Contains(err.Error(), "not enabled") {
		t.Fatalf("disabled auction budget error = %v", err)
	}
}

func TestDynastyFranchisesPlayersAndDraftOrderPersistAcrossRollover(t *testing.T) {
	store, err := draftsqlite.Open(t.TempDir() + "/draftmeld.db")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	configuration := DemoLeagueConfiguration()
	configuration.Rules.TeamCount = 2
	configuration.Rules.TeamNames = []string{"Marcus", "Rival"}
	configuration.Rules.UserTeamNumber = 1
	configuration.Rules.DraftOrder = []int{2, 1}
	configuration.Rules.DraftPosition = 2
	configuration.Rules.LeagueFormat = league.LeagueFormatDynasty
	configuration.Rules.FuturePickSeasons = 2
	configuration.Rules.RookieDraftRounds = 2
	configuration.Rules.RosterSlots = []league.RosterSlot{{Name: "Bench", Count: 2, Positions: []string{"RB", "WR"}}}
	if err = store.SaveLeague(t.Context(), configuration); err != nil {
		t.Fatal(err)
	}
	service, err := NewDraftServiceWithLeagues(store, store, draft.DemoCatalog())
	if err != nil {
		t.Fatal(err)
	}
	startTestDraft(t, service, "demo")
	if _, err = service.RecordForTeam(t.Context(), "demo", "p002", draft.ActionTaken, 0, 2); err != nil {
		t.Fatalf("record Rival pick from first draft slot: %v", err)
	}
	if _, err = service.RecordForTeam(t.Context(), "demo", "p001", draft.ActionDraft, 0, 1); err != nil {
		t.Fatalf("record Marcus pick from second draft slot: %v", err)
	}
	snapshot, err := service.CreateDraftTrade(t.Context(), "demo", 1, 2, nil, nil, nil, nil, []string{"p002"}, []string{"p001"}, nil, nil, 0, 0)
	if err != nil {
		t.Fatalf("trade dynasty players: %v", err)
	}
	if len(snapshot.Teams[0].Roster) != 1 || snapshot.Teams[0].Roster[0].ID != "p002" || snapshot.Teams[0].Name != "Marcus" {
		t.Fatalf("permanent Marcus franchise roster = %#v", snapshot.Teams[0])
	}

	next, err := service.AdvanceSeason(t.Context(), "demo", 2027, league.DraftTypeLinear, []int{1, 2})
	if err != nil {
		t.Fatalf("advance dynasty season: %v", err)
	}
	if next.Season != 2027 || next.OnClockTeamNumber != 0 || next.SessionStatus != draft.SessionNotStarted || next.TotalPicks != 4 || len(next.History) != 0 {
		t.Fatalf("unexpected next-season draft: %#v", next)
	}
	if len(next.Teams[0].Roster) != 1 || next.Teams[0].Roster[0].ID != "p002" || containsPlayer(next.Available, "p001") || containsPlayer(next.Available, "p002") {
		t.Fatalf("dynasty rosters did not carry forward: teams=%#v available=%#v", next.Teams, next.Available)
	}
}

func TestDynastyFutureFAABAndConditionalPickLifecycle(t *testing.T) {
	store, err := draftsqlite.Open(t.TempDir() + "/draftmeld.db")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	configuration := DemoLeagueConfiguration()
	configuration.Rules.LeagueFormat = league.LeagueFormatDynasty
	configuration.Rules.FuturePickSeasons = 2
	configuration.Rules.RookieDraftRounds = 2
	configuration.Rules.FAABBudget = 100
	configuration.Rules.FAABTrades = true
	if err = store.SaveLeague(t.Context(), configuration); err != nil {
		t.Fatal(err)
	}
	service, err := NewDraftServiceWithLeagues(store, store, draft.DemoCatalog())
	if err != nil {
		t.Fatal(err)
	}
	conditional := draft.FuturePick{Season: 2027, Round: 1, OriginalTeamNumber: 2, Condition: "Player appears in eight games"}
	snapshot, err := service.CreateDraftTrade(t.Context(), "demo", 1, 2, nil, nil, []draft.FuturePick{conditional}, nil, nil, nil,
		[]draft.BudgetAsset{{Kind: "faab", Season: 2027, Amount: 25}}, nil, 0, 0)
	if err != nil {
		t.Fatalf("trade conditional pick and future FAAB: %v", err)
	}
	trade := snapshot.PickTrades[0]
	if trade.TeamOneFuturePicks[0].ConditionStatus != "pending" || futureSlotOwner(snapshot.PickSlots, 2027, 1, 2) != 1 {
		t.Fatalf("conditional pick was not locked to recipient: %#v", trade)
	}
	if _, err = service.CreateDraftTrade(t.Context(), "demo", 1, 3, nil, nil, nil, []draft.FuturePick{conditional}, nil, nil, nil, nil, 0, 0); err == nil || !strings.Contains(err.Error(), "unresolved condition") {
		t.Fatalf("pending conditional pick was tradeable: %v", err)
	}
	resolved, err := service.ResolveTradeCondition(t.Context(), "demo", trade.ID, 2027, 1, 2, "not-met")
	if err != nil {
		t.Fatalf("resolve condition: %v", err)
	}
	if futureSlotOwner(resolved.PickSlots, 2027, 1, 2) != 2 {
		t.Fatal("unmet conditional pick did not return to its prior franchise")
	}
	next, err := service.AdvanceSeason(t.Context(), "demo", 2027, league.DraftTypeSnake, configuration.Rules.DraftOrder)
	if err != nil {
		t.Fatalf("advance season: %v", err)
	}
	if remainingBudget(next.BudgetBalances, 1, 2027, "faab") != 125 || remainingBudget(next.BudgetBalances, 2, 2027, "faab") != 75 {
		t.Fatalf("future FAAB did not carry forward: %#v", next.BudgetBalances)
	}
}

func remainingBudget(balances []draft.BudgetBalance, team, season int, kind string) float64 {
	for _, balance := range balances {
		if balance.TeamNumber == team && balance.Season == season && balance.Kind == kind {
			return balance.Remaining
		}
	}
	return -1
}

func futureSlotOwner(slots []draft.PickSlot, season, round, originalTeam int) int {
	for _, slot := range slots {
		if slot.Season == season && slot.Round == round && slot.OriginalTeamNumber == originalTeam {
			return slot.OwnerTeamNumber
		}
	}
	return 0
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
	startTestDraft(t, service, "demo")
	snapshot, err := service.Record(t.Context(), "demo", "p001", draft.ActionDraft, 37)
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.BudgetRemaining != 163 || snapshot.History[0].Cost != 37 {
		t.Fatalf("auction budget not tracked: %#v", snapshot)
	}
}

func TestDraftSessionStartResetAndRestore(t *testing.T) {
	store, err := draftsqlite.Open(t.TempDir() + "/draftmeld.db")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	configuration := DemoLeagueConfiguration()
	configuration.Rules.TeamCount = 2
	configuration.Rules.TeamNames = []string{"Marcus", "Opponent"}
	configuration.Rules.DraftOrder = []int{1, 2}
	configuration.Rules.RosterSlots = []league.RosterSlot{{Name: "FLEX", Count: 1, Positions: []string{"RB", "WR"}, IsStarting: true}}
	service, err := NewDraftService(store, draft.DemoCatalog(), configuration)
	if err != nil {
		t.Fatal(err)
	}

	initial, err := service.Snapshot(t.Context(), "demo")
	if err != nil || initial.SessionStatus != draft.SessionNotStarted || initial.CanReset {
		t.Fatalf("unexpected initial session: snapshot=%#v error=%v", initial, err)
	}
	if _, err = service.RecordForTeam(t.Context(), "demo", "p001", draft.ActionDraft, 0, 1); !errors.Is(err, ErrDraftNotStarted) {
		t.Fatalf("record before start error = %v", err)
	}
	started, err := service.StartDraft(t.Context(), "demo")
	if err != nil || started.SessionStatus != draft.SessionInProgress || !started.CanReset {
		t.Fatalf("unexpected started session: snapshot=%#v error=%v", started, err)
	}
	if _, err = service.RecordForTeam(t.Context(), "demo", "p001", draft.ActionDraft, 0, 1); err != nil {
		t.Fatal(err)
	}
	complete, err := service.RecordForTeam(t.Context(), "demo", "p002", draft.ActionTaken, 0, 2)
	if err != nil || complete.SessionStatus != draft.SessionComplete {
		t.Fatalf("unexpected completed session: snapshot=%#v error=%v", complete, err)
	}
	if _, err = service.ResetDraft(t.Context(), "demo", "wrong name"); err == nil {
		t.Fatal("expected an exact-name reset confirmation")
	}
	reset, err := service.ResetDraft(t.Context(), "demo", configuration.Rules.Name)
	if err != nil || reset.SessionStatus != draft.SessionNotStarted || len(reset.History) != 0 || !reset.CanUndoReset {
		t.Fatalf("unexpected reset session: snapshot=%#v error=%v", reset, err)
	}
	restored, err := service.UndoDraftReset(t.Context(), "demo")
	if err != nil || restored.SessionStatus != draft.SessionComplete || len(restored.History) != 2 || restored.CanUndoReset {
		t.Fatalf("unexpected restored session: snapshot=%#v error=%v", restored, err)
	}
	if _, err = service.ResetDraft(t.Context(), "demo", configuration.Rules.Name); err != nil {
		t.Fatal(err)
	}
	if _, err = service.StartDraft(t.Context(), "demo"); err != nil {
		t.Fatal(err)
	}
	if _, err = service.UndoDraftReset(t.Context(), "demo"); !errors.Is(err, ErrNoResetToUndo) {
		t.Fatalf("undo after restart error = %v", err)
	}
}

func TestDraftCannotStartUntilDraftPositionIsAssigned(t *testing.T) {
	store, err := draftsqlite.Open(t.TempDir() + "/draftmeld.db")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	configuration := DemoLeagueConfiguration()
	configuration.Rules.DraftPosition = 0
	service, err := NewDraftService(store, draft.DemoCatalog(), configuration)
	if err != nil {
		t.Fatal(err)
	}

	if _, err = service.StartDraft(t.Context(), "demo"); !errors.Is(err, ErrDraftPositionUnassigned) {
		t.Fatalf("start without a draft position error = %v", err)
	}
}

func TestLeagueDraftStructureLocksAfterStart(t *testing.T) {
	store, err := draftsqlite.Open(t.TempDir() + "/draftmeld.db")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	configuration := DemoLeagueConfiguration()
	if err = store.SaveLeague(t.Context(), configuration); err != nil {
		t.Fatal(err)
	}
	service, err := NewDraftServiceWithLeagues(store, store, draft.DemoCatalog())
	if err != nil {
		t.Fatal(err)
	}
	startTestDraft(t, service, "demo")
	leagues := NewLeagueService(store)

	structural := configuration.Rules
	structural.DraftType = league.DraftTypeLinear
	if _, err = leagues.Update(t.Context(), "demo", structural); !errors.Is(err, ErrInvalidLeague) {
		t.Fatalf("structural update error = %v", err)
	}
	safe := configuration.Rules
	safe.Name = "Renamed League"
	if _, err = leagues.Update(t.Context(), "demo", safe); err != nil {
		t.Fatalf("safe live update failed: %v", err)
	}
}

func newTestDraftService(t *testing.T, repository DraftEventRepository) *DraftService {
	t.Helper()
	service, err := NewDraftService(repository, draft.DemoCatalog(), DemoLeagueConfiguration())
	if err != nil {
		t.Fatalf("create draft service: %v", err)
	}
	if _, ok := repository.(DraftSessionRepository); ok {
		startTestDraft(t, service, "demo")
	}
	return service
}

func startTestDraft(t *testing.T, service *DraftService, leagueID string) {
	t.Helper()
	if _, err := service.StartDraft(t.Context(), leagueID); err != nil {
		t.Fatalf("start draft: %v", err)
	}
}

func containsPlayer(players []draft.Player, playerID string) bool {
	for _, player := range players {
		if player.ID == playerID {
			return true
		}
	}
	return false
}
