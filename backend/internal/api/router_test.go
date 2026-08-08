package api

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
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
	if len(sources) != 7 {
		t.Fatalf("expected seven built-in sources, got %d", len(sources))
	}
	for _, source := range sources {
		if source.ID == "" || source.License == "" || source.ProjectURL == "" {
			t.Fatalf("source is missing provenance: %#v", source)
		}
	}
}

func TestConsensusRankingsRequireKnownLeague(t *testing.T) {
	router, closeStore := testRouter(t)
	defer closeStore()
	for _, endpoint := range []string{"rankings", "ranking-watchlist"} {
		for _, test := range []struct {
			suffix string
			want   int
		}{{"", http.StatusBadRequest}, {"?leagueId=missing", http.StatusNotFound}, {"?leagueId=demo", http.StatusOK}} {
			path := "/api/v1/" + endpoint + test.suffix
			response := httptest.NewRecorder()
			router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
			if response.Code != test.want {
				t.Errorf("expected %d for %s, got %d: %s", test.want, path, response.Code, response.Body.String())
			}
		}
	}
}

func TestRankingPDFImportRequiresAFile(t *testing.T) {
	router, closeStore := testRouter(t)
	defer closeStore()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/ranking-sources/import-pdf", bytes.NewBufferString("not multipart"))
	router.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d: %s", http.StatusBadRequest, response.Code, response.Body.String())
	}
}

func TestRankingPDFImportRejectsMultipleFiles(t *testing.T) {
	router, closeStore := testRouter(t)
	defer closeStore()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	for _, filename := range []string{"first.pdf", "second.pdf"} {
		part, err := writer.CreateFormFile("file", filename)
		if err != nil {
			t.Fatalf("create multipart file: %v", err)
		}
		_, _ = part.Write([]byte("%PDF-test"))
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart body: %v", err)
	}
	request := httptest.NewRequest(http.MethodPost, "/api/v1/ranking-sources/import-pdf", &body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest || !strings.Contains(response.Body.String(), "one PDF at a time") {
		t.Fatalf("expected duplicate-file error, got %d: %s", response.Code, response.Body.String())
	}
}

func TestRankingCSVImportCreatesAWeightablePrivateSource(t *testing.T) {
	router, closeStore := testRouter(t)
	defer closeStore()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	_ = writer.WriteField("name", "Marcus rankings")
	_ = writer.WriteField("mapping", `{"rank":"RK","name":"Player","position":"POS","team":"TM","adp":"ADP","tier":"Tier","providerId":"Player ID"}`)
	file, err := writer.CreateFormFile("file", "rankings.csv")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = file.Write([]byte("RK,Player,POS,TM,ADP,Tier,Player ID\n1,Custom Runner,RB,ATL,4.2,1,runner-1\n2,Custom Receiver,WR,DAL,9.5,2,receiver-2\n"))
	_ = writer.Close()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/ranking-sources/import-csv", &body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusCreated {
		t.Fatalf("ranking import failed: %d %s", response.Code, response.Body.String())
	}
	if !strings.Contains(response.Body.String(), `"id":"custom-marcus-rankings"`) || !strings.Contains(response.Body.String(), `"isCustom":true`) {
		t.Fatalf("unexpected import response: %s", response.Body.String())
	}

	rankingsResponse := httptest.NewRecorder()
	router.ServeHTTP(rankingsResponse, httptest.NewRequest(http.MethodGet, "/api/v1/rankings?leagueId=demo", nil))
	if rankingsResponse.Code != http.StatusOK || !strings.Contains(rankingsResponse.Body.String(), `"name":"Custom Runner"`) || !strings.Contains(rankingsResponse.Body.String(), `"adp":4.2`) {
		t.Fatalf("custom source did not reach consensus: %d %s", rankingsResponse.Code, rankingsResponse.Body.String())
	}

	sourcesResponse := httptest.NewRecorder()
	router.ServeHTTP(sourcesResponse, httptest.NewRequest(http.MethodGet, "/api/v1/ranking-sources", nil))
	if sourcesResponse.Code != http.StatusOK || !strings.Contains(sourcesResponse.Body.String(), `"name":"Marcus rankings"`) {
		t.Fatalf("custom source was not persisted: %d %s", sourcesResponse.Code, sourcesResponse.Body.String())
	}
	directoryResponse := httptest.NewRecorder()
	router.ServeHTTP(directoryResponse, httptest.NewRequest(http.MethodGet, "/api/v1/player-directory/status", nil))
	if directoryResponse.Code != http.StatusOK || !strings.Contains(directoryResponse.Body.String(), `"playerCount":2`) || !strings.Contains(directoryResponse.Body.String(), `"providerIdCount":2`) {
		t.Fatalf("unexpected player directory status: %d %s", directoryResponse.Code, directoryResponse.Body.String())
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

func TestDraftPickTradeEndpointHandlesPickPackages(t *testing.T) {
	router, closeStore := testRouter(t)
	defer closeStore()
	body := bytes.NewBufferString(`{"leagueId":"demo","teamOneNumber":12,"teamTwoNumber":1,"teamOneReceives":[1],"teamTwoReceives":[12,13]}`)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/api/v1/draft/pick-trades", body))
	if response.Code != http.StatusCreated {
		t.Fatalf("expected trade status %d, got %d: %s", http.StatusCreated, response.Code, response.Body.String())
	}
	var snapshot draft.Snapshot
	if err := json.NewDecoder(response.Body).Decode(&snapshot); err != nil {
		t.Fatal(err)
	}
	if len(snapshot.PickTrades) != 1 || snapshot.PickSlots[0].OwnerTeamNumber != 12 || snapshot.PickSlots[11].OwnerTeamNumber != 1 {
		t.Fatalf("unexpected traded-pick snapshot: %#v", snapshot)
	}

	reverse := httptest.NewRecorder()
	router.ServeHTTP(reverse, httptest.NewRequest(http.MethodDelete, "/api/v1/draft/pick-trades/1?leagueId=demo", nil))
	if reverse.Code != http.StatusOK || !strings.Contains(reverse.Body.String(), `"pickTrades":[]`) {
		t.Fatalf("trade reversal failed: %d %s", reverse.Code, reverse.Body.String())
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
	if created.ID != "work-league" || created.DraftPosition != 4 || len(created.SourcePreferences) != len(application.BuiltInRankingSources()) {
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
	router, closeStore, _ := testRouterWithDraftService(t)
	return router, closeStore
}

func TestDraftWorkflowEndToEnd(t *testing.T) {
	router, closeStore, draftService := testRouterWithDraftService(t)
	defer closeStore()

	var projectionBody bytes.Buffer
	projectionWriter := multipart.NewWriter(&projectionBody)
	_ = projectionWriter.WriteField("name", "Mapped projections")
	_ = projectionWriter.WriteField("mapping", `{"name":"Player Name","position":"Pos","team":"Tm","reception":"REC"}`)
	projectionFile, err := projectionWriter.CreateFormFile("file", "projections.csv")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = projectionFile.Write([]byte("Player Name,Pos,Tm,REC\nAlex Rivers,RB,ATL,70\nJordan Hale,WR,MIN,80\n"))
	_ = projectionWriter.Close()
	projectionRequest := httptest.NewRequest(http.MethodPost, "/api/v1/projection-sources/import-csv", &projectionBody)
	projectionRequest.Header.Set("Content-Type", projectionWriter.FormDataContentType())
	projectionResponse := httptest.NewRecorder()
	router.ServeHTTP(projectionResponse, projectionRequest)
	if projectionResponse.Code != http.StatusCreated {
		t.Fatalf("projection import failed: %d %s", projectionResponse.Code, projectionResponse.Body.String())
	}

	preferenceResponse := httptest.NewRecorder()
	router.ServeHTTP(preferenceResponse, httptest.NewRequest(http.MethodPut, "/api/v1/draft/preferences", bytes.NewBufferString(`{"leagueId":"demo","playerId":"p001","preference":"target"}`)))
	if preferenceResponse.Code != http.StatusOK || !strings.Contains(preferenceResponse.Body.String(), `"preference":"target"`) {
		t.Fatalf("target preference failed: %d %s", preferenceResponse.Code, preferenceResponse.Body.String())
	}

	draftResponse := httptest.NewRecorder()
	router.ServeHTTP(draftResponse, httptest.NewRequest(http.MethodPost, "/api/v1/draft/actions", bytes.NewBufferString(`{"leagueId":"demo","playerId":"p001","action":"draft"}`)))
	if draftResponse.Code != http.StatusOK {
		t.Fatalf("draft action failed: %d %s", draftResponse.Code, draftResponse.Body.String())
	}
	mockResponse := httptest.NewRecorder()
	router.ServeHTTP(mockResponse, httptest.NewRequest(http.MethodPost, "/api/v1/draft/mock", bytes.NewBufferString(`{"leagueId":"demo"}`)))
	if mockResponse.Code != http.StatusOK {
		t.Fatalf("mock draft failed: %d %s", mockResponse.Code, mockResponse.Body.String())
	}

	sleeperServer := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		_, _ = response.Write([]byte(`[{"pick_no":1,"roster_id":7,"metadata":{"first_name":"Alex","last_name":"Rivers","position":"RB","team":"ATL"}},{"pick_no":2,"roster_id":2,"metadata":{"first_name":"Jordan","last_name":"Hale","position":"WR","team":"MIN"}}]`))
	}))
	defer sleeperServer.Close()
	draftService.ConfigureSleeperClient(sleeperServer.Client(), sleeperServer.URL)
	syncResponse := httptest.NewRecorder()
	router.ServeHTTP(syncResponse, httptest.NewRequest(http.MethodPost, "/api/v1/draft/sync/sleeper", bytes.NewBufferString(`{"leagueId":"demo","sleeperDraftId":"draft-1","rosterId":7}`)))
	if syncResponse.Code != http.StatusOK {
		t.Fatalf("Sleeper sync failed: %d %s", syncResponse.Code, syncResponse.Body.String())
	}
	var synced application.SleeperSyncResult
	if err = json.NewDecoder(syncResponse.Body).Decode(&synced); err != nil {
		t.Fatal(err)
	}
	if len(synced.Snapshot.History) != 2 || len(synced.Snapshot.MyTeam) != 1 || synced.Snapshot.MyTeam[0].ID != "p001" {
		t.Fatalf("workflow did not reconcile the Sleeper source of truth: %#v", synced)
	}

	undoResponse := httptest.NewRecorder()
	router.ServeHTTP(undoResponse, httptest.NewRequest(http.MethodPost, "/api/v1/draft/undo", bytes.NewBufferString(`{"leagueId":"demo"}`)))
	if undoResponse.Code != http.StatusOK {
		t.Fatalf("undo failed: %d %s", undoResponse.Code, undoResponse.Body.String())
	}
}

func TestExportAndRestoreWorkflow(t *testing.T) {
	router, closeStore := testRouter(t)
	defer closeStore()

	backupResponse := httptest.NewRecorder()
	router.ServeHTTP(backupResponse, httptest.NewRequest(http.MethodGet, "/api/v1/leagues/demo/backup", nil))
	if backupResponse.Code != http.StatusOK || !strings.Contains(backupResponse.Header().Get("Content-Disposition"), "backup.json") {
		t.Fatalf("backup export failed: %d %s", backupResponse.Code, backupResponse.Body.String())
	}
	var backup application.LeagueBackup
	if err := json.NewDecoder(backupResponse.Body).Decode(&backup); err != nil {
		t.Fatal(err)
	}
	backupBytes, err := json.Marshal(backup)
	if err != nil {
		t.Fatal(err)
	}
	restoreResponse := httptest.NewRecorder()
	router.ServeHTTP(restoreResponse, httptest.NewRequest(http.MethodPost, "/api/v1/leagues/import", bytes.NewReader(backupBytes)))
	if restoreResponse.Code != http.StatusCreated || !strings.Contains(restoreResponse.Body.String(), `"id":"demo-league"`) {
		t.Fatalf("backup restore failed: %d %s", restoreResponse.Code, restoreResponse.Body.String())
	}

	actionResponse := httptest.NewRecorder()
	router.ServeHTTP(actionResponse, httptest.NewRequest(http.MethodPost, "/api/v1/draft/actions", bytes.NewBufferString(`{"leagueId":"demo","playerId":"p001","action":"draft"}`)))
	if actionResponse.Code != http.StatusOK {
		t.Fatalf("draft action failed: %d %s", actionResponse.Code, actionResponse.Body.String())
	}
	for _, export := range []struct {
		path        string
		contentType string
		contains    string
	}{
		{"/api/v1/leagues/demo/exports/rankings.csv", "text/csv", "enabled_source_weights"},
		{"/api/v1/leagues/demo/exports/draft.csv", "text/csv", "Alex Rivers"},
		{"/api/v1/leagues/demo/exports/draft.json", "application/json", `"picks"`},
	} {
		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, export.path, nil))
		if response.Code != http.StatusOK || !strings.HasPrefix(response.Header().Get("Content-Type"), export.contentType) || !strings.Contains(response.Body.String(), export.contains) {
			t.Fatalf("export %s failed: %d %s", export.path, response.Code, response.Body.String())
		}
	}

	invalidResponse := httptest.NewRecorder()
	router.ServeHTTP(invalidResponse, httptest.NewRequest(http.MethodPost, "/api/v1/leagues/import", bytes.NewBufferString(`{"formatVersion":99,"exportedAt":"2026-08-07T12:00:00Z","originalLeagueId":"demo","rules":{}}`)))
	if invalidResponse.Code != http.StatusBadRequest {
		t.Fatalf("expected unsupported backup status %d, got %d", http.StatusBadRequest, invalidResponse.Code)
	}
}

func testRouterWithDraftService(t *testing.T) (http.Handler, func(), *application.DraftService) {
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
	rankingService := application.NewRankingService(store)
	projectionService := application.NewProjectionService(store)
	service.UseIntelligence(rankingService, projectionService)
	return NewRouter(slog.New(slog.NewTextHandler(io.Discard, nil)), "test", service, leagueService, rankingService, projectionService), func() { _ = store.Close() }, service
}
