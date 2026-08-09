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

type DraftTradeUpdateRepository interface {
	UpdatePickTrade(context.Context, draft.PickTrade) error
}

func (service *DraftService) CreatePickTrade(
	ctx context.Context,
	leagueID string,
	teamOne, teamTwo int,
	teamOneReceives, teamTwoReceives []int,
	teamOneFuture, teamTwoFuture []draft.FuturePick,
	teamOneBudget, teamTwoBudget float64,
) (draft.Snapshot, error) {
	return service.CreateDraftTrade(ctx, leagueID, teamOne, teamTwo, teamOneReceives, teamTwoReceives,
		teamOneFuture, teamTwoFuture, nil, nil, nil, nil, teamOneBudget, teamTwoBudget)
}

func (service *DraftService) CreateDraftTrade(
	ctx context.Context,
	leagueID string,
	teamOne, teamTwo int,
	teamOneReceives, teamTwoReceives []int,
	teamOneFuture, teamTwoFuture []draft.FuturePick,
	teamOnePlayers, teamTwoPlayers []string,
	teamOneBudgets, teamTwoBudgets []draft.BudgetAsset,
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
	if len(teamOneReceives)+len(teamTwoReceives)+len(teamOneFuture)+len(teamTwoFuture)+len(teamOnePlayers)+len(teamTwoPlayers)+len(teamOneBudgets)+len(teamTwoBudgets) == 0 && teamOneBudget == 0 && teamTwoBudget == 0 {
		return draft.Snapshot{}, errors.New("add at least one draft asset to the trade")
	}
	if configuration.Rules.DraftType == league.DraftTypeAuction && len(teamOneReceives)+len(teamTwoReceives) > 0 {
		return draft.Snapshot{}, errors.New("the active auction draft does not have numbered picks")
	}
	if configuration.Rules.LeagueFormat == league.LeagueFormatRedraft && len(teamOneFuture)+len(teamTwoFuture) > 0 {
		return draft.Snapshot{}, errors.New("redraft leagues cannot trade future-season picks")
	}
	if configuration.Rules.LeagueFormat != league.LeagueFormatDynasty && len(teamOnePlayers)+len(teamTwoPlayers) > 0 {
		return draft.Snapshot{}, errors.New("player assets are only available in dynasty leagues")
	}
	if (teamOneBudget != 0 || teamTwoBudget != 0) && (configuration.Rules.DraftType != league.DraftTypeAuction || !configuration.Rules.AuctionBudgetTrades) {
		return draft.Snapshot{}, errors.New("auction budget trading is not enabled for this league")
	}
	if teamOneBudget < 0 || teamTwoBudget < 0 {
		return draft.Snapshot{}, errors.New("auction budget amounts cannot be negative")
	}
	teamOneFuture = normalizeConditionalPicks(teamOneFuture)
	teamTwoFuture = normalizeConditionalPicks(teamTwoFuture)
	if teamOneBudget > 0 {
		teamOneBudgets = append(slices.Clone(teamOneBudgets), draft.BudgetAsset{Kind: "auction", Season: configuration.Rules.Season, Amount: teamOneBudget})
	}
	if teamTwoBudget > 0 {
		teamTwoBudgets = append(slices.Clone(teamTwoBudgets), draft.BudgetAsset{Kind: "auction", Season: configuration.Rules.Season, Amount: teamTwoBudget})
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
	lockedFuturePicks := pendingConditionalPicks(trades)
	futureSeen := make(map[string]bool, len(teamOneFuture)+len(teamTwoFuture))
	if err = validateFuturePicks(teamOneFuture, teamTwo, teamOne, futureOwners, futureSeen, lockedFuturePicks, configuration.Rules); err != nil {
		return draft.Snapshot{}, err
	}
	if err = validateFuturePicks(teamTwoFuture, teamOne, teamTwo, futureOwners, futureSeen, lockedFuturePicks, configuration.Rules); err != nil {
		return draft.Snapshot{}, err
	}
	if err = service.validateTradePlayers(ctx, leagueID, trades, teamOne, teamTwo, teamOnePlayers, teamTwoPlayers, configuration.Rules); err != nil {
		return draft.Snapshot{}, err
	}
	if err = validateTradeBudgets(events, trades, teamOne, teamTwo, teamOneBudgets, teamTwoBudgets, configuration.Rules); err != nil {
		return draft.Snapshot{}, err
	}
	trade := draft.PickTrade{
		LeagueID: leagueID, TeamOneNumber: teamOne, TeamTwoNumber: teamTwo,
		TeamOneReceives: slices.Clone(teamOneReceives), TeamTwoReceives: slices.Clone(teamTwoReceives),
		TeamOneFuturePicks: slices.Clone(teamOneFuture), TeamTwoFuturePicks: slices.Clone(teamTwoFuture),
		TeamOneAuctionBudget: teamOneBudget, TeamTwoAuctionBudget: teamTwoBudget, Season: configuration.Rules.Season,
		TeamOnePlayers: slices.Clone(teamOnePlayers), TeamTwoPlayers: slices.Clone(teamTwoPlayers),
		TeamOneBudgets: slices.Clone(teamOneBudgets), TeamTwoBudgets: slices.Clone(teamTwoBudgets),
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
		if playerTradesOverlap(*selected, trade) {
			return draft.Snapshot{}, errors.New("trade cannot be reversed because one of its players was traded again later")
		}
	}
	remainingTrades := make([]draft.PickTrade, 0, len(trades)-1)
	for _, trade := range trades {
		if trade.ID != selected.ID {
			remainingTrades = append(remainingTrades, trade)
		}
	}
	if budgetReversalWouldOverdraw(configuration.Rules, events, remainingTrades) {
		return draft.Snapshot{}, errors.New("trade cannot be reversed because a later budget transfer depends on it")
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

func (service *DraftService) ResolveTradeCondition(ctx context.Context, leagueID string, tradeID int64, season, round, originalTeam int, status string) (draft.Snapshot, error) {
	service.mu.Lock()
	defer service.mu.Unlock()
	configuration, err := service.configuration(ctx, leagueID)
	if err != nil {
		return draft.Snapshot{}, err
	}
	if configuration.Rules.LeagueFormat != league.LeagueFormatDynasty || (status != "met" && status != "not-met") {
		return draft.Snapshot{}, errors.New("choose whether the dynasty pick condition was met or not met")
	}
	repository, ok := service.repository.(DraftPickTradeRepository)
	if !ok {
		return draft.Snapshot{}, errors.New("draft trade storage is unavailable")
	}
	updater, ok := service.repository.(DraftTradeUpdateRepository)
	if !ok {
		return draft.Snapshot{}, errors.New("conditional pick updates are unavailable")
	}
	trades, err := repository.ListPickTrades(ctx, leagueID)
	if err != nil {
		return draft.Snapshot{}, err
	}
	for index := range trades {
		if trades[index].ID != tradeID {
			continue
		}
		updated := false
		resolve := func(picks []draft.FuturePick) {
			for pickIndex := range picks {
				pick := &picks[pickIndex]
				if pick.Season == season && pick.Round == round && pick.OriginalTeamNumber == originalTeam && pick.Condition != "" && pick.ConditionStatus == "pending" {
					pick.ConditionStatus = status
					updated = true
				}
			}
		}
		resolve(trades[index].TeamOneFuturePicks)
		resolve(trades[index].TeamTwoFuturePicks)
		if !updated {
			return draft.Snapshot{}, errors.New("that pending pick condition was not found")
		}
		if season < configuration.Rules.Season {
			return draft.Snapshot{}, errors.New("conditions from a completed season cannot be changed")
		}
		if season == configuration.Rules.Season {
			events, listErr := service.listDraftEvents(ctx, leagueID, season)
			if listErr != nil {
				return draft.Snapshot{}, listErr
			}
			if conditionalPickHasBeenUsed(configuration.Rules, season, round, originalTeam, len(replay(events).activeEvents)) {
				return draft.Snapshot{}, errors.New("that condition cannot be changed after its draft pick was used")
			}
		}
		if err = updater.UpdatePickTrade(ctx, trades[index]); err != nil {
			return draft.Snapshot{}, err
		}
		return service.Snapshot(ctx, leagueID)
	}
	return draft.Snapshot{}, errors.New("that draft trade was not found")
}
