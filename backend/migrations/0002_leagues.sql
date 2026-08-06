CREATE TABLE IF NOT EXISTS leagues (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    team_count INTEGER NOT NULL,
    draft_position INTEGER NOT NULL,
    draft_type TEXT NOT NULL,
    scoring_rules TEXT NOT NULL,
    recommendation_policy TEXT NOT NULL,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS league_roster_slots (
    league_id TEXT NOT NULL,
    slot_order INTEGER NOT NULL,
    name TEXT NOT NULL,
    slot_count INTEGER NOT NULL,
    positions TEXT NOT NULL,
    is_starting INTEGER NOT NULL,
    PRIMARY KEY (league_id, slot_order),
    FOREIGN KEY (league_id) REFERENCES leagues(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_league_roster_slots_league
ON league_roster_slots (league_id, slot_order);
