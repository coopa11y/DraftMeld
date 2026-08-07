package application

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/coopa11y/DraftMeld/backend/internal/domain/draft"
	"github.com/coopa11y/DraftMeld/backend/internal/domain/league"
)

const ExportFormatVersion = 1

var ErrUnsupportedBackupVersion = errors.New("unsupported backup format version")

type LeagueBackup struct {
	FormatVersion    int                         `json:"formatVersion"`
	ExportedAt       time.Time                   `json:"exportedAt"`
	OriginalLeagueID string                      `json:"originalLeagueId"`
	Rules            league.Rules                `json:"rules"`
	Recommendation   league.RecommendationPolicy `json:"recommendationPolicy"`
}

type DraftExport struct {
	FormatVersion int            `json:"formatVersion"`
	ExportedAt    time.Time      `json:"exportedAt"`
	LeagueID      string         `json:"leagueId"`
	LeagueName    string         `json:"leagueName"`
	DraftType     string         `json:"draftType"`
	Picks         []draft.Pick   `json:"picks"`
	MyTeam        []draft.Player `json:"myTeam"`
}

type ExportService struct {
	leagues  *LeagueService
	drafts   *DraftService
	rankings *RankingService
	now      func() time.Time
}

func NewExportService(leagues *LeagueService, drafts *DraftService, rankings *RankingService) *ExportService {
	return &ExportService{leagues: leagues, drafts: drafts, rankings: rankings, now: time.Now}
}

func (service *ExportService) Backup(ctx context.Context, leagueID string) (LeagueBackup, error) {
	configuration, err := service.leagues.Get(ctx, leagueID)
	if err != nil {
		return LeagueBackup{}, err
	}
	return LeagueBackup{
		FormatVersion: ExportFormatVersion, ExportedAt: service.now().UTC(),
		OriginalLeagueID: configuration.ID, Rules: configuration.Rules, Recommendation: configuration.Recommendation,
	}, nil
}

func (service *ExportService) ImportBackup(ctx context.Context, backup LeagueBackup) (LeagueConfiguration, error) {
	if backup.FormatVersion != ExportFormatVersion {
		return LeagueConfiguration{}, fmt.Errorf("%w: %d", ErrUnsupportedBackupVersion, backup.FormatVersion)
	}
	return service.leagues.Restore(ctx, backup.Rules, backup.Recommendation)
}

func (service *ExportService) RankingsCSV(ctx context.Context, leagueID string) ([]byte, error) {
	configuration, err := service.leagues.Get(ctx, leagueID)
	if err != nil {
		return nil, err
	}
	rankings, err := service.rankings.Consensus(ctx, configuration.Rules.SourcePreferences, configuration.Rules.ConsensusMethod)
	if err != nil {
		return nil, err
	}
	rows := [][]string{{
		"rank", "player", "position", "team", "consensus_score", "source_count", "coverage_percent",
		"rank_range", "confidence", "method", "enabled_source_weights", "source_ranks",
	}}
	weights := formatSourceWeights(configuration.Rules.SourcePreferences)
	for _, player := range rankings {
		rows = append(rows, []string{
			strconv.Itoa(player.Rank), safeCSVCell(player.Name), player.Position, safeCSVCell(player.Team),
			strconv.FormatFloat(player.Score, 'f', 2, 64), strconv.Itoa(player.SourceCount),
			strconv.FormatFloat(player.Coverage*100, 'f', 1, 64), strconv.Itoa(player.RankRange), player.Confidence,
			player.Method, weights, formatSourceRanks(player.SourceRanks),
		})
	}
	return encodeCSV(rows)
}

func (service *ExportService) Draft(ctx context.Context, leagueID string) (DraftExport, error) {
	snapshot, err := service.drafts.Snapshot(ctx, leagueID)
	if err != nil {
		return DraftExport{}, err
	}
	return DraftExport{
		FormatVersion: ExportFormatVersion, ExportedAt: service.now().UTC(), LeagueID: snapshot.LeagueID,
		LeagueName: snapshot.LeagueName, DraftType: snapshot.DraftType, Picks: snapshot.History, MyTeam: snapshot.MyTeam,
	}, nil
}

func (service *ExportService) DraftCSV(ctx context.Context, leagueID string) ([]byte, error) {
	exported, err := service.Draft(ctx, leagueID)
	if err != nil {
		return nil, err
	}
	rows := [][]string{{"pick", "action", "player", "position", "nfl_team", "cost", "recorded_at"}}
	for _, pick := range exported.Picks {
		rows = append(rows, []string{
			strconv.Itoa(pick.Number), string(pick.Action), safeCSVCell(pick.Player.Name), pick.Player.Position,
			safeCSVCell(pick.Player.NFLTeam), strconv.FormatFloat(pick.Cost, 'f', 2, 64), pick.CreatedAt.UTC().Format(time.RFC3339),
		})
	}
	return encodeCSV(rows)
}

func encodeCSV(rows [][]string) ([]byte, error) {
	var output bytes.Buffer
	writer := csv.NewWriter(&output)
	if err := writer.WriteAll(rows); err != nil {
		return nil, fmt.Errorf("write CSV export: %w", err)
	}
	return output.Bytes(), nil
}

func formatSourceWeights(preferences map[string]league.RankingSourcePreference) string {
	values := make([]string, 0, len(preferences))
	for sourceID, preference := range preferences {
		if preference.Enabled {
			values = append(values, fmt.Sprintf("%s=%s", sourceID, strconv.FormatFloat(preference.Weight, 'f', -1, 64)))
		}
	}
	sort.Strings(values)
	return safeCSVCell(strings.Join(values, "; "))
}

func formatSourceRanks(ranks map[string]int) string {
	values := make([]string, 0, len(ranks))
	for sourceID, rank := range ranks {
		values = append(values, fmt.Sprintf("%s=%d", sourceID, rank))
	}
	sort.Strings(values)
	return safeCSVCell(strings.Join(values, "; "))
}

func safeCSVCell(value string) string {
	trimmed := strings.TrimLeft(value, " \t\r\n")
	if trimmed != "" && strings.ContainsRune("=+-@", rune(trimmed[0])) {
		return "'" + value
	}
	return value
}
