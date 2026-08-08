package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/coopa11y/DraftMeld/backend/internal/domain/league"
)

func (store *DraftEventStore) ListLeagues(ctx context.Context) ([]league.Configuration, error) {
	rows, err := store.database.QueryContext(ctx, `SELECT id FROM leagues ORDER BY name, id`)
	if err != nil {
		return nil, fmt.Errorf("query leagues: %w", err)
	}
	ids := make([]string, 0)
	for rows.Next() {
		var id string
		if err = rows.Scan(&id); err != nil {
			_ = rows.Close()
			return nil, fmt.Errorf("scan league ID: %w", err)
		}
		ids = append(ids, id)
	}
	if err = rows.Close(); err != nil {
		return nil, fmt.Errorf("close league rows: %w", err)
	}

	leagues := make([]league.Configuration, 0, len(ids))
	for _, id := range ids {
		configuration, found, getErr := store.GetLeague(ctx, id)
		if getErr != nil {
			return nil, getErr
		}
		if found {
			leagues = append(leagues, configuration)
		}
	}
	return leagues, nil
}

func (store *DraftEventStore) GetLeague(ctx context.Context, id string) (league.Configuration, bool, error) {
	var configuration league.Configuration
	var draftType string
	var scoringJSON, sourcePreferencesJSON, recommendationJSON, draftSettingsJSON string
	err := store.database.QueryRowContext(ctx, `
SELECT id, name, team_count, draft_position, draft_type, scoring_rules, source_preferences, recommendation_policy, draft_settings
FROM leagues WHERE id = ?`, id).Scan(
		&configuration.ID, &configuration.Rules.Name, &configuration.Rules.TeamCount,
		&configuration.Rules.DraftPosition, &draftType, &scoringJSON, &sourcePreferencesJSON, &recommendationJSON, &draftSettingsJSON,
	)
	if err == sql.ErrNoRows {
		return league.Configuration{}, false, nil
	}
	if err != nil {
		return league.Configuration{}, false, fmt.Errorf("query league: %w", err)
	}
	configuration.Rules.DraftType = league.DraftType(draftType)
	if err = json.Unmarshal([]byte(scoringJSON), &configuration.Rules.ScoringRules); err != nil {
		return league.Configuration{}, false, fmt.Errorf("decode league scoring rules: %w", err)
	}
	if err = json.Unmarshal([]byte(sourcePreferencesJSON), &configuration.Rules.SourcePreferences); err != nil {
		return league.Configuration{}, false, fmt.Errorf("decode ranking source preferences: %w", err)
	}
	if err = json.Unmarshal([]byte(recommendationJSON), &configuration.Recommendation); err != nil {
		return league.Configuration{}, false, fmt.Errorf("decode recommendation policy: %w", err)
	}
	var settings leagueDraftSettings
	if err = json.Unmarshal([]byte(draftSettingsJSON), &settings); err != nil {
		return league.Configuration{}, false, fmt.Errorf("decode league draft settings: %w", err)
	}
	settings.apply(&configuration.Rules)

	rows, err := store.database.QueryContext(ctx, `
SELECT name, slot_count, positions, is_starting
FROM league_roster_slots WHERE league_id = ? ORDER BY slot_order`, id)
	if err != nil {
		return league.Configuration{}, false, fmt.Errorf("query league roster slots: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var slot league.RosterSlot
		var positionsJSON string
		if err = rows.Scan(&slot.Name, &slot.Count, &positionsJSON, &slot.IsStarting); err != nil {
			return league.Configuration{}, false, fmt.Errorf("scan league roster slot: %w", err)
		}
		if err = json.Unmarshal([]byte(positionsJSON), &slot.Positions); err != nil {
			return league.Configuration{}, false, fmt.Errorf("decode roster slot positions: %w", err)
		}
		configuration.Rules.RosterSlots = append(configuration.Rules.RosterSlots, slot)
	}
	if err = rows.Err(); err != nil {
		return league.Configuration{}, false, fmt.Errorf("iterate league roster slots: %w", err)
	}
	return configuration, true, nil
}

func (store *DraftEventStore) SaveLeague(ctx context.Context, configuration league.Configuration) error {
	if err := configuration.Validate(); err != nil {
		return err
	}
	scoringJSON, err := json.Marshal(configuration.Rules.ScoringRules)
	if err != nil {
		return fmt.Errorf("encode league scoring rules: %w", err)
	}
	sourcePreferencesJSON, err := json.Marshal(configuration.Rules.SourcePreferences)
	if err != nil {
		return fmt.Errorf("encode ranking source preferences: %w", err)
	}
	recommendationJSON, err := json.Marshal(configuration.Recommendation)
	if err != nil {
		return fmt.Errorf("encode recommendation policy: %w", err)
	}
	draftSettingsJSON, err := json.Marshal(newLeagueDraftSettings(configuration.Rules))
	if err != nil {
		return fmt.Errorf("encode league draft settings: %w", err)
	}

	transaction, err := store.database.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin save league: %w", err)
	}
	defer transaction.Rollback()
	now := time.Now().UTC().Format(time.RFC3339Nano)
	_, err = transaction.ExecContext(ctx, `
INSERT INTO leagues (id, name, team_count, draft_position, draft_type, scoring_rules, source_preferences, recommendation_policy, draft_settings, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(id) DO UPDATE SET
  name = excluded.name,
  team_count = excluded.team_count,
  draft_position = excluded.draft_position,
  draft_type = excluded.draft_type,
  scoring_rules = excluded.scoring_rules,
  source_preferences = excluded.source_preferences,
  recommendation_policy = excluded.recommendation_policy,
  draft_settings = excluded.draft_settings,
  updated_at = excluded.updated_at`,
		configuration.ID, configuration.Rules.Name, configuration.Rules.TeamCount,
		configuration.Rules.DraftPosition, configuration.Rules.DraftType, string(scoringJSON),
		string(sourcePreferencesJSON), string(recommendationJSON), string(draftSettingsJSON), now, now,
	)
	if err != nil {
		return fmt.Errorf("save league: %w", err)
	}
	if _, err = transaction.ExecContext(ctx, `DELETE FROM league_roster_slots WHERE league_id = ?`, configuration.ID); err != nil {
		return fmt.Errorf("replace league roster slots: %w", err)
	}
	for index, slot := range configuration.Rules.RosterSlots {
		positionsJSON, marshalErr := json.Marshal(slot.Positions)
		if marshalErr != nil {
			return fmt.Errorf("encode roster slot positions: %w", marshalErr)
		}
		_, err = transaction.ExecContext(ctx, `
INSERT INTO league_roster_slots (league_id, slot_order, name, slot_count, positions, is_starting)
VALUES (?, ?, ?, ?, ?, ?)`, configuration.ID, index, slot.Name, slot.Count, string(positionsJSON), slot.IsStarting)
		if err != nil {
			return fmt.Errorf("save league roster slot: %w", err)
		}
	}
	if err = transaction.Commit(); err != nil {
		return fmt.Errorf("commit league: %w", err)
	}
	return nil
}

type leagueDraftSettings struct {
	TeamNames           []string            `json:"teamNames"`
	ConsensusMethod     string              `json:"consensusMethod"`
	PlayerPreferences   map[string]string   `json:"playerPreferences"`
	AuctionBudget       float64             `json:"auctionBudget"`
	AuctionMinimumBid   float64             `json:"auctionMinimumBid"`
	KeeperBudgetSpent   float64             `json:"keeperBudgetSpent"`
	MyKeeperSpend       float64             `json:"myKeeperSpend"`
	KeeperValueRemoved  float64             `json:"keeperValueRemoved"`
	LeagueFormat        league.LeagueFormat `json:"leagueFormat"`
	Season              int                 `json:"season"`
	InitialSeason       int                 `json:"initialSeason"`
	FuturePickSeasons   int                 `json:"futurePickSeasons"`
	RookieDraftRounds   int                 `json:"rookieDraftRounds"`
	AuctionBudgetTrades bool                `json:"auctionBudgetTrades"`
	UserTeamNumber      int                 `json:"userTeamNumber"`
	DraftOrder          []int               `json:"draftOrder"`
	FAABBudget          float64             `json:"faabBudget"`
	FAABTrades          bool                `json:"faabTrades"`
}

func newLeagueDraftSettings(rules league.Rules) leagueDraftSettings {
	return leagueDraftSettings{
		TeamNames:       rules.TeamNames,
		ConsensusMethod: rules.ConsensusMethod, PlayerPreferences: rules.PlayerPreferences,
		AuctionBudget: rules.AuctionBudget, AuctionMinimumBid: rules.AuctionMinimumBid,
		KeeperBudgetSpent: rules.KeeperBudgetSpent, MyKeeperSpend: rules.MyKeeperSpend, KeeperValueRemoved: rules.KeeperValueRemoved,
		LeagueFormat: rules.LeagueFormat, Season: rules.Season, InitialSeason: rules.InitialSeason, FuturePickSeasons: rules.FuturePickSeasons,
		RookieDraftRounds: rules.RookieDraftRounds, AuctionBudgetTrades: rules.AuctionBudgetTrades,
		UserTeamNumber: rules.UserTeamNumber, DraftOrder: rules.DraftOrder,
		FAABBudget: rules.FAABBudget, FAABTrades: rules.FAABTrades,
	}
}

func (settings leagueDraftSettings) apply(rules *league.Rules) {
	rules.TeamNames = settings.TeamNames
	rules.ConsensusMethod, rules.PlayerPreferences = settings.ConsensusMethod, settings.PlayerPreferences
	rules.AuctionBudget, rules.AuctionMinimumBid = settings.AuctionBudget, settings.AuctionMinimumBid
	rules.KeeperBudgetSpent, rules.MyKeeperSpend, rules.KeeperValueRemoved = settings.KeeperBudgetSpent, settings.MyKeeperSpend, settings.KeeperValueRemoved
	rules.LeagueFormat, rules.Season, rules.InitialSeason = settings.LeagueFormat, settings.Season, settings.InitialSeason
	rules.FuturePickSeasons, rules.RookieDraftRounds = settings.FuturePickSeasons, settings.RookieDraftRounds
	rules.AuctionBudgetTrades = settings.AuctionBudgetTrades
	rules.UserTeamNumber, rules.DraftOrder = settings.UserTeamNumber, settings.DraftOrder
	rules.FAABBudget, rules.FAABTrades = settings.FAABBudget, settings.FAABTrades
}

func (store *DraftEventStore) DeleteLeague(ctx context.Context, id string) (bool, error) {
	transaction, err := store.database.BeginTx(ctx, nil)
	if err != nil {
		return false, fmt.Errorf("begin delete league: %w", err)
	}
	defer transaction.Rollback()
	if _, err = transaction.ExecContext(ctx, `DELETE FROM draft_events WHERE league_id = ?`, id); err != nil {
		return false, fmt.Errorf("delete league draft events: %w", err)
	}
	if _, err = transaction.ExecContext(ctx, `DELETE FROM draft_pick_trades WHERE league_id = ?`, id); err != nil {
		return false, fmt.Errorf("delete league draft trades: %w", err)
	}
	if _, err = transaction.ExecContext(ctx, `DELETE FROM league_roster_slots WHERE league_id = ?`, id); err != nil {
		return false, fmt.Errorf("delete league roster slots: %w", err)
	}
	result, err := transaction.ExecContext(ctx, `DELETE FROM leagues WHERE id = ?`, id)
	if err != nil {
		return false, fmt.Errorf("delete league: %w", err)
	}
	deleted, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("read deleted league count: %w", err)
	}
	if err = transaction.Commit(); err != nil {
		return false, fmt.Errorf("commit delete league: %w", err)
	}
	return deleted > 0, nil
}
