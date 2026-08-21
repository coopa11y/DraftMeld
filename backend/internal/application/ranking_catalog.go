package application

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/coopa11y/DraftMeld/backend/internal/domain/league"
	"github.com/coopa11y/DraftMeld/backend/internal/domain/ranking"
)

const (
	dynastyDataURL   = "https://raw.githubusercontent.com/dynastyprocess/data/master/files/values-players.csv"
	ecrDataURL       = "https://raw.githubusercontent.com/dynastyprocess/data/master/files/db_fpecr_latest.csv"
	opportunityURL   = "https://github.com/ffverse/ffopportunity/releases/download/latest-data/ep_weekly_2025.csv"
	cbsRankingsURL   = "https://www.cbssports.com/fantasy/football/rankings/"
	yahooRankingsURL = "https://football.fantasysports.yahoo.com/f1/public_prerank"
	espnDraftKitURL  = "https://www.espn.com/fantasy/football/"
	espnPPRDataURL   = "https://lm-api-reads.fantasy.espn.com/apis/v3/games/ffl/seasons/2026/segments/0/leaguedefaults/1?view=kona_player_info"
)

func BuiltInRankingSources() []ranking.SourceDefinition {
	sources := []ranking.SourceDefinition{
		{ID: "redraft-ecr", Name: "Redraft PPR expert consensus", Description: "Current PPR redraft consensus across every supported fantasy position.", Methodology: "Average PPR rank from participating fantasy experts", License: "GPL-3.0 open-data repository; upstream FantasyPros attribution", ProjectURL: "https://github.com/dynastyprocess/data", DataURL: ecrDataURL, DefaultWeight: 1, DefaultEnabled: true, ImportMode: "download", Role: "ranking"},
		{ID: "dynasty-1qb", Name: "Dynasty market - 1 QB", Description: "Long-term player market values for traditional one-quarterback leagues.", Methodology: "DynastyProcess normalized 1-QB player value", License: "GPL-3.0", ProjectURL: "https://github.com/dynastyprocess/data", DataURL: dynastyDataURL, DefaultWeight: 0.7, ImportMode: "download", Role: "market"},
		{ID: "dynasty-superflex", Name: "Dynasty market - Superflex", Description: "Long-term values that account for elevated quarterback demand.", Methodology: "DynastyProcess normalized 2-QB/Superflex player value", License: "GPL-3.0", ProjectURL: "https://github.com/dynastyprocess/data", DataURL: dynastyDataURL, DefaultWeight: 0.5, ImportMode: "download", Role: "market"},
		{ID: "expected-opportunity", Name: "Expected opportunity", Description: "Prior-season usage quality measured independently of box-score luck.", Methodology: "2025 ffopportunity expected fantasy points, summed by player", License: "CC-BY-SA-4.0", ProjectURL: "https://github.com/ffverse/ffopportunity", DataURL: opportunityURL, DefaultWeight: 0.6, DefaultEnabled: true, ImportMode: "download", Role: "usage"},
		{ID: "cbs-ppr", Name: "CBS Sports PPR Top 200", Description: "Current CBS Sports expert-consensus rankings for PPR redraft leagues.", Methodology: "CBS Fantasy Experts consensus order", License: "Proprietary; retrieved on demand and not redistributed", ProjectURL: cbsRankingsURL, DataURL: cbsRankingsURL, DefaultWeight: 0.9, DefaultEnabled: true, ImportMode: "download", Role: "ranking"},
		{ID: "espn-ppr-online", Name: "ESPN PPR draft rankings", Description: "Current ESPN PPR default draft order with ESPN player IDs and average draft position.", Methodology: "ESPN PPR draft rank from its public fantasy player service", License: "Proprietary; retrieved on demand and not redistributed", ProjectURL: espnDraftKitURL, DataURL: espnPPRDataURL, DefaultWeight: 0.9, DefaultEnabled: true, ImportMode: "download", Role: "ranking"},
		{ID: "espn-ppr-pdf", Name: "ESPN PPR Top 300 PDF", Description: "Overall PPR rankings imported from a user-supplied ESPN draft-kit PDF.", Methodology: "ESPN overall ordinal rank across QB, RB, WR, TE, K, and DST", License: "Proprietary; user-supplied and never redistributed", ProjectURL: espnDraftKitURL, DataURL: espnDraftKitURL, DefaultWeight: 0.9, ImportMode: "pdf-upload", Role: "ranking"},
		{ID: "espn-dynasty-pdf", Name: "ESPN Dynasty PDF", Description: "Long-term overall rankings imported from a user-supplied ESPN dynasty cheat sheet.", Methodology: "ESPN dynasty overall ordinal rank for supported positions present in the sheet", License: "Proprietary; user-supplied and never redistributed", ProjectURL: espnDraftKitURL, DataURL: espnDraftKitURL, DefaultWeight: 0.6, ImportMode: "pdf-upload", Role: "market"},
		{ID: "yahoo-standard", Name: "Yahoo default Standard Top 200", Description: "Yahoo's public default pre-draft order for Standard leagues.", Methodology: "Yahoo default platform draft order; league-specific Xrank requires Yahoo authorization", License: "Proprietary; retrieved on demand and not redistributed", ProjectURL: yahooRankingsURL, DataURL: yahooRankingsURL, DefaultWeight: 0.3, ImportMode: "download", Role: "market"},
	}
	sources = append(sources, draftSharksSources()...)
	return append(sources, sleeperADPSources()...)
}

func DefaultRankingSourcePreferences() map[string]league.RankingSourcePreference {
	return RecommendedRankingSourcePreferences(league.Rules{
		LeagueFormat: league.LeagueFormatRedraft,
		RosterSlots:  []league.RosterSlot{{Name: "QB", Count: 1, Positions: []string{"QB"}, IsStarting: true}},
		ScoringRules: map[string]float64{"reception": 1},
	})
}

type RankingSourceRecommendation struct {
	SourceID   string                         `json:"sourceId"`
	Fit        string                         `json:"fit"`
	Reason     string                         `json:"reason"`
	Preference league.RankingSourcePreference `json:"preference"`
}

type RankingSourceRecommendations struct {
	Profile string                        `json:"profile"`
	Sources []RankingSourceRecommendation `json:"sources"`
}

func (service *RankingService) Recommendations(ctx context.Context, rules league.Rules) (RankingSourceRecommendations, error) {
	definitions, err := service.definitions(ctx)
	if err != nil {
		return RankingSourceRecommendations{}, err
	}
	return rankingSourceRecommendations(rules, definitions), nil
}

func RecommendedRankingSourcePreferences(rules league.Rules) map[string]league.RankingSourcePreference {
	recommendations := rankingSourceRecommendations(rules, BuiltInRankingSources())
	preferences := make(map[string]league.RankingSourcePreference)
	for _, recommendation := range recommendations.Sources {
		preferences[recommendation.SourceID] = recommendation.Preference
	}
	return preferences
}

func rankingSourceRecommendations(rules league.Rules, definitions []ranking.SourceDefinition) RankingSourceRecommendations {
	dynasty := rules.LeagueFormat == league.LeagueFormatDynasty
	superflex := hasSuperflex(rules.RosterSlots)
	ppr := rules.ScoringRules["reception"] > 0
	draftSharksSourceID := matchingDraftSharksSourceID(rules.ScoringRules, superflex)
	sleeperSourceID := matchingSleeperADPSourceID(rules, superflex)
	profile := []string{map[bool]string{true: "dynasty", false: "redraft"}[dynasty]}
	if superflex {
		profile = append(profile, "Superflex/2-QB")
	} else {
		profile = append(profile, "1-QB")
	}
	if ppr {
		profile = append(profile, fmt.Sprintf("%g PPR", rules.ScoringRules["reception"]))
	} else {
		profile = append(profile, "standard scoring")
	}

	result := RankingSourceRecommendations{Profile: strings.Join(profile, ", "), Sources: make([]RankingSourceRecommendation, 0, len(definitions))}
	for _, definition := range definitions {
		recommendation := recommendationForSource(definition, dynasty, superflex, ppr, draftSharksSourceID, sleeperSourceID)
		result.Sources = append(result.Sources, recommendation)
	}
	return result
}

func recommendationForSource(source ranking.SourceDefinition, dynasty, superflex, ppr bool, draftSharksSourceID, sleeperSourceID string) RankingSourceRecommendation {
	preference := league.RankingSourcePreference{Weight: source.DefaultWeight, Enabled: false}
	fit, reason := "not recommended", "This source does not match the league format."
	if source.VariantGroup == "draft-sharks" {
		if source.ID != draftSharksSourceID {
			reason = "DraftMeld selected the Draft Sharks preset that more closely matches this league."
			return RankingSourceRecommendation{SourceID: source.ID, Fit: fit, Reason: reason, Preference: preference}
		}
		preference.Enabled = true
		if dynasty {
			preference.Weight, fit = 0.25, "context"
			reason = "This preset matches the scoring and quarterback rules, but its seasonal projections receive reduced influence in dynasty."
		} else {
			fit, reason = "recommended", "Draft Sharks tiers, projections, and 3D Value match this league's closest public scoring and quarterback preset."
		}
		return RankingSourceRecommendation{SourceID: source.ID, Fit: fit, Reason: reason, Preference: preference}
	}
	if source.VariantGroup == "sleeper-adp" {
		if source.ID != sleeperSourceID {
			reason = "DraftMeld selected the Sleeper ADP profile that matches this league."
			return RankingSourceRecommendation{SourceID: source.ID, Fit: fit, Reason: reason, Preference: preference}
		}
		preference.Enabled = true
		preference.Weight = 0.5
		fit = "context"
		reason = "Sleeper ADP shows where its community is selecting players in this league format; it complements expert rankings without overpowering them."
		if dynasty {
			preference.Weight, fit = 0.65, "recommended"
			reason = "Sleeper dynasty ADP is current market evidence matched to this league's scoring and quarterback demand."
		}
		return RankingSourceRecommendation{SourceID: source.ID, Fit: fit, Reason: reason, Preference: preference}
	}
	switch source.ID {
	case "redraft-ecr":
		preference.Enabled = true
		if dynasty {
			preference.Weight, fit, reason = 0.35, "context", "Keeps the current player pool grounded in seasonal value while dynasty rankings drive long-term value."
		} else if !ppr {
			fit, reason = "context", "This PPR list is the closest built-in current-season baseline, but a standard-scoring ranking CSV would be a better primary source."
		} else {
			fit, reason = "recommended", "Primary current-season PPR expert consensus for this redraft league."
		}
	case "cbs-ppr", "espn-ppr-online":
		preference.Enabled = ppr
		if ppr && dynasty {
			preference.Weight, fit, reason = 0.25, "context", "Useful current-season PPR context, with reduced influence in a dynasty league."
		} else if ppr {
			fit, reason = "recommended", "Matches this league's reception scoring and redraft format."
		} else {
			reason = "This is a PPR list, but the league awards no points per reception."
		}
	case "espn-ppr-pdf":
		preference.Enabled = false
		if ppr {
			fit, reason = "fallback", "Use this only when the online ESPN PPR source is unavailable; enabling both would count ESPN twice."
		} else {
			reason = "This is a PPR list, but the league awards no points per reception."
		}
	case "dynasty-1qb":
		preference.Enabled = dynasty && !superflex
		if preference.Enabled {
			preference.Weight, fit, reason = 1, "recommended", "Matches this dynasty league's one-quarterback starting structure."
		} else if !dynasty {
			reason = "Long-term dynasty values should not influence a redraft league."
		} else {
			reason = "This league gives quarterbacks Superflex/2-QB demand."
		}
	case "dynasty-superflex":
		preference.Enabled = dynasty && superflex
		if preference.Enabled {
			preference.Weight, fit, reason = 1, "recommended", "Matches this dynasty league's Superflex/2-QB starting structure."
		} else if !dynasty {
			reason = "Long-term dynasty values should not influence a redraft league."
		} else {
			reason = "This is a one-quarterback league, so Superflex values would overstate quarterback demand."
		}
	case "espn-dynasty-pdf":
		preference.Enabled = dynasty
		if dynasty {
			preference.Weight, fit, reason = 0.9, "recommended", "Long-term ESPN values match this dynasty league."
		} else {
			reason = "Long-term dynasty values should not influence a redraft league."
		}
	case "expected-opportunity":
		preference.Enabled, preference.Weight = true, 0.35
		fit, reason = "context", "Prior-season opportunity can break close calls, but receives less influence than current rankings."
	case "yahoo-standard":
		preference.Enabled = !ppr
		if preference.Enabled {
			preference.Weight, fit, reason = 0.3, "context", "Yahoo's Standard platform order is useful draft-room context, but should not outweigh expert rankings."
		} else {
			reason = "Yahoo labels this public default order Standard, while this league awards points per reception."
		}
	default:
		preference.Enabled = true
		fit, reason = "user source", "User-supplied rankings are included until the user chooses otherwise."
	}
	return RankingSourceRecommendation{SourceID: source.ID, Fit: fit, Reason: reason, Preference: preference}
}

func matchingDraftSharksSourceID(scoring map[string]float64, superflex bool) string {
	format := "standard"
	switch receptions := scoring["reception"]; {
	case scoring["tightEndReceptionBonus"] > 0:
		format = "tep"
	case receptions >= 0.75:
		format = "ppr"
	case receptions > 0:
		format = "half-ppr"
	}
	quarterbacks := "1qb"
	if superflex {
		quarterbacks = "superflex"
	}
	return "draft-sharks-" + format + "-" + quarterbacks
}

func hasSuperflex(slots []league.RosterSlot) bool {
	quarterbackStarters := 0
	for _, slot := range slots {
		if !slot.IsStarting || !slices.Contains(slot.Positions, "QB") {
			continue
		}
		if len(slot.Positions) > 1 {
			return true
		}
		quarterbackStarters += slot.Count
	}
	return quarterbackStarters > 1
}
