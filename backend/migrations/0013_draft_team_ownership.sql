ALTER TABLE draft_events ADD COLUMN team_number INTEGER NOT NULL DEFAULT 0;

CREATE INDEX draft_events_league_team_idx
    ON draft_events(league_id, team_number, id);
