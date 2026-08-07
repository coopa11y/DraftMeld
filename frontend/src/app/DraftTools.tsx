import type { DraftSnapshot } from "../shared/api/types";
import { Button } from "../shared/ui/Button";
import { Panel } from "../shared/ui/Panel";
import { SleeperSyncForm } from "./SleeperSyncForm";

interface DraftToolsProps {
  snapshot: DraftSnapshot;
  busy: boolean;
  onMock: () => void;
  onSleeperSync: (draftId: string, rosterId: number) => Promise<void>;
}

export function DraftTools({ snapshot, busy, onMock, onSleeperSync }: DraftToolsProps) {
  return (
    <Panel variant="side" aria-labelledby="draft-tools-title">
      <p className="eyebrow">Optional automation</p>
      <h2 id="draft-tools-title">Draft tools</h2>
      {snapshot.draftType === "auction"
        ? <AuctionSummary snapshot={snapshot} />
        : <MockDraftControl snapshot={snapshot} busy={busy} onMock={onMock} />}
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

function AuctionSummary({ snapshot }: Pick<DraftToolsProps, "snapshot">) {
  return (
    <dl className="auction-summary">
      <div><dt>My budget</dt><dd>${snapshot.budgetRemaining.toFixed(0)} of ${snapshot.auctionBudget.toFixed(0)}</dd></div>
      <div><dt>Maximum bid</dt><dd>${snapshot.maximumBid.toFixed(0)}</dd></div>
      <div><dt>Market inflation</dt><dd>{snapshot.auctionInflation.toFixed(2)}×</dd></div>
    </dl>
  );
}
