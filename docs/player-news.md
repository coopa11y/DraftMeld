# Player news and availability

DraftMeld keeps structured player availability separate from news stories. An injury or practice status can be shown even when no article exists, while an article does not silently mark a player unavailable or remove them from recommendations.

## Built-in sources

- **Sleeper player status** provides structured NFL team, injury, practice, and availability fields. It refreshes daily by default.
- **ESPN Fantasy Football RSS** provides attributed fantasy-football stories and refreshes every 10 minutes by default.
- **PFF Fantasy Football RSS** provides attributed fantasy-football stories and refreshes every 15 minutes by default.

Source availability and terms can change. Each adapter is isolated so an upstream failure is recorded against that source without blocking the draft board or other sources.

## Draft-board experience

The player table has no permanent news column or feed. A player with a meaningful status or linked story receives a concise status and one **Details** button. Activating it opens a focused dialog containing the latest structured status and up to five recent, dated stories with source attribution. Closing the dialog returns the user to the triggering control.

An active player with no linked stories adds no extra control. DraftMeld does not automatically draft, hide, or downgrade a player based only on a headline.

## Source management

**Player news sources** is available from the league overview. The compact table exposes one enable control, freshness or error status, and removal only for user-added feeds. **Refresh now** updates enabled sources without requiring a background worker or external account.

Users can add a public HTTPS RSS or Atom URL. DraftMeld validates the hostname, blocks private and local network destinations and unsafe redirects, limits downloads to 20 MiB, and stores normalized events rather than authentication data. Built-in sources cannot be deleted, but they can be disabled.

## Player matching and confidence

Structured Sleeper records use the canonical directory's provider IDs when available. News titles and summaries are matched against normalized canonical player names. Exact name matches receive high confidence; ambiguous or unmatched stories are not attached to a player. Team changes remain canonical-directory concerns, so a provider's current team can update without creating a second player identity.

## Current limitations

- RSS is supplemental and may omit or delay a story.
- Public feeds do not provide a complete transaction wire, official game designation history, or every practice report.
- Player-name linking deliberately favors missed stories over attaching a story to the wrong player.
- Recommendation scores remain explainable and league-driven; news is informational until a structured, tested adjustment policy is added.
