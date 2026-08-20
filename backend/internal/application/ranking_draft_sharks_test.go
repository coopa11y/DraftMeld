package application

import (
	"fmt"
	"strings"
	"testing"
)

func TestParseDraftSharksPreservesTiersAndProjectionEvidence(t *testing.T) {
	var input strings.Builder
	for rankValue := 1; rankValue <= 100; rankValue++ {
		name, position, team := fmt.Sprintf("Player %d", rankValue), "RB", "ATL"
		if rankValue == 1 {
			name, team = "Jahmyr Gibbs", "DET"
		}
		fmt.Fprintf(&input, `<tbody data-player-row data-key="%d" data-tier-overall="%d" data-fantasy-position="%s" data-player-name="%s">
<tr><td class="rank"><div class="column-title rank-index"><span>%d</span></div></td><td><img alt="%s logo"></td>
<td data-value="17" data-attribute="games_played"></td><td data-value="2.08" data-attribute="adp"></td>
<td data-value="6" data-attribute="player.team.bye"></td><td data-value="0.9%%" data-attribute="strength_of_schedule"></td>
<td data-value="54%%" data-attribute="player.sipPlayerProfile.injury_prob"></td><td data-value="231.2" data-attribute="floor_points"></td>
<td data-value="273" data-attribute="consensus_projection"></td><td data-value="285" data-attribute="fantasy_points"></td>
<td data-value="328.2" data-attribute="ceiling_points"></td><td data-value="100" data-attribute="dsValue"></td></tr></tbody>`,
			13000+rankValue, 1+(rankValue-1)/5, position, name, rankValue, team)
	}
	records, published, err := parseDraftSharks(strings.NewReader(input.String()), "draft-sharks-ppr-1qb")
	if err != nil {
		t.Fatal(err)
	}
	first := records[0]
	if len(records) != 100 || first.Name != "Jahmyr Gibbs" || first.Team != "DET" || first.ProviderID != "13001" {
		t.Fatalf("unexpected Draft Sharks identity data: %#v", first)
	}
	if first.Tier != 1 || first.ADP != 20 || first.Games != 17 || first.ByeWeek != 6 {
		t.Fatalf("unexpected Draft Sharks rank context: %#v", first)
	}
	if first.FloorProjection != 231.2 || first.ConsensusProjection != 273 || first.SourceProjection != 285 || first.CeilingProjection != 328.2 || first.SourceValue != 100 || first.InjuryRisk != 54 || first.ScheduleStrength != 0.9 {
		t.Fatalf("unexpected Draft Sharks projection evidence: %#v", first)
	}
	if published != "Current Draft Sharks PPR, 1QB preset" {
		t.Fatalf("unexpected publication label: %q", published)
	}
}

func TestParseDraftSharksRejectsIncompletePage(t *testing.T) {
	if _, _, err := parseDraftSharks(strings.NewReader(`<tbody data-player-row></tbody>`), "draft-sharks-ppr-1qb"); err == nil {
		t.Fatal("expected incomplete Draft Sharks response to fail")
	}
}

func TestDraftNotationToOverallPick(t *testing.T) {
	if got := draftNotationToOverall("10.12"); got != 120 {
		t.Fatalf("expected pick 120, got %g", got)
	}
	if got := draftNotationToOverall("bad"); got != 0 {
		t.Fatalf("expected invalid ADP to be zero, got %g", got)
	}
}
