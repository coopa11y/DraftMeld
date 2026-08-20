# Ranking sources

DraftMeld starts with transparent ranking signals from open-data projects and public provider pages. CBS, ESPN PPR, Yahoo Standard, the league-matched Draft Sharks preset, and league-matched Sleeper ADP are retrieved on demand. Sleeper's raw offensive projections are also scored with the active league rules. ESPN PPR and dynasty PDFs remain user-supplied fallbacks. Proprietary source files are never committed or redistributed by DraftMeld.

Users can also add any number of private ordinal ranking CSVs. These sources are named by the user, mapped interactively, stored only as normalized records, and exposed to the same per-league inclusion and influence controls as built-in sources.

| DraftMeld source | Signal | Project and license |
| --- | --- | --- |
| Redraft PPR expert consensus | Current PPR expert consensus across supported fantasy positions | [DynastyProcess data](https://github.com/dynastyprocess/data), GPL-3.0 repository with upstream FantasyPros attribution |
| Dynasty market - 1 QB | DynastyProcess normalized `value_1qb` | [DynastyProcess data](https://github.com/dynastyprocess/data), GPL-3.0 |
| Dynasty market - Superflex | DynastyProcess normalized `value_2qb` | [DynastyProcess data](https://github.com/dynastyprocess/data), GPL-3.0 |
| Expected opportunity | Prior-season ffopportunity expected fantasy points aggregated by player | [ffopportunity](https://github.com/ffverse/ffopportunity), CC-BY-SA-4.0 data |
| CBS Sports PPR Top 200 | Current CBS Fantasy Experts consensus order | [CBS Sports rankings](https://www.cbssports.com/fantasy/football/rankings/), proprietary; on-demand personal retrieval |
| ESPN PPR draft rankings | Current default PPR draft order and ADP from ESPN's public fantasy player service | [ESPN Fantasy Football](https://www.espn.com/fantasy/football/), proprietary; on-demand personal retrieval |
| ESPN PPR Top 300 PDF | Overall PPR rank for QB, RB, WR, TE, K, and DST entries | [ESPN Fantasy Football](https://www.espn.com/fantasy/football/), proprietary; user-supplied PDF |
| ESPN Dynasty PDF | Overall dynasty rank for supported positions present in the sheet | [ESPN Fantasy Football](https://www.espn.com/fantasy/football/), proprietary; user-supplied PDF |
| Yahoo default Standard Top 200 | Public default Yahoo pre-draft order for Standard leagues | [Yahoo Fantasy Football](https://football.fantasysports.yahoo.com/f1/public_prerank), proprietary; on-demand personal retrieval |
| Draft Sharks preset | One of eight public Standard, Half-PPR, PPR, or TE Premium and 1QB/Superflex variants; includes tier, ADP, floor, consensus, Draft Sharks, ceiling, 3D Value, injury risk, and schedule strength | [Draft Sharks rankings](https://www.draftsharks.com/rankings), proprietary; on-demand personal retrieval |
| Sleeper ADP | The closest public redraft/dynasty, Standard/Half-PPR/PPR, and 1QB/Superflex community draft market | [Sleeper Fantasy Football](https://sleeper.com/fantasy-football); public undocumented feed, noncommercial use unless separately licensed |
| Sleeper statistical projections | Raw QB, RB, WR, and TE passing, rushing, and receiving projections recalculated with the active league's supported scoring rules | [Sleeper API documentation](https://docs.sleeper.com/); public undocumented projection feed, noncommercial use unless separately licensed |

## Import behavior

- Online data is fetched from fixed public source URLs only when a user selects **Refresh online sources** or an individual **Update now** action. The league-scoped refresh downloads enabled sources, the matching Draft Sharks and Sleeper ADP profiles, and Sleeper's raw offensive projections; PDF sources are excluded.
- PDF uploads are limited to 20 MiB and 200 pages for selectable text. PDFs with no selectable text can use local English OCR for up to 25 pages. Temporary OCR files are removed after each attempt, and DraftMeld stores only normalized player records and source status.
- Ranking CSV uploads are limited to 10 MiB. Player name, overall rank, and position are required; team, ADP, and tier are optional. Reimporting the same source name replaces that source atomically.
- The importer detects a supported provider and document type from the extracted document text. Users do not need to choose column mappings or a parser.
- Provider columns are converted to a common record containing source, normalized player key, display name, position, team, and ordinal rank. Optional tier, ADP, provider ID, and projection evidence are retained when supplied.
- Team defenses use canonical NFL team identities, so values such as `DEN`, `Denver Defense`, and `Broncos D/ST` contribute to the same consensus player.
- Each source is replaced transactionally. A later source failing does not roll back sources that refreshed successfully earlier in the same request.
- Ordinal ranking sources, including private CSVs, define draft-board eligibility. Contextual market and usage feeds enrich that pool without introducing prior-season or dynasty-only names by themselves.
- The UI exposes methodology, license, project link, default weight, publication date, refresh time, record count, and the source's fit for the active league.
- New leagues receive source settings matched to league format, reception scoring, and one-QB versus Superflex/2-QB roster demand. Users can review and reapply those recommendations without overwriting preferences until they save.
- Redraft PPR rankings are primary for a redraft PPR league. Prior-season opportunity is lower-weight context. Dynasty sources are excluded from redraft leagues, and only the matching 1-QB or Superflex market is primary in dynasty leagues.
- Each league can include or exclude a source and assign a positive influence from 0.1 to 10. At least one source must remain included.
- A player's blended score uses the selected robust consensus method across every included, imported source that ranks that player. Higher influence gives a source more pull, while equal values provide equal influence.
- Excluded sources retain their weight. DraftMeld compares them with the active consensus and shows at most five players when an excluded source ranks them at least 10 spots higher, or ranks an otherwise-missing player in its top 50.

Ordinal rankings reflect the scoring assumptions of their publishers. Draft Sharks' preset projection range is retained as provider evidence, but DraftMeld does not pretend those pre-scored totals are raw statistics. Exact custom scoring is applied to raw projection sources, including Sleeper and user-imported CSVs. Sleeper's public feed does not currently expose enough granular kicking and team-defense distributions for trustworthy custom scoring, so K and DST remain covered by its ADP signal rather than its raw projection source.

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
- **Yahoo Fantasy:** the public default Standard Top 200 is supported without authentication. Yahoo states that a team's default pre-draft order is based on its league settings; importing that league-specific order requires the official OAuth flow and a registered Yahoo application.
- **NFL.com:** not included in consensus yet because its public overall draft table still identifies the prior-season board. DraftMeld will not treat that as a current 2026 ranking.
- **ESPN:** the current PPR draft order is refreshed online with ESPN player IDs, positions, teams, and ADP. PPR Top 300 and Dynasty Cheat Sheet PDFs remain supported as user-supplied fallbacks, and DraftMeld never retains the uploaded source files.
- **Draft Sharks:** the public 250-player tables are supported for Standard, Half-PPR, PPR, and TE Premium in both 1QB and Superflex. DraftMeld shows only the closest preset for the active league. Subscriber-only league sync is not scraped; a future private-import path can accept an export obtained by the authorized user.
- **Sleeper:** DraftMeld derives an ordinal market ranking from the matching public ADP field and separately imports raw offensive season projections. The projection endpoint is public but not part of Sleeper's documented API contract, so the adapter is isolated, validates minimum record counts, and preserves CSV projections as a fallback. Sleeper's documented API permits noncommercial use; commercial distribution requires contacting Sleeper.
