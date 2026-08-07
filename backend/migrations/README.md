# Database migrations

Versioned SQLite migrations live here and are embedded into the DraftMeld executable. They run in filename order, each in its own transaction, and applied filenames are recorded in `schema_migrations`.

Use a zero-padded numeric prefix such as `0002_leagues.sql`. Migrations must be forward-only in released versions and covered by upgrade tests; add a new migration instead of editing one that may already have run.

`0002_leagues.sql` adds persisted league rules and ordered roster slots. Deleting a league through the application also removes that league's draft-event history in the same transaction.

`0003_ranking_sources.sql` adds ranking-source provenance and normalized ranking entries. Replacing one source's entries and metadata occurs in one transaction.

`0004_league_source_weights.sql` adds per-league ranking-source influence settings without changing existing league behavior.
