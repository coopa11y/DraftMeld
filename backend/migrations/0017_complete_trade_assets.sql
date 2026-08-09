ALTER TABLE draft_pick_trades ADD COLUMN team_one_players TEXT NOT NULL DEFAULT '[]';
ALTER TABLE draft_pick_trades ADD COLUMN team_two_players TEXT NOT NULL DEFAULT '[]';
ALTER TABLE draft_pick_trades ADD COLUMN team_one_budgets TEXT NOT NULL DEFAULT '[]';
ALTER TABLE draft_pick_trades ADD COLUMN team_two_budgets TEXT NOT NULL DEFAULT '[]';
