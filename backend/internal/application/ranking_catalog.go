package application

import "github.com/coopa11y/DraftMeld/backend/internal/domain/ranking"

const (
	dynastyDataURL = "https://raw.githubusercontent.com/dynastyprocess/data/master/files/values-players.csv"
	ecrDataURL     = "https://raw.githubusercontent.com/dynastyprocess/data/master/files/db_fpecr_latest.csv"
	opportunityURL = "https://github.com/ffverse/ffopportunity/releases/download/latest-data/ep_weekly_2025.csv"
)

func BuiltInRankingSources() []ranking.SourceDefinition {
	return []ranking.SourceDefinition{
		{ID: "redraft-ecr", Name: "Redraft expert consensus", Description: "Current overall redraft consensus for QB, RB, WR, and TE.", Methodology: "Average rank from participating fantasy experts", License: "GPL-3.0 open-data repository; upstream FantasyPros attribution", ProjectURL: "https://github.com/dynastyprocess/data", DataURL: ecrDataURL, DefaultWeight: 1},
		{ID: "dynasty-1qb", Name: "Dynasty market — 1 QB", Description: "Long-term player market values for traditional one-quarterback leagues.", Methodology: "DynastyProcess normalized 1-QB player value", License: "GPL-3.0", ProjectURL: "https://github.com/dynastyprocess/data", DataURL: dynastyDataURL, DefaultWeight: 0.7},
		{ID: "dynasty-superflex", Name: "Dynasty market — Superflex", Description: "Long-term values that account for elevated quarterback demand.", Methodology: "DynastyProcess normalized 2-QB/Superflex player value", License: "GPL-3.0", ProjectURL: "https://github.com/dynastyprocess/data", DataURL: dynastyDataURL, DefaultWeight: 0.5},
		{ID: "expected-opportunity", Name: "Expected opportunity", Description: "Prior-season usage quality measured independently of box-score luck.", Methodology: "2025 ffopportunity expected fantasy points, summed by player", License: "CC-BY-SA-4.0", ProjectURL: "https://github.com/ffverse/ffopportunity", DataURL: opportunityURL, DefaultWeight: 0.6},
	}
}
