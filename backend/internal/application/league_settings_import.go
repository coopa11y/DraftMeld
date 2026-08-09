package application

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/coopa11y/DraftMeld/backend/internal/domain/league"
)

type leagueSettingKind string

const (
	settingText         leagueSettingKind = "text"
	settingNumber       leagueSettingKind = "number"
	settingDraftType    leagueSettingKind = "draft-type"
	settingLeagueFormat leagueSettingKind = "league-format"
	settingBoolean      leagueSettingKind = "boolean"
)

type leagueSettingDefinition struct {
	Key     string
	Label   string
	Kind    leagueSettingKind
	Aliases []string
}

var leagueSettingDefinitions = []leagueSettingDefinition{
	{Key: "auctionMinimumBid", Label: "Auction minimum bid", Kind: settingNumber, Aliases: []string{"auction minimum bid", "minimum auction bid", "minimum bid"}},
	{Key: "auctionBudgetTrades", Label: "Auction-budget trading", Kind: settingBoolean, Aliases: []string{"auction budget trades", "auction dollar trades", "trade auction budget"}},
	{Key: "auctionBudget", Label: "Auction budget", Kind: settingNumber, Aliases: []string{"auction budget", "salary cap", "draft budget"}},
	{Key: "faabTrades", Label: "FAAB trading", Kind: settingBoolean, Aliases: []string{"faab trades", "faab trading", "trade faab"}},
	{Key: "faabBudget", Label: "FAAB budget", Kind: settingNumber, Aliases: []string{"faab budget", "free agent budget", "waiver budget"}},
	{Key: "futurePickSeasons", Label: "Future pick seasons", Kind: settingNumber, Aliases: []string{"future pick seasons", "future draft pick years", "future picks years"}},
	{Key: "rookieDraftRounds", Label: "Rookie draft rounds", Kind: settingNumber, Aliases: []string{"rookie draft rounds", "rookie rounds"}},
	{Key: "teamCount", Label: "Number of teams", Kind: settingNumber, Aliases: []string{"number of teams", "team count", "league size", "teams"}},
	{Key: "draftType", Label: "Draft format", Kind: settingDraftType, Aliases: []string{"draft format", "draft type"}},
	{Key: "leagueFormat", Label: "League format", Kind: settingLeagueFormat, Aliases: []string{"league format", "league type"}},
	{Key: "name", Label: "League name", Kind: settingText, Aliases: []string{"league name"}},
}

type rosterSlotDefinition struct {
	Name       string
	Positions  []string
	IsStarting bool
	Aliases    []string
}

var rosterSlotDefinitions = []rosterSlotDefinition{
	{Name: "SUPERFLEX", Positions: []string{"QB", "RB", "WR", "TE"}, IsStarting: true, Aliases: []string{"superflex", "super flex"}},
	{Name: "FLEX", Positions: []string{"RB", "WR", "TE"}, IsStarting: true, Aliases: []string{"flex", "rb wr te flex"}},
	{Name: "DST", Positions: []string{"DST"}, IsStarting: true, Aliases: []string{"d st", "dst", "team defense", "defense special teams"}},
	{Name: "QB", Positions: []string{"QB"}, IsStarting: true, Aliases: []string{"quarterbacks", "quarterback", "qb"}},
	{Name: "RB", Positions: []string{"RB"}, IsStarting: true, Aliases: []string{"running backs", "running back", "rb"}},
	{Name: "WR", Positions: []string{"WR"}, IsStarting: true, Aliases: []string{"wide receivers", "wide receiver", "wr"}},
	{Name: "TE", Positions: []string{"TE"}, IsStarting: true, Aliases: []string{"tight ends", "tight end", "te"}},
	{Name: "K", Positions: []string{"K"}, IsStarting: true, Aliases: []string{"kickers", "kicker", "pk", "k"}},
	{Name: "Bench", Positions: []string{"QB", "RB", "WR", "TE", "K", "DST"}, IsStarting: false, Aliases: []string{"bench slots", "bench players", "bench"}},
	{Name: "IR", Positions: []string{"QB", "RB", "WR", "TE", "K", "DST"}, IsStarting: false, Aliases: []string{"injured reserve", "reserve slots", "ir slots", "ir"}},
}

func parseLeagueSettingsCSV(result *LeagueRuleImportResult, rows [][]string) {
	builder := newLeagueSettingsBuilder(result)
	headers := normalizedHeaders(rows[0])
	settingColumn := firstHeader(headers, "setting", "rule", "name", "category")
	valueColumn := firstHeader(headers, "value", "settingvalue", "selection", "count")
	if settingColumn >= 0 && valueColumn >= 0 {
		for _, row := range rows[1:] {
			if settingColumn < len(row) && valueColumn < len(row) {
				builder.parsePair(row[settingColumn], row[valueColumn], strings.Join(row, ", "), "high")
			}
		}
		builder.finish()
		return
	}
	if len(rows) > 1 {
		for column, header := range rows[0] {
			if column < len(rows[1]) {
				builder.parsePair(header, rows[1][column], header+": "+rows[1][column], "high")
			}
		}
		if len(builder.seen) > 0 {
			builder.finish()
			return
		}
	}
	for _, row := range rows {
		if len(row) >= 2 {
			builder.parsePair(row[0], row[1], strings.Join(row, ", "), "medium")
		}
	}
	builder.finish()
}

func parseLeagueSettingsLines(result *LeagueRuleImportResult, lines []string) {
	builder := newLeagueSettingsBuilder(result)
	for index, line := range lines {
		value := ""
		if definition, alias, ok := findLeagueSetting(line); ok {
			value = cleanSettingValue(expressionAfterAlias(line, alias))
			if value == "" && index+1 < len(lines) {
				value = strings.TrimSpace(lines[index+1])
			}
			builder.addSetting(definition, value, strings.TrimSpace(line), "medium")
		}
		builder.parseRosterText(line, "medium")
	}
	if strings.Contains(strings.ToLower(strings.Join(lines, " ")), "individual defensive player") || strings.Contains(strings.ToLower(strings.Join(lines, " ")), "idp") {
		result.Warnings = append(result.Warnings, "Individual defensive-player roster slots are not supported and were not imported.")
	}
	builder.finish()
}

type leagueSettingsBuilder struct {
	result      *LeagueRuleImportResult
	seen        map[string]LeagueSettingMatch
	rosterSlots map[string]ImportedRosterSlot
}

func newLeagueSettingsBuilder(result *LeagueRuleImportResult) *leagueSettingsBuilder {
	return &leagueSettingsBuilder{result: result, seen: map[string]LeagueSettingMatch{}, rosterSlots: map[string]ImportedRosterSlot{}}
}

func (builder *leagueSettingsBuilder) parsePair(label, value, source, confidence string) {
	if definition, _, ok := findLeagueSetting(label); ok {
		builder.addSetting(definition, value, source, confidence)
	}
	if slot, ok := findRosterSlot(label); ok {
		if count, valid := integerValue(value); valid && count >= 0 && count <= 40 {
			builder.addRoster(slot, count, source, confidence)
		}
	}
}

func (builder *leagueSettingsBuilder) addSetting(definition leagueSettingDefinition, rawValue, source, confidence string) {
	rawValue = cleanSettingValue(rawValue)
	var display string
	switch definition.Kind {
	case settingText:
		if rawValue == "" || len(rawValue) > 80 {
			return
		}
		value := rawValue
		builder.result.Settings.Name = &value
		display = value
	case settingNumber:
		value, ok := numericValue(rawValue)
		if !ok || !builder.setNumber(definition.Key, value) {
			return
		}
		display = strconv.FormatFloat(value, 'f', -1, 64)
	case settingDraftType:
		value, ok := parseDraftType(rawValue)
		if !ok {
			return
		}
		builder.result.Settings.DraftType = &value
		display = string(value)
	case settingLeagueFormat:
		value, ok := parseLeagueFormat(rawValue)
		if !ok {
			return
		}
		builder.result.Settings.LeagueFormat = &value
		display = string(value)
	case settingBoolean:
		value, ok := booleanValue(rawValue)
		if !ok || !builder.setBoolean(definition.Key, value) {
			return
		}
		display = strconv.FormatBool(value)
	}
	builder.addMatch(LeagueSettingMatch{Key: definition.Key, Label: definition.Label, Value: display, Source: source, Confidence: confidence})
}

func (builder *leagueSettingsBuilder) setNumber(key string, value float64) bool {
	integer := int(value)
	switch key {
	case "teamCount":
		if value != float64(integer) || integer < 2 || integer > 32 {
			return false
		}
		builder.result.Settings.TeamCount = &integer
	case "futurePickSeasons":
		if value != float64(integer) || integer < 0 || integer > 5 {
			return false
		}
		builder.result.Settings.FuturePickSeasons = &integer
	case "rookieDraftRounds":
		if value != float64(integer) || integer < 1 || integer > 10 {
			return false
		}
		builder.result.Settings.RookieDraftRounds = &integer
	case "auctionBudget":
		if value <= 0 {
			return false
		}
		builder.result.Settings.AuctionBudget = &value
	case "auctionMinimumBid":
		if value <= 0 {
			return false
		}
		builder.result.Settings.AuctionMinimumBid = &value
	case "faabBudget":
		if value < 0 {
			return false
		}
		builder.result.Settings.FAABBudget = &value
	default:
		return false
	}
	return true
}

func (builder *leagueSettingsBuilder) setBoolean(key string, value bool) bool {
	switch key {
	case "auctionBudgetTrades":
		builder.result.Settings.AuctionBudgetTrades = &value
	case "faabTrades":
		builder.result.Settings.FAABTrades = &value
	default:
		return false
	}
	return true
}

func (builder *leagueSettingsBuilder) parseRosterText(line, confidence string) {
	normalized := normalizeRuleText(line)
	for _, definition := range rosterSlotDefinitions {
		for _, alias := range definition.Aliases {
			normalizedAlias := normalizeRuleText(alias)
			quoted := regexp.QuoteMeta(normalizedAlias)
			matcher := regexp.MustCompile(`(?:^|\s)(?:(\d+)\s+` + quoted + `|` + quoted + `\s+(\d+))(?:\s|$)`)
			match := matcher.FindStringSubmatch(normalized)
			if len(match) == 0 {
				continue
			}
			countText := match[1]
			if countText == "" {
				countText = match[2]
			}
			count, err := strconv.Atoi(countText)
			if err == nil && count <= 40 {
				builder.addRoster(definition, count, strings.TrimSpace(line), confidence)
			}
			break
		}
	}
}

func (builder *leagueSettingsBuilder) addRoster(definition rosterSlotDefinition, count int, source, confidence string) {
	slot := ImportedRosterSlot{Name: definition.Name, Count: count, Positions: append([]string(nil), definition.Positions...), IsStarting: definition.IsStarting}
	builder.rosterSlots[definition.Name] = slot
	builder.addMatch(LeagueSettingMatch{Key: "rosterSlots." + definition.Name, Label: definition.Name + " roster slots", Value: strconv.Itoa(count), Source: source, Confidence: confidence})
}

func (builder *leagueSettingsBuilder) addMatch(match LeagueSettingMatch) {
	if existing, ok := builder.seen[match.Key]; ok && existing.Value != match.Value {
		builder.result.Warnings = append(builder.result.Warnings, fmt.Sprintf("Conflicting values were found for %s; using %s.", match.Label, match.Value))
	}
	builder.seen[match.Key] = match
}

func (builder *leagueSettingsBuilder) finish() {
	keys := make([]string, 0, len(builder.seen))
	for key := range builder.seen {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		builder.result.SettingMatches = append(builder.result.SettingMatches, builder.seen[key])
	}
	rosterNames := make([]string, 0, len(builder.rosterSlots))
	for name := range builder.rosterSlots {
		rosterNames = append(rosterNames, name)
	}
	sort.Strings(rosterNames)
	for _, name := range rosterNames {
		builder.result.Settings.RosterSlots = append(builder.result.Settings.RosterSlots, builder.rosterSlots[name])
	}
	if len(builder.result.SettingMatches) > 0 && len(builder.result.Matches) == 0 {
		builder.result.Warnings = append(builder.result.Warnings, "Review every imported value against your league settings before saving.")
	}
}

func findLeagueSetting(value string) (leagueSettingDefinition, string, bool) {
	normalized := normalizeRuleText(value)
	bestIndex := len(normalized) + 1
	var bestDefinition leagueSettingDefinition
	bestAlias := ""
	for _, definition := range leagueSettingDefinitions {
		for _, alias := range definition.Aliases {
			normalizedAlias := normalizeRuleText(alias)
			index := strings.Index(" "+normalized+" ", " "+normalizedAlias+" ")
			if index >= 0 && index < bestIndex {
				bestIndex = index
				bestDefinition = definition
				bestAlias = normalizedAlias
			}
		}
	}
	return bestDefinition, bestAlias, bestAlias != ""
}

func findRosterSlot(value string) (rosterSlotDefinition, bool) {
	normalized := normalizeRuleText(value)
	for _, definition := range rosterSlotDefinitions {
		for _, alias := range definition.Aliases {
			if normalized == normalizeRuleText(alias) || strings.HasSuffix(normalized, " "+normalizeRuleText(alias)) || strings.HasPrefix(normalized, normalizeRuleText(alias)+" ") {
				return definition, true
			}
		}
	}
	return rosterSlotDefinition{}, false
}

func cleanSettingValue(value string) string {
	return strings.TrimSpace(strings.TrimLeft(value, ":=- "))
}

func integerValue(value string) (int, bool) {
	number, ok := numericValue(value)
	integer := int(number)
	return integer, ok && number == float64(integer)
}

func parseDraftType(value string) (league.DraftType, bool) {
	normalized := normalizeRuleText(value)
	switch {
	case strings.Contains(normalized, "auction"), strings.Contains(normalized, "salary cap"):
		return league.DraftTypeAuction, true
	case strings.Contains(normalized, "snake"), strings.Contains(normalized, "serpentine"):
		return league.DraftTypeSnake, true
	case strings.Contains(normalized, "linear"), strings.Contains(normalized, "straight"):
		return league.DraftTypeLinear, true
	default:
		return "", false
	}
}

func parseLeagueFormat(value string) (league.LeagueFormat, bool) {
	normalized := normalizeRuleText(value)
	switch {
	case strings.Contains(normalized, "dynasty"):
		return league.LeagueFormatDynasty, true
	case strings.Contains(normalized, "redraft"), strings.Contains(normalized, "re draft"):
		return league.LeagueFormatRedraft, true
	default:
		return "", false
	}
}

func booleanValue(value string) (bool, bool) {
	normalized := normalizeRuleText(value)
	switch normalized {
	case "yes", "true", "enabled", "allowed", "allow":
		return true, true
	case "no", "false", "disabled", "not allowed", "disallow":
		return false, true
	default:
		return false, false
	}
}
