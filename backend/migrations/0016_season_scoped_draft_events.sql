ALTER TABLE draft_events ADD COLUMN season INTEGER NOT NULL DEFAULT 0;

UPDATE draft_events
SET season = 2026
WHERE season = 0;

CREATE INDEX IF NOT EXISTS idx_draft_events_league_season
    ON draft_events (league_id, season, id);
