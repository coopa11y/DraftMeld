# Maintainability guide

DraftMeld favors small, cohesive feature modules over layers created only for abstraction. Shared code belongs at the lowest layer that can own its meaning without importing a higher layer.

## Backend boundaries

- `domain` owns business data, validation, and model-level operations such as deep copies.
- `application` coordinates use cases. Draft orchestration, trade validation, asset ownership, and budget accounting live in separate files even though they share one `DraftService`.
- `persistence` translates domain records to and from SQLite and owns transactional integrity.
- `api` translates HTTP requests and responses. Route registration is grouped by resource: leagues, rankings, projections, draft actions, draft sessions/seasons, trades, and exports.
- `document` owns reusable PDF validation, selectable-text extraction, and bounded OCR. Provider parsers remain in the application layer.

HTTP JSON bodies are decoded through one strict generic helper. Unknown fields, multiple JSON values, malformed input, and oversized requests are rejected consistently. Service errors are mapped centrally to stable HTTP statuses and user-facing messages.

`league.Rules.Clone` is the canonical deep-copy operation for mutable league data. Repositories and services must use it rather than duplicating slice and map copy loops.

## Frontend boundaries

- `shared/ui` contains reusable visual and accessibility primitives.
- `shared/api` contains transport-only functions and generated contract types.
- `shared/domain` contains framework-independent operations on shared application data.
- `app` components own feature behavior. Large features are split by user responsibility, not by arbitrary markup fragments.

The draft trade feature, for example, separates its composer, asset selector, ledger, shared summary controls, and pure label/key transformations. Pure transformations receive direct unit tests; rendered behavior remains covered through interaction and accessibility tests.

## Automated guardrails

`npm run verify` enforces:

- generated OpenAPI type consistency;
- Prettier, ESLint, TypeScript, gofmt, and Go vet;
- a 500-line maximum for maintained production Go files;
- a 425-line maximum for maintained production TypeScript and TSX files;
- a 250-line maximum for maintained TypeScript and TSX functions and components, excluding blank lines and comments;
- backend statement coverage of at least 68%;
- frontend coverage of at least 80% statements, 73% branches, 80% functions, and 82% lines;
- backend, frontend, SQLite integrity, accessibility, and production-build tests.

Tests and generated files are excluded from source-size and function-size limits. Generated OpenAPI code and test harness files are excluded from frontend coverage because they are not maintained production logic.

Coverage thresholds are regression floors, not finish lines. Changed business logic should cover successful behavior, expected validation failures, dependency failures where practical, and preservation of stored data after a failed operation. A higher percentage does not replace assertions about data integrity.

## Review checklist

Before merging a feature:

1. Put new behavior in the owning domain or feature module rather than the nearest existing file.
2. Search for an existing shared operation before adding a helper.
3. Prefer a concrete shared function; introduce a generic only when it removes repeated type-independent behavior.
4. Preserve the OpenAPI contract or version an intentional breaking change.
5. Add happy-path, failure-path, and integrity assertions for changed business behavior.
6. Run `npm run verify` and rendered keyboard/accessibility QA.
