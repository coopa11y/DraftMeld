package application

import "github.com/coopa11y/DraftMeld/backend/internal/domain/ranking"

const (
	dynastyDataURL  = "https://raw.githubusercontent.com/dynastyprocess/data/master/files/values-players.csv"
	ecrDataURL      = "https://raw.githubusercontent.com/dynastyprocess/data/master/files/db_fpecr_latest.csv"
	opportunityURL  = "https://github.com/ffverse/ffopportunity/releases/download/latest-data/ep_weekly_2025.csv"
	cbsRankingsURL  = "https://www.cbssports.com/fantasy/football/rankings/"
	espnDraftKitURL = "https://www.espn.com/fantasy/football/"
)

func BuiltInRankingSources() []ranking.SourceDefinition {
	return []ranking.SourceDefinition{
		{ID: "redraft-ecr", Name: "Redraft expert consensus", Description: "Current overall redraft consensus across every supported fantasy position.", Methodology: "Average rank from participating fantasy experts", License: "GPL-3.0 open-data repository; upstream FantasyPros attribution", ProjectURL: "https://github.com/dynastyprocess/data", DataURL: ecrDataURL, DefaultWeight: 1, ImportMode: "download"},
		{ID: "dynasty-1qb", Name: "Dynasty market - 1 QB", Description: "Long-term player market values for traditional one-quarterback leagues.", Methodology: "DynastyProcess normalized 1-QB player value", License: "GPL-3.0", ProjectURL: "https://github.com/dynastyprocess/data", DataURL: dynastyDataURL, DefaultWeight: 0.7, ImportMode: "download"},
		{ID: "dynasty-superflex", Name: "Dynasty market - Superflex", Description: "Long-term values that account for elevated quarterback demand.", Methodology: "DynastyProcess normalized 2-QB/Superflex player value", License: "GPL-3.0", ProjectURL: "https://github.com/dynastyprocess/data", DataURL: dynastyDataURL, DefaultWeight: 0.5, ImportMode: "download"},
		{ID: "expected-opportunity", Name: "Expected opportunity", Description: "Prior-season usage quality measured independently of box-score luck.", Methodology: "2025 ffopportunity expected fantasy points, summed by player", License: "CC-BY-SA-4.0", ProjectURL: "https://github.com/ffverse/ffopportunity", DataURL: opportunityURL, DefaultWeight: 0.6, ImportMode: "download"},
		{ID: "cbs-ppr", Name: "CBS Sports PPR Top 200", Description: "Current CBS Sports expert-consensus rankings for PPR redraft leagues.", Methodology: "CBS Fantasy Experts consensus order", License: "Proprietary; retrieved on demand and not redistributed", ProjectURL: cbsRankingsURL, DataURL: cbsRankingsURL, DefaultWeight: 0.9, ImportMode: "download"},
		{ID: "espn-ppr-pdf", Name: "ESPN PPR Top 300 PDF", Description: "Overall PPR rankings imported from a user-supplied ESPN draft-kit PDF.", Methodology: "ESPN overall ordinal rank across QB, RB, WR, TE, K, and DST", License: "Proprietary; user-supplied and never redistributed", ProjectURL: espnDraftKitURL, DataURL: espnDraftKitURL, DefaultWeight: 0.9, ImportMode: "pdf-upload"},
		{ID: "espn-dynasty-pdf", Name: "ESPN Dynasty PDF", Description: "Long-term overall rankings imported from a user-supplied ESPN dynasty cheat sheet.", Methodology: "ESPN dynasty overall ordinal rank for supported positions present in the sheet", License: "Proprietary; user-supplied and never redistributed", ProjectURL: espnDraftKitURL, DataURL: espnDraftKitURL, DefaultWeight: 0.6, ImportMode: "pdf-upload"},
	}
}

func DefaultRankingSourceWeights() map[string]float64 {
	weights := make(map[string]float64)
	for _, source := range BuiltInRankingSources() {
		weights[source.ID] = source.DefaultWeight
	}
	return weights
}
