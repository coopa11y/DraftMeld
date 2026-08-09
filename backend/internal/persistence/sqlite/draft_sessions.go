package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/coopa11y/DraftMeld/backend/internal/domain/draft"
)

func (store *DraftEventStore) GetDraftSession(ctx context.Context, leagueID string, season int) (draft.Session, bool, error) {
	var session draft.Session
	var status, startedAt, updatedAt, resetJSON, resetStatus string
	err := store.database.QueryRowContext(ctx, `
SELECT league_id, season, status, COALESCE(started_at, ''), updated_at, reset_events, reset_status
FROM draft_sessions WHERE league_id = ? AND season = ?`, leagueID, season).Scan(
		&session.LeagueID, &session.Season, &status, &startedAt, &updatedAt, &resetJSON, &resetStatus,
	)
	if err == sql.ErrNoRows {
		return draft.Session{}, false, nil
	}
	if err != nil {
		return draft.Session{}, false, fmt.Errorf("load draft session: %w", err)
	}
	session.Status = draft.SessionStatus(status)
	session.ResetStatus = draft.SessionStatus(resetStatus)
	if startedAt != "" {
		if session.StartedAt, err = time.Parse(time.RFC3339Nano, startedAt); err != nil {
			return draft.Session{}, false, fmt.Errorf("parse draft session start time: %w", err)
		}
	}
	if session.UpdatedAt, err = time.Parse(time.RFC3339Nano, updatedAt); err != nil {
		return draft.Session{}, false, fmt.Errorf("parse draft session update time: %w", err)
	}
	if err = json.Unmarshal([]byte(resetJSON), &session.ResetEvents); err != nil {
		return draft.Session{}, false, fmt.Errorf("decode reset draft events: %w", err)
	}
	return session, true, nil
}

func (store *DraftEventStore) SaveDraftSession(ctx context.Context, session draft.Session) error {
	resetJSON, err := json.Marshal(session.ResetEvents)
	if err != nil {
		return fmt.Errorf("encode reset draft events: %w", err)
	}
	session.UpdatedAt = time.Now().UTC()
	var startedAt any
	if !session.StartedAt.IsZero() {
		startedAt = session.StartedAt.Format(time.RFC3339Nano)
	}
	_, err = store.database.ExecContext(ctx, `
INSERT INTO draft_sessions (league_id, season, status, started_at, updated_at, reset_events, reset_status)
VALUES (?, ?, ?, ?, ?, ?, ?)
ON CONFLICT (league_id, season) DO UPDATE SET
    status = excluded.status,
    started_at = excluded.started_at,
    updated_at = excluded.updated_at,
    reset_events = excluded.reset_events,
    reset_status = excluded.reset_status`, session.LeagueID, session.Season, session.Status, startedAt,
		session.UpdatedAt.Format(time.RFC3339Nano), string(resetJSON), session.ResetStatus)
	if err != nil {
		return fmt.Errorf("save draft session: %w", err)
	}
	return nil
}

func (store *DraftEventStore) ResetDraftSession(ctx context.Context, session draft.Session) error {
	tx, err := store.database.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin draft reset: %w", err)
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, `DELETE FROM draft_events WHERE league_id = ? AND season = ?`, session.LeagueID, session.Season); err != nil {
		return fmt.Errorf("clear current draft events: %w", err)
	}
	if err = saveDraftSessionTx(ctx, tx, session); err != nil {
		return err
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit draft reset: %w", err)
	}
	return nil
}

func (store *DraftEventStore) RestoreDraftSession(ctx context.Context, session draft.Session, events []draft.Event) error {
	tx, err := store.database.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin draft reset restore: %w", err)
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, `DELETE FROM draft_events WHERE league_id = ? AND season = ?`, session.LeagueID, session.Season); err != nil {
		return fmt.Errorf("clear reset draft events: %w", err)
	}
	createdAt := time.Now().UTC()
	for index, event := range events {
		if _, err = tx.ExecContext(ctx, `
INSERT INTO draft_events (league_id, season, player_id, action, target_event_id, created_at, cost, team_number)
VALUES (?, ?, ?, ?, NULL, ?, ?, ?)`, session.LeagueID, session.Season, event.PlayerID, event.Action,
			createdAt.Add(time.Duration(index)*time.Nanosecond).Format(time.RFC3339Nano), event.Cost, event.TeamNumber); err != nil {
			return fmt.Errorf("restore reset draft event: %w", err)
		}
	}
	if err = saveDraftSessionTx(ctx, tx, session); err != nil {
		return err
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit draft reset restore: %w", err)
	}
	return nil
}

func saveDraftSessionTx(ctx context.Context, tx *sql.Tx, session draft.Session) error {
	resetJSON, err := json.Marshal(session.ResetEvents)
	if err != nil {
		return fmt.Errorf("encode reset draft events: %w", err)
	}
	session.UpdatedAt = time.Now().UTC()
	var startedAt any
	if !session.StartedAt.IsZero() {
		startedAt = session.StartedAt.Format(time.RFC3339Nano)
	}
	_, err = tx.ExecContext(ctx, `
INSERT INTO draft_sessions (league_id, season, status, started_at, updated_at, reset_events, reset_status)
VALUES (?, ?, ?, ?, ?, ?, ?)
ON CONFLICT (league_id, season) DO UPDATE SET
    status = excluded.status,
    started_at = excluded.started_at,
    updated_at = excluded.updated_at,
    reset_events = excluded.reset_events,
    reset_status = excluded.reset_status`, session.LeagueID, session.Season, session.Status, startedAt,
		session.UpdatedAt.Format(time.RFC3339Nano), string(resetJSON), session.ResetStatus)
	if err != nil {
		return fmt.Errorf("save draft session transaction: %w", err)
	}
	return nil
}
