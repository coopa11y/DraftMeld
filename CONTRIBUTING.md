# Contributing to DraftMeld

DraftMeld is early in development. Contributions should strengthen the product model, data correctness, accessibility, tests, or documentation without obscuring how recommendations are calculated.

## Before contributing

1. Search existing issues and discussions.
2. Open an issue for substantial features or architectural changes.
3. Keep provider integrations legally and technically independent. Do not commit paid or copyrighted ranking data.
4. Include tests for ranking, scoring, identity-matching, and draft-state logic.

## Development expectations

- Use TypeScript for application and shared-package code.
- Keep league rules and ranking calculations independent from the UI.
- Treat accessibility as a release requirement.
- Document new data sources, required credentials, rate limits, and failure modes.
- Never log provider secrets or private league data.

## Pull requests

Keep pull requests focused. Explain what changed, why it changed, how it was tested, and any user-visible or data-model impact.

DraftMeld uses a two-stage branch flow:

1. Create a feature branch from `dev` and open the feature pull request back to `dev`.
2. Keep `dev` green while related work is integrated and tested together.
3. Open a release pull request from `dev` to `main`; `main` represents releasable code.

Run `npm run verify` from the repository root before opening a pull request. Generated API files must be committed whenever `contracts/openapi.yaml` changes.
