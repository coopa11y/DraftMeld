# Draft intelligence

DraftMeld keeps rankings, projections, market timing, and recommendations as separate explainable layers. No AI service is required to reproduce a recommendation.

## Consensus v2

Each built-in signal has a role:

- `ranking` sources define current redraft preference and conservatively score an omitted eligible player below the source's published list.
- `market` sources add dynasty or trade-market context only when they rank a player.
- `usage` sources add historical opportunity context only when they rank a player.

Source ranks are normalized to the current eligible player-pool depth before combination, so rank 100 in a 200-player list is comparable with rank 250 in a 500-player list. A league can select weighted median, trimmed mean, or weighted average. Every player reports source coverage, normalized rank range, and high, medium, or low confidence. The current redraft feeds remain the eligibility anchor so stale dynasty-only players cannot enter the board.

## Projection CSV format

Projection files are user-supplied CSVs. The import screen detects the source headers and lets the user map them to DraftMeld's player identity, metadata, and scoring fields. Required mappings are `name`, `position`, and `team`; `adp`, `byeWeek`, and statistic mappings are optional:

```text
reception,passingYard,passingTouchdown,interception,rushingYard,rushingTouchdown,receivingYard,receivingTouchdown,fieldGoalMade,extraPointMade,defenseSack,defenseInterception,defenseFumbleRecovery,defenseTouchdown,defenseSafety
```

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

Player names, common suffixes, positions, and NFL defense aliases are normalized into canonical keys. Similar names sharing a team and position are surfaced in the identity-review queue for a human decision. A user can keep candidates separate or merge aliases under one selected canonical player; ranking and projection signals then resolve through the persisted alias map.

## Validation scenarios

Golden backend scenarios lock expected behavior for source normalization, robust consensus, roster-aware VOR, targets and avoids, next-pick timing, keeper-adjusted auction inflation, identity aliases, and projection scoring. API workflow coverage exercises projection import, preferences, a user pick, mock opponents, Sleeper reconciliation, and undo against a real temporary SQLite database. Frontend interaction tests cover the accessible column mapper, identity merging, sync reconciliation messages, draft focus management, source weighting, and axe checks.

## Current limitations

- Mock opponents provide varied deterministic behavior, not historical manager-specific models.
- Sleeper polling runs only while the draft page is open and depends on the public Sleeper API.
- Keeper calibration accepts aggregate spend and removed value; selecting the actual kept-player set is future work.
