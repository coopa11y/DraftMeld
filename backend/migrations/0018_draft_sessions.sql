CREATE TABLE draft_sessions (
    league_id TEXT NOT NULL,
    season INTEGER NOT NULL,
    status TEXT NOT NULL,
    started_at TEXT,
    updated_at TEXT NOT NULL,
    reset_events TEXT NOT NULL DEFAULT '[]',
    reset_status TEXT NOT NULL DEFAULT '',
    PRIMARY KEY (league_id, season),
    FOREIGN KEY (league_id) REFERENCES leagues(id) ON DELETE CASCADE
);
