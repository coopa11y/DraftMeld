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
