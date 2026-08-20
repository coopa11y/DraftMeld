package application

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/coopa11y/DraftMeld/backend/internal/domain/news"
	"github.com/coopa11y/DraftMeld/backend/internal/domain/player"
)

var (
	ErrNewsSourceNotFound = errors.New("player news source not found")
	ErrBuiltInNewsSource  = errors.New("built-in player news sources cannot be deleted")
)

type PlayerNewsRepository interface {
	NewsSources(context.Context) ([]news.Source, error)
	SaveNewsSource(context.Context, news.Source) error
	SetNewsSourceEnabled(context.Context, string, bool) (bool, error)
	DeleteNewsSource(context.Context, string) (bool, error)
	RecordNewsRefresh(context.Context, string, time.Time, string) error
	SaveNewsEvents(context.Context, []news.Event) error
	SaveAvailability(context.Context, []news.Availability) error
	PlayerNews(context.Context) ([]news.Event, []news.Availability, error)
	AllCanonicalPlayers(context.Context) ([]player.Player, error)
}

type PlayerNewsService struct {
	repository PlayerNewsRepository
	client     *http.Client
	now        func() time.Time
}

func NewPlayerNewsService(repository PlayerNewsRepository) *PlayerNewsService {
	return &PlayerNewsService{repository: repository, client: newPublicFeedClient(), now: func() time.Time { return time.Now().UTC() }}
}

func (service *PlayerNewsService) Sources(ctx context.Context) ([]news.Source, error) {
	return service.repository.NewsSources(ctx)
}

func (service *PlayerNewsService) AddRSSSource(ctx context.Context, name, rawURL, attribution string, refreshMinutes int) (news.Source, error) {
	name, attribution = strings.TrimSpace(name), strings.TrimSpace(attribution)
	if name == "" {
		return news.Source{}, errors.New("source name is required")
	}
	if attribution == "" {
		attribution = name
	}
	if refreshMinutes == 0 {
		refreshMinutes = 15
	}
	if refreshMinutes < 5 || refreshMinutes > 1440 {
		return news.Source{}, errors.New("refresh interval must be between 5 and 1440 minutes")
	}
	normalizedURL, err := validatePublicFeedURL(rawURL)
	if err != nil {
		return news.Source{}, err
	}
	source := news.Source{
		ID: sourceID(name, normalizedURL), Name: name, Kind: news.SourceRSS, URL: normalizedURL,
		Attribution: attribution, Enabled: true, RefreshMinutes: refreshMinutes,
	}
	if err = service.repository.SaveNewsSource(ctx, source); err != nil {
		return news.Source{}, err
	}
	return source, nil
}

func (service *PlayerNewsService) SetSourceEnabled(ctx context.Context, id string, enabled bool) error {
	updated, err := service.repository.SetNewsSourceEnabled(ctx, id, enabled)
	if err != nil {
		return err
	}
	if !updated {
		return ErrNewsSourceNotFound
	}
	return nil
}

func (service *PlayerNewsService) DeleteSource(ctx context.Context, id string) error {
	sources, err := service.repository.NewsSources(ctx)
	if err != nil {
		return err
	}
	for _, source := range sources {
		if source.ID == id && source.BuiltIn {
			return ErrBuiltInNewsSource
		}
	}
	deleted, err := service.repository.DeleteNewsSource(ctx, id)
	if err != nil {
		return err
	}
	if !deleted {
		return ErrNewsSourceNotFound
	}
	return nil
}

func (service *PlayerNewsService) Feed(ctx context.Context) (news.Feed, error) {
	events, availability, err := service.repository.PlayerNews(ctx)
	if err != nil {
		return news.Feed{}, err
	}
	players, err := service.repository.AllCanonicalPlayers(ctx)
	if err != nil {
		return news.Feed{}, err
	}
	updates := make(map[string]*news.PlayerUpdate)
	for _, item := range players {
		updates[item.ID] = &news.PlayerUpdate{PlayerID: item.ID, PlayerName: item.Name, Events: []news.Event{}}
	}
	var updatedAt time.Time
	for index := range availability {
		item := availability[index]
		if update := updates[item.PlayerID]; update != nil {
			update.Availability = &item
		}
		if item.UpdatedAt.After(updatedAt) {
			updatedAt = item.UpdatedAt
		}
	}
	for _, event := range events {
		for _, playerID := range event.PlayerIDs {
			if update := updates[playerID]; update != nil && len(update.Events) < 5 {
				update.Events = append(update.Events, event)
			}
		}
		if event.ObservedAt.After(updatedAt) {
			updatedAt = event.ObservedAt
		}
	}
	feed := news.Feed{UpdatedAt: updatedAt, Updates: make([]news.PlayerUpdate, 0)}
	for _, update := range updates {
		if update.Availability != nil || len(update.Events) > 0 {
			feed.Updates = append(feed.Updates, *update)
		}
	}
	sort.Slice(feed.Updates, func(i, j int) bool {
		return strings.ToLower(feed.Updates[i].PlayerName) < strings.ToLower(feed.Updates[j].PlayerName)
	})
	return feed, nil
}

func (service *PlayerNewsService) Refresh(ctx context.Context, force bool) news.RefreshResult {
	sources, err := service.repository.NewsSources(ctx)
	if err != nil {
		return news.RefreshResult{Errors: []string{err.Error()}}
	}
	result := news.RefreshResult{Errors: []string{}}
	now := service.now()
	for _, source := range sources {
		if !source.Enabled || (!force && !sourceDue(source, now)) {
			result.Skipped++
			continue
		}
		if err = service.refreshSource(ctx, source, now); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("%s: %v", source.Name, err))
			_ = service.repository.RecordNewsRefresh(ctx, source.ID, now, err.Error())
			continue
		}
		_ = service.repository.RecordNewsRefresh(ctx, source.ID, now, "")
		result.Refreshed++
	}
	return result
}

func (service *PlayerNewsService) refreshSource(ctx context.Context, source news.Source, observedAt time.Time) error {
	players, err := service.repository.AllCanonicalPlayers(ctx)
	if err != nil {
		return err
	}
	switch source.Kind {
	case news.SourceRSS:
		events, fetchErr := service.fetchRSS(ctx, source, players, observedAt)
		if fetchErr != nil {
			return fetchErr
		}
		return service.repository.SaveNewsEvents(ctx, events)
	case news.SourceSleeper:
		events, availability, fetchErr := service.fetchSleeper(ctx, source, players, observedAt)
		if fetchErr != nil {
			return fetchErr
		}
		if fetchErr = service.repository.SaveAvailability(ctx, availability); fetchErr != nil {
			return fetchErr
		}
		return service.repository.SaveNewsEvents(ctx, events)
	default:
		return fmt.Errorf("unsupported player news source kind %q", source.Kind)
	}
}

func sourceDue(source news.Source, now time.Time) bool {
	return source.LastRefreshedAt.IsZero() || now.Sub(source.LastRefreshedAt) >= time.Duration(source.RefreshMinutes)*time.Minute
}

func sourceID(name, rawURL string) string {
	sum := sha256.Sum256([]byte(strings.ToLower(name) + "\x00" + rawURL))
	return "rss-" + hex.EncodeToString(sum[:8])
}
