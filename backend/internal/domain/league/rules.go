package league

import (
	"errors"
	"fmt"
	"math"
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

type Rules struct {
	Name          string             `json:"name"`
	TeamCount     int                `json:"teamCount"`
	DraftPosition int                `json:"draftPosition"`
	DraftType     DraftType          `json:"draftType"`
	RosterSlots   []RosterSlot       `json:"rosterSlots"`
	ScoringRules  map[string]float64 `json:"scoringRules"`
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
	switch rules.DraftType {
	case DraftTypeSnake, DraftTypeLinear, DraftTypeAuction:
	default:
		return fmt.Errorf("unsupported draft type: %q", rules.DraftType)
	}
	if len(rules.RosterSlots) == 0 {
		return errors.New("at least one roster slot is required")
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
	return nil
}
