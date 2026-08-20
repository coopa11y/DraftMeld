package application

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

type espnRoundTripper func(*http.Request) (*http.Response, error)

func (roundTripper espnRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	return roundTripper(request)
}

const espnSettingsFixture = `{
  "settings": {
    "name": "Accessible Champions",
    "size": 12,
    "draftSettings": {"type": 2, "auctionBudget": 250},
    "acquisitionSettings": {"acquisitionBudget": 100},
    "rosterSettings": {"lineupSlotCounts": {"0": 1, "2": 2, "4": 3, "6": 1, "7": 1, "16": 1, "17": 1, "20": 7, "21": 2, "23": 1, "9": 2}},
    "scoringSettings": {"scoringItems": [
      {"statId": 3, "points": 0.04},
      {"statId": 4, "points": 6},
      {"statId": 41, "points": 1},
      {"statId": 92, "points": 0}
    ]}
  }
}`

func TestESPNJSONImportsCanonicalSettingsRosterAndScoring(t *testing.T) {
	result, err := NewESPNLeagueImporter(nil).ImportJSON([]byte(espnSettingsFixture))
	if err != nil {
		t.Fatal(err)
	}
	if result.FileType != "espn" || result.Settings.Name == nil || *result.Settings.Name != "Accessible Champions" {
		t.Fatalf("unexpected basic settings: %#v", result)
	}
	if result.Settings.TeamCount == nil || *result.Settings.TeamCount != 12 || result.Settings.DraftType == nil || *result.Settings.DraftType != "auction" {
		t.Fatalf("unexpected team or draft settings: %#v", result.Settings)
	}
	if result.Settings.AuctionBudget == nil || *result.Settings.AuctionBudget != 250 || result.Settings.FAABBudget == nil || *result.Settings.FAABBudget != 100 {
		t.Fatalf("unexpected budgets: %#v", result.Settings)
	}
	if len(result.Settings.RosterSlots) != 10 || result.Rules["passingYard"] != 0.04 || result.Rules["passingTouchdown"] != 6 || result.Rules["reception"] != 1 {
		t.Fatalf("unexpected roster or scoring import: %#v", result)
	}
	warnings := strings.Join(result.Warnings, " ")
	if !strings.Contains(warnings, "roster slot IDs") || !strings.Contains(warnings, "scoring stat IDs") || !strings.Contains(warnings, "92") {
		t.Fatalf("expected unsupported values to be disclosed: %#v", result.Warnings)
	}
}

func TestESPNJSONRejectsInvalidAndUnsupportedContent(t *testing.T) {
	importer := NewESPNLeagueImporter(nil)
	if _, err := importer.ImportJSON([]byte("not json")); err == nil {
		t.Fatal("expected invalid JSON to fail")
	}
	if _, err := importer.ImportJSON([]byte(`{"settings":{"scoringSettings":{"scoringItems":[{"statId":999,"points":1}]}}}`)); !errorsIsNoRules(err) {
		t.Fatalf("expected unsupported payload to fail, got %v", err)
	}
}

func TestESPNLeagueFetchUsesOnlyValidatedLeagueID(t *testing.T) {
	client := &http.Client{Transport: espnRoundTripper(func(request *http.Request) (*http.Response, error) {
		if !strings.HasSuffix(request.URL.Path, "/seasons/2026/segments/0/leagues/793949449") || request.URL.Query().Get("view") != "mSettings" {
			t.Fatalf("unexpected ESPN request: %s", request.URL)
		}
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(espnSettingsFixture)), Header: make(http.Header)}, nil
	})}
	result, err := NewLeagueRuleImportServiceWithESPNClient(client).ImportESPNLeague(context.Background(), "https://fantasy.espn.com/football/league/settings?leagueId=793949449", 2026)
	if err != nil || result.Settings.TeamCount == nil || *result.Settings.TeamCount != 12 {
		t.Fatalf("unexpected public import: %#v, %v", result, err)
	}
	if _, err := parseESPNLeagueID("https://notespn.com/?leagueId=1"); err == nil {
		t.Fatal("expected lookalike domain to be rejected")
	}
}

func TestESPNLeagueFetchOffersPrivateLeagueFallback(t *testing.T) {
	client := &http.Client{Transport: espnRoundTripper(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusForbidden, Body: io.NopCloser(strings.NewReader("")), Header: make(http.Header)}, nil
	})}
	_, err := NewLeagueRuleImportServiceWithESPNClient(client).ImportESPNLeague(context.Background(), "https://fantasy.espn.com/football/league?leagueId=1", 2026)
	if err != ErrESPNLeaguePrivate {
		t.Fatalf("expected private league error, got %v", err)
	}
}

func errorsIsNoRules(err error) bool { return err == ErrNoLeagueRulesFound }
