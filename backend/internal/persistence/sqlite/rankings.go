package sqlite

import (
	"context"
	"fmt"
	"time"

	"github.com/coopa11y/DraftMeld/backend/internal/domain/ranking"
)

func (store *DraftEventStore) ReplaceRankings(ctx context.Context, source ranking.SourceDefinition, records []ranking.Record, published string, refreshed time.Time) error {
	tx, err := store.database.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin ranking refresh: %w", err)
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, `INSERT INTO ranking_sources (id, refreshed_at, published_at, record_count)
VALUES (?, ?, ?, ?) ON CONFLICT(id) DO UPDATE SET refreshed_at=excluded.refreshed_at, published_at=excluded.published_at, record_count=excluded.record_count`, source.ID, refreshed.Format(time.RFC3339Nano), published, len(records)); err != nil {
		return fmt.Errorf("save ranking source: %w", err)
	}
	if _, err = tx.ExecContext(ctx, "DELETE FROM ranking_entries WHERE source_id = ?", source.ID); err != nil {
		return fmt.Errorf("clear ranking entries: %w", err)
	}
	statement, err := tx.PrepareContext(ctx, `INSERT INTO ranking_entries (source_id, player_key, player_name, position, nfl_team, source_rank) VALUES (?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return fmt.Errorf("prepare ranking entries: %w", err)
	}
	defer statement.Close()
	for _, record := range records {
		if _, err = statement.ExecContext(ctx, source.ID, record.PlayerKey, record.Name, record.Position, record.Team, record.Rank); err != nil {
			return fmt.Errorf("save ranking entry: %w", err)
		}
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit ranking refresh: %w", err)
	}
	return nil
}

func (store *DraftEventStore) RankingRecords(ctx context.Context) ([]ranking.Record, error) {
	rows, err := store.database.QueryContext(ctx, `SELECT source_id, player_key, player_name, position, nfl_team, source_rank FROM ranking_entries ORDER BY source_id, source_rank`)
	if err != nil {
		return nil, fmt.Errorf("query ranking entries: %w", err)
	}
	defer rows.Close()
	records := make([]ranking.Record, 0)
	for rows.Next() {
		var record ranking.Record
		if err = rows.Scan(&record.SourceID, &record.PlayerKey, &record.Name, &record.Position, &record.Team, &record.Rank); err != nil {
			return nil, fmt.Errorf("scan ranking entry: %w", err)
		}
		records = append(records, record)
	}
	return records, rows.Err()
}

func (store *DraftEventStore) RankingStatuses(ctx context.Context) (map[string]ranking.SourceStatus, error) {
	rows, err := store.database.QueryContext(ctx, `SELECT id, refreshed_at, published_at, record_count FROM ranking_sources`)
	if err != nil {
		return nil, fmt.Errorf("query ranking sources: %w", err)
	}
	defer rows.Close()
	statuses := make(map[string]ranking.SourceStatus)
	for rows.Next() {
		var id, refreshed, published string
		var status ranking.SourceStatus
		if err = rows.Scan(&id, &refreshed, &published, &status.RecordCount); err != nil {
			return nil, fmt.Errorf("scan ranking source: %w", err)
		}
		parsed, parseErr := time.Parse(time.RFC3339Nano, refreshed)
		if parseErr != nil {
			return nil, fmt.Errorf("parse ranking refresh time: %w", parseErr)
		}
		status.RefreshedAt, status.PublishedAt = &parsed, published
		statuses[id] = status
	}
	return statuses, rows.Err()
}
