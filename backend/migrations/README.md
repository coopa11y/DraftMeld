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

`0010_custom_ranking_sources.sql` stores user-imported ordinal source metadata plus optional ADP and tier values.

`0011_player_directory.sql` stores stable canonical players, normalized identity keys, and provider-specific player identifiers.

`0019_ranking_projection_evidence.sql` retains optional games, bye week, projection range, source value, injury risk, and schedule-strength evidence attached to ranking records.

`0020_projection_provenance.sql` retains an online projection feed's upstream publication timestamp across restarts.

`0021_player_news.sql` stores configurable news feeds, player-linked articles, and the latest structured availability snapshot.

`0022_custom_ranking_source_profiles.sql` preserves provider and scoring-profile labels for private ranking imports.
