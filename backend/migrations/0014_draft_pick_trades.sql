CREATE TABLE draft_pick_trades (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    league_id TEXT NOT NULL,
    team_one_number INTEGER NOT NULL,
    team_two_number INTEGER NOT NULL,
    team_one_receives TEXT NOT NULL,
    team_two_receives TEXT NOT NULL,
    created_at TEXT NOT NULL,
    FOREIGN KEY (league_id) REFERENCES leagues(id) ON DELETE CASCADE
);

CREATE INDEX draft_pick_trades_league_idx
    ON draft_pick_trades(league_id, id);
