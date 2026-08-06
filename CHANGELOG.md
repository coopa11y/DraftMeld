# Changelog

All notable DraftMeld changes will be documented here. DraftMeld follows Semantic Versioning.

## [0.2.0] - Unreleased

### Added

- Accessible manual draft board with overall and position rankings
- Draft and Taken actions with descriptive screen-reader labels
- My Team, recommendations, and draft-history panels
- Persistent SQLite draft-event storage and Undo support
- Dynamic recommendation updates for roster needs, ADP value, and positional scarcity
- Skip links, live announcements, keyboard focus recovery, forced-colors support, and reduced-motion support
- Automated axe, keyboard interaction, API, domain, and SQLite tests
- Persistent multi-league setup with create, edit, duplicate, delete, and active-league selection
- Custom team count, draft position, format, roster slots, and scoring controls with beginner-friendly presets
- On-demand imports for four open ranking signals with source URLs, licenses, timestamps, and record counts
- Weighted consensus preview anchored to the current redraft player pool
- Ranking-source and consensus REST endpoints with generated TypeScript types

### Changed

- League configuration now drives roster needs and recommendation policy
- OpenAPI now generates the frontend API types used by a typed fetch client
- Database migrations are ordered, transactional, and tracked in the database
- Draft UI and backend application logic are split into cohesive feature modules
- Shared CSS tokens and a root `npm run verify` contributor workflow
- Draft configuration is loaded from persisted league rules instead of startup-only configuration
- Ranking refreshes normalize provider CSV formats and replace each source atomically

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
