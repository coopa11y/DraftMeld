package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/coopa11y/DraftMeld/backend/internal/domain/league"
)

func (store *DraftEventStore) ListLeagues(ctx context.Context) ([]league.Configuration, error) {
	rows, err := store.database.QueryContext(ctx, `SELECT id FROM leagues ORDER BY name, id`)
	if err != nil {
		return nil, fmt.Errorf("query leagues: %w", err)
	}
	ids := make([]string, 0)
	for rows.Next() {
		var id string
		if err = rows.Scan(&id); err != nil {
			_ = rows.Close()
			return nil, fmt.Errorf("scan league ID: %w", err)
		}
		ids = append(ids, id)
	}
	if err = rows.Close(); err != nil {
		return nil, fmt.Errorf("close league rows: %w", err)
	}

	leagues := make([]league.Configuration, 0, len(ids))
	for _, id := range ids {
		configuration, found, getErr := store.GetLeague(ctx, id)
		if getErr != nil {
			return nil, getErr
		}
		if found {
			leagues = append(leagues, configuration)
		}
	}
	return leagues, nil
}

func (store *DraftEventStore) GetLeague(ctx context.Context, id string) (league.Configuration, bool, error) {
	var configuration league.Configuration
	var draftType string
	var scoringJSON, recommendationJSON string
	err := store.database.QueryRowContext(ctx, `
SELECT id, name, team_count, draft_position, draft_type, scoring_rules, recommendation_policy
FROM leagues WHERE id = ?`, id).Scan(
		&configuration.ID, &configuration.Rules.Name, &configuration.Rules.TeamCount,
		&configuration.Rules.DraftPosition, &draftType, &scoringJSON, &recommendationJSON,
	)
	if err == sql.ErrNoRows {
		return league.Configuration{}, false, nil
	}
	if err != nil {
		return league.Configuration{}, false, fmt.Errorf("query league: %w", err)
	}
	configuration.Rules.DraftType = league.DraftType(draftType)
	if err = json.Unmarshal([]byte(scoringJSON), &configuration.Rules.ScoringRules); err != nil {
		return league.Configuration{}, false, fmt.Errorf("decode league scoring rules: %w", err)
	}
	if err = json.Unmarshal([]byte(recommendationJSON), &configuration.Recommendation); err != nil {
		return league.Configuration{}, false, fmt.Errorf("decode recommendation policy: %w", err)
	}

	rows, err := store.database.QueryContext(ctx, `
SELECT name, slot_count, positions, is_starting
FROM league_roster_slots WHERE league_id = ? ORDER BY slot_order`, id)
	if err != nil {
		return league.Configuration{}, false, fmt.Errorf("query league roster slots: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var slot league.RosterSlot
		var positionsJSON string
		if err = rows.Scan(&slot.Name, &slot.Count, &positionsJSON, &slot.IsStarting); err != nil {
			return league.Configuration{}, false, fmt.Errorf("scan league roster slot: %w", err)
		}
		if err = json.Unmarshal([]byte(positionsJSON), &slot.Positions); err != nil {
			return league.Configuration{}, false, fmt.Errorf("decode roster slot positions: %w", err)
		}
		configuration.Rules.RosterSlots = append(configuration.Rules.RosterSlots, slot)
	}
	if err = rows.Err(); err != nil {
		return league.Configuration{}, false, fmt.Errorf("iterate league roster slots: %w", err)
	}
	return configuration, true, nil
}

func (store *DraftEventStore) SaveLeague(ctx context.Context, configuration league.Configuration) error {
	if err := configuration.Validate(); err != nil {
		return err
	}
	scoringJSON, err := json.Marshal(configuration.Rules.ScoringRules)
	if err != nil {
		return fmt.Errorf("encode league scoring rules: %w", err)
	}
	recommendationJSON, err := json.Marshal(configuration.Recommendation)
	if err != nil {
		return fmt.Errorf("encode recommendation policy: %w", err)
	}

	transaction, err := store.database.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin save league: %w", err)
	}
	defer transaction.Rollback()
	now := time.Now().UTC().Format(time.RFC3339Nano)
	_, err = transaction.ExecContext(ctx, `
INSERT INTO leagues (id, name, team_count, draft_position, draft_type, scoring_rules, recommendation_policy, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(id) DO UPDATE SET
  name = excluded.name,
  team_count = excluded.team_count,
  draft_position = excluded.draft_position,
  draft_type = excluded.draft_type,
  scoring_rules = excluded.scoring_rules,
  recommendation_policy = excluded.recommendation_policy,
  updated_at = excluded.updated_at`,
		configuration.ID, configuration.Rules.Name, configuration.Rules.TeamCount,
		configuration.Rules.DraftPosition, configuration.Rules.DraftType, string(scoringJSON),
		string(recommendationJSON), now, now,
	)
	if err != nil {
		return fmt.Errorf("save league: %w", err)
	}
	if _, err = transaction.ExecContext(ctx, `DELETE FROM league_roster_slots WHERE league_id = ?`, configuration.ID); err != nil {
		return fmt.Errorf("replace league roster slots: %w", err)
	}
	for index, slot := range configuration.Rules.RosterSlots {
		positionsJSON, marshalErr := json.Marshal(slot.Positions)
		if marshalErr != nil {
			return fmt.Errorf("encode roster slot positions: %w", marshalErr)
		}
		_, err = transaction.ExecContext(ctx, `
INSERT INTO league_roster_slots (league_id, slot_order, name, slot_count, positions, is_starting)
VALUES (?, ?, ?, ?, ?, ?)`, configuration.ID, index, slot.Name, slot.Count, string(positionsJSON), slot.IsStarting)
		if err != nil {
			return fmt.Errorf("save league roster slot: %w", err)
		}
	}
	if err = transaction.Commit(); err != nil {
		return fmt.Errorf("commit league: %w", err)
	}
	return nil
}

func (store *DraftEventStore) DeleteLeague(ctx context.Context, id string) (bool, error) {
	transaction, err := store.database.BeginTx(ctx, nil)
	if err != nil {
		return false, fmt.Errorf("begin delete league: %w", err)
	}
	defer transaction.Rollback()
	if _, err = transaction.ExecContext(ctx, `DELETE FROM draft_events WHERE league_id = ?`, id); err != nil {
		return false, fmt.Errorf("delete league draft events: %w", err)
	}
	if _, err = transaction.ExecContext(ctx, `DELETE FROM league_roster_slots WHERE league_id = ?`, id); err != nil {
		return false, fmt.Errorf("delete league roster slots: %w", err)
	}
	result, err := transaction.ExecContext(ctx, `DELETE FROM leagues WHERE id = ?`, id)
	if err != nil {
		return false, fmt.Errorf("delete league: %w", err)
	}
	deleted, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("read deleted league count: %w", err)
	}
	if err = transaction.Commit(); err != nil {
		return false, fmt.Errorf("commit delete league: %w", err)
	}
	return deleted > 0, nil
}
