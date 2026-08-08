package application

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

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

func budgetReversalWouldOverdraw(rules league.Rules, events []draft.Event, trades []draft.PickTrade) bool {
	for _, remaining := range tradeBudgetBalances(rules, events, trades) {
		if remaining < 0 {
			return true
		}
	}
	return false
}

func playerTradesOverlap(first, second draft.PickTrade) bool {
	players := make(map[string]bool)
	for _, playerID := range append(slices.Clone(first.TeamOnePlayers), first.TeamTwoPlayers...) {
		players[playerID] = true
	}
	for _, playerID := range append(slices.Clone(second.TeamOnePlayers), second.TeamTwoPlayers...) {
		if players[playerID] {
			return true
		}
	}
	return false
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

func conditionalPickHasBeenUsed(rules league.Rules, season, round, originalTeam, used int) bool {
	if season != rules.Season || rules.DraftType == league.DraftTypeAuction {
		return false
	}
	for overall := 1; overall <= totalDraftPicks(rules); overall++ {
		if (overall-1)/rules.TeamCount+1 == round && pickOwner(overall, rules) == originalTeam {
			return overall <= used
		}
	}
	return false
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

func validateFuturePicks(picks []draft.FuturePick, expectedOwner, newOwner int, owners map[string]int, seen, locked map[string]bool, rules league.Rules) error {
	for _, pick := range picks {
		if pick.Season <= rules.Season || pick.Season > rules.Season+rules.FuturePickSeasons || pick.Round < 1 || pick.Round > rules.RookieDraftRounds || pick.OriginalTeamNumber < 1 || pick.OriginalTeamNumber > rules.TeamCount {
			return fmt.Errorf("%d round %d is not a tradeable future pick in this league", pick.Season, pick.Round)
		}
		key := futurePickKey(pick.Season, pick.Round, pick.OriginalTeamNumber)
		if locked[key] {
			return fmt.Errorf("%d round %d pick from %s is locked by an unresolved condition", pick.Season, pick.Round, teamName(rules, pick.OriginalTeamNumber))
		}
		if len(strings.TrimSpace(pick.Condition)) > 300 {
			return errors.New("conditional pick descriptions cannot exceed 300 characters")
		}
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

func pendingConditionalPicks(trades []draft.PickTrade) map[string]bool {
	locked := make(map[string]bool)
	for _, trade := range trades {
		for _, pick := range append(slices.Clone(trade.TeamOneFuturePicks), trade.TeamTwoFuturePicks...) {
			if pick.Condition != "" && pick.ConditionStatus == "pending" {
				locked[futurePickKey(pick.Season, pick.Round, pick.OriginalTeamNumber)] = true
			}
		}
	}
	return locked
}

func normalizeConditionalPicks(picks []draft.FuturePick) []draft.FuturePick {
	result := slices.Clone(picks)
	for index := range result {
		if result[index].Condition != "" {
			result[index].ConditionStatus = "pending"
		} else {
			result[index].ConditionStatus = ""
		}
	}
	return result
}

func (service *DraftService) validateTradePlayers(ctx context.Context, leagueID string, trades []draft.PickTrade, teamOne, teamTwo int, teamOneReceives, teamTwoReceives []string, rules league.Rules) error {
	if len(teamOneReceives)+len(teamTwoReceives) == 0 {
		return nil
	}
	events, err := service.repository.List(ctx, leagueID)
	if err != nil {
		return err
	}
	events, err = resolveDraftEventAliases(ctx, service.repository, events)
	if err != nil {
		return err
	}
	owners := dynastyPlayerOwnership(events, trades)
	seen := make(map[string]bool, len(teamOneReceives)+len(teamTwoReceives))
	validate := func(players []string, expected, receiver int) error {
		for _, playerID := range players {
			if playerID == "" || seen[playerID] {
				return errors.New("each player can appear only once in a trade")
			}
			seen[playerID] = true
			if owners[playerID] != expected {
				return fmt.Errorf("player %s is currently owned by %s", playerID, teamName(rules, owners[playerID]))
			}
			owners[playerID] = receiver
		}
		return nil
	}
	if err = validate(teamOneReceives, teamTwo, teamOne); err != nil {
		return err
	}
	return validate(teamTwoReceives, teamOne, teamTwo)
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

func validateTradeBudgets(events []draft.Event, trades []draft.PickTrade, teamOne, teamTwo int, teamOneReceives, teamTwoReceives []draft.BudgetAsset, rules league.Rules) error {
	if len(teamOneReceives)+len(teamTwoReceives) == 0 {
		return nil
	}
	balances := tradeBudgetBalances(rules, events, trades)
	seen := make(map[string]bool, len(teamOneReceives)+len(teamTwoReceives))
	apply := func(assets []draft.BudgetAsset, source, receiver int) error {
		for _, asset := range assets {
			if asset.Amount <= 0 {
				return errors.New("trade budget amounts must be positive")
			}
			if asset.Season != rules.Season && (rules.LeagueFormat != league.LeagueFormatDynasty || asset.Season <= rules.Season || asset.Season > rules.Season+rules.FuturePickSeasons) {
				return fmt.Errorf("%d budget is outside this league's tradeable seasons", asset.Season)
			}
			switch asset.Kind {
			case "auction":
				if rules.DraftType != league.DraftTypeAuction || !rules.AuctionBudgetTrades {
					return errors.New("auction budget trading is not enabled for this league")
				}
			case "faab":
				if !rules.FAABTrades || rules.FAABBudget <= 0 {
					return errors.New("FAAB trading is not enabled for this league")
				}
			default:
				return fmt.Errorf("unsupported trade budget kind: %q", asset.Kind)
			}
			key := budgetKey(asset.Kind, asset.Season, source)
			requestKey := fmt.Sprintf("%s:%d:%d", asset.Kind, asset.Season, receiver)
			if seen[requestKey] {
				return fmt.Errorf("combine duplicate %s budget entries for %d", asset.Kind, asset.Season)
			}
			seen[requestKey] = true
			if balances[key] < asset.Amount {
				return fmt.Errorf("%s does not have enough %d %s budget", teamName(rules, source), asset.Season, strings.ToUpper(asset.Kind))
			}
			balances[key] -= asset.Amount
			balances[budgetKey(asset.Kind, asset.Season, receiver)] += asset.Amount
		}
		return nil
	}
	if err := apply(teamOneReceives, teamTwo, teamOne); err != nil {
		return err
	}
	return apply(teamTwoReceives, teamOne, teamTwo)
}

func budgetKey(kind string, season, team int) string {
	return fmt.Sprintf("%s:%d:%d", kind, season, team)
}

func tradeBudgetBalances(rules league.Rules, events []draft.Event, trades []draft.PickTrade) map[string]float64 {
	balances := make(map[string]float64)
	lastSeason := rules.Season
	if rules.LeagueFormat == league.LeagueFormatDynasty {
		lastSeason += rules.FuturePickSeasons
	}
	for season := rules.Season; season <= lastSeason; season++ {
		for team := 1; team <= rules.TeamCount; team++ {
			if rules.DraftType == league.DraftTypeAuction {
				balances[budgetKey("auction", season, team)] = rules.AuctionBudget
			}
			if rules.FAABBudget > 0 {
				balances[budgetKey("faab", season, team)] = rules.FAABBudget
			}
		}
	}
	balances[budgetKey("auction", rules.Season, userTeamNumber(rules))] -= rules.MyKeeperSpend
	for _, event := range replay(events).activeEvents {
		balances[budgetKey("auction", rules.Season, event.TeamNumber)] -= event.Cost
	}
	for _, trade := range trades {
		oneBudgets, twoBudgets := tradeBudgetAssets(trade)
		for _, asset := range oneBudgets {
			balances[budgetKey(asset.Kind, asset.Season, trade.TeamTwoNumber)] -= asset.Amount
			balances[budgetKey(asset.Kind, asset.Season, trade.TeamOneNumber)] += asset.Amount
		}
		for _, asset := range twoBudgets {
			balances[budgetKey(asset.Kind, asset.Season, trade.TeamOneNumber)] -= asset.Amount
			balances[budgetKey(asset.Kind, asset.Season, trade.TeamTwoNumber)] += asset.Amount
		}
	}
	return balances
}

func draftBudgetBalances(rules league.Rules, events []draft.Event, trades []draft.PickTrade) []draft.BudgetBalance {
	balances := tradeBudgetBalances(rules, events, trades)
	result := make([]draft.BudgetBalance, 0, len(balances))
	lastSeason := rules.Season
	if rules.LeagueFormat == league.LeagueFormatDynasty {
		lastSeason += rules.FuturePickSeasons
	}
	for season := rules.Season; season <= lastSeason; season++ {
		for team := 1; team <= rules.TeamCount; team++ {
			if rules.DraftType == league.DraftTypeAuction && rules.AuctionBudgetTrades {
				result = append(result, draft.BudgetBalance{TeamNumber: team, Season: season, Kind: "auction", Remaining: balances[budgetKey("auction", season, team)]})
			}
			if rules.FAABTrades && rules.FAABBudget > 0 {
				result = append(result, draft.BudgetBalance{TeamNumber: team, Season: season, Kind: "faab", Remaining: balances[budgetKey("faab", season, team)]})
			}
		}
	}
	return result
}

func tradeBudgetAssets(trade draft.PickTrade) ([]draft.BudgetAsset, []draft.BudgetAsset) {
	one, two := slices.Clone(trade.TeamOneBudgets), slices.Clone(trade.TeamTwoBudgets)
	if len(one) == 0 && trade.TeamOneAuctionBudget > 0 {
		one = append(one, draft.BudgetAsset{Kind: "auction", Season: trade.Season, Amount: trade.TeamOneAuctionBudget})
	}
	if len(two) == 0 && trade.TeamTwoAuctionBudget > 0 {
		two = append(two, draft.BudgetAsset{Kind: "auction", Season: trade.Season, Amount: trade.TeamTwoAuctionBudget})
	}
	return one, two
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
		if ownerForPick(pick, rules, trades) == userTeamNumber(rules) {
			return pick
		}
	}
	return 0
}

func applyFuturePickToCurrentOwners(owners map[int]int, pick draft.FuturePick, owner int, rules league.Rules) {
	if pick.Season != rules.Season || rules.DraftType == league.DraftTypeAuction || pick.ConditionStatus == "not-met" {
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
			if pick.ConditionStatus != "not-met" {
				owners[futurePickKey(pick.Season, pick.Round, pick.OriginalTeamNumber)] = trade.TeamOneNumber
			}
		}
		for _, pick := range trade.TeamTwoFuturePicks {
			if pick.ConditionStatus != "not-met" {
				owners[futurePickKey(pick.Season, pick.Round, pick.OriginalTeamNumber)] = trade.TeamTwoNumber
			}
		}
	}
	return owners
}

func dynastyPlayerOwnership(events []draft.Event, trades []draft.PickTrade) map[string]int {
	owners := make(map[string]int)
	for _, event := range replay(events).activeEvents {
		if event.TeamNumber > 0 {
			owners[event.PlayerID] = event.TeamNumber
		}
	}
	for _, trade := range trades {
		for _, playerID := range trade.TeamOnePlayers {
			owners[playerID] = trade.TeamOneNumber
		}
		for _, playerID := range trade.TeamTwoPlayers {
			owners[playerID] = trade.TeamTwoNumber
		}
	}
	return owners
}

func auctionBudgetAdjustments(trades []draft.PickTrade, season int) map[int]float64 {
	adjustments := make(map[int]float64)
	for _, trade := range trades {
		one, two := tradeBudgetAssets(trade)
		for _, asset := range one {
			if asset.Kind == "auction" && asset.Season == season {
				adjustments[trade.TeamOneNumber] += asset.Amount
				adjustments[trade.TeamTwoNumber] -= asset.Amount
			}
		}
		for _, asset := range two {
			if asset.Kind == "auction" && asset.Season == season {
				adjustments[trade.TeamTwoNumber] += asset.Amount
				adjustments[trade.TeamOneNumber] -= asset.Amount
			}
		}
	}
	return adjustments
}

func auctionTeamBudgets(rules league.Rules, events []draft.Event, trades []draft.PickTrade) map[int]float64 {
	adjustments := auctionBudgetAdjustments(trades, rules.Season)
	budgets := make(map[int]float64, rules.TeamCount)
	for team := 1; team <= rules.TeamCount; team++ {
		budgets[team] = rules.AuctionBudget + adjustments[team]
	}
	budgets[userTeamNumber(rules)] -= rules.MyKeeperSpend
	for _, event := range replay(events).activeEvents {
		budgets[event.TeamNumber] -= event.Cost
	}
	return budgets
}
