CREATE TABLE ranking_sources (
    id TEXT PRIMARY KEY,
    refreshed_at TEXT NOT NULL,
    published_at TEXT NOT NULL DEFAULT '',
    record_count INTEGER NOT NULL CHECK (record_count >= 0)
);

CREATE TABLE ranking_entries (
    source_id TEXT NOT NULL REFERENCES ranking_sources(id) ON DELETE CASCADE,
    player_key TEXT NOT NULL,
    player_name TEXT NOT NULL,
    position TEXT NOT NULL,
    nfl_team TEXT NOT NULL,
    source_rank INTEGER NOT NULL CHECK (source_rank > 0),
    PRIMARY KEY (source_id, player_key)
);

CREATE INDEX ranking_entries_player_key_idx ON ranking_entries(player_key);
