package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/coopa11y/DraftMeld/backend/internal/domain/news"
)

func (store *DraftEventStore) NewsSources(ctx context.Context) ([]news.Source, error) {
	rows, err := store.database.QueryContext(ctx, `SELECT id, name, kind, source_url, attribution, enabled, built_in, refresh_minutes, last_refreshed_at, last_error FROM player_news_sources ORDER BY built_in DESC, name, id`)
	if err != nil {
		return nil, fmt.Errorf("query player news sources: %w", err)
	}
	defer rows.Close()
	sources := make([]news.Source, 0)
	for rows.Next() {
		var source news.Source
		var lastRefreshed sql.NullString
		if err = rows.Scan(&source.ID, &source.Name, &source.Kind, &source.URL, &source.Attribution, &source.Enabled, &source.BuiltIn, &source.RefreshMinutes, &lastRefreshed, &source.LastError); err != nil {
			return nil, fmt.Errorf("scan player news source: %w", err)
		}
		if lastRefreshed.Valid {
			source.LastRefreshedAt, _ = time.Parse(time.RFC3339Nano, lastRefreshed.String)
		}
		sources = append(sources, source)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate player news sources: %w", err)
	}
	return sources, nil
}

func (store *DraftEventStore) SaveNewsSource(ctx context.Context, source news.Source) error {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	_, err := store.database.ExecContext(ctx, `INSERT INTO player_news_sources (id, name, kind, source_url, attribution, enabled, built_in, refresh_minutes, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, 0, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET name=excluded.name, source_url=excluded.source_url, attribution=excluded.attribution, enabled=excluded.enabled, refresh_minutes=excluded.refresh_minutes, updated_at=excluded.updated_at
		WHERE player_news_sources.built_in = 0`, source.ID, source.Name, source.Kind, source.URL, source.Attribution, source.Enabled, source.RefreshMinutes, now, now)
	if err != nil {
		return fmt.Errorf("save player news source: %w", err)
	}
	return nil
}

func (store *DraftEventStore) SetNewsSourceEnabled(ctx context.Context, id string, enabled bool) (bool, error) {
	result, err := store.database.ExecContext(ctx, `UPDATE player_news_sources SET enabled = ?, updated_at = ? WHERE id = ?`, enabled, time.Now().UTC().Format(time.RFC3339Nano), id)
	if err != nil {
		return false, fmt.Errorf("update player news source: %w", err)
	}
	count, err := result.RowsAffected()
	return count > 0, err
}

func (store *DraftEventStore) DeleteNewsSource(ctx context.Context, id string) (bool, error) {
	result, err := store.database.ExecContext(ctx, `DELETE FROM player_news_sources WHERE id = ? AND built_in = 0`, id)
	if err != nil {
		return false, fmt.Errorf("delete player news source: %w", err)
	}
	count, err := result.RowsAffected()
	return count > 0, err
}

func (store *DraftEventStore) RecordNewsRefresh(ctx context.Context, id string, refreshedAt time.Time, refreshError string) error {
	_, err := store.database.ExecContext(ctx, `UPDATE player_news_sources SET last_refreshed_at = ?, last_error = ?, updated_at = ? WHERE id = ?`, refreshedAt.UTC().Format(time.RFC3339Nano), refreshError, refreshedAt.UTC().Format(time.RFC3339Nano), id)
	if err != nil {
		return fmt.Errorf("record player news refresh: %w", err)
	}
	return nil
}

func (store *DraftEventStore) SaveNewsEvents(ctx context.Context, events []news.Event) error {
	tx, err := store.database.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin player news save: %w", err)
	}
	defer tx.Rollback()
	for _, event := range events {
		if _, err = tx.ExecContext(ctx, `INSERT INTO player_news_events (id, source_id, category, title, summary, article_url, published_at, observed_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?) ON CONFLICT(id) DO UPDATE SET category=excluded.category, title=excluded.title, summary=excluded.summary, article_url=excluded.article_url, published_at=excluded.published_at`,
			event.ID, event.SourceID, event.Category, event.Title, event.Summary, event.URL, event.PublishedAt.UTC().Format(time.RFC3339Nano), event.ObservedAt.UTC().Format(time.RFC3339Nano)); err != nil {
			return fmt.Errorf("save player news event: %w", err)
		}
		for _, playerID := range event.PlayerIDs {
			if _, err = tx.ExecContext(ctx, `INSERT INTO player_news_event_players (event_id, player_id) VALUES (?, ?) ON CONFLICT DO NOTHING`, event.ID, playerID); err != nil {
				return fmt.Errorf("link player news event: %w", err)
			}
		}
	}
	if _, err = tx.ExecContext(ctx, `DELETE FROM player_news_events WHERE published_at < ?`, time.Now().UTC().AddDate(0, -2, 0).Format(time.RFC3339Nano)); err != nil {
		return fmt.Errorf("prune player news: %w", err)
	}
	return tx.Commit()
}

func (store *DraftEventStore) SaveAvailability(ctx context.Context, items []news.Availability) error {
	tx, err := store.database.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin availability save: %w", err)
	}
	defer tx.Rollback()
	for _, item := range items {
		_, err = tx.ExecContext(ctx, `INSERT INTO player_availability (player_id, status, injury, practice_participation, source_id, updated_at)
			VALUES (?, ?, ?, ?, ?, ?) ON CONFLICT(player_id) DO UPDATE SET status=excluded.status, injury=excluded.injury, practice_participation=excluded.practice_participation, source_id=excluded.source_id, updated_at=excluded.updated_at
			WHERE excluded.updated_at >= player_availability.updated_at`, item.PlayerID, item.Status, item.Injury, item.PracticeParticipation, item.SourceID, item.UpdatedAt.UTC().Format(time.RFC3339Nano))
		if err != nil {
			return fmt.Errorf("save player availability: %w", err)
		}
	}
	return tx.Commit()
}

func (store *DraftEventStore) PlayerNews(ctx context.Context) ([]news.Event, []news.Availability, error) {
	events, err := store.newsEvents(ctx)
	if err != nil {
		return nil, nil, err
	}
	availability, err := store.playerAvailability(ctx)
	return events, availability, err
}

func (store *DraftEventStore) newsEvents(ctx context.Context) ([]news.Event, error) {
	rows, err := store.database.QueryContext(ctx, `SELECT e.id, e.source_id, s.name, e.category, e.title, e.summary, e.article_url, e.published_at, e.observed_at, p.id, p.name
		FROM player_news_events e JOIN player_news_sources s ON s.id=e.source_id
		JOIN player_news_event_players ep ON ep.event_id=e.id JOIN canonical_players p ON p.id=ep.player_id
		ORDER BY e.published_at DESC, e.id LIMIT 500`)
	if err != nil {
		return nil, fmt.Errorf("query player news: %w", err)
	}
	defer rows.Close()
	byID := make(map[string]int)
	events := make([]news.Event, 0)
	for rows.Next() {
		var event news.Event
		var published, observed, playerID, playerName string
		if err = rows.Scan(&event.ID, &event.SourceID, &event.SourceName, &event.Category, &event.Title, &event.Summary, &event.URL, &published, &observed, &playerID, &playerName); err != nil {
			return nil, fmt.Errorf("scan player news: %w", err)
		}
		if index, exists := byID[event.ID]; exists {
			events[index].PlayerIDs = append(events[index].PlayerIDs, playerID)
			events[index].PlayerNames = append(events[index].PlayerNames, playerName)
			continue
		}
		event.PublishedAt, _ = time.Parse(time.RFC3339Nano, published)
		event.ObservedAt, _ = time.Parse(time.RFC3339Nano, observed)
		event.PlayerIDs, event.PlayerNames = []string{playerID}, []string{playerName}
		event.Confidence = eventConfidence(event.SourceID, event.Category)
		byID[event.ID] = len(events)
		events = append(events, event)
	}
	return events, rows.Err()
}

func (store *DraftEventStore) playerAvailability(ctx context.Context) ([]news.Availability, error) {
	rows, err := store.database.QueryContext(ctx, `SELECT a.player_id, a.status, a.injury, a.practice_participation, a.source_id, s.name, a.updated_at FROM player_availability a JOIN player_news_sources s ON s.id=a.source_id`)
	if err != nil {
		return nil, fmt.Errorf("query player availability: %w", err)
	}
	defer rows.Close()
	items := make([]news.Availability, 0)
	for rows.Next() {
		var item news.Availability
		var updated string
		if err = rows.Scan(&item.PlayerID, &item.Status, &item.Injury, &item.PracticeParticipation, &item.SourceID, &item.SourceName, &updated); err != nil {
			return nil, fmt.Errorf("scan player availability: %w", err)
		}
		item.UpdatedAt, _ = time.Parse(time.RFC3339Nano, updated)
		items = append(items, item)
	}
	return items, rows.Err()
}

func eventConfidence(sourceID, category string) string {
	if sourceID == "sleeper-player-status" || category == "official-status" {
		return "high"
	}
	return "medium"
}
