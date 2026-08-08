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

func TestLeagueRuleImportRejectsUnsupportedContent(t *testing.T) {
	service := NewLeagueRuleImportService()
	if _, err := service.ImportCSV(strings.NewReader("Name,Value\nWaiver period,2\n")); err == nil {
		t.Fatal("expected unsupported scoring CSV to be rejected")
	}
}
