CREATE TABLE player_team_observations (
    player_id TEXT NOT NULL REFERENCES canonical_players(id),
    source TEXT NOT NULL,
    nfl_team TEXT NOT NULL,
    confidence INTEGER NOT NULL,
    observed_at INTEGER NOT NULL,
    PRIMARY KEY (player_id, source)
);

CREATE INDEX player_team_observations_winner_idx
    ON player_team_observations(player_id, observed_at DESC, confidence DESC, source);
