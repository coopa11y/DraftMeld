package application

import (
	"context"
	"sort"

	"github.com/coopa11y/DraftMeld/backend/internal/domain/league"
	"github.com/coopa11y/DraftMeld/backend/internal/domain/ranking"
)

const (
	watchlistLimit       = 5
	watchlistRankGap     = 10
	watchlistMissingRank = 50
)

type watchlistCandidate struct {
	player   ranking.WatchlistPlayer
	bestRank int
	maxGap   int
}

func (service *RankingService) Watchlist(ctx context.Context, preferences map[string]league.RankingSourcePreference, methods ...string) ([]ranking.WatchlistPlayer, error) {
	records, err := service.repository.RankingRecords(ctx)
	if err != nil {
		return nil, err
	}
	allDefinitions, err := service.definitions(ctx)
	if err != nil {
		return nil, err
	}
	consensus, err := service.Consensus(ctx, preferences, methods...)
	if err != nil {
		return nil, err
	}
	consensusRanks := make(map[string]int, len(consensus))
	for _, player := range consensus {
		consensusRanks[player.PlayerKey] = player.Rank
	}
	definitions := make(map[string]ranking.SourceDefinition, len(allDefinitions))
	for _, definition := range allDefinitions {
		definitions[definition.ID] = definition
	}
	eligible := make(map[string]bool)
	metadata := make(map[string]ranking.Record)
	for _, record := range records {
		record = canonicalizeRankingRecord(record)
		if definitions[record.SourceID].Role == "ranking" {
			eligible[record.PlayerKey] = true
			metadata[record.PlayerKey] = record
		}
	}
	candidates := make(map[string]*watchlistCandidate)
	for _, record := range records {
		record = canonicalizeRankingRecord(record)
		definition, exists := definitions[record.SourceID]
		if !exists || !eligible[record.PlayerKey] || effectiveSourcePreference(definition, preferences).Enabled {
			continue
		}
		consensusRank, ranked := consensusRanks[record.PlayerKey]
		spotsHigher := consensusRank - record.Rank
		if (ranked && spotsHigher < watchlistRankGap) || (!ranked && record.Rank > watchlistMissingRank) {
			continue
		}
		candidate := candidates[record.PlayerKey]
		if candidate == nil {
			player := metadata[record.PlayerKey]
			candidate = &watchlistCandidate{player: ranking.WatchlistPlayer{
				PlayerKey: record.PlayerKey, Name: player.Name, Position: player.Position, Team: player.Team,
				Signals: make([]ranking.WatchlistSignal, 0, 1),
			}}
			if ranked {
				rank := consensusRank
				candidate.player.ConsensusRank = &rank
			}
			candidates[record.PlayerKey] = candidate
		}
		candidate.player.Signals = append(candidate.player.Signals, ranking.WatchlistSignal{
			SourceID: record.SourceID, SourceName: definition.Name, SourceRank: record.Rank, SpotsHigher: spotsHigher,
		})
		if candidate.bestRank == 0 || record.Rank < candidate.bestRank {
			candidate.bestRank = record.Rank
		}
		if ranked {
			candidate.maxGap = max(candidate.maxGap, spotsHigher)
		}
	}
	ordered := make([]watchlistCandidate, 0, len(candidates))
	for _, candidate := range candidates {
		sort.Slice(candidate.player.Signals, func(left, right int) bool {
			return candidate.player.Signals[left].SourceRank < candidate.player.Signals[right].SourceRank
		})
		ordered = append(ordered, *candidate)
	}
	sort.Slice(ordered, func(left, right int) bool {
		if ordered[left].bestRank != ordered[right].bestRank {
			return ordered[left].bestRank < ordered[right].bestRank
		}
		if ordered[left].maxGap != ordered[right].maxGap {
			return ordered[left].maxGap > ordered[right].maxGap
		}
		return ordered[left].player.Name < ordered[right].player.Name
	})
	if len(ordered) > watchlistLimit {
		ordered = ordered[:watchlistLimit]
	}
	result := make([]ranking.WatchlistPlayer, len(ordered))
	for index, candidate := range ordered {
		result[index] = candidate.player
	}
	return result, nil
}
