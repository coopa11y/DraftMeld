package application

import (
	"fmt"
	"sort"

	"github.com/coopa11y/DraftMeld/backend/internal/domain/draft"
	"github.com/coopa11y/DraftMeld/backend/internal/domain/league"
)

func recommend(
	available []draft.Player,
	myTeam []draft.Player,
	rules league.Rules,
	policy RecommendationPolicy,
) []draft.Recommendation {
	neededPositions := openStartingPositions(rules.RosterSlots, myTeam)
	positionRanks := make(map[string][]int)
	for _, player := range available {
		positionRanks[player.Position] = append(positionRanks[player.Position], player.OverallRank)
	}

	recommendations := make([]draft.Recommendation, 0, len(available))
	for _, player := range available {
		score := policy.BaseScore - float64(player.OverallRank)
		reasons := make([]string, 0, 3)
		if neededPositions[player.Position] {
			score += policy.StartingNeedBonus
			reasons = append(reasons, "Fills an open starting roster need")
		}
		value := player.ADP - float64(player.OverallRank)
		if value >= policy.ADPValueThreshold {
			score += value
			reasons = append(reasons, fmt.Sprintf("Ranks %.0f spots above draft-room ADP", value))
		}
		ranks := positionRanks[player.Position]
		if len(ranks) > 1 && ranks[0] == player.OverallRank && ranks[1]-ranks[0] >= policy.ScarcityDropOff {
			score += policy.ScarcityBonus
			reasons = append(reasons, "Top option before a positional drop-off")
		}
		if len(reasons) == 0 {
			reasons = append(reasons, "Best available league-adjusted value")
		}
		recommendations = append(recommendations, draft.Recommendation{Player: player, Score: score, Reasons: reasons})
	}
	sort.SliceStable(recommendations, func(left, right int) bool {
		return recommendations[left].Score > recommendations[right].Score
	})
	if len(recommendations) > policy.RecommendationLimit {
		recommendations = recommendations[:policy.RecommendationLimit]
	}
	return recommendations
}

type startingSlot struct {
	positions map[string]bool
}

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
	sort.SliceStable(openSlots, func(left, right int) bool {
		return len(openSlots[left].positions) < len(openSlots[right].positions)
	})
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
