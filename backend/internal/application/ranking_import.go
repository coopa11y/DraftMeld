package application

import (
	"encoding/csv"
	"fmt"
	"html"
	"io"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"github.com/coopa11y/DraftMeld/backend/internal/domain/ranking"
)

type rankingCandidate struct {
	key, name, position, team string
	score                     float64
}

func parseRankingSource(sourceID string, input io.Reader) ([]ranking.Record, string, error) {
	switch sourceID {
	case "redraft-ecr":
		return parseECR(input)
	case "dynasty-1qb":
		return parseDynasty(input, sourceID, "value_1qb")
	case "dynasty-superflex":
		return parseDynasty(input, sourceID, "value_2qb")
	case "expected-opportunity":
		return parseOpportunity(input)
	case "cbs-ppr":
		return parseCBS(input)
	case "espn-ppr-online":
		return parseESPNPPR(input)
	case "yahoo-standard":
		return parseYahooStandard(input)
	default:
		if strings.HasPrefix(sourceID, "draft-sharks-") {
			return parseDraftSharks(input, sourceID)
		}
		if strings.HasPrefix(sourceID, "sleeper-adp-") {
			return parseSleeperADP(input, sourceID)
		}
		return nil, "", fmt.Errorf("unsupported ranking source %q", sourceID)
	}
}

var (
	cbsRowPattern     = regexp.MustCompile(`(?s)<div class="player-row[^"]*">.*?<div class="rank">([0-9]+)</div>.*?<a href="/nfl/players/[0-9]+/([^/]+)/fantasy/">.*?<span class="team position">(QB|RB|WR|TE|K|DST|D/ST|DEF)(?:\s+\$[0-9]+)?</span>`)
	cbsUpdatedPattern = regexp.MustCompile(`Updated\s+([^<]+)`)
)

func parseCBS(input io.Reader) ([]ranking.Record, string, error) {
	contents, err := io.ReadAll(input)
	if err != nil {
		return nil, "", fmt.Errorf("read CBS rankings: %w", err)
	}
	matches := cbsRowPattern.FindAllStringSubmatch(string(contents), -1)
	records := make([]ranking.Record, 0, 200)
	lastRank := 0
	for _, match := range matches {
		rank, rankErr := strconv.Atoi(match[1])
		if rankErr != nil {
			continue
		}
		if lastRank > 0 && rank <= lastRank {
			break
		}
		name := displayNameFromSlug(match[2])
		records = append(records, canonicalizeRankingRecord(ranking.Record{SourceID: "cbs-ppr", Name: name, Position: match[3], Rank: rank}))
		lastRank = rank
	}
	published := "Current CBS page"
	if updated := cbsUpdatedPattern.FindSubmatch(contents); len(updated) == 2 {
		published = "Updated " + strings.TrimSpace(html.UnescapeString(string(updated[1])))
	}
	return records, published, nil
}

func displayNameFromSlug(slug string) string {
	words := strings.Split(slug, "-")
	for index, word := range words {
		if word == "ii" || word == "iii" || word == "iv" {
			words[index] = strings.ToUpper(word)
			continue
		}
		if word != "" {
			words[index] = strings.ToUpper(word[:1]) + word[1:]
		}
	}
	return strings.Join(words, " ")
}

func readCSV(input io.Reader) ([]map[string]string, error) {
	reader := csv.NewReader(input)
	reader.ReuseRecord = false
	headings, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("read CSV headings: %w", err)
	}
	rows := make([]map[string]string, 0)
	for {
		values, readErr := reader.Read()
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return nil, fmt.Errorf("read CSV row: %w", readErr)
		}
		row := make(map[string]string, len(headings))
		for index, heading := range headings {
			if index < len(values) {
				row[heading] = values[index]
			}
		}
		rows = append(rows, row)
	}
	return rows, nil
}

func parseECR(input io.Reader) ([]ranking.Record, string, error) {
	rows, err := readCSV(input)
	if err != nil {
		return nil, "", err
	}
	candidates := make([]rankingCandidate, 0)
	published := ""
	for _, row := range rows {
		if row["page_type"] != "redraft-overall" || !supportedPosition(row["pos"]) {
			continue
		}
		score, scoreErr := strconv.ParseFloat(row["ecr"], 64)
		if scoreErr != nil || score <= 0 {
			continue
		}
		candidates = append(candidates, candidate(row["player"], row["pos"], row["team"], score))
		if row["scrape_date"] > published {
			published = row["scrape_date"]
		}
	}
	return rankCandidates("redraft-ecr", candidates, false), published, nil
}

func parseDynasty(input io.Reader, sourceID, valueColumn string) ([]ranking.Record, string, error) {
	rows, err := readCSV(input)
	if err != nil {
		return nil, "", err
	}
	candidates := make([]rankingCandidate, 0)
	published := ""
	for _, row := range rows {
		if !supportedPosition(row["pos"]) {
			continue
		}
		value, valueErr := strconv.ParseFloat(row[valueColumn], 64)
		if valueErr != nil || value <= 0 {
			continue
		}
		candidates = append(candidates, candidate(row["player"], row["pos"], row["team"], value))
		if row["scrape_date"] > published {
			published = row["scrape_date"]
		}
	}
	return rankCandidates(sourceID, candidates, true), published, nil
}

func parseOpportunity(input io.Reader) ([]ranking.Record, string, error) {
	rows, err := readCSV(input)
	if err != nil {
		return nil, "", err
	}
	aggregated := make(map[string]rankingCandidate)
	for _, row := range rows {
		if !supportedPosition(row["position"]) {
			continue
		}
		value, valueErr := strconv.ParseFloat(row["total_fantasy_points_exp"], 64)
		if valueErr != nil {
			continue
		}
		entry := candidate(row["full_name"], row["position"], row["posteam"], value)
		current := aggregated[entry.key]
		entry.score += current.score
		aggregated[entry.key] = entry
	}
	candidates := make([]rankingCandidate, 0, len(aggregated))
	for _, entry := range aggregated {
		candidates = append(candidates, entry)
	}
	return rankCandidates("expected-opportunity", candidates, true), "2025 season", nil
}

func candidate(name, position, team string, score float64) rankingCandidate {
	position = normalizePosition(position)
	team = strings.ToUpper(strings.TrimSpace(team))
	return rankingCandidate{key: canonicalRankingKey(name, position, team), name: strings.TrimSpace(name), position: position, team: team, score: score}
}

func rankCandidates(sourceID string, candidates []rankingCandidate, descending bool) []ranking.Record {
	sort.Slice(candidates, func(left, right int) bool {
		if candidates[left].score == candidates[right].score {
			return candidates[left].key < candidates[right].key
		}
		if descending {
			return candidates[left].score > candidates[right].score
		}
		return candidates[left].score < candidates[right].score
	})
	records := make([]ranking.Record, 0, len(candidates))
	seen := make(map[string]bool)
	for _, entry := range candidates {
		if entry.key == "" || seen[entry.key] || math.IsNaN(entry.score) {
			continue
		}
		seen[entry.key] = true
		records = append(records, ranking.Record{SourceID: sourceID, PlayerKey: entry.key, Name: entry.name, Position: entry.position, Team: entry.team, Rank: len(records) + 1})
	}
	return records
}

func normalizePlayerKey(name string) string {
	parts := strings.Fields(strings.TrimSpace(name))
	if len(parts) > 1 {
		suffix := strings.Trim(strings.ToLower(parts[len(parts)-1]), ".,")
		switch suffix {
		case "jr", "sr", "ii", "iii", "iv", "v":
			parts = parts[:len(parts)-1]
			name = strings.Join(parts, " ")
		}
	}
	return strings.Map(func(value rune) rune {
		if unicode.IsLetter(value) || unicode.IsDigit(value) {
			return unicode.ToLower(value)
		}
		return -1
	}, name)
}

func supportedPosition(position string) bool {
	switch normalizePosition(position) {
	case "QB", "RB", "WR", "TE", "K", "DST":
		return true
	}
	return false
}
