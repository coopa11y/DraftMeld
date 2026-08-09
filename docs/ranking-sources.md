# Ranking sources

DraftMeld starts with seven transparent ranking signals. Four come from two open-data ecosystems; CBS is retrieved from its public provider page on demand. The two ESPN signals are imported only from PDFs supplied by the user. Proprietary source files are never committed or redistributed by DraftMeld.

Users can also add any number of private ordinal ranking CSVs. These sources are named by the user, mapped interactively, stored only as normalized records, and exposed to the same per-league inclusion and influence controls as built-in sources.

| DraftMeld source | Signal | Project and license |
| --- | --- | --- |
| Redraft expert consensus | Current overall expert consensus across supported fantasy positions | [DynastyProcess data](https://github.com/dynastyprocess/data), GPL-3.0 repository with upstream FantasyPros attribution |
| Dynasty market - 1 QB | DynastyProcess normalized `value_1qb` | [DynastyProcess data](https://github.com/dynastyprocess/data), GPL-3.0 |
| Dynasty market - Superflex | DynastyProcess normalized `value_2qb` | [DynastyProcess data](https://github.com/dynastyprocess/data), GPL-3.0 |
| Expected opportunity | Prior-season ffopportunity expected fantasy points aggregated by player | [ffopportunity](https://github.com/ffverse/ffopportunity), CC-BY-SA-4.0 data |
| CBS Sports PPR Top 200 | Current CBS Fantasy Experts consensus order | [CBS Sports rankings](https://www.cbssports.com/fantasy/football/rankings/), proprietary; on-demand personal retrieval |
| ESPN PPR Top 300 PDF | Overall PPR rank for QB, RB, WR, TE, K, and DST entries | [ESPN Fantasy Football](https://www.espn.com/fantasy/football/), proprietary; user-supplied PDF |
| ESPN Dynasty PDF | Overall dynasty rank for supported positions present in the sheet | [ESPN Fantasy Football](https://www.espn.com/fantasy/football/), proprietary; user-supplied PDF |

## Import behavior

- Online data is fetched from fixed public source URLs only when a user selects **Refresh all sources**. PDF sources are excluded from automatic refresh.
- PDF uploads are limited to 20 MiB and 200 pages for selectable text. PDFs with no selectable text can use local English OCR for up to 25 pages. Temporary OCR files are removed after each attempt, and DraftMeld stores only normalized player records and source status.
- Ranking CSV uploads are limited to 10 MiB. Player name, overall rank, and position are required; team, ADP, and tier are optional. Reimporting the same source name replaces that source atomically.
- The importer detects a supported provider and document type from the extracted document text. Users do not need to choose column mappings or a parser.
- Provider columns are converted to a small common record: source, normalized player key, display name, position, team, and ordinal rank.
- Team defenses use canonical NFL team identities, so values such as `DEN`, `Denver Defense`, and `Broncos D/ST` contribute to the same consensus player.
- Each source is replaced transactionally. A later source failing does not roll back sources that refreshed successfully earlier in the same request.
- Ordinal ranking sources, including private CSVs, define draft-board eligibility. Contextual market and usage feeds enrich that pool without introducing prior-season or dynasty-only names by themselves.
- The UI exposes methodology, license, project link, default weight, publication date, refresh time, and record count.
- Each league can include or exclude a source and assign a positive influence from 0.1 to 10. Missing settings use the enabled published default, and at least one source must remain included.
- A player's blended score is the weighted average of every included, imported source that ranks that player. Higher influence gives a source more pull, while equal values provide equal influence.
- Excluded sources retain their weight. DraftMeld compares them with the active consensus and shows at most five players when an excluded source ranks them at least 10 spots higher, or ranks an otherwise-missing player in its top 50.

## Reusable PDF adapter design

PDF mechanics live in `backend/internal/document`; ranking-specific interpretation lives in `backend/internal/application`. A new site or document layout requires a small adapter implementing the internal `pdfRankingParser` contract:

1. Detect the provider and document type using stable title or publisher markers.
2. Convert extracted text into normalized `ranking.Record` values.
3. Register a source definition with `ImportMode: "pdf-upload"` and register the parser in `defaultPDFRankingParsers`.
4. Add synthetic parser tests and an optional local-file integration test. Never commit proprietary fixtures.

This separation keeps file validation, size/page limits, panic recovery, selectable-text extraction, and bounded OCR fallback shared across every provider. Provider adapters remain focused on one document layout.

## Known limitations

DraftMeld resolves imported players into a persistent canonical directory before storing new ranking and projection data. Each player receives an opaque stable DraftMeld ID; normalized names remain searchable identity keys, and an optional source player ID can bind renamed records from the same provider. Existing normalized keys are retained as aliases so upgraded databases continue to match prior rankings and draft history. Uncertain matches enter the identity review queue, where a merge redirects name aliases and provider IDs to the selected canonical player. Team defenses remain canonicalized by NFL team. OCR can recover text from scanned pages, but provider adapters still reject unsupported layouts. ESPN's projection guide and positional-only PPR sheet are intentionally rejected because they do not provide the supported overall-ranking layout.

Source terms and upstream availability can change. Maintainers should verify licenses and attribution before adding a connector, and should never commit or redistribute paid rankings.

## Platform coverage

- **CBS Sports:** connected to the current public PPR Top 200. The parser deliberately takes the first consensus group and ignores the separate expert lists later in the page.
- **Yahoo Fantasy:** requires the official OAuth 2.0 API because public pre-draft pages are tied to individual league settings. Integration can proceed after the DraftMeld application is approved and credentials can be configured securely.
- **NFL.com:** not included in consensus yet because its public overall draft table still identifies the prior-season board. DraftMeld will not treat that as a current 2026 ranking.
- **ESPN:** PPR Top 300 and Dynasty Cheat Sheet PDFs are supported as user-supplied imports. DraftMeld does not automatically retrieve or retain the source files.
