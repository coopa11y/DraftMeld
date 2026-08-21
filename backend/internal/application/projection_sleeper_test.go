package application

import (
	"errors"
	"fmt"
	"testing"
)

func TestRefreshSleeperProjectionRejectsUnknownSource(t *testing.T) {
	service := NewProjectionService(&projectionRepositoryStub{})
	if _, err := service.RefreshSource(t.Context(), "private-projections"); !errors.Is(err, ErrProjectionSourceNotFound) {
		t.Fatalf("expected projection source validation, got %v", err)
	}
}

func TestParseSleeperProjectionsMapsRawOffensiveStatistics(t *testing.T) {
	items := make([]sleeperFeedItem, 0, 302)
	for index := 1; index <= 301; index++ {
		items = append(items, sleeperFeedItem{
			Stats: map[string]float64{
				"pts_ppr": 100, "rec": 70, "rec_yd": 900, "rec_td": 7, "fum_lost": 1,
			},
			Season: "2026", PlayerID: fmt.Sprintf("receiver-%03d", index),
			Player: sleeperPlayer{FirstName: "Receiver", LastName: fmt.Sprint(index), Position: "WR", Team: "LAR"},
		})
	}
	items = append(items, sleeperFeedItem{
		Stats: map[string]float64{"pts_ppr": 80, "xpm": 30}, PlayerID: "kicker",
		Player: sleeperPlayer{FirstName: "Test", LastName: "Kicker", Position: "K", Team: "DAL"},
	})
	records, _, err := parseSleeperProjections(sleeperJSON(t, items))
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 301 {
		t.Fatalf("expected only complete offensive projections, got %d", len(records))
	}
	first := records[0]
	if first.SourceID != sleeperProjectionID || first.ProviderID != "receiver-001" || first.Stats["reception"] != 70 || first.Stats["receivingYard"] != 900 || first.Stats["receivingTouchdown"] != 7 || first.Stats["fumbleLost"] != 1 {
		t.Fatalf("unexpected mapped Sleeper projection: %#v", first)
	}
}

func TestSleeperParsersRejectIncompleteResponses(t *testing.T) {
	items := []sleeperFeedItem{{Stats: map[string]float64{"adp_ppr": 1}, PlayerID: "one", Player: sleeperPlayer{FirstName: "One", LastName: "Player", Position: "RB", Team: "ATL"}}}
	if _, _, err := parseSleeperADP(sleeperJSON(t, items), "sleeper-adp-redraft-ppr-1qb"); err == nil {
		t.Fatal("expected incomplete ADP feed to fail")
	}
	if _, _, err := parseSleeperProjections(sleeperJSON(t, items)); err == nil {
		t.Fatal("expected incomplete projection feed to fail")
	}
}
