package application

import (
	"context"
	"testing"

	"github.com/coopa11y/DraftMeld/backend/internal/domain/draft"
)

type identityAliasReaderStub map[string]string

func (stub identityAliasReaderStub) IdentityAliases(context.Context) (map[string]string, error) {
	return stub, nil
}

func TestPlayerDirectoryAliasesPreserveDraftHistoryAndPreferences(t *testing.T) {
	repository := identityAliasReaderStub{"old-player-key": "player-stable"}
	events, err := resolveDraftEventAliases(t.Context(), repository, []draft.Event{{PlayerID: "old-player-key", Action: draft.ActionDraft}})
	if err != nil || events[0].PlayerID != "player-stable" {
		t.Fatalf("draft history alias was not resolved: %#v %v", events, err)
	}
	preferences, err := resolvePlayerPreferences(t.Context(), repository, map[string]string{"old-player-key": "target"})
	if err != nil || preferences["player-stable"] != "target" {
		t.Fatalf("player preference alias was not resolved: %#v %v", preferences, err)
	}
}
