ALTER TABLE ranking_entries ADD COLUMN adp REAL NOT NULL DEFAULT 0 CHECK (adp >= 0);
ALTER TABLE ranking_entries ADD COLUMN tier INTEGER NOT NULL DEFAULT 0 CHECK (tier >= 0);

CREATE TABLE custom_ranking_sources (
    id TEXT PRIMARY KEY REFERENCES ranking_sources(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT NOT NULL,
    methodology TEXT NOT NULL,
    license TEXT NOT NULL,
    project_url TEXT NOT NULL DEFAULT '',
    data_url TEXT NOT NULL DEFAULT '',
    default_weight REAL NOT NULL CHECK (default_weight > 0 AND default_weight <= 10),
    import_mode TEXT NOT NULL CHECK (import_mode = 'csv-upload'),
    role TEXT NOT NULL CHECK (role = 'ranking')
);
