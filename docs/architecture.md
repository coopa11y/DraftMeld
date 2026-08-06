# Architecture overview

DraftMeld will begin as a TypeScript monorepo with a web application and independently testable domain packages.

## Boundaries

### Web application

Owns league setup, imports, draft-board interaction, visualizations, authentication when deployed, and accessibility.

### Core package

Owns scoring rules, roster constraints, consensus algorithms, VOR, tiers, scarcity, draft-state transitions, and recommendation evidence. It must not depend on a browser or a specific fantasy platform.

### Connectors package

Owns provider adapters, CSV column mapping, canonical-player matching inputs, rate limiting, and provider-specific error handling. Raw provider records must not leak into the core model.

## Core entities

- User workspace
- League
- Team entry
- League ruleset
- Canonical player
- Provider player identity
- Ranking source
- Projection source
- Consensus profile
- Draft
- Pick
- Roster

## Consensus model

The first model should support mean rank, median rank, trimmed mean, and weighted rank. Missing players, source coverage, ties, and outliers must be explicit. Projection aggregation and ordinal ranking aggregation remain separate operations.

## Persistence

Local development should use a relational database with migrations. Draft picks are append-only events with compensating undo events, allowing the current board to be reconstructed and audited.

## Integration policy

Connectors should prefer documented APIs and user-supplied exports. Scrapers must be isolated, optional, rate-limited, and accompanied by source-specific tests. Paid source data must never be committed or redistributed.
