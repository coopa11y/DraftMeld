package application

import (
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
