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
	rows, err := store.database.QueryContext(ctx, `
SELECT id, league_id, player_id, action, target_event_id, created_at, cost
FROM draft_events
WHERE league_id = ?
ORDER BY id`, leagueID)
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
		if err = rows.Scan(&event.ID, &event.LeagueID, &event.PlayerID, &action, &target, &created, &event.Cost); err != nil {
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
INSERT INTO draft_events (league_id, player_id, action, target_event_id, created_at, cost)
VALUES (?, ?, ?, ?, ?, ?)`, event.LeagueID, event.PlayerID, event.Action, event.TargetEventID, event.CreatedAt.Format(time.RFC3339Nano), event.Cost)
	if err != nil {
		return draft.Event{}, fmt.Errorf("append draft event: %w", err)
	}
	event.ID, err = result.LastInsertId()
	if err != nil {
		return draft.Event{}, fmt.Errorf("read draft event ID: %w", err)
	}
	return event, nil
}
