import type { DraftSnapshot, WatchlistPlayer } from "../shared/api/types";
import { Panel } from "../shared/ui/Panel";

interface DraftFocusSidebarProps {
  snapshot: DraftSnapshot;
  watchlist: WatchlistPlayer[];
}

export function DraftFocusSidebar({ snapshot, watchlist }: DraftFocusSidebarProps) {
  const availableNames = new Set(snapshot.available.map((player) => `${player.name}|${player.position}`));
  const availableOutliers = watchlist
    .filter((player) => availableNames.has(`${player.name}|${player.position}`))
    .slice(0, 5);
  return (
    <aside className="draft-focus-sidebar" aria-label="My draft summary">
      <Panel variant="side" id="my-team">
        <p className="eyebrow">{snapshot.myTeam.length} players</p>
        <h2>My team</h2>
        {snapshot.myTeam.length > 0 ? (
          <ul className="team-list">
            {snapshot.myTeam.map((player) => (
              <li key={player.id}>
                <strong>{player.position}</strong>
                <span>
                  {player.name}, {player.nflTeam}
                </span>
              </li>
            ))}
          </ul>
        ) : (
          <p className="empty-state">No players drafted yet.</p>
        )}
      </Panel>
      {availableOutliers.length > 0 ? (
        <Panel variant="side" id="draft-outliers">
          <p className="eyebrow">Excluded-source signals</p>
          <h2>Outliers</h2>
          <ul className="draft-outlier-list">
            {availableOutliers.map((player) => (
              <li key={player.playerKey}>
                <strong>{player.name}</strong>
                <span>
                  {player.position} · {player.team}
                </span>
                <small>
                  {player.signals[0]?.spotsHigher} spots above consensus in {player.signals[0]?.sourceName}
                </small>
              </li>
            ))}
          </ul>
        </Panel>
      ) : null}
    </aside>
  );
}
