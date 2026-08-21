package application

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/coopa11y/DraftMeld/backend/internal/domain/league"
	"github.com/coopa11y/DraftMeld/backend/internal/domain/ranking"
)

type sleeperADPVariant struct {
	id, name, profile, statKey string
	format                     league.LeagueFormat
	scoring                    string
	superflex                  bool
}

var sleeperADPVariants = buildSleeperADPVariants()

func buildSleeperADPVariants() []sleeperADPVariant {
	variants := make([]sleeperADPVariant, 0, 12)
	for _, format := range []league.LeagueFormat{league.LeagueFormatRedraft, league.LeagueFormatDynasty} {
		for _, scoring := range []string{"standard", "half-ppr", "ppr"} {
			for _, superflex := range []bool{false, true} {
				formatName := map[league.LeagueFormat]string{league.LeagueFormatRedraft: "Redraft", league.LeagueFormatDynasty: "Dynasty"}[format]
				scoringName := map[string]string{"standard": "Standard", "half-ppr": "Half-PPR", "ppr": "PPR"}[scoring]
				quarterbacks := map[bool]string{false: "1QB", true: "Superflex"}[superflex]
				id := fmt.Sprintf("sleeper-adp-%s-%s-%s", format, scoring, map[bool]string{false: "1qb", true: "superflex"}[superflex])
				variants = append(variants, sleeperADPVariant{
					id: id, name: fmt.Sprintf("Sleeper ADP - %s %s, %s", formatName, scoringName, quarterbacks),
					profile: fmt.Sprintf("%s %s, %s", formatName, scoringName, quarterbacks),
					statKey: sleeperADPStatKey(format, scoring, superflex), format: format, scoring: scoring, superflex: superflex,
				})
			}
		}
	}
	return variants
}

func sleeperADPStatKey(format league.LeagueFormat, scoring string, superflex bool) string {
	if format == league.LeagueFormatDynasty {
		if superflex {
			return "adp_dynasty_2qb"
		}
		return map[string]string{"standard": "adp_dynasty_std", "half-ppr": "adp_dynasty_half_ppr", "ppr": "adp_dynasty_ppr"}[scoring]
	}
	if superflex {
		return "adp_2qb"
	}
	return map[string]string{"standard": "adp_std", "half-ppr": "adp_half_ppr", "ppr": "adp_ppr"}[scoring]
}

func sleeperADPSources() []ranking.SourceDefinition {
	sources := make([]ranking.SourceDefinition, 0, len(sleeperADPVariants))
	for _, variant := range sleeperADPVariants {
		sources = append(sources, ranking.SourceDefinition{
			ID: variant.id, Name: variant.name, Profile: variant.profile, VariantGroup: "sleeper-adp",
			Description: "Sleeper community average draft position for the active league's closest public format.",
			Methodology: "Observed Sleeper draft position ordered from earliest to latest selection",
			License:     "Public undocumented feed; noncommercial use only unless licensed by Sleeper",
			ProjectURL:  sleeperProjectURL, DataURL: sleeperProjectionURL, DefaultWeight: 0.5,
			ImportMode: "download", Role: "market",
		})
	}
	return sources
}

func parseSleeperADP(input io.Reader, sourceID string) ([]ranking.Record, string, error) {
	items, err := decodeSleeperFeed(input)
	if err != nil {
		return nil, "", err
	}
	variant, exists := sleeperADPVariantByID(sourceID)
	if !exists {
		return nil, "", fmt.Errorf("unknown Sleeper ADP profile %q", sourceID)
	}
	type candidate struct {
		item sleeperFeedItem
		adp  float64
	}
	candidates := make([]candidate, 0, 500)
	for _, item := range items {
		position := normalizePosition(item.Player.Position)
		adp := item.Stats[variant.statKey]
		if !supportedPosition(position) || adp <= 0 || adp >= 999 {
			continue
		}
		candidates = append(candidates, candidate{item: item, adp: adp})
	}
	sort.Slice(candidates, func(left, right int) bool {
		if candidates[left].adp == candidates[right].adp {
			return candidates[left].item.PlayerID < candidates[right].item.PlayerID
		}
		return candidates[left].adp < candidates[right].adp
	})
	records := make([]ranking.Record, 0, len(candidates))
	for _, candidate := range candidates {
		item := candidate.item
		name := strings.TrimSpace(item.Player.FirstName + " " + item.Player.LastName)
		if name == "" || item.PlayerID == "" {
			continue
		}
		records = append(records, canonicalizeRankingRecord(ranking.Record{
			SourceID: sourceID, ProviderID: item.PlayerID, Name: name,
			Position: normalizePosition(item.Player.Position), Team: canonicalNFLTeam(item.Player.Team),
			Rank: len(records) + 1, ADP: candidate.adp,
		}))
	}
	if len(records) < 100 {
		return nil, "", fmt.Errorf("Sleeper ADP feed contained only %d usable players", len(records))
	}
	return records, sleeperPublishedAt(items), nil
}

func sleeperADPVariantByID(sourceID string) (sleeperADPVariant, bool) {
	for _, variant := range sleeperADPVariants {
		if variant.id == sourceID {
			return variant, true
		}
	}
	return sleeperADPVariant{}, false
}

func matchingSleeperADPSourceID(rules league.Rules, superflex bool) string {
	scoring := "standard"
	if receptions := rules.ScoringRules["reception"]; receptions >= 0.75 {
		scoring = "ppr"
	} else if receptions > 0 {
		scoring = "half-ppr"
	}
	return fmt.Sprintf("sleeper-adp-%s-%s-%s", rules.LeagueFormat, scoring, map[bool]string{false: "1qb", true: "superflex"}[superflex])
}
