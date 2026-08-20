package sqlite

import (
	"testing"
	"time"

	"github.com/coopa11y/DraftMeld/backend/internal/domain/news"
	"github.com/coopa11y/DraftMeld/backend/internal/domain/player"
)

func TestPlayerNewsPersistenceLifecycle(t *testing.T) {
	store, err := Open(t.TempDir() + "/draftmeld.db")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	ctx := t.Context()
	resolved, err := store.ResolvePlayer(ctx, player.Candidate{IdentityKey: "jordan-hale-wr", Name: "Jordan Hale", Position: "WR", Team: "MIN", Provider: "sleeper", ProviderID: "100"}, "player-1")
	if err != nil {
		t.Fatal(err)
	}
	providerID, found, err := store.PlayerIDByProvider(ctx, "sleeper", "100")
	if err != nil || !found || providerID != resolved.ID {
		t.Fatalf("provider player = %q, %v, %v", providerID, found, err)
	}
	allPlayers, err := store.AllCanonicalPlayers(ctx)
	if err != nil || len(allPlayers) != 1 {
		t.Fatalf("canonical players = %#v, %v", allPlayers, err)
	}
	sources, err := store.NewsSources(ctx)
	if err != nil || len(sources) != 3 {
		t.Fatalf("built-in sources = %#v, %v", sources, err)
	}
	custom := news.Source{ID: "rss-test", Name: "Test feed", Kind: news.SourceRSS, URL: "https://example.com/feed", Attribution: "Example", Enabled: true, RefreshMinutes: 15}
	if err = store.SaveNewsSource(ctx, custom); err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 8, 20, 15, 0, 0, 0, time.UTC)
	event := news.Event{ID: "event-1", SourceID: custom.ID, PlayerIDs: []string{resolved.ID}, Category: "injury", Title: "Jordan Hale limited", URL: "https://example.com/story", PublishedAt: now, ObservedAt: now}
	if err = store.SaveNewsEvents(ctx, []news.Event{event, event}); err != nil {
		t.Fatal(err)
	}
	availability := news.Availability{PlayerID: resolved.ID, Status: "Questionable", Injury: "Hamstring", PracticeParticipation: "Limited", SourceID: "sleeper-player-status", UpdatedAt: now}
	if err = store.SaveAvailability(ctx, []news.Availability{availability}); err != nil {
		t.Fatal(err)
	}
	events, statuses, err := store.PlayerNews(ctx)
	if err != nil || len(events) != 1 || len(statuses) != 1 || statuses[0].Status != "Questionable" {
		t.Fatalf("stored news = %#v, %#v, %v", events, statuses, err)
	}
	if err = store.RecordNewsRefresh(ctx, custom.ID, now, "temporary failure"); err != nil {
		t.Fatal(err)
	}
	sources, _ = store.NewsSources(ctx)
	if len(sources) != 4 {
		t.Fatalf("sources after custom add = %d", len(sources))
	}
	updated, err := store.SetNewsSourceEnabled(ctx, custom.ID, false)
	if err != nil || !updated {
		t.Fatalf("disable source = %v, %v", updated, err)
	}
	deleted, err := store.DeleteNewsSource(ctx, custom.ID)
	if err != nil || !deleted {
		t.Fatalf("delete source = %v, %v", deleted, err)
	}
	events, _, err = store.PlayerNews(ctx)
	if err != nil || len(events) != 0 {
		t.Fatalf("events after source delete = %#v, %v", events, err)
	}
}
