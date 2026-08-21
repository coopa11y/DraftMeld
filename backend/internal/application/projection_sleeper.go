package application

import (
	"fmt"
	"io"
	"strings"

	"github.com/coopa11y/DraftMeld/backend/internal/domain/projection"
)

var sleeperProjectionStatMap = map[string]string{
	"rec":      "reception",
	"pass_yd":  "passingYard",
	"pass_td":  "passingTouchdown",
	"pass_int": "interception",
	"rush_yd":  "rushingYard",
	"rush_td":  "rushingTouchdown",
	"rec_yd":   "receivingYard",
	"rec_td":   "receivingTouchdown",
	"pass_2pt": "passingTwoPointConversion",
	"rush_2pt": "rushingTwoPointConversion",
	"rec_2pt":  "receivingTwoPointConversion",
	"fum_lost": "fumbleLost",
}

func parseSleeperProjections(input io.Reader) ([]projection.Record, string, error) {
	items, err := decodeSleeperFeed(input)
	if err != nil {
		return nil, "", err
	}
	records := make([]projection.Record, 0, 600)
	for _, item := range items {
		position := normalizePosition(item.Player.Position)
		if !sleeperProjectionPosition(position) || item.Stats["pts_ppr"] <= 0 || item.PlayerID == "" {
			continue
		}
		name := strings.TrimSpace(item.Player.FirstName + " " + item.Player.LastName)
		if name == "" {
			continue
		}
		stats := make(map[string]float64, len(sleeperProjectionStatMap))
		for sleeperName, draftMeldName := range sleeperProjectionStatMap {
			if value := item.Stats[sleeperName]; value != 0 {
				stats[draftMeldName] = value
			}
		}
		records = append(records, projection.Record{
			SourceID: sleeperProjectionID, ProviderID: item.PlayerID,
			PlayerKey: canonicalRankingKey(name, position, item.Player.Team), Name: name,
			Position: position, Team: canonicalNFLTeam(item.Player.Team), Stats: stats,
		})
	}
	if len(records) < 300 {
		return nil, "", fmt.Errorf("Sleeper projection feed contained only %d usable offensive players", len(records))
	}
	return records, sleeperPublishedAt(items), nil
}

func sleeperProjectionPosition(position string) bool {
	switch position {
	case "QB", "RB", "WR", "TE":
		return true
	default:
		return false
	}
}
