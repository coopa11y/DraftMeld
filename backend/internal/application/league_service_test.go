package application

import (
	"testing"

	"github.com/coopa11y/DraftMeld/backend/internal/domain/league"
)

func TestLeagueServiceCreatesUniqueIDsAndDuplicatesIndependentRules(t *testing.T) {
	repository := NewMemoryLeagueRepository()
	service := NewLeagueService(repository)
	rules := DemoLeagueConfiguration().Rules
	rules.Name = "Home League"

	first, err := service.Create(t.Context(), rules)
	if err != nil {
		t.Fatalf("create first league: %v", err)
	}
	second, err := service.Create(t.Context(), rules)
	if err != nil {
		t.Fatalf("create second league: %v", err)
	}
	if first.ID != "home-league" || second.ID != "home-league-2" {
		t.Fatalf("unexpected unique IDs: %q and %q", first.ID, second.ID)
	}

	copy, err := service.Duplicate(t.Context(), first.ID)
	if err != nil {
		t.Fatalf("duplicate league: %v", err)
	}
	copy.Rules.RosterSlots[0].Positions[0] = "WR"
	copy.Rules.ScoringRules["reception"] = 0
	copy.Rules.SourcePreferences["cbs-ppr"] = league.RankingSourcePreference{Weight: 10, Enabled: false}
	stored, err := service.Get(t.Context(), first.ID)
	if err != nil {
		t.Fatalf("reload source league: %v", err)
	}
	if stored.Rules.RosterSlots[0].Positions[0] != "QB" || stored.Rules.ScoringRules["reception"] != 1 || stored.Rules.SourcePreferences["cbs-ppr"].Weight == 10 {
		t.Fatalf("duplicate mutated source configuration: %#v", stored.Rules)
	}
}

func TestLeagueServiceEnsuresDefaultOnlyWhenEmpty(t *testing.T) {
	repository := NewMemoryLeagueRepository()
	service := NewLeagueService(repository)
	if err := service.EnsureDefault(t.Context()); err != nil {
		t.Fatalf("ensure default league: %v", err)
	}
	if err := service.EnsureDefault(t.Context()); err != nil {
		t.Fatalf("ensure default league twice: %v", err)
	}
	leagues, err := service.List(t.Context())
	if err != nil || len(leagues) != 1 || leagues[0].ID != "demo" {
		t.Fatalf("unexpected default leagues: %#v err=%v", leagues, err)
	}
}

func TestLeagueServiceCreatesHiddenMockDraftConfiguration(t *testing.T) {
	repository := NewMemoryLeagueRepository()
	service := NewLeagueService(repository)
	rules := DemoLeagueConfiguration().Rules
	rules.Name = "Home League"
	created, err := service.Create(t.Context(), rules)
	if err != nil {
		t.Fatalf("create league: %v", err)
	}
	mock, err := service.CreateMockDraft(t.Context(), created.ID)
	if err != nil {
		t.Fatalf("create mock draft: %v", err)
	}
	if mock.Kind != league.ConfigurationKindMock || mock.ParentLeagueID != created.ID || mock.ID == created.ID {
		t.Fatalf("unexpected mock configuration: %#v", mock)
	}
	listed, err := service.List(t.Context())
	if err != nil || len(listed) != 1 || listed[0].ID != created.ID {
		t.Fatalf("mock draft leaked into league list: %#v err=%v", listed, err)
	}
	stored, err := service.Get(t.Context(), mock.ID)
	if err != nil || stored.ParentLeagueID != created.ID {
		t.Fatalf("mock draft was not retrievable: %#v err=%v", stored, err)
	}
}

func TestMemoryLeagueRepositoryDoesNotExposeStoredCollections(t *testing.T) {
	original := DemoLeagueConfiguration()
	original.Rules.TeamNames = []string{"Marcus", "Team 2"}
	repository := NewMemoryLeagueRepository(original)
	loaded, found, err := repository.GetLeague(t.Context(), original.ID)
	if err != nil || !found {
		t.Fatalf("load league: found=%v err=%v", found, err)
	}
	loaded.Rules.TeamNames[0] = "Mutated"
	loaded.Rules.RosterSlots[0].Positions[0] = "WR"
	loaded.Rules.ScoringRules["reception"] = 0

	reloaded, found, err := repository.GetLeague(t.Context(), original.ID)
	if err != nil || !found || reloaded.Rules.TeamNames[0] == "Mutated" || reloaded.Rules.RosterSlots[0].Positions[0] == "WR" || reloaded.Rules.ScoringRules["reception"] == 0 {
		t.Fatalf("repository exposed mutable stored state: %#v found=%v err=%v", reloaded.Rules, found, err)
	}
}

func TestLeagueServiceDefaultsUnknownTeamNamesAroundUserDraftPosition(t *testing.T) {
	repository := NewMemoryLeagueRepository()
	service := NewLeagueService(repository)
	rules := DemoLeagueConfiguration().Rules
	rules.Name = "Unknown Opponents League"
	rules.DraftPosition = 7
	rules.UserTeamNumber = 7
	rules.TeamNames = make([]string, rules.TeamCount)
	rules.TeamNames[6] = "Marcus"

	created, err := service.Create(t.Context(), rules)
	if err != nil {
		t.Fatalf("create league with unknown opponent names: %v", err)
	}
	if created.Rules.TeamNames[0] != "Team 1" || created.Rules.TeamNames[6] != "Marcus" || created.Rules.TeamNames[11] != "Team 12" {
		t.Fatalf("unexpected normalized team names: %#v", created.Rules.TeamNames)
	}
}

func TestLeagueServiceRequiresGuidedDynastySeasonRollover(t *testing.T) {
	repository := NewMemoryLeagueRepository()
	service := NewLeagueService(repository)
	rules := DemoLeagueConfiguration().Rules
	rules.Name = "Dynasty League"
	rules.LeagueFormat = league.LeagueFormatDynasty
	rules.FuturePickSeasons = 2
	rules.RookieDraftRounds = 4
	created, err := service.Create(t.Context(), rules)
	if err != nil {
		t.Fatalf("create dynasty league: %v", err)
	}
	rules = created.Rules
	rules.Season++
	if _, err = service.Update(t.Context(), created.ID, rules); err == nil {
		t.Fatal("expected direct dynasty season update to require the rollover workflow")
	}
}
