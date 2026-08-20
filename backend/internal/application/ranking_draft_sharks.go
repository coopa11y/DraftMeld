package application

import (
	"fmt"
	"html"
	"io"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/coopa11y/DraftMeld/backend/internal/domain/ranking"
)

const (
	draftSharksProjectURL = "https://www.draftsharks.com/rankings"
	draftSharksTableURL   = "https://www.draftsharks.com/rankings/load-table"
)

type draftSharksVariant struct {
	id, name, slug, profile string
}

var draftSharksPresetVariants = []draftSharksVariant{
	{id: "draft-sharks-standard-1qb", name: "Draft Sharks - Standard, 1QB", profile: "Standard, 1QB"},
	{id: "draft-sharks-half-ppr-1qb", name: "Draft Sharks - Half-PPR, 1QB", slug: "half-ppr", profile: "Half-PPR, 1QB"},
	{id: "draft-sharks-ppr-1qb", name: "Draft Sharks - PPR, 1QB", slug: "ppr", profile: "PPR, 1QB"},
	{id: "draft-sharks-tep-1qb", name: "Draft Sharks - TE Premium, 1QB", slug: "te-premium", profile: "TE Premium, 1QB"},
	{id: "draft-sharks-standard-superflex", name: "Draft Sharks - Standard, Superflex", slug: "superflex", profile: "Standard, Superflex"},
	{id: "draft-sharks-half-ppr-superflex", name: "Draft Sharks - Half-PPR, Superflex", slug: "half-ppr-superflex", profile: "Half-PPR, Superflex"},
	{id: "draft-sharks-ppr-superflex", name: "Draft Sharks - PPR, Superflex", slug: "ppr-superflex", profile: "PPR, Superflex"},
	{id: "draft-sharks-tep-superflex", name: "Draft Sharks - TE Premium, Superflex", slug: "te-premium-superflex", profile: "TE Premium, Superflex"},
}

func draftSharksSources() []ranking.SourceDefinition {
	sources := make([]ranking.SourceDefinition, 0, len(draftSharksPresetVariants))
	for _, variant := range draftSharksPresetVariants {
		parameters := url.Values{
			"pprSuperflexSlug": {variant.slug}, "fantasyPosition": {""}, "researchDepth": {"rankings"},
			"playerGroup": {"all"}, "sort": {""}, "selectedTeam": {""}, "playerSearchTerm": {""},
		}
		sources = append(sources, ranking.SourceDefinition{
			ID: variant.id, Name: variant.name, Profile: variant.profile, VariantGroup: "draft-sharks",
			Description: "Tiered Draft Sharks rankings with ADP, projection range, injury risk, schedule strength, and 3D Value.",
			Methodology: "Draft Sharks 3D Value ranking for the selected public scoring and quarterback preset",
			License:     "Proprietary; retrieved on demand and not redistributed", ProjectURL: draftSharksProjectURL,
			DataURL: draftSharksTableURL + "?" + parameters.Encode(), DefaultWeight: 0.9,
			ImportMode: "download", Role: "ranking",
		})
	}
	return sources
}

var (
	draftSharksRowPattern  = regexp.MustCompile(`(?s)<tbody\s+data-player-row(.*?)</tbody>`)
	draftSharksRankPattern = regexp.MustCompile(`(?s)rank-index[^>]*>\s*<span>([0-9]+)</span>`)
	draftSharksTeamPattern = regexp.MustCompile(`alt="([A-Z]{2,3}) logo"`)
	draftSharksCellPattern = regexp.MustCompile(`(?s)<td\b([^>]*)>.*?</td>`)
	draftNotationPattern   = regexp.MustCompile(`^([0-9]+)\.([0-9]{2})$`)
	htmlAttributePattern   = regexp.MustCompile(`([A-Za-z0-9_.:-]+)="([^"]*)"`)
)

func parseDraftSharks(input io.Reader, sourceID string) ([]ranking.Record, string, error) {
	contents, err := io.ReadAll(input)
	if err != nil {
		return nil, "", fmt.Errorf("read Draft Sharks rankings: %w", err)
	}
	profile := draftSharksProfile(sourceID)
	records := make([]ranking.Record, 0, 250)
	seenRanks := make(map[int]bool)
	for _, match := range draftSharksRowPattern.FindAllSubmatch(contents, -1) {
		attributes, body := string(match[1]), string(match[0])
		rankMatch := draftSharksRankPattern.FindStringSubmatch(body)
		if len(rankMatch) != 2 {
			continue
		}
		rankValue, _ := strconv.Atoi(rankMatch[1])
		position := normalizePosition(htmlAttribute(attributes, "data-fantasy-position"))
		name := strings.TrimSpace(html.UnescapeString(htmlAttribute(attributes, "data-player-name")))
		if rankValue < 1 || seenRanks[rankValue] || name == "" || !supportedPosition(position) {
			continue
		}
		seenRanks[rankValue] = true
		team := ""
		if teamMatch := draftSharksTeamPattern.FindStringSubmatch(body); len(teamMatch) == 2 {
			team = canonicalNFLTeam(teamMatch[1])
		}
		values := draftSharksCellValues(body)
		record := ranking.Record{
			SourceID: sourceID, ProviderID: htmlAttribute(attributes, "data-key"), Name: name,
			Position: position, Team: team, Rank: rankValue,
			Tier:                positiveInteger(htmlAttribute(attributes, "data-tier-overall")),
			ADP:                 draftNotationToOverall(values["adp"]),
			Games:               positiveInteger(values["games_played"]),
			ByeWeek:             positiveInteger(values["player.team.bye"]),
			FloorProjection:     draftSharksNumber(values["floor_points"]),
			ConsensusProjection: draftSharksNumber(values["consensus_projection"]),
			SourceProjection:    draftSharksNumber(values["fantasy_points"]),
			CeilingProjection:   draftSharksNumber(values["ceiling_points"]),
			SourceValue:         draftSharksNumber(values["dsValue"]),
			InjuryRisk:          percentageValue(values["player.sipPlayerProfile.injury_prob"]),
			ScheduleStrength:    percentageValue(values["strength_of_schedule"]),
		}
		records = append(records, canonicalizeRankingRecord(record))
	}
	sort.Slice(records, func(left, right int) bool { return records[left].Rank < records[right].Rank })
	if len(records) < 100 {
		return nil, "", fmt.Errorf("parse Draft Sharks rankings: found only %d usable players", len(records))
	}
	return records, "Current Draft Sharks " + profile + " preset", nil
}

func draftSharksCellValues(body string) map[string]string {
	values := make(map[string]string)
	for _, cell := range draftSharksCellPattern.FindAllStringSubmatch(body, -1) {
		attribute, value := htmlAttribute(cell[1], "data-attribute"), htmlAttribute(cell[1], "data-value")
		if attribute != "" {
			values[attribute] = value
		}
	}
	return values
}

func htmlAttribute(attributes, name string) string {
	for _, match := range htmlAttributePattern.FindAllStringSubmatch(attributes, -1) {
		if match[1] == name {
			return html.UnescapeString(match[2])
		}
	}
	return ""
}

func draftSharksProfile(sourceID string) string {
	for _, variant := range draftSharksPresetVariants {
		if variant.id == sourceID {
			return variant.profile
		}
	}
	return "public"
}

func positiveInteger(value string) int {
	parsed, _ := strconv.Atoi(strings.TrimSpace(value))
	return max(parsed, 0)
}

func draftSharksNumber(value string) float64 {
	parsed, _ := strconv.ParseFloat(strings.TrimSpace(strings.TrimSuffix(value, "%")), 64)
	return parsed
}

func percentageValue(value string) float64 { return draftSharksNumber(value) }

func draftNotationToOverall(value string) float64 {
	match := draftNotationPattern.FindStringSubmatch(strings.TrimSpace(value))
	if len(match) != 3 {
		return draftSharksNumber(value)
	}
	round, _ := strconv.Atoi(match[1])
	pick, _ := strconv.Atoi(match[2])
	if round < 1 || pick < 1 || pick > 12 {
		return 0
	}
	return float64((round-1)*12 + pick)
}
