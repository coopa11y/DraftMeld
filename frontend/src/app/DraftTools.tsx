import type { DraftSnapshot } from "../shared/api/types";
import { Button } from "../shared/ui/Button";
import { Panel } from "../shared/ui/Panel";
import { SleeperSyncForm } from "./SleeperSyncForm";

interface DraftToolsProps {
  snapshot: DraftSnapshot;
  busy: boolean;
  onMock: () => void;
  selectedTeamNumber: number;
  onSelectedTeamChange: (teamNumber: number) => void;
  onSleeperSync: (draftId: string, rosterId: number) => Promise<void>;
}

export function DraftTools({
  snapshot,
  busy,
  selectedTeamNumber,
  onSelectedTeamChange,
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
          selectedTeamNumber={selectedTeamNumber}
          onSelectedTeamChange={onSelectedTeamChange}
        />
      ) : (
        <>
          <PickOwnerControl
            snapshot={snapshot}
            busy={busy}
            selectedTeamNumber={selectedTeamNumber}
            onSelectedTeamChange={onSelectedTeamChange}
          />
          <MockDraftControl snapshot={snapshot} busy={busy} onMock={onMock} />
        </>
      )}
      <SleeperSyncForm busy={busy} onSync={onSleeperSync} />
    </Panel>
  );
}

function PickOwnerControl({
  snapshot,
  busy,
  selectedTeamNumber,
  onSelectedTeamChange,
}: Pick<DraftToolsProps, "snapshot" | "busy" | "selectedTeamNumber" | "onSelectedTeamChange">) {
  const scheduledTeam = snapshot.teams.find((team) => team.number === snapshot.onClockTeamNumber);
  const selectedTeam = snapshot.teams.find((team) => team.number === selectedTeamNumber);
  return (
    <label className="auction-team-select">
      <span>Owner of pick {snapshot.pickNumber}</span>
      <select
        value={selectedTeamNumber}
        disabled={busy || snapshot.isComplete}
        onChange={(event) => onSelectedTeamChange(Number(event.target.value))}
      >
        {snapshot.teams.map((team) => (
          <option key={team.number} value={team.number}>
            {team.name}
          </option>
        ))}
      </select>
      <small>
        {selectedTeamNumber !== snapshot.onClockTeamNumber
          ? `Traded from ${scheduledTeam?.name ?? "the scheduled team"} to ${selectedTeam?.name ?? "the selected team"}.`
          : "Change this only when the current pick was traded."}
      </small>
    </label>
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
  selectedTeamNumber,
  onSelectedTeamChange,
}: Pick<DraftToolsProps, "snapshot" | "busy" | "selectedTeamNumber" | "onSelectedTeamChange">) {
  return (
    <>
      <label className="auction-team-select">
        <span>Opponent winning the next player</span>
        <select
          value={selectedTeamNumber}
          disabled={busy}
          onChange={(event) => onSelectedTeamChange(Number(event.target.value))}
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
