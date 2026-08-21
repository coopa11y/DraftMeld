package application

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/coopa11y/DraftMeld/backend/internal/domain/news"
	"github.com/coopa11y/DraftMeld/backend/internal/domain/player"
)

type sleeperNewsPlayer struct {
	FirstName             string `json:"first_name"`
	LastName              string `json:"last_name"`
	Position              string `json:"position"`
	Team                  string `json:"team"`
	Status                string `json:"status"`
	InjuryStatus          string `json:"injury_status"`
	InjuryBodyPart        string `json:"injury_body_part"`
	InjuryNotes           string `json:"injury_notes"`
	PracticeParticipation string `json:"practice_participation"`
	NewsUpdated           int64  `json:"news_updated"`
}

func (service *PlayerNewsService) fetchSleeper(ctx context.Context, source news.Source, players []player.Player, observedAt time.Time) ([]news.Event, []news.Availability, error) {
	contents, err := service.fetchFeed(ctx, source)
	if err != nil {
		return nil, nil, err
	}
	var upstream map[string]sleeperNewsPlayer
	if err = json.Unmarshal(contents, &upstream); err != nil {
		return nil, nil, fmt.Errorf("parse Sleeper player status: %w", err)
	}
	known := make(map[string][]player.Player, len(players))
	for _, item := range players {
		known[normalizeNewsText(item.Name)] = append(known[normalizeNewsText(item.Name)], item)
	}
	availability := make([]news.Availability, 0, len(players))
	for _, upstreamPlayer := range upstream {
		name := strings.TrimSpace(upstreamPlayer.FirstName + " " + upstreamPlayer.LastName)
		matched, exists := bestSleeperMatch(known[normalizeNewsText(name)], upstreamPlayer)
		if !exists {
			continue
		}
		updatedAt := observedAt
		if upstreamPlayer.NewsUpdated > 0 {
			updatedAt = time.UnixMilli(upstreamPlayer.NewsUpdated).UTC()
		}
		status := strings.TrimSpace(upstreamPlayer.InjuryStatus)
		if status == "" {
			status = strings.TrimSpace(upstreamPlayer.Status)
		}
		if status == "" {
			status = "Active"
		}
		injury := strings.TrimSpace(upstreamPlayer.InjuryBodyPart)
		if injury == "" {
			injury = strings.TrimSpace(upstreamPlayer.InjuryNotes)
		}
		availability = append(availability, news.Availability{
			PlayerID: matched.ID, Status: status, Injury: injury,
			PracticeParticipation: strings.TrimSpace(upstreamPlayer.PracticeParticipation),
			SourceID:              source.ID, SourceName: source.Name, UpdatedAt: updatedAt,
		})
	}
	return nil, availability, nil
}

func bestSleeperMatch(candidates []player.Player, upstream sleeperNewsPlayer) (player.Player, bool) {
	if len(candidates) == 1 {
		return candidates[0], true
	}
	for _, candidate := range candidates {
		if strings.EqualFold(candidate.Position, upstream.Position) && (candidate.Team == "" || upstream.Team == "" || strings.EqualFold(candidate.Team, upstream.Team)) {
			return candidate, true
		}
	}
	return player.Player{}, false
}
