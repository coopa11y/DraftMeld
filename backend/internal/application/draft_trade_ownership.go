package application

import (
	"context"
	"fmt"
	"slices"

	"github.com/coopa11y/DraftMeld/backend/internal/domain/draft"
	"github.com/coopa11y/DraftMeld/backend/internal/domain/league"
)

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
