package application

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/coopa11y/DraftMeld/backend/internal/domain/draft"
)

func TestLeagueBackupRoundTripCreatesIndependentLeague(t *testing.T) {
	ctx := context.Background()
	repository := NewMemoryLeagueRepository(DemoLeagueConfiguration())
	leagues := NewLeagueService(repository)
	drafts, err := NewDraftServiceWithLeagues(&memoryDraftEvents{}, repository, draft.DemoCatalog())
	if err != nil {
		t.Fatal(err)
	}
	service := NewExportService(leagues, drafts, NewRankingService(nil))
	service.now = func() time.Time { return time.Date(2026, 8, 7, 12, 0, 0, 0, time.UTC) }

	backup, err := service.Backup(ctx, "demo")
	if err != nil {
		t.Fatal(err)
	}
	if backup.FormatVersion != 1 || backup.OriginalLeagueID != "demo" || backup.Rules.Name != "Demo League" || backup.Recommendation.RecommendationLimit != 5 {
		t.Fatalf("unexpected backup: %#v", backup)
	}
	restored, err := service.ImportBackup(ctx, backup)
	if err != nil {
		t.Fatal(err)
	}
	if restored.ID == backup.OriginalLeagueID || restored.Rules.Name != backup.Rules.Name || restored.Recommendation != backup.Recommendation {
		t.Fatalf("restore should create an independent league: %#v", restored)
	}
}

func TestLeagueBackupRejectsUnknownVersion(t *testing.T) {
	service := NewExportService(NewLeagueService(NewMemoryLeagueRepository()), nil, nil)
	_, err := service.ImportBackup(context.Background(), LeagueBackup{FormatVersion: 99})
	if !errors.Is(err, ErrUnsupportedBackupVersion) {
		t.Fatalf("expected unsupported version error, got %v", err)
	}
}

func TestDraftCSVUsesStableColumnsAndProtectsSpreadsheetCells(t *testing.T) {
	if got := safeCSVCell("=IMPORTXML(example)"); got != "'=IMPORTXML(example)" {
		t.Fatalf("expected formula-safe cell, got %q", got)
	}
	rows, err := encodeCSV([][]string{{"player", "cost"}, {safeCSVCell("+Risk"), "1.00"}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(rows), "'+Risk,1.00") {
		t.Fatalf("unexpected CSV: %s", rows)
	}
}

type memoryDraftEvents struct{ events []draft.Event }

func (repository *memoryDraftEvents) List(context.Context, string) ([]draft.Event, error) {
	return append([]draft.Event(nil), repository.events...), nil
}

func (repository *memoryDraftEvents) Append(_ context.Context, event draft.Event) (draft.Event, error) {
	event.ID = int64(len(repository.events) + 1)
	repository.events = append(repository.events, event)
	return event, nil
}
