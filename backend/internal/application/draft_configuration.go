package application

import (
	"errors"
	"fmt"

	"github.com/coopa11y/DraftMeld/backend/internal/domain/league"
)

type RecommendationPolicy struct {
	BaseScore           float64
	StartingNeedBonus   float64
	ADPValueThreshold   float64
	ScarcityBonus       float64
	ScarcityDropOff     int
	RecommendationLimit int
}

func DefaultRecommendationPolicy() RecommendationPolicy {
	return RecommendationPolicy{
		BaseScore: 200, StartingNeedBonus: 24, ADPValueThreshold: 5,
		ScarcityBonus: 8, ScarcityDropOff: 5, RecommendationLimit: 5,
	}
}

type LeagueConfiguration struct {
	ID             string
	Rules          league.Rules
	Recommendation RecommendationPolicy
}

func (configuration LeagueConfiguration) Validate() error {
	if configuration.ID == "" {
		return errors.New("league configuration requires an ID")
	}
	if err := configuration.Rules.Validate(); err != nil {
		return fmt.Errorf("validate league %s: %w", configuration.ID, err)
	}
	policy := configuration.Recommendation
	if policy.BaseScore <= 0 || policy.StartingNeedBonus < 0 || policy.ADPValueThreshold < 0 ||
		policy.ScarcityBonus < 0 || policy.ScarcityDropOff < 1 || policy.RecommendationLimit < 1 {
		return fmt.Errorf("league %s has an invalid recommendation policy", configuration.ID)
	}
	return nil
}

func DemoLeagueConfiguration() LeagueConfiguration {
	return LeagueConfiguration{
		ID: "demo",
		Rules: league.Rules{
			Name: "Demo League", TeamCount: 12, DraftType: league.DraftTypeSnake,
			RosterSlots: []league.RosterSlot{
				{Name: "QB", Count: 1, Positions: []string{"QB"}, IsStarting: true},
				{Name: "RB", Count: 2, Positions: []string{"RB"}, IsStarting: true},
				{Name: "WR", Count: 2, Positions: []string{"WR"}, IsStarting: true},
				{Name: "TE", Count: 1, Positions: []string{"TE"}, IsStarting: true},
				{Name: "FLEX", Count: 1, Positions: []string{"RB", "WR", "TE"}, IsStarting: true},
				{Name: "Bench", Count: 6, Positions: []string{"QB", "RB", "WR", "TE"}, IsStarting: false},
			},
			ScoringRules: map[string]float64{"reception": 1},
		},
		Recommendation: DefaultRecommendationPolicy(),
	}
}
