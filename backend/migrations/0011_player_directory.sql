CREATE TABLE canonical_players (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    position TEXT NOT NULL,
    nfl_team TEXT NOT NULL DEFAULT '',
    merged_into TEXT REFERENCES canonical_players(id),
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    CHECK (merged_into IS NULL OR merged_into <> id)
);

CREATE TABLE player_identity_keys (
    identity_key TEXT PRIMARY KEY,
    player_id TEXT NOT NULL REFERENCES canonical_players(id),
    created_at TEXT NOT NULL
);

CREATE INDEX player_identity_keys_player_id_idx ON player_identity_keys(player_id);

CREATE TABLE player_provider_ids (
    provider TEXT NOT NULL,
    provider_player_id TEXT NOT NULL,
    player_id TEXT NOT NULL REFERENCES canonical_players(id),
    created_at TEXT NOT NULL,
    PRIMARY KEY (provider, provider_player_id)
);

CREATE INDEX player_provider_ids_player_id_idx ON player_provider_ids(player_id);
