# DraftMeld

> Every ranking. Every rule. One draft board.

DraftMeld is an open-source fantasy football draft command center. It combines rankings, projections, average draft position, league rules, roster construction, and live draft state into one explainable board.

## Project status

DraftMeld is in its foundation phase. The product model and architecture are being established before the first application milestone.

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

## Repository layout

```text
apps/
  web/          DraftMeld web application
packages/
  core/         League rules, rankings, and draft intelligence
  connectors/   Ranking and fantasy-platform integrations
docs/           Product and architecture decisions
```

See the [product brief](docs/product-brief.md) and [architecture overview](docs/architecture.md) for the initial direction.

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
