import { useState, type FormEvent } from "react";
import type { DraftSnapshot } from "../shared/api/types";

interface DraftToolsProps {
  snapshot: DraftSnapshot;
  busy: boolean;
  onMock: () => void;
  onSleeperSync: (draftId: string, rosterId: number) => void;
}

export function DraftTools({ snapshot, busy, onMock, onSleeperSync }: DraftToolsProps) {
  const [draftId, setDraftId] = useState("");
  const [rosterId, setRosterId] = useState(1);
  function submit(event: FormEvent) {
    event.preventDefault();
    onSleeperSync(draftId, rosterId);
  }
  return (
    <section className="side-panel" aria-labelledby="draft-tools-title">
      <p className="eyebrow">Optional automation</p><h2 id="draft-tools-title">Draft tools</h2>
      {snapshot.draftType !== "auction" ? <><button className="secondary-button full-width-button" type="button" disabled={busy || snapshot.nextUserPick === 0 || snapshot.isUserTurn} onClick={onMock}>Simulate to my next turn{snapshot.nextUserPick ? ` (pick ${snapshot.nextUserPick})` : ""}</button>{snapshot.isUserTurn ? <p className="tool-help">Make your pick before simulating opponent selections.</p> : null}</> : <dl className="auction-summary"><div><dt>My budget</dt><dd>${snapshot.budgetRemaining.toFixed(0)} of ${snapshot.auctionBudget.toFixed(0)}</dd></div><div><dt>Market inflation</dt><dd>{snapshot.auctionInflation.toFixed(2)}×</dd></div></dl>}
      <form className="sleeper-sync-form" onSubmit={submit}><h3>Sleeper live sync</h3><p>Imports public picks only. DraftMeld never submits a selection.</p><label>Draft ID<input required value={draftId} onChange={(event) => setDraftId(event.target.value)} /></label><label>Your roster ID<input required type="number" min="1" value={rosterId} onChange={(event) => setRosterId(Number(event.target.value))} /></label><button className="secondary-button" type="submit" disabled={busy || !draftId.trim()}>Sync picks</button></form>
    </section>
  );
}
