package application

import (
	"strings"
	"testing"
	"time"

	"github.com/coopa11y/DraftMeld/backend/internal/domain/draft"
	"github.com/coopa11y/DraftMeld/backend/internal/domain/league"
	"github.com/coopa11y/DraftMeld/backend/internal/domain/ranking"
	draftsqlite "github.com/coopa11y/DraftMeld/backend/internal/persistence/sqlite"
)

func TestGoldenRecommendationScenarioHonorsNeedValueTargetsAndAvoids(t *testing.T) {
	policy := DefaultRecommendationPolicy()
	players := []draft.Player{
		{ID: "target-rb", Name: "Target RB", Position: "RB", OverallRank: 4, PositionRank: 2, ADP: 15, ValueOverReplacement: 18, Preference: "target"},
		{ID: "elite-qb", Name: "Elite QB", Position: "QB", OverallRank: 1, PositionRank: 1, ADP: 2, ValueOverReplacement: 22},
		{ID: "avoid-wr", Name: "Avoid WR", Position: "WR", OverallRank: 2, PositionRank: 1, ADP: 20, ValueOverReplacement: 25, Preference: "avoid"},
	}
	rules := league.Rules{DraftType: league.DraftTypeSnake, RosterSlots: []league.RosterSlot{{Name: "RB", Count: 1, Positions: []string{"RB"}, IsStarting: true}}}
	result := recommend(players, nil, rules, policy, recommendationContext{NextUserPick: 12})
	if result[0].Player.ID != "target-rb" {
		t.Fatalf("golden recommendation changed: got %s, want target-rb", result[0].Player.ID)
	}
	if result[len(result)-1].Player.ID != "avoid-wr" {
		t.Fatalf("avoid preference should remain last: %#v", result)
	}
}

func TestGoldenAuctionScenarioAccountsForRosterReserveAndKeepers(t *testing.T) {
	rules := league.Rules{
		TeamCount: 2, DraftType: league.DraftTypeAuction, AuctionBudget: 100, AuctionMinimumBid: 2,
		KeeperBudgetSpent: 40, KeeperValueRemoved: 25,
		RosterSlots: []league.RosterSlot{{Name: "RB", Count: 2, Positions: []string{"RB"}, IsStarting: true}},
	}
	players := []draft.Player{{ID: "rb1", Position: "RB", ProjectedPoints: 100}, {ID: "rb2", Position: "RB", ProjectedPoints: 80}, {ID: "rb3", Position: "RB", ProjectedPoints: 60}, {ID: "rb4", Position: "RB", ProjectedPoints: 40}, {ID: "rb5", Position: "RB", ProjectedPoints: 20}}
	applyReplacementValues(players, rules)
	budget, inflation, maximumBid := auctionState(rules, nil, players, 0)
	if budget != 100 || maximumBid != 98 {
		t.Fatalf("roster reserve changed: budget=%.0f maximum=%.0f", budget, maximumBid)
	}
	if inflation <= 0 || inflation >= 1 {
		t.Fatalf("keeper spend/value should calibrate the remaining market, inflation=%.3f", inflation)
	}
}

func TestIdentityMergeCombinesAliasesAcrossRankingsAndProjections(t *testing.T) {
	store, err := draftsqlite.Open(t.TempDir() + "/draftmeld.db")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	rankingService := NewRankingService(store)
	now := time.Now().UTC()
	redraft := ranking.SourceDefinition{ID: "redraft-ecr", Name: "Redraft", Role: "ranking", DefaultWeight: 1}
	if err = store.ReplaceRankings(t.Context(), redraft, []ranking.Record{
		{SourceID: "redraft-ecr", PlayerKey: "nathanieldell", Name: "Nathaniel Dell", Position: "WR", Team: "HOU", Rank: 10},
		{SourceID: "redraft-ecr", PlayerKey: "tankdell", Name: "Tank Dell", Position: "WR", Team: "HOU", Rank: 11},
	}, "test", now); err != nil {
		t.Fatal(err)
	}
	issues, err := rankingService.IdentityIssues(t.Context())
	if err != nil || len(issues) != 1 {
		t.Fatalf("expected one identity issue: %#v %v", issues, err)
	}
	if err = rankingService.ReviewIdentity(t.Context(), issues[0].IssueKey, "merged", "tankdell"); err != nil {
		t.Fatal(err)
	}
	issues, err = rankingService.IdentityIssues(t.Context())
	if err != nil || issues[0].Resolution != "merged" || issues[0].CanonicalPlayerKey != "tankdell" {
		t.Fatalf("merge was not persisted: %#v %v", issues, err)
	}
	projectionService := NewProjectionService(store)
	if _, err = projectionService.ImportCSV(t.Context(), "Alias projections", strings.NewReader("name,position,team,reception\nNathaniel Dell,WR,HOU,80\n")); err != nil {
		t.Fatal(err)
	}
	values, err := projectionService.LeagueValues(t.Context(), map[string]float64{"reception": 1})
	if err != nil || values["tankdell"].ProjectedPoints != 80 {
		t.Fatalf("projection alias did not resolve: %#v %v", values, err)
	}
	if _, err = projectionService.ImportCSV(t.Context(), "Canonical alias projections", strings.NewReader("name,position,team,reception\nTank Dell,WR,HOU,90\n")); err != nil {
		t.Fatal(err)
	}
	values, err = projectionService.LeagueValues(t.Context(), map[string]float64{"reception": 1})
	if err != nil || values["tankdell"].ProjectedPoints != 85 {
		t.Fatalf("legacy canonical identity was not preserved: %#v %v", values, err)
	}
}
