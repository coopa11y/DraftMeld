import type { DraftAction, DraftSnapshot, Player } from "../shared/api/types";
import { Panel } from "../shared/ui/Panel";
import { PlayerActions } from "./PlayerActions";
import { DraftTools } from "./DraftTools";

interface DraftSidebarProps {
  snapshot: DraftSnapshot;
  busy: boolean;
  opponentTeamNumber: number;
  onOpponentTeamChange: (teamNumber: number) => void;
  onAction: (player: Player, action: DraftAction, cost?: number, teamNumber?: number) => void;
  onMock: () => void;
  onSleeperSync: (draftId: string, rosterId: number) => Promise<void>;
}

export function DraftSidebar({
  snapshot,
  busy,
  opponentTeamNumber,
  onOpponentTeamChange,
  onAction,
  onMock,
  onSleeperSync,
}: DraftSidebarProps) {
  return (
    <aside className="sidebar" aria-label="Draft assistant">
      <Panel variant="side" id="recommendations" aria-labelledby="recommendations-title">
        <p className="eyebrow">Updated after every pick</p>
        <h2 id="recommendations-title">Recommended</h2>
        <ol className="recommendation-list">
          {snapshot.recommendations.map((recommendation) => (
            <li key={recommendation.player.id}>
              <article>
                <h3>
                  {recommendation.player.name}{" "}
                  <span>
                    {recommendation.player.position}
                    {recommendation.player.positionRank}
                  </span>
                </h3>
                <ul className="reason-list">
                  {recommendation.reasons.map((reason) => (
                    <li key={reason}>{reason}</li>
                  ))}
                </ul>
                <PlayerActions
                  player={recommendation.player}
                  busy={busy}
                  auction={snapshot.draftType === "auction"}
                  inflation={snapshot.auctionInflation}
                  minimumBid={snapshot.auctionMinimumBid}
                  maximumBid={snapshot.maximumBid}
                  opponentTeamNumber={opponentTeamNumber}
                  isUserTurn={snapshot.isUserTurn}
                  isComplete={snapshot.isComplete}
                  onAction={onAction}
                />
              </article>
            </li>
          ))}
        </ol>
      </Panel>

      <DraftTools
        snapshot={snapshot}
        busy={busy}
        opponentTeamNumber={opponentTeamNumber}
        onOpponentTeamChange={onOpponentTeamChange}
        onMock={onMock}
        onSleeperSync={onSleeperSync}
      />

      <Panel variant="side" id="my-team" aria-labelledby="my-team-title">
        <p className="eyebrow">{snapshot.myTeam.length} players</p>
        <h2 id="my-team-title">My team</h2>
        {snapshot.myTeam.length ? (
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

      <Panel variant="side" aria-labelledby="history-title">
        <h2 id="history-title">Draft history</h2>
        {snapshot.history.length ? (
          <ol className="history-list">
            {[...snapshot.history]
              .reverse()
              .slice(0, 8)
              .map((pick) => (
                <li key={pick.eventId}>
                  <strong>Pick {pick.number}:</strong> {pick.player.name} {`to ${pick.teamName}`}
                  {snapshot.draftType === "auction" ? ` for $${pick.cost.toFixed(0)}` : ""}
                </li>
              ))}
          </ol>
        ) : (
          <p className="empty-state">No picks recorded yet.</p>
        )}
      </Panel>
    </aside>
  );
}
