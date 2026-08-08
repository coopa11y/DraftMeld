package sqlite

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/coopa11y/DraftMeld/backend/internal/domain/draft"
)

func (store *DraftEventStore) ListPickTrades(ctx context.Context, leagueID string) ([]draft.PickTrade, error) {
	rows, err := store.database.QueryContext(ctx, `
SELECT id, league_id, team_one_number, team_two_number, team_one_receives, team_two_receives, created_at
FROM draft_pick_trades
WHERE league_id = ?
ORDER BY id`, leagueID)
	if err != nil {
		return nil, fmt.Errorf("query draft pick trades: %w", err)
	}
	defer rows.Close()
	trades := make([]draft.PickTrade, 0)
	for rows.Next() {
		var trade draft.PickTrade
		var teamOneJSON, teamTwoJSON, created string
		if err = rows.Scan(&trade.ID, &trade.LeagueID, &trade.TeamOneNumber, &trade.TeamTwoNumber, &teamOneJSON, &teamTwoJSON, &created); err != nil {
			return nil, fmt.Errorf("scan draft pick trade: %w", err)
		}
		if err = json.Unmarshal([]byte(teamOneJSON), &trade.TeamOneReceives); err != nil {
			return nil, fmt.Errorf("decode team one draft picks: %w", err)
		}
		if err = json.Unmarshal([]byte(teamTwoJSON), &trade.TeamTwoReceives); err != nil {
			return nil, fmt.Errorf("decode team two draft picks: %w", err)
		}
		if trade.CreatedAt, err = time.Parse(time.RFC3339Nano, created); err != nil {
			return nil, fmt.Errorf("parse draft pick trade time: %w", err)
		}
		trades = append(trades, trade)
	}
	return trades, rows.Err()
}

func (store *DraftEventStore) SavePickTrade(ctx context.Context, trade draft.PickTrade) (draft.PickTrade, error) {
	one, err := json.Marshal(trade.TeamOneReceives)
	if err != nil {
		return draft.PickTrade{}, fmt.Errorf("encode team one draft picks: %w", err)
	}
	two, err := json.Marshal(trade.TeamTwoReceives)
	if err != nil {
		return draft.PickTrade{}, fmt.Errorf("encode team two draft picks: %w", err)
	}
	trade.CreatedAt = time.Now().UTC()
	result, err := store.database.ExecContext(ctx, `
INSERT INTO draft_pick_trades (league_id, team_one_number, team_two_number, team_one_receives, team_two_receives, created_at)
VALUES (?, ?, ?, ?, ?, ?)`, trade.LeagueID, trade.TeamOneNumber, trade.TeamTwoNumber, string(one), string(two), trade.CreatedAt.Format(time.RFC3339Nano))
	if err != nil {
		return draft.PickTrade{}, fmt.Errorf("save draft pick trade: %w", err)
	}
	trade.ID, err = result.LastInsertId()
	if err != nil {
		return draft.PickTrade{}, fmt.Errorf("read draft pick trade ID: %w", err)
	}
	return trade, nil
}

func (store *DraftEventStore) DeletePickTrade(ctx context.Context, leagueID string, tradeID int64) (bool, error) {
	result, err := store.database.ExecContext(ctx, `DELETE FROM draft_pick_trades WHERE league_id = ? AND id = ?`, leagueID, tradeID)
	if err != nil {
		return false, fmt.Errorf("delete draft pick trade: %w", err)
	}
	count, err := result.RowsAffected()
	return count > 0, err
}

var _ interface {
	ListPickTrades(context.Context, string) ([]draft.PickTrade, error)
	SavePickTrade(context.Context, draft.PickTrade) (draft.PickTrade, error)
	DeletePickTrade(context.Context, string, int64) (bool, error)
} = (*DraftEventStore)(nil)
