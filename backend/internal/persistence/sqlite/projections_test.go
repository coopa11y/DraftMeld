package sqlite

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/coopa11y/DraftMeld/backend/internal/domain/projection"
)

func TestReplaceProjectionsPreservesAtomicSourceIntegrity(t *testing.T) {
	store, err := Open(t.TempDir() + "/draftmeld.db")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	now := time.Date(2026, time.August, 9, 12, 0, 0, 0, time.UTC)
	source := projection.SourceStatus{ID: "private", Name: "Private projections", ImportedAt: now}
	initial := []projection.Record{
		{SourceID: source.ID, PlayerKey: "one", Name: "One", Position: "RB", Team: "AAA", ByeWeek: 7, ADP: 12.5, Stats: map[string]float64{"rushingYards": 900}},
		{SourceID: source.ID, PlayerKey: "two", Name: "Two", Position: "WR", Team: "BBB", Stats: map[string]float64{"receptions": 80}},
	}
	if err = store.ReplaceProjections(t.Context(), source, initial); err != nil {
		t.Fatal(err)
	}
	replacement := []projection.Record{{SourceID: source.ID, PlayerKey: "three", Name: "Three", Position: "QB", Team: "CCC", ADP: 4, Stats: map[string]float64{"passingTouchdowns": 30}}}
	if err = store.ReplaceProjections(t.Context(), source, replacement); err != nil {
		t.Fatal(err)
	}
	records, err := store.ProjectionRecords(t.Context())
	if err != nil || len(records) != 1 || records[0].PlayerKey != "three" || records[0].Stats["passingTouchdowns"] != 30 {
		t.Fatalf("replacement was not atomic: %#v err=%v", records, err)
	}
	statuses, err := store.ProjectionStatuses(t.Context())
	if err != nil || len(statuses) != 1 || statuses[0].RecordCount != 1 || !statuses[0].ImportedAt.Equal(now) {
		t.Fatalf("source status does not match stored records: %#v err=%v", statuses, err)
	}

	cancelled, cancel := context.WithCancel(t.Context())
	cancel()
	if err = store.ReplaceProjections(cancelled, source, initial); err == nil {
		t.Fatal("expected a cancelled replacement to fail")
	}
	records, err = store.ProjectionRecords(t.Context())
	if err != nil || len(records) != 1 || records[0].PlayerKey != "three" {
		t.Fatalf("failed replacement corrupted existing projections: %#v err=%v", records, err)
	}
}

func TestProjectionRecordsRejectCorruptStoredStats(t *testing.T) {
	store, err := Open(t.TempDir() + "/draftmeld.db")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	if _, err = store.database.Exec(`INSERT INTO projection_sources (id, name, imported_at, record_count) VALUES ('bad', 'Bad', '2026-08-09T12:00:00Z', 1)`); err != nil {
		t.Fatal(err)
	}
	if _, err = store.database.Exec(`INSERT INTO player_projections (source_id, player_key, player_name, position, nfl_team, bye_week, adp, stats_json) VALUES ('bad', 'player', 'Player', 'RB', 'AAA', 0, 0, '{')`); err != nil {
		t.Fatal(err)
	}
	if _, err = store.ProjectionRecords(t.Context()); err == nil || !strings.Contains(err.Error(), "decode projection stats") {
		t.Fatalf("expected corrupt stats to be rejected, got %v", err)
	}
}
