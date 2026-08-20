package application

import (
	"fmt"
	"strings"
	"testing"

	"github.com/coopa11y/DraftMeld/backend/internal/document"
)

type leagueRulePDFExtractorStub struct {
	document document.TextDocument
}

func (stub leagueRulePDFExtractorStub) Extract([]byte) (document.TextDocument, error) {
	return stub.document, nil
}

func TestLeagueRuleCSVImportsRowAndPerUnitValues(t *testing.T) {
	service := NewLeagueRuleImportService()
	result, err := service.ImportCSV(strings.NewReader("Statistic,Points,Per\nPassing yards,1,25\nPassing touchdowns,4,1\nReceptions,0.5,1\nInterceptions thrown,-2,1\n"))
	if err != nil {
		t.Fatal(err)
	}
	if result.Rules["passingYard"] != 0.04 || result.Rules["passingTouchdown"] != 4 || result.Rules["reception"] != 0.5 || result.Rules["interception"] != -2 {
		t.Fatalf("unexpected imported rules: %#v", result.Rules)
	}
	if len(result.Matches) != 4 || len(result.Warnings) == 0 {
		t.Fatalf("expected reviewable matches and warning: %#v", result)
	}
}

func TestLeagueRuleCSVImportsWideFormat(t *testing.T) {
	service := NewLeagueRuleImportService()
	result, err := service.ImportCSV(strings.NewReader("Passing TD,Reception,Field goals made 50+ yards\n6,1,5\n"))
	if err != nil {
		t.Fatal(err)
	}
	if result.Rules["passingTouchdown"] != 6 || result.Rules["reception"] != 1 || result.Rules["fieldGoal50Plus"] != 5 {
		t.Fatalf("unexpected imported rules: %#v", result.Rules)
	}
}

func TestLeagueRulePDFImportsSelectableScoringText(t *testing.T) {
	extractor := leagueRulePDFExtractorStub{document: document.TextDocument{PageCount: 2, Text: `Passing Yards: 1 point for every 25 yards
Passing Touchdowns: 4
Interceptions Thrown: -2
TE Premium: 0.5
Points Allowed 35+: -4`}}
	result, err := NewLeagueRuleImportServiceWithExtractor(extractor).ImportPDF([]byte("pdf"))
	if err != nil {
		t.Fatal(err)
	}
	if result.Rules["passingYard"] != 0.04 || result.Rules["interception"] != -2 || result.Rules["tightEndReceptionBonus"] != 0.5 || result.Rules["defensePointsAllowed35Plus"] != -4 {
		t.Fatalf("unexpected imported rules: %#v", result.Rules)
	}
	if result.FileType != "pdf" || len(result.Matches) != 5 {
		t.Fatalf("unexpected PDF import result: %#v", result)
	}
}

func TestLeagueRulePDFHandlesLabelsAndValuesOnAdjacentLines(t *testing.T) {
	extractor := leagueRulePDFExtractorStub{document: document.TextDocument{PageCount: 1, Text: "Passing Touchdowns\n6 points\nInterceptions Thrown\n-1"}}
	result, err := NewLeagueRuleImportServiceWithExtractor(extractor).ImportPDF([]byte("pdf"))
	if err != nil {
		t.Fatal(err)
	}
	if result.Rules["passingTouchdown"] != 6 || result.Rules["interception"] != -1 {
		t.Fatalf("unexpected adjacent-line rules: %#v", result.Rules)
	}
}

func TestLeagueRulePDFWarnsWhenOCRWasApplied(t *testing.T) {
	extractor := leagueRulePDFExtractorStub{document: document.TextDocument{PageCount: 1, OCRApplied: true, Text: "Passing Touchdowns: 6"}}
	result, err := NewLeagueRuleImportServiceWithExtractor(extractor).ImportPDF([]byte("pdf"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(strings.Join(result.Warnings, " "), "recognized locally with OCR") {
		t.Fatalf("expected OCR review warning: %#v", result.Warnings)
	}
}

func TestLeagueRuleImportRejectsUnsupportedContent(t *testing.T) {
	service := NewLeagueRuleImportService()
	if _, err := service.ImportCSV(strings.NewReader("Name,Value\nWaiver period,2\n")); err == nil {
		t.Fatal("expected unsupported scoring CSV to be rejected")
	}
}

func TestLeagueRuleCSVImportsLeagueSettingsAndRoster(t *testing.T) {
	service := NewLeagueRuleImportService()
	result, err := service.ImportCSV(strings.NewReader(`Setting,Value
League name,Saturday League
Number of teams,10
League format,Dynasty
Draft format,Snake
Future pick seasons,3
Rookie draft rounds,4
FAAB budget,125
FAAB trades,Yes
QB,1
RB,2
WR,3
TE,1
FLEX,1
K,0
DST,1
Bench,7
`))
	if err != nil {
		t.Fatal(err)
	}
	if result.Settings.Name == nil || *result.Settings.Name != "Saturday League" || result.Settings.TeamCount == nil || *result.Settings.TeamCount != 10 {
		t.Fatalf("unexpected basic settings: %#v", result.Settings)
	}
	if result.Settings.LeagueFormat == nil || *result.Settings.LeagueFormat != "dynasty" || result.Settings.FAABTrades == nil || !*result.Settings.FAABTrades {
		t.Fatalf("unexpected dynasty settings: %#v", result.Settings)
	}
	if len(result.Settings.RosterSlots) != 8 || len(result.SettingMatches) < 10 {
		t.Fatalf("expected reviewable roster and settings: %#v", result)
	}
}

func TestLeagueRulePDFImportsAuctionAndInlineRoster(t *testing.T) {
	extractor := leagueRulePDFExtractorStub{document: document.TextDocument{PageCount: 1, Text: `League Name: Office League
Number of Teams: 12
League Type: Redraft
Draft Type: Salary Cap Auction
Auction Budget: $250
Minimum Bid: $2
Starting roster: 1 QB, 2 RB, 3 WR, 1 TE, 1 SUPERFLEX, 0 K, 1 D/ST, 8 Bench
Passing Touchdowns: 6`}}
	result, err := NewLeagueRuleImportServiceWithExtractor(extractor).ImportPDF([]byte("pdf"))
	if err != nil {
		t.Fatal(err)
	}
	if result.Settings.DraftType == nil || *result.Settings.DraftType != "auction" || result.Settings.AuctionBudget == nil || *result.Settings.AuctionBudget != 250 {
		t.Fatalf("unexpected auction settings: %#v", result.Settings)
	}
	if len(result.Settings.RosterSlots) != 8 || result.Rules["passingTouchdown"] != 6 {
		t.Fatalf("unexpected mixed PDF import: %#v", result)
	}
}

func TestESPNTextUsesExactLeagueLabelsAndVerticalRosterColumns(t *testing.T) {
	result, err := NewLeagueRuleImportService().ImportESPNText([]byte(`Opposing Teams
Knockin' On Evans DoorManager
Nacua MatataManager
Team 20Manager
Team 21Manager
J's Scary TeamManager
Joseph's Finest TeamManager
April Showers Bring Maye FlowerManager
Das HabichtsnestManager
Team 27Manager
One Hitter Quitters
League Name
Copper Canyon East
Number of Teams
10
Roster
POSITION
STARTERS
MAXIMUMS
Quarterback (QB)
1
4
Running Back (RB)
2
8
Running Back/Wide Receiver (RB/WR)
0
N/A
Wide Receiver (WR)
2
8
Wide Receiver/Tight End (WR/TE)
0
N/A
Tight End (TE)
1
3
Flex (FLEX)
1
N/A
Offensive Player Utility (OP)
0
N/A
Team Defense/Special Teams (D/ST)
1
3
Place Kicker (K)
1
3
Bench (BE)
7
N/A
Injured Reserve (IR)
2
N/A
Safety (S)
0
No Limit
Scoring
TD Pass (PTD)
4
2pt Passing Conversion (2PC)
2
TD Rush (RTD)
6
Each reception (REC)
1
TD Reception (RETD)
6
Each PAT Made (PAT)
1
FG Made (50-59 yards) (FG50)
5
FG Made (60+ yards) (FG60)
6
FG Made (0-39 yards) (FG0)
3
FG Made (40-49 yards) (FG40)
4
Each Sack (SK)
1
Each Interception (INT)
2
Each Fumble Recovered (FR)
2
Each Safety (SF)
2
1pt Safety (1PSF)
1
0 points allowed (PA0)
5
1-6 points allowed (PA1)
4
7-13 points allowed (PA7)
3
14-17 points allowed (PA14)
1
28-34 points allowed (PA28)
-1
35-45 points allowed (PA35)
-3
46+ points allowed (PA46)
-5
Less than 100 total yards allowed (YA100)
5
100-199 total yards allowed (YA199)
3
550+ total yards allowed (YA550)
-7
Fumble Recovered for TD (FTD)
6
Playoff Bracket Setup
Playoff Teams
4
Teams And Divisions Settings
EAST
Knockin' On Evans Door
Nacua Matata
Team 20
Team 21
J's Scary Team
Joseph's Finest Team
April Showers Bring Maye Flower
Das Habichtsnest
One Hitter Quitters
Team 27
Player Rules`))
	if err != nil {
		t.Fatal(err)
	}
	if result.Settings.TeamCount == nil || *result.Settings.TeamCount != 10 {
		t.Fatalf("expected the league team count, not playoff teams: %#v", result.Settings.TeamCount)
	}
	if result.Settings.Name == nil || *result.Settings.Name != "Copper Canyon East" {
		t.Fatalf("unexpected league name: %#v", result.Settings.Name)
	}
	if len(result.Settings.TeamNames) != 10 || result.Settings.TeamNames[8] != "One Hitter Quitters" {
		t.Fatalf("unexpected team names: %#v", result.Settings.TeamNames)
	}
	if result.Settings.UserTeamNumber == nil || *result.Settings.UserTeamNumber != 9 {
		t.Fatalf("expected the current ESPN team to be team 9: %#v", result.Settings.UserTeamNumber)
	}
	roster := map[string]int{}
	for _, slot := range result.Settings.RosterSlots {
		roster[slot.Name] = slot.Count
	}
	for name, count := range map[string]int{"QB": 1, "RB": 2, "WR": 2, "TE": 1, "FLEX": 1, "SUPERFLEX": 0, "DST": 1, "K": 1, "Bench": 7, "IR": 2} {
		if roster[name] != count {
			t.Fatalf("expected %s roster count %d, got %#v", name, count, roster)
		}
	}
	for key, value := range map[string]float64{
		"passingTouchdown":            4,
		"passingTwoPointConversion":   2,
		"rushingTouchdown":            6,
		"reception":                   1,
		"receivingTouchdown":          6,
		"extraPointMade":              1,
		"defenseSack":                 1,
		"defenseInterception":         2,
		"defenseFumbleRecovery":       2,
		"defenseSafety":               2,
		"defenseOnePointSafety":       1,
		"defensePointsAllowed0":       5,
		"defensePointsAllowed1To6":    4,
		"defensePointsAllowed7To13":   3,
		"defensePointsAllowed14To17":  1,
		"defensePointsAllowed28To34":  -1,
		"defensePointsAllowed35To45":  -3,
		"defensePointsAllowed46Plus":  -5,
		"defenseYardsAllowedUnder100": 5,
		"defenseYardsAllowed100To199": 3,
		"defenseYardsAllowed550Plus":  -7,
		"fieldGoal0To39":              3,
		"fieldGoal40To49":             4,
		"fieldGoal50To59":             5,
		"fieldGoal60Plus":             6,
		"fieldGoalMade":               0,
		"fieldGoal50Plus":             0,
		"defensePointsAllowed14To20":  0,
		"defensePointsAllowed21To27":  0,
		"defensePointsAllowed35Plus":  0,
	} {
		if result.Rules[key] != value {
			t.Fatalf("expected %s scoring value %g, got %#v", key, value, result.Rules)
		}
	}
	if _, imported := result.Rules["fumble"]; imported {
		t.Fatalf("a fumble-return touchdown must not become a generic fumble rule: %#v", result.Rules)
	}
	if strings.Contains(strings.Join(result.Warnings, " "), "Conflicting values") {
		t.Fatalf("ESPN roster and playoff labels must not create conflicting imports: %#v", result.Warnings)
	}
	for _, match := range append(result.SettingMatches, settingMatchesFromRules(result.Matches)...) {
		if match.Confidence != "high" {
			t.Fatalf("expected deterministic ESPN match to be high confidence: %#v", match)
		}
	}
}

func settingMatchesFromRules(matches []LeagueRuleMatch) []LeagueSettingMatch {
	result := make([]LeagueSettingMatch, 0, len(matches))
	for _, match := range matches {
		result = append(result, LeagueSettingMatch{Key: match.Key, Label: match.Label, Value: fmt.Sprint(match.Value), Source: match.Source, Confidence: match.Confidence})
	}
	return result
}
