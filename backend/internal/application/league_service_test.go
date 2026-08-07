package application

import "testing"

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
	copy.Rules.SourceWeights["cbs-ppr"] = 10
	stored, err := service.Get(t.Context(), first.ID)
	if err != nil {
		t.Fatalf("reload source league: %v", err)
	}
	if stored.Rules.RosterSlots[0].Positions[0] != "QB" || stored.Rules.ScoringRules["reception"] != 1 || stored.Rules.SourceWeights["cbs-ppr"] == 10 {
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
