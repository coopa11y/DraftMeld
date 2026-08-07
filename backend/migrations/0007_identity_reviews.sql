CREATE TABLE identity_reviews (
    issue_key TEXT PRIMARY KEY,
    resolution TEXT NOT NULL CHECK (resolution IN ('confirmed-separate', 'acknowledged')),
    reviewed_at TEXT NOT NULL
);
