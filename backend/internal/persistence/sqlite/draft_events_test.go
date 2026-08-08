package sqlite

import (
	"context"
	"reflect"
	"testing"

	"github.com/coopa11y/DraftMeld/backend/internal/domain/draft"
)

func TestDraftTradeAssetsAndConditionsPersist(t *testing.T) {
	store, err := Open(t.TempDir() + "/draftmeld.db")
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	defer store.Close()

	written, err := store.SavePickTrade(t.Context(), draft.PickTrade{
		LeagueID:       "league-a",
		TeamOneNumber:  1,
		TeamTwoNumber:  2,
		Season:         2026,
		TeamOnePlayers: []string{"p002"},
		TeamTwoPlayers: []string{"p001"},
		TeamOneFuturePicks: []draft.FuturePick{{
			Season: 2027, Round: 1, OriginalTeamNumber: 2, Condition: "If Team 2 reaches the final", ConditionStatus: "pending",
		}},
		TeamTwoBudgets: []draft.BudgetAsset{{Kind: "faab", Season: 2027, Amount: 25}},
	})
	if err != nil {
		t.Fatalf("save draft trade: %v", err)
	}

	trades, err := store.ListPickTrades(t.Context(), "league-a")
	if err != nil || len(trades) != 1 {
		t.Fatalf("list draft trades: trades=%#v error=%v", trades, err)
	}
	if !reflect.DeepEqual(trades[0].TeamOnePlayers, []string{"p002"}) ||
		!reflect.DeepEqual(trades[0].TeamTwoPlayers, []string{"p001"}) ||
		!reflect.DeepEqual(trades[0].TeamTwoBudgets, []draft.BudgetAsset{{Kind: "faab", Season: 2027, Amount: 25}}) ||
		trades[0].TeamOneFuturePicks[0].ConditionStatus != "pending" {
		t.Fatalf("trade assets did not round-trip: %#v", trades[0])
	}

	written.TeamOneFuturePicks[0].ConditionStatus = "met"
	if err = store.UpdatePickTrade(t.Context(), written); err != nil {
		t.Fatalf("update draft trade: %v", err)
	}
	trades, err = store.ListPickTrades(t.Context(), "league-a")
	if err != nil || trades[0].TeamOneFuturePicks[0].ConditionStatus != "met" {
		t.Fatalf("condition resolution did not persist: trades=%#v error=%v", trades, err)
	}
}

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

func TestDraftEventsCanBeListedAndReplacedBySeason(t *testing.T) {
	store, err := Open(t.TempDir() + "/draftmeld.db")
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	defer store.Close()

	for _, event := range []draft.Event{
		{LeagueID: "league-a", Season: 2026, PlayerID: "p001", Action: draft.ActionDraft},
		{LeagueID: "league-a", Season: 2027, PlayerID: "p002", Action: draft.ActionDraft},
	} {
		if _, err = store.Append(t.Context(), event); err != nil {
			t.Fatalf("append event: %v", err)
		}
	}
	events, err := store.ListSeason(t.Context(), "league-a", 2027)
	if err != nil || len(events) != 1 || events[0].PlayerID != "p002" {
		t.Fatalf("unexpected 2027 events: events=%#v error=%v", events, err)
	}
	if err = store.ReplaceDraftEventsForSeason(t.Context(), "league-a", 2027, []draft.Event{{PlayerID: "p003", Action: draft.ActionTaken}}); err != nil {
		t.Fatalf("replace 2027 events: %v", err)
	}
	events, err = store.ListSeason(t.Context(), "league-a", 2026)
	if err != nil || len(events) != 1 || events[0].PlayerID != "p001" {
		t.Fatalf("2026 events changed: events=%#v error=%v", events, err)
	}
	events, err = store.ListSeason(t.Context(), "league-a", 2027)
	if err != nil || len(events) != 1 || events[0].PlayerID != "p003" || events[0].Season != 2027 {
		t.Fatalf("unexpected replacement: events=%#v error=%v", events, err)
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
