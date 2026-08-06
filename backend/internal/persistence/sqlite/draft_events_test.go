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
		LeagueID: "league-a", PlayerID: "p001", Action: draft.ActionDraft,
	})
	if err != nil {
		t.Fatalf("append event: %v", err)
	}
	events, err := store.List(context.Background(), "league-a")
	if err != nil {
		t.Fatalf("list events: %v", err)
	}
	if len(events) != 1 || events[0].ID != written.ID || events[0].PlayerID != "p001" {
		t.Fatalf("unexpected events: %#v", events)
	}
}
