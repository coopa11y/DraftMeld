ALTER TABLE draft_pick_trades ADD COLUMN season INTEGER NOT NULL DEFAULT 0;
ALTER TABLE draft_pick_trades ADD COLUMN team_one_future_picks TEXT NOT NULL DEFAULT '[]';
ALTER TABLE draft_pick_trades ADD COLUMN team_two_future_picks TEXT NOT NULL DEFAULT '[]';
ALTER TABLE draft_pick_trades ADD COLUMN team_one_auction_budget REAL NOT NULL DEFAULT 0;
ALTER TABLE draft_pick_trades ADD COLUMN team_two_auction_budget REAL NOT NULL DEFAULT 0;

UPDATE draft_pick_trades
SET season = 2026
WHERE season = 0;
