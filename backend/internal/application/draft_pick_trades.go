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

func (service *DraftService) CreatePickTrade(
	ctx context.Context,
	leagueID string,
	teamOne, teamTwo int,
	teamOneReceives, teamTwoReceives []int,
	teamOneFuture, teamTwoFuture []draft.FuturePick,
	teamOneBudget, teamTwoBudget float64,
) (draft.Snapshot, error) {
	service.mu.Lock()
	defer service.mu.Unlock()
	configuration, err := service.configuration(ctx, leagueID)
	if err != nil {
		return draft.Snapshot{}, err
	}
	repository, ok := service.repository.(DraftPickTradeRepository)
	if !ok {
		return draft.Snapshot{}, errors.New("draft-pick trade storage is unavailable")
	}
	if teamOne < 1 || teamOne > configuration.Rules.TeamCount || teamTwo < 1 || teamTwo > configuration.Rules.TeamCount || teamOne == teamTwo {
		return draft.Snapshot{}, errors.New("choose two different league teams")
	}
	if len(teamOneReceives)+len(teamTwoReceives)+len(teamOneFuture)+len(teamTwoFuture) == 0 && teamOneBudget == 0 && teamTwoBudget == 0 {
		return draft.Snapshot{}, errors.New("add at least one draft asset to the trade")
	}
	if configuration.Rules.DraftType == league.DraftTypeAuction && len(teamOneReceives)+len(teamTwoReceives) > 0 {
		return draft.Snapshot{}, errors.New("the active auction draft does not have numbered picks")
	}
	if configuration.Rules.LeagueFormat == league.LeagueFormatRedraft && len(teamOneFuture)+len(teamTwoFuture) > 0 {
		return draft.Snapshot{}, errors.New("redraft leagues cannot trade future-season picks")
	}
	if (teamOneBudget != 0 || teamTwoBudget != 0) && (configuration.Rules.DraftType != league.DraftTypeAuction || !configuration.Rules.AuctionBudgetTrades) {
		return draft.Snapshot{}, errors.New("auction budget trading is not enabled for this league")
	}
	if teamOneBudget < 0 || teamTwoBudget < 0 {
		return draft.Snapshot{}, errors.New("auction budget amounts cannot be negative")
	}
	events, err := service.listDraftEvents(ctx, leagueID, configuration.Rules.Season)
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
	futureOwners := futurePickOwnership(configuration.Rules, trades)
	futureSeen := make(map[string]bool, len(teamOneFuture)+len(teamTwoFuture))
	if err = validateFuturePicks(teamOneFuture, teamTwo, teamOne, futureOwners, futureSeen, configuration.Rules); err != nil {
		return draft.Snapshot{}, err
	}
	if err = validateFuturePicks(teamTwoFuture, teamOne, teamTwo, futureOwners, futureSeen, configuration.Rules); err != nil {
		return draft.Snapshot{}, err
	}
	if err = validateAuctionBudgetTrade(events, trades, teamOne, teamTwo, teamOneBudget, teamTwoBudget, configuration.Rules); err != nil {
		return draft.Snapshot{}, err
	}
	trade := draft.PickTrade{
		LeagueID: leagueID, TeamOneNumber: teamOne, TeamTwoNumber: teamTwo,
		TeamOneReceives: slices.Clone(teamOneReceives), TeamTwoReceives: slices.Clone(teamTwoReceives),
		TeamOneFuturePicks: slices.Clone(teamOneFuture), TeamTwoFuturePicks: slices.Clone(teamTwoFuture),
		TeamOneAuctionBudget: teamOneBudget, TeamTwoAuctionBudget: teamTwoBudget, Season: configuration.Rules.Season,
	}
	if _, err = repository.SavePickTrade(ctx, trade); err != nil {
		return draft.Snapshot{}, err
	}
	return service.Snapshot(ctx, leagueID)
}

func (service *DraftService) DeletePickTrade(ctx context.Context, leagueID string, tradeID int64) (draft.Snapshot, error) {
	service.mu.Lock()
	defer service.mu.Unlock()
	configuration, err := service.configuration(ctx, leagueID)
	if err != nil {
		return draft.Snapshot{}, err
	}
	repository, ok := service.repository.(DraftPickTradeRepository)
	if !ok {
		return draft.Snapshot{}, errors.New("draft-pick trade storage is unavailable")
	}
	events, err := service.listDraftEvents(ctx, leagueID, configuration.Rules.Season)
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
	if selected.Season != 0 && selected.Season < configuration.Rules.Season {
		return draft.Snapshot{}, errors.New("trade cannot be reversed because its season is complete")
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
		if futureTradesOverlap(*selected, trade) {
			return draft.Snapshot{}, errors.New("trade cannot be reversed because one of its future picks was traded again later")
		}
		if (selected.TeamOneAuctionBudget != 0 || selected.TeamTwoAuctionBudget != 0) &&
			(trade.TeamOneAuctionBudget != 0 || trade.TeamTwoAuctionBudget != 0) &&
			(trade.TeamOneNumber == selected.TeamOneNumber || trade.TeamOneNumber == selected.TeamTwoNumber || trade.TeamTwoNumber == selected.TeamOneNumber || trade.TeamTwoNumber == selected.TeamTwoNumber) {
			return draft.Snapshot{}, errors.New("trade cannot be reversed because its auction budget was traded again later")
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

func futureTradesOverlap(first, second draft.PickTrade) bool {
	keys := make(map[string]bool)
	for _, pick := range append(slices.Clone(first.TeamOneFuturePicks), first.TeamTwoFuturePicks...) {
		keys[futurePickKey(pick.Season, pick.Round, pick.OriginalTeamNumber)] = true
	}
	for _, pick := range append(slices.Clone(second.TeamOneFuturePicks), second.TeamTwoFuturePicks...) {
		if keys[futurePickKey(pick.Season, pick.Round, pick.OriginalTeamNumber)] {
			return true
		}
	}
	return false
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

func validateFuturePicks(picks []draft.FuturePick, expectedOwner, newOwner int, owners map[string]int, seen map[string]bool, rules league.Rules) error {
	for _, pick := range picks {
		if pick.Season <= rules.Season || pick.Season > rules.Season+rules.FuturePickSeasons || pick.Round < 1 || pick.Round > rules.RookieDraftRounds || pick.OriginalTeamNumber < 1 || pick.OriginalTeamNumber > rules.TeamCount {
			return fmt.Errorf("%d round %d is not a tradeable future pick in this league", pick.Season, pick.Round)
		}
		key := futurePickKey(pick.Season, pick.Round, pick.OriginalTeamNumber)
		if seen[key] {
			return fmt.Errorf("%d round %d pick from %s appears more than once", pick.Season, pick.Round, teamName(rules, pick.OriginalTeamNumber))
		}
		seen[key] = true
		if owners[key] != expectedOwner {
			return fmt.Errorf("%d round %d pick from %s is currently owned by %s", pick.Season, pick.Round, teamName(rules, pick.OriginalTeamNumber), teamName(rules, owners[key]))
		}
		owners[key] = newOwner
	}
	return nil
}

func validateAuctionBudgetTrade(events []draft.Event, trades []draft.PickTrade, teamOne, teamTwo int, teamOneReceives, teamTwoReceives float64, rules league.Rules) error {
	if teamOneReceives == 0 && teamTwoReceives == 0 {
		return nil
	}
	remaining := auctionTeamBudgets(rules, events, trades)
	remaining[teamOne] += teamOneReceives - teamTwoReceives
	remaining[teamTwo] += teamTwoReceives - teamOneReceives
	if remaining[teamOne] < 0 {
		return fmt.Errorf("%s does not have enough auction budget for this trade", teamName(rules, teamOne))
	}
	if remaining[teamTwo] < 0 {
		return fmt.Errorf("%s does not have enough auction budget for this trade", teamName(rules, teamTwo))
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
		if trade.Season == 0 || trade.Season == rules.Season {
			for _, pick := range trade.TeamOneReceives {
				owners[pick] = trade.TeamOneNumber
			}
			for _, pick := range trade.TeamTwoReceives {
				owners[pick] = trade.TeamTwoNumber
			}
		}
		for _, pick := range trade.TeamOneFuturePicks {
			applyFuturePickToCurrentOwners(owners, pick, trade.TeamOneNumber, rules)
		}
		for _, pick := range trade.TeamTwoFuturePicks {
			applyFuturePickToCurrentOwners(owners, pick, trade.TeamTwoNumber, rules)
		}
	}
	return owners
}

func enrichPickTrades(rules league.Rules, trades []draft.PickTrade) []draft.PickTrade {
	result := slices.Clone(trades)
	for index := range result {
		result[index].TeamOneName = teamName(rules, result[index].TeamOneNumber)
		result[index].TeamTwoName = teamName(rules, result[index].TeamTwoNumber)
		if result[index].Season == 0 {
			result[index].Season = rules.Season
		}
		for pickIndex := range result[index].TeamOneFuturePicks {
			result[index].TeamOneFuturePicks[pickIndex].OriginalTeamName = teamName(rules, result[index].TeamOneFuturePicks[pickIndex].OriginalTeamNumber)
		}
		for pickIndex := range result[index].TeamTwoFuturePicks {
			result[index].TeamTwoFuturePicks[pickIndex].OriginalTeamName = teamName(rules, result[index].TeamTwoFuturePicks[pickIndex].OriginalTeamNumber)
		}
	}
	return result
}

func draftPickSlots(rules league.Rules, trades []draft.PickTrade, used int) []draft.PickSlot {
	owners := pickOwnership(rules, trades)
	slots := make([]draft.PickSlot, 0, totalDraftPicks(rules)+rules.FuturePickSeasons*rules.RookieDraftRounds*rules.TeamCount)
	if rules.DraftType != league.DraftTypeAuction {
		for pick := 1; pick <= totalDraftPicks(rules); pick++ {
			original := pickOwner(pick, rules)
			owner := owners[pick]
			slots = append(slots, draft.PickSlot{
				Season: rules.Season, OverallNumber: pick, Round: (pick-1)/rules.TeamCount + 1, PickInRound: (pick-1)%rules.TeamCount + 1,
				OriginalTeamNumber: original, OriginalTeamName: teamName(rules, original),
				OwnerTeamNumber: owner, OwnerTeamName: teamName(rules, owner), IsUsed: pick <= used,
			})
		}
	}
	if rules.LeagueFormat == league.LeagueFormatDynasty {
		futureOwners := futurePickOwnership(rules, trades)
		for season := rules.Season + 1; season <= rules.Season+rules.FuturePickSeasons; season++ {
			for round := 1; round <= rules.RookieDraftRounds; round++ {
				for original := 1; original <= rules.TeamCount; original++ {
					owner := futureOwners[futurePickKey(season, round, original)]
					slots = append(slots, draft.PickSlot{
						Season: season, Round: round, OriginalTeamNumber: original, OriginalTeamName: teamName(rules, original),
						OwnerTeamNumber: owner, OwnerTeamName: teamName(rules, owner),
					})
				}
			}
		}
	}
	return slots
}

func ownerForPick(pick int, rules league.Rules, trades []draft.PickTrade) int {
	return pickOwnership(rules, trades)[pick]
}

func nextUserPickWithTrades(current int, rules league.Rules, trades []draft.PickTrade) int {
	for pick := current + 1; pick <= totalDraftPicks(rules); pick++ {
		if ownerForPick(pick, rules, trades) == rules.DraftPosition {
			return pick
		}
	}
	return 0
}

func applyFuturePickToCurrentOwners(owners map[int]int, pick draft.FuturePick, owner int, rules league.Rules) {
	if pick.Season != rules.Season || rules.DraftType == league.DraftTypeAuction {
		return
	}
	for overall := 1; overall <= totalDraftPicks(rules); overall++ {
		if (overall-1)/rules.TeamCount+1 == pick.Round && pickOwner(overall, rules) == pick.OriginalTeamNumber {
			owners[overall] = owner
			return
		}
	}
}

func futurePickKey(season, round, originalTeam int) string {
	return fmt.Sprintf("%d:%d:%d", season, round, originalTeam)
}

func futurePickOwnership(rules league.Rules, trades []draft.PickTrade) map[string]int {
	owners := make(map[string]int)
	for season := rules.Season + 1; season <= rules.Season+rules.FuturePickSeasons; season++ {
		for round := 1; round <= rules.RookieDraftRounds; round++ {
			for team := 1; team <= rules.TeamCount; team++ {
				owners[futurePickKey(season, round, team)] = team
			}
		}
	}
	for _, trade := range trades {
		for _, pick := range trade.TeamOneFuturePicks {
			owners[futurePickKey(pick.Season, pick.Round, pick.OriginalTeamNumber)] = trade.TeamOneNumber
		}
		for _, pick := range trade.TeamTwoFuturePicks {
			owners[futurePickKey(pick.Season, pick.Round, pick.OriginalTeamNumber)] = trade.TeamTwoNumber
		}
	}
	return owners
}

func auctionBudgetAdjustments(trades []draft.PickTrade, season int) map[int]float64 {
	adjustments := make(map[int]float64)
	for _, trade := range trades {
		if trade.Season != 0 && trade.Season != season {
			continue
		}
		adjustments[trade.TeamOneNumber] += trade.TeamOneAuctionBudget - trade.TeamTwoAuctionBudget
		adjustments[trade.TeamTwoNumber] += trade.TeamTwoAuctionBudget - trade.TeamOneAuctionBudget
	}
	return adjustments
}

func auctionTeamBudgets(rules league.Rules, events []draft.Event, trades []draft.PickTrade) map[int]float64 {
	adjustments := auctionBudgetAdjustments(trades, rules.Season)
	budgets := make(map[int]float64, rules.TeamCount)
	for team := 1; team <= rules.TeamCount; team++ {
		budgets[team] = rules.AuctionBudget + adjustments[team]
	}
	budgets[rules.DraftPosition] -= rules.MyKeeperSpend
	for _, event := range replay(events).activeEvents {
		budgets[event.TeamNumber] -= event.Cost
	}
	return budgets
}
