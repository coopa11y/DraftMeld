package sqlite

import (
	"strings"
	"testing"

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
