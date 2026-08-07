CREATE TABLE projection_sources (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    imported_at TEXT NOT NULL,
    record_count INTEGER NOT NULL CHECK (record_count >= 0)
);

CREATE TABLE player_projections (
    source_id TEXT NOT NULL REFERENCES projection_sources(id) ON DELETE CASCADE,
    player_key TEXT NOT NULL,
    player_name TEXT NOT NULL,
    position TEXT NOT NULL,
    nfl_team TEXT NOT NULL,
    bye_week INTEGER NOT NULL DEFAULT 0,
    adp REAL NOT NULL DEFAULT 0,
    stats_json TEXT NOT NULL,
    PRIMARY KEY (source_id, player_key)
);

CREATE INDEX player_projections_player_key_idx ON player_projections(player_key);
