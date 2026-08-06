CREATE TABLE IF NOT EXISTS draft_events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    league_id TEXT NOT NULL,
    player_id TEXT NOT NULL,
    action TEXT NOT NULL CHECK (action IN ('draft', 'taken', 'undo')),
    target_event_id INTEGER,
    created_at TEXT NOT NULL,
    FOREIGN KEY (target_event_id) REFERENCES draft_events(id)
);

CREATE INDEX IF NOT EXISTS draft_events_league_id_id
    ON draft_events (league_id, id);
