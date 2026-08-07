package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/coopa11y/DraftMeld/backend/internal/domain/player"
)

func (store *DraftEventStore) ResolvePlayer(ctx context.Context, candidate player.Candidate, proposedID string) (player.Player, error) {
	tx, err := store.database.BeginTx(ctx, nil)
	if err != nil {
		return player.Player{}, fmt.Errorf("begin player resolution: %w", err)
	}
	defer tx.Rollback()
	playerID := ""
	if candidate.Provider != "" && candidate.ProviderID != "" {
		err = tx.QueryRowContext(ctx, `SELECT player_id FROM player_provider_ids WHERE provider = ? AND provider_player_id = ?`, candidate.Provider, candidate.ProviderID).Scan(&playerID)
		if err != nil && err != sql.ErrNoRows {
			return player.Player{}, fmt.Errorf("resolve provider player ID: %w", err)
		}
	}
	if playerID == "" && candidate.LegacyKey != "" {
		legacyKey := candidate.LegacyKey
		for range 16 {
			var canonicalKey string
			err = tx.QueryRowContext(ctx, `SELECT canonical_key FROM identity_aliases WHERE alias_key = ?`, legacyKey).Scan(&canonicalKey)
			if err == sql.ErrNoRows || canonicalKey == "" || canonicalKey == legacyKey {
				break
			}
			if err != nil {
				return player.Player{}, fmt.Errorf("resolve legacy player alias: %w", err)
			}
			playerID, legacyKey = canonicalKey, canonicalKey
		}
		if playerID == "" {
			var isCanonical bool
			if err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM identity_aliases WHERE canonical_key = ?)`, candidate.LegacyKey).Scan(&isCanonical); err != nil {
				return player.Player{}, fmt.Errorf("check legacy canonical player: %w", err)
			}
			if isCanonical {
				playerID = candidate.LegacyKey
			}
		}
	}
	if playerID == "" {
		err = tx.QueryRowContext(ctx, `SELECT player_id FROM player_identity_keys WHERE identity_key = ?`, candidate.IdentityKey).Scan(&playerID)
		if err != nil && err != sql.ErrNoRows {
			return player.Player{}, fmt.Errorf("resolve player identity: %w", err)
		}
	}
	nowTime := time.Now().UTC()
	now := nowTime.Format(time.RFC3339Nano)
	team := normalizedObservedTeam(candidate.Team)
	if playerID == "" {
		playerID = proposedID
		if _, err = tx.ExecContext(ctx, `INSERT INTO canonical_players (id, name, position, nfl_team, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)`, playerID, candidate.Name, candidate.Position, team, now, now); err != nil {
			return player.Player{}, fmt.Errorf("create canonical player: %w", err)
		}
	} else if _, err = tx.ExecContext(ctx, `INSERT INTO canonical_players (id, name, position, nfl_team, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?) ON CONFLICT(id) DO NOTHING`, playerID, candidate.Name, candidate.Position, team, now, now); err != nil {
		return player.Player{}, fmt.Errorf("preserve legacy canonical player: %w", err)
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO player_identity_keys (identity_key, player_id, created_at) VALUES (?, ?, ?) ON CONFLICT(identity_key) DO UPDATE SET player_id=excluded.player_id`, candidate.IdentityKey, playerID, now); err != nil {
		return player.Player{}, fmt.Errorf("save player identity key: %w", err)
	}
	if candidate.LegacyKey != "" && candidate.LegacyKey != playerID {
		if _, err = tx.ExecContext(ctx, `INSERT INTO identity_aliases (alias_key, canonical_key, created_at) VALUES (?, ?, ?) ON CONFLICT(alias_key) DO UPDATE SET canonical_key=excluded.canonical_key, created_at=excluded.created_at`, candidate.LegacyKey, playerID, now); err != nil {
			return player.Player{}, fmt.Errorf("save legacy player alias: %w", err)
		}
	}
	if candidate.Provider != "" && candidate.ProviderID != "" {
		if _, err = tx.ExecContext(ctx, `INSERT INTO player_provider_ids (provider, provider_player_id, player_id, created_at) VALUES (?, ?, ?, ?) ON CONFLICT(provider, provider_player_id) DO UPDATE SET player_id=excluded.player_id`, candidate.Provider, candidate.ProviderID, playerID, now); err != nil {
			return player.Player{}, fmt.Errorf("save provider player ID: %w", err)
		}
	}
	team, err = recordTeamObservation(ctx, tx, playerID, candidate, nowTime)
	if err != nil {
		return player.Player{}, err
	}
	if _, err = tx.ExecContext(ctx, `UPDATE canonical_players SET name = ?, position = ?, nfl_team = CASE WHEN ? <> '' THEN ? ELSE nfl_team END, updated_at = ? WHERE id = ?`, candidate.Name, candidate.Position, team, team, now, playerID); err != nil {
		return player.Player{}, fmt.Errorf("update canonical player: %w", err)
	}
	var resolved player.Player
	if err = tx.QueryRowContext(ctx, `SELECT id, name, position, nfl_team FROM canonical_players WHERE id = ?`, playerID).Scan(&resolved.ID, &resolved.Name, &resolved.Position, &resolved.Team); err != nil {
		return player.Player{}, fmt.Errorf("load canonical player: %w", err)
	}
	if err = tx.Commit(); err != nil {
		return player.Player{}, fmt.Errorf("commit player resolution: %w", err)
	}
	return resolved, nil
}

func recordTeamObservation(ctx context.Context, tx *sql.Tx, playerID string, candidate player.Candidate, fallbackTime time.Time) (string, error) {
	team := normalizedObservedTeam(candidate.Team)
	if team == "" {
		return "", nil
	}
	observedAt := candidate.ObservedAt.UTC()
	if observedAt.IsZero() {
		observedAt = fallbackTime
	}
	source := strings.TrimSpace(candidate.Provider)
	if source == "" {
		source = candidate.IdentityKey
	}
	confidence := 1
	if candidate.ProviderID != "" {
		confidence = 2
	}
	if strings.EqualFold(source, "sleeper") && candidate.ProviderID != "" {
		confidence = 3
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO player_team_observations (player_id, source, nfl_team, confidence, observed_at)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(player_id, source) DO UPDATE SET nfl_team=excluded.nfl_team, confidence=excluded.confidence, observed_at=excluded.observed_at
		WHERE excluded.observed_at > player_team_observations.observed_at
		   OR (excluded.observed_at = player_team_observations.observed_at AND excluded.confidence >= player_team_observations.confidence)`,
		playerID, source, team, confidence, observedAt.UnixNano()); err != nil {
		return "", fmt.Errorf("save player team observation: %w", err)
	}
	if err := tx.QueryRowContext(ctx, `SELECT nfl_team FROM player_team_observations WHERE player_id = ? ORDER BY observed_at DESC, confidence DESC, source LIMIT 1`, playerID).Scan(&team); err != nil {
		return "", fmt.Errorf("resolve current player team: %w", err)
	}
	return team, nil
}

func normalizedObservedTeam(team string) string {
	team = strings.ToUpper(strings.TrimSpace(team))
	switch team {
	case "", "-", "N/A", "NA", "FA", "FREE AGENT":
		return ""
	default:
		return team
	}
}

func (store *DraftEventStore) CanonicalPlayers(ctx context.Context, ids []string) (map[string]player.Player, error) {
	players := make(map[string]player.Player, len(ids))
	if len(ids) == 0 {
		return players, nil
	}
	wanted := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		wanted[id] = struct{}{}
	}
	rows, err := store.database.QueryContext(ctx, `SELECT id, name, position, nfl_team FROM canonical_players WHERE merged_into IS NULL`)
	if err != nil {
		return nil, fmt.Errorf("load canonical players: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var item player.Player
		if err = rows.Scan(&item.ID, &item.Name, &item.Position, &item.Team); err != nil {
			return nil, fmt.Errorf("scan canonical player: %w", err)
		}
		if _, exists := wanted[item.ID]; exists {
			players[item.ID] = item
		}
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate canonical players: %w", err)
	}
	return players, nil
}

func (store *DraftEventStore) PlayerDirectoryStatus(ctx context.Context) (player.DirectoryStatus, error) {
	var status player.DirectoryStatus
	queries := []struct {
		query  string
		target *int
	}{
		{`SELECT COUNT(*) FROM canonical_players WHERE merged_into IS NULL`, &status.PlayerCount},
		{`SELECT COUNT(*) FROM player_identity_keys`, &status.IdentityCount},
		{`SELECT COUNT(*) FROM player_provider_ids`, &status.ProviderIDCount},
	}
	for _, query := range queries {
		if err := store.database.QueryRowContext(ctx, query.query).Scan(query.target); err != nil {
			return player.DirectoryStatus{}, fmt.Errorf("load player directory status: %w", err)
		}
	}
	return status, nil
}
