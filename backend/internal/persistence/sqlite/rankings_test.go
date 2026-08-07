package sqlite

import (
	"context"
	"testing"
	"time"

	"github.com/coopa11y/DraftMeld/backend/internal/domain/ranking"
)

func TestReplaceRankingsIsAtomicPerSource(t *testing.T) {
	store, err := Open(t.TempDir() + "/draftmeld.db")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer store.Close()
	source := ranking.SourceDefinition{ID: "test", Name: "Test", Description: "Private", Methodology: "Ordinal", License: "Private", DefaultWeight: 1, ImportMode: "csv-upload", Role: "ranking", IsCustom: true}
	now := time.Now().UTC()
	if err = store.ReplaceRankings(context.Background(), source, []ranking.Record{{SourceID: "test", PlayerKey: "one", Name: "One", Position: "RB", Team: "AAA", Rank: 1}}, "today", now); err != nil {
		t.Fatalf("replace rankings: %v", err)
	}
	if err = store.ReplaceRankings(context.Background(), source, []ranking.Record{{SourceID: "test", PlayerKey: "two", Name: "Two", Position: "WR", Team: "BBB", Rank: 1, ADP: 7.5, Tier: 2}}, "tomorrow", now); err != nil {
		t.Fatalf("replace rankings again: %v", err)
	}
	records, err := store.RankingRecords(context.Background())
	if err != nil {
		t.Fatalf("list rankings: %v", err)
	}
	if len(records) != 1 || records[0].PlayerKey != "two" || records[0].ADP != 7.5 || records[0].Tier != 2 {
		t.Fatalf("unexpected records: %#v", records)
	}
	statuses, err := store.RankingStatuses(context.Background())
	if err != nil {
		t.Fatalf("list statuses: %v", err)
	}
	if statuses["test"].RecordCount != 1 || statuses["test"].PublishedAt != "tomorrow" {
		t.Fatalf("unexpected status: %#v", statuses["test"])
	}
	definitions, err := store.CustomRankingSources(context.Background())
	if err != nil || len(definitions) != 1 || definitions[0].ID != "test" || !definitions[0].IsCustom {
		t.Fatalf("unexpected custom definitions: %#v err=%v", definitions, err)
	}
}
