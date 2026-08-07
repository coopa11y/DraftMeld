package application

import (
	"fmt"
	"math"
	"sort"

	"github.com/coopa11y/DraftMeld/backend/internal/domain/draft"
	"github.com/coopa11y/DraftMeld/backend/internal/domain/league"
)

type recommendationContext struct {
	NextUserPick int
	RecentPicks  []draft.Pick
}

func recommend(available []draft.Player, myTeam []draft.Player, rules league.Rules, policy RecommendationPolicy, contexts ...recommendationContext) []draft.Recommendation {
	neededPositions := openStartingPositions(rules.RosterSlots, myTeam)
	positionRanks := make(map[string][]int)
	for _, player := range available {
		positionRanks[player.Position] = append(positionRanks[player.Position], player.OverallRank)
	}
	context := recommendationContext{}
	if len(contexts) > 0 {
		context = contexts[0]
	}
	recentPositions := make(map[string]int)
	start := max(0, len(context.RecentPicks)-5)
	for _, pick := range context.RecentPicks[start:] {
		recentPositions[pick.Player.Position]++
	}

	recommendations := make([]draft.Recommendation, 0, len(available))
	for _, player := range available {
		score := policy.BaseScore - float64(player.OverallRank) + player.ValueOverReplacement*0.35
		reasons := make([]string, 0, 6)
		if player.Preference == "target" {
			score += 30
			reasons = append(reasons, "Marked as one of your targets")
		} else if player.Preference == "avoid" {
			score -= 500
			reasons = append(reasons, "On your avoid list")
		}
		if player.ValueOverReplacement > 0 {
			reasons = append(reasons, fmt.Sprintf("Adds %.1f projected points over replacement", player.ValueOverReplacement))
		}
		if neededPositions[player.Position] {
			score += policy.StartingNeedBonus
			reasons = append(reasons, "Fills an open starting roster need")
		}
		value := player.ADP - float64(player.OverallRank)
		if value >= policy.ADPValueThreshold {
			score += value
			reasons = append(reasons, fmt.Sprintf("Ranks %.0f spots above draft-room ADP", value))
		}
		if context.NextUserPick > 0 && player.ADP > 0 && player.ADP < float64(context.NextUserPick)-1 {
			score += math.Min(15, float64(context.NextUserPick)-player.ADP)
			reasons = append(reasons, fmt.Sprintf("Unlikely to remain available at pick %d", context.NextUserPick))
		}
		if recentPositions[player.Position] >= 2 {
			score += 6
			reasons = append(reasons, fmt.Sprintf("Responds to a recent %s run", player.Position))
		}
		ranks := positionRanks[player.Position]
		if len(ranks) > 1 && ranks[0] == player.OverallRank && ranks[1]-ranks[0] >= policy.ScarcityDropOff {
			score += policy.ScarcityBonus
			reasons = append(reasons, "Top option before a positional drop-off")
		}
		if rules.DraftType == league.DraftTypeAuction && player.AuctionValue > 0 {
			reasons = append(reasons, fmt.Sprintf("Baseline auction value is $%.0f before inflation", player.AuctionValue))
		}
		if len(reasons) == 0 {
			reasons = append(reasons, "Best available league-adjusted value")
		}
		recommendations = append(recommendations, draft.Recommendation{Player: player, Score: score, Reasons: reasons})
	}
	sort.SliceStable(recommendations, func(left, right int) bool { return recommendations[left].Score > recommendations[right].Score })
	if len(recommendations) > policy.RecommendationLimit {
		recommendations = recommendations[:policy.RecommendationLimit]
	}
	return recommendations
}

type startingSlot struct{ positions map[string]bool }

func openStartingPositions(slots []league.RosterSlot, myTeam []draft.Player) map[string]bool {
	openSlots := make([]startingSlot, 0)
	for _, slot := range slots {
		if !slot.IsStarting {
			continue
		}
		for range slot.Count {
			positions := make(map[string]bool, len(slot.Positions))
			for _, position := range slot.Positions {
				positions[position] = true
			}
			openSlots = append(openSlots, startingSlot{positions: positions})
		}
	}
	sort.SliceStable(openSlots, func(left, right int) bool { return len(openSlots[left].positions) < len(openSlots[right].positions) })
	for _, player := range myTeam {
		for index, slot := range openSlots {
			if slot.positions[player.Position] {
				openSlots = append(openSlots[:index], openSlots[index+1:]...)
				break
			}
		}
	}
	needed := make(map[string]bool)
	for _, slot := range openSlots {
		for position := range slot.positions {
			needed[position] = true
		}
	}
	return needed
}
