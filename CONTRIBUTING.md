# Contributing to DraftMeld

DraftMeld is early in development. Contributions should strengthen the product model, data correctness, accessibility, tests, or documentation without obscuring how recommendations are calculated.

Meaningful accepted contributions—including documentation, translations, testing, ranking adapters, and accessibility work—may qualify for access to future official supported builds. This recognition is discretionary and is described in [docs/distribution.md](docs/distribution.md).

## Before contributing

1. Search existing issues and discussions.
2. Open an issue for substantial features or architectural changes.
3. Keep provider integrations legally and technically independent. Do not commit paid or copyrighted ranking data.
4. Include tests for ranking, scoring, identity-matching, and draft-state logic.

## Development expectations

- Use TypeScript for application and shared-package code.
- Keep league rules and ranking calculations independent from the UI.
- Treat accessibility as a release requirement.
- Prefer native HTML semantics, visible focus, programmatic form labels, and concise accessible names. Do not add redundant ARIA or live regions to frequently changing values.
- Preserve keyboard, mouse, touch, and responsive behavior when changing an interactive component. Test focus placement and return for view changes, dialogs, and menus.
- Fix repeated behavior in the shared component or helper that owns it. Audit every consumer before changing a shared default.
- Extract a component or function when it represents an independent responsibility or a repeated pattern; do not split code merely to reduce line counts.
- Use repository-pinned Node 24 and Go 1.26 runtimes and install npm dependencies with `npm ci`.
- Document new data sources, required credentials, rate limits, and failure modes.
- Never log provider secrets or private league data.

## Pull requests

Keep pull requests focused. Explain what changed, why it changed, how it was tested, and any user-visible or data-model impact.

DraftMeld uses a two-stage branch flow:

1. Create a feature branch from `dev` and open the feature pull request back to `dev`.
2. Keep `dev` green while related work is integrated and tested together.
3. Open a release pull request from `dev` to `main`; `main` represents releasable code.

Run `npm run verify` from the repository root before opening a pull request. Generated API files must be committed whenever `contracts/openapi.yaml` changes.

Before committing, also review `git status`, the complete diff, and `git diff --check`. Keep Unix LF line endings and stage only intended paths. Automated accessibility checks supplement, but do not replace, keyboard and screen-reader testing.

DraftMeld defaults to draft pull requests. Verify the base branch, title, description, checklist, and initial CI jobs after publishing. Do not mark a pull request ready or merge it until its intended reviewer has approved that action.
