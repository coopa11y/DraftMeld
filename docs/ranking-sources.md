# Ranking sources

DraftMeld starts with five transparent ranking signals. Four come from two open-data ecosystems; CBS is a proprietary ranking retrieved from its public provider page on demand and is not redistributed by DraftMeld.

| DraftMeld source | Signal | Project and license |
| --- | --- | --- |
| Redraft expert consensus | Current overall expert consensus for QB, RB, WR, and TE | [DynastyProcess data](https://github.com/dynastyprocess/data), GPL-3.0 repository with upstream FantasyPros attribution |
| Dynasty market - 1 QB | DynastyProcess normalized `value_1qb` | [DynastyProcess data](https://github.com/dynastyprocess/data), GPL-3.0 |
| Dynasty market - Superflex | DynastyProcess normalized `value_2qb` | [DynastyProcess data](https://github.com/dynastyprocess/data), GPL-3.0 |
| Expected opportunity | Prior-season ffopportunity expected fantasy points aggregated by player | [ffopportunity](https://github.com/ffverse/ffopportunity), CC-BY-SA-4.0 data |
| CBS Sports PPR Top 200 | Current CBS Fantasy Experts consensus order | [CBS Sports rankings](https://www.cbssports.com/fantasy/football/rankings/), proprietary; on-demand personal retrieval |

## Import behavior

- Data is fetched from fixed public source URLs only when a user selects **Refresh all sources**. Third-party CSV or HTML files are not committed to DraftMeld.
- Provider columns are converted to a small common record: source, normalized player key, display name, position, team, and ordinal rank.
- Each source is replaced transactionally. A later source failing does not roll back sources that refreshed successfully earlier in the same request.
- The current redraft feed anchors eligibility so prior-season or dynasty-only names cannot enter the draft board by themselves.
- The UI exposes methodology, license, project link, default weight, publication date, refresh time, and record count.

## Known limitations

The first matcher uses a normalized player name because these feeds do not share one universal identifier. Suffixes, name changes, and collisions can prevent a valid match. Future work will introduce a canonical player table, provider identifiers, a review queue for uncertain matches, league-specific source weights, and normalized/robust consensus methods.

Source terms and upstream availability can change. Maintainers should verify licenses and attribution before adding a connector, and should never commit or redistribute paid rankings.

## Platform coverage

- **CBS Sports:** connected to the current public PPR Top 200. The parser deliberately takes the first consensus group and ignores the separate expert lists later in the page.
- **Yahoo Fantasy:** requires the official OAuth 2.0 API because public pre-draft pages are tied to individual league settings. Integration can proceed after the DraftMeld application is approved and credentials can be configured securely.
- **NFL.com:** not included in consensus yet because its public overall draft table still identifies the prior-season board. DraftMeld will not treat that as a current 2026 ranking.
- **ESPN:** not automatically retrieved because the current free 2026 material is positional/article/PDF content rather than a stable overall export, and ESPN's published crawler policy restricts automated AI retrieval. A future user-supplied import can support material the user is entitled to use.
