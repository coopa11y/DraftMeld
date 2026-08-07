package application

import (
	"github.com/coopa11y/DraftMeld/backend/internal/domain/league"
)

type RecommendationPolicy = league.RecommendationPolicy
type LeagueConfiguration = league.Configuration

func DefaultRecommendationPolicy() RecommendationPolicy {
	return RecommendationPolicy{
		BaseScore: 200, StartingNeedBonus: 24, ADPValueThreshold: 5,
		ScarcityBonus: 8, ScarcityDropOff: 5, RecommendationLimit: 5,
	}
}

func DemoLeagueConfiguration() LeagueConfiguration {
	return LeagueConfiguration{
		ID: "demo",
		Rules: league.Rules{
			Name: "Demo League", TeamCount: 12, DraftPosition: 1, DraftType: league.DraftTypeSnake,
			RosterSlots: []league.RosterSlot{
				{Name: "QB", Count: 1, Positions: []string{"QB"}, IsStarting: true},
				{Name: "RB", Count: 2, Positions: []string{"RB"}, IsStarting: true},
				{Name: "WR", Count: 2, Positions: []string{"WR"}, IsStarting: true},
				{Name: "TE", Count: 1, Positions: []string{"TE"}, IsStarting: true},
				{Name: "FLEX", Count: 1, Positions: []string{"RB", "WR", "TE"}, IsStarting: true},
				{Name: "K", Count: 1, Positions: []string{"K"}, IsStarting: true},
				{Name: "DST", Count: 1, Positions: []string{"DST"}, IsStarting: true},
				{Name: "Bench", Count: 6, Positions: []string{"QB", "RB", "WR", "TE", "K", "DST"}, IsStarting: false},
			},
			ScoringRules:      map[string]float64{"reception": 1},
			SourcePreferences: DefaultRankingSourcePreferences(),
			ConsensusMethod:   "weighted-median",
			PlayerPreferences: map[string]string{},
			AuctionBudget:     200,
			AuctionMinimumBid: 1,
		},
		Recommendation: DefaultRecommendationPolicy(),
	}
}
