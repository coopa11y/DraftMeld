# Changelog

All notable DraftMeld changes will be documented here. DraftMeld follows Semantic Versioning.

## [0.3.0] - 2026-08-07

### Added

- Role-aware Consensus v2 with list-depth normalization, conservative missing-rank handling, weighted median, trimmed mean, and weighted average
- Per-player coverage, disagreement range, and confidence evidence
- Granular projection CSV imports scored with each league's custom rules
- Real consensus players on the live draft board with dedicated, FLEX, and SUPERFLEX replacement allocation
- Projection-based VOR, positional tiers, auction values, budget tracking, and live inflation
- Persistent target and avoid lists with recommendation influence
- Explainable next-pick availability, positional-run, tier-cliff, roster-need, VOR, ADP, and auction reasons
- Deterministic mock opponents and simulate-to-next-turn workflow
- Read-only Sleeper pick synchronization using the public draft endpoint
- Persistent player-identity review queue for uncertain matches
- Interactive projection CSV column mapping with common-header detection
- Explicit canonical-player alias merging across rankings and projections
- Sleeper reconciliation counts, remote deletion/change handling, and optional 15-second polling
- Configurable minimum bids, keeper spend/value calibration, and maximum legal bid protection
- Golden intelligence scenarios and an API-level draft workflow test backed by temporary SQLite
- Versioned league configuration backups with non-destructive restore
- Consensus ranking CSV exports with source weights, evidence, and spreadsheet-formula protection
- Current draft result exports in CSV and versioned JSON formats
- Automated Windows and Linux release archives, checksums, provenance attestations, GHCR images, and smoke tests
- Reusable private ranking CSV imports with interactive column mapping, optional ADP and tiers, and persistent source metadata
- Persistent canonical player directory with stable DraftMeld IDs, source aliases, and provider-specific player IDs
- Named league teams, enforced pick ownership, opponent rosters, and accessible snake, linear, and auction draft-room views
- Accessible manual draft board with overall and position rankings
- Draft and Taken actions with descriptive screen-reader labels
- My Team, recommendations, and draft-history panels
- Persistent SQLite draft-event storage and Undo support
- Dynamic recommendation updates for roster needs, ADP value, and positional scarcity
- Skip links, live announcements, keyboard focus recovery, forced-colors support, and reduced-motion support
- Automated axe, keyboard interaction, API, domain, and SQLite tests
- Persistent multi-league setup with create, edit, duplicate, delete, and active-league selection
- Custom team count, draft position, format, roster slots, and scoring controls with beginner-friendly presets
- Draft-position-first setup with optional opponent names and automatic team placeholders
- Separate league, draft, and team settings plus current-pick reassignment for draft-day trades
- On-demand imports for five ranking signals with source URLs, licenses, timestamps, and record counts
- Current CBS Sports PPR Top 200 connector that retains attribution without redistributing raw data
- Weighted consensus preview anchored to the current redraft player pool
- Ranking-source and consensus REST endpoints with generated TypeScript types
- Private, in-memory imports for user-supplied ESPN PPR and dynasty ranking PDFs
- Reusable PDF extraction and provider-adapter boundary for adding future ranking sheets
- Position-only draft views for QB, RB, WR, TE, K, and DST with position-relative ranks
- Accessible per-league inclusion and influence controls for every ranking source
- Compact disabled-source watchlist for players ranked meaningfully above the active consensus

### Changed

- DraftMeld now separates ordinal rankings, contextual market and usage signals, and granular projections
- Common player suffixes and defense aliases normalize into more stable canonical identities
- The ranking workspace now exposes consensus methodology, source roles, projection inputs, identity exceptions, and uncertainty
- Advanced league settings now persist consensus method, player preferences, and auction calibration together with the existing league configuration
- League configuration now drives roster needs and recommendation policy
- OpenAPI now generates the frontend API types used by a typed fetch client
- Database migrations are ordered, transactional, and tracked in the database
- Draft UI and backend application logic are split into cohesive feature modules
- Shared CSS tokens and a root `npm run verify` contributor workflow
- Draft configuration is loaded from persisted league rules instead of startup-only configuration
- Ranking refreshes normalize provider CSV formats and replace each source atomically
- Ranking downloads use an identifying user agent, bounded response size, timeout, and one transient-failure retry
- Defense rankings now share one team-based identity across provider names, abbreviations, and position labels
- Weighted consensus uses each league's enabled sources and permits equal influence values
- Excluded sources retain their saved weight and continue powering outlier insights

## [0.1.0] - 2026-08-05

### Added

- Public AGPL-3.0 project foundation and product brief
- React and TypeScript frontend project
- Go backend with a versioned REST health endpoint
- Initial league-rule validation and weighted-consensus domain logic
- OpenAPI contract
- Native Windows/Linux build scripts with embedded frontend assets
- Docker and Compose definitions
- GitHub Actions validation for the frontend, backend, and container image
