package application

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/coopa11y/DraftMeld/backend/internal/domain/projection"
)

var projectionStatColumns = []string{
	"reception", "passingYard", "passingTouchdown", "interception", "rushingYard", "rushingTouchdown",
	"receivingYard", "receivingTouchdown", "fieldGoalMade", "extraPointMade", "defenseSack",
	"defenseInterception", "defenseFumbleRecovery", "defenseTouchdown", "defenseSafety",
}

type ProjectionRepository interface {
	ReplaceProjections(context.Context, projection.SourceStatus, []projection.Record) error
	ProjectionRecords(context.Context) ([]projection.Record, error)
	ProjectionStatuses(context.Context) ([]projection.SourceStatus, error)
}

type ProjectionService struct{ repository ProjectionRepository }

func NewProjectionService(repository ProjectionRepository) *ProjectionService {
	return &ProjectionService{repository: repository}
}

func (service *ProjectionService) Sources(ctx context.Context) ([]projection.SourceStatus, error) {
	return service.repository.ProjectionStatuses(ctx)
}

func (service *ProjectionService) ImportCSV(ctx context.Context, name string, input io.Reader) (projection.SourceStatus, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return projection.SourceStatus{}, errors.New("projection source name is required")
	}
	reader := csv.NewReader(input)
	rows, err := reader.ReadAll()
	if err != nil || len(rows) < 2 {
		return projection.SourceStatus{}, errors.New("projection CSV requires a header and at least one player")
	}
	headers := make(map[string]int, len(rows[0]))
	for index, header := range rows[0] {
		headers[strings.TrimSpace(header)] = index
	}
	for _, required := range []string{"name", "position", "team"} {
		if _, exists := headers[required]; !exists {
			return projection.SourceStatus{}, fmt.Errorf("projection CSV is missing %s", required)
		}
	}
	sourceID := slugify(name)
	records := make([]projection.Record, 0, len(rows)-1)
	seen := make(map[string]bool)
	for rowIndex, row := range rows[1:] {
		nameValue, position, team := value(row, headers["name"]), normalizePosition(value(row, headers["position"])), value(row, headers["team"])
		key := canonicalRankingKey(nameValue, position, team)
		if key == "" || seen[key] {
			continue
		}
		seen[key] = true
		record := projection.Record{SourceID: sourceID, PlayerKey: key, Name: nameValue, Position: position, Team: team, Stats: make(map[string]float64)}
		if index, exists := headers["byeWeek"]; exists {
			record.ByeWeek, _ = strconv.Atoi(value(row, index))
		}
		if index, exists := headers["adp"]; exists {
			record.ADP, _ = strconv.ParseFloat(value(row, index), 64)
		}
		for _, column := range projectionStatColumns {
			if index, exists := headers[column]; exists {
				parsed, parseErr := strconv.ParseFloat(value(row, index), 64)
				if parseErr != nil && value(row, index) != "" {
					return projection.SourceStatus{}, fmt.Errorf("projection row %d has invalid %s", rowIndex+2, column)
				}
				record.Stats[column] = parsed
			}
		}
		records = append(records, record)
	}
	if len(records) == 0 {
		return projection.SourceStatus{}, errors.New("projection CSV contained no usable players")
	}
	status := projection.SourceStatus{ID: sourceID, Name: name, RecordCount: len(records), ImportedAt: time.Now().UTC()}
	if err = service.repository.ReplaceProjections(ctx, status, records); err != nil {
		return projection.SourceStatus{}, err
	}
	return status, nil
}

func (service *ProjectionService) LeagueValues(ctx context.Context, scoring map[string]float64) (map[string]projection.LeagueValue, error) {
	records, err := service.repository.ProjectionRecords(ctx)
	if err != nil {
		return nil, err
	}
	totals := make(map[string]projection.LeagueValue)
	counts := make(map[string]int)
	adpCounts := make(map[string]int)
	for _, record := range records {
		points := 0.0
		for stat, amount := range record.Stats {
			points += amount * scoring[stat]
		}
		value := totals[record.PlayerKey]
		value.ProjectedPoints += points
		if record.ADP > 0 {
			value.ADP += record.ADP
			adpCounts[record.PlayerKey]++
		}
		if value.ByeWeek == 0 {
			value.ByeWeek = record.ByeWeek
		}
		totals[record.PlayerKey] = value
		counts[record.PlayerKey]++
	}
	for playerID, total := range totals {
		count := float64(counts[playerID])
		total.ProjectedPoints /= count
		if adpCounts[playerID] > 0 {
			total.ADP /= float64(adpCounts[playerID])
		}
		totals[playerID] = total
	}
	return totals, nil
}

func value(row []string, index int) string {
	if index < 0 || index >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[index])
}
