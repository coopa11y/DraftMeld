package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/coopa11y/DraftMeld/backend/internal/domain/draft"
	_ "modernc.org/sqlite"
)

type DraftEventStore struct {
	database *sql.DB
}

func Open(path string) (*DraftEventStore, error) {
	database, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open SQLite database: %w", err)
	}
	database.SetMaxOpenConns(1)
	store := &DraftEventStore{database: database}
	if err = migrate(context.Background(), database); err != nil {
		_ = database.Close()
		return nil, err
	}
	return store, nil
}

func (store *DraftEventStore) Close() error {
	return store.database.Close()
}

func (store *DraftEventStore) List(ctx context.Context, leagueID string) ([]draft.Event, error) {
	return store.list(ctx, leagueID, 0)
}

func (store *DraftEventStore) ListSeason(ctx context.Context, leagueID string, season int) ([]draft.Event, error) {
	return store.list(ctx, leagueID, season)
}

func (store *DraftEventStore) list(ctx context.Context, leagueID string, season int) ([]draft.Event, error) {
	query := `
SELECT id, league_id, season, player_id, action, target_event_id, created_at, cost, team_number
FROM draft_events
WHERE league_id = ?`
	arguments := []any{leagueID}
	if season > 0 {
		query += ` AND season = ?`
		arguments = append(arguments, season)
	}
	query += ` ORDER BY id`
	rows, err := store.database.QueryContext(ctx, query, arguments...)
	if err != nil {
		return nil, fmt.Errorf("query draft events: %w", err)
	}
	defer rows.Close()

	events := make([]draft.Event, 0)
	for rows.Next() {
		var event draft.Event
		var action string
		var target sql.NullInt64
		var created string
		if err = rows.Scan(&event.ID, &event.LeagueID, &event.Season, &event.PlayerID, &action, &target, &created, &event.Cost, &event.TeamNumber); err != nil {
			return nil, fmt.Errorf("scan draft event: %w", err)
		}
		event.Action = draft.Action(action)
		if target.Valid {
			event.TargetEventID = &target.Int64
		}
		event.CreatedAt, err = time.Parse(time.RFC3339Nano, created)
		if err != nil {
			return nil, fmt.Errorf("parse draft event time: %w", err)
		}
		events = append(events, event)
	}
	return events, rows.Err()
}

func (store *DraftEventStore) Append(ctx context.Context, event draft.Event) (draft.Event, error) {
	event.CreatedAt = time.Now().UTC()
	result, err := store.database.ExecContext(ctx, `
INSERT INTO draft_events (league_id, season, player_id, action, target_event_id, created_at, cost, team_number)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)`, event.LeagueID, event.Season, event.PlayerID, event.Action, event.TargetEventID, event.CreatedAt.Format(time.RFC3339Nano), event.Cost, event.TeamNumber)
	if err != nil {
		return draft.Event{}, fmt.Errorf("append draft event: %w", err)
	}
	event.ID, err = result.LastInsertId()
	if err != nil {
		return draft.Event{}, fmt.Errorf("read draft event ID: %w", err)
	}
	return event, nil
}

func (store *DraftEventStore) ReplaceDraftEvents(ctx context.Context, leagueID string, events []draft.Event) error {
	season := 0
	if len(events) > 0 {
		season = events[0].Season
	}
	return store.ReplaceDraftEventsForSeason(ctx, leagueID, season, events)
}

func (store *DraftEventStore) ReplaceDraftEventsForSeason(ctx context.Context, leagueID string, season int, events []draft.Event) error {
	tx, err := store.database.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin draft reconciliation: %w", err)
	}
	defer tx.Rollback()
	deleteQuery := `DELETE FROM draft_events WHERE league_id = ?`
	deleteArguments := []any{leagueID}
	if season > 0 {
		deleteQuery += ` AND season = ?`
		deleteArguments = append(deleteArguments, season)
	}
	if _, err = tx.ExecContext(ctx, deleteQuery, deleteArguments...); err != nil {
		return fmt.Errorf("clear draft events: %w", err)
	}
	createdAt := time.Now().UTC()
	for index, event := range events {
		if _, err = tx.ExecContext(ctx, `INSERT INTO draft_events (league_id, season, player_id, action, target_event_id, created_at, cost, team_number) VALUES (?, ?, ?, ?, NULL, ?, ?, ?)`, leagueID, season, event.PlayerID, event.Action, createdAt.Add(time.Duration(index)*time.Nanosecond).Format(time.RFC3339Nano), event.Cost, event.TeamNumber); err != nil {
			return fmt.Errorf("replace draft event: %w", err)
		}
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit draft reconciliation: %w", err)
	}
	return nil
}
