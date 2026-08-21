package application

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/coopa11y/DraftMeld/backend/internal/domain/league"
)

func TestParseSleeperADPUsesTheRequestedLeagueProfile(t *testing.T) {
	items := make([]sleeperFeedItem, 0, 121)
	for index := 1; index <= 120; index++ {
		items = append(items, sleeperFeedItem{
			Stats:  map[string]float64{"adp_ppr": float64(index) + 0.4, "adp_std": float64(121 - index)},
			Season: "2026", LastModified: 1787212238024, PlayerID: fmt.Sprintf("player-%03d", index),
			Player: sleeperPlayer{FirstName: "Player", LastName: fmt.Sprint(index), Position: "WR", Team: "BUF"},
		})
	}
	items = append(items, sleeperFeedItem{
		Stats: map[string]float64{"adp_ppr": 999}, PlayerID: "unranked",
		Player: sleeperPlayer{FirstName: "Not", LastName: "Ranked", Position: "RB", Team: "ATL"},
	})
	records, published, err := parseSleeperADP(sleeperJSON(t, items), "sleeper-adp-redraft-ppr-1qb")
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 120 || records[0].ProviderID != "player-001" || records[0].ADP != 1.4 || records[0].Rank != 1 {
		t.Fatalf("unexpected Sleeper PPR rankings: %#v", records[:1])
	}
	if published == "" {
		t.Fatal("expected Sleeper update provenance")
	}
}

func TestMatchingSleeperADPProfileUsesFormatScoringAndQuarterbackDemand(t *testing.T) {
	rules := league.Rules{LeagueFormat: league.LeagueFormatDynasty, ScoringRules: map[string]float64{"reception": 0.5}}
	if got := matchingSleeperADPSourceID(rules, true); got != "sleeper-adp-dynasty-half-ppr-superflex" {
		t.Fatalf("unexpected Sleeper profile: %s", got)
	}
}

func sleeperJSON(t *testing.T, items []sleeperFeedItem) *bytes.Reader {
	t.Helper()
	contents, err := json.Marshal(items)
	if err != nil {
		t.Fatal(err)
	}
	return bytes.NewReader(contents)
}
