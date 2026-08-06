package ranking

import (
	"errors"
	"sort"
)

type Source struct {
	ID     string
	Weight float64
	Ranks  map[string]int
}

type Entry struct {
	PlayerID    string  `json:"playerId"`
	Score       float64 `json:"score"`
	SourceCount int     `json:"sourceCount"`
}

func WeightedAverage(sources []Source) ([]Entry, error) {
	weightedRanks := make(map[string]float64)
	weights := make(map[string]float64)
	counts := make(map[string]int)

	for _, source := range sources {
		if source.ID == "" || source.Weight <= 0 {
			return nil, errors.New("each ranking source requires an ID and positive weight")
		}
		for playerID, rank := range source.Ranks {
			if playerID == "" || rank < 1 {
				return nil, errors.New("player IDs must be non-empty and ranks must be positive")
			}
			weightedRanks[playerID] += float64(rank) * source.Weight
			weights[playerID] += source.Weight
			counts[playerID]++
		}
	}

	entries := make([]Entry, 0, len(weightedRanks))
	for playerID, total := range weightedRanks {
		entries = append(entries, Entry{
			PlayerID:    playerID,
			Score:       total / weights[playerID],
			SourceCount: counts[playerID],
		})
	}
	sort.Slice(entries, func(left, right int) bool {
		if entries[left].Score == entries[right].Score {
			return entries[left].PlayerID < entries[right].PlayerID
		}
		return entries[left].Score < entries[right].Score
	})
	return entries, nil
}
