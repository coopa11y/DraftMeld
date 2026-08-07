package sqlite

import (
	"strings"
	"testing"
	"time"

	"github.com/coopa11y/DraftMeld/backend/internal/domain/player"
)

func TestPlayerDirectoryResolvesAliasesAndProviderIDsToStablePlayer(t *testing.T) {
	store, err := Open(t.TempDir() + "/draftmeld.db")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	first, err := store.ResolvePlayer(t.Context(), player.Candidate{
		IdentityKey: "tankdell", LegacyKey: "tankdell", Name: "Tank Dell", Position: "WR", Team: "HOU",
		Provider: "source-a", ProviderID: "101",
	}, "player-first")
	if err != nil {
		t.Fatal(err)
	}
	second, err := store.ResolvePlayer(t.Context(), player.Candidate{
		IdentityKey: "tankdell", LegacyKey: "tankdell", Name: "Tank Dell", Position: "WR", Team: "HOU",
		Provider: "source-b", ProviderID: "abc",
	}, "player-should-not-be-used")
	if err != nil {
		t.Fatal(err)
	}
	if first.ID != "player-first" || second.ID != first.ID || !strings.HasPrefix(second.ID, "player-") {
		t.Fatalf("expected one stable canonical player, got %#v %#v", first, second)
	}
	status, err := store.PlayerDirectoryStatus(t.Context())
	if err != nil || status.PlayerCount != 1 || status.IdentityCount != 1 || status.ProviderIDCount != 2 {
		t.Fatalf("unexpected directory status: %#v %v", status, err)
	}
}

func TestPlayerDirectoryKeepsIdentityWhenPlayerChangesTeams(t *testing.T) {
	store, err := Open(t.TempDir() + "/draftmeld.db")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	weekOne := time.Date(2026, time.August, 1, 12, 0, 0, 0, time.UTC)
	first, err := store.ResolvePlayer(t.Context(), player.Candidate{
		IdentityKey: "tradedplayer", Name: "Traded Player", Position: "RB", Team: "DAL",
		Provider: "rankings-a", ProviderID: "99", ObservedAt: weekOne,
	}, "player-traded")
	if err != nil {
		t.Fatal(err)
	}
	traded, err := store.ResolvePlayer(t.Context(), player.Candidate{
		IdentityKey: "tradedplayer", Name: "Traded Player", Position: "RB", Team: "PIT",
		Provider: "rankings-b", ObservedAt: weekOne.Add(7 * 24 * time.Hour),
	}, "player-unused")
	if err != nil {
		t.Fatal(err)
	}
	if traded.ID != first.ID || traded.Team != "PIT" {
		t.Fatalf("expected the same player on the current team, got %#v %#v", first, traded)
	}
	for _, ignoredTeam := range []string{"", "FA", "Free Agent"} {
		current, resolveErr := store.ResolvePlayer(t.Context(), player.Candidate{
			IdentityKey: "tradedplayer", Name: "Traded Player", Position: "RB", Team: ignoredTeam,
			Provider: "rankings-c", ObservedAt: weekOne.Add(14 * 24 * time.Hour),
		}, "player-unused")
		if resolveErr != nil || current.Team != "PIT" {
			t.Fatalf("placeholder team %q replaced current team: %#v %v", ignoredTeam, current, resolveErr)
		}
	}
	older, err := store.ResolvePlayer(t.Context(), player.Candidate{
		IdentityKey: "tradedplayer", Name: "Traded Player", Position: "RB", Team: "DAL",
		Provider: "rankings-old", ObservedAt: weekOne,
	}, "player-unused")
	if err != nil || older.Team != "PIT" {
		t.Fatalf("older team observation replaced current team: %#v %v", older, err)
	}
}

func TestPlayerDirectoryUsesProviderIDAcrossANameChange(t *testing.T) {
	store, err := Open(t.TempDir() + "/draftmeld.db")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	first, err := store.ResolvePlayer(t.Context(), player.Candidate{
		IdentityKey: "oldname", LegacyKey: "oldname", Name: "Old Name", Position: "RB", Team: "ATL",
		Provider: "provider", ProviderID: "44",
	}, "player-stable")
	if err != nil {
		t.Fatal(err)
	}
	renamed, err := store.ResolvePlayer(t.Context(), player.Candidate{
		IdentityKey: "newname", LegacyKey: "newname", Name: "New Name", Position: "RB", Team: "ATL",
		Provider: "provider", ProviderID: "44",
	}, "player-new")
	if err != nil || renamed.ID != first.ID || renamed.Name != "New Name" {
		t.Fatalf("provider ID did not preserve the player: %#v %#v %v", first, renamed, err)
	}
}
