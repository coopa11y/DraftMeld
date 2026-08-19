package application

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/coopa11y/DraftMeld/backend/internal/document"
	"github.com/coopa11y/DraftMeld/backend/internal/domain/league"
)

var ErrNoLeagueRulesFound = errors.New("no supported league rules were found")

type LeagueRuleMatch struct {
	Key        string  `json:"key"`
	Label      string  `json:"label"`
	Value      float64 `json:"value"`
	Source     string  `json:"source"`
	Confidence string  `json:"confidence"`
}

type LeagueRuleImportResult struct {
	FileType       string                 `json:"fileType"`
	Rules          map[string]float64     `json:"rules"`
	Matches        []LeagueRuleMatch      `json:"matches"`
	Settings       ImportedLeagueSettings `json:"settings"`
	SettingMatches []LeagueSettingMatch   `json:"settingMatches"`
	Warnings       []string               `json:"warnings"`
}

type LeagueSettingMatch struct {
	Key        string `json:"key"`
	Label      string `json:"label"`
	Value      string `json:"value"`
	Source     string `json:"source"`
	Confidence string `json:"confidence"`
}

type ImportedRosterSlot struct {
	Name       string   `json:"name"`
	Count      int      `json:"count"`
	Positions  []string `json:"positions"`
	IsStarting bool     `json:"isStarting"`
}

type ImportedLeagueSettings struct {
	Name                *string              `json:"name,omitempty"`
	TeamCount           *int                 `json:"teamCount,omitempty"`
	DraftType           *league.DraftType    `json:"draftType,omitempty"`
	LeagueFormat        *league.LeagueFormat `json:"leagueFormat,omitempty"`
	FuturePickSeasons   *int                 `json:"futurePickSeasons,omitempty"`
	RookieDraftRounds   *int                 `json:"rookieDraftRounds,omitempty"`
	AuctionBudget       *float64             `json:"auctionBudget,omitempty"`
	AuctionMinimumBid   *float64             `json:"auctionMinimumBid,omitempty"`
	AuctionBudgetTrades *bool                `json:"auctionBudgetTrades,omitempty"`
	FAABBudget          *float64             `json:"faabBudget,omitempty"`
	FAABTrades          *bool                `json:"faabTrades,omitempty"`
	RosterSlots         []ImportedRosterSlot `json:"rosterSlots,omitempty"`
}

type leagueRuleDefinition struct {
	Key     string
	Label   string
	Aliases []string
}

var leagueRuleDefinitions = []leagueRuleDefinition{
	{Key: "tightEndReceptionBonus", Label: "Tight end reception bonus", Aliases: []string{"tight end reception bonus", "te reception bonus", "te premium"}},
	{Key: "passingTwoPointConversion", Label: "Passing two-point conversion", Aliases: []string{"passing two point conversion", "passing 2 point conversion", "pass 2pt"}},
	{Key: "rushingTwoPointConversion", Label: "Rushing two-point conversion", Aliases: []string{"rushing two point conversion", "rushing 2 point conversion", "rush 2pt"}},
	{Key: "receivingTwoPointConversion", Label: "Receiving two-point conversion", Aliases: []string{"receiving two point conversion", "receiving 2 point conversion", "receiving 2pt"}},
	{Key: "passing300YardGame", Label: "300-yard passing game", Aliases: []string{"300 yard passing game", "passing 300 yard game", "300 passing yards"}},
	{Key: "passing400YardGame", Label: "400-yard passing game", Aliases: []string{"400 yard passing game", "passing 400 yard game", "400 passing yards"}},
	{Key: "rushing100YardGame", Label: "100-yard rushing game", Aliases: []string{"100 yard rushing game", "rushing 100 yard game", "100 rushing yards"}},
	{Key: "rushing200YardGame", Label: "200-yard rushing game", Aliases: []string{"200 yard rushing game", "rushing 200 yard game", "200 rushing yards"}},
	{Key: "receiving100YardGame", Label: "100-yard receiving game", Aliases: []string{"100 yard receiving game", "receiving 100 yard game", "100 receiving yards"}},
	{Key: "receiving200YardGame", Label: "200-yard receiving game", Aliases: []string{"200 yard receiving game", "receiving 200 yard game", "200 receiving yards"}},
	{Key: "fieldGoal0To39", Label: "Field goal made, 0-39 yards", Aliases: []string{"field goal made 0 39 yards", "field goals made 0 39 yards", "fg made 0 39", "fg 0 39"}},
	{Key: "fieldGoal40To49", Label: "Field goal made, 40-49 yards", Aliases: []string{"field goal made 40 49 yards", "field goals made 40 49 yards", "fg made 40 49", "fg 40 49"}},
	{Key: "fieldGoal50Plus", Label: "Field goal made, 50+ yards", Aliases: []string{"field goal made 50 yards", "field goals made 50 yards", "field goal 50 plus", "fg made 50", "fg 50 plus"}},
	{Key: "defensePointsAllowed1To6", Label: "Game allowing 1-6 points", Aliases: []string{"points allowed 1 6", "1 6 points allowed"}},
	{Key: "defensePointsAllowed7To13", Label: "Game allowing 7-13 points", Aliases: []string{"points allowed 7 13", "7 13 points allowed"}},
	{Key: "defensePointsAllowed14To20", Label: "Game allowing 14-20 points", Aliases: []string{"points allowed 14 20", "14 20 points allowed"}},
	{Key: "defensePointsAllowed21To27", Label: "Game allowing 21-27 points", Aliases: []string{"points allowed 21 27", "21 27 points allowed"}},
	{Key: "defensePointsAllowed28To34", Label: "Game allowing 28-34 points", Aliases: []string{"points allowed 28 34", "28 34 points allowed"}},
	{Key: "defensePointsAllowed35Plus", Label: "Game allowing 35+ points", Aliases: []string{"points allowed 35 plus", "35 plus points allowed", "points allowed 35"}},
	{Key: "defensePointsAllowed0", Label: "Game allowing 0 points", Aliases: []string{"points allowed 0", "0 points allowed", "shutout"}},
	{Key: "defenseFumbleRecovery", Label: "Defense fumble recovery", Aliases: []string{"defensive fumble recovery", "defense fumble recovery", "fumble recovery"}},
	{Key: "defenseInterception", Label: "Defense interception", Aliases: []string{"defensive interception", "defense interception", "interception return"}},
	{Key: "defenseTouchdown", Label: "Defense or special-teams touchdown", Aliases: []string{"defensive touchdown", "defense touchdown", "special teams touchdown", "return touchdown"}},
	{Key: "defenseTwoPointReturn", Label: "Defensive two-point return", Aliases: []string{"defensive two point return", "defense two point return", "defensive 2 point return"}},
	{Key: "defenseBlockedKick", Label: "Blocked kick", Aliases: []string{"blocked kick", "blocked punt", "blocked field goal"}},
	{Key: "fieldGoalMissed", Label: "Field goal missed", Aliases: []string{"field goal missed", "missed field goal", "fg missed"}},
	{Key: "extraPointMissed", Label: "Extra point missed", Aliases: []string{"extra point missed", "missed extra point", "pat missed"}},
	{Key: "passingTouchdown", Label: "Passing touchdown", Aliases: []string{"passing touchdown", "passing touchdowns", "passing td", "pass td"}},
	{Key: "rushingTouchdown", Label: "Rushing touchdown", Aliases: []string{"rushing touchdown", "rushing touchdowns", "rush td"}},
	{Key: "receivingTouchdown", Label: "Receiving touchdown", Aliases: []string{"receiving touchdown", "receiving touchdowns", "receive td"}},
	{Key: "interception", Label: "Interception thrown", Aliases: []string{"interception thrown", "interceptions thrown", "pass intercepted", "passing interception"}},
	{Key: "fumbleLost", Label: "Fumble lost", Aliases: []string{"fumble lost", "fumbles lost"}},
	{Key: "fieldGoalMade", Label: "Field goal made", Aliases: []string{"field goal made", "field goals made", "fg made"}},
	{Key: "extraPointMade", Label: "Extra point made", Aliases: []string{"extra point made", "extra points made", "pat made"}},
	{Key: "defenseSack", Label: "Sack", Aliases: []string{"defensive sack", "defense sack", "sacks"}},
	{Key: "defenseSafety", Label: "Safety", Aliases: []string{"defensive safety", "defense safety", "safety"}},
	{Key: "passingYard", Label: "Passing yard", Aliases: []string{"passing yards", "passing yard", "pass yards", "pass yds"}},
	{Key: "rushingYard", Label: "Rushing yard", Aliases: []string{"rushing yards", "rushing yard", "rush yards", "rush yds"}},
	{Key: "receivingYard", Label: "Receiving yard", Aliases: []string{"receiving yards", "receiving yard", "receive yards", "rec yds"}},
	{Key: "reception", Label: "Reception", Aliases: []string{"receptions", "reception", "points per reception"}},
	{Key: "fumble", Label: "Fumble", Aliases: []string{"fumbles", "fumble"}},
}

var (
	ruleSeparators = regexp.MustCompile(`[^a-z0-9]+`)
	decimalNumber  = regexp.MustCompile(`-?\d+(?:\.\d+)?`)
	perExpression  = regexp.MustCompile(`(?i)(-?\d+(?:\.\d+)?)\s*(?:points?|pts?)?\s*(?:per|for every|every|for each)\s*(\d+(?:\.\d+)?)`)
)

type LeagueRuleImportService struct {
	pdfExtractor document.PDFExtractor
	espnImporter *ESPNLeagueImporter
}

func NewLeagueRuleImportService() *LeagueRuleImportService {
	return &LeagueRuleImportService{
		pdfExtractor: document.NewPDFExtractor(),
		espnImporter: NewESPNLeagueImporter(nil),
	}
}

func NewLeagueRuleImportServiceWithExtractor(extractor document.PDFExtractor) *LeagueRuleImportService {
	return &LeagueRuleImportService{pdfExtractor: extractor, espnImporter: NewESPNLeagueImporter(nil)}
}

func NewLeagueRuleImportServiceWithESPNClient(client *http.Client) *LeagueRuleImportService {
	return &LeagueRuleImportService{pdfExtractor: document.NewPDFExtractor(), espnImporter: NewESPNLeagueImporter(client)}
}

func (service *LeagueRuleImportService) ImportESPNLeague(ctx context.Context, leagueURL string, season int) (LeagueRuleImportResult, error) {
	return service.espnImporter.ImportLeague(ctx, leagueURL, season)
}

func (service *LeagueRuleImportService) ImportESPNJSON(contents []byte) (LeagueRuleImportResult, error) {
	return service.espnImporter.ImportJSON(contents)
}

func (service *LeagueRuleImportService) ImportESPNText(contents []byte) (LeagueRuleImportResult, error) {
	result := parseLeagueRuleLines(strings.Split(string(contents), "\n"), "espn-text")
	parseLeagueSettingsLines(&result, strings.Split(string(contents), "\n"))
	if len(result.Matches) == 0 && len(result.SettingMatches) == 0 {
		return LeagueRuleImportResult{}, ErrNoLeagueRulesFound
	}
	result.Warnings = append(result.Warnings, "Pasted ESPN text can omit settings that are collapsed or off screen. Compare the review with ESPN before saving.")
	return result, nil
}

func (service *LeagueRuleImportService) ImportESPNPDF(contents []byte) (LeagueRuleImportResult, error) {
	result, err := service.ImportPDF(contents)
	if err != nil {
		return LeagueRuleImportResult{}, err
	}
	result.FileType = "espn-pdf"
	result.Warnings = append(result.Warnings, "ESPN print layouts can omit collapsed settings. Compare the review with ESPN before saving.")
	return result, nil
}

func (service *LeagueRuleImportService) ImportPDF(contents []byte) (LeagueRuleImportResult, error) {
	extracted, err := service.pdfExtractor.Extract(contents)
	if err != nil {
		return LeagueRuleImportResult{}, err
	}
	result := parseLeagueRuleLines(strings.Split(extracted.Text, "\n"), "pdf")
	parseLeagueSettingsLines(&result, strings.Split(extracted.Text, "\n"))
	if len(result.Matches) == 0 && len(result.SettingMatches) == 0 {
		return LeagueRuleImportResult{}, ErrNoLeagueRulesFound
	}
	if extracted.PageCount > 1 {
		result.Warnings = append(result.Warnings, fmt.Sprintf("Reviewed text from %d PDF pages.", extracted.PageCount))
	}
	if extracted.OCRApplied {
		result.Warnings = append(result.Warnings, "Scanned PDF text was recognized locally with OCR. Verify every imported value carefully.")
	}
	return result, nil
}

func (service *LeagueRuleImportService) ImportCSV(input io.Reader) (LeagueRuleImportResult, error) {
	reader := csv.NewReader(input)
	reader.FieldsPerRecord = -1
	rows, err := reader.ReadAll()
	if err != nil {
		return LeagueRuleImportResult{}, fmt.Errorf("read league rules CSV: %w", err)
	}
	if len(rows) == 0 {
		return LeagueRuleImportResult{}, ErrNoLeagueRulesFound
	}
	result := parseLeagueRuleCSV(rows)
	parseLeagueSettingsCSV(&result, rows)
	if len(result.Matches) == 0 && len(result.SettingMatches) == 0 {
		return LeagueRuleImportResult{}, ErrNoLeagueRulesFound
	}
	return result, nil
}

func parseLeagueRuleCSV(rows [][]string) LeagueRuleImportResult {
	result := newLeagueRuleResult("csv")
	headers := normalizedHeaders(rows[0])
	ruleColumn := firstHeader(headers, "rule", "stat", "statistic", "category", "setting", "name")
	valueColumn := firstHeader(headers, "points", "pointvalue", "value", "score")
	perColumn := firstHeader(headers, "per", "every", "unit", "yardsperpoint")
	if ruleColumn >= 0 && valueColumn >= 0 {
		for _, row := range rows[1:] {
			if ruleColumn >= len(row) || valueColumn >= len(row) {
				continue
			}
			definition, _, ok := findLeagueRule(row[ruleColumn])
			if !ok {
				continue
			}
			value, valid := numericValue(row[valueColumn])
			if !valid {
				continue
			}
			if perColumn >= 0 && perColumn < len(row) {
				if divisor, divisorOK := numericValue(row[perColumn]); divisorOK && divisor != 0 {
					value /= divisor
				}
			}
			result.add(definition, value, strings.Join(row, ", "), "high")
		}
		return result.finish()
	}

	if len(rows) > 1 {
		for column, header := range rows[0] {
			definition, _, ok := findLeagueRule(header)
			if !ok || column >= len(rows[1]) {
				continue
			}
			if value, valid := scoringValue(rows[1][column]); valid {
				result.add(definition, value, header+": "+rows[1][column], "high")
			}
		}
		if len(result.seen) > 0 {
			return result.finish()
		}
	}
	for _, row := range rows {
		if len(row) < 2 {
			continue
		}
		definition, _, ok := findLeagueRule(row[0])
		if !ok {
			continue
		}
		if value, valid := scoringValue(strings.Join(row[1:], " per ")); valid {
			result.add(definition, value, strings.Join(row, ", "), "medium")
		}
	}
	return result.finish()
}

func parseLeagueRuleLines(lines []string, fileType string) LeagueRuleImportResult {
	result := newLeagueRuleResult(fileType)
	for index, line := range lines {
		definition, alias, ok := findLeagueRule(line)
		if !ok {
			continue
		}
		remainder := expressionAfterAlias(line, alias)
		value, valid := scoringValue(remainder)
		source := strings.TrimSpace(line)
		if !valid && index+1 < len(lines) {
			nextLine := strings.TrimSpace(lines[index+1])
			_, _, nextIsRule := findLeagueRule(nextLine)
			if !nextIsRule {
				value, valid = scoringValue(nextLine)
				if valid {
					remainder = nextLine
					source += " " + nextLine
				}
			}
		}
		if !valid {
			continue
		}
		confidence := "medium"
		if perExpression.MatchString(remainder) {
			confidence = "high"
		}
		result.add(definition, value, source, confidence)
	}
	return result.finish()
}

func expressionAfterAlias(line, normalizedAlias string) string {
	tokens := strings.Fields(normalizedAlias)
	parts := make([]string, 0, len(tokens))
	for _, token := range tokens {
		parts = append(parts, regexp.QuoteMeta(token))
	}
	matcher := regexp.MustCompile(`(?i)` + strings.Join(parts, `[^a-z0-9]+`))
	location := matcher.FindStringIndex(line)
	if location == nil {
		return line
	}
	return strings.TrimSpace(line[location[1]:])
}

type leagueRuleResultBuilder struct {
	result LeagueRuleImportResult
	seen   map[string]LeagueRuleMatch
}

func newLeagueRuleResult(fileType string) *leagueRuleResultBuilder {
	return &leagueRuleResultBuilder{result: LeagueRuleImportResult{FileType: fileType, Rules: map[string]float64{}, Warnings: []string{}}, seen: map[string]LeagueRuleMatch{}}
}

func (builder *leagueRuleResultBuilder) add(definition leagueRuleDefinition, value float64, source, confidence string) {
	match := LeagueRuleMatch{Key: definition.Key, Label: definition.Label, Value: value, Source: source, Confidence: confidence}
	if existing, ok := builder.seen[definition.Key]; ok && existing.Value != value {
		builder.result.Warnings = append(builder.result.Warnings, fmt.Sprintf("Conflicting values were found for %s; using %g.", definition.Label, value))
	}
	builder.seen[definition.Key] = match
	builder.result.Rules[definition.Key] = value
}

func (builder *leagueRuleResultBuilder) finish() LeagueRuleImportResult {
	keys := make([]string, 0, len(builder.seen))
	for key := range builder.seen {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		builder.result.Matches = append(builder.result.Matches, builder.seen[key])
	}
	if len(builder.result.Matches) > 0 || len(builder.result.SettingMatches) > 0 {
		builder.result.Warnings = append(builder.result.Warnings, "Review every imported value against your league settings before saving.")
	}
	return builder.result
}

func findLeagueRule(value string) (leagueRuleDefinition, string, bool) {
	normalized := normalizeRuleText(value)
	for _, definition := range leagueRuleDefinitions {
		for _, alias := range definition.Aliases {
			normalizedAlias := normalizeRuleText(alias)
			if strings.Contains(normalized, normalizedAlias) {
				return definition, normalizedAlias, true
			}
		}
	}
	return leagueRuleDefinition{}, "", false
}

func leagueRuleByKey(key string) (leagueRuleDefinition, bool) {
	for _, definition := range leagueRuleDefinitions {
		if definition.Key == key {
			return definition, true
		}
	}
	return leagueRuleDefinition{}, false
}

func normalizeRuleText(value string) string {
	return strings.TrimSpace(ruleSeparators.ReplaceAllString(strings.ToLower(value), " "))
}

func scoringValue(value string) (float64, bool) {
	if match := perExpression.FindStringSubmatch(value); len(match) == 3 {
		points, pointsOK := numericValue(match[1])
		divisor, divisorOK := numericValue(match[2])
		if pointsOK && divisorOK && divisor != 0 {
			return points / divisor, true
		}
	}
	return numericValue(value)
}

func numericValue(value string) (float64, bool) {
	match := decimalNumber.FindString(value)
	if match == "" {
		return 0, false
	}
	parsed, err := strconv.ParseFloat(match, 64)
	return parsed, err == nil
}

func normalizedHeaders(row []string) map[string]int {
	headers := make(map[string]int, len(row))
	for index, header := range row {
		headers[strings.ReplaceAll(normalizeRuleText(header), " ", "")] = index
	}
	return headers
}

func firstHeader(headers map[string]int, names ...string) int {
	for _, name := range names {
		if index, ok := headers[name]; ok {
			return index
		}
	}
	return -1
}
