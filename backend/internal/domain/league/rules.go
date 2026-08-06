package league

import (
	"errors"
	"fmt"
)

type DraftType string

const (
	DraftTypeSnake   DraftType = "snake"
	DraftTypeLinear  DraftType = "linear"
	DraftTypeAuction DraftType = "auction"
)

type RosterSlot struct {
	Name       string
	Count      int
	Positions  []string
	IsStarting bool
}

type Rules struct {
	Name         string
	TeamCount    int
	DraftType    DraftType
	RosterSlots  []RosterSlot
	ScoringRules map[string]float64
}

func (rules Rules) Validate() error {
	if rules.Name == "" {
		return errors.New("league name is required")
	}
	if rules.TeamCount < 2 || rules.TeamCount > 32 {
		return fmt.Errorf("team count must be between 2 and 32: %d", rules.TeamCount)
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
	return nil
}
