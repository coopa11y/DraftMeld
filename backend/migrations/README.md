# Database migrations

Versioned SQLite migrations live here and are embedded into the DraftMeld executable. They run in filename order, each in its own transaction, and applied filenames are recorded in `schema_migrations`.

Use a zero-padded numeric prefix such as `0002_leagues.sql`. Migrations must be forward-only in released versions and covered by upgrade tests; add a new migration instead of editing one that may already have run.

`0002_leagues.sql` adds persisted league rules and ordered roster slots. Deleting a league through the application also removes that league's draft-event history in the same transaction.

`0003_ranking_sources.sql` adds ranking-source provenance and normalized ranking entries. Replacing one source's entries and metadata occurs in one transaction.

`0004_league_source_preferences.sql` adds per-league ranking-source influence and inclusion settings without changing existing league behavior.

`0005_projection_sources.sql` stores user-imported granular statistics separately from ordinal ranking inputs.

`0006_auction_costs.sql` records an optional winning bid on each append-only draft event.

`0007_identity_reviews.sql` persists human decisions from the canonical-player review queue.

`0008_identity_aliases.sql` stores explicit alias-to-canonical player mappings used by rankings and projections.

`0009_league_draft_settings.sql` persists consensus selection, player preferences, and calibrated auction/keeper settings that were previously defaulted at load time.
