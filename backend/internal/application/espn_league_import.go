package application

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/coopa11y/DraftMeld/backend/internal/domain/league"
)

const (
	espnAPIBase       = "https://lm-api-reads.fantasy.espn.com/apis/v3/games/ffl"
	espnResponseLimit = 5 << 20
)

var (
	ErrInvalidESPNLeagueURL = errors.New("enter an ESPN fantasy football league URL containing a leagueId")
	ErrESPNLeaguePrivate    = errors.New("ESPN did not make this league available publicly; use pasted settings or an ESPN PDF instead")
)

type ESPNLeagueImporter struct {
	client  *http.Client
	baseURL string
}

type espnLeaguePayload struct {
	Settings espnSettings `json:"settings"`
}

type espnSettings struct {
	Name                string                  `json:"name"`
	Size                int                     `json:"size"`
	DraftSettings       espnDraftSettings       `json:"draftSettings"`
	RosterSettings      espnRosterSettings      `json:"rosterSettings"`
	ScoringSettings     espnScoringSettings     `json:"scoringSettings"`
	AcquisitionSettings espnAcquisitionSettings `json:"acquisitionSettings"`
}

type espnDraftSettings struct {
	Type          int     `json:"type"`
	AuctionBudget float64 `json:"auctionBudget"`
}

type espnRosterSettings struct {
	LineupSlotCounts map[string]int `json:"lineupSlotCounts"`
}

type espnScoringSettings struct {
	ScoringItems []espnScoringItem `json:"scoringItems"`
}

type espnScoringItem struct {
	StatID          int                `json:"statId"`
	Points          float64            `json:"points"`
	PointsOverrides map[string]float64 `json:"pointsOverrides"`
}

type espnAcquisitionSettings struct {
	AcquisitionBudget float64 `json:"acquisitionBudget"`
}

type espnRosterMapping struct {
	definition rosterSlotDefinition
	supported  bool
}

var espnRosterSlots = map[int]espnRosterMapping{
	0:  {definition: rosterSlotDefinition{Name: "QB", Positions: []string{"QB"}, IsStarting: true}, supported: true},
	2:  {definition: rosterSlotDefinition{Name: "RB", Positions: []string{"RB"}, IsStarting: true}, supported: true},
	4:  {definition: rosterSlotDefinition{Name: "WR", Positions: []string{"WR"}, IsStarting: true}, supported: true},
	6:  {definition: rosterSlotDefinition{Name: "TE", Positions: []string{"TE"}, IsStarting: true}, supported: true},
	7:  {definition: rosterSlotDefinition{Name: "SUPERFLEX", Positions: []string{"QB", "RB", "WR", "TE"}, IsStarting: true}, supported: true},
	16: {definition: rosterSlotDefinition{Name: "DST", Positions: []string{"DST"}, IsStarting: true}, supported: true},
	17: {definition: rosterSlotDefinition{Name: "K", Positions: []string{"K"}, IsStarting: true}, supported: true},
	20: {definition: rosterSlotDefinition{Name: "Bench", Positions: []string{"QB", "RB", "WR", "TE", "K", "DST"}, IsStarting: false}, supported: true},
	21: {definition: rosterSlotDefinition{Name: "IR", Positions: []string{"QB", "RB", "WR", "TE", "K", "DST"}, IsStarting: false}, supported: true},
	23: {definition: rosterSlotDefinition{Name: "FLEX", Positions: []string{"RB", "WR", "TE"}, IsStarting: true}, supported: true},
}

var espnScoringRules = map[int]string{
	3: "passingYard", 4: "passingTouchdown", 17: "passing300YardGame", 18: "passing400YardGame",
	19: "passingTwoPointConversion", 20: "interception", 24: "rushingYard", 25: "rushingTouchdown",
	26: "rushingTwoPointConversion", 37: "rushing100YardGame", 38: "rushing200YardGame", 41: "reception",
	42: "receivingYard", 43: "receivingTouchdown", 44: "receivingTwoPointConversion", 53: "reception",
	56: "receiving100YardGame", 57: "receiving200YardGame", 68: "fumble", 72: "fumbleLost",
	77: "fieldGoal40To49", 80: "fieldGoal0To39", 83: "fieldGoalMade", 85: "fieldGoalMissed",
	86: "extraPointMade", 88: "extraPointMissed", 89: "defensePointsAllowed0", 90: "defensePointsAllowed1To6",
	91: "defensePointsAllowed7To13", 94: "defenseTouchdown", 95: "defenseInterception", 96: "defenseFumbleRecovery",
	97: "defenseBlockedKick", 98: "defenseSafety", 99: "defenseSack", 105: "defenseTouchdown",
	123: "defensePointsAllowed28To34", 201: "fieldGoal50Plus", 205: "defenseTwoPointReturn", 206: "defenseTwoPointReturn",
}

func NewESPNLeagueImporter(client *http.Client) *ESPNLeagueImporter {
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}
	return &ESPNLeagueImporter{client: client, baseURL: espnAPIBase}
}

func (importer *ESPNLeagueImporter) ImportLeague(ctx context.Context, leagueURL string, season int) (LeagueRuleImportResult, error) {
	leagueID, err := parseESPNLeagueID(leagueURL)
	if err != nil {
		return LeagueRuleImportResult{}, err
	}
	if season < 2000 || season > time.Now().Year()+1 {
		return LeagueRuleImportResult{}, errors.New("choose a valid ESPN season")
	}
	endpoint := fmt.Sprintf("%s/seasons/%d/segments/0/leagues/%s?view=mSettings", importer.baseURL, season, leagueID)
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return LeagueRuleImportResult{}, err
	}
	request.Header.Set("Accept", "application/json")
	response, err := importer.client.Do(request)
	if err != nil {
		return LeagueRuleImportResult{}, fmt.Errorf("contact ESPN: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode == http.StatusUnauthorized || response.StatusCode == http.StatusForbidden || response.StatusCode == http.StatusNotFound {
		return LeagueRuleImportResult{}, ErrESPNLeaguePrivate
	}
	if response.StatusCode != http.StatusOK {
		return LeagueRuleImportResult{}, fmt.Errorf("ESPN returned status %d", response.StatusCode)
	}
	contents, err := io.ReadAll(io.LimitReader(response.Body, espnResponseLimit+1))
	if err != nil || len(contents) > espnResponseLimit {
		return LeagueRuleImportResult{}, errors.New("ESPN returned an unreadable or unexpectedly large response")
	}
	return importer.ImportJSON(contents)
}

func (importer *ESPNLeagueImporter) ImportJSON(contents []byte) (LeagueRuleImportResult, error) {
	var payload espnLeaguePayload
	if err := json.Unmarshal(contents, &payload); err != nil {
		return LeagueRuleImportResult{}, fmt.Errorf("read ESPN settings JSON: %w", err)
	}
	ruleBuilder := newLeagueRuleResult("espn")
	unsupportedStats := importESPNScoring(ruleBuilder, payload.Settings.ScoringSettings.ScoringItems)
	result := ruleBuilder.finish()
	settings := newLeagueSettingsBuilder(&result)
	importESPNBasics(settings, payload.Settings)
	unsupportedSlots := importESPNRoster(settings, payload.Settings.RosterSettings.LineupSlotCounts)
	settings.finish()
	if len(result.Matches) == 0 && len(result.SettingMatches) == 0 {
		return LeagueRuleImportResult{}, ErrNoLeagueRulesFound
	}
	result.Warnings = append(result.Warnings, "ESPN does not publish a supported settings API. Review every imported value before saving.")
	if len(unsupportedSlots) > 0 {
		result.Warnings = append(result.Warnings, "Unsupported ESPN roster slot IDs were not imported: "+joinInts(unsupportedSlots)+".")
	}
	if len(unsupportedStats) > 0 {
		result.Warnings = append(result.Warnings, "Unsupported or ambiguous ESPN scoring stat IDs were not imported: "+joinInts(unsupportedStats)+".")
	}
	return result, nil
}

func importESPNBasics(builder *leagueSettingsBuilder, settings espnSettings) {
	if settings.Name != "" {
		builder.addSetting(leagueSettingByKey("name"), settings.Name, "settings.name", "high")
	}
	if settings.Size > 0 {
		builder.addSetting(leagueSettingByKey("teamCount"), strconv.Itoa(settings.Size), "settings.size", "high")
	}
	if settings.DraftSettings.Type == 2 {
		builder.addSetting(leagueSettingByKey("draftType"), string(league.DraftTypeAuction), "settings.draftSettings.type", "high")
	} else if settings.DraftSettings.Type == 1 {
		builder.addSetting(leagueSettingByKey("draftType"), string(league.DraftTypeSnake), "settings.draftSettings.type", "high")
	}
	if settings.DraftSettings.AuctionBudget > 0 {
		builder.addSetting(leagueSettingByKey("auctionBudget"), strconv.FormatFloat(settings.DraftSettings.AuctionBudget, 'f', -1, 64), "settings.draftSettings.auctionBudget", "high")
	}
	if settings.AcquisitionSettings.AcquisitionBudget > 0 {
		builder.addSetting(leagueSettingByKey("faabBudget"), strconv.FormatFloat(settings.AcquisitionSettings.AcquisitionBudget, 'f', -1, 64), "settings.acquisitionSettings.acquisitionBudget", "high")
	}
}

func importESPNRoster(builder *leagueSettingsBuilder, counts map[string]int) []int {
	unsupported := []int{}
	for rawID, count := range counts {
		id, err := strconv.Atoi(rawID)
		mapping, ok := espnRosterSlots[id]
		if err != nil || !ok || !mapping.supported {
			if count > 0 && err == nil {
				unsupported = append(unsupported, id)
			}
			continue
		}
		if count >= 0 && count <= 40 {
			builder.addRoster(mapping.definition, count, "settings.rosterSettings.lineupSlotCounts["+rawID+"]", "high")
		}
	}
	sort.Ints(unsupported)
	return unsupported
}

func importESPNScoring(builder *leagueRuleResultBuilder, items []espnScoringItem) []int {
	unsupported := []int{}
	for _, item := range items {
		key, ok := espnScoringRules[item.StatID]
		definition, found := leagueRuleByKey(key)
		if !ok || !found || len(item.PointsOverrides) > 0 {
			unsupported = append(unsupported, item.StatID)
			continue
		}
		builder.add(definition, item.Points, fmt.Sprintf("settings.scoringSettings.scoringItems[statId=%d]", item.StatID), "high")
	}
	sort.Ints(unsupported)
	return uniqueInts(unsupported)
}

func parseESPNLeagueID(value string) (string, error) {
	parsed, err := url.Parse(strings.TrimSpace(value))
	host := strings.ToLower(parsed.Hostname())
	if err != nil || parsed.Scheme != "https" || (host != "espn.com" && !strings.HasSuffix(host, ".espn.com")) {
		return "", ErrInvalidESPNLeagueURL
	}
	leagueID := parsed.Query().Get("leagueId")
	if leagueID == "" {
		return "", ErrInvalidESPNLeagueURL
	}
	if _, err := strconv.ParseUint(leagueID, 10, 64); err != nil {
		return "", ErrInvalidESPNLeagueURL
	}
	return leagueID, nil
}

func uniqueInts(values []int) []int {
	result := values[:0]
	for _, value := range values {
		if len(result) == 0 || result[len(result)-1] != value {
			result = append(result, value)
		}
	}
	return result
}

func joinInts(values []int) string {
	parts := make([]string, len(values))
	for index, value := range values {
		parts[index] = strconv.Itoa(value)
	}
	return strings.Join(parts, ", ")
}
