package api

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/coopa11y/DraftMeld/backend/internal/application"
	"github.com/coopa11y/DraftMeld/backend/internal/domain/draft"
	draftsqlite "github.com/coopa11y/DraftMeld/backend/internal/persistence/sqlite"
)

func TestHealth(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	response := httptest.NewRecorder()
	router, closeStore := testRouter(t)
	defer closeStore()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}

	var body healthResponse
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode health response: %v", err)
	}
	if body.Status != "ok" || body.Version != "test" {
		t.Fatalf("unexpected health response: %#v", body)
	}
}

func TestDraftActionAndUndoEndpoints(t *testing.T) {
	router, closeStore := testRouter(t)
	defer closeStore()

	actionBody := bytes.NewBufferString(`{"leagueId":"demo","playerId":"p001","action":"draft"}`)
	actionRequest := httptest.NewRequest(http.MethodPost, "/api/v1/draft/actions", actionBody)
	actionResponse := httptest.NewRecorder()
	router.ServeHTTP(actionResponse, actionRequest)
	if actionResponse.Code != http.StatusOK {
		t.Fatalf("expected draft status %d, got %d: %s", http.StatusOK, actionResponse.Code, actionResponse.Body.String())
	}
	var afterAction draft.Snapshot
	if err := json.NewDecoder(actionResponse.Body).Decode(&afterAction); err != nil {
		t.Fatalf("decode draft snapshot: %v", err)
	}
	if len(afterAction.MyTeam) != 1 || afterAction.MyTeam[0].ID != "p001" {
		t.Fatalf("draft endpoint did not update team: %#v", afterAction.MyTeam)
	}

	undoRequest := httptest.NewRequest(http.MethodPost, "/api/v1/draft/undo", bytes.NewBufferString(`{"leagueId":"demo"}`))
	undoResponse := httptest.NewRecorder()
	router.ServeHTTP(undoResponse, undoRequest)
	if undoResponse.Code != http.StatusOK {
		t.Fatalf("expected undo status %d, got %d: %s", http.StatusOK, undoResponse.Code, undoResponse.Body.String())
	}
	var afterUndo draft.Snapshot
	if err := json.NewDecoder(undoResponse.Body).Decode(&afterUndo); err != nil {
		t.Fatalf("decode undo snapshot: %v", err)
	}
	if len(afterUndo.MyTeam) != 0 || len(afterUndo.Available) != len(draft.DemoCatalog()) {
		t.Fatalf("undo endpoint did not restore draft state")
	}
}

func testRouter(t *testing.T) (http.Handler, func()) {
	t.Helper()
	store, err := draftsqlite.Open(t.TempDir() + "/draftmeld.db")
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	service := application.NewDraftService(store, draft.DemoCatalog())
	return NewRouter(slog.New(slog.NewTextHandler(io.Discard, nil)), "test", service), func() { _ = store.Close() }
}
