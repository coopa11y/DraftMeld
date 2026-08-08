package sqlite

import (
	"context"
	"testing"

	"github.com/coopa11y/DraftMeld/backend/internal/domain/draft"
)

func TestDraftEventsPersist(t *testing.T) {
	store, err := Open(t.TempDir() + "/draftmeld.db")
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	defer store.Close()

	written, err := store.Append(context.Background(), draft.Event{
		LeagueID: "league-a", PlayerID: "p001", Action: draft.ActionDraft, TeamNumber: 4,
	})
	if err != nil {
		t.Fatalf("append event: %v", err)
	}
	events, err := store.List(context.Background(), "league-a")
	if err != nil {
		t.Fatalf("list events: %v", err)
	}
	if len(events) != 1 || events[0].ID != written.ID || events[0].PlayerID != "p001" || events[0].TeamNumber != 4 {
		t.Fatalf("unexpected events: %#v", events)
	}
}

func TestMigrationsAreRecordedAndIdempotent(t *testing.T) {
	path := t.TempDir() + "/draftmeld.db"
	store, err := Open(path)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	var migrationCount int
	if err = store.database.QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&migrationCount); err != nil {
		t.Fatalf("count migrations: %v", err)
	}
	if migrationCount < 1 {
		t.Fatal("expected at least one applied migration")
	}
	var initialMigration int
	if err = store.database.QueryRow(
		"SELECT COUNT(*) FROM schema_migrations WHERE version = '0001_draft_events.sql'",
	).Scan(&initialMigration); err != nil || initialMigration != 1 {
		t.Fatalf("initial migration was not recorded: count=%d error=%v", initialMigration, err)
	}
	initialCount := migrationCount
	if err = store.Close(); err != nil {
		t.Fatalf("close database: %v", err)
	}

	reopened, err := Open(path)
	if err != nil {
		t.Fatalf("reopen database: %v", err)
	}
	defer reopened.Close()
	if err = reopened.database.QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&migrationCount); err != nil {
		t.Fatalf("count migrations after reopen: %v", err)
	}
	if migrationCount != initialCount {
		t.Fatalf("migration count changed after reopen: before=%d after=%d", initialCount, migrationCount)
	}
}
