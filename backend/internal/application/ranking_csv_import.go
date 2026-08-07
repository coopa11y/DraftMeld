package application

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/coopa11y/DraftMeld/backend/internal/domain/ranking"
)

var rankingCSVHeaderAliases = map[string][]string{
	"name":       {"name", "player", "playername", "playerfullname"},
	"rank":       {"rank", "overallrank", "rk"},
	"position":   {"position", "pos"},
	"team":       {"team", "nflteam", "tm"},
	"adp":        {"adp", "averagedraftposition"},
	"tier":       {"tier"},
	"providerId": {"providerid", "playerid", "id"},
}

func (service *RankingService) ImportCSV(ctx context.Context, name string, input io.Reader, mapping map[string]string) (ranking.SourceStatus, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return ranking.SourceStatus{}, errors.New("ranking source name is required")
	}
	if len(name) > 80 {
		return ranking.SourceStatus{}, errors.New("ranking source name must be 80 characters or fewer")
	}
	rows, err := csv.NewReader(input).ReadAll()
	if err != nil || len(rows) < 2 {
		return ranking.SourceStatus{}, errors.New("ranking CSV requires a header and at least one player")
	}
	headers := make(map[string]int, len(rows[0]))
	for index, header := range rows[0] {
		headers[normalizeCSVHeader(header)] = index
	}
	columns := rankingCSVColumnIndexes(headers, mapping)
	for _, required := range []string{"name", "rank", "position"} {
		if _, exists := columns[required]; !exists {
			return ranking.SourceStatus{}, fmt.Errorf("ranking CSV is missing %s", required)
		}
	}
	sourceID := "custom-" + slugify(name)
	if sourceID == "custom-" {
		return ranking.SourceStatus{}, errors.New("ranking source name requires at least one letter or number")
	}
	customSources, err := service.repository.CustomRankingSources(ctx)
	if err != nil {
		return ranking.SourceStatus{}, err
	}
	for _, source := range customSources {
		if source.ID == sourceID && !strings.EqualFold(source.Name, name) {
			return ranking.SourceStatus{}, fmt.Errorf("ranking source name conflicts with existing source %q", source.Name)
		}
	}
	recordsByPlayer := make(map[string]ranking.Record)
	for rowIndex, row := range rows[1:] {
		playerName := value(row, columns["name"])
		if playerName == "" {
			continue
		}
		position := normalizePosition(value(row, columns["position"]))
		if !supportedPosition(position) {
			return ranking.SourceStatus{}, fmt.Errorf("ranking row %d has unsupported position %q", rowIndex+2, position)
		}
		rankValue, rankErr := strconv.Atoi(value(row, columns["rank"]))
		if rankErr != nil || rankValue < 1 {
			return ranking.SourceStatus{}, fmt.Errorf("ranking row %d requires a positive rank", rowIndex+2)
		}
		team := ""
		if index, exists := columns["team"]; exists {
			team = strings.ToUpper(value(row, index))
		}
		record := canonicalizeRankingRecord(ranking.Record{SourceID: sourceID, Name: playerName, Position: position, Team: team, Rank: rankValue})
		if index, exists := columns["adp"]; exists && value(row, index) != "" {
			record.ADP, err = strconv.ParseFloat(value(row, index), 64)
			if err != nil || record.ADP < 0 {
				return ranking.SourceStatus{}, fmt.Errorf("ranking row %d has invalid ADP", rowIndex+2)
			}
		}
		if index, exists := columns["tier"]; exists && value(row, index) != "" {
			record.Tier, err = strconv.Atoi(value(row, index))
			if err != nil || record.Tier < 1 {
				return ranking.SourceStatus{}, fmt.Errorf("ranking row %d has invalid tier", rowIndex+2)
			}
		}
		if index, exists := columns["providerId"]; exists {
			record.ProviderID = value(row, index)
		}
		if current, exists := recordsByPlayer[record.PlayerKey]; !exists || record.Rank < current.Rank {
			recordsByPlayer[record.PlayerKey] = record
		}
	}
	if len(recordsByPlayer) == 0 {
		return ranking.SourceStatus{}, errors.New("ranking CSV contained no usable players")
	}
	records := make([]ranking.Record, 0, len(recordsByPlayer))
	for _, record := range recordsByPlayer {
		records = append(records, record)
	}
	sort.Slice(records, func(left, right int) bool {
		if records[left].Rank == records[right].Rank {
			return records[left].PlayerKey < records[right].PlayerKey
		}
		return records[left].Rank < records[right].Rank
	})
	refreshed := time.Now().UTC()
	records, err = resolveRankingPlayers(ctx, service.repository, records, refreshed)
	if err != nil {
		return ranking.SourceStatus{}, err
	}
	definition := ranking.SourceDefinition{
		ID: sourceID, Name: name, Description: "Private ranking CSV uploaded by the user.",
		Methodology: "User-supplied ordinal player ranking", License: "Private user data",
		DefaultWeight: 1, ImportMode: "csv-upload", Role: "ranking", IsCustom: true,
	}
	if err = service.repository.ReplaceRankings(ctx, definition, records, "Private CSV import", refreshed); err != nil {
		return ranking.SourceStatus{}, err
	}
	return ranking.SourceStatus{SourceDefinition: definition, RecordCount: len(records), RefreshedAt: &refreshed, PublishedAt: "Private CSV import"}, nil
}

func rankingCSVColumnIndexes(headers map[string]int, mapping map[string]string) map[string]int {
	indexes := make(map[string]int)
	for _, column := range []string{"name", "rank", "position", "team", "adp", "tier", "providerId"} {
		if sourceHeader, explicitlyMapped := mapping[column]; explicitlyMapped {
			if index, exists := headers[normalizeCSVHeader(sourceHeader)]; exists && sourceHeader != "" {
				indexes[column] = index
			}
			continue
		}
		for _, alias := range rankingCSVHeaderAliases[column] {
			if index, exists := headers[alias]; exists {
				indexes[column] = index
				break
			}
		}
	}
	return indexes
}
