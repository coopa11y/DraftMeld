# Ranking sources

DraftMeld starts with four transparent ranking signals. These are four distinct methods from two open-data ecosystems, not four independent publishers.

| DraftMeld source | Signal | Project and license |
| --- | --- | --- |
| Redraft expert consensus | Current overall expert consensus for QB, RB, WR, and TE | [DynastyProcess data](https://github.com/dynastyprocess/data), GPL-3.0 repository with upstream FantasyPros attribution |
| Dynasty market - 1 QB | DynastyProcess normalized `value_1qb` | [DynastyProcess data](https://github.com/dynastyprocess/data), GPL-3.0 |
| Dynasty market - Superflex | DynastyProcess normalized `value_2qb` | [DynastyProcess data](https://github.com/dynastyprocess/data), GPL-3.0 |
| Expected opportunity | Prior-season ffopportunity expected fantasy points aggregated by player | [ffopportunity](https://github.com/ffverse/ffopportunity), CC-BY-SA-4.0 data |

## Import behavior

- Data is fetched from fixed public project URLs only when a user selects **Refresh all sources**. Third-party CSV files are not committed to DraftMeld.
- Provider columns are converted to a small common record: source, normalized player key, display name, position, team, and ordinal rank.
- Each source is replaced transactionally. A later source failing does not roll back sources that refreshed successfully earlier in the same request.
- The current redraft feed anchors eligibility so prior-season or dynasty-only names cannot enter the draft board by themselves.
- The UI exposes methodology, license, project link, default weight, publication date, refresh time, and record count.

## Known limitations

The first matcher uses a normalized player name because these feeds do not share one universal identifier. Suffixes, name changes, and collisions can prevent a valid match. Future work will introduce a canonical player table, provider identifiers, a review queue for uncertain matches, league-specific source weights, and normalized/robust consensus methods.

Source terms and upstream availability can change. Maintainers should verify licenses and attribution before adding a connector, and should never commit or redistribute paid rankings.
