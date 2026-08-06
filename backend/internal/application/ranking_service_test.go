package application

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/coopa11y/DraftMeld/backend/internal/domain/ranking"
)

type rankingRepositoryStub struct {
	records []ranking.Record
}

func TestRankingDownloadRetriesTransientServerFailure(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		attempts++
		if request.Header.Get("User-Agent") != "DraftMeld/0.2 (+https://github.com/coopa11y/DraftMeld)" {
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

func (repository *rankingRepositoryStub) ReplaceRankings(context.Context, ranking.SourceDefinition, []ranking.Record, string, time.Time) error {
	return nil
}

func (repository *rankingRepositoryStub) RankingRecords(context.Context) ([]ranking.Record, error) {
	return repository.records, nil
}

func (repository *rankingRepositoryStub) RankingStatuses(context.Context) (map[string]ranking.SourceStatus, error) {
	return map[string]ranking.SourceStatus{}, nil
}

func TestConsensusUsesCurrentRedraftPoolAsEligibilityAnchor(t *testing.T) {
	repository := &rankingRepositoryStub{records: []ranking.Record{
		{SourceID: "expected-opportunity", PlayerKey: "current-player", Name: "Old Player Name", Position: "RB", Team: "OLD", Rank: 4},
		{SourceID: "expected-opportunity", PlayerKey: "historical-player", Name: "Historical Player", Position: "RB", Rank: 1},
		{SourceID: "redraft-ecr", PlayerKey: "current-player", Name: "Current Player", Position: "RB", Team: "NEW", Rank: 10},
	}}
	service := NewRankingService(repository)

	consensus, err := service.Consensus(t.Context())
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

	consensus, err := service.Consensus(t.Context())
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
