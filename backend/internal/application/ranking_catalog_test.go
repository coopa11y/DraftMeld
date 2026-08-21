package application

import (
	"testing"

	"github.com/coopa11y/DraftMeld/backend/internal/domain/league"
)

func TestRecommendedRankingSourcesMatchRedraftPPRLeague(t *testing.T) {
	rules := DemoLeagueConfiguration().Rules
	rules.TeamCount = 10
	preferences := RecommendedRankingSourcePreferences(rules)

	assertSourcePreference(t, preferences, "redraft-ecr", true, 1)
	assertSourcePreference(t, preferences, "cbs-ppr", true, 0.9)
	assertSourcePreference(t, preferences, "espn-ppr-online", true, 0.9)
	assertSourcePreference(t, preferences, "espn-ppr-pdf", false, 0.9)
	assertSourcePreference(t, preferences, "expected-opportunity", true, 0.35)
	assertSourcePreference(t, preferences, "dynasty-1qb", false, 0.7)
	assertSourcePreference(t, preferences, "dynasty-superflex", false, 0.5)
	assertSourcePreference(t, preferences, "espn-dynasty-pdf", false, 0.6)
	assertSourcePreference(t, preferences, "yahoo-standard", false, 0.3)
	assertSourcePreference(t, preferences, "draft-sharks-ppr-1qb", true, 0.9)
	assertSourcePreference(t, preferences, "draft-sharks-half-ppr-1qb", false, 0.9)
	assertSourcePreference(t, preferences, "sleeper-adp-redraft-ppr-1qb", true, 0.5)
	assertSourcePreference(t, preferences, "sleeper-adp-redraft-standard-1qb", false, 0.5)
}

func TestRecommendedRankingSourcesDistinguishDynastyQuarterbackDemand(t *testing.T) {
	rules := DemoLeagueConfiguration().Rules
	rules.LeagueFormat = league.LeagueFormatDynasty
	rules.FuturePickSeasons = 3
	oneQB := RecommendedRankingSourcePreferences(rules)
	assertSourcePreference(t, oneQB, "dynasty-1qb", true, 1)
	assertSourcePreference(t, oneQB, "dynasty-superflex", false, 0.5)

	rules.RosterSlots = append(rules.RosterSlots, league.RosterSlot{
		Name: "SUPERFLEX", Count: 1, Positions: []string{"QB", "RB", "WR", "TE"}, IsStarting: true,
	})
	superflex := RecommendedRankingSourcePreferences(rules)
	assertSourcePreference(t, superflex, "dynasty-1qb", false, 0.7)
	assertSourcePreference(t, superflex, "dynasty-superflex", true, 1)
	assertSourcePreference(t, superflex, "draft-sharks-ppr-superflex", true, 0.25)
	assertSourcePreference(t, superflex, "sleeper-adp-dynasty-ppr-superflex", true, 0.65)
}

func TestRecommendedRankingSourcesDoNotTreatPPRListsAsStandardRankings(t *testing.T) {
	rules := DemoLeagueConfiguration().Rules
	rules.ScoringRules["reception"] = 0
	preferences := RecommendedRankingSourcePreferences(rules)
	assertSourcePreference(t, preferences, "redraft-ecr", true, 1)
	assertSourcePreference(t, preferences, "cbs-ppr", false, 0.9)
	assertSourcePreference(t, preferences, "espn-ppr-pdf", false, 0.9)
	assertSourcePreference(t, preferences, "espn-ppr-online", false, 0.9)
	assertSourcePreference(t, preferences, "yahoo-standard", true, 0.3)
	assertSourcePreference(t, preferences, "draft-sharks-standard-1qb", true, 0.9)
	assertSourcePreference(t, preferences, "sleeper-adp-redraft-standard-1qb", true, 0.5)
}

func TestRecommendedRankingSourcesChooseHalfPPRAndTEPremiumPresets(t *testing.T) {
	rules := DemoLeagueConfiguration().Rules
	rules.ScoringRules["reception"] = 0.5
	halfPPR := RecommendedRankingSourcePreferences(rules)
	assertSourcePreference(t, halfPPR, "draft-sharks-half-ppr-1qb", true, 0.9)
	assertSourcePreference(t, halfPPR, "draft-sharks-ppr-1qb", false, 0.9)

	rules.ScoringRules["tightEndReceptionBonus"] = 0.5
	tePremium := RecommendedRankingSourcePreferences(rules)
	assertSourcePreference(t, tePremium, "draft-sharks-tep-1qb", true, 0.9)
}

func assertSourcePreference(t *testing.T, preferences map[string]league.RankingSourcePreference, id string, enabled bool, weight float64) {
	t.Helper()
	preference, exists := preferences[id]
	if !exists || preference.Enabled != enabled || preference.Weight != weight {
		t.Fatalf("unexpected %s preference: %#v (exists=%v)", id, preference, exists)
	}
}
