package application

import (
	"fmt"
	"math"

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
	adpTotals   map[string]float64
	adpWeights  map[string]float64
	tierTotals  map[string]float64
	tierWeights map[string]float64
	projections map[string]ranking.ProjectionEvidence
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
	definitionsByID := make(map[string]ranking.SourceDefinition, len(definitions))
	for _, definition := range definitions {
		definitionsByID[definition.ID] = definition
	}
	eligible := eligiblePlayerKeys(records, definitionsByID)

	sources := make(map[string]ranking.Source)
	metadata := make(map[string]ranking.Record)
	metadataPriority := make(map[string]int)
	sourceRanks := make(map[string]map[string]int)
	adpTotals, adpWeights := make(map[string]float64), make(map[string]float64)
	tierTotals, tierWeights := make(map[string]float64), make(map[string]float64)
	projections, projectionWeights := make(map[string]ranking.ProjectionEvidence), make(map[string]float64)
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
		addPlayerMetadata(metadata, metadataPriority, record, definition, resolved.isCanonical)
		addPlayerSourceRank(sourceRanks, record)
		if record.ADP > 0 {
			adpTotals[record.PlayerKey] += record.ADP * preference.Weight
			adpWeights[record.PlayerKey] += preference.Weight
		}
		if record.Tier > 0 {
			tierTotals[record.PlayerKey] += float64(record.Tier) * preference.Weight
			tierWeights[record.PlayerKey] += preference.Weight
		}
		if record.SourceProjection > 0 && preference.Weight > projectionWeights[record.PlayerKey] {
			projections[record.PlayerKey] = projectionEvidence(record, definition)
			projectionWeights[record.PlayerKey] = preference.Weight
		}
	}

	weighted := make([]ranking.Source, 0, len(sources))
	for _, source := range sources {
		weighted = append(weighted, source)
	}
	players := make([]string, 0, len(eligible))
	for playerID := range eligible {
		players = append(players, playerID)
	}
	return consensusInputs{sources: weighted, players: players, metadata: metadata, sourceRanks: sourceRanks,
		adpTotals: adpTotals, adpWeights: adpWeights, tierTotals: tierTotals, tierWeights: tierWeights,
		projections: projections}
}

func projectionEvidence(record ranking.Record, definition ranking.SourceDefinition) ranking.ProjectionEvidence {
	return ranking.ProjectionEvidence{
		SourceID: record.SourceID, SourceName: definition.Name, Profile: definition.Profile,
		Games: record.Games, ByeWeek: record.ByeWeek, FloorProjection: record.FloorProjection,
		ConsensusProjection: record.ConsensusProjection, SourceProjection: record.SourceProjection,
		CeilingProjection: record.CeilingProjection, SourceValue: record.SourceValue,
		InjuryRisk: record.InjuryRisk, ScheduleStrength: record.ScheduleStrength,
	}
}

func eligiblePlayerKeys(records []resolvedRankingRecord, definitions map[string]ranking.SourceDefinition) map[string]struct{} {
	// Ordinal ranking lists define the eligible pool. Contextual market and usage
	// feeds enrich those players without introducing retired players on their own.
	eligible := make(map[string]struct{})
	for _, resolved := range records {
		record := resolved.record
		if definitions[record.SourceID].Role == "ranking" {
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

func addPlayerMetadata(metadata map[string]ranking.Record, priorities map[string]int, record ranking.Record, definition ranking.SourceDefinition, isCanonical bool) {
	priority := 0
	if definition.Role == "ranking" {
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
		adp, tier := 0.0, 0
		if inputs.adpWeights[entry.PlayerID] > 0 {
			adp = inputs.adpTotals[entry.PlayerID] / inputs.adpWeights[entry.PlayerID]
		}
		if inputs.tierWeights[entry.PlayerID] > 0 {
			tier = int(math.Round(inputs.tierTotals[entry.PlayerID] / inputs.tierWeights[entry.PlayerID]))
		}
		var projection *ranking.ProjectionEvidence
		if evidence, exists := inputs.projections[entry.PlayerID]; exists {
			projection = &evidence
		}
		result = append(result, ranking.PlayerRanking{
			PlayerKey: entry.PlayerID, Name: player.Name, Position: player.Position, Team: player.Team,
			Rank: index + 1, Score: entry.Score, SourceCount: entry.SourceCount,
			SourceRanks: inputs.sourceRanks[entry.PlayerID], Coverage: entry.Coverage,
			RankRange: entry.RankRange, Confidence: entry.Confidence, Method: method, ADP: adp, Tier: tier,
			Projection: projection,
		})
	}
	return result
}
