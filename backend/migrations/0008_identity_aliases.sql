CREATE TABLE identity_aliases (
    alias_key TEXT PRIMARY KEY,
    canonical_key TEXT NOT NULL,
    created_at TEXT NOT NULL,
    CHECK (alias_key <> canonical_key)
);
