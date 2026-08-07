package sqlite

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/coopa11y/DraftMeld/backend/internal/domain/projection"
)

func (store *DraftEventStore) ReplaceProjections(ctx context.Context, source projection.SourceStatus, records []projection.Record) error {
	tx, err := store.database.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin projection import: %w", err)
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, `INSERT INTO projection_sources (id, name, imported_at, record_count)
VALUES (?, ?, ?, ?) ON CONFLICT(id) DO UPDATE SET name=excluded.name, imported_at=excluded.imported_at, record_count=excluded.record_count`, source.ID, source.Name, source.ImportedAt.Format(time.RFC3339Nano), len(records)); err != nil {
		return fmt.Errorf("save projection source: %w", err)
	}
	if _, err = tx.ExecContext(ctx, "DELETE FROM player_projections WHERE source_id = ?", source.ID); err != nil {
		return fmt.Errorf("clear projections: %w", err)
	}
	statement, err := tx.PrepareContext(ctx, `INSERT INTO player_projections (source_id, player_key, player_name, position, nfl_team, bye_week, adp, stats_json) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return fmt.Errorf("prepare projections: %w", err)
	}
	defer statement.Close()
	for _, record := range records {
		stats, marshalErr := json.Marshal(record.Stats)
		if marshalErr != nil {
			return fmt.Errorf("encode projection stats: %w", marshalErr)
		}
		if _, err = statement.ExecContext(ctx, source.ID, record.PlayerKey, record.Name, record.Position, record.Team, record.ByeWeek, record.ADP, string(stats)); err != nil {
			return fmt.Errorf("save projection: %w", err)
		}
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit projections: %w", err)
	}
	return nil
}

func (store *DraftEventStore) ProjectionRecords(ctx context.Context) ([]projection.Record, error) {
	rows, err := store.database.QueryContext(ctx, `SELECT source_id, player_key, player_name, position, nfl_team, bye_week, adp, stats_json FROM player_projections ORDER BY source_id, player_key`)
	if err != nil {
		return nil, fmt.Errorf("query projections: %w", err)
	}
	defer rows.Close()
	records := make([]projection.Record, 0)
	for rows.Next() {
		var record projection.Record
		var stats string
		if err = rows.Scan(&record.SourceID, &record.PlayerKey, &record.Name, &record.Position, &record.Team, &record.ByeWeek, &record.ADP, &stats); err != nil {
			return nil, fmt.Errorf("scan projection: %w", err)
		}
		if err = json.Unmarshal([]byte(stats), &record.Stats); err != nil {
			return nil, fmt.Errorf("decode projection stats: %w", err)
		}
		records = append(records, record)
	}
	return records, rows.Err()
}

func (store *DraftEventStore) ProjectionStatuses(ctx context.Context) ([]projection.SourceStatus, error) {
	rows, err := store.database.QueryContext(ctx, `SELECT id, name, imported_at, record_count FROM projection_sources ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("query projection sources: %w", err)
	}
	defer rows.Close()
	statuses := make([]projection.SourceStatus, 0)
	for rows.Next() {
		var status projection.SourceStatus
		var imported string
		if err = rows.Scan(&status.ID, &status.Name, &imported, &status.RecordCount); err != nil {
			return nil, fmt.Errorf("scan projection source: %w", err)
		}
		status.ImportedAt, err = time.Parse(time.RFC3339Nano, imported)
		if err != nil {
			return nil, fmt.Errorf("parse projection import time: %w", err)
		}
		statuses = append(statuses, status)
	}
	return statuses, rows.Err()
}
