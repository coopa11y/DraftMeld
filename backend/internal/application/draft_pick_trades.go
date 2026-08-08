package application

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/coopa11y/DraftMeld/backend/internal/domain/draft"
	"github.com/coopa11y/DraftMeld/backend/internal/domain/league"
)

type DraftPickTradeRepository interface {
	ListPickTrades(context.Context, string) ([]draft.PickTrade, error)
	SavePickTrade(context.Context, draft.PickTrade) (draft.PickTrade, error)
	DeletePickTrade(context.Context, string, int64) (bool, error)
}

func (service *DraftService) CreatePickTrade(ctx context.Context, leagueID string, teamOne, teamTwo int, teamOneReceives, teamTwoReceives []int) (draft.Snapshot, error) {
	service.mu.Lock()
	defer service.mu.Unlock()
	configuration, err := service.configuration(ctx, leagueID)
	if err != nil {
		return draft.Snapshot{}, err
	}
	if configuration.Rules.DraftType == league.DraftTypeAuction {
		return draft.Snapshot{}, errors.New("draft-pick trades are only available for snake and linear drafts")
	}
	repository, ok := service.repository.(DraftPickTradeRepository)
	if !ok {
		return draft.Snapshot{}, errors.New("draft-pick trade storage is unavailable")
	}
	if teamOne < 1 || teamOne > configuration.Rules.TeamCount || teamTwo < 1 || teamTwo > configuration.Rules.TeamCount || teamOne == teamTwo {
		return draft.Snapshot{}, errors.New("choose two different league teams")
	}
	if len(teamOneReceives) == 0 || len(teamTwoReceives) == 0 {
		return draft.Snapshot{}, errors.New("each team must receive at least one draft pick")
	}
	events, err := service.repository.List(ctx, leagueID)
	if err != nil {
		return draft.Snapshot{}, err
	}
	usedPicks := len(replay(events).activeEvents)
	trades, err := repository.ListPickTrades(ctx, leagueID)
	if err != nil {
		return draft.Snapshot{}, err
	}
	owners := pickOwnership(configuration.Rules, trades)
	seen := make(map[int]bool, len(teamOneReceives)+len(teamTwoReceives))
	if err = validateTradePicks(teamOneReceives, teamTwo, teamOne, usedPicks, owners, seen, configuration.Rules); err != nil {
		return draft.Snapshot{}, err
	}
	if err = validateTradePicks(teamTwoReceives, teamOne, teamTwo, usedPicks, owners, seen, configuration.Rules); err != nil {
		return draft.Snapshot{}, err
	}
	trade := draft.PickTrade{
		LeagueID: leagueID, TeamOneNumber: teamOne, TeamTwoNumber: teamTwo,
		TeamOneReceives: slices.Clone(teamOneReceives), TeamTwoReceives: slices.Clone(teamTwoReceives),
	}
	if _, err = repository.SavePickTrade(ctx, trade); err != nil {
		return draft.Snapshot{}, err
	}
	return service.Snapshot(ctx, leagueID)
}

func (service *DraftService) DeletePickTrade(ctx context.Context, leagueID string, tradeID int64) (draft.Snapshot, error) {
	service.mu.Lock()
	defer service.mu.Unlock()
	if _, err := service.configuration(ctx, leagueID); err != nil {
		return draft.Snapshot{}, err
	}
	repository, ok := service.repository.(DraftPickTradeRepository)
	if !ok {
		return draft.Snapshot{}, errors.New("draft-pick trade storage is unavailable")
	}
	events, err := service.repository.List(ctx, leagueID)
	if err != nil {
		return draft.Snapshot{}, err
	}
	trades, err := repository.ListPickTrades(ctx, leagueID)
	if err != nil {
		return draft.Snapshot{}, err
	}
	var selected *draft.PickTrade
	for index := range trades {
		if trades[index].ID == tradeID {
			selected = &trades[index]
			break
		}
	}
	if selected == nil {
		return draft.Snapshot{}, errors.New("that draft-pick trade was not found")
	}
	usedPicks := len(replay(events).activeEvents)
	affected := append(slices.Clone(selected.TeamOneReceives), selected.TeamTwoReceives...)
	for _, pick := range affected {
		if pick <= usedPicks {
			return draft.Snapshot{}, fmt.Errorf("trade cannot be reversed because pick %d has already been used", pick)
		}
	}
	for _, trade := range trades {
		if trade.ID <= selected.ID {
			continue
		}
		for _, pick := range affected {
			if slices.Contains(trade.TeamOneReceives, pick) || slices.Contains(trade.TeamTwoReceives, pick) {
				return draft.Snapshot{}, fmt.Errorf("trade cannot be reversed because pick %d was traded again later", pick)
			}
		}
	}
	deleted, err := repository.DeletePickTrade(ctx, leagueID, tradeID)
	if err != nil {
		return draft.Snapshot{}, err
	}
	if !deleted {
		return draft.Snapshot{}, errors.New("that draft-pick trade was not found")
	}
	return service.Snapshot(ctx, leagueID)
}

func validateTradePicks(picks []int, expectedOwner, newOwner, used int, owners map[int]int, seen map[int]bool, rules league.Rules) error {
	for _, pick := range picks {
		if pick <= used || pick > totalDraftPicks(rules) {
			return fmt.Errorf("pick %d is not an unused pick in this draft", pick)
		}
		if seen[pick] {
			return fmt.Errorf("pick %d appears more than once in this trade", pick)
		}
		seen[pick] = true
		if owners[pick] != expectedOwner {
			return fmt.Errorf("pick %d is currently owned by %s, not %s", pick, teamName(rules, owners[pick]), teamName(rules, expectedOwner))
		}
		owners[pick] = newOwner
	}
	return nil
}

func (service *DraftService) pickTrades(ctx context.Context, leagueID string) ([]draft.PickTrade, error) {
	repository, ok := service.repository.(DraftPickTradeRepository)
	if !ok {
		return []draft.PickTrade{}, nil
	}
	return repository.ListPickTrades(ctx, leagueID)
}

func pickOwnership(rules league.Rules, trades []draft.PickTrade) map[int]int {
	owners := make(map[int]int, totalDraftPicks(rules))
	for pick := 1; pick <= totalDraftPicks(rules); pick++ {
		owners[pick] = pickOwner(pick, rules)
	}
	for _, trade := range trades {
		for _, pick := range trade.TeamOneReceives {
			owners[pick] = trade.TeamOneNumber
		}
		for _, pick := range trade.TeamTwoReceives {
			owners[pick] = trade.TeamTwoNumber
		}
	}
	return owners
}

func enrichPickTrades(rules league.Rules, trades []draft.PickTrade) []draft.PickTrade {
	result := slices.Clone(trades)
	for index := range result {
		result[index].TeamOneName = teamName(rules, result[index].TeamOneNumber)
		result[index].TeamTwoName = teamName(rules, result[index].TeamTwoNumber)
	}
	return result
}

func draftPickSlots(rules league.Rules, trades []draft.PickTrade, used int) []draft.PickSlot {
	owners := pickOwnership(rules, trades)
	slots := make([]draft.PickSlot, 0, totalDraftPicks(rules))
	for pick := 1; pick <= totalDraftPicks(rules); pick++ {
		original := pickOwner(pick, rules)
		owner := owners[pick]
		slots = append(slots, draft.PickSlot{
			OverallNumber: pick, Round: (pick-1)/rules.TeamCount + 1, PickInRound: (pick-1)%rules.TeamCount + 1,
			OriginalTeamNumber: original, OriginalTeamName: teamName(rules, original),
			OwnerTeamNumber: owner, OwnerTeamName: teamName(rules, owner), IsUsed: pick <= used,
		})
	}
	return slots
}

func ownerForPick(pick int, rules league.Rules, trades []draft.PickTrade) int {
	for index := len(trades) - 1; index >= 0; index-- {
		if slices.Contains(trades[index].TeamOneReceives, pick) {
			return trades[index].TeamOneNumber
		}
		if slices.Contains(trades[index].TeamTwoReceives, pick) {
			return trades[index].TeamTwoNumber
		}
	}
	return pickOwner(pick, rules)
}

func nextUserPickWithTrades(current int, rules league.Rules, trades []draft.PickTrade) int {
	for pick := current + 1; pick <= totalDraftPicks(rules); pick++ {
		if ownerForPick(pick, rules, trades) == rules.DraftPosition {
			return pick
		}
	}
	return 0
}
