package application

import (
	"encoding/json"
	"fmt"
	"io"
	"time"
)

const (
	sleeperProjectURL     = "https://sleeper.com/fantasy-football"
	sleeperAPIURL         = "https://docs.sleeper.com/"
	sleeperProjectionURL  = "https://api.sleeper.com/projections/nfl/2026?season_type=regular"
	sleeperProjectionID   = "sleeper-projections"
	sleeperProjectionName = "Sleeper statistical projections"
)

type sleeperFeedItem struct {
	Stats        map[string]float64 `json:"stats"`
	Category     string             `json:"category"`
	LastModified int64              `json:"last_modified"`
	Season       string             `json:"season"`
	SeasonType   string             `json:"season_type"`
	PlayerID     string             `json:"player_id"`
	Player       sleeperPlayer      `json:"player"`
}

type sleeperPlayer struct {
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	Position     string `json:"position"`
	Team         string `json:"team"`
	InjuryStatus string `json:"injury_status"`
}

func decodeSleeperFeed(input io.Reader) ([]sleeperFeedItem, error) {
	decoder := json.NewDecoder(io.LimitReader(input, (20<<20)+1))
	var items []sleeperFeedItem
	if err := decoder.Decode(&items); err != nil {
		return nil, fmt.Errorf("decode Sleeper feed: %w", err)
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("decode Sleeper feed: response contained no players")
	}
	return items, nil
}

func sleeperPublishedAt(items []sleeperFeedItem) string {
	var latest int64
	season := ""
	for _, item := range items {
		if item.LastModified > latest {
			latest = item.LastModified
		}
		if item.Season > season {
			season = item.Season
		}
	}
	if latest > 0 {
		return fmt.Sprintf("%s season; updated %s", season, time.UnixMilli(latest).UTC().Format(time.RFC3339))
	}
	return season + " season"
}
