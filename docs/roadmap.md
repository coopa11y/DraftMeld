# Feature-completeness roadmap

DraftMeld is keeping the `v0.3.0` release unpublished while the remaining draft-day workflows are completed.

## Current priorities

1. **Release readiness** — complete a manual NVDA walkthrough before publishing the prerelease.

## Remaining release-readiness work

- Complete a hands-on NVDA walkthrough using realistic draft scenarios.
- Provide clear first-run empty states when no league or ranking source exists.
- Validate PDF and CSV imports against messy real-world files, with actionable recovery guidance for invalid or partial data.
- Preserve and restore interrupted draft sessions and partially completed import configuration.
- Make loading, offline, and partial-failure states understandable and recoverable.
- Run an end-to-end draft simulation that covers pick trades, drafting, marking players gone, and undo.
- Add privacy-conscious diagnostic logging so import and synchronization failures can be investigated.
- Complete release packaging, installation documentation, backup and restore verification, and upgrade migration testing.

These items are ordered by user risk. Accessibility and draft-data integrity should be validated before packaging and distribution work is considered complete.

## Completed initial-release capabilities

- Multiple persistent league profiles
- Custom roster slots and core scoring settings
- Projection CSV mapping and league-scored values
- Weighted, explainable multi-source consensus
- Accessible manual drafting with position views, recommendations, persistence, and undo
- Explicit draft start, completion, reset, and recoverable reset states with active-draft configuration safeguards
- Simple-to-advanced scoring with optional TE premium, turnovers, bonuses, kicking bands, and defense tiers
- Resumable first-run guidance with PDF/CSV league-settings and scoring import plus a full manual setup path
- Ranking, league backup, and draft-result exports
- Generic private ranking CSV imports with reusable column mapping
- Stable canonical player identities, aliases, and provider IDs

This list is intentionally outcome-focused. Implementation details and smaller refinements belong in the pull request that delivers each priority.
