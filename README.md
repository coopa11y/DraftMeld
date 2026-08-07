# DraftMeld

> Every ranking. Every rule. One draft board.

DraftMeld is an open-source fantasy football draft command center. It combines rankings, projections, average draft position, league rules, roster construction, and live draft state into one explainable board.

Current development version: **0.3.0**

## Project status

DraftMeld is in active `0.3.0` development. It includes an accessible live board powered by normalized multi-source consensus, projection-based league scoring, roster-aware VOR and tiers, targets and avoids, mock opponents, auction tracking, and read-only Sleeper pick synchronization. Five downloadable feeds and two user-supplied ESPN PDF formats are currently supported.

## Product goals

- Support multiple leagues and teams from one account.
- Model custom scoring, roster slots, keepers, budgets, and draft formats.
- Import rankings and projections through files and provider connectors.
- Normalize provider-specific player records into stable player identities.
- Produce weighted, median, and trimmed-mean consensus rankings.
- Explain recommendations with VOR, tiers, scarcity, ADP value, roster fit, and source disagreement.
- Track live drafts without making picks on a user's behalf.
- Remain useful without an AI provider and transparent when AI is enabled.
- Run locally or as a self-hosted web application.

## Planned formats

- Redraft and dynasty
- Snake and linear drafts
- Auction and salary-cap drafts
- Keeper leagues
- Standard, PPR, half-PPR, superflex, and TE-premium scoring

## Technology

- React and TypeScript frontend
- Go backend and REST API
- OpenAPI contract with generated-client support
- SQLite persistence target
- One native executable with the compiled frontend embedded
- One multi-stage Docker image

## Repository layout

```text
frontend/       React and TypeScript application
backend/        Go API, domain, application, and adapters
contracts/      Language-independent OpenAPI contract
deployments/    Docker and Compose definitions
docs/           Product briefs and architecture decisions
scripts/        Native build entrypoints
```

See the [product brief](docs/product-brief.md) and [architecture overview](docs/architecture.md) for the initial direction.

The formulas, projection CSV schema, draft-day integrations, and current limitations are documented in [draft intelligence](docs/draft-intelligence.md).

Accessibility requirements and current limitations are documented in [docs/accessibility.md](docs/accessibility.md).

## Development

Run the frontend development server:

```bash
cd frontend
npm install
npm run dev
```

Run the backend after installing Go:

```bash
cd backend
go run ./cmd/draftmeld
```

The frontend proxies `/api` requests to `http://localhost:8080`. A production build compiles the frontend into the Go executable.

On first launch, DraftMeld creates a customizable demo league. Use **Manage leagues** to create, edit, duplicate, or delete leagues and switch the active draft board.

Use **Ranking sources** to review each feed's method, license, weight, freshness, and project link before refreshing the local data or privately importing a supported PDF. See [docs/ranking-sources.md](docs/ranking-sources.md) for the source set, supported PDF formats, and current matching limitations.

Run the same contract, type, unit-test, vet, and production-build checks used for pull requests:

```bash
npm run verify
```

Run the complete application with Docker:

```bash
docker compose -f deployments/compose.yaml up --build
```

## Principles

1. **League rules are data.** Recommendations must come from the actual league configuration.
2. **Consensus must be inspectable.** Every combined ranking should show its sources, weights, and uncertainty.
3. **Draft state belongs to the user.** Draft history must persist and be exportable.
4. **Integrations degrade gracefully.** Manual CSV import remains a first-class path.
5. **Intelligence should explain itself.** Users should be able to understand why a player is recommended.

## Contributing

DraftMeld welcomes ideas and contributions. Read [CONTRIBUTING.md](CONTRIBUTING.md) before opening a change. Please report security concerns through the process in [SECURITY.md](SECURITY.md).

## License

DraftMeld is licensed under the GNU Affero General Public License v3.0. See [LICENSE](LICENSE).

## Versioning

DraftMeld follows Semantic Versioning. While the project remains below `1.0.0`, minor releases may contain breaking changes. Patch releases contain compatible fixes to the current minor line.
