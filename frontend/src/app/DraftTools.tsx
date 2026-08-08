import type { DraftSnapshot } from "../shared/api/types";
import { Button } from "../shared/ui/Button";
import { Panel } from "../shared/ui/Panel";
import { SleeperSyncForm } from "./SleeperSyncForm";

interface DraftToolsProps {
  snapshot: DraftSnapshot;
  busy: boolean;
  onMock: () => void;
  opponentTeamNumber: number;
  onOpponentTeamChange: (teamNumber: number) => void;
  onSleeperSync: (draftId: string, rosterId: number) => Promise<void>;
}

export function DraftTools({
  snapshot,
  busy,
  opponentTeamNumber,
  onOpponentTeamChange,
  onMock,
  onSleeperSync,
}: DraftToolsProps) {
  return (
    <Panel variant="side" aria-labelledby="draft-tools-title">
      <p className="eyebrow">Optional automation</p>
      <h2 id="draft-tools-title">Draft tools</h2>
      {snapshot.draftType === "auction" ? (
        <AuctionSummary
          snapshot={snapshot}
          busy={busy}
          opponentTeamNumber={opponentTeamNumber}
          onOpponentTeamChange={onOpponentTeamChange}
        />
      ) : (
        <MockDraftControl snapshot={snapshot} busy={busy} onMock={onMock} />
      )}
      <SleeperSyncForm busy={busy} onSync={onSleeperSync} />
    </Panel>
  );
}

function MockDraftControl({ snapshot, busy, onMock }: Pick<DraftToolsProps, "snapshot" | "busy" | "onMock">) {
  return (
    <>
      <Button fullWidth disabled={busy || snapshot.nextUserPick === 0 || snapshot.isUserTurn} onClick={onMock}>
        Simulate to my next turn{snapshot.nextUserPick ? ` (pick ${snapshot.nextUserPick})` : ""}
      </Button>
      {snapshot.isUserTurn ? <p className="tool-help">Make your pick before simulating opponent selections.</p> : null}
    </>
  );
}

function AuctionSummary({
  snapshot,
  busy,
  opponentTeamNumber,
  onOpponentTeamChange,
}: Pick<DraftToolsProps, "snapshot" | "busy" | "opponentTeamNumber" | "onOpponentTeamChange">) {
  return (
    <>
      <label className="auction-team-select">
        <span>Opponent winning the next player</span>
        <select
          value={opponentTeamNumber}
          disabled={busy}
          onChange={(event) => onOpponentTeamChange(Number(event.target.value))}
        >
          {snapshot.teams
            .filter((team) => !team.isUser)
            .map((team) => (
              <option key={team.number} value={team.number}>
                {team.name}
              </option>
            ))}
        </select>
      </label>
      <dl className="auction-summary">
        <div>
          <dt>My budget</dt>
          <dd>
            ${snapshot.budgetRemaining.toFixed(0)} of ${snapshot.auctionBudget.toFixed(0)}
          </dd>
        </div>
        <div>
          <dt>Maximum bid</dt>
          <dd>${snapshot.maximumBid.toFixed(0)}</dd>
        </div>
        <div>
          <dt>Market inflation</dt>
          <dd>{snapshot.auctionInflation.toFixed(2)}×</dd>
        </div>
      </dl>
    </>
  );
}
