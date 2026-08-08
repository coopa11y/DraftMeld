package league

import (
	"errors"
	"fmt"
	"math"
	"strings"
)

type DraftType string

const (
	DraftTypeSnake   DraftType = "snake"
	DraftTypeLinear  DraftType = "linear"
	DraftTypeAuction DraftType = "auction"
)

type RosterSlot struct {
	Name       string   `json:"name"`
	Count      int      `json:"count"`
	Positions  []string `json:"positions"`
	IsStarting bool     `json:"isStarting"`
}

type RankingSourcePreference struct {
	Weight  float64 `json:"weight"`
	Enabled bool    `json:"enabled"`
}

type Rules struct {
	Name               string                             `json:"name"`
	TeamCount          int                                `json:"teamCount"`
	DraftPosition      int                                `json:"draftPosition"`
	TeamNames          []string                           `json:"teamNames"`
	DraftType          DraftType                          `json:"draftType"`
	RosterSlots        []RosterSlot                       `json:"rosterSlots"`
	ScoringRules       map[string]float64                 `json:"scoringRules"`
	SourcePreferences  map[string]RankingSourcePreference `json:"sourcePreferences"`
	ConsensusMethod    string                             `json:"consensusMethod"`
	PlayerPreferences  map[string]string                  `json:"playerPreferences"`
	AuctionBudget      float64                            `json:"auctionBudget"`
	AuctionMinimumBid  float64                            `json:"auctionMinimumBid"`
	KeeperBudgetSpent  float64                            `json:"keeperBudgetSpent"`
	MyKeeperSpend      float64                            `json:"myKeeperSpend"`
	KeeperValueRemoved float64                            `json:"keeperValueRemoved"`
}

type RecommendationPolicy struct {
	BaseScore           float64 `json:"baseScore"`
	StartingNeedBonus   float64 `json:"startingNeedBonus"`
	ADPValueThreshold   float64 `json:"adpValueThreshold"`
	ScarcityBonus       float64 `json:"scarcityBonus"`
	ScarcityDropOff     int     `json:"scarcityDropOff"`
	RecommendationLimit int     `json:"recommendationLimit"`
}

type Configuration struct {
	ID             string               `json:"id"`
	Rules          Rules                `json:"rules"`
	Recommendation RecommendationPolicy `json:"recommendation"`
}

func (configuration Configuration) Validate() error {
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

func (rules Rules) Validate() error {
	if rules.Name == "" {
		return errors.New("league name is required")
	}
	if rules.TeamCount < 2 || rules.TeamCount > 32 {
		return fmt.Errorf("team count must be between 2 and 32: %d", rules.TeamCount)
	}
	if rules.DraftPosition < 1 || rules.DraftPosition > rules.TeamCount {
		return fmt.Errorf("draft position must be between 1 and %d: %d", rules.TeamCount, rules.DraftPosition)
	}
	if len(rules.TeamNames) != 0 && len(rules.TeamNames) != rules.TeamCount {
		return fmt.Errorf("team names must contain exactly %d entries", rules.TeamCount)
	}
	seenTeamNames := make(map[string]bool, len(rules.TeamNames))
	for _, name := range rules.TeamNames {
		normalized := strings.ToLower(strings.TrimSpace(name))
		if normalized == "" || len(name) > 80 {
			return errors.New("team names must be between 1 and 80 characters")
		}
		if seenTeamNames[normalized] {
			return fmt.Errorf("team names must be unique: %q", name)
		}
		seenTeamNames[normalized] = true
	}
	switch rules.DraftType {
	case DraftTypeSnake, DraftTypeLinear, DraftTypeAuction:
	default:
		return fmt.Errorf("unsupported draft type: %q", rules.DraftType)
	}
	if rules.DraftType == DraftTypeAuction && rules.AuctionBudget <= 0 {
		return errors.New("auction leagues require a positive team budget")
	}
	if rules.DraftType == DraftTypeAuction && (rules.AuctionMinimumBid <= 0 || rules.AuctionMinimumBid > rules.AuctionBudget) {
		return errors.New("auction minimum bid must be positive and no greater than the team budget")
	}
	if rules.KeeperBudgetSpent < 0 || rules.KeeperBudgetSpent > rules.AuctionBudget*float64(rules.TeamCount) {
		return errors.New("league-wide keeper spend exceeds the available auction budget")
	}
	if rules.MyKeeperSpend < 0 || rules.MyKeeperSpend > rules.AuctionBudget || rules.MyKeeperSpend > rules.KeeperBudgetSpent {
		return errors.New("your keeper spend must fit within both the team budget and league-wide keeper spend")
	}
	if rules.KeeperValueRemoved < 0 {
		return errors.New("keeper value removed cannot be negative")
	}
	if len(rules.RosterSlots) == 0 {
		return errors.New("at least one roster slot is required")
	}
	switch rules.ConsensusMethod {
	case "", "weighted-average", "weighted-median", "trimmed-mean":
	default:
		return fmt.Errorf("unsupported consensus method: %q", rules.ConsensusMethod)
	}
	for playerID, preference := range rules.PlayerPreferences {
		if playerID == "" || (preference != "target" && preference != "avoid") {
			return fmt.Errorf("invalid player preference for %q", playerID)
		}
	}
	for _, slot := range rules.RosterSlots {
		if slot.Name == "" || slot.Count < 1 || len(slot.Positions) == 0 {
			return fmt.Errorf("invalid roster slot: %#v", slot)
		}
	}
	for name, value := range rules.ScoringRules {
		if name == "" || math.IsNaN(value) || math.IsInf(value, 0) {
			return fmt.Errorf("invalid scoring rule: %q", name)
		}
	}
	enabledSources := 0
	for sourceID, preference := range rules.SourcePreferences {
		if sourceID == "" || math.IsNaN(preference.Weight) || math.IsInf(preference.Weight, 0) || preference.Weight <= 0 || preference.Weight > 10 {
			return fmt.Errorf("ranking source weight must be greater than 0 and no more than 10: %q", sourceID)
		}
		if preference.Enabled {
			enabledSources++
		}
	}
	if len(rules.SourcePreferences) > 0 && enabledSources == 0 {
		return errors.New("at least one ranking source must be enabled")
	}
	return nil
}
