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

func TestImportRankingCSVRecognizesFantasyFootballersTop200(t *testing.T) {
	repository := &rankingRepositoryStub{}
	service := NewRankingService(repository)
	input := "Rank,Name,Bye,Team,Pos,Andy,Jason,Mike,Markers\n1,Example Runner,7,BUF,RB,1,2,1,Favorite\n"
	status, err := service.ImportCSV(t.Context(), "Fantasy Footballers UDK Top 200", strings.NewReader(input), nil)
	if err != nil {
		t.Fatalf("import UDK Top 200: %v", err)
	}
	if status.Profile != "UDK Top 200" || status.ProjectURL == "" || status.PublishedAt != "Private UDK Top 200 CSV import" {
		t.Fatalf("unexpected UDK source metadata: %#v", status)
	}
	if len(repository.records) != 1 || repository.records[0].Name != "Example Runner" {
		t.Fatalf("unexpected UDK records: %#v", repository.records)
	}
}

func TestImportRankingCSVRejectsFantasyFootballersPositionRanks(t *testing.T) {
	repository := &rankingRepositoryStub{}
	service := NewRankingService(repository)
	input := "Name,Position,Team,Bye Week,Rank,Points,Risk,Upside,ADP,Tier,Outlook,Dynasty,Markers\nExample Quarterback,QB,BUF,7,1,350,2,9,2.12,1,Private notes,,\n"
	_, err := service.ImportCSV(t.Context(), "UDK QB", strings.NewReader(input), nil)
	if err == nil || !strings.Contains(err.Error(), "export the UDK Top 200 CSV") {
		t.Fatalf("expected position-ranking guidance, got %v", err)
	}
	if len(repository.records) != 0 {
		t.Fatalf("position-relative rankings changed the repository: %#v", repository.records)
	}
}
