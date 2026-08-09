package application

import (
	"fmt"
	"slices"

	"github.com/coopa11y/DraftMeld/backend/internal/domain/draft"
	"github.com/coopa11y/DraftMeld/backend/internal/domain/league"
)

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
