# DraftMeld

> Every ranking. Every rule. One draft board.

DraftMeld is an open-source fantasy football draft command center. It combines rankings, projections, average draft position, league rules, roster construction, and live draft state into one explainable board.

Current prerelease version: **0.3.0**

## Project status

DraftMeld is in active `0.3.0` development. It includes an accessible live board powered by normalized multi-source consensus, mapped projection imports, roster-aware VOR and tiers, targets and avoids, mock opponents, keeper-aware auction tracking, canonical alias review, and read-only reconciled Sleeper synchronization. Five downloadable feeds and two user-supplied ESPN PDF formats are currently supported.

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
- Accessible draft-day and dynasty trades for picks, players, and configured auction/FAAB budgets
- Guided dynasty season rollover with permanent franchise identities and season-specific draft order
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

The remaining release-readiness work is tracked in the [feature-completeness roadmap](docs/roadmap.md).

The formulas, projection CSV schema, draft-day integrations, and current limitations are documented in [draft intelligence](docs/draft-intelligence.md).

Accessibility requirements and current limitations are documented in [docs/accessibility.md](docs/accessibility.md).

Code ownership boundaries, source-size limits, coverage floors, and review expectations are documented in the [maintainability guide](docs/maintainability.md).

## Development

Run the frontend development server:

```bash
cd frontend
npm ci
npm run dev
```

DraftMeld targets Node.js 24 and Go 1.26. The repository root pins the supported Node major and npm package-manager line; CI and Docker use the same runtime families.

Run the backend after installing Go:

```bash
cd backend
go run ./cmd/draftmeld
```

The frontend proxies `/api` requests to `http://localhost:8080`. A production build compiles the frontend into the Go executable.

On first launch, DraftMeld creates a customizable demo league. The resumable setup guide can import common league, draft, roster, auction, dynasty, FAAB, and scoring settings from a PDF or CSV before creation. Use **Manage leagues** to create, edit, duplicate, or delete leagues and switch the active draft board.

The **Export and backup center** under Manage leagues downloads versioned league backups, consensus ranking CSVs, and draft results. Restores always create a new league rather than overwriting existing data. See [data portability](docs/data-portability.md) for formats and compatibility guarantees.

PDF imports read selectable text first and use local OCR for scanned pages when Poppler and Tesseract are available. The Docker image includes both tools. See [local PDF OCR](docs/ocr.md) for native setup, limits, and privacy behavior.

Use **Ranking sources** to review each feed's method, license, weight, freshness, and project link before refreshing the local data or privately importing a supported PDF. See [docs/ranking-sources.md](docs/ranking-sources.md) for the source set, supported PDF formats, and current matching limitations.

Run the same contract, type, unit-test, vet, and production-build checks used for pull requests:

```bash
npm run verify
```

Run the published community image with Docker:

```bash
docker compose -f deployments/compose.yaml up -d
```

To build the same image from your checkout, add the build override:

```bash
docker compose -f deployments/compose.yaml -f deployments/compose.build.yaml up --build
```

Tagged releases publish versioned multi-architecture images in GitHub Container Registry. Signed Windows and Linux packages are planned as optional official supported builds; source, Docker, and all product features remain free. See the [community and official distribution policy](docs/distribution.md) and [release guide](docs/releasing.md).

## Principles

1. **League rules are data.** Recommendations must come from the actual league configuration.
2. **Consensus must be inspectable.** Every combined ranking should show its sources, weights, and uncertainty.
3. **Draft state belongs to the user.** Draft history must persist and be exportable.
4. **Integrations degrade gracefully.** Manual CSV import remains a first-class path.
5. **Intelligence should explain itself.** Users should be able to understand why a player is recommended.

## Contributing

DraftMeld welcomes ideas and contributions. Read [CONTRIBUTING.md](CONTRIBUTING.md) before opening a change. Please report security concerns through the process in [SECURITY.md](SECURITY.md).

Accepted code, documentation, translation, testing, ranking-adapter, and accessibility work may qualify for contributor access to future official supported builds. The criteria are documented in the [distribution policy](docs/distribution.md).

## License

DraftMeld is licensed under the GNU Affero General Public License v3.0. See [LICENSE](LICENSE).

Use of the DraftMeld name and official-build designation is covered by [TRADEMARKS.md](TRADEMARKS.md).

## Versioning

DraftMeld follows Semantic Versioning. While the project remains below `1.0.0`, minor releases may contain breaking changes. Patch releases contain compatible fixes to the current minor line.
