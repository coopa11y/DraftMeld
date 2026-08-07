package application

import (
	"fmt"

	"github.com/coopa11y/DraftMeld/backend/internal/domain/league"
	"github.com/coopa11y/DraftMeld/backend/internal/domain/ranking"
)

type resolvedRankingRecord struct {
	record      ranking.Record
	isCanonical bool
}

type consensusInputs struct {
	sources     []ranking.Source
	players     []string
	metadata    map[string]ranking.Record
	sourceRanks map[string]map[string]int
}

func validateRankingSourcePreferences(definitions []ranking.SourceDefinition, preferences map[string]league.RankingSourcePreference) error {
	enabledSources := 0
	for _, definition := range definitions {
		preference := effectiveSourcePreference(definition, preferences)
		if preference.Weight <= 0 || preference.Weight > 10 {
			return fmt.Errorf("ranking source %s requires a weight greater than 0 and no more than 10", definition.ID)
		}
		if preference.Enabled {
			enabledSources++
		}
	}
	if enabledSources == 0 {
		return fmt.Errorf("at least one ranking source must be enabled")
	}
	return nil
}

func resolveRankingRecords(records []ranking.Record, aliases map[string]string) []resolvedRankingRecord {
	resolved := make([]resolvedRankingRecord, 0, len(records))
	for _, record := range records {
		record = canonicalizeRankingRecord(record)
		originalPlayerKey := record.PlayerKey
		record.PlayerKey = resolveIdentityAlias(record.PlayerKey, aliases)
		resolved = append(resolved, resolvedRankingRecord{record: record, isCanonical: originalPlayerKey == record.PlayerKey})
	}
	return resolved
}

func buildConsensusInputs(definitions []ranking.SourceDefinition, records []resolvedRankingRecord, preferences map[string]league.RankingSourcePreference) consensusInputs {
	eligible := eligiblePlayerKeys(records)
	definitionsByID := make(map[string]ranking.SourceDefinition, len(definitions))
	for _, definition := range definitions {
		definitionsByID[definition.ID] = definition
	}

	sources := make(map[string]ranking.Source)
	metadata := make(map[string]ranking.Record)
	metadataPriority := make(map[string]int)
	sourceRanks := make(map[string]map[string]int)
	for _, resolved := range records {
		record := resolved.record
		if _, exists := eligible[record.PlayerKey]; !exists {
			continue
		}
		definition, exists := definitionsByID[record.SourceID]
		if !exists {
			continue
		}
		preference := effectiveSourcePreference(definition, preferences)
		if !preference.Enabled {
			continue
		}
		addSourceRank(sources, record, definition, preference)
		addPlayerMetadata(metadata, metadataPriority, record, resolved.isCanonical)
		addPlayerSourceRank(sourceRanks, record)
	}

	weighted := make([]ranking.Source, 0, len(sources))
	for _, source := range sources {
		weighted = append(weighted, source)
	}
	players := make([]string, 0, len(eligible))
	for playerID := range eligible {
		players = append(players, playerID)
	}
	return consensusInputs{sources: weighted, players: players, metadata: metadata, sourceRanks: sourceRanks}
}

func eligiblePlayerKeys(records []resolvedRankingRecord) map[string]struct{} {
	// Current redraft lists define the eligible pool. Contextual and dynasty
	// feeds enrich those players without introducing retired players on their own.
	eligible := make(map[string]struct{})
	for _, resolved := range records {
		record := resolved.record
		if record.SourceID == "redraft-ecr" || record.SourceID == "espn-ppr-pdf" {
			eligible[record.PlayerKey] = struct{}{}
		}
	}
	return eligible
}

func addSourceRank(sources map[string]ranking.Source, record ranking.Record, definition ranking.SourceDefinition, preference league.RankingSourcePreference) {
	source := sources[record.SourceID]
	source.ID, source.Role, source.Weight = record.SourceID, definition.Role, preference.Weight
	if source.Ranks == nil {
		source.Ranks = make(map[string]int)
	}
	if currentRank, ranked := source.Ranks[record.PlayerKey]; !ranked || record.Rank < currentRank {
		source.Ranks[record.PlayerKey] = record.Rank
	}
	sources[record.SourceID] = source
}

func addPlayerMetadata(metadata map[string]ranking.Record, priorities map[string]int, record ranking.Record, isCanonical bool) {
	priority := 0
	if record.SourceID == "espn-ppr-pdf" || record.SourceID == "redraft-ecr" {
		priority += 2
	}
	if isCanonical {
		priority += 4
	}
	if _, exists := metadata[record.PlayerKey]; !exists || priority > priorities[record.PlayerKey] {
		metadata[record.PlayerKey] = record
		priorities[record.PlayerKey] = priority
	}
}

func addPlayerSourceRank(sourceRanks map[string]map[string]int, record ranking.Record) {
	if sourceRanks[record.PlayerKey] == nil {
		sourceRanks[record.PlayerKey] = make(map[string]int)
	}
	if currentRank, ranked := sourceRanks[record.PlayerKey][record.SourceID]; !ranked || record.Rank < currentRank {
		sourceRanks[record.PlayerKey][record.SourceID] = record.Rank
	}
}

func requestedConsensusMethod(methods []string) string {
	if len(methods) > 0 && methods[0] != "" {
		return methods[0]
	}
	return ranking.MethodWeightedAverage
}

func playerRankings(entries []ranking.Entry, inputs consensusInputs, method string) []ranking.PlayerRanking {
	result := make([]ranking.PlayerRanking, 0, len(entries))
	for index, entry := range entries {
		player := inputs.metadata[entry.PlayerID]
		result = append(result, ranking.PlayerRanking{
			PlayerKey: entry.PlayerID, Name: player.Name, Position: player.Position, Team: player.Team,
			Rank: index + 1, Score: entry.Score, SourceCount: entry.SourceCount,
			SourceRanks: inputs.sourceRanks[entry.PlayerID], Coverage: entry.Coverage,
			RankRange: entry.RankRange, Confidence: entry.Confidence, Method: method,
		})
	}
	return result
}
