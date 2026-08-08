package sqlite

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/coopa11y/DraftMeld/backend/internal/domain/draft"
)

func (store *DraftEventStore) ListPickTrades(ctx context.Context, leagueID string) ([]draft.PickTrade, error) {
	rows, err := store.database.QueryContext(ctx, `
SELECT id, league_id, team_one_number, team_two_number, team_one_receives, team_two_receives,
	       season, team_one_future_picks, team_two_future_picks, team_one_auction_budget, team_two_auction_budget, created_at
	       , team_one_players, team_two_players, team_one_budgets, team_two_budgets
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
		var teamOneJSON, teamTwoJSON, teamOneFutureJSON, teamTwoFutureJSON, teamOnePlayersJSON, teamTwoPlayersJSON, teamOneBudgetsJSON, teamTwoBudgetsJSON, created string
		if err = rows.Scan(&trade.ID, &trade.LeagueID, &trade.TeamOneNumber, &trade.TeamTwoNumber, &teamOneJSON, &teamTwoJSON,
			&trade.Season, &teamOneFutureJSON, &teamTwoFutureJSON, &trade.TeamOneAuctionBudget, &trade.TeamTwoAuctionBudget, &created,
			&teamOnePlayersJSON, &teamTwoPlayersJSON, &teamOneBudgetsJSON, &teamTwoBudgetsJSON); err != nil {
			return nil, fmt.Errorf("scan draft pick trade: %w", err)
		}
		if err = json.Unmarshal([]byte(teamOneJSON), &trade.TeamOneReceives); err != nil {
			return nil, fmt.Errorf("decode team one draft picks: %w", err)
		}
		if err = json.Unmarshal([]byte(teamTwoJSON), &trade.TeamTwoReceives); err != nil {
			return nil, fmt.Errorf("decode team two draft picks: %w", err)
		}
		if err = json.Unmarshal([]byte(teamOneFutureJSON), &trade.TeamOneFuturePicks); err != nil {
			return nil, fmt.Errorf("decode team one future picks: %w", err)
		}
		if err = json.Unmarshal([]byte(teamTwoFutureJSON), &trade.TeamTwoFuturePicks); err != nil {
			return nil, fmt.Errorf("decode team two future picks: %w", err)
		}
		if err = json.Unmarshal([]byte(teamOnePlayersJSON), &trade.TeamOnePlayers); err != nil {
			return nil, fmt.Errorf("decode team one players: %w", err)
		}
		if err = json.Unmarshal([]byte(teamTwoPlayersJSON), &trade.TeamTwoPlayers); err != nil {
			return nil, fmt.Errorf("decode team two players: %w", err)
		}
		if err = json.Unmarshal([]byte(teamOneBudgetsJSON), &trade.TeamOneBudgets); err != nil {
			return nil, fmt.Errorf("decode team one budgets: %w", err)
		}
		if err = json.Unmarshal([]byte(teamTwoBudgetsJSON), &trade.TeamTwoBudgets); err != nil {
			return nil, fmt.Errorf("decode team two budgets: %w", err)
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
	oneFuture, err := json.Marshal(trade.TeamOneFuturePicks)
	if err != nil {
		return draft.PickTrade{}, fmt.Errorf("encode team one future picks: %w", err)
	}
	twoFuture, err := json.Marshal(trade.TeamTwoFuturePicks)
	if err != nil {
		return draft.PickTrade{}, fmt.Errorf("encode team two future picks: %w", err)
	}
	onePlayers, err := json.Marshal(trade.TeamOnePlayers)
	if err != nil {
		return draft.PickTrade{}, fmt.Errorf("encode team one players: %w", err)
	}
	twoPlayers, err := json.Marshal(trade.TeamTwoPlayers)
	if err != nil {
		return draft.PickTrade{}, fmt.Errorf("encode team two players: %w", err)
	}
	oneBudgets, err := json.Marshal(trade.TeamOneBudgets)
	if err != nil {
		return draft.PickTrade{}, fmt.Errorf("encode team one budgets: %w", err)
	}
	twoBudgets, err := json.Marshal(trade.TeamTwoBudgets)
	if err != nil {
		return draft.PickTrade{}, fmt.Errorf("encode team two budgets: %w", err)
	}
	trade.CreatedAt = time.Now().UTC()
	result, err := store.database.ExecContext(ctx, `
INSERT INTO draft_pick_trades (
    league_id, team_one_number, team_two_number, team_one_receives, team_two_receives, season,
	    team_one_future_picks, team_two_future_picks, team_one_auction_budget, team_two_auction_budget, created_at,
	    team_one_players, team_two_players, team_one_budgets, team_two_budgets
)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, trade.LeagueID, trade.TeamOneNumber, trade.TeamTwoNumber,
		string(one), string(two), trade.Season, string(oneFuture), string(twoFuture), trade.TeamOneAuctionBudget,
		trade.TeamTwoAuctionBudget, trade.CreatedAt.Format(time.RFC3339Nano), string(onePlayers), string(twoPlayers), string(oneBudgets), string(twoBudgets))
	if err != nil {
		return draft.PickTrade{}, fmt.Errorf("save draft pick trade: %w", err)
	}
	trade.ID, err = result.LastInsertId()
	if err != nil {
		return draft.PickTrade{}, fmt.Errorf("read draft pick trade ID: %w", err)
	}
	return trade, nil
}

func (store *DraftEventStore) UpdatePickTrade(ctx context.Context, trade draft.PickTrade) error {
	oneFuture, err := json.Marshal(trade.TeamOneFuturePicks)
	if err != nil {
		return fmt.Errorf("encode team one future picks: %w", err)
	}
	twoFuture, err := json.Marshal(trade.TeamTwoFuturePicks)
	if err != nil {
		return fmt.Errorf("encode team two future picks: %w", err)
	}
	result, err := store.database.ExecContext(ctx, `UPDATE draft_pick_trades SET team_one_future_picks = ?, team_two_future_picks = ? WHERE id = ? AND league_id = ?`, string(oneFuture), string(twoFuture), trade.ID, trade.LeagueID)
	if err != nil {
		return fmt.Errorf("update draft trade: %w", err)
	}
	updated, err := result.RowsAffected()
	if err != nil || updated != 1 {
		return errors.New("draft trade was not found")
	}
	return nil
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
	UpdatePickTrade(context.Context, draft.PickTrade) error
	DeletePickTrade(context.Context, string, int64) (bool, error)
} = (*DraftEventStore)(nil)
