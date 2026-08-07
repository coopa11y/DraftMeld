# Data portability

DraftMeld keeps league configuration and draft results portable without requiring an account or hosted service. Open **Manage leagues** and use the **Export and backup center** to download or restore data.

## League backups

A league backup is versioned JSON containing:

- league, roster, scoring, auction, and keeper rules;
- ranking-source inclusion and weights;
- consensus method and target or avoid preferences;
- recommendation policy settings;
- the original league identifier and export timestamp.

Restoring a backup always creates a new league with a unique identifier. It never overwrites an existing league. Version `1` is the only currently supported backup format, and uploads are limited to 256 KiB.

League backups intentionally do not contain imported provider data or draft history. Ranking and projection data may have separate licensing restrictions, while draft results have dedicated exports.

## Consensus ranking CSV

The consensus export includes each player's normalized consensus rank and score, position, team, coverage, disagreement range, confidence, method, enabled source weights, and available per-source ranks. CSV text cells that spreadsheet applications could interpret as formulas are escaped before download.

## Draft result exports

The CSV export provides a compact pick-by-pick table suitable for a spreadsheet. The versioned JSON export includes pick history, player details, costs, timestamps, and the user's current roster.

Exports reflect the current active draft state. Picks removed through Undo or Sleeper reconciliation are not included.

## Compatibility

The `formatVersion` field controls compatibility independently of the DraftMeld application version. Future breaking backup changes will use a new format version rather than silently interpreting older data differently.
