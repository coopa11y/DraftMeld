# Draft intelligence

DraftMeld keeps rankings, projections, market timing, and recommendations as separate explainable layers. No AI service is required to reproduce a recommendation.

## Consensus v2

Each built-in signal has a role:

- `ranking` sources define current redraft preference and conservatively score an omitted eligible player below the source's published list.
- `market` sources add dynasty or trade-market context only when they rank a player.
- `usage` sources add historical opportunity context only when they rank a player.

Source ranks are normalized to the current eligible player-pool depth before combination, so rank 100 in a 200-player list is comparable with rank 250 in a 500-player list. A league can select weighted median, trimmed mean, or weighted average. Every player reports source coverage, normalized rank range, and high, medium, or low confidence. The current redraft feeds remain the eligibility anchor so stale dynasty-only players cannot enter the board.

## Projection CSV format

Projection sources can be refreshed from Sleeper or supplied as user CSVs. The CSV import screen detects the source headers and lets the user map them to DraftMeld's player identity, metadata, and scoring fields. Required mappings are `name`, `position`, and `team`; `adp`, `byeWeek`, and every statistic mapping are optional. Missing statistics contribute zero, so a league and its projection source can be as simple or detailed as needed.

Core statistics include:

```text
reception,passingYard,passingTouchdown,interception,rushingYard,rushingTouchdown,receivingYard,receivingTouchdown,fieldGoalMade,extraPointMade,defenseSack,defenseInterception,defenseFumbleRecovery,defenseTouchdown,defenseSafety
```

Expanded optional statistics include passing, rushing, and receiving two-point conversions; fumbles and fumbles lost; 300/400-yard passing games; 100/200-yard rushing and receiving games; field-goal distance bands and misses; blocked kicks and defensive two-point returns; and games in configurable points-allowed bands. TE premium does not require a separate projection column: DraftMeld applies the configured bonus to receptions by players at TE.

Every scoring value can be positive, negative, or zero. Zero disables a category. Field goals can use one any-distance value or distance bands; setting the unused approach to zero prevents double scoring. Milestone columns are event counts and apply independently, allowing leagues to make higher thresholds cumulative or non-cumulative through their projection data.

Multiple imported projection sources are averaged per player. DraftMeld then multiplies each projected statistic by the active league's scoring value. Projection data remains distinct from ordinal consensus ranks.

## Replacement value and tiers

Replacement demand starts with every dedicated starting slot across all teams. FLEX and SUPERFLEX slots are allocated one at a time to the position with the highest next projected player, matching how a legal lineup is actually filled. VOR is the player's league-scored projection minus the first player beyond that position's allocated starter demand.

Position tiers split when the VOR curve has a material absolute or proportional drop. Auction values reserve the configured minimum bid for every roster spot and distribute the remaining league budget in proportion to positive VOR. Inflation compares remaining league dollars with the baseline value of the remaining player pool, including league-wide keeper spend and removed keeper value. Personal keeper spend reduces the user's remaining budget, and DraftMeld calculates the current maximum legal bid while reserving enough money to complete the roster.

## Recommendations

Recommendations are recalculated after every draft event. Their visible reasons can include:

- projected points over replacement;
- an open starting slot;
- value against ADP;
- likelihood of surviving to the user's next pick;
- recent positional runs and tier drop-offs;
- a persisted target or avoid preference;
- baseline auction value before live inflation.

Mock opponents are deterministic and combine ADP with a small rotating position preference. Sleeper synchronization performs GET requests against the public draft-picks endpoint and only imports known canonical players. Sleeper is authoritative while connected: changed and deleted remote picks are reconciled, unmatched players are reported, and the user can enable 15-second polling while the page remains open. DraftMeld never submits a pick.

## Identity review

Ranking and projection imports resolve into a persistent canonical player directory. Stable DraftMeld player IDs survive source-name changes, normalized names remain aliases for compatibility, and optional provider player IDs offer the strongest match when a source supplies them. Similar names sharing a team and position are surfaced in the identity-review queue for a human decision. A user can keep candidates separate or merge aliases under one selected canonical player; the merge also redirects stored provider identifiers.

## Validation scenarios

Golden backend scenarios lock expected behavior for source normalization, robust consensus, roster-aware VOR, targets and avoids, next-pick timing, keeper-adjusted auction inflation, identity aliases, and projection scoring. API workflow coverage exercises projection import, preferences, a user pick, mock opponents, Sleeper reconciliation, and undo against a real temporary SQLite database. Frontend interaction tests cover the accessible column mapper, identity merging, sync reconciliation messages, draft focus management, source weighting, and axe checks.

## Current limitations

- Mock opponents provide varied deterministic behavior, not historical manager-specific models.
- Sleeper polling runs only while the draft page is open and depends on the public Sleeper API.
- Keeper calibration accepts aggregate spend and removed value; selecting the actual kept-player set is future work.
