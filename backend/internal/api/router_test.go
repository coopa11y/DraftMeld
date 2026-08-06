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

func TestRankingSourcesExposeBuiltInProvenance(t *testing.T) {
	router, closeStore := testRouter(t)
	defer closeStore()
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v1/ranking-sources", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}
	var sources []struct {
		ID         string `json:"id"`
		License    string `json:"license"`
		ProjectURL string `json:"projectUrl"`
	}
	if err := json.NewDecoder(response.Body).Decode(&sources); err != nil {
		t.Fatalf("decode ranking sources: %v", err)
	}
	if len(sources) != 5 {
		t.Fatalf("expected five built-in sources, got %d", len(sources))
	}
	for _, source := range sources {
		if source.ID == "" || source.License == "" || source.ProjectURL == "" {
			t.Fatalf("source is missing provenance: %#v", source)
		}
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

func TestDraftEndpointRequiresKnownLeague(t *testing.T) {
	router, closeStore := testRouter(t)
	defer closeStore()

	for _, path := range []string{"/api/v1/draft", "/api/v1/draft?leagueId=missing"} {
		request := httptest.NewRequest(http.MethodGet, path, nil)
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		expected := http.StatusBadRequest
		if path != "/api/v1/draft" {
			expected = http.StatusNotFound
		}
		if response.Code != expected {
			t.Fatalf("expected %d for %s, got %d", expected, path, response.Code)
		}
	}
}

func TestLeagueLifecycleEndpoints(t *testing.T) {
	router, closeStore := testRouter(t)
	defer closeStore()

	rules := `{"name":"Work League","teamCount":10,"draftPosition":4,"draftType":"snake","rosterSlots":[{"name":"QB","count":1,"positions":["QB"],"isStarting":true},{"name":"Bench","count":5,"positions":["QB","RB","WR","TE"],"isStarting":false}],"scoringRules":{"reception":0.5}}`
	createResponse := httptest.NewRecorder()
	router.ServeHTTP(createResponse, httptest.NewRequest(http.MethodPost, "/api/v1/leagues", bytes.NewBufferString(rules)))
	if createResponse.Code != http.StatusCreated {
		t.Fatalf("expected create status %d, got %d: %s", http.StatusCreated, createResponse.Code, createResponse.Body.String())
	}
	var created leagueResponse
	if err := json.NewDecoder(createResponse.Body).Decode(&created); err != nil {
		t.Fatalf("decode created league: %v", err)
	}
	if created.ID != "work-league" || created.DraftPosition != 4 {
		t.Fatalf("unexpected created league: %#v", created)
	}

	draftResponse := httptest.NewRecorder()
	router.ServeHTTP(draftResponse, httptest.NewRequest(http.MethodGet, "/api/v1/draft?leagueId=work-league", nil))
	if draftResponse.Code != http.StatusOK {
		t.Fatalf("new league was not available to draft service: %s", draftResponse.Body.String())
	}

	duplicateResponse := httptest.NewRecorder()
	router.ServeHTTP(duplicateResponse, httptest.NewRequest(http.MethodPost, "/api/v1/leagues/work-league/duplicate", nil))
	if duplicateResponse.Code != http.StatusCreated {
		t.Fatalf("expected duplicate status %d, got %d", http.StatusCreated, duplicateResponse.Code)
	}

	deleteResponse := httptest.NewRecorder()
	router.ServeHTTP(deleteResponse, httptest.NewRequest(http.MethodDelete, "/api/v1/leagues/work-league", nil))
	if deleteResponse.Code != http.StatusNoContent {
		t.Fatalf("expected delete status %d, got %d", http.StatusNoContent, deleteResponse.Code)
	}
	missingResponse := httptest.NewRecorder()
	router.ServeHTTP(missingResponse, httptest.NewRequest(http.MethodGet, "/api/v1/leagues/work-league", nil))
	if missingResponse.Code != http.StatusNotFound {
		t.Fatalf("expected deleted league status %d, got %d", http.StatusNotFound, missingResponse.Code)
	}
}

func testRouter(t *testing.T) (http.Handler, func()) {
	t.Helper()
	store, err := draftsqlite.Open(t.TempDir() + "/draftmeld.db")
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	leagueService := application.NewLeagueService(store)
	if err = leagueService.EnsureDefault(t.Context()); err != nil {
		t.Fatalf("initialize leagues: %v", err)
	}
	service, err := application.NewDraftServiceWithLeagues(store, store, draft.DemoCatalog())
	if err != nil {
		t.Fatalf("create persisted draft service: %v", err)
	}
	return NewRouter(slog.New(slog.NewTextHandler(io.Discard, nil)), "test", service, leagueService, application.NewRankingService(store)), func() { _ = store.Close() }
}
