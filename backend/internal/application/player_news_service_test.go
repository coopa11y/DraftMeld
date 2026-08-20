package application

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coopa11y/DraftMeld/backend/internal/domain/news"
	"github.com/coopa11y/DraftMeld/backend/internal/domain/player"
)

type playerNewsRepositoryStub struct {
	sources      []news.Source
	events       []news.Event
	availability []news.Availability
	players      []player.Player
}

func (stub *playerNewsRepositoryStub) NewsSources(context.Context) ([]news.Source, error) {
	return append([]news.Source(nil), stub.sources...), nil
}
func (stub *playerNewsRepositoryStub) SaveNewsSource(_ context.Context, source news.Source) error {
	stub.sources = append(stub.sources, source)
	return nil
}
func (stub *playerNewsRepositoryStub) SetNewsSourceEnabled(_ context.Context, id string, enabled bool) (bool, error) {
	for index := range stub.sources {
		if stub.sources[index].ID == id {
			stub.sources[index].Enabled = enabled
			return true, nil
		}
	}
	return false, nil
}
func (stub *playerNewsRepositoryStub) DeleteNewsSource(_ context.Context, id string) (bool, error) {
	for index := range stub.sources {
		if stub.sources[index].ID == id && !stub.sources[index].BuiltIn {
			stub.sources = append(stub.sources[:index], stub.sources[index+1:]...)
			return true, nil
		}
	}
	return false, nil
}
func (stub *playerNewsRepositoryStub) RecordNewsRefresh(_ context.Context, id string, refreshedAt time.Time, refreshError string) error {
	for index := range stub.sources {
		if stub.sources[index].ID == id {
			stub.sources[index].LastRefreshedAt, stub.sources[index].LastError = refreshedAt, refreshError
		}
	}
	return nil
}
func (stub *playerNewsRepositoryStub) SaveNewsEvents(_ context.Context, events []news.Event) error {
	stub.events = append(stub.events, events...)
	return nil
}
func (stub *playerNewsRepositoryStub) SaveAvailability(_ context.Context, availability []news.Availability) error {
	stub.availability = append(stub.availability, availability...)
	return nil
}
func (stub *playerNewsRepositoryStub) PlayerNews(context.Context) ([]news.Event, []news.Availability, error) {
	return stub.events, stub.availability, nil
}
func (stub *playerNewsRepositoryStub) AllCanonicalPlayers(context.Context) ([]player.Player, error) {
	return stub.players, nil
}

func TestPlayerNewsParsesRSSAndMatchesCanonicalPlayers(t *testing.T) {
	published := "Thu, 20 Aug 2026 14:00:00 +0000"
	contents := []byte(`<rss><channel><item><guid>one</guid><title>Jordan Hale limited by hamstring injury</title><description>Jordan Hale was limited in practice.</description><link>https://example.com/one</link><pubDate>` + published + `</pubDate></item><item><guid>two</guid><title>Team preview</title></item></channel></rss>`)
	items, err := parseSyndicatedItems(contents, time.Now())
	if err != nil || len(items) != 2 {
		t.Fatalf("parse RSS: %v, items=%d", err, len(items))
	}
	if got := classifyPlayerNews(items[0].Title); got != "injury" {
		t.Fatalf("classify injury = %q", got)
	}
	matches := newPlayerNameMatcher([]player.Player{{ID: "p1", Name: "Jordan Hale"}, {ID: "p2", Name: "Al"}}).match(items[0].Title)
	if len(matches) != 1 || matches[0].ID != "p1" {
		t.Fatalf("matched players = %#v", matches)
	}
}

func TestPlayerNewsParsesAtomAndRejectsUnknownXML(t *testing.T) {
	items, err := parseSyndicatedItems([]byte(`<feed><entry><id>a</id><title>Player traded</title><summary>Details</summary><updated>2026-08-20T14:00:00Z</updated><link href="https://example.com/a"/></entry></feed>`), time.Now())
	if err != nil || len(items) != 1 || items[0].URL != "https://example.com/a" {
		t.Fatalf("parse Atom: %#v, %v", items, err)
	}
	if _, err = parseSyndicatedItems([]byte(`<document/>`), time.Now()); err == nil {
		t.Fatal("expected unsupported XML error")
	}
}

func TestPlayerNewsRefreshesRSSAndSleeper(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.URL.Path == "/rss" {
			_, _ = response.Write([]byte(`<rss><channel><item><guid>one</guid><title>Jordan Hale ruled out</title><link>https://example.com/one</link></item></channel></rss>`))
			return
		}
		_, _ = response.Write([]byte(`{"100":{"first_name":"Jordan","last_name":"Hale","position":"WR","team":"MIN","status":"Active","injury_status":"Questionable","injury_body_part":"Hamstring","practice_participation":"Limited","news_updated":1787234400000}}`))
	}))
	defer server.Close()
	now := time.Date(2026, 8, 20, 15, 0, 0, 0, time.UTC)
	repository := &playerNewsRepositoryStub{
		players: []player.Player{{ID: "p1", Name: "Jordan Hale", Position: "WR", Team: "MIN"}},
		sources: []news.Source{
			{ID: "rss", Name: "Test RSS", Kind: news.SourceRSS, URL: server.URL + "/rss", Enabled: true, RefreshMinutes: 5},
			{ID: "sleeper", Name: "Sleeper", Kind: news.SourceSleeper, URL: server.URL + "/players", Enabled: true, RefreshMinutes: 1440},
		},
	}
	service := NewPlayerNewsService(repository)
	service.client, service.now = server.Client(), func() time.Time { return now }
	result := service.Refresh(t.Context(), false)
	if result.Refreshed != 2 || len(result.Errors) != 0 {
		t.Fatalf("refresh result = %#v", result)
	}
	if len(repository.availability) != 1 || repository.availability[0].Status != "Questionable" {
		t.Fatalf("availability = %#v", repository.availability)
	}
	if len(repository.events) != 1 {
		t.Fatalf("events = %#v", repository.events)
	}
	second := service.Refresh(t.Context(), false)
	if second.Skipped != 2 {
		t.Fatalf("second refresh = %#v", second)
	}
}

func TestPlayerNewsFeedAndSourceManagement(t *testing.T) {
	now := time.Date(2026, 8, 20, 15, 0, 0, 0, time.UTC)
	repository := &playerNewsRepositoryStub{
		players:      []player.Player{{ID: "p1", Name: "Jordan Hale"}},
		events:       []news.Event{{ID: "e1", PlayerIDs: []string{"p1"}, Title: "Update", ObservedAt: now}},
		availability: []news.Availability{{PlayerID: "p1", Status: "Out", UpdatedAt: now.Add(time.Minute)}},
		sources:      []news.Source{{ID: "built-in", BuiltIn: true}, {ID: "custom", Kind: news.SourceRSS}},
	}
	service := NewPlayerNewsService(repository)
	feed, err := service.Feed(t.Context())
	if err != nil || len(feed.Updates) != 1 || feed.Updates[0].Availability.Status != "Out" || !feed.UpdatedAt.Equal(now.Add(time.Minute)) {
		t.Fatalf("feed = %#v, %v", feed, err)
	}
	if err = service.SetSourceEnabled(t.Context(), "custom", false); err != nil || repository.sources[1].Enabled {
		t.Fatalf("disable source: %v", err)
	}
	if err = service.SetSourceEnabled(t.Context(), "missing", false); err != ErrNewsSourceNotFound {
		t.Fatalf("missing source error = %v", err)
	}
	if err = service.DeleteSource(t.Context(), "built-in"); err != ErrBuiltInNewsSource {
		t.Fatalf("built-in delete error = %v", err)
	}
	if err = service.DeleteSource(t.Context(), "custom"); err != nil {
		t.Fatalf("delete custom source: %v", err)
	}
}

func TestPlayerNewsSourceValidation(t *testing.T) {
	for _, rawURL := range []string{"http://example.com/feed", "https://127.0.0.1/feed", "https://user:pass@example.com/feed"} {
		if _, err := validatePublicFeedURL(rawURL); err == nil {
			t.Fatalf("expected URL rejection for %s", rawURL)
		}
	}
	repository := &playerNewsRepositoryStub{}
	service := NewPlayerNewsService(repository)
	if _, err := service.AddRSSSource(t.Context(), "", "https://8.8.8.8/feed", "", 15); err == nil {
		t.Fatal("expected missing name error")
	}
	if _, err := service.AddRSSSource(t.Context(), "Feed", "https://8.8.8.8/feed", "", 2); err == nil {
		t.Fatal("expected refresh interval error")
	}
	source, err := service.AddRSSSource(t.Context(), "Feed", "https://8.8.8.8/feed#fragment", "", 15)
	if err != nil || source.Attribution != "Feed" || strings.Contains(source.URL, "#") {
		t.Fatalf("add source = %#v, %v", source, err)
	}
}
