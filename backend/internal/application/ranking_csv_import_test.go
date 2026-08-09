package application

import (
	"strings"
	"testing"

	"github.com/coopa11y/DraftMeld/backend/internal/domain/league"
	"github.com/coopa11y/DraftMeld/backend/internal/domain/ranking"
)

func TestImportRankingCSVMapsPrivateSourceIntoConsensus(t *testing.T) {
	repository := &rankingRepositoryStub{}
	service := NewRankingService(repository)
	csv := "RK,Player Full Name,POS,TM,Average Draft Position,Tier\n2,Second Player,WR,DAL,8.5,2\n1,First Player,RB,ATL,3.2,1\n"
	status, err := service.ImportCSV(t.Context(), "My Board", strings.NewReader(csv), map[string]string{
		"rank": "RK", "name": "Player Full Name", "position": "POS", "team": "TM",
		"adp": "Average Draft Position", "tier": "Tier",
	})
	if err != nil {
		t.Fatalf("import ranking CSV: %v", err)
	}
	if status.ID != "custom-my-board" || !status.IsCustom || status.RecordCount != 2 {
		t.Fatalf("unexpected custom source: %#v", status)
	}
	consensus, err := service.Consensus(t.Context(), map[string]league.RankingSourcePreference{
		"custom-my-board": {Weight: 1, Enabled: true},
	})
	if err != nil {
		t.Fatalf("build custom consensus: %v", err)
	}
	if len(consensus) != 2 || consensus[0].Name != "First Player" || consensus[0].ADP != 3.2 || consensus[0].Tier != 1 {
		t.Fatalf("unexpected custom consensus: %#v", consensus)
	}
}

func TestImportRankingCSVRejectsNameSlugCollision(t *testing.T) {
	repository := &rankingRepositoryStub{custom: []ranking.SourceDefinition{{ID: "custom-my-board", Name: "My Board!", IsCustom: true}}}
	service := NewRankingService(repository)
	_, err := service.ImportCSV(t.Context(), "My Board?", strings.NewReader("rank,name,position\n1,Player,RB\n"), nil)
	if err == nil || !strings.Contains(err.Error(), `conflicts with existing source "My Board!"`) {
		t.Fatalf("expected a source name conflict, got %v", err)
	}
}

func TestImportRankingCSVRejectsInvalidMappedRowsWithoutReplacingSource(t *testing.T) {
	repository := &rankingRepositoryStub{}
	service := NewRankingService(repository)
	_, err := service.ImportCSV(t.Context(), "Bad Board", strings.NewReader("rank,name,position\nnope,Player,RB\n"), nil)
	if err == nil || !strings.Contains(err.Error(), "row 2 requires a positive rank") {
		t.Fatalf("expected a row-specific rank error, got %v", err)
	}
	if len(repository.records) != 0 || len(repository.custom) != 0 {
		t.Fatalf("invalid import changed the repository: %#v %#v", repository.records, repository.custom)
	}
}
