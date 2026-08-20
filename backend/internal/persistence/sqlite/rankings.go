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
	if source.IsCustom {
		if _, err = tx.ExecContext(ctx, `INSERT INTO custom_ranking_sources
(id, name, description, methodology, license, project_url, data_url, default_weight, import_mode, role)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(id) DO UPDATE SET name=excluded.name, description=excluded.description,
methodology=excluded.methodology, license=excluded.license, project_url=excluded.project_url,
data_url=excluded.data_url, default_weight=excluded.default_weight,
import_mode=excluded.import_mode, role=excluded.role`, source.ID, source.Name, source.Description,
			source.Methodology, source.License, source.ProjectURL, source.DataURL, source.DefaultWeight,
			source.ImportMode, source.Role); err != nil {
			return fmt.Errorf("save custom ranking source: %w", err)
		}
	}
	if _, err = tx.ExecContext(ctx, "DELETE FROM ranking_entries WHERE source_id = ?", source.ID); err != nil {
		return fmt.Errorf("clear ranking entries: %w", err)
	}
	statement, err := tx.PrepareContext(ctx, `INSERT INTO ranking_entries
(source_id, player_key, player_name, position, nfl_team, source_rank, adp, tier, games, bye_week,
 floor_projection, consensus_projection, source_projection, ceiling_projection, source_value, injury_risk, schedule_strength)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return fmt.Errorf("prepare ranking entries: %w", err)
	}
	defer statement.Close()
	for _, record := range records {
		if _, err = statement.ExecContext(ctx, source.ID, record.PlayerKey, record.Name, record.Position, record.Team,
			record.Rank, record.ADP, record.Tier, record.Games, record.ByeWeek, record.FloorProjection,
			record.ConsensusProjection, record.SourceProjection, record.CeilingProjection, record.SourceValue,
			record.InjuryRisk, record.ScheduleStrength); err != nil {
			return fmt.Errorf("save ranking entry: %w", err)
		}
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit ranking refresh: %w", err)
	}
	return nil
}

func (store *DraftEventStore) RankingRecords(ctx context.Context) ([]ranking.Record, error) {
	rows, err := store.database.QueryContext(ctx, `SELECT source_id, player_key, player_name, position, nfl_team,
source_rank, adp, tier, games, bye_week, floor_projection, consensus_projection, source_projection,
ceiling_projection, source_value, injury_risk, schedule_strength FROM ranking_entries ORDER BY source_id, source_rank`)
	if err != nil {
		return nil, fmt.Errorf("query ranking entries: %w", err)
	}
	defer rows.Close()
	records := make([]ranking.Record, 0)
	for rows.Next() {
		var record ranking.Record
		if err = rows.Scan(&record.SourceID, &record.PlayerKey, &record.Name, &record.Position, &record.Team,
			&record.Rank, &record.ADP, &record.Tier, &record.Games, &record.ByeWeek, &record.FloorProjection,
			&record.ConsensusProjection, &record.SourceProjection, &record.CeilingProjection,
			&record.SourceValue, &record.InjuryRisk, &record.ScheduleStrength); err != nil {
			return nil, fmt.Errorf("scan ranking entry: %w", err)
		}
		records = append(records, record)
	}
	return records, rows.Err()
}

func (store *DraftEventStore) CustomRankingSources(ctx context.Context) ([]ranking.SourceDefinition, error) {
	rows, err := store.database.QueryContext(ctx, `SELECT id, name, description, methodology, license,
project_url, data_url, default_weight, import_mode, role FROM custom_ranking_sources ORDER BY name, id`)
	if err != nil {
		return nil, fmt.Errorf("query custom ranking sources: %w", err)
	}
	defer rows.Close()
	definitions := make([]ranking.SourceDefinition, 0)
	for rows.Next() {
		var definition ranking.SourceDefinition
		if err = rows.Scan(&definition.ID, &definition.Name, &definition.Description, &definition.Methodology,
			&definition.License, &definition.ProjectURL, &definition.DataURL, &definition.DefaultWeight,
			&definition.ImportMode, &definition.Role); err != nil {
			return nil, fmt.Errorf("scan custom ranking source: %w", err)
		}
		definition.IsCustom = true
		definition.DefaultEnabled = true
		definitions = append(definitions, definition)
	}
	return definitions, rows.Err()
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
