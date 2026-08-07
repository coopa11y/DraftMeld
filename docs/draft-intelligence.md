# Draft intelligence

DraftMeld keeps rankings, projections, market timing, and recommendations as separate explainable layers. No AI service is required to reproduce a recommendation.

## Consensus v2

Each built-in signal has a role:

- `ranking` sources define current redraft preference and conservatively score an omitted eligible player below the source's published list.
- `market` sources add dynasty or trade-market context only when they rank a player.
- `usage` sources add historical opportunity context only when they rank a player.

Source ranks are normalized to the current eligible player-pool depth before combination, so rank 100 in a 200-player list is comparable with rank 250 in a 500-player list. A league can select weighted median, trimmed mean, or weighted average. Every player reports source coverage, normalized rank range, and high, medium, or low confidence. The current redraft feeds remain the eligibility anchor so stale dynasty-only players cannot enter the board.

## Projection CSV format

Projection files are user-supplied CSVs. Required headers are `name`, `position`, and `team`. Optional metadata headers are `adp` and `byeWeek`. Statistic headers use the same names as league scoring rules:

```text
reception,passingYard,passingTouchdown,interception,rushingYard,rushingTouchdown,receivingYard,receivingTouchdown,fieldGoalMade,extraPointMade,defenseSack,defenseInterception,defenseFumbleRecovery,defenseTouchdown,defenseSafety
```

Multiple imported projection sources are averaged per player. DraftMeld then multiplies each projected statistic by the active league's scoring value. Projection data remains distinct from ordinal consensus ranks.

## Replacement value and tiers

Replacement demand starts with every dedicated starting slot across all teams. FLEX and SUPERFLEX slots are allocated one at a time to the position with the highest next projected player, matching how a legal lineup is actually filled. VOR is the player's league-scored projection minus the first player beyond that position's allocated starter demand.

Position tiers split when the VOR curve has a material absolute or proportional drop. Auction values reserve one dollar per roster spot and distribute the remaining league budget in proportion to positive VOR. Inflation compares remaining league dollars with the baseline value of the remaining player pool.

## Recommendations

Recommendations are recalculated after every draft event. Their visible reasons can include:

- projected points over replacement;
- an open starting slot;
- value against ADP;
- likelihood of surviving to the user's next pick;
- recent positional runs and tier drop-offs;
- a persisted target or avoid preference;
- baseline auction value before live inflation.

Mock opponents are deterministic and combine ADP with a small rotating position preference. Sleeper synchronization performs GET requests against the public draft-picks endpoint and only imports known canonical players. DraftMeld never submits a pick.

## Identity review

Player names, common suffixes, positions, and NFL defense aliases are normalized into canonical keys. Similar names sharing a team and position are surfaced in the identity-review queue for a human decision. Acknowledgements are persisted so exceptions remain auditable.

## Current limitations

- Projection imports currently use the documented canonical headers rather than an interactive column mapper.
- Sleeper synchronization is user-triggered rather than a continuous polling connection.
- Identity review can confirm that candidates are separate; explicit alias merging remains future work.
- Mock opponents provide varied deterministic behavior, not historical manager-specific models.
