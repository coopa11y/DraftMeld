package application

import (
	"strings"
	"testing"
)

func TestParseECRFiltersSupportedPlayers(t *testing.T) {
	input := "page_type,player,pos,team,ecr,scrape_date\nredraft-overall,Alpha Runner,RB,AAA,2.4,2026-07-31\nredraft-overall,Beta Passer,QB,BBB,1.2,2026-07-31\nredraft-dst,Defense,DST,CCC,1,2026-07-31\n"
	records, published, err := parseRankingSource("redraft-ecr", strings.NewReader(input))
	if err != nil {
		t.Fatalf("parse ECR: %v", err)
	}
	if len(records) != 2 || records[0].Name != "Beta Passer" || records[0].Rank != 1 {
		t.Fatalf("unexpected records: %#v", records)
	}
	if published != "2026-07-31" {
		t.Fatalf("unexpected published date: %s", published)
	}
}

func TestParseDynastyRanksHighestValueFirst(t *testing.T) {
	input := "player,pos,team,value_1qb,value_2qb,scrape_date\nAlpha Runner,RB,AAA,8000,7000,2026-07-31\nBeta Passer,QB,BBB,4000,9000,2026-07-31\n"
	records, _, err := parseRankingSource("dynasty-superflex", strings.NewReader(input))
	if err != nil {
		t.Fatalf("parse dynasty: %v", err)
	}
	if records[0].Name != "Beta Passer" {
		t.Fatalf("expected superflex QB first, got %#v", records)
	}
}

func TestParseOpportunityAggregatesWeeklyExpectedPoints(t *testing.T) {
	input := "season,posteam,full_name,position,total_fantasy_points_exp\n2025,AAA,Alpha Runner,RB,10.5\n2025,AAA,Alpha Runner,RB,12.5\n2025,BBB,Beta Passer,QB,20\n"
	records, _, err := parseRankingSource("expected-opportunity", strings.NewReader(input))
	if err != nil {
		t.Fatalf("parse opportunity: %v", err)
	}
	if records[0].Name != "Alpha Runner" || records[0].Rank != 1 {
		t.Fatalf("expected aggregated leader, got %#v", records)
	}
}

func TestParseCBSUsesOnlyTheConsensusRankingGroup(t *testing.T) {
	row := func(rank, slug, position string) string {
		return `<div class="player-row"><div class="rank">` + rank + `</div><div><a href="/nfl/players/1/` + slug + `/fantasy/"><span>Player</span></a><span class="team position">` + position + ` $20</span></div></div>`
	}
	input := `Updated 2h ago` + row("1", "alpha-runner", "RB") + row("2", "beta-passer", "QB") + row("1", "expert-favorite", "WR")
	records, published, err := parseRankingSource("cbs-ppr", strings.NewReader(input))
	if err != nil {
		t.Fatalf("parse CBS: %v", err)
	}
	if len(records) != 2 || records[0].PlayerKey != "alpharunner" || records[1].Position != "QB" {
		t.Fatalf("unexpected CBS consensus: %#v", records)
	}
	if published != "Updated 2h ago" {
		t.Fatalf("unexpected CBS update label: %s", published)
	}
}
