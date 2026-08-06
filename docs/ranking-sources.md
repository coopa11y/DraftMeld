# Ranking sources

DraftMeld starts with seven transparent ranking signals. Four come from two open-data ecosystems; CBS is retrieved from its public provider page on demand. The two ESPN signals are imported only from PDFs supplied by the user. Proprietary source files are never committed or redistributed by DraftMeld.

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
- PDF uploads are limited to 20 MiB and 200 pages, processed in memory, and discarded immediately after text extraction. DraftMeld stores only normalized player records and source status.
- The importer detects a supported provider and document type from the extracted document text. Users do not need to choose column mappings or a parser.
- Provider columns are converted to a small common record: source, normalized player key, display name, position, team, and ordinal rank.
- Each source is replaced transactionally. A later source failing does not roll back sources that refreshed successfully earlier in the same request.
- The current redraft feed anchors eligibility so prior-season or dynasty-only names cannot enter the draft board by themselves.
- The UI exposes methodology, license, project link, default weight, publication date, refresh time, and record count.

## Reusable PDF adapter design

PDF mechanics live in `backend/internal/document`; ranking-specific interpretation lives in `backend/internal/application`. A new site or document layout requires a small adapter implementing the internal `pdfRankingParser` contract:

1. Detect the provider and document type using stable title or publisher markers.
2. Convert extracted text into normalized `ranking.Record` values.
3. Register a source definition with `ImportMode: "pdf-upload"` and register the parser in `defaultPDFRankingParsers`.
4. Add synthetic parser tests and an optional local-file integration test. Never commit proprietary fixtures.

This separation keeps file validation, size/page limits, panic recovery, and text extraction shared across every provider. Provider adapters remain focused on one document layout.

## Known limitations

The first matcher uses a normalized player name because these feeds do not share one universal identifier. Suffixes, name changes, and collisions can prevent a valid match. Scanned PDFs are rejected because DraftMeld does not bundle OCR. ESPN's projection guide and positional-only PPR sheet are intentionally rejected because they do not provide the supported overall-ranking layout. Future work will introduce a canonical player table, provider identifiers, a review queue for uncertain matches, league-specific source weights, and normalized/robust consensus methods.

Source terms and upstream availability can change. Maintainers should verify licenses and attribution before adding a connector, and should never commit or redistribute paid rankings.

## Platform coverage

- **CBS Sports:** connected to the current public PPR Top 200. The parser deliberately takes the first consensus group and ignores the separate expert lists later in the page.
- **Yahoo Fantasy:** requires the official OAuth 2.0 API because public pre-draft pages are tied to individual league settings. Integration can proceed after the DraftMeld application is approved and credentials can be configured securely.
- **NFL.com:** not included in consensus yet because its public overall draft table still identifies the prior-season board. DraftMeld will not treat that as a current 2026 ranking.
- **ESPN:** PPR Top 300 and Dynasty Cheat Sheet PDFs are supported as user-supplied imports. DraftMeld does not automatically retrieve or retain the source files.
