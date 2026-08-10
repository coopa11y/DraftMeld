package application

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/coopa11y/DraftMeld/backend/internal/domain/league"
	"github.com/coopa11y/DraftMeld/backend/internal/domain/player"
	"github.com/coopa11y/DraftMeld/backend/internal/domain/ranking"
)

type rankingRepositoryStub struct {
	records  []ranking.Record
	custom   []ranking.SourceDefinition
	players  map[string]player.Player
	statuses map[string]ranking.SourceStatus
}

func TestRankingDownloadRetriesTransientServerFailure(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		attempts++
		if request.Header.Get("User-Agent") != "DraftMeld/0.3 (+https://github.com/coopa11y/DraftMeld)" {
			t.Errorf("unexpected user agent: %s", request.Header.Get("User-Agent"))
		}
		if attempts == 1 {
			response.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		_, _ = response.Write([]byte("rankings"))
	}))
	defer server.Close()
	service := NewRankingService(&rankingRepositoryStub{})

	contents, err := service.download(t.Context(), ranking.SourceDefinition{Name: "Test source", DataURL: server.URL})
	if err != nil {
		t.Fatalf("download after transient failure: %v", err)
	}
	if string(contents) != "rankings" || attempts != 2 {
		t.Fatalf("expected one retry and downloaded contents, got %q after %d attempts", contents, attempts)
	}
}

func (repository *rankingRepositoryStub) ReplaceRankings(_ context.Context, source ranking.SourceDefinition, records []ranking.Record, published string, refreshed time.Time) error {
	repository.records = append([]ranking.Record(nil), records...)
	if repository.statuses == nil {
		repository.statuses = make(map[string]ranking.SourceStatus)
	}
	repository.statuses[source.ID] = ranking.SourceStatus{SourceDefinition: source, RecordCount: len(records), RefreshedAt: &refreshed, PublishedAt: published}
	if source.IsCustom {
		repository.custom = []ranking.SourceDefinition{source}
	}
	return nil
}

func (repository *rankingRepositoryStub) RankingRecords(context.Context) ([]ranking.Record, error) {
	return repository.records, nil
}

func (repository *rankingRepositoryStub) RankingStatuses(context.Context) (map[string]ranking.SourceStatus, error) {
	return repository.statuses, nil
}

func TestRefreshSourceOnlyUpdatesTheRequestedFeed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		_, _ = response.Write([]byte(`<div class="player-row"><div class="rank">1</div><a href="/nfl/players/1/alex-rivers/fantasy/"><span class="team position">RB</span></a></div><div class="player-row"><div class="rank">2</div><a href="/nfl/players/2/jordan-hale/fantasy/"><span class="team position">WR</span></a></div>`))
	}))
	defer server.Close()
	repository := &rankingRepositoryStub{}
	service := NewRankingService(repository)
	service.sources = []ranking.SourceDefinition{
		{ID: "cbs-ppr", Name: "Requested", DataURL: server.URL, ImportMode: "download"},
		{ID: "other", Name: "Other", DataURL: "http://invalid.local", ImportMode: "download"},
	}

	status, err := service.RefreshSource(t.Context(), "cbs-ppr")
	if err != nil {
		t.Fatalf("refresh source: %v", err)
	}
	if status.ID != "cbs-ppr" || status.RecordCount != 2 || repository.records[0].SourceID != "cbs-ppr" {
		t.Fatalf("unexpected refreshed source: %#v records=%#v", status, repository.records)
	}
}

func TestRefreshSourceRejectsUnknownAndUploadSources(t *testing.T) {
	service := NewRankingService(&rankingRepositoryStub{})
	service.sources = []ranking.SourceDefinition{{ID: "upload", ImportMode: "pdf-upload"}}
	if _, err := service.RefreshSource(t.Context(), "missing"); !errors.Is(err, ErrRankingSourceNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
	if _, err := service.RefreshSource(t.Context(), "upload"); !errors.Is(err, ErrRankingSourceNotRefreshable) {
		t.Fatalf("expected not refreshable, got %v", err)
	}
}

func (repository *rankingRepositoryStub) CustomRankingSources(context.Context) ([]ranking.SourceDefinition, error) {
	return repository.custom, nil
}

func (repository *rankingRepositoryStub) CanonicalPlayers(_ context.Context, ids []string) (map[string]player.Player, error) {
	result := make(map[string]player.Player)
	for _, id := range ids {
		if current, exists := repository.players[id]; exists {
			result[id] = current
		}
	}
	return result, nil
}

func TestConsensusDisplaysCanonicalCurrentTeam(t *testing.T) {
	repository := &rankingRepositoryStub{
		records: []ranking.Record{
			{SourceID: "redraft-ecr", PlayerKey: "player-traded", Name: "Traded Player", Position: "RB", Team: "DAL", Rank: 12},
			{SourceID: "cbs-ppr", PlayerKey: "player-traded", Name: "Traded Player", Position: "RB", Team: "PIT", Rank: 14},
		},
		players: map[string]player.Player{"player-traded": {ID: "player-traded", Name: "Traded Player", Position: "RB", Team: "PIT"}},
	}
	consensus, err := NewRankingService(repository).Consensus(t.Context(), nil)
	if err != nil || len(consensus) != 1 || consensus[0].Team != "PIT" {
		t.Fatalf("expected canonical current team in consensus, got %#v %v", consensus, err)
	}
}

func TestConsensusUsesCurrentRedraftPoolAsEligibilityAnchor(t *testing.T) {
	repository := &rankingRepositoryStub{records: []ranking.Record{
		{SourceID: "expected-opportunity", PlayerKey: "current-player", Name: "Old Player Name", Position: "RB", Team: "OLD", Rank: 4},
		{SourceID: "expected-opportunity", PlayerKey: "historical-player", Name: "Historical Player", Position: "RB", Rank: 1},
		{SourceID: "redraft-ecr", PlayerKey: "current-player", Name: "Current Player", Position: "RB", Team: "NEW", Rank: 10},
	}}
	service := NewRankingService(repository)

	consensus, err := service.Consensus(t.Context(), nil)
	if err != nil {
		t.Fatalf("build consensus: %v", err)
	}
	if len(consensus) != 1 || consensus[0].PlayerKey != "current-player" {
		t.Fatalf("expected only the current redraft player, got %#v", consensus)
	}
	if consensus[0].SourceCount != 2 {
		t.Fatalf("expected both current-player signals, got %d", consensus[0].SourceCount)
	}
	if consensus[0].Name != "Current Player" || consensus[0].Team != "NEW" {
		t.Fatalf("expected current redraft metadata, got %#v", consensus[0])
	}
}

func TestConsensusMergesStoredDefenseAliases(t *testing.T) {
	repository := &rankingRepositoryStub{records: []ranking.Record{
		{SourceID: "espn-ppr-pdf", PlayerKey: "denverdefense", Name: "Denver Defense", Position: "DST", Team: "DEN", Rank: 140},
		{SourceID: "cbs-ppr", PlayerKey: "broncosdst", Name: "Broncos D/ST", Position: "DST", Rank: 132},
	}}
	service := NewRankingService(repository)

	consensus, err := service.Consensus(t.Context(), nil)
	if err != nil {
		t.Fatalf("build consensus: %v", err)
	}
	if len(consensus) != 1 || consensus[0].PlayerKey != "dstden" {
		t.Fatalf("expected one canonical Denver defense, got %#v", consensus)
	}
	if consensus[0].SourceCount != 2 || consensus[0].Team != "DEN" {
		t.Fatalf("expected both defense signals and canonical metadata, got %#v", consensus[0])
	}
}

func TestConsensusUsesEqualAndCustomWeightsWithoutDroppingEnabledSources(t *testing.T) {
	repository := &rankingRepositoryStub{records: []ranking.Record{
		{SourceID: "redraft-ecr", PlayerKey: "player-a", Name: "Player A", Position: "RB", Rank: 1},
		{SourceID: "redraft-ecr", PlayerKey: "player-b", Name: "Player B", Position: "RB", Rank: 3},
		{SourceID: "cbs-ppr", PlayerKey: "player-a", Name: "Player A", Position: "RB", Rank: 3},
		{SourceID: "cbs-ppr", PlayerKey: "player-b", Name: "Player B", Position: "RB", Rank: 1},
	}}
	service := NewRankingService(repository)

	equalConsensus, err := service.Consensus(t.Context(), map[string]league.RankingSourcePreference{
		"redraft-ecr": {Weight: 1, Enabled: true},
		"cbs-ppr":     {Weight: 1, Enabled: true},
	})
	if err != nil || len(equalConsensus) != 2 || equalConsensus[0].Score != equalConsensus[1].Score {
		t.Fatalf("expected equal source weights to produce equal blended scores, got %#v err=%v", equalConsensus, err)
	}

	consensus, err := service.Consensus(t.Context(), map[string]league.RankingSourcePreference{
		"redraft-ecr": {Weight: 1, Enabled: true},
		"cbs-ppr":     {Weight: 5, Enabled: true},
	})
	if err != nil {
		t.Fatalf("build weighted consensus: %v", err)
	}
	if len(consensus) != 2 || consensus[0].PlayerKey != "player-b" {
		t.Fatalf("expected CBS preference to move player-b first, got %#v", consensus)
	}
	for _, player := range consensus {
		if player.SourceCount != 2 {
			t.Fatalf("expected every available source to contribute, got %#v", player)
		}
	}
}

func TestWatchlistHighlightsPlayersPreferredByDisabledSources(t *testing.T) {
	records := []ranking.Record{
		{SourceID: "redraft-ecr", PlayerKey: "player-a", Name: "Player A", Position: "RB", Rank: 40},
		{SourceID: "redraft-ecr", PlayerKey: "player-b", Name: "Player B", Position: "WR", Rank: 1},
		{SourceID: "cbs-ppr", PlayerKey: "player-a", Name: "Player A", Position: "RB", Rank: 5},
		{SourceID: "cbs-ppr", PlayerKey: "player-b", Name: "Player B", Position: "WR", Rank: 2},
	}
	for rank := 2; rank < 40; rank++ {
		key := fmt.Sprintf("filler-%d", rank)
		records = append(records, ranking.Record{SourceID: "redraft-ecr", PlayerKey: key, Name: key, Position: "WR", Rank: rank})
	}
	repository := &rankingRepositoryStub{records: records}
	service := NewRankingService(repository)
	preferences := DefaultRankingSourcePreferences()
	preferences["cbs-ppr"] = league.RankingSourcePreference{Weight: 0.9, Enabled: false}

	watchlist, err := service.Watchlist(t.Context(), preferences)
	if err != nil {
		t.Fatalf("build watchlist: %v", err)
	}
	if len(watchlist) != 1 || watchlist[0].PlayerKey != "player-a" || watchlist[0].Signals[0].SpotsHigher != 35 {
		t.Fatalf("expected the CBS outlier only, got %#v", watchlist)
	}
}
