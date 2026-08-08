package application

import (
	"context"
	"sort"

	"github.com/coopa11y/DraftMeld/backend/internal/domain/draft"
	"github.com/coopa11y/DraftMeld/backend/internal/domain/league"
)

func (service *DraftService) playersForLeague(ctx context.Context, configuration LeagueConfiguration) ([]draft.Player, map[string]draft.Player, string, int, error) {
	preferences, err := resolvePlayerPreferences(ctx, service.repository, configuration.Rules.PlayerPreferences)
	if err != nil {
		return nil, nil, "", 0, err
	}
	if service.rankings == nil {
		return staticPlayers(service.players, preferences)
	}
	consensus, err := service.rankings.Consensus(ctx, configuration.Rules.SourcePreferences, configuration.Rules.ConsensusMethod)
	if err != nil {
		return nil, nil, "", 0, err
	}
	if len(consensus) == 0 {
		return staticPlayers(service.players, preferences)
	}
	values := make(map[string]leaguePlayerValue)
	if service.projections != nil {
		projected, projectionErr := service.projections.LeagueValues(ctx, configuration.Rules.ScoringRules)
		if projectionErr != nil {
			return nil, nil, "", 0, projectionErr
		}
		for playerID, value := range projected {
			values[playerID] = leaguePlayerValue{points: value.ProjectedPoints, adp: value.ADP, byeWeek: value.ByeWeek}
		}
	}
	positionCounts := make(map[string]int)
	players := make([]draft.Player, 0, len(consensus))
	projectionCount := 0
	for _, ranked := range consensus {
		positionCounts[ranked.Position]++
		value := values[ranked.PlayerKey]
		adp := value.adp
		if adp <= 0 {
			adp = ranked.ADP
		}
		if adp <= 0 {
			adp = float64(ranked.Rank)
		}
		if value.points != 0 {
			projectionCount++
		}
		players = append(players, draft.Player{
			ID: ranked.PlayerKey, Name: ranked.Name, NFLTeam: ranked.Team, Position: ranked.Position,
			ByeWeek: value.byeWeek, OverallRank: ranked.Rank, PositionRank: positionCounts[ranked.Position], ADP: adp,
			Tier: ranked.Tier, ProjectedPoints: value.points, Confidence: ranked.Confidence, RankRange: ranked.RankRange,
			Preference: preferences[ranked.PlayerKey],
		})
	}
	applyReplacementValues(players, configuration.Rules)
	playerByID := indexPlayers(players)
	mode := "consensus"
	if projectionCount > 0 {
		mode = "league-scored projections"
	}
	return players, playerByID, mode, projectionCount, nil
}

type leaguePlayerValue struct {
	points  float64
	adp     float64
	byeWeek int
}

func staticPlayers(source []draft.Player, preferences map[string]string) ([]draft.Player, map[string]draft.Player, string, int, error) {
	players := append([]draft.Player(nil), source...)
	for index := range players {
		players[index].Preference = preferences[players[index].ID]
		if players[index].Confidence == "" {
			players[index].Confidence = "demo"
		}
	}
	return players, indexPlayers(players), "demo", 0, nil
}

func indexPlayers(players []draft.Player) map[string]draft.Player {
	indexed := make(map[string]draft.Player, len(players))
	for _, player := range players {
		indexed[player.ID] = player
	}
	return indexed
}

func applyReplacementValues(players []draft.Player, rules league.Rules) {
	byPosition := make(map[string][]int)
	projected := 0
	for index, player := range players {
		byPosition[player.Position] = append(byPosition[player.Position], index)
		if player.ProjectedPoints != 0 {
			projected++
		}
	}
	if projected == 0 {
		for index := range players {
			if players[index].Tier == 0 {
				players[index].Tier = 1 + (players[index].PositionRank-1)/5
			}
		}
		return
	}
	for position := range byPosition {
		sort.SliceStable(byPosition[position], func(left, right int) bool {
			return players[byPosition[position][left]].ProjectedPoints > players[byPosition[position][right]].ProjectedPoints
		})
	}
	demand := make(map[string]int)
	flexSlots := make([]league.RosterSlot, 0)
	for _, slot := range rules.RosterSlots {
		if !slot.IsStarting {
			continue
		}
		if len(slot.Positions) == 1 {
			demand[slot.Positions[0]] += slot.Count * rules.TeamCount
		} else {
			for range slot.Count * rules.TeamCount {
				flexSlots = append(flexSlots, slot)
			}
		}
	}
	for _, slot := range flexSlots {
		bestPosition, bestPoints := "", -1.0
		for _, position := range slot.Positions {
			indices := byPosition[position]
			if demand[position] >= len(indices) {
				continue
			}
			points := players[indices[demand[position]]].ProjectedPoints
			if bestPosition == "" || points > bestPoints {
				bestPosition, bestPoints = position, points
			}
		}
		if bestPosition != "" {
			demand[bestPosition]++
		}
	}
	for position, indices := range byPosition {
		replacementIndex := demand[position]
		if replacementIndex >= len(indices) {
			replacementIndex = len(indices) - 1
		}
		baseline := players[indices[replacementIndex]].ProjectedPoints
		previous, tier := 0.0, 1
		for order, index := range indices {
			players[index].ValueOverReplacement = players[index].ProjectedPoints - baseline
			if order > 0 {
				gap := previous - players[index].ValueOverReplacement
				if gap >= 4 || (previous > 0 && gap/previous >= 0.18) {
					tier++
				}
			}
			players[index].Tier = tier
			previous = players[index].ValueOverReplacement
		}
	}
	applyAuctionValues(players, rules)
}

func applyAuctionValues(players []draft.Player, rules league.Rules) {
	if rules.AuctionBudget <= 0 {
		return
	}
	totalRosterSlots := 0
	for _, slot := range rules.RosterSlots {
		totalRosterSlots += slot.Count
	}
	positiveVOR := 0.0
	for _, player := range players {
		if player.ValueOverReplacement > 0 {
			positiveVOR += player.ValueOverReplacement
		}
	}
	minimumBid := rules.AuctionMinimumBid
	if minimumBid <= 0 {
		minimumBid = 1
	}
	discretionary := rules.AuctionBudget*float64(rules.TeamCount) - float64(totalRosterSlots*rules.TeamCount)*minimumBid
	if positiveVOR <= 0 || discretionary <= 0 {
		return
	}
	for index := range players {
		players[index].AuctionValue = minimumBid
		if players[index].ValueOverReplacement > 0 {
			players[index].AuctionValue += discretionary * players[index].ValueOverReplacement / positiveVOR
		}
	}
}

func nextUserPick(current int, rules league.Rules) int {
	if rules.DraftType == league.DraftTypeAuction {
		return 0
	}
	lastPick := totalDraftPicks(rules)
	if lastPick == 0 {
		lastPick = current + rules.TeamCount*2
	}
	for pick := current + 1; pick <= min(current+rules.TeamCount*2, lastPick); pick++ {
		if pickOwner(pick, rules) == rules.DraftPosition {
			return pick
		}
	}
	return 0
}

func isUserTurn(pick int, rules league.Rules) bool {
	if rules.DraftType == league.DraftTypeAuction {
		return true
	}
	return pickOwner(pick, rules) == rules.DraftPosition
}

func auctionState(rules league.Rules, history []draft.Pick, available []draft.Player, myRosterSize int) (float64, float64, float64) {
	if rules.DraftType != league.DraftTypeAuction {
		return 0, 1, 0
	}
	minimumBid := rules.AuctionMinimumBid
	if minimumBid <= 0 {
		minimumBid = 1
	}
	mySpent, leagueSpent := 0.0, 0.0
	for _, pick := range history {
		leagueSpent += pick.Cost
		if pick.Action == draft.ActionDraft {
			mySpent += pick.Cost
		}
	}
	remainingMarket := -rules.KeeperValueRemoved
	for _, player := range available {
		remainingMarket += player.AuctionValue
	}
	remainingLeagueDollars := rules.AuctionBudget*float64(rules.TeamCount) - rules.KeeperBudgetSpent - leagueSpent
	inflation := 1.0
	if remainingMarket > 0 {
		inflation = remainingLeagueDollars / remainingMarket
	}
	if inflation < 0.1 {
		inflation = 0.1
	}
	totalRosterSlots := 0
	for _, slot := range rules.RosterSlots {
		totalRosterSlots += slot.Count
	}
	remainingSlots := max(0, totalRosterSlots-myRosterSize)
	budgetRemaining := max(0, rules.AuctionBudget-rules.MyKeeperSpend-mySpent)
	maximumBid := 0.0
	if remainingSlots > 0 {
		maximumBid = max(minimumBid, budgetRemaining-float64(remainingSlots-1)*minimumBid)
	}
	return budgetRemaining, inflation, maximumBid
}
