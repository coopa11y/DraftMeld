package application

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
	"unicode"

	"github.com/coopa11y/DraftMeld/backend/internal/domain/news"
	"github.com/coopa11y/DraftMeld/backend/internal/domain/player"
)

const maximumNewsFeedSize = 20 << 20

type syndicatedItem struct {
	ID          string
	Title       string
	Summary     string
	URL         string
	PublishedAt time.Time
}

type rssDocument struct {
	Channel struct {
		Items []struct {
			GUID        string `xml:"guid"`
			Title       string `xml:"title"`
			Description string `xml:"description"`
			Link        string `xml:"link"`
			Published   string `xml:"pubDate"`
		} `xml:"item"`
	} `xml:"channel"`
}

type atomDocument struct {
	Entries []struct {
		ID      string `xml:"id"`
		Title   string `xml:"title"`
		Summary string `xml:"summary"`
		Content string `xml:"content"`
		Updated string `xml:"updated"`
		Link    struct {
			Href string `xml:"href,attr"`
		} `xml:"link"`
	} `xml:"entry"`
}

func (service *PlayerNewsService) fetchRSS(ctx context.Context, source news.Source, players []player.Player, observedAt time.Time) ([]news.Event, error) {
	contents, err := service.fetchFeed(ctx, source)
	if err != nil {
		return nil, err
	}
	items, err := parseSyndicatedItems(contents, observedAt)
	if err != nil {
		return nil, err
	}
	matcher := newPlayerNameMatcher(players)
	events := make([]news.Event, 0, len(items))
	for _, item := range items {
		matches := matcher.match(item.Title + " " + item.Summary)
		if len(matches) == 0 {
			continue
		}
		event := news.Event{
			ID: newsEventID(source.ID, item.ID, item.URL, item.Title), SourceID: source.ID,
			Category: classifyPlayerNews(item.Title + " " + item.Summary), Title: strings.TrimSpace(item.Title),
			Summary: strings.TrimSpace(item.Summary), URL: strings.TrimSpace(item.URL),
			PublishedAt: item.PublishedAt, ObservedAt: observedAt, Confidence: "medium",
		}
		for _, matched := range matches {
			event.PlayerIDs = append(event.PlayerIDs, matched.ID)
			event.PlayerNames = append(event.PlayerNames, matched.Name)
		}
		events = append(events, event)
	}
	return events, nil
}

func (service *PlayerNewsService) fetchFeed(ctx context.Context, source news.Source) ([]byte, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, source.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("build feed request: %w", err)
	}
	if source.Kind == news.SourceSleeper {
		request.Header.Set("Accept", "application/json")
	} else {
		request.Header.Set("Accept", "application/rss+xml, application/atom+xml, application/xml, text/xml")
	}
	request.Header.Set("User-Agent", "DraftMeld/0.3 (+https://github.com/coopa11y/DraftMeld)")
	response, err := service.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("download feed: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download feed: HTTP %d", response.StatusCode)
	}
	contents, err := io.ReadAll(io.LimitReader(response.Body, maximumNewsFeedSize+1))
	if err != nil {
		return nil, fmt.Errorf("read feed: %w", err)
	}
	if len(contents) > maximumNewsFeedSize {
		return nil, errorsNewFeedTooLarge()
	}
	return contents, nil
}

func errorsNewFeedTooLarge() error {
	return fmt.Errorf("feed exceeds %d MiB", maximumNewsFeedSize>>20)
}

func parseSyndicatedItems(contents []byte, fallback time.Time) ([]syndicatedItem, error) {
	var rss rssDocument
	if err := xml.Unmarshal(contents, &rss); err == nil && len(rss.Channel.Items) > 0 {
		items := make([]syndicatedItem, 0, len(rss.Channel.Items))
		for _, item := range rss.Channel.Items {
			items = append(items, syndicatedItem{ID: item.GUID, Title: item.Title, Summary: item.Description, URL: item.Link, PublishedAt: parseNewsTime(item.Published, fallback)})
		}
		return items, nil
	}
	var atom atomDocument
	if err := xml.Unmarshal(contents, &atom); err != nil {
		return nil, fmt.Errorf("parse feed XML: %w", err)
	}
	if len(atom.Entries) == 0 {
		return nil, fmt.Errorf("parse feed XML: no RSS items or Atom entries")
	}
	items := make([]syndicatedItem, 0, len(atom.Entries))
	for _, entry := range atom.Entries {
		summary := entry.Summary
		if summary == "" {
			summary = entry.Content
		}
		items = append(items, syndicatedItem{ID: entry.ID, Title: entry.Title, Summary: summary, URL: entry.Link.Href, PublishedAt: parseNewsTime(entry.Updated, fallback)})
	}
	return items, nil
}

func parseNewsTime(value string, fallback time.Time) time.Time {
	for _, layout := range []string{time.RFC1123Z, time.RFC1123, time.RFC3339Nano, time.RFC3339, "Mon, 02 Jan 2006 15:04:05 MST"} {
		if parsed, err := time.Parse(layout, strings.TrimSpace(value)); err == nil {
			return parsed.UTC()
		}
	}
	return fallback
}

type playerNameMatcher struct {
	players []player.Player
}

func newPlayerNameMatcher(players []player.Player) playerNameMatcher {
	return playerNameMatcher{players: players}
}

func (matcher playerNameMatcher) match(contents string) []player.Player {
	normalizedContents := " " + normalizeNewsText(contents) + " "
	matches := make([]player.Player, 0)
	seen := make(map[string]bool)
	for _, candidate := range matcher.players {
		name := normalizeNewsText(candidate.Name)
		if len(name) < 5 || seen[candidate.ID] || !strings.Contains(normalizedContents, " "+name+" ") {
			continue
		}
		seen[candidate.ID] = true
		matches = append(matches, candidate)
	}
	return matches
}

func normalizeNewsText(value string) string {
	var builder strings.Builder
	space := true
	for _, current := range strings.ToLower(value) {
		if unicode.IsLetter(current) || unicode.IsDigit(current) {
			builder.WriteRune(current)
			space = false
		} else if !space {
			builder.WriteByte(' ')
			space = true
		}
	}
	return strings.TrimSpace(builder.String())
}

func newsEventID(sourceID, upstreamID, articleURL, title string) string {
	key := upstreamID
	if key == "" {
		key = articleURL
	}
	if key == "" {
		key = title
	}
	sum := sha256.Sum256([]byte(sourceID + "\x00" + key))
	return "news-" + hex.EncodeToString(sum[:12])
}

func classifyPlayerNews(value string) string {
	normalized := normalizeNewsText(value)
	for category, terms := range map[string][]string{
		"injury":      {"injury", "injured", "questionable", "doubtful", "ruled out", "practice", "concussion", "hamstring", "ankle", "knee"},
		"transaction": {"traded", "trade", "signed", "released", "waived", "claimed"},
		"suspension":  {"suspended", "suspension"},
		"retirement":  {"retired", "retirement"},
		"depth-chart": {"starter", "starting", "depth chart", "backup"},
	} {
		for _, term := range terms {
			if strings.Contains(normalized, term) {
				return category
			}
		}
	}
	return "news"
}
