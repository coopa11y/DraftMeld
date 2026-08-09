package application

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/coopa11y/DraftMeld/backend/internal/domain/projection"
)

var projectionStatColumns = []string{
	"reception", "passingYard", "passingTouchdown", "interception", "rushingYard", "rushingTouchdown",
	"receivingYard", "receivingTouchdown", "fieldGoalMade", "extraPointMade", "defenseSack",
	"defenseInterception", "defenseFumbleRecovery", "defenseTouchdown", "defenseSafety",
	"passingTwoPointConversion", "rushingTwoPointConversion", "receivingTwoPointConversion", "fumble", "fumbleLost",
	"passing300YardGame", "passing400YardGame", "rushing100YardGame", "rushing200YardGame",
	"receiving100YardGame", "receiving200YardGame", "fieldGoal0To39", "fieldGoal40To49", "fieldGoal50Plus",
	"fieldGoalMissed", "extraPointMissed", "defenseBlockedKick", "defenseTwoPointReturn", "defensePointsAllowed0",
	"defensePointsAllowed1To6", "defensePointsAllowed7To13", "defensePointsAllowed14To20",
	"defensePointsAllowed21To27", "defensePointsAllowed28To34", "defensePointsAllowed35Plus",
}

var nonCSVHeaderCharacter = regexp.MustCompile(`[^a-z0-9]+`)

var projectionHeaderAliases = map[string][]string{
	"name":                        {"name", "player", "playername", "playerfullname"},
	"position":                    {"position", "pos"},
	"team":                        {"team", "nflteam", "tm"},
	"adp":                         {"adp", "averagedraftposition"},
	"byeWeek":                     {"byeweek", "bye"},
	"providerId":                  {"providerid", "playerid", "id"},
	"passingTwoPointConversion":   {"passingtwoPointconversion", "passing2pt", "pass2pt"},
	"rushingTwoPointConversion":   {"rushingtwopointconversion", "rushing2pt", "rush2pt"},
	"receivingTwoPointConversion": {"receivingtwopointconversion", "receiving2pt", "rec2pt"},
	"fumbleLost":                  {"fumblelost", "fumbleslost", "fumlost"},
	"passing300YardGame":          {"passing300yardgame", "pass300games"},
	"passing400YardGame":          {"passing400yardgame", "pass400games"},
	"rushing100YardGame":          {"rushing100yardgame", "rush100games"},
	"rushing200YardGame":          {"rushing200yardgame", "rush200games"},
	"receiving100YardGame":        {"receiving100yardgame", "rec100games"},
	"receiving200YardGame":        {"receiving200yardgame", "rec200games"},
	"fieldGoal0To39":              {"fieldgoal0to39", "fg0to39", "fg039"},
	"fieldGoal40To49":             {"fieldgoal40to49", "fg40to49", "fg4049"},
	"fieldGoal50Plus":             {"fieldgoal50plus", "fg50plus", "fg50"},
	"defensePointsAllowed0":       {"defensepointsallowed0", "dstpa0"},
	"defensePointsAllowed1To6":    {"defensepointsallowed1to6", "dstpa1to6"},
	"defensePointsAllowed7To13":   {"defensepointsallowed7to13", "dstpa7to13"},
	"defensePointsAllowed14To20":  {"defensepointsallowed14to20", "dstpa14to20"},
	"defensePointsAllowed21To27":  {"defensepointsallowed21to27", "dstpa21to27"},
	"defensePointsAllowed28To34":  {"defensepointsallowed28to34", "dstpa28to34"},
	"defensePointsAllowed35Plus":  {"defensepointsallowed35plus", "dstpa35plus"},
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

func (service *ProjectionService) ImportCSV(ctx context.Context, name string, input io.Reader, mappings ...map[string]string) (projection.SourceStatus, error) {
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
		headers[normalizeCSVHeader(header)] = index
	}
	columnIndexes := projectionColumnIndexes(headers, firstMapping(mappings))
	for _, required := range []string{"name", "position", "team"} {
		if _, exists := columnIndexes[required]; !exists {
			return projection.SourceStatus{}, fmt.Errorf("projection CSV is missing %s", required)
		}
	}
	sourceID := slugify(name)
	records := make([]projection.Record, 0, len(rows)-1)
	seen := make(map[string]bool)
	for rowIndex, row := range rows[1:] {
		nameValue, position, team := value(row, columnIndexes["name"]), normalizePosition(value(row, columnIndexes["position"])), value(row, columnIndexes["team"])
		key := canonicalRankingKey(nameValue, position, team)
		if key == "" || seen[key] {
			continue
		}
		seen[key] = true
		record := projection.Record{SourceID: sourceID, PlayerKey: key, Name: nameValue, Position: position, Team: team, Stats: make(map[string]float64)}
		if index, exists := columnIndexes["providerId"]; exists {
			record.ProviderID = value(row, index)
		}
		if index, exists := columnIndexes["byeWeek"]; exists {
			record.ByeWeek, _ = strconv.Atoi(value(row, index))
		}
		if index, exists := columnIndexes["adp"]; exists {
			record.ADP, _ = strconv.ParseFloat(value(row, index), 64)
		}
		for _, column := range projectionStatColumns {
			if index, exists := columnIndexes[column]; exists {
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
	importedAt := time.Now().UTC()
	records, err = resolveProjectionPlayers(ctx, service.repository, records, importedAt)
	if err != nil {
		return projection.SourceStatus{}, err
	}
	status := projection.SourceStatus{ID: sourceID, Name: name, RecordCount: len(records), ImportedAt: importedAt}
	if err = service.repository.ReplaceProjections(ctx, status, records); err != nil {
		return projection.SourceStatus{}, err
	}
	return status, nil
}

func firstMapping(mappings []map[string]string) map[string]string {
	if len(mappings) == 0 || mappings[0] == nil {
		return map[string]string{}
	}
	return mappings[0]
}

func projectionColumnIndexes(headers map[string]int, mapping map[string]string) map[string]int {
	columns := append([]string{"name", "position", "team", "adp", "byeWeek", "providerId"}, projectionStatColumns...)
	indexes := make(map[string]int)
	for _, canonical := range columns {
		if sourceHeader, explicitlyMapped := mapping[canonical]; explicitlyMapped {
			if index, exists := headers[normalizeCSVHeader(sourceHeader)]; exists && sourceHeader != "" {
				indexes[canonical] = index
			}
			continue
		}
		aliases := projectionHeaderAliases[canonical]
		if len(aliases) == 0 {
			aliases = []string{canonical}
		}
		for _, alias := range aliases {
			if index, exists := headers[normalizeCSVHeader(alias)]; exists {
				indexes[canonical] = index
				break
			}
		}
	}
	return indexes
}

func normalizeCSVHeader(value string) string {
	return nonCSVHeaderCharacter.ReplaceAllString(strings.ToLower(strings.TrimSpace(value)), "")
}

func (service *ProjectionService) LeagueValues(ctx context.Context, scoring map[string]float64) (map[string]projection.LeagueValue, error) {
	records, err := service.repository.ProjectionRecords(ctx)
	if err != nil {
		return nil, err
	}
	aliases := map[string]string{}
	if repository, ok := service.repository.(interface {
		IdentityAliases(context.Context) (map[string]string, error)
	}); ok {
		aliases, err = repository.IdentityAliases(ctx)
		if err != nil {
			return nil, err
		}
	}
	totals := make(map[string]projection.LeagueValue)
	counts := make(map[string]int)
	adpCounts := make(map[string]int)
	canonicalRecords := make(map[string]projection.Record, len(records))
	canonicalPriority := make(map[string]bool, len(records))
	for _, record := range records {
		originalPlayerKey := record.PlayerKey
		record.PlayerKey = resolveIdentityAlias(record.PlayerKey, aliases)
		key := record.SourceID + "|" + record.PlayerKey
		isCanonical := originalPlayerKey == record.PlayerKey
		if _, exists := canonicalRecords[key]; !exists || (isCanonical && !canonicalPriority[key]) {
			canonicalRecords[key], canonicalPriority[key] = record, isCanonical
		}
	}
	for _, record := range canonicalRecords {
		points := projectionPoints(record, scoring)
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

func projectionPoints(record projection.Record, scoring map[string]float64) float64 {
	points := 0.0
	for stat, amount := range record.Stats {
		points += amount * scoring[stat]
	}
	if record.Position == "TE" {
		points += record.Stats["reception"] * scoring["tightEndReceptionBonus"]
	}
	return points
}

func value(row []string, index int) string {
	if index < 0 || index >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[index])
}
