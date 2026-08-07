package application

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/coopa11y/DraftMeld/backend/internal/domain/league"
	"github.com/coopa11y/DraftMeld/backend/internal/domain/ranking"
)

type rankingRepositoryStub struct {
	records []ranking.Record
	custom  []ranking.SourceDefinition
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

func (repository *rankingRepositoryStub) ReplaceRankings(_ context.Context, source ranking.SourceDefinition, records []ranking.Record, _ string, _ time.Time) error {
	repository.records = append([]ranking.Record(nil), records...)
	if source.IsCustom {
		repository.custom = []ranking.SourceDefinition{source}
	}
	return nil
}

func (repository *rankingRepositoryStub) RankingRecords(context.Context) ([]ranking.Record, error) {
	return repository.records, nil
}

func (repository *rankingRepositoryStub) RankingStatuses(context.Context) (map[string]ranking.SourceStatus, error) {
	return map[string]ranking.SourceStatus{}, nil
}

func (repository *rankingRepositoryStub) CustomRankingSources(context.Context) ([]ranking.SourceDefinition, error) {
	return repository.custom, nil
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
