ALTER TABLE ranking_entries ADD COLUMN games INTEGER NOT NULL DEFAULT 0 CHECK (games >= 0);
ALTER TABLE ranking_entries ADD COLUMN bye_week INTEGER NOT NULL DEFAULT 0 CHECK (bye_week >= 0);
ALTER TABLE ranking_entries ADD COLUMN floor_projection REAL NOT NULL DEFAULT 0;
ALTER TABLE ranking_entries ADD COLUMN consensus_projection REAL NOT NULL DEFAULT 0;
ALTER TABLE ranking_entries ADD COLUMN source_projection REAL NOT NULL DEFAULT 0;
ALTER TABLE ranking_entries ADD COLUMN ceiling_projection REAL NOT NULL DEFAULT 0;
ALTER TABLE ranking_entries ADD COLUMN source_value REAL NOT NULL DEFAULT 0;
ALTER TABLE ranking_entries ADD COLUMN injury_risk REAL NOT NULL DEFAULT 0;
ALTER TABLE ranking_entries ADD COLUMN schedule_strength REAL NOT NULL DEFAULT 0;
