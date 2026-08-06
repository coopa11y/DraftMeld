# Database migrations

Versioned SQLite migrations live here and are embedded into the DraftMeld executable. They run in filename order, each in its own transaction, and applied filenames are recorded in `schema_migrations`.

Use a zero-padded numeric prefix such as `0002_leagues.sql`. Migrations must be forward-only in released versions and covered by upgrade tests; add a new migration instead of editing one that may already have run.
