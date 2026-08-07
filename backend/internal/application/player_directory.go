package application

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/coopa11y/DraftMeld/backend/internal/domain/draft"
	"github.com/coopa11y/DraftMeld/backend/internal/domain/player"
	"github.com/coopa11y/DraftMeld/backend/internal/domain/projection"
	"github.com/coopa11y/DraftMeld/backend/internal/domain/ranking"
)

type PlayerDirectoryRepository interface {
	ResolvePlayer(context.Context, player.Candidate, string) (player.Player, error)
	PlayerDirectoryStatus(context.Context) (player.DirectoryStatus, error)
}

func newCanonicalPlayerID() (string, error) {
	bytes := make([]byte, 12)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("create canonical player ID: %w", err)
	}
	return "player-" + hex.EncodeToString(bytes), nil
}

func resolveRankingPlayers(ctx context.Context, repository any, records []ranking.Record) ([]ranking.Record, error) {
	directory, ok := repository.(PlayerDirectoryRepository)
	if !ok {
		return records, nil
	}
	resolvedRecords := make([]ranking.Record, 0, len(records))
	seen := make(map[string]int, len(records))
	for index := range records {
		record := canonicalizeRankingRecord(records[index])
		id, err := newCanonicalPlayerID()
		if err != nil {
			return nil, err
		}
		resolved, err := directory.ResolvePlayer(ctx, player.Candidate{
			IdentityKey: canonicalRankingKey(record.Name, record.Position, record.Team), LegacyKey: record.PlayerKey,
			Name: record.Name, Position: record.Position, Team: record.Team,
			Provider: record.SourceID, ProviderID: record.ProviderID,
		}, id)
		if err != nil {
			return nil, err
		}
		record.PlayerKey = resolved.ID
		record.Team = resolved.Team
		if existing, exists := seen[record.PlayerKey]; exists {
			if record.Rank < resolvedRecords[existing].Rank {
				resolvedRecords[existing] = record
			}
			continue
		}
		seen[record.PlayerKey] = len(resolvedRecords)
		resolvedRecords = append(resolvedRecords, record)
	}
	return resolvedRecords, nil
}

func resolveProjectionPlayers(ctx context.Context, repository any, records []projection.Record) ([]projection.Record, error) {
	directory, ok := repository.(PlayerDirectoryRepository)
	if !ok {
		return records, nil
	}
	resolvedRecords := make([]projection.Record, 0, len(records))
	seen := make(map[string]bool, len(records))
	for index := range records {
		record := records[index]
		id, err := newCanonicalPlayerID()
		if err != nil {
			return nil, err
		}
		resolved, err := directory.ResolvePlayer(ctx, player.Candidate{
			IdentityKey: canonicalRankingKey(record.Name, record.Position, record.Team), LegacyKey: record.PlayerKey,
			Name: record.Name, Position: record.Position, Team: record.Team,
			Provider: record.SourceID, ProviderID: record.ProviderID,
		}, id)
		if err != nil {
			return nil, err
		}
		record.PlayerKey = resolved.ID
		record.Team = resolved.Team
		if seen[record.PlayerKey] {
			continue
		}
		seen[record.PlayerKey] = true
		resolvedRecords = append(resolvedRecords, record)
	}
	return resolvedRecords, nil
}

func (service *RankingService) PlayerDirectoryStatus(ctx context.Context) (player.DirectoryStatus, error) {
	directory, ok := service.repository.(PlayerDirectoryRepository)
	if !ok {
		return player.DirectoryStatus{}, nil
	}
	return directory.PlayerDirectoryStatus(ctx)
}

type identityAliasReader interface {
	IdentityAliases(context.Context) (map[string]string, error)
}

func resolveDraftEventAliases(ctx context.Context, repository any, events []draft.Event) ([]draft.Event, error) {
	reader, ok := repository.(identityAliasReader)
	if !ok {
		return events, nil
	}
	aliases, err := reader.IdentityAliases(ctx)
	if err != nil {
		return nil, err
	}
	resolved := append([]draft.Event(nil), events...)
	for index := range resolved {
		resolved[index].PlayerID = resolveIdentityAlias(resolved[index].PlayerID, aliases)
	}
	return resolved, nil
}

func resolvePlayerPreferences(ctx context.Context, repository any, preferences map[string]string) (map[string]string, error) {
	reader, ok := repository.(identityAliasReader)
	if !ok {
		return preferences, nil
	}
	aliases, err := reader.IdentityAliases(ctx)
	if err != nil {
		return nil, err
	}
	resolved := make(map[string]string, len(preferences))
	for playerID, preference := range preferences {
		resolved[resolveIdentityAlias(playerID, aliases)] = preference
	}
	return resolved, nil
}
