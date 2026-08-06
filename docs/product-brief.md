# Product brief

## Problem

Fantasy football players routinely juggle expert rankings, projections, platform ADP, spreadsheets, league settings, and a live draft room. Most tools either lock users into one data provider or hide how their recommendations are produced.

## Vision

DraftMeld gives every league a transparent, customizable draft model. It turns heterogeneous sources into a normalized player pool, applies the league's exact rules, and presents one explainable live board across multiple leagues and teams.

## Primary users

- A manager drafting several leagues on different platforms
- A commissioner with unusual scoring or roster rules
- An analyst who wants to control source weights and ranking methodology
- A self-hoster who wants local ownership of league and draft data

## Initial release scope

1. Create and manage multiple league profiles.
2. Configure scoring and roster rules.
3. Import multiple ranking or projection CSV files with column mapping.
4. Resolve imported players against a canonical player directory.
5. Build an explainable consensus board with adjustable source weights.
6. Run a manual snake draft with persistent pick history and undo.
7. Show tiers, VOR, positional scarcity, source disagreement, and roster needs.
8. Export league configuration, consensus rankings, and draft results.

## Later milestones

- Sleeper live draft synchronization
- ESPN and Yahoo companion integrations where technically permitted
- Auction values, inflation, and budget strategy
- Keeper and dynasty valuation
- Optional AI-assisted comparisons and strategy explanations
- Historical source-accuracy weighting

## Non-goals for the first release

- Automatically submitting draft picks
- Redistributing paid projection or ranking data
- Pretending unsupported scoring rules are equivalent to standard formats
- Making opaque recommendations that cannot be reproduced without an AI model

## Success criteria

- A user can configure an unusual league without editing code.
- A player can be reconciled across multiple imported sources with reviewable exceptions.
- Every consensus rank can be traced to its source values and weights.
- Draft state survives refreshes, reconnects, and configuration navigation.
- The application remains usable with manual imports and no paid services.
