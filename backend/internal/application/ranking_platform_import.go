package application

import (
	"encoding/json"
	"fmt"
	"html"
	"io"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/coopa11y/DraftMeld/backend/internal/domain/ranking"
)

const espnPPRFilter = `{"players":{"filterActive":{"value":true},"filterSlotIds":{"value":[0,2,4,6,16,17]},"limit":500,"sortDraftRanks":{"sortPriority":100,"sortAsc":true,"value":"PPR"}}}`

var yahooStandardRankPattern = regexp.MustCompile(`<li class="Listitem Phone-fz-lg">([1-9][0-9]{0,2})\.\s*([^<]{2,80})</li>`)

type espnPlayerPool struct {
	Players []espnPlayerPoolEntry `json:"players"`
}

type espnPlayerPoolEntry struct {
	Player espnRankedPlayer `json:"player"`
}

type espnRankedPlayer struct {
	ID              int                      `json:"id"`
	FullName        string                   `json:"fullName"`
	DefaultPosition int                      `json:"defaultPositionId"`
	ProTeamID       int                      `json:"proTeamId"`
	DraftRanks      map[string]espnDraftRank `json:"draftRanksByRankType"`
	Ownership       *espnOwnership           `json:"ownership"`
}

type espnDraftRank struct {
	Rank int `json:"rank"`
}

type espnOwnership struct {
	AverageDraftPosition float64 `json:"averageDraftPosition"`
}

func parseESPNPPR(input io.Reader) ([]ranking.Record, string, error) {
	var response espnPlayerPool
	if err := json.NewDecoder(input).Decode(&response); err != nil {
		return nil, "", fmt.Errorf("decode ESPN PPR rankings: %w", err)
	}
	records := make([]ranking.Record, 0, len(response.Players))
	seenRanks := make(map[int]bool)
	for _, entry := range response.Players {
		position := espnPosition(entry.Player.DefaultPosition)
		rankValue := entry.Player.DraftRanks["PPR"].Rank
		if !supportedPosition(position) || rankValue < 1 || rankValue > 500 || seenRanks[rankValue] || strings.TrimSpace(entry.Player.FullName) == "" {
			continue
		}
		seenRanks[rankValue] = true
		record := ranking.Record{
			SourceID: "espn-ppr-online", Name: strings.TrimSpace(entry.Player.FullName), Position: position,
			Team: espnNFLTeam(entry.Player.ProTeamID), Rank: rankValue, ProviderID: strconv.Itoa(entry.Player.ID),
		}
		if entry.Player.Ownership != nil && entry.Player.Ownership.AverageDraftPosition > 0 {
			record.ADP = entry.Player.Ownership.AverageDraftPosition
		}
		records = append(records, canonicalizeRankingRecord(record))
	}
	sort.Slice(records, func(left, right int) bool { return records[left].Rank < records[right].Rank })
	if len(records) < 100 {
		return nil, "", fmt.Errorf("parse ESPN PPR rankings: found only %d usable players", len(records))
	}
	return records, "Current ESPN PPR draft order", nil
}

func parseYahooStandard(input io.Reader) ([]ranking.Record, string, error) {
	contents, err := io.ReadAll(input)
	if err != nil {
		return nil, "", fmt.Errorf("read Yahoo rankings: %w", err)
	}
	matches := yahooStandardRankPattern.FindAllSubmatch(contents, -1)
	records := make([]ranking.Record, 0, len(matches))
	for _, match := range matches {
		rankValue, rankErr := strconv.Atoi(string(match[1]))
		if rankErr != nil || rankValue != len(records)+1 {
			continue
		}
		name := strings.TrimSpace(html.UnescapeString(string(match[2])))
		position, team := "", ""
		if team = canonicalNFLTeam(name); team != "" {
			position = "DST"
		}
		records = append(records, canonicalizeRankingRecord(ranking.Record{
			SourceID: "yahoo-standard", Name: name, Position: position, Team: team, Rank: rankValue,
		}))
	}
	if len(records) < 100 {
		return nil, "", fmt.Errorf("parse Yahoo Standard rankings: found only %d usable players", len(records))
	}
	return records, "Current Yahoo default Standard order", nil
}

func espnPosition(id int) string {
	return map[int]string{1: "QB", 2: "RB", 3: "WR", 4: "TE", 5: "K", 16: "DST"}[id]
}

func espnNFLTeam(id int) string {
	return map[int]string{
		1: "ATL", 2: "BUF", 3: "CHI", 4: "CIN", 5: "CLE", 6: "DAL", 7: "DEN", 8: "DET",
		9: "GB", 10: "TEN", 11: "IND", 12: "KC", 13: "LV", 14: "LAR", 15: "MIA", 16: "MIN",
		17: "NE", 18: "NO", 19: "NYG", 20: "NYJ", 21: "PHI", 22: "ARI", 23: "PIT", 24: "LAC",
		25: "SF", 26: "SEA", 27: "TB", 28: "WAS", 29: "CAR", 30: "JAX", 33: "BAL", 34: "HOU",
	}[id]
}
